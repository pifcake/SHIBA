package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SHIBA/internal/domain"
	"SHIBA/internal/repository"
	"SHIBA/pkg/hash"
	jwtpkg "SHIBA/pkg/jwt"
	"SHIBA/pkg/mail"

	"github.com/google/uuid"
)

var (
	ErrEmailNotVerified  = errors.New("email not verified")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired or used")
	ErrTokenRevoked      = errors.New("token revoked")
	ErrAccountBlocked    = errors.New("account is blocked")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtManager  *jwtpkg.Manager
	mailSender  *mail.Sender
	frontendURL string
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtManager *jwtpkg.Manager,
	mailSender *mail.Sender,
	frontendURL string,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtManager:  jwtManager,
		mailSender:  mailSender,
		frontendURL: frontendURL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error) {
	passwordHash, err := hash.Password(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       domain.StatusPendingEmail,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send email verification
	if err := s.sendVerificationEmail(ctx, user); err != nil {
		// Don't fail registration if email fails, but log it
		fmt.Printf("WARNING: failed to send verification email to %s: %v\n", email, err)
	}

	return user, nil
}

func (s *AuthService) sendVerificationEmail(ctx context.Context, user *domain.User) error {
	token, err := hash.RandomToken(32)
	if err != nil {
		return err
	}

	v := &domain.EmailVerification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.userRepo.CreateVerification(ctx, v); err != nil {
		return err
	}

	return s.mailSender.SendVerification(user.Email, token, s.frontendURL)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	v, err := s.userRepo.GetVerificationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTokenExpired
		}
		return err
	}

	if v.UsedAt != nil {
		return ErrTokenExpired
	}
	if time.Now().After(v.ExpiresAt) {
		return ErrTokenExpired
	}

	if err := s.userRepo.MarkVerificationUsed(ctx, v.ID); err != nil {
		return err
	}

	return s.userRepo.UpdateStatus(ctx, v.UserID, domain.StatusActive)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user *domain.User, err error) {
	user, err = s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", "", nil, ErrInvalidCredentials
		}
		return "", "", nil, err
	}

	if !hash.CheckPassword(password, user.PasswordHash) {
		return "", "", nil, ErrInvalidCredentials
	}

	if user.Status == domain.StatusPendingEmail {
		return "", "", nil, ErrEmailNotVerified
	}
	if user.Status == domain.StatusBlocked {
		return "", "", nil, ErrAccountBlocked
	}

	accessToken, err = s.jwtManager.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("generate access token: %w", err)
	}

	rawRefresh, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("generate refresh token: %w", err)
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash.HashToken(rawRefresh),
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTTL()),
	}
	if err := s.userRepo.CreateRefreshToken(ctx, rt); err != nil {
		return "", "", nil, err
	}

	return accessToken, rawRefresh, user, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (accessToken, newRefreshToken string, err error) {
	if _, err = s.jwtManager.ParseRefreshToken(rawRefreshToken); err != nil {
		return "", "", ErrTokenExpired
	}

	tokenHash := hash.HashToken(rawRefreshToken)
	stored, err := s.userRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", "", ErrTokenRevoked
		}
		return "", "", err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return "", "", ErrTokenRevoked
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return "", "", err
	}
	if user.Status == domain.StatusBlocked {
		return "", "", ErrAccountBlocked
	}

	// Rotate: revoke old, issue new
	if err = s.userRepo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return "", "", fmt.Errorf("revoke old refresh token: %w", err)
	}

	accessToken, err = s.jwtManager.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err = s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash.HashToken(newRefreshToken),
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTTL()),
	}
	if err = s.userRepo.CreateRefreshToken(ctx, rt); err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.RevokeAllRefreshTokens(ctx, userID)
}

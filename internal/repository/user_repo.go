package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SHIBA/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	q := `INSERT INTO users (id, email, password_hash, role, status)
		  VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, q, user.ID, user.Email, user.PasswordHash, user.Role, user.Status)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	q := `SELECT id, email, password_hash, role, status, created_at, updated_at
		  FROM users WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `SELECT id, email, password_hash, role, status, created_at, updated_at
		  FROM users WHERE email = $1`
	row := r.pool.QueryRow(ctx, q, email)
	return scanUser(row)
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	q := `UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, role string, status string, limit, offset int) ([]domain.User, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if role != "" {
		where += fmt.Sprintf(" AND role = $%d", argIdx)
		args = append(args, role)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT u.id, u.email, u.password_hash, u.role, u.status, u.created_at, u.updated_at,
		  COALESCE(
		    CASE u.role
		      WHEN 'model' THEN (
		        SELECT mp.storage_key FROM model_profiles mpr
		        JOIN model_photos mp ON mp.model_profile_id = mpr.id
		        WHERE mpr.user_id = u.id AND mp.is_cover = true AND mp.moderation_status = 'approved'
		        LIMIT 1
		      )
		      WHEN 'agency' THEN (
		        SELECT ap.logo_url FROM agency_profiles ap WHERE ap.user_id = u.id LIMIT 1
		      )
		      ELSE NULL
		    END, ''
		  ) AS photo_ref
		  FROM users u %s ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.PhotoRef)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user list: %w", err)
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

// Email verifications

func (r *UserRepository) CreateVerification(ctx context.Context, v *domain.EmailVerification) error {
	q := `INSERT INTO email_verifications (id, user_id, token, expires_at)
		  VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, q, v.ID, v.UserID, v.Token, v.ExpiresAt)
	return err
}

func (r *UserRepository) GetVerificationByToken(ctx context.Context, token string) (*domain.EmailVerification, error) {
	q := `SELECT id, user_id, token, expires_at, used_at
		  FROM email_verifications WHERE token = $1`
	row := r.pool.QueryRow(ctx, q, token)
	var v domain.EmailVerification
	err := row.Scan(&v.ID, &v.UserID, &v.Token, &v.ExpiresAt, &v.UsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *UserRepository) MarkVerificationUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	q := `UPDATE email_verifications SET used_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, now, id)
	return err
}

// Refresh tokens

func (r *UserRepository) CreateRefreshToken(ctx context.Context, rt *domain.RefreshToken) error {
	q := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		  VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, q, rt.ID, rt.UserID, rt.TokenHash, rt.ExpiresAt)
	return err
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	q := `SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		  FROM refresh_tokens WHERE token_hash = $1`
	row := r.pool.QueryRow(ctx, q, tokenHash)
	var rt domain.RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	q := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, now, id)
	return err
}

func (r *UserRepository) RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	q := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, q, now, userID)
	return err
}

// Helpers

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return fmt.Sprintf("%s", err) != "" && (contains(err.Error(), "23505") || contains(err.Error(), "unique"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"SHIBA/internal/domain"
	"SHIBA/internal/repository"
	"SHIBA/internal/storage"

	"github.com/google/uuid"
)

type AgencyService struct {
	agencyRepo *repository.AgencyRepository
	storage    *storage.MinIOClient
}

func NewAgencyService(agencyRepo *repository.AgencyRepository, storage *storage.MinIOClient) *AgencyService {
	return &AgencyService{agencyRepo: agencyRepo, storage: storage}
}

func (s *AgencyService) GetOrCreateProfile(ctx context.Context, userID uuid.UUID) (*domain.AgencyProfile, error) {
	p, err := s.agencyRepo.GetByUserID(ctx, userID)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	p = &domain.AgencyProfile{
		ID:     uuid.New(),
		UserID: userID,
		Status: domain.AgencyStatusPendingApproval,
	}
	if err := s.agencyRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AgencyService) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*domain.AgencyProfile, error) {
	return s.agencyRepo.GetByUserID(ctx, userID)
}

func (s *AgencyService) GetProfileByID(ctx context.Context, id uuid.UUID) (*domain.AgencyProfile, error) {
	return s.agencyRepo.GetByID(ctx, id)
}

func (s *AgencyService) UpdateProfile(ctx context.Context, userID uuid.UUID, companyName, description, phone string) (*domain.AgencyProfile, error) {
	p, err := s.agencyRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			p = &domain.AgencyProfile{
				ID:     uuid.New(),
				UserID: userID,
				Status: domain.AgencyStatusPendingApproval,
			}
			if createErr := s.agencyRepo.Create(ctx, p); createErr != nil {
				return nil, createErr
			}
		} else {
			return nil, err
		}
	}

	p.CompanyName = companyName
	p.Description = description
	p.Phone = phone

	if err := s.agencyRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AgencyService) UploadLogo(ctx context.Context, userID uuid.UUID, fh *multipart.FileHeader) (string, error) {
	p, err := s.agencyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	mimeType := fh.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return "", ErrInvalidMimeType
	}

	f, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	result, err := s.storage.UploadLogo(ctx, p.ID, fh.Filename, f, fh.Size, mimeType)
	if err != nil {
		return "", err
	}

	p.LogoURL = result.URL
	if err := s.agencyRepo.Update(ctx, p); err != nil {
		return "", err
	}

	return result.URL, nil
}

func (s *AgencyService) GetPending(ctx context.Context, limit, offset int) ([]domain.AgencyProfile, int, error) {
	return s.agencyRepo.GetPending(ctx, limit, offset)
}

func (s *AgencyService) Approve(ctx context.Context, agencyID, adminID uuid.UUID) error {
	return s.agencyRepo.Approve(ctx, agencyID, adminID)
}

func (s *AgencyService) Reject(ctx context.Context, agencyID uuid.UUID) error {
	return s.agencyRepo.Reject(ctx, agencyID)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"SHIBA/internal/domain"
	"SHIBA/internal/repository"
	"SHIBA/internal/storage"

	"github.com/google/uuid"
)

const (
	MaxPhotoCount = 10
	MaxPhotoSize  = 10 * 1024 * 1024 // 10MB
)

var (
	ErrPhotoLimitReached = errors.New("maximum photo limit reached")
	ErrInvalidMimeType   = errors.New("invalid file type: only JPEG and PNG allowed")
	ErrFileTooLarge      = errors.New("file too large: maximum 10MB allowed")
	ErrNotOwner          = errors.New("not the owner of this resource")
)

type ModelService struct {
	modelRepo *repository.ModelRepository
	storage   *storage.MinIOClient
}

func NewModelService(modelRepo *repository.ModelRepository, storage *storage.MinIOClient) *ModelService {
	return &ModelService{modelRepo: modelRepo, storage: storage}
}

func (s *ModelService) GetOrCreateProfile(ctx context.Context, userID uuid.UUID) (*domain.ModelProfile, error) {
	p, err := s.modelRepo.GetByUserID(ctx, userID)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	// Create empty profile
	p = &domain.ModelProfile{
		ID:                uuid.New(),
		UserID:            userID,
		ShootRestrictions: []string{},
	}
	if err := s.modelRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ModelService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.ModelProfile, error) {
	p, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	cats, err := s.modelRepo.GetCategories(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Categories = cats
	return p, nil
}

func (s *ModelService) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*domain.ModelProfile, error) {
	p, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	cats, err := s.modelRepo.GetCategories(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Categories = cats
	return p, nil
}

func (s *ModelService) UpdateProfile(ctx context.Context, userID uuid.UUID, update func(*domain.ModelProfile)) (*domain.ModelProfile, error) {
	p, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			p = &domain.ModelProfile{
				ID:                uuid.New(),
				UserID:            userID,
				ShootRestrictions: []string{},
			}
			if err := s.modelRepo.Create(ctx, p); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	update(p)

	if err := s.modelRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	cats, err := s.modelRepo.GetCategories(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Categories = cats
	return p, nil
}

func (s *ModelService) AddCategory(ctx context.Context, userID uuid.UUID, categoryID int16) error {
	p, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.modelRepo.AddCategory(ctx, p.ID, categoryID)
}

func (s *ModelService) RemoveCategory(ctx context.Context, userID uuid.UUID, categoryID int16) error {
	p, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.modelRepo.RemoveCategory(ctx, p.ID, categoryID)
}

func (s *ModelService) UploadPhoto(ctx context.Context, userID uuid.UUID, fh *multipart.FileHeader) (*domain.ModelPhoto, error) {
	p, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	count, err := s.modelRepo.CountPhotos(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if count >= MaxPhotoCount {
		return nil, ErrPhotoLimitReached
	}

	if fh.Size > MaxPhotoSize {
		return nil, ErrFileTooLarge
	}

	mimeType := fh.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return nil, ErrInvalidMimeType
	}

	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	result, err := s.storage.UploadPhoto(ctx, p.ID, fh.Filename, f, fh.Size, mimeType)
	if err != nil {
		return nil, err
	}

	photo := &domain.ModelPhoto{
		ID:               uuid.New(),
		ModelProfileID:   p.ID,
		StorageKey:       result.StorageKey,
		OriginalName:     fh.Filename,
		SizeBytes:        result.SizeBytes,
		MimeType:         mimeType,
		ModerationStatus: domain.ModerationPending,
		IsCover:          count == 0, // first photo is cover by default
		SortOrder:        int16(count),
		CreatedAt:        time.Now(),
	}

	if err := s.modelRepo.AddPhoto(ctx, photo); err != nil {
		return nil, err
	}

	return photo, nil
}

func (s *ModelService) GetPhotos(ctx context.Context, profileID uuid.UUID, approvedOnly bool) ([]domain.ModelPhoto, error) {
	return s.modelRepo.GetPhotos(ctx, profileID, approvedOnly)
}

func (s *ModelService) UpdatePhoto(ctx context.Context, userID, photoID uuid.UUID, isCover *bool, sortOrder *int16, tagIDs []int16) (*domain.ModelPhoto, error) {
	profile, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	photo, err := s.modelRepo.GetPhoto(ctx, photoID)
	if err != nil {
		return nil, err
	}

	if photo.ModelProfileID != profile.ID {
		return nil, ErrNotOwner
	}

	if isCover != nil {
		photo.IsCover = *isCover
	}
	if sortOrder != nil {
		photo.SortOrder = *sortOrder
	}

	if err := s.modelRepo.UpdatePhoto(ctx, photo); err != nil {
		return nil, err
	}

	if tagIDs != nil {
		if err := s.modelRepo.SetPhotoTags(ctx, photoID, tagIDs); err != nil {
			return nil, err
		}
	}

	return photo, nil
}

func (s *ModelService) DeletePhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	profile, err := s.modelRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	photo, err := s.modelRepo.GetPhoto(ctx, photoID)
	if err != nil {
		return err
	}

	if photo.ModelProfileID != profile.ID {
		return ErrNotOwner
	}

	if err := s.storage.DeleteObject(ctx, photo.StorageKey); err != nil {
		fmt.Printf("WARNING: failed to delete photo from storage: %v\n", err)
	}

	return s.modelRepo.DeletePhoto(ctx, photoID)
}

func (s *ModelService) Search(ctx context.Context, f repository.ModelSearchFilter) ([]domain.ModelProfile, int, error) {
	profiles, total, err := s.modelRepo.Search(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	for i := range profiles {
		cats, err := s.modelRepo.GetCategories(ctx, profiles[i].ID)
		if err != nil {
			return nil, 0, err
		}
		profiles[i].Categories = cats
	}

	return profiles, total, nil
}

func (s *ModelService) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	return s.modelRepo.GetAllCategories(ctx)
}

func (s *ModelService) GetAllPhotoTags(ctx context.Context) ([]domain.PhotoTag, error) {
	return s.modelRepo.GetAllPhotoTags(ctx)
}

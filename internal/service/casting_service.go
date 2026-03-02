package service

import (
	"context"
	"errors"

	"SHIBA/internal/domain"
	"SHIBA/internal/repository"
	"SHIBA/pkg/mail"

	"github.com/google/uuid"
)

type CastingService struct {
	castingRepo *repository.CastingRepository
	agencyRepo  *repository.AgencyRepository
	modelRepo   *repository.ModelRepository
	userRepo    *repository.UserRepository
	mailSender  *mail.Sender
}

func NewCastingService(
	castingRepo *repository.CastingRepository,
	agencyRepo *repository.AgencyRepository,
	modelRepo *repository.ModelRepository,
	userRepo *repository.UserRepository,
	mailSender *mail.Sender,
) *CastingService {
	return &CastingService{
		castingRepo: castingRepo,
		agencyRepo:  agencyRepo,
		modelRepo:   modelRepo,
		userRepo:    userRepo,
		mailSender:  mailSender,
	}
}

func (s *CastingService) CreateCasting(ctx context.Context, agencyUserID uuid.UUID, c *domain.Casting) (*domain.Casting, error) {
	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return nil, err
	}
	if agency.Status != domain.AgencyStatusActive {
		return nil, errors.New("agency is not approved")
	}

	c.ID = uuid.New()
	c.AgencyProfileID = agency.ID
	c.Status = domain.CastingActive

	if err := s.castingRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CastingService) GetCasting(ctx context.Context, id uuid.UUID) (*domain.Casting, error) {
	return s.castingRepo.GetByID(ctx, id)
}

func (s *CastingService) UpdateCasting(ctx context.Context, castingID, agencyUserID uuid.UUID, update func(*domain.Casting)) (*domain.Casting, error) {
	casting, err := s.castingRepo.GetByID(ctx, castingID)
	if err != nil {
		return nil, err
	}

	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return nil, err
	}
	if casting.AgencyProfileID != agency.ID {
		return nil, ErrNotOwner
	}

	update(casting)
	if err := s.castingRepo.Update(ctx, casting); err != nil {
		return nil, err
	}
	return casting, nil
}

func (s *CastingService) DeleteCasting(ctx context.Context, castingID, agencyUserID uuid.UUID) error {
	casting, err := s.castingRepo.GetByID(ctx, castingID)
	if err != nil {
		return err
	}

	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return err
	}
	if casting.AgencyProfileID != agency.ID {
		return ErrNotOwner
	}

	return s.castingRepo.Delete(ctx, castingID)
}

func (s *CastingService) ListCastings(ctx context.Context, f repository.CastingFilter) ([]domain.Casting, int, error) {
	return s.castingRepo.List(ctx, f)
}

func (s *CastingService) GetApplicationsByCasting(ctx context.Context, castingID, agencyUserID uuid.UUID, limit, offset int) ([]domain.Application, int, error) {
	casting, err := s.castingRepo.GetByID(ctx, castingID)
	if err != nil {
		return nil, 0, err
	}
	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return nil, 0, err
	}
	if casting.AgencyProfileID != agency.ID {
		return nil, 0, ErrNotOwner
	}
	return s.castingRepo.GetApplicationsByCasting(ctx, castingID, limit, offset)
}

func (s *CastingService) CreateApplication(ctx context.Context, modelUserID, castingID uuid.UUID, message string) (*domain.Application, error) {
	modelProfile, err := s.modelRepo.GetByUserID(ctx, modelUserID)
	if err != nil {
		return nil, err
	}

	casting, err := s.castingRepo.GetByID(ctx, castingID)
	if err != nil {
		return nil, err
	}
	if casting.Status != domain.CastingActive {
		return nil, errors.New("casting is not active")
	}

	app := &domain.Application{
		ID:             uuid.New(),
		CastingID:      castingID,
		ModelProfileID: modelProfile.ID,
		Status:         domain.AppPending,
		ModelMessage:   message,
	}

	if err := s.castingRepo.CreateApplication(ctx, app); err != nil {
		return nil, err
	}

	// Notify agency
	go func() {
		agency, _ := s.agencyRepo.GetByID(context.Background(), casting.AgencyProfileID)
		if agency != nil {
			user, _ := s.userRepo.GetByID(context.Background(), agency.UserID)
			if user != nil {
				modelName := modelProfile.FirstName + " " + modelProfile.LastName
				_ = s.mailSender.SendNewApplication(user.Email, modelName, casting.Title)
			}
		}
	}()

	return app, nil
}

func (s *CastingService) GetApplicationsByModel(ctx context.Context, modelUserID uuid.UUID, limit, offset int) ([]domain.Application, int, error) {
	modelProfile, err := s.modelRepo.GetByUserID(ctx, modelUserID)
	if err != nil {
		return nil, 0, err
	}
	return s.castingRepo.GetApplicationsByModel(ctx, modelProfile.ID, limit, offset)
}

func (s *CastingService) DeleteApplication(ctx context.Context, appID, modelUserID uuid.UUID) error {
	modelProfile, err := s.modelRepo.GetByUserID(ctx, modelUserID)
	if err != nil {
		return err
	}

	app, err := s.castingRepo.GetApplication(ctx, appID)
	if err != nil {
		return err
	}
	if app.ModelProfileID != modelProfile.ID {
		return ErrNotOwner
	}
	if app.Status != domain.AppPending {
		return errors.New("can only delete pending applications")
	}

	return s.castingRepo.DeleteApplication(ctx, appID)
}

func (s *CastingService) UpdateApplicationStatus(ctx context.Context, castingID, appID, agencyUserID uuid.UUID, status domain.ApplicationStatus, response string) (*domain.Application, error) {
	casting, err := s.castingRepo.GetByID(ctx, castingID)
	if err != nil {
		return nil, err
	}
	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return nil, err
	}
	if casting.AgencyProfileID != agency.ID {
		return nil, ErrNotOwner
	}

	if err := s.castingRepo.UpdateApplicationStatus(ctx, appID, status, response); err != nil {
		return nil, err
	}

	app, err := s.castingRepo.GetApplication(ctx, appID)
	if err != nil {
		return nil, err
	}

	// Notify model
	go func() {
		model, _ := s.modelRepo.GetByID(context.Background(), app.ModelProfileID)
		if model != nil {
			user, _ := s.userRepo.GetByID(context.Background(), model.UserID)
			if user != nil {
				_ = s.mailSender.SendApplicationStatusChanged(user.Email, casting.Title, string(status))
			}
		}
	}()

	return app, nil
}

func (s *CastingService) CreateInvitation(ctx context.Context, agencyUserID, modelProfileID uuid.UUID, message string, castingID *uuid.UUID) (*domain.Invitation, error) {
	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return nil, err
	}
	if agency.Status != domain.AgencyStatusActive {
		return nil, errors.New("agency is not approved")
	}

	// Verify model exists
	model, err := s.modelRepo.GetByID(ctx, modelProfileID)
	if err != nil {
		return nil, err
	}

	inv := &domain.Invitation{
		ID:              uuid.New(),
		AgencyProfileID: agency.ID,
		ModelProfileID:  modelProfileID,
		CastingID:       castingID,
		Message:         message,
		Status:          domain.AppPending,
	}

	if err := s.castingRepo.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	// Notify model
	go func() {
		user, _ := s.userRepo.GetByID(context.Background(), model.UserID)
		if user != nil {
			_ = s.mailSender.SendNewInvitation(user.Email, agency.CompanyName, message)
		}
	}()

	return inv, nil
}

func (s *CastingService) GetInvitationsByModel(ctx context.Context, modelUserID uuid.UUID, limit, offset int) ([]domain.Invitation, int, error) {
	modelProfile, err := s.modelRepo.GetByUserID(ctx, modelUserID)
	if err != nil {
		return nil, 0, err
	}
	return s.castingRepo.GetInvitationsByModel(ctx, modelProfile.ID, limit, offset)
}

func (s *CastingService) GetInvitationsByAgency(ctx context.Context, agencyUserID uuid.UUID, limit, offset int) ([]domain.Invitation, int, error) {
	agency, err := s.agencyRepo.GetByUserID(ctx, agencyUserID)
	if err != nil {
		return nil, 0, err
	}
	return s.castingRepo.GetInvitationsByAgency(ctx, agency.ID, limit, offset)
}

func (s *CastingService) UpdateInvitationStatus(ctx context.Context, invID, modelUserID uuid.UUID, status domain.ApplicationStatus) error {
	modelProfile, err := s.modelRepo.GetByUserID(ctx, modelUserID)
	if err != nil {
		return err
	}

	inv, err := s.castingRepo.GetInvitation(ctx, invID)
	if err != nil {
		return err
	}
	if inv.ModelProfileID != modelProfile.ID {
		return ErrNotOwner
	}

	return s.castingRepo.UpdateInvitationStatus(ctx, invID, status)
}

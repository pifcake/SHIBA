package service

import (
	"context"

	"SHIBA/internal/domain"
	"SHIBA/internal/repository"
	"SHIBA/pkg/mail"

	"github.com/google/uuid"
)

type AdminService struct {
	userRepo      *repository.UserRepository
	modelRepo     *repository.ModelRepository
	agencyRepo    *repository.AgencyRepository
	complaintRepo *repository.ComplaintRepository
	mailSender    *mail.Sender
}

func NewAdminService(
	userRepo *repository.UserRepository,
	modelRepo *repository.ModelRepository,
	agencyRepo *repository.AgencyRepository,
	complaintRepo *repository.ComplaintRepository,
	mailSender *mail.Sender,
) *AdminService {
	return &AdminService{
		userRepo:      userRepo,
		modelRepo:     modelRepo,
		agencyRepo:    agencyRepo,
		complaintRepo: complaintRepo,
		mailSender:    mailSender,
	}
}

func (s *AdminService) ListUsers(ctx context.Context, role, status string, limit, offset int) ([]domain.User, int, error) {
	return s.userRepo.List(ctx, role, status, limit, offset)
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	return s.userRepo.UpdateStatus(ctx, userID, status)
}

func (s *AdminService) GetPendingAgencies(ctx context.Context, limit, offset int) ([]domain.AgencyProfile, int, error) {
	return s.agencyRepo.GetPending(ctx, limit, offset)
}

func (s *AdminService) ApproveAgency(ctx context.Context, agencyID, adminID uuid.UUID) error {
	return s.agencyRepo.Approve(ctx, agencyID, adminID)
}

func (s *AdminService) RejectAgency(ctx context.Context, agencyID uuid.UUID) error {
	return s.agencyRepo.Reject(ctx, agencyID)
}

func (s *AdminService) GetPendingPhotos(ctx context.Context, limit, offset int) ([]domain.ModelPhoto, int, error) {
	return s.modelRepo.GetPendingPhotos(ctx, limit, offset)
}

func (s *AdminService) ModeratePhoto(ctx context.Context, photoID, adminID uuid.UUID, status domain.ModerationStatus, reason string) error {
	photo, err := s.modelRepo.GetPhoto(ctx, photoID)
	if err != nil {
		return err
	}

	if err := s.modelRepo.ModeratePhoto(ctx, photoID, status, adminID, reason); err != nil {
		return err
	}

	if status == domain.ModerationRejected && reason != "" {
		// Notify model owner
		go func() {
			profile, _ := s.modelRepo.GetByID(context.Background(), photo.ModelProfileID)
			if profile != nil {
				user, _ := s.userRepo.GetByID(context.Background(), profile.UserID)
				if user != nil {
					_ = s.mailSender.SendPhotoRejected(user.Email, reason)
				}
			}
		}()
	}

	return nil
}

func (s *AdminService) ListComplaints(ctx context.Context, status string, limit, offset int) ([]domain.Complaint, int, error) {
	f := repository.ComplaintFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	return s.complaintRepo.List(ctx, f)
}

func (s *AdminService) UpdateComplaint(ctx context.Context, complaintID, adminID uuid.UUID, status domain.ComplaintStatus, note string) error {
	return s.complaintRepo.Update(ctx, complaintID, status, &adminID, note)
}

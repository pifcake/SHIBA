package service

import (
	"context"

	"SHIBA/internal/domain"
	"SHIBA/internal/repository"

	"github.com/google/uuid"
)

type ComplaintService struct {
	complaintRepo *repository.ComplaintRepository
}

func NewComplaintService(complaintRepo *repository.ComplaintRepository) *ComplaintService {
	return &ComplaintService{complaintRepo: complaintRepo}
}

func (s *ComplaintService) Create(ctx context.Context, reporterID uuid.UUID, targetType domain.ComplaintTargetType, targetID uuid.UUID, reason domain.ComplaintReason, description string) (*domain.Complaint, error) {
	c := &domain.Complaint{
		ID:             uuid.New(),
		ReporterID:     reporterID,
		TargetType:     targetType,
		TargetID:       targetID,
		ReasonCategory: reason,
		Description:    description,
		Status:         domain.ComplaintOpen,
	}
	if err := s.complaintRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

type ComplaintTargetType string

const (
	TargetCasting    ComplaintTargetType = "casting"
	TargetInvitation ComplaintTargetType = "invitation"
	TargetAgency     ComplaintTargetType = "agency"
	TargetModel      ComplaintTargetType = "model"
)

type ComplaintReason string

const (
	ReasonInappropriate ComplaintReason = "inappropriate_content"
	ReasonHarassment    ComplaintReason = "harassment"
	ReasonFraud         ComplaintReason = "fraud"
	ReasonOther         ComplaintReason = "other"
)

type ComplaintStatus string

const (
	ComplaintOpen      ComplaintStatus = "open"
	ComplaintReviewed  ComplaintStatus = "reviewed"
	ComplaintResolved  ComplaintStatus = "resolved"
	ComplaintDismissed ComplaintStatus = "dismissed"
)

type Complaint struct {
	ID             uuid.UUID
	ReporterID     uuid.UUID
	TargetType     ComplaintTargetType
	TargetID       uuid.UUID
	ReasonCategory ComplaintReason
	Description    string
	Status         ComplaintStatus
	ReviewedBy     *uuid.UUID
	ResolutionNote string
	ResolvedAt     *time.Time
	CreatedAt      time.Time
}

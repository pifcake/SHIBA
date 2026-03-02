package domain

import (
	"time"

	"github.com/google/uuid"
)

type CastingStatus string

const (
	CastingActive    CastingStatus = "active"
	CastingClosed    CastingStatus = "closed"
	CastingCancelled CastingStatus = "cancelled"
)

type Casting struct {
	ID              uuid.UUID
	AgencyProfileID uuid.UUID
	Title           string
	Description     string
	City            string
	CastingDate     *time.Time
	CategoryID      *int16
	Status          CastingStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// Populated joins
	AgencyName string
	Category   *Category
}

type ApplicationStatus string

const (
	AppPending  ApplicationStatus = "pending"
	AppViewed   ApplicationStatus = "viewed"
	AppAccepted ApplicationStatus = "accepted"
	AppRejected ApplicationStatus = "rejected"
)

type Application struct {
	ID             uuid.UUID
	CastingID      uuid.UUID
	ModelProfileID uuid.UUID
	Status         ApplicationStatus
	ModelMessage   string
	AgencyResponse string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	// Populated joins
	Casting      *Casting
	ModelProfile *ModelProfile
}

type Invitation struct {
	ID              uuid.UUID
	AgencyProfileID uuid.UUID
	ModelProfileID  uuid.UUID
	CastingID       *uuid.UUID
	Message         string
	Status          ApplicationStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// Populated joins
	AgencyProfile *AgencyProfile
	ModelProfile  *ModelProfile
	Casting       *Casting
}

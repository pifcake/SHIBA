package domain

import (
	"time"

	"github.com/google/uuid"
)

// Reference data

type Category struct {
	ID   int16  `json:"id"`
	Name string `json:"name"`
}

type PhotoTag struct {
	ID   int16  `json:"id"`
	Name string `json:"name"`
}

// Model profile

type ModelProfile struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	FirstName         string
	LastName          string
	BirthDate         *time.Time
	City              string
	Country           string
	WillingToRelocate bool
	HeightCm          *int16
	WeightKg          *int16
	ChestCm           *int16
	WaistCm           *int16
	HipsCm            *int16
	ShoeSize          *float32
	ClothingSize      string
	Gender            string
	HairColor         string
	HairLength        string
	HairStructure     string
	EyeColor          string
	ClothingSizeTop   string
	ClothingSizeBot   string
	Phone             string
	Bio               string
	ShootRestrictions    []string
	Categories           []Category
	CoverPhotoStorageKey string // populated only by Search
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Agency profile

type AgencyStatus string

const (
	AgencyStatusPendingApproval AgencyStatus = "pending_approval"
	AgencyStatusActive          AgencyStatus = "active"
	AgencyStatusBlocked         AgencyStatus = "blocked"
)

type AgencyProfile struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CompanyName string
	Description string
	Phone       string
	LogoURL     string
	Status      AgencyStatus
	ApprovedBy  *uuid.UUID
	ApprovedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Photos

type ModerationStatus string

const (
	ModerationPending  ModerationStatus = "pending"
	ModerationApproved ModerationStatus = "approved"
	ModerationRejected ModerationStatus = "rejected"
)

type ModelPhoto struct {
	ID               uuid.UUID
	ModelProfileID   uuid.UUID
	StorageKey       string
	OriginalName     string
	SizeBytes        int64
	MimeType         string
	WidthPx          *int32
	HeightPx         *int32
	ModerationStatus ModerationStatus
	ModeratedBy      *uuid.UUID
	ModeratedAt      *time.Time
	RejectionReason  string
	IsCover          bool
	SortOrder        int16
	Tags             []PhotoTag
	CreatedAt        time.Time
}

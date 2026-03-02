package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleAgency UserRole = "agency"
	RoleModel  UserRole = "model"
)

type UserStatus string

const (
	StatusPendingEmail UserStatus = "pending_email"
	StatusActive       UserStatus = "active"
	StatusBlocked      UserStatus = "blocked"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         UserRole
	Status       UserStatus
	PhotoRef     string // populated only by List: storage_key for model, logo_url for agency
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EmailVerification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

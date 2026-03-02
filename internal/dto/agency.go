package dto

import "time"

type UpdateAgencyProfileRequest struct {
	CompanyName string `json:"company_name"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
}

type AgencyProfileResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CompanyName string    `json:"company_name"`
	Description string    `json:"description"`
	Phone       string    `json:"phone"`
	LogoURL     string    `json:"logo_url"`
	Status      string    `json:"status"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

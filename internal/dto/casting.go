package dto

import "time"

type CreateCastingRequest struct {
	Title       string  `json:"title" binding:"required,min=3,max=255"`
	Description string  `json:"description"`
	City        string  `json:"city"`
	CastingDate *string `json:"casting_date"` // "YYYY-MM-DD"
	CategoryID  *int16  `json:"category_id"`
}

type UpdateCastingRequest struct {
	Title       string  `json:"title" binding:"required,min=3,max=255"`
	Description string  `json:"description"`
	City        string  `json:"city"`
	CastingDate *string `json:"casting_date"`
	CategoryID  *int16  `json:"category_id"`
	Status      string  `json:"status" binding:"omitempty,oneof=active closed cancelled"`
}

type CastingResponse struct {
	ID          string     `json:"id"`
	AgencyID    string     `json:"agency_id"`
	AgencyName  string     `json:"agency_name,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	City        string     `json:"city"`
	CastingDate *string    `json:"casting_date,omitempty"`
	CategoryID  *int16     `json:"category_id,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CastingFilter struct {
	Status     string `form:"status"`
	City       string `form:"city"`
	CategoryID *int16 `form:"category_id"`
	Limit      int    `form:"limit,default=20"`
	Offset     int    `form:"offset,default=0"`
}

type CreateApplicationRequest struct {
	CastingID string `json:"casting_id" binding:"required,uuid"`
	Message   string `json:"message"`
}

type UpdateApplicationStatusRequest struct {
	Status   string `json:"status" binding:"required,oneof=viewed accepted rejected"`
	Response string `json:"response"`
}

type ApplicationResponse struct {
	ID             string    `json:"id"`
	CastingID      string    `json:"casting_id"`
	ModelProfileID string    `json:"model_profile_id"`
	Status         string    `json:"status"`
	ModelMessage   string    `json:"model_message"`
	AgencyResponse string    `json:"agency_response,omitempty"`
	Casting        *CastingResponse      `json:"casting,omitempty"`
	Model          *ModelProfileResponse `json:"model,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateInvitationRequest struct {
	ModelID   string  `json:"model_id" binding:"required,uuid"`
	Message   string  `json:"message"`
	CastingID *string `json:"casting_id"`
}

type UpdateInvitationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected"`
}

type InvitationResponse struct {
	ID        string    `json:"id"`
	AgencyID  string    `json:"agency_id"`
	ModelID   string    `json:"model_id"`
	CastingID *string   `json:"casting_id,omitempty"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Agency    *AgencyProfileResponse `json:"agency,omitempty"`
	Casting   *CastingResponse       `json:"casting,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateComplaintRequest struct {
	TargetType     string `json:"target_type" binding:"required,oneof=casting invitation agency model"`
	TargetID       string `json:"target_id" binding:"required,uuid"`
	ReasonCategory string `json:"reason_category" binding:"required,oneof=inappropriate_content harassment fraud other"`
	Description    string `json:"description"`
}

type UpdateComplaintRequest struct {
	Status         string `json:"status" binding:"required,oneof=reviewed resolved dismissed"`
	ResolutionNote string `json:"resolution_note"`
}

type ComplaintResponse struct {
	ID             string     `json:"id"`
	ReporterID     string     `json:"reporter_id"`
	TargetType     string     `json:"target_type"`
	TargetID       string     `json:"target_id"`
	ReasonCategory string     `json:"reason_category"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	ResolutionNote string     `json:"resolution_note,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

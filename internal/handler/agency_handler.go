package handler

import (
	"errors"

	"SHIBA/internal/domain"
	"SHIBA/internal/dto"
	"SHIBA/internal/middleware"
	"SHIBA/internal/repository"
	"SHIBA/internal/service"
	"SHIBA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgencyHandler struct {
	agencyService *service.AgencyService
}

func NewAgencyHandler(agencyService *service.AgencyService) *AgencyHandler {
	return &AgencyHandler{agencyService: agencyService}
}

func (h *AgencyHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	profile, err := h.agencyService.GetOrCreateProfile(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, agencyToResponse(profile))
}

func (h *AgencyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid agency id")
		return
	}
	profile, err := h.agencyService.GetProfileByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "agency not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, agencyToResponse(profile))
}

func (h *AgencyHandler) UpdateMe(c *gin.Context) {
	var req dto.UpdateAgencyProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	profile, err := h.agencyService.UpdateProfile(c.Request.Context(), userID, req.CompanyName, req.Description, req.Phone)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, agencyToResponse(profile))
}

func (h *AgencyHandler) UploadLogo(c *gin.Context) {
	fh, err := c.FormFile("logo")
	if err != nil {
		response.BadRequest(c, "logo file is required")
		return
	}
	userID := middleware.GetUserID(c)
	logoURL, err := h.agencyService.UploadLogo(c.Request.Context(), userID, fh)
	if err != nil {
		if errors.Is(err, service.ErrInvalidMimeType) {
			response.UnprocessableEntity(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"logo_url": logoURL})
}

func agencyToResponse(p *domain.AgencyProfile) dto.AgencyProfileResponse {
	return dto.AgencyProfileResponse{
		ID:          p.ID.String(),
		UserID:      p.UserID.String(),
		CompanyName: p.CompanyName,
		Description: p.Description,
		Phone:       p.Phone,
		LogoURL:     p.LogoURL,
		Status:      string(p.Status),
		ApprovedAt:  p.ApprovedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

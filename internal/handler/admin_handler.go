package handler

import (
	"SHIBA/internal/domain"
	"SHIBA/internal/dto"
	"SHIBA/internal/middleware"
	"SHIBA/internal/repository"
	"SHIBA/internal/service"
	"SHIBA/pkg/response"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService *service.AdminService
	storageURL   string
}

func NewAdminHandler(adminService *service.AdminService, storageURL string) *AdminHandler {
	return &AdminHandler{adminService: adminService, storageURL: storageURL}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	var f dto.AdminUsersFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	users, total, err := h.adminService.ListUsers(c.Request.Context(), f.Role, f.Status, f.Limit, f.Offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.AdminUserResponse, 0, len(users))
	for _, u := range users {
		var photoURL string
		if u.PhotoRef != "" {
			if u.Role == domain.RoleModel {
				photoURL = h.storageURL + "/" + u.PhotoRef
			} else {
				photoURL = u.PhotoRef // agency logo_url is already a full URL
			}
		}
		items = append(items, dto.AdminUserResponse{
			ID:        u.ID.String(),
			Email:     u.Email,
			Role:      string(u.Role),
			Status:    string(u.Status),
			PhotoURL:  photoURL,
			CreatedAt: u.CreatedAt.String(),
		})
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: f.Limit, Offset: f.Offset})
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.adminService.UpdateUserStatus(c.Request.Context(), id, domain.UserStatus(req.Status)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "user status updated")
}

func (h *AdminHandler) GetPendingAgencies(c *gin.Context) {
	limit, offset := parsePagination(c)
	agencies, total, err := h.adminService.GetPendingAgencies(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.AgencyProfileResponse, 0, len(agencies))
	for _, a := range agencies {
		aa := a
		items = append(items, agencyToResponse(&aa))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *AdminHandler) ApproveAgency(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid agency id")
		return
	}
	adminID := middleware.GetUserID(c)
	if err := h.adminService.ApproveAgency(c.Request.Context(), id, adminID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "agency not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "agency approved")
}

func (h *AdminHandler) RejectAgency(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid agency id")
		return
	}
	if err := h.adminService.RejectAgency(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "agency not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "agency rejected")
}

func (h *AdminHandler) GetPendingPhotos(c *gin.Context) {
	limit, offset := parsePagination(c)
	photos, total, err := h.adminService.GetPendingPhotos(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.PhotoResponse, 0, len(photos))
	for _, p := range photos {
		pp := p
		items = append(items, toPhotoResponse(&pp, h.storageURL))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *AdminHandler) ModeratePhoto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid photo id")
		return
	}

	var req dto.ModeratePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	adminID := middleware.GetUserID(c)
	if err := h.adminService.ModeratePhoto(c.Request.Context(), id, adminID, domain.ModerationStatus(req.Status), req.RejectionReason); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "photo not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "photo moderated")
}

func (h *AdminHandler) ListComplaints(c *gin.Context) {
	limit, offset := parsePagination(c)
	statusFilter := c.Query("status")

	complaints, total, err := h.adminService.ListComplaints(c.Request.Context(), statusFilter, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.ComplaintResponse, 0, len(complaints))
	for _, comp := range complaints {
		cc := comp
		items = append(items, complaintToResponse(&cc))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *AdminHandler) UpdateComplaint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid complaint id")
		return
	}

	var req dto.UpdateComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	adminID := middleware.GetUserID(c)
	if err := h.adminService.UpdateComplaint(c.Request.Context(), id, adminID, domain.ComplaintStatus(req.Status), req.ResolutionNote); err != nil {
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "complaint updated")
}

func complaintToResponse(c *domain.Complaint) dto.ComplaintResponse {
	return dto.ComplaintResponse{
		ID:             c.ID.String(),
		ReporterID:     c.ReporterID.String(),
		TargetType:     string(c.TargetType),
		TargetID:       c.TargetID.String(),
		ReasonCategory: string(c.ReasonCategory),
		Description:    c.Description,
		Status:         string(c.Status),
		ResolutionNote: c.ResolutionNote,
		ResolvedAt:     c.ResolvedAt,
		CreatedAt:      c.CreatedAt,
	}
}

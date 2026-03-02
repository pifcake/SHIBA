package handler

import (
	"errors"
	"time"

	"SHIBA/internal/domain"
	"SHIBA/internal/dto"
	"SHIBA/internal/middleware"
	"SHIBA/internal/repository"
	"SHIBA/internal/service"
	"SHIBA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CastingHandler struct {
	castingService *service.CastingService
}

func NewCastingHandler(castingService *service.CastingService) *CastingHandler {
	return &CastingHandler{castingService: castingService}
}

func (h *CastingHandler) CreateCasting(c *gin.Context) {
	var req dto.CreateCastingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	casting := &domain.Casting{
		Title:       req.Title,
		Description: req.Description,
		City:        req.City,
		CategoryID:  req.CategoryID,
	}
	if req.CastingDate != nil {
		t, err := time.Parse("2006-01-02", *req.CastingDate)
		if err == nil {
			casting.CastingDate = &t
		}
	}

	userID := middleware.GetUserID(c)
	result, err := h.castingService.CreateCasting(c.Request.Context(), userID, casting)
	if err != nil {
		if err.Error() == "agency is not approved" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, castingToResponse(result))
}

func (h *CastingHandler) GetCasting(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid casting id")
		return
	}

	casting, err := h.castingService.GetCasting(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "casting not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, castingToResponse(casting))
}

func (h *CastingHandler) UpdateCasting(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid casting id")
		return
	}

	var req dto.UpdateCastingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	casting, err := h.castingService.UpdateCasting(c.Request.Context(), id, userID, func(c *domain.Casting) {
		c.Title = req.Title
		c.Description = req.Description
		c.City = req.City
		c.CategoryID = req.CategoryID
		if req.Status != "" {
			c.Status = domain.CastingStatus(req.Status)
		}
		if req.CastingDate != nil {
			t, err := time.Parse("2006-01-02", *req.CastingDate)
			if err == nil {
				c.CastingDate = &t
			}
		}
	})
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "casting not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, castingToResponse(casting))
}

func (h *CastingHandler) DeleteCasting(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid casting id")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.castingService.DeleteCasting(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *CastingHandler) ListCastings(c *gin.Context) {
	var f dto.CastingFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	repoFilter := repository.CastingFilter{
		Status:     f.Status,
		City:       f.City,
		CategoryID: f.CategoryID,
		Limit:      f.Limit,
		Offset:     f.Offset,
	}

	castings, total, err := h.castingService.ListCastings(c.Request.Context(), repoFilter)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.CastingResponse, 0, len(castings))
	for _, c := range castings {
		cc := c
		items = append(items, castingToResponse(&cc))
	}

	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: f.Limit, Offset: f.Offset})
}

func (h *CastingHandler) GetApplicationsByCasting(c *gin.Context) {
	castingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid casting id")
		return
	}

	limit, offset := parsePagination(c)
	userID := middleware.GetUserID(c)

	apps, total, err := h.castingService.GetApplicationsByCasting(c.Request.Context(), castingID, userID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}

	items := make([]dto.ApplicationResponse, 0, len(apps))
	for _, a := range apps {
		aa := a
		items = append(items, applicationToResponse(&aa))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *CastingHandler) UpdateApplicationStatus(c *gin.Context) {
	castingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid casting id")
		return
	}
	appID, err := uuid.Parse(c.Param("app_id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	var req dto.UpdateApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	app, err := h.castingService.UpdateApplicationStatus(
		c.Request.Context(), castingID, appID, userID,
		domain.ApplicationStatus(req.Status), req.Response,
	)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, applicationToResponse(app))
}

// Applications

type ApplicationHandler struct {
	castingService *service.CastingService
}

func NewApplicationHandler(castingService *service.CastingService) *ApplicationHandler {
	return &ApplicationHandler{castingService: castingService}
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	castingID, _ := uuid.Parse(req.CastingID)
	userID := middleware.GetUserID(c)

	app, err := h.castingService.CreateApplication(c.Request.Context(), userID, castingID, req.Message)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			response.Conflict(c, "already applied to this casting")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, applicationToResponse(app))
}

func (h *ApplicationHandler) GetMyApplications(c *gin.Context) {
	limit, offset := parsePagination(c)
	userID := middleware.GetUserID(c)

	apps, total, err := h.castingService.GetApplicationsByModel(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.ApplicationResponse, 0, len(apps))
	for _, a := range apps {
		aa := a
		items = append(items, applicationToResponse(&aa))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.castingService.DeleteApplication(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.NoContent(c)
}

// Invitations

type InvitationHandler struct {
	castingService *service.CastingService
}

func NewInvitationHandler(castingService *service.CastingService) *InvitationHandler {
	return &InvitationHandler{castingService: castingService}
}

func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	modelID, _ := uuid.Parse(req.ModelID)
	userID := middleware.GetUserID(c)

	var castingID *uuid.UUID
	if req.CastingID != nil {
		id, err := uuid.Parse(*req.CastingID)
		if err == nil {
			castingID = &id
		}
	}

	inv, err := h.castingService.CreateInvitation(c.Request.Context(), userID, modelID, req.Message, castingID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, invitationToResponse(inv))
}

func (h *InvitationHandler) GetMyInvitations(c *gin.Context) {
	limit, offset := parsePagination(c)
	userID := middleware.GetUserID(c)

	invs, total, err := h.castingService.GetInvitationsByModel(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.InvitationResponse, 0, len(invs))
	for _, inv := range invs {
		ii := inv
		items = append(items, invitationToResponse(&ii))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *InvitationHandler) GetMyOutgoingInvitations(c *gin.Context) {
	limit, offset := parsePagination(c)
	userID := middleware.GetUserID(c)

	invs, total, err := h.castingService.GetInvitationsByAgency(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.InvitationResponse, 0, len(invs))
	for _, inv := range invs {
		ii := inv
		items = append(items, invitationToResponse(&ii))
	}
	response.OK(c, dto.PaginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (h *InvitationHandler) UpdateInvitationStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid invitation id")
		return
	}

	var req dto.UpdateInvitationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.castingService.UpdateInvitationStatus(c.Request.Context(), id, userID, domain.ApplicationStatus(req.Status)); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "invitation status updated")
}

// Helpers

func parsePagination(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0
	if l := c.Query("limit"); l != "" {
		if v, err := parseInt(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := parseInt(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}

func parseInt(s string) (int, error) {
	var v int
	_, err := (&v), (func() error {
		for _, ch := range s {
			if ch < '0' || ch > '9' {
				return errors.New("invalid")
			}
			v = v*10 + int(ch-'0')
		}
		return nil
	})()
	return v, err
}

func castingToResponse(c *domain.Casting) dto.CastingResponse {
	r := dto.CastingResponse{
		ID:          c.ID.String(),
		AgencyID:    c.AgencyProfileID.String(),
		AgencyName:  c.AgencyName,
		Title:       c.Title,
		Description: c.Description,
		City:        c.City,
		CategoryID:  c.CategoryID,
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.CastingDate != nil {
		s := c.CastingDate.Format("2006-01-02")
		r.CastingDate = &s
	}
	return r
}

func applicationToResponse(a *domain.Application) dto.ApplicationResponse {
	return dto.ApplicationResponse{
		ID:             a.ID.String(),
		CastingID:      a.CastingID.String(),
		ModelProfileID: a.ModelProfileID.String(),
		Status:         string(a.Status),
		ModelMessage:   a.ModelMessage,
		AgencyResponse: a.AgencyResponse,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func invitationToResponse(inv *domain.Invitation) dto.InvitationResponse {
	r := dto.InvitationResponse{
		ID:        inv.ID.String(),
		AgencyID:  inv.AgencyProfileID.String(),
		ModelID:   inv.ModelProfileID.String(),
		Message:   inv.Message,
		Status:    string(inv.Status),
		CreatedAt: inv.CreatedAt,
		UpdatedAt: inv.UpdatedAt,
	}
	if inv.CastingID != nil {
		s := inv.CastingID.String()
		r.CastingID = &s
	}
	return r
}

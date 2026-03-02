package handler

import (
	"SHIBA/internal/domain"
	"SHIBA/internal/dto"
	"SHIBA/internal/middleware"
	"SHIBA/internal/service"
	"SHIBA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ComplaintHandler struct {
	complaintService *service.ComplaintService
}

func NewComplaintHandler(complaintService *service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{complaintService: complaintService}
}

func (h *ComplaintHandler) CreateComplaint(c *gin.Context) {
	var req dto.CreateComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		response.BadRequest(c, "invalid target_id")
		return
	}

	userID := middleware.GetUserID(c)
	complaint, err := h.complaintService.Create(
		c.Request.Context(),
		userID,
		domain.ComplaintTargetType(req.TargetType),
		targetID,
		domain.ComplaintReason(req.ReasonCategory),
		req.Description,
	)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, complaintToResponse(complaint))
}

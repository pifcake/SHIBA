package handler

import (
	"SHIBA/internal/service"
	"SHIBA/pkg/response"

	"github.com/gin-gonic/gin"
)

type RefHandler struct {
	modelService *service.ModelService
}

func NewRefHandler(modelService *service.ModelService) *RefHandler {
	return &RefHandler{modelService: modelService}
}

func (h *RefHandler) GetCategories(c *gin.Context) {
	cats, err := h.modelService.GetAllCategories(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, cats)
}

func (h *RefHandler) GetPhotoTags(c *gin.Context) {
	tags, err := h.modelService.GetAllPhotoTags(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, tags)
}

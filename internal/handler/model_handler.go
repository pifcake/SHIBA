package handler

import (
	"errors"
	"strconv"
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

type ModelHandler struct {
	modelService *service.ModelService
	storageURL   string
}

func NewModelHandler(modelService *service.ModelService, storageURL string) *ModelHandler {
	return &ModelHandler{modelService: modelService, storageURL: storageURL}
}

func (h *ModelHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	profile, err := h.modelService.GetOrCreateProfile(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, toModelProfileResponse(profile, h.storageURL))
}

func (h *ModelHandler) UpdateMe(c *gin.Context) {
	var req dto.UpdateModelProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	profile, err := h.modelService.UpdateProfile(c.Request.Context(), userID, func(p *domain.ModelProfile) {
		p.FirstName = req.FirstName
		p.LastName = req.LastName
		p.City = req.City
		p.Country = req.Country
		p.WillingToRelocate = req.WillingToRelocate
		p.HeightCm = req.HeightCm
		p.WeightKg = req.WeightKg
		p.ChestCm = req.ChestCm
		p.WaistCm = req.WaistCm
		p.HipsCm = req.HipsCm
		p.ShoeSize = req.ShoeSize
		p.ClothingSize = req.ClothingSize
		p.Gender = req.Gender
		p.HairColor = req.HairColor
		p.HairLength = req.HairLength
		p.HairStructure = req.HairStructure
		p.EyeColor = req.EyeColor
		p.ClothingSizeTop = req.ClothingSizeTop
		p.ClothingSizeBot = req.ClothingSizeBot
		p.Phone = req.Phone
		p.Bio = req.Bio
		if req.ShootRestrictions != nil {
			p.ShootRestrictions = req.ShootRestrictions
		}
		if req.BirthDate != nil {
			t, err := time.Parse("2006-01-02", *req.BirthDate)
			if err == nil {
				p.BirthDate = &t
			}
		}
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, toModelProfileResponse(profile, h.storageURL))
}

func (h *ModelHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid profile id")
		return
	}

	profile, err := h.modelService.GetProfile(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "model not found")
			return
		}
		response.InternalError(c)
		return
	}

	// For agencies: only show basic info + approved photos handled separately
	response.OK(c, toModelProfileResponse(profile, h.storageURL))
}

func (h *ModelHandler) Search(c *gin.Context) {
	var f dto.ModelSearchFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	repoFilter := repository.ModelSearchFilter{
		City:              f.City,
		MinAge:            f.MinAge,
		MaxAge:            f.MaxAge,
		CategoryID:        f.CategoryID,
		WillingToRelocate: f.WillingToRelocate,
		Limit:             f.Limit,
		Offset:            f.Offset,
	}

	profiles, total, err := h.modelService.Search(c.Request.Context(), repoFilter)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.ModelProfileResponse, 0, len(profiles))
	for _, p := range profiles {
		pp := p
		items = append(items, toModelProfileResponse(&pp, h.storageURL))
	}

	response.OK(c, dto.PaginatedResponse{
		Data:   items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
}

func (h *ModelHandler) AddCategory(c *gin.Context) {
	var req dto.AddCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.modelService.AddCategory(c.Request.Context(), userID, req.CategoryID); err != nil {
		response.InternalError(c)
		return
	}
	response.Message(c, 200, "category added")
}

func (h *ModelHandler) RemoveCategory(c *gin.Context) {
	catID, err := strconv.ParseInt(c.Param("id"), 10, 16)
	if err != nil {
		response.BadRequest(c, "invalid category id")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.modelService.RemoveCategory(c.Request.Context(), userID, int16(catID)); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *ModelHandler) UploadPhoto(c *gin.Context) {
	fh, err := c.FormFile("photo")
	if err != nil {
		response.BadRequest(c, "photo file is required")
		return
	}

	userID := middleware.GetUserID(c)
	photo, err := h.modelService.UploadPhoto(c.Request.Context(), userID, fh)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhotoLimitReached):
			response.BadRequest(c, "maximum photo limit reached")
		case errors.Is(err, service.ErrInvalidMimeType):
			response.UnprocessableEntity(c, err.Error())
		case errors.Is(err, service.ErrFileTooLarge):
			response.UnprocessableEntity(c, err.Error())
		default:
			response.InternalError(c)
		}
		return
	}

	response.Created(c, toPhotoResponse(photo, h.storageURL))
}

func (h *ModelHandler) GetPhotos(c *gin.Context) {
	profileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid profile id")
		return
	}

	role := middleware.GetUserRole(c)
	approvedOnly := role == domain.RoleAgency

	photos, err := h.modelService.GetPhotos(c.Request.Context(), profileID, approvedOnly)
	if err != nil {
		response.InternalError(c)
		return
	}

	items := make([]dto.PhotoResponse, 0, len(photos))
	for _, p := range photos {
		pp := p
		items = append(items, toPhotoResponse(&pp, h.storageURL))
	}
	response.OK(c, items)
}

func (h *ModelHandler) UpdatePhoto(c *gin.Context) {
	photoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid photo id")
		return
	}

	var req dto.UpdatePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	photo, err := h.modelService.UpdatePhoto(c.Request.Context(), userID, photoID, req.IsCover, req.SortOrder, req.TagIDs)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, toPhotoResponse(photo, h.storageURL))
}

func (h *ModelHandler) DeletePhoto(c *gin.Context) {
	photoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid photo id")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.modelService.DeletePhoto(c.Request.Context(), userID, photoID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			response.Forbidden(c)
			return
		}
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

// Converters

func toModelProfileResponse(p *domain.ModelProfile, storageURL string) dto.ModelProfileResponse {
	r := dto.ModelProfileResponse{
		ID:                p.ID.String(),
		UserID:            p.UserID.String(),
		FirstName:         p.FirstName,
		LastName:          p.LastName,
		City:              p.City,
		Country:           p.Country,
		WillingToRelocate: p.WillingToRelocate,
		HeightCm:          p.HeightCm,
		WeightKg:          p.WeightKg,
		ChestCm:           p.ChestCm,
		WaistCm:           p.WaistCm,
		HipsCm:            p.HipsCm,
		ShoeSize:          p.ShoeSize,
		ClothingSize:      p.ClothingSize,
		Gender:            p.Gender,
		HairColor:         p.HairColor,
		HairLength:        p.HairLength,
		HairStructure:     p.HairStructure,
		EyeColor:          p.EyeColor,
		ClothingSizeTop:   p.ClothingSizeTop,
		ClothingSizeBot:   p.ClothingSizeBot,
		Phone:             p.Phone,
		Bio:               p.Bio,
		ShootRestrictions: p.ShootRestrictions,
		Categories:        make([]dto.CategoryResponse, 0, len(p.Categories)),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	if p.BirthDate != nil {
		s := p.BirthDate.Format("2006-01-02")
		r.BirthDate = &s
	}
	for _, cat := range p.Categories {
		r.Categories = append(r.Categories, dto.CategoryResponse{ID: cat.ID, Name: cat.Name})
	}
	if p.CoverPhotoStorageKey != "" {
		r.CoverPhotoURL = storageURL + "/" + p.CoverPhotoStorageKey
	}
	return r
}

func toPhotoResponse(p *domain.ModelPhoto, storageURL string) dto.PhotoResponse {
	r := dto.PhotoResponse{
		ID:               p.ID.String(),
		StorageKey:       p.StorageKey,
		URL:              storageURL + "/" + p.StorageKey,
		OriginalName:     p.OriginalName,
		SizeBytes:        p.SizeBytes,
		MimeType:         p.MimeType,
		WidthPx:          p.WidthPx,
		HeightPx:         p.HeightPx,
		ModerationStatus: string(p.ModerationStatus),
		RejectionReason:  p.RejectionReason,
		IsCover:          p.IsCover,
		SortOrder:        p.SortOrder,
		Tags:             make([]dto.PhotoTagResponse, 0, len(p.Tags)),
		CreatedAt:        p.CreatedAt,
	}
	for _, tag := range p.Tags {
		r.Tags = append(r.Tags, dto.PhotoTagResponse{ID: tag.ID, Name: tag.Name})
	}
	return r
}

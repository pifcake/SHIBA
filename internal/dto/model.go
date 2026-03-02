package dto

import "time"

type UpdateModelProfileRequest struct {
	FirstName         string   `json:"first_name"`
	LastName          string   `json:"last_name"`
	BirthDate         *string  `json:"birth_date"` // "YYYY-MM-DD"
	City              string   `json:"city"`
	Country           string   `json:"country"`
	WillingToRelocate bool     `json:"willing_to_relocate"`
	HeightCm          *int16   `json:"height_cm"`
	WeightKg          *int16   `json:"weight_kg"`
	ChestCm           *int16   `json:"chest_cm"`
	WaistCm           *int16   `json:"waist_cm"`
	HipsCm            *int16   `json:"hips_cm"`
	ShoeSize          *float32 `json:"shoe_size"`
	ClothingSize      string   `json:"clothing_size"`
	Gender            string   `json:"gender"`
	HairColor         string   `json:"hair_color"`
	HairLength        string   `json:"hair_length"`
	HairStructure     string   `json:"hair_structure"`
	EyeColor          string   `json:"eye_color"`
	ClothingSizeTop   string   `json:"clothing_size_top"`
	ClothingSizeBot   string   `json:"clothing_size_bot"`
	Phone             string   `json:"phone"`
	Bio               string   `json:"bio"`
	ShootRestrictions []string `json:"shoot_restrictions"`
}

type ModelProfileResponse struct {
	ID                string             `json:"id"`
	UserID            string             `json:"user_id"`
	FirstName         string             `json:"first_name"`
	LastName          string             `json:"last_name"`
	BirthDate         *string            `json:"birth_date,omitempty"`
	City              string             `json:"city"`
	Country           string             `json:"country"`
	WillingToRelocate bool               `json:"willing_to_relocate"`
	HeightCm          *int16             `json:"height_cm,omitempty"`
	WeightKg          *int16             `json:"weight_kg,omitempty"`
	ChestCm           *int16             `json:"chest_cm,omitempty"`
	WaistCm           *int16             `json:"waist_cm,omitempty"`
	HipsCm            *int16             `json:"hips_cm,omitempty"`
	ShoeSize          *float32           `json:"shoe_size,omitempty"`
	ClothingSize      string             `json:"clothing_size"`
	Gender            string             `json:"gender"`
	HairColor         string             `json:"hair_color"`
	HairLength        string             `json:"hair_length"`
	HairStructure     string             `json:"hair_structure"`
	EyeColor          string             `json:"eye_color"`
	ClothingSizeTop   string             `json:"clothing_size_top"`
	ClothingSizeBot   string             `json:"clothing_size_bot"`
	Phone             string             `json:"phone"`
	Bio               string             `json:"bio"`
	ShootRestrictions []string           `json:"shoot_restrictions"`
	Categories        []CategoryResponse `json:"categories"`
	CoverPhotoURL     string             `json:"cover_photo_url,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type CategoryResponse struct {
	ID   int16  `json:"id"`
	Name string `json:"name"`
}

type PhotoResponse struct {
	ID               string             `json:"id"`
	StorageKey       string             `json:"storage_key"`
	URL              string             `json:"url"`
	OriginalName     string             `json:"original_name"`
	SizeBytes        int64              `json:"size_bytes"`
	MimeType         string             `json:"mime_type"`
	WidthPx          *int32             `json:"width_px,omitempty"`
	HeightPx         *int32             `json:"height_px,omitempty"`
	ModerationStatus string             `json:"moderation_status"`
	RejectionReason  string             `json:"rejection_reason,omitempty"`
	IsCover          bool               `json:"is_cover"`
	SortOrder        int16              `json:"sort_order"`
	Tags             []PhotoTagResponse `json:"tags"`
	CreatedAt        time.Time          `json:"created_at"`
}

type PhotoTagResponse struct {
	ID   int16  `json:"id"`
	Name string `json:"name"`
}

type UpdatePhotoRequest struct {
	IsCover   *bool   `json:"is_cover"`
	SortOrder *int16  `json:"sort_order"`
	TagIDs    []int16 `json:"tag_ids"`
}

type AddCategoryRequest struct {
	CategoryID int16 `json:"category_id" binding:"required"`
}

type ModelSearchFilter struct {
	City              string `form:"city"`
	MinAge            *int   `form:"min_age"`
	MaxAge            *int   `form:"max_age"`
	CategoryID        *int16 `form:"category_id"`
	WillingToRelocate *bool  `form:"willing_to_relocate"`
	Limit             int    `form:"limit,default=20"`
	Offset            int    `form:"offset,default=0"`
}

type PaginatedResponse struct {
	Data   any `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

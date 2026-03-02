package dto

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active blocked"`
}

type AdminUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	PhotoURL  string `json:"photo_url,omitempty"`
	CreatedAt string `json:"created_at"`
}

type AdminUsersFilter struct {
	Role   string `form:"role"`
	Status string `form:"status"`
	Limit  int    `form:"limit,default=20"`
	Offset int    `form:"offset,default=0"`
}

type ModeratePhotoRequest struct {
	Status          string `json:"status" binding:"required,oneof=approved rejected"`
	RejectionReason string `json:"rejection_reason"`
}

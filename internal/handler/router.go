package handler

import (
	"SHIBA/internal/domain"
	"SHIBA/internal/middleware"
	jwtpkg "SHIBA/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type Router struct {
	auth        *AuthHandler
	model       *ModelHandler
	agency      *AgencyHandler
	casting     *CastingHandler
	application *ApplicationHandler
	invitation  *InvitationHandler
	complaint   *ComplaintHandler
	ref         *RefHandler
	admin       *AdminHandler
	jwtManager  *jwtpkg.Manager
	corsHandler gin.HandlerFunc
	rateLimiter *middleware.RateLimiter
}

func NewRouter(
	auth *AuthHandler,
	model *ModelHandler,
	agency *AgencyHandler,
	casting *CastingHandler,
	application *ApplicationHandler,
	invitation *InvitationHandler,
	complaint *ComplaintHandler,
	ref *RefHandler,
	admin *AdminHandler,
	jwtManager *jwtpkg.Manager,
	corsHandler gin.HandlerFunc,
	rateLimiter *middleware.RateLimiter,
) *Router {
	return &Router{
		auth:        auth,
		model:       model,
		agency:      agency,
		casting:     casting,
		application: application,
		invitation:  invitation,
		complaint:   complaint,
		ref:         ref,
		admin:       admin,
		jwtManager:  jwtManager,
		corsHandler: corsHandler,
		rateLimiter: rateLimiter,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(r.corsHandler)
	engine.Use(r.rateLimiter.Middleware())

	authMiddleware := middleware.Auth(r.jwtManager)
	requireAdmin := middleware.RequireRoles(domain.RoleAdmin)
	requireAgency := middleware.RequireRoles(domain.RoleAgency)
	requireModel := middleware.RequireRoles(domain.RoleModel)
	requireAgencyOrAdmin := middleware.RequireRoles(domain.RoleAgency, domain.RoleAdmin)
	requireModelOrAgency := middleware.RequireRoles(domain.RoleModel, domain.RoleAgency, domain.RoleAdmin)

	api := engine.Group("/api/v1")

	// Auth
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", r.auth.Register)
		authGroup.GET("/verify-email", r.auth.VerifyEmail)
		authGroup.POST("/login", r.auth.Login)
		authGroup.POST("/refresh", r.auth.Refresh)
		authGroup.POST("/logout", authMiddleware, r.auth.Logout)
	}

	// Reference data (public)
	refGroup := api.Group("/ref")
	{
		refGroup.GET("/categories", r.ref.GetCategories)
		refGroup.GET("/photo-tags", r.ref.GetPhotoTags)
	}

	// Models
	modelsGroup := api.Group("/models")
	{
		modelsGroup.GET("", authMiddleware, requireAgencyOrAdmin, r.model.Search)
		modelsGroup.GET("/me", authMiddleware, requireModel, r.model.GetMe)
		modelsGroup.PUT("/me", authMiddleware, requireModel, r.model.UpdateMe)
		modelsGroup.POST("/me/categories", authMiddleware, requireModel, r.model.AddCategory)
		modelsGroup.DELETE("/me/categories/:id", authMiddleware, requireModel, r.model.RemoveCategory)
		modelsGroup.POST("/me/photos", authMiddleware, requireModel, r.model.UploadPhoto)
		modelsGroup.GET("/:id", authMiddleware, r.model.GetByID)
		modelsGroup.GET("/:id/photos", authMiddleware, r.model.GetPhotos)
		modelsGroup.PUT("/me/photos/:id", authMiddleware, requireModel, r.model.UpdatePhoto)
		modelsGroup.DELETE("/me/photos/:id", authMiddleware, requireModel, r.model.DeletePhoto)
	}

	// Agencies
	agenciesGroup := api.Group("/agencies")
	{
		agenciesGroup.GET("/me", authMiddleware, requireAgency, r.agency.GetMe)
		agenciesGroup.PUT("/me", authMiddleware, requireAgency, r.agency.UpdateMe)
		agenciesGroup.POST("/me/logo", authMiddleware, requireAgency, r.agency.UploadLogo)
		agenciesGroup.GET("/me/invitations", authMiddleware, requireAgency, r.invitation.GetMyOutgoingInvitations)
		agenciesGroup.GET("/:id", authMiddleware, r.agency.GetByID)
	}

	// Castings
	castingsGroup := api.Group("/castings")
	{
		castingsGroup.POST("", authMiddleware, requireAgency, r.casting.CreateCasting)
		castingsGroup.GET("", authMiddleware, requireModelOrAgency, r.casting.ListCastings)
		castingsGroup.GET("/:id", authMiddleware, r.casting.GetCasting)
		castingsGroup.PUT("/:id", authMiddleware, requireAgency, r.casting.UpdateCasting)
		castingsGroup.DELETE("/:id", authMiddleware, requireAgency, r.casting.DeleteCasting)
		castingsGroup.GET("/:id/applications", authMiddleware, requireAgency, r.casting.GetApplicationsByCasting)
		castingsGroup.PUT("/:id/applications/:app_id", authMiddleware, requireAgency, r.casting.UpdateApplicationStatus)
	}

	// Applications
	applicationsGroup := api.Group("/applications")
	{
		applicationsGroup.POST("", authMiddleware, requireModel, r.application.CreateApplication)
		applicationsGroup.GET("/me", authMiddleware, requireModel, r.application.GetMyApplications)
		applicationsGroup.DELETE("/:id", authMiddleware, requireModel, r.application.DeleteApplication)
	}

	// Invitations
	invitationsGroup := api.Group("/invitations")
	{
		invitationsGroup.POST("", authMiddleware, requireAgency, r.invitation.CreateInvitation)
		invitationsGroup.GET("/me", authMiddleware, requireModel, r.invitation.GetMyInvitations)
		invitationsGroup.PUT("/:id", authMiddleware, requireModel, r.invitation.UpdateInvitationStatus)
	}

	// Complaints
	api.POST("/complaints", authMiddleware, r.complaint.CreateComplaint)

	// Admin
	adminGroup := api.Group("/admin")
	adminGroup.Use(authMiddleware, requireAdmin)
	{
		adminGroup.GET("/users", r.admin.ListUsers)
		adminGroup.PUT("/users/:id/status", r.admin.UpdateUserStatus)
		adminGroup.GET("/agencies/pending", r.admin.GetPendingAgencies)
		adminGroup.PUT("/agencies/:id/approve", r.admin.ApproveAgency)
		adminGroup.PUT("/agencies/:id/reject", r.admin.RejectAgency)
		adminGroup.GET("/photos/pending", r.admin.GetPendingPhotos)
		adminGroup.PUT("/photos/:id/moderate", r.admin.ModeratePhoto)
		adminGroup.GET("/complaints", r.admin.ListComplaints)
		adminGroup.PUT("/complaints/:id", r.admin.UpdateComplaint)
	}
}

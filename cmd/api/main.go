package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SHIBA/internal/config"
	"SHIBA/internal/handler"
	"SHIBA/internal/middleware"
	"SHIBA/internal/repository"
	"SHIBA/internal/service"
	"SHIBA/internal/storage"
	jwtpkg "SHIBA/pkg/jwt"
	"SHIBA/pkg/mail"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database
	ctx := context.Background()
	pool, err := repository.NewPool(ctx, cfg.DB.URL())
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := runMigrations(cfg.DB.URL()); err != nil {
		log.Printf("WARNING: migrations: %v", err)
	}

	// MinIO
	minioClient, err := storage.NewMinIOClient(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKeyID,
		cfg.MinIO.SecretAccessKey,
		cfg.MinIO.BucketName,
		cfg.MinIO.PublicURL,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		log.Fatalf("init minio: %v", err)
	}
	if err := minioClient.EnsureBucket(ctx); err != nil {
		log.Printf("WARNING: ensure minio bucket: %v", err)
	}
	log.Println("Connected to MinIO")

	// JWT
	jwtManager := jwtpkg.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Mail
	mailSender := mail.NewSender(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	modelRepo := repository.NewModelRepository(pool)
	agencyRepo := repository.NewAgencyRepository(pool)
	castingRepo := repository.NewCastingRepository(pool)
	complaintRepo := repository.NewComplaintRepository(pool)

	// Services
	authService := service.NewAuthService(userRepo, jwtManager, mailSender, cfg.App.FrontendURL)
	modelService := service.NewModelService(modelRepo, minioClient)
	agencyService := service.NewAgencyService(agencyRepo, minioClient)
	castingService := service.NewCastingService(castingRepo, agencyRepo, modelRepo, userRepo, mailSender)
	complaintService := service.NewComplaintService(complaintRepo)
	adminService := service.NewAdminService(userRepo, modelRepo, agencyRepo, complaintRepo, mailSender)

	// Middleware
	storageURL := fmt.Sprintf("%s/%s", cfg.MinIO.PublicURL, cfg.MinIO.BucketName)
	corsMiddleware := middleware.CORS(cfg.App.FrontendURL)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	modelHandler := handler.NewModelHandler(modelService, storageURL)
	agencyHandler := handler.NewAgencyHandler(agencyService)
	castingHandler := handler.NewCastingHandler(castingService)
	applicationHandler := handler.NewApplicationHandler(castingService)
	invitationHandler := handler.NewInvitationHandler(castingService)
	complaintHandler := handler.NewComplaintHandler(complaintService)
	refHandler := handler.NewRefHandler(modelService)
	adminHandler := handler.NewAdminHandler(adminService, storageURL)

	// Router
	router := handler.NewRouter(
		authHandler,
		modelHandler,
		agencyHandler,
		castingHandler,
		applicationHandler,
		invitationHandler,
		complaintHandler,
		refHandler,
		adminHandler,
		jwtManager,
		corsMiddleware,
		rateLimiter,
	)

	engine := gin.Default()
	engine.MaxMultipartMemory = 15 << 20 // 15MB
	router.Setup(engine)

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "env": cfg.App.Env})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on :%s (env=%s)\n", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Println("Migrations applied")
	return nil
}

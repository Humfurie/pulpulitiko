package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/humfurie/pulpulitiko/api/internal/config"
	"github.com/humfurie/pulpulitiko/api/internal/handlers"
	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/internal/services"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
	"github.com/humfurie/pulpulitiko/api/pkg/storage"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	if os.Getenv("APP_ENV") == "development" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	}

	// Load configuration
	cfg := config.Load()

	ctx := context.Background()

	// Initialize database
	logger.Info().Msg("Connecting to database...")
	db, err := repository.NewDBPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	logger.Info().Msg("Database connected")

	// Initialize Redis cache
	logger.Info().Msg("Connecting to Redis...")
	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisCache.Close()
	logger.Info().Msg("Redis connected")

	// Initialize MinIO storage
	logger.Info().Msg("Connecting to MinIO...")
	minioStorage, err := storage.NewMinioStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to MinIO")
	}
	logger.Info().Msg("MinIO connected")

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	articleService := services.NewArticleService(articleRepo, redisCache)
	categoryService := services.NewCategoryService(categoryRepo, redisCache)
	tagService := services.NewTagService(tagRepo)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	uploadService := services.NewUploadService(minioStorage)

	// Initialize handlers
	articleHandler := handlers.NewArticleHandler(articleService)
	categoryHandler := handlers.NewCategoryHandler(categoryService, articleService)
	tagHandler := handlers.NewTagHandler(tagService, articleService)
	authHandler := handlers.NewAuthHandler(authService)
	uploadHandler := handlers.NewUploadHandler(uploadService)
	healthHandler := handlers.NewHealthHandler()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	rateLimiter := middleware.NewRateLimiter(redisCache, 100, 60) // 100 requests per minute

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(rateLimiter.Limit)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify exact origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", healthHandler.Health)

	// Public API routes
	r.Route("/api", func(r chi.Router) {
		// Articles
		r.Get("/articles", articleHandler.List)
		r.Get("/articles/trending", articleHandler.GetTrending)
		r.Get("/articles/{slug}", articleHandler.GetBySlug)

		// Categories
		r.Get("/categories", categoryHandler.List)
		r.Get("/categories/{slug}", categoryHandler.GetArticlesBySlug)

		// Tags
		r.Get("/tags", tagHandler.List)
		r.Get("/tags/{slug}", tagHandler.GetArticlesBySlug)

		// Search
		r.Get("/search", articleHandler.Search)

		// Auth
		r.Post("/auth/login", authHandler.Login)
		r.With(authMiddleware.Authenticate).Get("/auth/me", authHandler.GetCurrentUser)
	})

	// Admin API routes (authenticated)
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)

		// Articles
		r.Get("/articles", articleHandler.AdminList)
		r.Get("/articles/{id}", articleHandler.AdminGetByID)
		r.Post("/articles", articleHandler.Create)
		r.Put("/articles/{id}", articleHandler.Update)
		r.Delete("/articles/{id}", articleHandler.Delete)

		// Categories
		r.Post("/categories", categoryHandler.Create)
		r.Put("/categories/{id}", categoryHandler.Update)
		r.Delete("/categories/{id}", categoryHandler.Delete)

		// Tags
		r.Post("/tags", tagHandler.Create)
		r.Put("/tags/{id}", tagHandler.Update)
		r.Delete("/tags/{id}", tagHandler.Delete)

		// Upload
		r.Post("/upload", uploadHandler.Upload)

		// Users (admin only)
		r.With(authMiddleware.RequireAdmin).Post("/users", authHandler.CreateUser)
	})

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info().Str("port", cfg.AppPort).Msg("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

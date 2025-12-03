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
	"github.com/humfurie/pulpulitiko/api/pkg/email"
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
		cfg.MinioPublicEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to MinIO")
	}
	logger.Info().Msg("MinIO connected")

	// Initialize email service
	emailService := email.NewEmailService(
		cfg.ResendAPIKey,
		cfg.EmailFromEmail,
		cfg.EmailFromName,
		cfg.FrontendURL,
	)
	if emailService.IsConfigured() {
		logger.Info().Msg("Email service configured")
	} else {
		logger.Warn().Msg("Email service not configured (RESEND_API_KEY not set)")
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	userRepo := repository.NewUserRepository(db)
	authorRepo := repository.NewAuthorRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	articleService := services.NewArticleService(articleRepo, redisCache)
	categoryService := services.NewCategoryService(categoryRepo, redisCache)
	tagService := services.NewTagService(tagRepo)
	authService := services.NewAuthService(userRepo, roleRepo, authorRepo, emailService, cfg.JWTSecret)
	uploadService := services.NewUploadService(minioStorage)
	authorService := services.NewAuthorService(authorRepo)
	roleService := services.NewRoleService(roleRepo, permissionRepo)
	commentService := services.NewCommentService(commentRepo, articleRepo)
	messageService := services.NewMessageService(messageRepo)

	// Initialize WebSocket hub
	wsHub := handlers.NewHub()
	go wsHub.Run()

	// Initialize handlers
	articleHandler := handlers.NewArticleHandler(articleService)
	categoryHandler := handlers.NewCategoryHandler(categoryService, articleService)
	tagHandler := handlers.NewTagHandler(tagService, articleService)
	authHandler := handlers.NewAuthHandler(authService)
	uploadHandler := handlers.NewUploadHandler(uploadService)
	healthHandler := handlers.NewHealthHandler()
	authorHandler := handlers.NewAuthorHandler(authorService, articleService)
	metricsHandler := handlers.NewMetricsHandler(metricsRepo)
	roleHandler := handlers.NewRoleHandler(roleService)
	commentHandler := handlers.NewCommentHandler(commentService)
	rssHandler := handlers.NewRSSHandler(articleService, cfg.SiteURL)
	userHandler := handlers.NewUserHandler(userRepo)
	messageHandler := handlers.NewMessageHandler(messageService, wsHub)
	wsHandler := handlers.NewWebSocketHandler(wsHub, authService, messageService)

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

	// RSS Feed
	r.Get("/rss", rssHandler.Feed)
	r.Get("/feed", rssHandler.Feed)

	// WebSocket endpoint
	r.Get("/ws", wsHandler.HandleWebSocket)

	// Public API routes
	r.Route("/api", func(r chi.Router) {
		// Articles - use nested routing to avoid route conflicts
		r.Get("/articles", articleHandler.List)
		r.Get("/articles/trending", articleHandler.GetTrending)
		r.Route("/articles/{slug}", func(r chi.Router) {
			r.Get("/", articleHandler.GetBySlug)
			r.Post("/view", articleHandler.IncrementViewCount)
			r.Get("/related", articleHandler.GetRelatedArticles)
			// Comments for this article - use OptionalAuth to identify user for reaction status
			r.With(authMiddleware.OptionalAuth).Get("/comments", commentHandler.ListComments)
			r.Get("/comments/count", commentHandler.GetCommentCount)
			r.With(authMiddleware.Authenticate).Post("/comments", commentHandler.CreateComment)
		})

		// Categories
		r.Get("/categories", categoryHandler.List)
		r.Get("/categories/{slug}", categoryHandler.GetArticlesBySlug)

		// Tags
		r.Get("/tags", tagHandler.List)
		r.Get("/tags/{slug}", tagHandler.GetArticlesBySlug)

		// Authors
		r.Get("/authors", authorHandler.List)
		r.Get("/authors/{slug}", authorHandler.GetArticlesBySlug)

		// Search
		r.Get("/search", articleHandler.Search)

		// Comments - standalone routes (by ID) - use OptionalAuth for reaction status
		r.With(authMiddleware.OptionalAuth).Get("/comments/{id}", commentHandler.GetComment)
		r.With(authMiddleware.OptionalAuth).Get("/comments/{id}/replies", commentHandler.GetReplies)
		r.With(authMiddleware.Authenticate).Put("/comments/{id}", commentHandler.UpdateComment)
		r.With(authMiddleware.Authenticate).Delete("/comments/{id}", commentHandler.DeleteComment)
		r.With(authMiddleware.Authenticate).Post("/comments/{id}/reactions", commentHandler.AddReaction)
		r.With(authMiddleware.Authenticate).Delete("/comments/{id}/reactions/{reaction}", commentHandler.RemoveReaction)

		// Auth
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)
		r.Post("/auth/reset-password", authHandler.ResetPassword)
		r.With(authMiddleware.Authenticate).Get("/auth/me", authHandler.GetCurrentUser)
		r.With(authMiddleware.Authenticate).Get("/auth/account", authorHandler.GetAccount)
		r.With(authMiddleware.Authenticate).Put("/auth/account", authorHandler.UpdateAccount)

		// User profiles (public)
		r.Get("/users/mentionable", userHandler.GetMentionableUsers)
		r.Get("/users/{slug}/profile", userHandler.GetUserProfile)
		r.Get("/users/{slug}/comments", userHandler.GetUserComments)
		r.Get("/users/{slug}/replies", userHandler.GetUserReplies)

		// Messaging (authenticated users)
		r.Route("/messages", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/unread", messageHandler.GetUnreadCounts)
			r.Get("/conversations", messageHandler.GetMyConversations)
			r.Post("/conversations", messageHandler.CreateConversation)
			r.Get("/conversations/{id}", messageHandler.GetConversation)
			r.Get("/conversations/{id}/messages", messageHandler.GetMessages)
			r.Post("/conversations/{id}/messages", messageHandler.SendMessage)
			r.Post("/conversations/{id}/read", messageHandler.MarkAsRead)
		})
	})

	// Admin API routes (authenticated)
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)

		// Metrics
		r.Get("/metrics", metricsHandler.GetDashboardMetrics)
		r.Get("/metrics/top-articles", metricsHandler.GetTopArticles)
		r.Get("/metrics/categories", metricsHandler.GetCategoryMetrics)
		r.Get("/metrics/tags", metricsHandler.GetTagMetrics)

		// Articles
		r.Get("/articles", articleHandler.AdminList)
		r.Get("/articles/{id}", articleHandler.AdminGetByID)
		r.Post("/articles", articleHandler.Create)
		r.Put("/articles/{id}", articleHandler.Update)
		r.Delete("/articles/{id}", articleHandler.Delete)
		r.Post("/articles/{id}/restore", articleHandler.Restore)

		// Categories
		r.Get("/categories", categoryHandler.AdminList)
		r.Get("/categories/{id}", categoryHandler.AdminGetByID)
		r.Post("/categories", categoryHandler.Create)
		r.Put("/categories/{id}", categoryHandler.Update)
		r.Delete("/categories/{id}", categoryHandler.Delete)
		r.Post("/categories/{id}/restore", categoryHandler.Restore)

		// Tags
		r.Get("/tags/{id}", tagHandler.AdminGetByID)
		r.Post("/tags", tagHandler.Create)
		r.Put("/tags/{id}", tagHandler.Update)
		r.Delete("/tags/{id}", tagHandler.Delete)
		r.Post("/tags/{id}/restore", tagHandler.Restore)

		// Upload
		r.Post("/upload", uploadHandler.Upload)

		// Users management (admin only)
		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/", authorHandler.AdminList)
			r.Get("/{id}", authorHandler.AdminGetByID)
			r.Post("/", authorHandler.AdminCreate)
			r.Put("/{id}", authorHandler.AdminUpdate)
			r.Delete("/{id}", authorHandler.AdminDelete)
			r.Post("/{id}/restore", authorHandler.AdminRestore)
		})

		// Roles management (admin only)
		r.Route("/roles", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/", roleHandler.List)
			r.Get("/permissions", roleHandler.ListPermissions)
			r.Get("/{id}", roleHandler.GetByID)
			r.Post("/", roleHandler.Create)
			r.Put("/{id}", roleHandler.Update)
			r.Delete("/{id}", roleHandler.Delete)
			r.Post("/{id}/restore", roleHandler.Restore)
		})

		// Comments moderation (admin only)
		r.Route("/comments", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/", commentHandler.ListAllComments)
			r.Put("/{id}/moderate", commentHandler.ModerateComment)
		})

		// Messaging management (admin only)
		r.Route("/messages", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/conversations", messageHandler.AdminListConversations)
			r.Get("/conversations/{id}", messageHandler.GetConversation)
			r.Get("/conversations/{id}/messages", messageHandler.GetMessages)
			r.Post("/conversations/{id}/messages", messageHandler.SendMessage)
			r.Post("/conversations/{id}/read", messageHandler.MarkAsRead)
			r.Patch("/conversations/{id}/status", messageHandler.AdminUpdateConversationStatus)
		})
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

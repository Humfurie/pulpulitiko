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
	politicianRepo := repository.NewPoliticianRepository(db)
	searchAnalyticsRepo := repository.NewSearchAnalyticsRepository(db)
	politicianCommentRepo := repository.NewPoliticianCommentRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	politicalPartyRepo := repository.NewPoliticalPartyRepository(db)
	billRepo := repository.NewBillRepository(db)
	electionRepo := repository.NewElectionRepository(db)
	pollRepo := repository.NewPollRepository(db)

	// Initialize HTML sanitizer for XSS protection
	htmlSanitizer := services.NewHTMLSanitizer()

	// Initialize services
	politicianService := services.NewPoliticianService(politicianRepo, redisCache)
	articleService := services.NewArticleService(articleRepo, politicianRepo, redisCache, htmlSanitizer)
	categoryService := services.NewCategoryService(categoryRepo, redisCache)
	tagService := services.NewTagService(tagRepo)
	authService := services.NewAuthService(userRepo, roleRepo, authorRepo, emailService, cfg.JWTSecret)
	uploadService := services.NewUploadService(minioStorage)
	authorService := services.NewAuthorService(authorRepo)
	roleService := services.NewRoleService(roleRepo, permissionRepo)
	messageService := services.NewMessageService(messageRepo)
	searchAnalyticsService := services.NewSearchAnalyticsService(searchAnalyticsRepo)
	notificationService := services.NewNotificationService(notificationRepo, userRepo)
	commentService := services.NewCommentService(commentRepo, articleRepo, notificationService, htmlSanitizer)
	politicianCommentService := services.NewPoliticianCommentService(politicianCommentRepo, politicianRepo, notificationService)
	locationService := services.NewLocationService(locationRepo, redisCache)
	politicalPartyService := services.NewPoliticalPartyService(politicalPartyRepo, redisCache)
	billService := services.NewBillService(billRepo, redisCache)
	electionService := services.NewElectionService(electionRepo, redisCache)
	pollService := services.NewPollService(pollRepo, redisCache)

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
	politicianHandler := handlers.NewPoliticianHandler(politicianService, articleService)
	searchAnalyticsHandler := handlers.NewSearchAnalyticsHandler(searchAnalyticsService)
	politicianCommentHandler := handlers.NewPoliticianCommentHandler(politicianCommentService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	locationHandler := handlers.NewLocationHandler(locationService)
	politicalPartyHandler := handlers.NewPoliticalPartyHandler(politicalPartyService)
	billHandler := handlers.NewBillHandler(billService)
	electionHandler := handlers.NewElectionHandler(electionService)
	pollHandler := handlers.NewPollHandler(pollService)

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

		// Politicians
		r.Get("/politicians", politicianHandler.List)
		r.Get("/politicians/search", politicianHandler.Search)
		r.Route("/politicians/{slug}", func(r chi.Router) {
			r.Get("/", politicianHandler.GetBySlug)
			// Politician comments
			r.With(authMiddleware.OptionalAuth).Get("/comments", politicianCommentHandler.ListComments)
			r.Get("/comments/count", politicianCommentHandler.GetCommentCount)
			r.With(authMiddleware.Authenticate).Post("/comments", politicianCommentHandler.CreateComment)
		})

		// Locations (Philippine Geographic Hierarchy)
		r.Route("/locations", func(r chi.Router) {
			r.Get("/regions", locationHandler.ListRegions)
			r.Get("/regions/{slug}", locationHandler.GetRegionBySlug)
			r.Get("/provinces", locationHandler.ListAllProvinces)
			r.Get("/provinces/{slug}", locationHandler.GetProvinceBySlug)
			r.Get("/provinces/by-region/{region_id}", locationHandler.GetProvincesByRegion)
			r.Get("/cities/{slug}", locationHandler.GetCityBySlug)
			r.Get("/cities/by-province/{province_id}", locationHandler.GetCitiesByProvince)
			r.Get("/barangays/{slug}", locationHandler.GetBarangayBySlug)
			r.Get("/barangays/by-city/{city_id}", locationHandler.GetBarangaysByCity)
			r.Get("/districts/{slug}", locationHandler.GetDistrictBySlug)
			r.Get("/districts/by-province/{province_id}", locationHandler.GetDistrictsByProvince)
			r.Get("/search", locationHandler.SearchLocations)
			r.Get("/hierarchy/{barangay_id}", locationHandler.GetHierarchy)
		})

		// Political Parties
		r.Route("/parties", func(r chi.Router) {
			r.Get("/", politicalPartyHandler.GetParties)
			r.Get("/all", politicalPartyHandler.GetAllParties)
			r.Get("/{slug}", politicalPartyHandler.GetPartyBySlug)
		})

		// Government Positions
		r.Route("/positions", func(r chi.Router) {
			r.Get("/", politicalPartyHandler.GetAllPositions)
			r.Get("/level/{level}", politicalPartyHandler.GetPositionsByLevel)
			r.Get("/{slug}", politicalPartyHandler.GetPositionBySlug)
		})

		// Find My Representatives
		r.Get("/my-representatives", politicalPartyHandler.FindMyRepresentatives)

		// Legislation / Bills
		r.Route("/legislation", func(r chi.Router) {
			// Sessions
			r.Get("/sessions", billHandler.ListSessions)
			r.Get("/sessions/current", billHandler.GetCurrentSession)

			// Committees
			r.Get("/committees", billHandler.ListCommittees)
			r.Get("/committees/{slug}", billHandler.GetCommitteeBySlug)

			// Topics
			r.Get("/topics", billHandler.ListAllTopics)

			// Bills
			r.Get("/bills", billHandler.ListBills)
			r.Get("/bills/{slug}", billHandler.GetBillBySlug)
			r.Get("/bills/id/{id}", billHandler.GetBillByID)
			r.Get("/bills/{id}/votes", billHandler.GetBillVotes)
			r.Get("/votes/{voteId}/politicians", billHandler.GetPoliticianVotesForBillVote)

			// Politician voting records
			r.Get("/politicians/{id}/votes", billHandler.GetPoliticianVotingHistory)
			r.Get("/politicians/{id}/voting-record", billHandler.GetPoliticianVotingRecord)
		})

		// Elections
		r.Route("/elections", func(r chi.Router) {
			r.Get("/", electionHandler.ListElections)
			r.Get("/upcoming", electionHandler.GetUpcomingElections)
			r.Get("/featured", electionHandler.GetFeaturedElections)
			r.Get("/calendar", electionHandler.GetElectionCalendar)
			r.Get("/slug/{slug}", electionHandler.GetElectionBySlug)
			r.Get("/{id}", electionHandler.GetElectionByID)
			r.Get("/{id}/positions", electionHandler.GetElectionPositions)
		})

		// Candidates
		r.Route("/candidates", func(r chi.Router) {
			r.Get("/", electionHandler.ListCandidates)
			r.Get("/{id}", electionHandler.GetCandidateByID)
			r.Get("/position/{positionId}", electionHandler.GetCandidatesForPosition)
		})

		// Voter Education
		r.Route("/voter-education", func(r chi.Router) {
			r.Get("/", electionHandler.ListVoterEducation)
			r.Get("/{slug}", electionHandler.GetVoterEducationBySlug)
		})

		// Polls
		r.Route("/polls", func(r chi.Router) {
			r.Get("/", pollHandler.ListPolls)
			r.Get("/featured", pollHandler.GetFeaturedPolls)
			r.Get("/slug/{slug}", pollHandler.GetPollBySlug)
			r.Get("/{id}", pollHandler.GetPollByID)
			r.Get("/{id}/results", pollHandler.GetPollResults)
			r.With(authMiddleware.OptionalAuth).Post("/{id}/vote", pollHandler.CastVote)
			// Poll comments
			r.With(authMiddleware.OptionalAuth).Get("/{id}/comments", pollHandler.GetPollComments)
			r.With(authMiddleware.Authenticate).Post("/{id}/comments", pollHandler.CreatePollComment)
		})

		// Authenticated user poll routes
		r.Route("/my-polls", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/", pollHandler.GetMyPolls)
			r.Post("/", pollHandler.CreatePoll)
			r.Put("/{id}", pollHandler.UpdatePoll)
			r.Post("/{id}/submit", pollHandler.SubmitForApproval)
			r.Delete("/{id}", pollHandler.DeletePoll)
		})

		// Search
		r.Get("/search", articleHandler.Search)

		// Search analytics tracking (public, uses OptionalAuth to identify user)
		r.With(authMiddleware.OptionalAuth).Post("/search/track", searchAnalyticsHandler.TrackSearch)
		r.Post("/search/click", searchAnalyticsHandler.TrackClick)

		// Comments - standalone routes (by ID) - use OptionalAuth for reaction status
		r.With(authMiddleware.OptionalAuth).Get("/comments/{id}", commentHandler.GetComment)
		r.With(authMiddleware.OptionalAuth).Get("/comments/{id}/replies", commentHandler.GetReplies)
		r.With(authMiddleware.Authenticate).Put("/comments/{id}", commentHandler.UpdateComment)
		r.With(authMiddleware.Authenticate).Delete("/comments/{id}", commentHandler.DeleteComment)
		r.With(authMiddleware.Authenticate).Post("/comments/{id}/reactions", commentHandler.AddReaction)
		r.With(authMiddleware.Authenticate).Delete("/comments/{id}/reactions/{reaction}", commentHandler.RemoveReaction)

		// Politician comments - standalone routes (by ID)
		r.With(authMiddleware.OptionalAuth).Get("/politician-comments/{id}", politicianCommentHandler.GetComment)
		r.With(authMiddleware.OptionalAuth).Get("/politician-comments/{id}/replies", politicianCommentHandler.GetReplies)
		r.With(authMiddleware.Authenticate).Put("/politician-comments/{id}", politicianCommentHandler.UpdateComment)
		r.With(authMiddleware.Authenticate).Delete("/politician-comments/{id}", politicianCommentHandler.DeleteComment)
		r.With(authMiddleware.Authenticate).Post("/politician-comments/{id}/reactions", politicianCommentHandler.AddReaction)
		r.With(authMiddleware.Authenticate).Delete("/politician-comments/{id}/reactions/{reaction}", politicianCommentHandler.RemoveReaction)

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

		// Notifications (authenticated users)
		r.Route("/notifications", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/", notificationHandler.ListNotifications)
			r.Get("/unread-count", notificationHandler.GetUnreadCount)
			r.Post("/{id}/read", notificationHandler.MarkAsRead)
			r.Post("/read-all", notificationHandler.MarkAllAsRead)
			r.Delete("/{id}", notificationHandler.DeleteNotification)
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

		// Search Analytics (admin only)
		r.Get("/analytics/search", searchAnalyticsHandler.GetAnalytics)

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
		r.Get("/tags", tagHandler.AdminList)
		r.Get("/tags/{id}", tagHandler.AdminGetByID)
		r.Post("/tags", tagHandler.Create)
		r.Put("/tags/{id}", tagHandler.Update)
		r.Delete("/tags/{id}", tagHandler.Delete)
		r.Post("/tags/{id}/restore", tagHandler.Restore)

		// Politicians
		r.Get("/politicians", politicianHandler.AdminList)
		r.Get("/politicians/{id}", politicianHandler.AdminGetByID)
		r.Post("/politicians", politicianHandler.Create)
		r.Put("/politicians/{id}", politicianHandler.Update)
		r.Delete("/politicians/{id}", politicianHandler.Delete)
		r.Post("/politicians/{id}/restore", politicianHandler.Restore)

		// Locations management (admin only)
		r.Route("/locations", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			// Regions
			r.Get("/regions/{id}", locationHandler.AdminGetRegionByID)
			r.Post("/regions", locationHandler.CreateRegion)
			r.Put("/regions/{id}", locationHandler.UpdateRegion)
			r.Delete("/regions/{id}", locationHandler.DeleteRegion)
			// Provinces
			r.Get("/provinces/{id}", locationHandler.AdminGetProvinceByID)
			r.Post("/provinces", locationHandler.CreateProvince)
			r.Put("/provinces/{id}", locationHandler.UpdateProvince)
			r.Delete("/provinces/{id}", locationHandler.DeleteProvince)
			// Cities
			r.Get("/cities/{id}", locationHandler.AdminGetCityByID)
			r.Post("/cities", locationHandler.CreateCity)
			r.Put("/cities/{id}", locationHandler.UpdateCity)
			r.Delete("/cities/{id}", locationHandler.DeleteCity)
			// Barangays
			r.Get("/barangays/{id}", locationHandler.AdminGetBarangayByID)
			r.Post("/barangays", locationHandler.CreateBarangay)
			r.Put("/barangays/{id}", locationHandler.UpdateBarangay)
			r.Delete("/barangays/{id}", locationHandler.DeleteBarangay)
			// Districts
			r.Get("/districts/{id}", locationHandler.AdminGetDistrictByID)
			r.Post("/districts", locationHandler.CreateDistrict)
		})

		// Political Parties management (admin only)
		r.Route("/parties", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Post("/", politicalPartyHandler.CreateParty)
			r.Put("/{id}", politicalPartyHandler.UpdateParty)
			r.Delete("/{id}", politicalPartyHandler.DeleteParty)
		})

		// Government Positions management (admin only)
		r.Route("/positions", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/{id}", politicalPartyHandler.GetPositionByID)
			r.Post("/", politicalPartyHandler.CreatePosition)
			r.Put("/{id}", politicalPartyHandler.UpdatePosition)
			r.Delete("/{id}", politicalPartyHandler.DeletePosition)
		})

		// Politician Jurisdictions management (admin only)
		r.Route("/jurisdictions", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Post("/", politicalPartyHandler.CreateJurisdiction)
			r.Get("/politician/{politicianId}", politicalPartyHandler.GetJurisdictionsByPolitician)
			r.Delete("/{id}", politicalPartyHandler.DeleteJurisdiction)
		})

		// Legislation / Bills management (admin only)
		r.Route("/legislation", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			// Bills CRUD
			r.Post("/bills", billHandler.CreateBill)
			r.Put("/bills/{id}", billHandler.UpdateBill)
			r.Delete("/bills/{id}", billHandler.DeleteBill)
			// Bill status updates
			r.Post("/bills/{id}/status", billHandler.AddBillStatus)
			// Bill votes
			r.Post("/bills/{id}/votes", billHandler.AddBillVote)
		})

		// Elections management (admin only)
		r.Route("/elections", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			// Elections CRUD
			r.Post("/", electionHandler.CreateElection)
			r.Put("/{id}", electionHandler.UpdateElection)
			r.Delete("/{id}", electionHandler.DeleteElection)
			// Election positions
			r.Post("/positions", electionHandler.CreateElectionPosition)
			// Candidates
			r.Post("/candidates", electionHandler.CreateCandidate)
			r.Put("/candidates/{id}", electionHandler.UpdateCandidate)
			// Voter education
			r.Post("/voter-education", electionHandler.CreateVoterEducation)
		})

		// Polls management (admin only)
		r.Route("/polls", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/", pollHandler.AdminListPolls)
			r.Put("/{id}", pollHandler.AdminUpdatePoll)
			r.Post("/{id}/approve", pollHandler.ApprovePoll)
			r.Post("/{id}/close", pollHandler.ClosePoll)
			r.Delete("/{id}", pollHandler.DeletePoll)
			r.Delete("/comments/{id}", pollHandler.DeletePollComment)
		})

		// Upload
		r.Post("/upload", uploadHandler.Upload)

		// Users management (admin only)
		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/", userHandler.AdminList)
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

		// Politician comments moderation (admin only)
		r.Route("/politician-comments", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Put("/{id}/moderate", politicianCommentHandler.ModerateComment)
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

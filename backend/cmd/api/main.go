package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/config"
	"perchly-backend/internal/crypto"
	"perchly-backend/internal/email"
	"perchly-backend/internal/embedding"
	"perchly-backend/internal/handler"
	"perchly-backend/internal/llm"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/requestlog"
	"perchly-backend/internal/service"

	_ "perchly-backend/docs" // swag init tarafından üretilir; @title vb. burada değil docs.go'da tutulur
)

// @title Perchly Backend API
// @version 1.0
// @description Perchly sohbet uygulamasının REST API'si — kimlik doğrulama, onboarding, persona, konuşma/mesajlaşma ve kota uçları.
// @description Kimlik doğrulama gerektiren uçlar için "Authorize" düğmesinden `Bearer <access_token>` girin.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.JWTSecretIsEphemeral {
		log.Printf("warning: JWT_SECRET not set, using a random per-process secret — tokens won't survive a restart")
	}
	if cfg.LLMTokenEncryptionKeyIsEphemeral {
		log.Printf("warning: LLM_TOKEN_ENCRYPTION_KEY not set, using a random per-process key — stored LLM credentials won't survive a restart")
	}

	ctx := context.Background()
	pool, err := repository.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("create postgres pool: %v", err)
	}
	defer pool.Close()

	llmTokenBox, err := crypto.NewSecretBox(cfg.LLMTokenEncryptionKey)
	if err != nil {
		log.Fatalf("build LLM token secret box: %v", err)
	}

	healthService := service.NewHealthService(pool)
	healthHandler := handler.NewHealthHandler(healthService)

	onboardingProfileRepo := repository.NewOnboardingProfileRepository(pool)

	personaRepo := repository.NewPersonaRepository(pool)
	personaService := service.NewPersonaService(personaRepo, onboardingProfileRepo)
	personaHandler := handler.NewPersonaHandler(personaService)

	personaTraitsRepo := repository.NewPersonaTraitsRepository(pool)
	personaTraitsService := service.NewPersonaTraitsService(personaTraitsRepo, personaRepo)
	personaTraitsHandler := handler.NewPersonaTraitsHandler(personaTraitsService)

	var emailSender email.Sender
	if cfg.SMTPUsername == "" || cfg.SMTPPassword == "" {
		log.Printf("warning: SMTP_USERNAME/SMTP_PASSWORD not set, verification codes will only be logged, not emailed")
		emailSender = email.NewConsoleSender()
	} else {
		emailSender = email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFromName)
	}

	userRepo := repository.NewUserRepository(pool)
	verificationCodeRepo := repository.NewVerificationCodeRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	accessTokenIssuer := auth.NewAccessTokenIssuer(cfg.JWTSecret, cfg.AccessTokenTTL)

	authService := service.NewAuthService(
		userRepo, verificationCodeRepo, refreshTokenRepo, onboardingProfileRepo,
		accessTokenIssuer, emailSender, cfg.RefreshTokenTTL,
	)
	authHandler := handler.NewAuthHandler(authService)

	onboardingService := service.NewOnboardingService(onboardingProfileRepo)
	onboardingHandler := handler.NewOnboardingHandler(onboardingService)

	conversationRepo := repository.NewConversationRepository(pool)

	llmClient, err := llm.New(llm.Config{
		Provider: cfg.LLMProvider,
		APIKey:   cfg.LLMAPIKey,
		Model:    cfg.LLMModel,
		BaseURL:  cfg.LLMBaseURL,
	})
	if err != nil {
		// Don't crash the whole server over missing LLM config: personas
		// and every other endpoint should keep working. Requests that
		// actually need the LLM will fail with this same error.
		log.Printf("warning: LLM client not configured: %v", err)
		llmClient = llm.NewUnconfiguredClient(err)
	}

	embeddingClient, err := embedding.New(embedding.Config{
		Provider: cfg.EmbeddingProvider,
		APIKey:   cfg.EmbeddingAPIKey,
		Model:    cfg.EmbeddingModel,
		BaseURL:  cfg.EmbeddingBaseURL,
	})
	if err != nil {
		// Same reasoning as the LLM client: embeddings improve context
		// quality but chatting must keep working without them.
		log.Printf("warning: embedding client not configured: %v", err)
		embeddingClient = embedding.NewUnconfiguredClient(err)
	}

	messageRepo := repository.NewMessageRepository(pool)
	conversationService := service.NewConversationService(conversationRepo, personaRepo, onboardingProfileRepo, messageRepo)
	conversationHandler := handler.NewConversationHandler(conversationService)
	messageEmbeddingRepo := repository.NewMessageEmbeddingRepository(pool)
	summaryRepo := repository.NewConversationSummaryRepository(pool)

	embeddingService := service.NewEmbeddingService(embeddingClient, messageEmbeddingRepo)
	summaryService := service.NewSummaryService(messageRepo, summaryRepo, llmClient)
	contextBuilder := service.NewContextBuilder(summaryRepo, messageEmbeddingRepo, embeddingClient)

	quotaRepo := repository.NewQuotaRepository(pool)
	creditRepo := repository.NewCreditRepository(pool)
	quotaService := service.NewQuotaService(quotaRepo, creditRepo, personaRepo, userRepo, cfg.DefaultDailyMessageLimit)

	quotaResetJob := service.NewQuotaResetJob(quotaRepo, cfg.QuotaResetInterval)
	go quotaResetJob.Run(ctx)

	chatService := service.NewChatService(
		messageRepo, personaRepo, personaTraitsRepo, onboardingProfileRepo, llmClient,
		embeddingService, summaryService, contextBuilder, quotaService,
	)
	messageHandler := handler.NewMessageHandler(chatService)
	chatEnergyHandler := handler.NewChatEnergyHandler(quotaService)

	requestLogRepo := repository.NewRequestLogRepository(pool)

	adminUserRepo := repository.NewAdminUserRepository(pool)
	adminSessionRepo := repository.NewAdminSessionRepository(pool)
	adminAuditLogRepo := repository.NewAdminAuditLogRepository(pool)
	adminAuthService := service.NewAdminAuthService(adminUserRepo, adminSessionRepo, adminAuditLogRepo, cfg.AdminSessionTTL)
	adminAuthHandler := handler.NewAdminAuthHandler(adminAuthService)

	// The admin services below reuse the same repositories the public
	// API already constructed above (userRepo, quotaRepo, creditRepo,
	// personaRepo) — the admin dashboard's extra read/write methods
	// were added directly onto those repository types, not split into
	// separate ones, since they operate on the exact same tables.
	adminUserService := service.NewAdminUserService(userRepo, quotaRepo, creditRepo, adminAuditLogRepo)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)

	llmCredentialRepo := repository.NewLLMCredentialRepository(pool)
	llmModelRepo := repository.NewLLMModelRepository(pool)

	adminPersonaService := service.NewAdminPersonaService(personaRepo, llmModelRepo, llmCredentialRepo, adminAuditLogRepo)
	adminPersonaHandler := handler.NewAdminPersonaHandler(adminPersonaService)

	adminMetricsRepo := repository.NewAdminMetricsRepository(pool)
	adminDashboardService := service.NewAdminDashboardService(adminMetricsRepo)
	adminDashboardHandler := handler.NewAdminDashboardHandler(adminDashboardService)

	adminConversationService := service.NewAdminConversationService(conversationRepo, messageRepo, adminAuditLogRepo)
	adminConversationHandler := handler.NewAdminConversationHandler(adminConversationService)

	adminLogService := service.NewAdminLogService(requestLogRepo)
	adminLogHandler := handler.NewAdminLogHandler(adminLogService)

	adminActivityService := service.NewAdminActivityService(adminAuditLogRepo)
	adminActivityHandler := handler.NewAdminActivityHandler(adminActivityService)

	adminOnboardingInsightsRepo := repository.NewAdminOnboardingInsightsRepository(pool)
	adminOnboardingService := service.NewAdminOnboardingService(adminOnboardingInsightsRepo)
	adminOnboardingHandler := handler.NewAdminOnboardingHandler(adminOnboardingService)

	adminLLMService := service.NewAdminLLMService(llmCredentialRepo, llmModelRepo, llmTokenBox, adminAuditLogRepo)
	adminLLMHandler := handler.NewAdminLLMHandler(adminLLMService)

	r := chi.NewRouter()
	// requestlog.Middleware is mounted before Recoverer deliberately —
	// see its doc comment for why that's required for it to log
	// panicking (500) requests too, not just normal ones.
	r.Use(requestlog.Middleware(requestLogRepo, cfg.LogRedactSensitiveFields))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler.Health)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/personas", func(r chi.Router) {
		// Soft auth: GET / and GET /{id} stay public either way; List
		// itself only enforces auth for the ?recommend=true variant.
		r.Use(auth.OptionalMiddleware(accessTokenIssuer))
		r.Get("/", personaHandler.List)
		r.Get("/{id}", personaHandler.Get)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(accessTokenIssuer))
			r.Get("/{id}/traits", personaTraitsHandler.Get)
			r.Put("/{id}/traits", personaTraitsHandler.Set)
			r.Delete("/{id}/traits", personaTraitsHandler.Reset)
		})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/anonymous", authHandler.CreateAnonymousSession)
		r.Post("/email/request-code", authHandler.RequestEmailCode)
		r.Post("/email/verify-code", authHandler.VerifyEmailCode)
		r.Post("/refresh", authHandler.RefreshSession)
		r.Post("/logout", authHandler.Logout)

		r.With(auth.Middleware(accessTokenIssuer)).Post("/register/complete", authHandler.CompleteRegistration)
	})

	r.Route("/conversations", func(r chi.Router) {
		r.Use(auth.Middleware(accessTokenIssuer))
		r.Get("/", conversationHandler.List)
		r.Post("/", conversationHandler.Create)
		r.Get("/{id}/messages", conversationHandler.ListMessages)
		r.With(handler.QuotaMiddleware(conversationService, quotaService)).Post("/{id}/messages", messageHandler.Create)
		r.Post("/{id}/messages/{messageID}/reaction", conversationHandler.SetReaction)
		r.Delete("/{id}/messages/{messageID}/reaction", conversationHandler.ClearReaction)
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(auth.Middleware(accessTokenIssuer))
		r.Post("/onboarding-profile", onboardingHandler.SaveProfile)
		r.Get("/onboarding-profile", onboardingHandler.GetProfile)
		r.Get("/chat-energy", chatEnergyHandler.Get)
	})

	// Admin dashboard API — entirely separate auth from everything above
	// (see AdminAuthService). Only login/logout are public; every other
	// /admin/* route sits inside the AdminMiddleware-gated group below.
	r.Route("/admin", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", adminAuthHandler.Login)
			r.Post("/logout", adminAuthHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(handler.AdminMiddleware(adminAuthService))

			r.Get("/auth/me", adminAuthHandler.Me)

			r.Route("/users", func(r chi.Router) {
				r.Get("/", adminUserHandler.List)
				r.Get("/{id}", adminUserHandler.Get)
				r.Patch("/{id}/quota", adminUserHandler.SetQuota)
				r.Patch("/{id}/credits", adminUserHandler.SetCredits)
			})

			r.Route("/personas", func(r chi.Router) {
				r.Get("/", adminPersonaHandler.List)
				r.Get("/{id}", adminPersonaHandler.Get)
				r.Post("/", adminPersonaHandler.Create)
				r.Put("/{id}", adminPersonaHandler.Update)
			})

			r.Get("/dashboard/metrics", adminDashboardHandler.Metrics)

			r.Route("/conversations", func(r chi.Router) {
				r.Get("/", adminConversationHandler.List)
				r.Get("/{id}", adminConversationHandler.Get)
				r.Delete("/{id}/messages/{messageID}", adminConversationHandler.DeleteMessage)
			})

			r.Get("/logs", adminLogHandler.List)
			r.Get("/activity", adminActivityHandler.List)
			r.Get("/onboarding/insights", adminOnboardingHandler.Insights)

			r.Route("/llm", func(r chi.Router) {
				r.Route("/credentials", func(r chi.Router) {
					r.Get("/", adminLLMHandler.ListCredentials)
					r.Get("/{id}", adminLLMHandler.GetCredential)
					r.Post("/", adminLLMHandler.CreateCredential)
					r.Put("/{id}", adminLLMHandler.UpdateCredential)
					r.Delete("/{id}", adminLLMHandler.DeleteCredential)
				})
				r.Route("/models", func(r chi.Router) {
					r.Get("/", adminLLMHandler.ListModels)
					r.Post("/", adminLLMHandler.CreateModel)
					r.Put("/{id}", adminLLMHandler.UpdateModel)
					r.Delete("/{id}", adminLLMHandler.DeleteModel)
				})
			})
		})
	})

	addr := ":" + cfg.ServerPort
	log.Printf("perchly-backend listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/config"
	"perchly-backend/internal/email"
	"perchly-backend/internal/embedding"
	"perchly-backend/internal/handler"
	"perchly-backend/internal/llm"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/requestlog"
	"perchly-backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.JWTSecretIsEphemeral {
		log.Printf("warning: JWT_SECRET not set, using a random per-process secret — tokens won't survive a restart")
	}

	ctx := context.Background()
	pool, err := repository.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("create postgres pool: %v", err)
	}
	defer pool.Close()

	healthService := service.NewHealthService(pool)
	healthHandler := handler.NewHealthHandler(healthService)

	personaRepo := repository.NewPersonaRepository(pool)
	personaService := service.NewPersonaService(personaRepo)
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
	onboardingProfileRepo := repository.NewOnboardingProfileRepository(pool)
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
	quotaService := service.NewQuotaService(quotaRepo, creditRepo, cfg.DefaultDailyMessageLimit)

	quotaResetJob := service.NewQuotaResetJob(quotaRepo, cfg.QuotaResetInterval)
	go quotaResetJob.Run(ctx)

	chatService := service.NewChatService(
		messageRepo, personaRepo, personaTraitsRepo, llmClient,
		embeddingService, summaryService, contextBuilder, quotaService,
	)
	messageHandler := handler.NewMessageHandler(chatService)

	requestLogRepo := repository.NewRequestLogRepository(pool)

	r := chi.NewRouter()
	// requestlog.Middleware is mounted before Recoverer deliberately —
	// see its doc comment for why that's required for it to log
	// panicking (500) requests too, not just normal ones.
	r.Use(requestlog.Middleware(requestLogRepo, cfg.LogRedactSensitiveFields))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler.Health)

	r.Route("/personas", func(r chi.Router) {
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
	})

	addr := ":" + cfg.ServerPort
	log.Printf("perchly-backend listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

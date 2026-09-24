package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the service.
type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	LLMProvider string
	LLMAPIKey   string
	LLMModel    string
	LLMBaseURL  string

	EmbeddingProvider string
	EmbeddingAPIKey   string
	EmbeddingModel    string
	EmbeddingBaseURL  string

	// JWTSecret signs access tokens. Empty means Load generated a random
	// one for this process only — fine for local dev, but every restart
	// invalidates existing tokens, so set a stable JWT_SECRET anywhere
	// tokens need to survive a restart.
	JWTSecret            string
	JWTSecretIsEphemeral bool
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration

	// AdminSessionTTL is how long an admin dashboard login lasts before
	// requiring another login — there's no refresh flow for admin
	// sessions (unlike user access/refresh tokens), just a longer-lived
	// opaque token, since admin accounts are few and it's fine to just
	// log in again after it lapses.
	AdminSessionTTL time.Duration

	// SMTP* configure the verification-code email sender. If
	// SMTPUsername or SMTPPassword is empty, the server falls back to
	// logging codes to stdout instead of sending real email (see
	// internal/email.ConsoleSender) — fine for local dev, not for
	// production.
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFromName string

	DefaultDailyMessageLimit int
	QuotaResetInterval       time.Duration

	// LogRedactSensitiveFields controls internal/requestlog.Middleware:
	// when true (the default), known-sensitive field values (code,
	// password, token, content, ...) in logged request bodies/query
	// params are replaced with "[REDACTED]". Flip it off only for a
	// short, deliberate debugging session — never in a shared or
	// production environment.
	LogRedactSensitiveFields bool
}

// Load reads configuration from a .env file (if present) and the process
// environment, and returns the resulting Config. Environment variables take
// precedence over values in .env.
func Load() (*Config, error) {
	// Missing .env is not an error: in production, env vars are usually
	// injected by the platform instead of a file.
	_ = godotenv.Load()

	dailyLimit, err := strconv.Atoi(getEnv("DAILY_MESSAGE_LIMIT", "20"))
	if err != nil {
		return nil, fmt.Errorf("parse DAILY_MESSAGE_LIMIT: %w", err)
	}

	resetInterval, err := time.ParseDuration(getEnv("QUOTA_RESET_INTERVAL", "1m"))
	if err != nil {
		return nil, fmt.Errorf("parse QUOTA_RESET_INTERVAL: %w", err)
	}

	accessTokenTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "20m"))
	if err != nil {
		return nil, fmt.Errorf("parse ACCESS_TOKEN_TTL: %w", err)
	}

	refreshTokenTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "1440h")) // 60 days
	if err != nil {
		return nil, fmt.Errorf("parse REFRESH_TOKEN_TTL: %w", err)
	}

	adminSessionTTL, err := time.ParseDuration(getEnv("ADMIN_SESSION_TTL", "168h")) // 7 days
	if err != nil {
		return nil, fmt.Errorf("parse ADMIN_SESSION_TTL: %w", err)
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("parse SMTP_PORT: %w", err)
	}

	logRedact, err := strconv.ParseBool(getEnv("LOG_REDACT_SENSITIVE_FIELDS", "true"))
	if err != nil {
		return nil, fmt.Errorf("parse LOG_REDACT_SENSITIVE_FIELDS: %w", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	ephemeral := jwtSecret == ""
	if ephemeral {
		secret, err := randomSecret(32)
		if err != nil {
			return nil, fmt.Errorf("generate ephemeral JWT secret: %w", err)
		}
		jwtSecret = secret
	}

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "perchly"),
		DBPassword: getEnv("DB_PASSWORD", "perchly"),
		DBName:     getEnv("DB_NAME", "perchly"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		LLMProvider: getEnv("LLM_PROVIDER", "anthropic"),
		LLMAPIKey:   getEnv("LLM_API_KEY", getEnv("ANTHROPIC_API_KEY", getEnv("OPENAI_API_KEY", getEnv("HF_TOKEN", getEnv("GEMINI_API_KEY", ""))))),
		LLMModel:    getEnv("LLM_MODEL", "claude-sonnet-5"),
		LLMBaseURL:  getEnv("LLM_BASE_URL", ""),

		EmbeddingProvider: getEnv("EMBEDDING_PROVIDER", "openai"),
		EmbeddingAPIKey:   getEnv("EMBEDDING_API_KEY", getEnv("OPENAI_API_KEY", "")),
		EmbeddingModel:    getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingBaseURL:  getEnv("EMBEDDING_BASE_URL", ""),

		JWTSecret:            jwtSecret,
		JWTSecretIsEphemeral: ephemeral,
		AccessTokenTTL:       accessTokenTTL,
		RefreshTokenTTL:      refreshTokenTTL,
		AdminSessionTTL:      adminSessionTTL,

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: normalizeSMTPPassword(getEnv("SMTP_PASSWORD", "")),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "Perchly"),

		DefaultDailyMessageLimit: dailyLimit,
		QuotaResetInterval:       resetInterval,

		LogRedactSensitiveFields: logRedact,
	}

	return cfg, nil
}

// DSN builds a PostgreSQL connection string suitable for pgx.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// normalizeSMTPPassword accepts Google's display form of an App Password
// ("xxxx xxxx xxxx xxxx", often copied with NBSP) and returns the 16-char
// secret Gmail actually authenticates with.
func normalizeSMTPPassword(raw string) string {
	s := strings.TrimSpace(raw)
	if len(s) >= 2 {
		if q := s[0]; (q == '"' || q == '\'') && s[len(s)-1] == q {
			s = s[1 : len(s)-1]
		}
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

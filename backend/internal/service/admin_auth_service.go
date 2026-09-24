package service

import (
	"context"
	"errors"
	"time"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

var (
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrInvalidAdminSession     = errors.New("invalid admin session")
)

type AdminUserRepo interface {
	GetByEmail(ctx context.Context, email string) (model.AdminUser, error)
	GetByID(ctx context.Context, id string) (model.AdminUser, error)
}

type AdminSessionRepo interface {
	Create(ctx context.Context, adminUserID, tokenHash string, expiresAt time.Time) (model.AdminSession, error)
	GetByHash(ctx context.Context, tokenHash string) (model.AdminSession, error)
	Revoke(ctx context.Context, id string) error
}

type AdminAuditLogRepo interface {
	Create(ctx context.Context, adminUserID, action, targetType string, targetID *string, detail map[string]any) error
}

// AdminAuthService issues and verifies admin dashboard sessions. It
// deliberately doesn't share anything with AuthService (public user
// auth) beyond the opaque-token scheme in internal/auth — admin
// accounts, sessions, and audit trail are their own tables.
type AdminAuthService struct {
	adminUsers AdminUserRepo
	sessions   AdminSessionRepo
	auditLog   AdminAuditLogRepo
	sessionTTL time.Duration
}

func NewAdminAuthService(adminUsers AdminUserRepo, sessions AdminSessionRepo, auditLog AdminAuditLogRepo, sessionTTL time.Duration) *AdminAuthService {
	return &AdminAuthService{adminUsers: adminUsers, sessions: sessions, auditLog: auditLog, sessionTTL: sessionTTL}
}

// Login verifies email/password and, on success, issues a new session
// token (returned raw — only its hash is stored) and records an
// admin.login audit entry. expiresAt is the session's expiry, so a
// caller (the admin/ web app) can set its own cookie to match rather
// than guessing sessionTTL independently.
func (s *AdminAuthService) Login(ctx context.Context, email, password string) (rawToken string, admin model.AdminUser, expiresAt time.Time, err error) {
	admin, err = s.adminUsers.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			return "", model.AdminUser{}, time.Time{}, ErrInvalidAdminCredentials
		}
		return "", model.AdminUser{}, time.Time{}, err
	}

	if !auth.VerifyAdminPassword(admin.PasswordHash, password) {
		return "", model.AdminUser{}, time.Time{}, ErrInvalidAdminCredentials
	}

	// Reuses the same opaque-token generation as user refresh tokens
	// (internal/auth/refresh_token.go) — the scheme is identical, only
	// which table the hash is stored in differs.
	rawToken, tokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", model.AdminUser{}, time.Time{}, err
	}

	session, err := s.sessions.Create(ctx, admin.ID, tokenHash, time.Now().Add(s.sessionTTL))
	if err != nil {
		return "", model.AdminUser{}, time.Time{}, err
	}

	_ = s.auditLog.Create(ctx, admin.ID, "admin.login", "admin_user", &admin.ID, nil)

	return rawToken, admin, session.ExpiresAt, nil
}

// VerifySession returns the admin who owns rawToken, if it's a live,
// unrevoked, unexpired session.
func (s *AdminAuthService) VerifySession(ctx context.Context, rawToken string) (model.AdminUser, error) {
	session, err := s.sessions.GetByHash(ctx, auth.HashRefreshToken(rawToken))
	if err != nil {
		if errors.Is(err, repository.ErrAdminSessionNotFound) {
			return model.AdminUser{}, ErrInvalidAdminSession
		}
		return model.AdminUser{}, err
	}
	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return model.AdminUser{}, ErrInvalidAdminSession
	}

	admin, err := s.adminUsers.GetByID(ctx, session.AdminUserID)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			return model.AdminUser{}, ErrInvalidAdminSession
		}
		return model.AdminUser{}, err
	}
	return admin, nil
}

// CurrentAdmin returns the admin dashboard account by id — used by GET
// /admin/auth/me to render the logged-in admin's email after
// AdminMiddleware has already verified the session token (and resolved
// the same row once, but discarded everything but the id — see that
// middleware's doc).
func (s *AdminAuthService) CurrentAdmin(ctx context.Context, adminUserID string) (model.AdminUser, error) {
	return s.adminUsers.GetByID(ctx, adminUserID)
}

// Logout revokes the session rawToken belongs to. A token that doesn't
// match any session (already logged out, or never valid) is not an
// error — logout is idempotent from the caller's point of view.
func (s *AdminAuthService) Logout(ctx context.Context, rawToken string) error {
	session, err := s.sessions.GetByHash(ctx, auth.HashRefreshToken(rawToken))
	if err != nil {
		if errors.Is(err, repository.ErrAdminSessionNotFound) {
			return nil
		}
		return err
	}
	return s.sessions.Revoke(ctx, session.ID)
}

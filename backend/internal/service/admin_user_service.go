package service

import (
	"context"
	"errors"

	"perchly-backend/internal/model"
)

var (
	ErrInvalidDailyLimit   = errors.New("daily limit must be zero or greater")
	ErrInvalidCreditAmount = errors.New("credit amount must be zero or greater")
)

// AdminAppUserRepo is the app-user (not admin-account) repository this
// service reads/writes — named to avoid colliding with
// AdminAuthService's own AdminUserRepo, which is about admin_users
// (dashboard accounts), a completely different table.
type AdminAppUserRepo interface {
	GetByID(ctx context.Context, id string) (model.User, error)
	ListForAdmin(ctx context.Context, search string, limit, offset int) ([]model.User, int, error)
}

type AdminUserQuotaRepo interface {
	ListForUser(ctx context.Context, userID string) ([]model.UserQuota, error)
	SetDailyLimit(ctx context.Context, userID, personaID string, dailyLimit int) (model.UserQuota, error)
}

type AdminUserCreditRepo interface {
	GetBalance(ctx context.Context, userID string) (int, error)
	SetBalance(ctx context.Context, userID string, amount int) (int, error)
}

// AdminUserDetail is a user plus everything the admin dashboard's user
// detail page shows about them — assembled here rather than left to
// three separate frontend calls, since all three always render together.
type AdminUserDetail struct {
	User    model.User        `json:"user"`
	Quotas  []model.UserQuota `json:"quotas"`
	Credits int               `json:"credits"`
}

// AdminUserService is the admin dashboard's view onto app users —
// listing/inspecting them and overriding their quota/credits. Every
// mutation is written to admin_audit_log, since these change another
// user's account state directly.
type AdminUserService struct {
	users    AdminAppUserRepo
	quotas   AdminUserQuotaRepo
	credits  AdminUserCreditRepo
	auditLog AdminAuditLogRepo
}

func NewAdminUserService(users AdminAppUserRepo, quotas AdminUserQuotaRepo, credits AdminUserCreditRepo, auditLog AdminAuditLogRepo) *AdminUserService {
	return &AdminUserService{users: users, quotas: quotas, credits: credits, auditLog: auditLog}
}

func (s *AdminUserService) List(ctx context.Context, search string, limit, offset int) ([]model.User, int, error) {
	return s.users.ListForAdmin(ctx, search, limit, offset)
}

func (s *AdminUserService) Get(ctx context.Context, userID string) (AdminUserDetail, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return AdminUserDetail{}, err
	}

	quotas, err := s.quotas.ListForUser(ctx, userID)
	if err != nil {
		return AdminUserDetail{}, err
	}

	credits, err := s.credits.GetBalance(ctx, userID)
	if err != nil {
		return AdminUserDetail{}, err
	}

	return AdminUserDetail{User: user, Quotas: quotas, Credits: credits}, nil
}

// SetQuota overrides userID's daily message limit for personaID — see
// UserQuota.DailyLimit's doc comment, which is exactly this operation.
func (s *AdminUserService) SetQuota(ctx context.Context, adminUserID, userID, personaID string, dailyLimit int) (model.UserQuota, error) {
	if dailyLimit < 0 {
		return model.UserQuota{}, ErrInvalidDailyLimit
	}

	quota, err := s.quotas.SetDailyLimit(ctx, userID, personaID, dailyLimit)
	if err != nil {
		return model.UserQuota{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.user.set_quota", "user", &userID, map[string]any{
		"persona_id":  personaID,
		"daily_limit": dailyLimit,
	})

	return quota, nil
}

// SetCredits overrides userID's credit balance to amount (not a
// delta — the admin dashboard always shows and submits the absolute
// number, same as SetQuota).
func (s *AdminUserService) SetCredits(ctx context.Context, adminUserID, userID string, amount int) (int, error) {
	if amount < 0 {
		return 0, ErrInvalidCreditAmount
	}

	balance, err := s.credits.SetBalance(ctx, userID, amount)
	if err != nil {
		return 0, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.user.set_credits", "user", &userID, map[string]any{
		"amount": amount,
	})

	return balance, nil
}

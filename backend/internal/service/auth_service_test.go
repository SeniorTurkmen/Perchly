package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

// ---- in-memory fakes ----
//
// AuthService's real dependencies are thin Postgres wrappers with no
// interesting logic of their own (see internal/repository) — these fakes
// reimplement just enough of their contract in memory so AuthService's
// actual branching logic (Durum A/B, rate limits, rotation, attempt
// caps) can be tested without a database.

type fakeUserRepo struct {
	mu    sync.Mutex
	users map[string]model.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]model.User)}
}

func (r *fakeUserRepo) GetByID(_ context.Context, id string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return model.User{}, repository.ErrUserNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) GetByDeviceID(_ context.Context, deviceID string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.DeviceID != nil && *u.DeviceID == deviceID {
			return u, nil
		}
	}
	return model.User{}, repository.ErrUserNotFound
}

func (r *fakeUserRepo) GetVerifiedByEmail(_ context.Context, email string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email != nil && *u.Email == email && u.EmailVerifiedAt != nil {
			return u, nil
		}
	}
	return model.User{}, repository.ErrUserNotFound
}

func (r *fakeUserRepo) CreateAnonymous(_ context.Context, deviceID, timezone string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if timezone == "" {
		timezone = "UTC"
	}
	u := model.User{ID: uuid.NewString(), IsAnonymous: true, Timezone: timezone, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if deviceID != "" {
		u.DeviceID = &deviceID
	}
	r.users[u.ID] = u
	return u, nil
}

func (r *fakeUserRepo) CreateVerifiedEmailUser(_ context.Context, email, timezone string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if timezone == "" {
		timezone = "UTC"
	}
	now := time.Now()
	u := model.User{ID: uuid.NewString(), Email: &email, EmailVerifiedAt: &now, IsAnonymous: false, Timezone: timezone, CreatedAt: now, UpdatedAt: now}
	r.users[u.ID] = u
	return u, nil
}

func (r *fakeUserRepo) UpgradeAnonymousToVerifiedEmail(_ context.Context, userID, email string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return model.User{}, repository.ErrUserNotFound
	}
	now := time.Now()
	u.Email = &email
	u.EmailVerifiedAt = &now
	u.IsAnonymous = false
	r.users[userID] = u
	return u, nil
}

func (r *fakeUserRepo) UpdateDisplayName(_ context.Context, userID, displayName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return repository.ErrUserNotFound
	}
	u.DisplayName = &displayName
	r.users[userID] = u
	return nil
}

type fakeVerificationCodeRepo struct {
	mu      sync.Mutex
	codes   map[string]model.EmailVerificationCode
	byEmail map[string][]string // creation order, oldest first
}

func newFakeVerificationCodeRepo() *fakeVerificationCodeRepo {
	return &fakeVerificationCodeRepo{codes: make(map[string]model.EmailVerificationCode), byEmail: make(map[string][]string)}
}

func (r *fakeVerificationCodeRepo) CountSince(_ context.Context, email string, since time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, id := range r.byEmail[email] {
		if r.codes[id].CreatedAt.After(since) {
			count++
		}
	}
	return count, nil
}

func (r *fakeVerificationCodeRepo) Create(_ context.Context, email, code string, expiresAt time.Time) (model.EmailVerificationCode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := model.EmailVerificationCode{ID: uuid.NewString(), Email: email, Code: code, ExpiresAt: expiresAt, CreatedAt: time.Now()}
	r.codes[c.ID] = c
	r.byEmail[email] = append(r.byEmail[email], c.ID)
	return c, nil
}

func (r *fakeVerificationCodeRepo) GetLatest(_ context.Context, email string) (model.EmailVerificationCode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := r.byEmail[email]
	if len(ids) == 0 {
		return model.EmailVerificationCode{}, repository.ErrVerificationCodeNotFound
	}
	return r.codes[ids[len(ids)-1]], nil
}

func (r *fakeVerificationCodeRepo) IncrementAttempt(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.codes[id]
	if !ok {
		return errors.New("code not found")
	}
	c.AttemptCount++
	r.codes[id] = c
	return nil
}

func (r *fakeVerificationCodeRepo) MarkUsed(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.codes[id]
	if !ok {
		return errors.New("code not found")
	}
	now := time.Now()
	c.UsedAt = &now
	r.codes[id] = c
	return nil
}

type fakeRefreshTokenRepo struct {
	mu     sync.Mutex
	tokens map[string]model.RefreshToken // by id
	byHash map[string]string             // hash -> id
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{tokens: make(map[string]model.RefreshToken), byHash: make(map[string]string)}
}

func (r *fakeRefreshTokenRepo) Create(_ context.Context, userID, tokenHash string, expiresAt time.Time) (model.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := model.RefreshToken{ID: uuid.NewString(), UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, CreatedAt: time.Now()}
	r.tokens[t.ID] = t
	r.byHash[tokenHash] = t.ID
	return t, nil
}

func (r *fakeRefreshTokenRepo) GetByHash(_ context.Context, tokenHash string) (model.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byHash[tokenHash]
	if !ok {
		return model.RefreshToken{}, repository.ErrRefreshTokenNotFound
	}
	return r.tokens[id], nil
}

func (r *fakeRefreshTokenRepo) Revoke(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[id]
	if !ok {
		return nil
	}
	now := time.Now()
	t.RevokedAt = &now
	r.tokens[id] = t
	return nil
}

func (r *fakeRefreshTokenRepo) RevokeAllForUser(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for id, t := range r.tokens {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &now
			r.tokens[id] = t
		}
	}
	return nil
}

// expire is a test-only helper reaching past the Sender/Repo interfaces
// to backdate a stored refresh token's expiry, so expiry can be tested
// without actually waiting out a TTL.
func (r *fakeRefreshTokenRepo) expire(rawToken string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	hash := auth.HashRefreshToken(rawToken)
	id := r.byHash[hash]
	t := r.tokens[id]
	t.ExpiresAt = time.Now().Add(-time.Minute)
	r.tokens[id] = t
}

type fakeEmailSender struct {
	mu   sync.Mutex
	sent []struct{ To, Code string }
}

func (s *fakeEmailSender) SendVerificationCode(_ context.Context, toEmail, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, struct{ To, Code string }{toEmail, code})
	return nil
}

func (s *fakeEmailSender) lastCode() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sent) == 0 {
		return ""
	}
	return s.sent[len(s.sent)-1].Code
}

func (s *fakeEmailSender) sentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

// fakeOnboardingProfileRepoForAuth is a minimal in-memory stand-in — only
// Exists is used by AuthService, so that's all this tracks.
type fakeOnboardingProfileRepoForAuth struct {
	mu        sync.Mutex
	completed map[string]bool
}

func newFakeOnboardingProfileRepoForAuth() *fakeOnboardingProfileRepoForAuth {
	return &fakeOnboardingProfileRepoForAuth{completed: make(map[string]bool)}
}

func (r *fakeOnboardingProfileRepoForAuth) Upsert(_ context.Context, p model.OnboardingProfile) (model.OnboardingProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed[p.UserID] = true
	return p, nil
}

func (r *fakeOnboardingProfileRepoForAuth) GetByUserID(context.Context, string) (model.OnboardingProfile, error) {
	return model.OnboardingProfile{}, repository.ErrOnboardingProfileNotFound
}

func (r *fakeOnboardingProfileRepoForAuth) Exists(_ context.Context, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.completed[userID], nil
}

// ---- test harness ----

type authServiceHarness struct {
	service     *AuthService
	users       *fakeUserRepo
	codes       *fakeVerificationCodeRepo
	refreshes   *fakeRefreshTokenRepo
	onboarding  *fakeOnboardingProfileRepoForAuth
	sender      *fakeEmailSender
	tokenIssuer *auth.AccessTokenIssuer
}

func newAuthServiceHarness() *authServiceHarness {
	users := newFakeUserRepo()
	codes := newFakeVerificationCodeRepo()
	refreshes := newFakeRefreshTokenRepo()
	onboarding := newFakeOnboardingProfileRepoForAuth()
	sender := &fakeEmailSender{}
	tokenIssuer := auth.NewAccessTokenIssuer("test-secret", 20*time.Minute)

	return &authServiceHarness{
		service:     NewAuthService(users, codes, refreshes, onboarding, tokenIssuer, sender, 60*24*time.Hour),
		users:       users,
		codes:       codes,
		refreshes:   refreshes,
		onboarding:  onboarding,
		sender:      sender,
		tokenIssuer: tokenIssuer,
	}
}

func (h *authServiceHarness) userIDFromAccessToken(t *testing.T, token string) string {
	t.Helper()
	claims, err := h.tokenIssuer.Verify(token)
	if err != nil {
		t.Fatalf("verify access token: %v", err)
	}
	return claims.UserID
}

// ---- tests ----

func TestAuthService_CreateAnonymousSession_ReturningDeviceReusesUser(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	_, user1, err := h.service.CreateAnonymousSession(ctx, "device-1", "Europe/Istanbul")
	if err != nil {
		t.Fatalf("first CreateAnonymousSession: %v", err)
	}
	if !user1.IsAnonymous {
		t.Fatalf("expected new user to be anonymous")
	}

	_, user2, err := h.service.CreateAnonymousSession(ctx, "device-1", "")
	if err != nil {
		t.Fatalf("second CreateAnonymousSession: %v", err)
	}
	if user2.ID != user1.ID {
		t.Fatalf("expected same user for same device_id, got %s and %s", user1.ID, user2.ID)
	}

	_, user3, err := h.service.CreateAnonymousSession(ctx, "device-2", "")
	if err != nil {
		t.Fatalf("third CreateAnonymousSession: %v", err)
	}
	if user3.ID == user1.ID {
		t.Fatalf("expected a different device_id to get a different user")
	}
}

func TestAuthService_VerifyEmailCode_DurumA_ExistingVerifiedAccountWins(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	existing, err := h.users.CreateVerifiedEmailUser(ctx, "a@example.com", "UTC")
	if err != nil {
		t.Fatalf("seed existing verified user: %v", err)
	}

	anonTokens, anonUser, err := h.service.CreateAnonymousSession(ctx, "device-a", "")
	if err != nil {
		t.Fatalf("create anonymous session: %v", err)
	}

	if err := h.service.RequestEmailCode(ctx, "a@example.com"); err != nil {
		t.Fatalf("RequestEmailCode: %v", err)
	}
	code := h.sender.lastCode()

	tokens, isNew, err := h.service.VerifyEmailCode(ctx, "a@example.com", code, anonTokens.AccessToken)
	if err != nil {
		t.Fatalf("VerifyEmailCode: %v", err)
	}
	if isNew {
		t.Errorf("Durum A should not report is_new_registration")
	}

	gotUserID := h.userIDFromAccessToken(t, tokens.AccessToken)
	if gotUserID != existing.ID {
		t.Errorf("session belongs to %s, want the existing verified user %s", gotUserID, existing.ID)
	}

	// The anonymous session's data must be left completely alone — never
	// merged, never deleted, never linked to the account just logged into.
	untouched, err := h.users.GetByID(ctx, anonUser.ID)
	if err != nil {
		t.Fatalf("load anonymous user: %v", err)
	}
	if !untouched.IsAnonymous || untouched.Email != nil {
		t.Errorf("anonymous user should be untouched, got is_anonymous=%v email=%v", untouched.IsAnonymous, untouched.Email)
	}
}

func TestAuthService_VerifyEmailCode_DurumB_WithAnonymousToken_UpgradesInPlace(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	anonTokens, anonUser, err := h.service.CreateAnonymousSession(ctx, "device-b", "")
	if err != nil {
		t.Fatalf("create anonymous session: %v", err)
	}

	if err := h.service.RequestEmailCode(ctx, "b@example.com"); err != nil {
		t.Fatalf("RequestEmailCode: %v", err)
	}
	code := h.sender.lastCode()

	tokens, isNew, err := h.service.VerifyEmailCode(ctx, "b@example.com", code, anonTokens.AccessToken)
	if err != nil {
		t.Fatalf("VerifyEmailCode: %v", err)
	}
	if !isNew {
		t.Errorf("Durum B should report is_new_registration")
	}

	gotUserID := h.userIDFromAccessToken(t, tokens.AccessToken)
	if gotUserID != anonUser.ID {
		t.Errorf("expected the SAME user id to be upgraded in place, got %s want %s", gotUserID, anonUser.ID)
	}

	upgraded, err := h.users.GetByID(ctx, anonUser.ID)
	if err != nil {
		t.Fatalf("load upgraded user: %v", err)
	}
	if upgraded.IsAnonymous {
		t.Errorf("expected user to no longer be anonymous")
	}
	if upgraded.Email == nil || *upgraded.Email != "b@example.com" {
		t.Errorf("expected email to be set to b@example.com, got %v", upgraded.Email)
	}
	if upgraded.EmailVerifiedAt == nil {
		t.Errorf("expected email_verified_at to be set")
	}
}

func TestAuthService_VerifyEmailCode_DurumB_WithoutAnonymousToken_CreatesNewUser(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	if err := h.service.RequestEmailCode(ctx, "c@example.com"); err != nil {
		t.Fatalf("RequestEmailCode: %v", err)
	}
	code := h.sender.lastCode()

	tokens, isNew, err := h.service.VerifyEmailCode(ctx, "c@example.com", code, "")
	if err != nil {
		t.Fatalf("VerifyEmailCode: %v", err)
	}
	if !isNew {
		t.Errorf("expected is_new_registration for a fresh account")
	}

	userID := h.userIDFromAccessToken(t, tokens.AccessToken)
	user, err := h.users.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("load new user: %v", err)
	}
	if user.IsAnonymous || user.Email == nil || *user.Email != "c@example.com" {
		t.Errorf("expected a new verified user for c@example.com, got %+v", user)
	}
}

func TestAuthService_VerifyEmailCode_ExpiredCodeRejected(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	if _, err := h.codes.Create(ctx, "d@example.com", "123456", time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("seed expired code: %v", err)
	}

	_, _, err := h.service.VerifyEmailCode(ctx, "d@example.com", "123456", "")
	if !errors.Is(err, ErrCodeExpired) {
		t.Fatalf("VerifyEmailCode error = %v, want ErrCodeExpired", err)
	}
}

func TestAuthService_VerifyEmailCode_AttemptLimitExceeded(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	if _, err := h.codes.Create(ctx, "e@example.com", "999999", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatalf("seed code: %v", err)
	}

	for i := 0; i < verificationCodeMaxAttempts; i++ {
		_, _, err := h.service.VerifyEmailCode(ctx, "e@example.com", "000000", "")
		if !errors.Is(err, ErrInvalidCode) {
			t.Fatalf("attempt %d: error = %v, want ErrInvalidCode", i+1, err)
		}
	}

	// Attempt count has now hit the cap; even the CORRECT code must be
	// rejected until a new code is requested.
	_, _, err := h.service.VerifyEmailCode(ctx, "e@example.com", "999999", "")
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("final attempt error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_RequestEmailCode_RateLimitsSilently(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	if err := h.service.RequestEmailCode(ctx, "f@example.com"); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if got := h.sender.sentCount(); got != 1 {
		t.Fatalf("sent count = %d, want 1", got)
	}

	// Second request within the same minute must be silently dropped —
	// no error, but no second email either.
	if err := h.service.RequestEmailCode(ctx, "f@example.com"); err != nil {
		t.Fatalf("second (rate-limited) request returned an error, want nil: %v", err)
	}
	if got := h.sender.sentCount(); got != 1 {
		t.Fatalf("sent count after rate-limited request = %d, want still 1", got)
	}
}

func TestAuthService_RequestEmailCode_InvalidEmailFormat(t *testing.T) {
	h := newAuthServiceHarness()
	err := h.service.RequestEmailCode(context.Background(), "not-an-email")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("error = %v, want ErrInvalidEmail", err)
	}
}

func TestAuthService_IssueSession_ReflectsOnboardingCompletion(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	initial, user, err := h.service.CreateAnonymousSession(ctx, "device-onboarding", "")
	if err != nil {
		t.Fatalf("create anonymous session: %v", err)
	}
	if initial.HasCompletedOnboarding {
		t.Errorf("HasCompletedOnboarding = true before onboarding was ever saved, want false")
	}

	// Simulate OnboardingService.SaveProfile having been called for this user.
	if _, err := h.onboarding.Upsert(ctx, model.OnboardingProfile{UserID: user.ID}); err != nil {
		t.Fatalf("seed onboarding profile: %v", err)
	}

	refreshed, err := h.service.RefreshSession(ctx, initial.RefreshToken)
	if err != nil {
		t.Fatalf("refresh session: %v", err)
	}
	if !refreshed.HasCompletedOnboarding {
		t.Errorf("HasCompletedOnboarding = false after onboarding was saved and the session refreshed, want true")
	}
}

func TestAuthService_RefreshSession_RotatesAndDetectsReuse(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	initial, _, err := h.service.CreateAnonymousSession(ctx, "device-r", "")
	if err != nil {
		t.Fatalf("create anonymous session: %v", err)
	}

	rotated, err := h.service.RefreshSession(ctx, initial.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if rotated.RefreshToken == initial.RefreshToken {
		t.Fatalf("expected rotation to issue a new refresh token")
	}

	// Reusing the now-rotated-away token must fail...
	if _, err := h.service.RefreshSession(ctx, initial.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("reuse error = %v, want ErrInvalidRefreshToken", err)
	}

	// ...and that reuse must have revoked EVERY token for the user as a
	// security response, including the one just issued by the rotation
	// above.
	if _, err := h.service.RefreshSession(ctx, rotated.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("post-reuse-detection refresh error = %v, want ErrInvalidRefreshToken (all sessions should be revoked)", err)
	}
}

func TestAuthService_RefreshSession_ExpiredTokenRejected(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	tokens, _, err := h.service.CreateAnonymousSession(ctx, "device-x", "")
	if err != nil {
		t.Fatalf("create anonymous session: %v", err)
	}

	h.refreshes.expire(tokens.RefreshToken)

	if _, err := h.service.RefreshSession(ctx, tokens.RefreshToken); !errors.Is(err, ErrRefreshTokenExpired) {
		t.Fatalf("error = %v, want ErrRefreshTokenExpired", err)
	}
}

func TestAuthService_CompleteRegistration(t *testing.T) {
	h := newAuthServiceHarness()
	ctx := context.Background()

	_, user, err := h.service.CreateAnonymousSession(ctx, "device-p", "")
	if err != nil {
		t.Fatalf("create anonymous session: %v", err)
	}

	if err := h.service.CompleteRegistration(ctx, user.ID, "Ada"); err != nil {
		t.Fatalf("CompleteRegistration: %v", err)
	}

	updated, err := h.users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("load user: %v", err)
	}
	if updated.DisplayName == nil || *updated.DisplayName != "Ada" {
		t.Errorf("display_name = %v, want Ada", updated.DisplayName)
	}

	if err := h.service.CompleteRegistration(ctx, user.ID, "   "); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("blank display name error = %v, want ErrInvalidDisplayName", err)
	}
}

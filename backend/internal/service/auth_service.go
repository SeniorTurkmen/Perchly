package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/email"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

const (
	verificationCodeTTL         = 10 * time.Minute
	verificationCodeMaxAttempts = 5

	requestCodeLimitPerMinute = 1
	requestCodeLimitPerHour   = 5
)

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidCode         = errors.New("invalid code")
	ErrCodeExpired         = errors.New("code expired")
	ErrTooManyAttempts     = errors.New("too many attempts")
	ErrInvalidDisplayName  = errors.New("invalid display name")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

type AuthUserRepo interface {
	GetByID(ctx context.Context, id string) (model.User, error)
	GetByDeviceID(ctx context.Context, deviceID string) (model.User, error)
	GetVerifiedByEmail(ctx context.Context, email string) (model.User, error)
	CreateAnonymous(ctx context.Context, deviceID, timezone string) (model.User, error)
	CreateVerifiedEmailUser(ctx context.Context, email, timezone string) (model.User, error)
	UpgradeAnonymousToVerifiedEmail(ctx context.Context, userID, email string) (model.User, error)
	UpdateDisplayName(ctx context.Context, userID, displayName string) error
}

type VerificationCodeRepo interface {
	CountSince(ctx context.Context, email string, since time.Time) (int, error)
	Create(ctx context.Context, email, code string, expiresAt time.Time) (model.EmailVerificationCode, error)
	GetLatest(ctx context.Context, email string) (model.EmailVerificationCode, error)
	IncrementAttempt(ctx context.Context, id string) error
	MarkUsed(ctx context.Context, id string) error
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (model.RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// SessionTokens is the access/refresh pair returned by every AuthService
// method that starts or renews a session.
type SessionTokens struct {
	AccessToken  string
	RefreshToken string
	// HasCompletedOnboarding is filled in on every session issuance (see
	// issueSession) so a client can tell whether to show onboarding
	// without a separate round trip — in particular so it self-corrects
	// after a reinstall, a new device, or an email being linked, where
	// local client state saying "onboarding done" may have been lost
	// even though the server already has an onboarding_profiles row for
	// this user id.
	HasCompletedOnboarding bool
}

// AuthService is the whole auth story: anonymous device sessions, email
// OTP sign-in (with the anonymous-session upgrade rules below), profile
// completion, and refresh token rotation.
type AuthService struct {
	users              AuthUserRepo
	codes              VerificationCodeRepo
	refreshTokens      RefreshTokenRepo
	onboardingProfiles OnboardingProfileRepo
	accessTokens       *auth.AccessTokenIssuer
	emailSender        email.Sender
	refreshTTL         time.Duration
}

func NewAuthService(
	users AuthUserRepo,
	codes VerificationCodeRepo,
	refreshTokens RefreshTokenRepo,
	onboardingProfiles OnboardingProfileRepo,
	accessTokens *auth.AccessTokenIssuer,
	emailSender email.Sender,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:              users,
		codes:              codes,
		refreshTokens:      refreshTokens,
		onboardingProfiles: onboardingProfiles,
		accessTokens:       accessTokens,
		emailSender:        emailSender,
		refreshTTL:         refreshTTL,
	}
}

// CreateAnonymousSession returns the existing user's session if deviceID
// already has one, or creates a brand-new anonymous user otherwise.
func (s *AuthService) CreateAnonymousSession(ctx context.Context, deviceID, timezone string) (SessionTokens, model.User, error) {
	user, err := s.resolveAnonymousUser(ctx, deviceID, timezone)
	if err != nil {
		return SessionTokens{}, model.User{}, err
	}

	// Not hardcoded to true: a device id can belong to a user who has
	// since verified an email (device_id is never cleared on upgrade),
	// and the token must reflect that user's actual current state.
	tokens, err := s.issueSession(ctx, user.ID, user.IsAnonymous)
	if err != nil {
		return SessionTokens{}, model.User{}, err
	}
	return tokens, user, nil
}

func (s *AuthService) resolveAnonymousUser(ctx context.Context, deviceID, timezone string) (model.User, error) {
	if deviceID != "" {
		user, err := s.users.GetByDeviceID(ctx, deviceID)
		switch {
		case err == nil:
			return user, nil
		case errors.Is(err, repository.ErrUserNotFound):
			// fall through to create
		default:
			return model.User{}, fmt.Errorf("look up device: %w", err)
		}
	}

	user, err := s.users.CreateAnonymous(ctx, deviceID, timezone)
	if err != nil {
		return model.User{}, fmt.Errorf("create anonymous user: %w", err)
	}
	return user, nil
}

// RequestEmailCode generates and emails a 6-digit code, unless email is
// being rate-limited (more than 1 request in the last minute, or more
// than 5 in the last hour) — in which case it silently does nothing and
// still returns nil.
//
// This single rate-limit check is deliberately what also defeats email
// enumeration: the caller (handler) always responds with the same
// generic success regardless of whether this returns nil because a code
// was actually sent, or nil because the request was silently throttled
// — an attacker probing for registered emails can't tell those apart
// from the response. A non-nil error here is reserved for things that
// have nothing to do with any specific email's existence: a malformed
// address (ErrInvalidEmail, a 400 — rejecting garbage input isn't an
// enumeration leak) or an infrastructure failure (DB/SMTP down, which
// fails identically no matter which email was requested).
func (s *AuthService) RequestEmailCode(ctx context.Context, rawEmail string) error {
	emailAddr := normalizeEmail(rawEmail)
	if !isValidEmailFormat(emailAddr) {
		return ErrInvalidEmail
	}

	now := time.Now()

	perMinute, err := s.codes.CountSince(ctx, emailAddr, now.Add(-time.Minute))
	if err != nil {
		return fmt.Errorf("check per-minute rate limit: %w", err)
	}
	if perMinute >= requestCodeLimitPerMinute {
		return nil
	}

	perHour, err := s.codes.CountSince(ctx, emailAddr, now.Add(-time.Hour))
	if err != nil {
		return fmt.Errorf("check per-hour rate limit: %w", err)
	}
	if perHour >= requestCodeLimitPerHour {
		return nil
	}

	code, err := generateSixDigitCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	if _, err := s.codes.Create(ctx, emailAddr, code, now.Add(verificationCodeTTL)); err != nil {
		return fmt.Errorf("store verification code: %w", err)
	}

	if err := s.emailSender.SendVerificationCode(ctx, emailAddr, code); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

// VerifyEmailCode checks a code and, if valid, returns a session. Which
// user that session belongs to follows two branches:
//
//   - Durum A: emailAddr already belongs to a verified account. That
//     account is logged into, full stop — any anonymousAccessToken
//     supplied is silently discarded (not deleted, just never linked;
//     its data stays exactly as it was, orphaned).
//   - Durum B: no verified account owns emailAddr yet. If
//     anonymousAccessToken is a valid, still-anonymous session token,
//     that user is upgraded in place (same id, so every
//     conversation/quota/credit row carries over) to a verified
//     account. Otherwise, a brand new verified user is created.
//
// The bool return is isNewRegistration: true only for Durum B, so the
// client knows whether to show a "complete your profile" step.
func (s *AuthService) VerifyEmailCode(ctx context.Context, rawEmail, code, anonymousAccessToken string) (SessionTokens, bool, error) {
	emailAddr := normalizeEmail(rawEmail)
	if !isValidEmailFormat(emailAddr) {
		return SessionTokens{}, false, ErrInvalidEmail
	}

	if err := s.checkAndConsumeCode(ctx, emailAddr, code); err != nil {
		return SessionTokens{}, false, err
	}

	if existing, err := s.users.GetVerifiedByEmail(ctx, emailAddr); err == nil {
		// Durum A.
		tokens, err := s.issueSession(ctx, existing.ID, false)
		return tokens, false, err
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return SessionTokens{}, false, fmt.Errorf("look up verified user: %w", err)
	}

	// Durum B.
	var (
		user model.User
		err  error
	)
	if anonymousUserID, ok := s.tryResolveAnonymousUser(anonymousAccessToken); ok {
		user, err = s.users.UpgradeAnonymousToVerifiedEmail(ctx, anonymousUserID, emailAddr)
	} else {
		user, err = s.users.CreateVerifiedEmailUser(ctx, emailAddr, "")
	}
	if err != nil {
		return SessionTokens{}, false, fmt.Errorf("finalize email registration: %w", err)
	}

	tokens, err := s.issueSession(ctx, user.ID, false)
	return tokens, true, err
}

// checkAndConsumeCode validates emailAddr's latest code against code,
// enforcing expiry and the attempt cap, and marks it used on success.
func (s *AuthService) checkAndConsumeCode(ctx context.Context, emailAddr, code string) error {
	verification, err := s.codes.GetLatest(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, repository.ErrVerificationCodeNotFound) {
			return ErrInvalidCode
		}
		return fmt.Errorf("load verification code: %w", err)
	}

	if verification.UsedAt != nil {
		return ErrInvalidCode
	}
	if time.Now().After(verification.ExpiresAt) {
		return ErrCodeExpired
	}
	if verification.AttemptCount >= verificationCodeMaxAttempts {
		return ErrTooManyAttempts
	}

	if verification.Code != code {
		if err := s.codes.IncrementAttempt(ctx, verification.ID); err != nil {
			return fmt.Errorf("record failed attempt: %w", err)
		}
		return ErrInvalidCode
	}

	if err := s.codes.MarkUsed(ctx, verification.ID); err != nil {
		return fmt.Errorf("mark verification code used: %w", err)
	}
	return nil
}

// tryResolveAnonymousUser returns the user id encoded in accessToken,
// but only if it's a valid token for an anonymous session — a verified
// user's token, or a garbage/expired token, both just mean "no anonymous
// session to upgrade" rather than an error, so VerifyEmailCode can fall
// through to creating a fresh account.
func (s *AuthService) tryResolveAnonymousUser(accessToken string) (string, bool) {
	if accessToken == "" {
		return "", false
	}
	claims, err := s.accessTokens.Verify(accessToken)
	if err != nil || !claims.IsAnonymous {
		return "", false
	}
	return claims.UserID, true
}

// CompleteRegistration fills in profile fields (currently just
// display_name) for an already-authenticated user.
func (s *AuthService) CompleteRegistration(ctx context.Context, userID, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return ErrInvalidDisplayName
	}
	if err := s.users.UpdateDisplayName(ctx, userID, displayName); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("update display name: %w", err)
	}
	return nil
}

// RefreshSession rotates a refresh token: the presented one is revoked
// and a brand new access/refresh pair is issued. Presenting a token that
// was already revoked (i.e. already used once) is treated as evidence of
// token theft — every refresh token for that user is revoked as a
// precaution, forcing every session to re-authenticate.
func (s *AuthService) RefreshSession(ctx context.Context, rawToken string) (SessionTokens, error) {
	if rawToken == "" {
		return SessionTokens{}, ErrInvalidRefreshToken
	}

	stored, err := s.refreshTokens.GetByHash(ctx, auth.HashRefreshToken(rawToken))
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return SessionTokens{}, ErrInvalidRefreshToken
		}
		return SessionTokens{}, fmt.Errorf("look up refresh token: %w", err)
	}

	if stored.RevokedAt != nil {
		if err := s.refreshTokens.RevokeAllForUser(ctx, stored.UserID); err != nil {
			log.Printf("auth: failed to revoke sessions after refresh token reuse (user=%s): %v", stored.UserID, err)
		}
		return SessionTokens{}, ErrInvalidRefreshToken
	}
	if time.Now().After(stored.ExpiresAt) {
		return SessionTokens{}, ErrRefreshTokenExpired
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil {
		return SessionTokens{}, fmt.Errorf("revoke used refresh token: %w", err)
	}

	user, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return SessionTokens{}, fmt.Errorf("load user for refresh: %w", err)
	}

	return s.issueSession(ctx, user.ID, user.IsAnonymous)
}

// Logout revokes a refresh token. Idempotent: an unknown or already-gone
// token isn't an error — logging out twice, or with a stale token,
// should just quietly succeed.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	stored, err := s.refreshTokens.GetByHash(ctx, auth.HashRefreshToken(rawToken))
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return nil
		}
		return fmt.Errorf("look up refresh token: %w", err)
	}
	return s.refreshTokens.Revoke(ctx, stored.ID)
}

func (s *AuthService) issueSession(ctx context.Context, userID string, isAnonymous bool) (SessionTokens, error) {
	accessToken, err := s.accessTokens.Issue(userID, isAnonymous)
	if err != nil {
		return SessionTokens{}, fmt.Errorf("issue access token: %w", err)
	}

	rawRefresh, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return SessionTokens{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if _, err := s.refreshTokens.Create(ctx, userID, hash, time.Now().Add(s.refreshTTL)); err != nil {
		return SessionTokens{}, fmt.Errorf("store refresh token: %w", err)
	}

	// Best-effort: a lookup failure here shouldn't block issuing a
	// session over what's just a UX nicety (skipping onboarding when
	// it's already done) — default to false, which just means the
	// client might show onboarding again unnecessarily, never worse.
	hasCompletedOnboarding, err := s.onboardingProfiles.Exists(ctx, userID)
	if err != nil {
		log.Printf("auth: failed to check onboarding status (user=%s): %v", userID, err)
	}

	return SessionTokens{
		AccessToken:            accessToken,
		RefreshToken:           rawRefresh,
		HasCompletedOnboarding: hasCompletedOnboarding,
	}, nil
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func isValidEmailFormat(e string) bool {
	at := strings.IndexByte(e, '@')
	return at > 0 && at < len(e)-1 && !strings.ContainsAny(e, " \t\n")
}

func generateSixDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

import Foundation

/// The single source of truth for the app's session: which auth state
/// we're in, and every operation that can change it (bootstrap, linking
/// an email, completing registration, and the refresh-or-fall-back-to-
/// anonymous recovery APIClient calls into on a 401). RootView branches
/// on `state`; screens elsewhere read `email` for display.
@MainActor
final class AuthManager: ObservableObject {
    static let shared = AuthManager()

    @Published private(set) var state: AuthState = .loading
    @Published private(set) var email: String?
    /// Whether onboarding is done, per the last session response we saw
    /// (or, before any network call, whatever was last cached — see
    /// OnboardingStatusStore). RootView reads this to decide whether to
    /// show the onboarding flow.
    @Published private(set) var hasCompletedOnboarding: Bool

    private let apiClient: APIClient
    private let tokenStorage: TokenStorage
    private let onboardingStatusStore = OnboardingStatusStore()

    /// Coalesces concurrent 401s (e.g. several requests in flight at
    /// once) into a single in-flight refresh, so they all await the same
    /// outcome instead of each independently rotating the refresh token
    /// out from under the others.
    private var refreshTask: Task<Bool, Never>?

    init(apiClient: APIClient = .shared, tokenStorage: TokenStorage = TokenStorage()) {
        self.apiClient = apiClient
        self.tokenStorage = tokenStorage
        self.email = tokenStorage.email
        self.hasCompletedOnboarding = onboardingStatusStore.hasCompletedOnboarding

        #if DEBUG
        // UI tests that aren't exercising onboarding itself opt into
        // this via `--uitest-skip-onboarding` (see
        // PerchlyUITests/XCTestCase+Onboarding.swift) so they aren't
        // hostage to a real tap making it through the onboarding flow —
        // never set outside of a test target's launch arguments, and
        // compiled out of Release entirely.
        if ProcessInfo.processInfo.arguments.contains("--uitest-skip-onboarding") {
            self.hasCompletedOnboarding = true
            onboardingStatusStore.hasCompletedOnboarding = true
        }
        #endif
    }

    // MARK: - Bootstrap

    /// Call once, at launch (see AnonymousBootstrapView). If a session
    /// already exists, its validity is trusted without asking the
    /// backend first — a stale access token is only discovered lazily,
    /// on whichever API call happens to need it first, via APIClient's
    /// 401 -> refresh interceptor.
    func bootstrap() async {
        state = .loading

        if let accessToken = tokenStorage.accessToken {
            state = state(forAccessToken: accessToken)
            return
        }

        await createAnonymousSession()
    }

    private func state(forAccessToken accessToken: String) -> AuthState {
        guard let claims = AccessTokenClaims.decode(from: accessToken) else {
            return .anonymous
        }
        return claims.isAnonymous ? .anonymous : .authenticated
    }

    // MARK: - Anonymous session

    private struct CreateAnonymousSessionRequest: Encodable {
        let device_id: String
    }

    private struct SessionResponse: Decodable {
        let access_token: String
        let refresh_token: String
        let has_completed_onboarding: Bool
    }

    private func createAnonymousSession() async {
        do {
            let body = try JSONEncoder().encode(CreateAnonymousSessionRequest(device_id: tokenStorage.deviceID))
            let request = APIRequest(path: "/auth/anonymous", method: .post, body: body)
            let response: SessionResponse = try await apiClient.send(request, skipAuthRetry: true)
            try tokenStorage.saveSession(accessToken: response.access_token, refreshToken: response.refresh_token)
            updateHasCompletedOnboarding(response.has_completed_onboarding)
            state = .anonymous
        } catch {
            // Nothing meaningful to fall back to here — without any
            // session at all the app can't call the backend regardless.
            // Land in .anonymous with no stored tokens rather than
            // getting stuck in .loading forever; every API call will
            // fail until the user retries (e.g. relaunching once the
            // backend is reachable again).
            state = .anonymous
        }
    }

    // MARK: - Email OTP

    private struct RequestEmailCodeRequest: Encodable { let email: String }

    /// Requests a 6-digit code be emailed. The backend always responds
    /// with a generic success (even when it silently rate-limited the
    /// request, by design — see the backend's AuthService) except for a
    /// malformed address or a genuine server failure, both of which
    /// throw here as a normal APIError.
    func requestEmailCode(email: String) async throws {
        let body = try JSONEncoder().encode(RequestEmailCodeRequest(email: email))
        let request = APIRequest(path: "/auth/email/request-code", method: .post, body: body)
        try await apiClient.send(request, skipAuthRetry: true)
    }

    private struct VerifyEmailCodeRequest: Encodable {
        let email: String
        let code: String
        let access_token: String
    }

    private struct VerifyEmailCodeResponse: Decodable {
        let access_token: String
        let refresh_token: String
        let has_completed_onboarding: Bool
        let is_new_registration: Bool
    }

    /// Verifies a 6-digit code and stores the resulting session either
    /// way. Returns whether the caller still needs to collect a
    /// display_name (RegisterCompleteView) before considering sign-in
    /// complete.
    ///
    /// Deliberately does NOT flip `state` to `.authenticated` when the
    /// result is "new registration": per the backend's Durum A/B
    /// contract, this call already returns a fully valid session in
    /// both cases (the account is verified either way) — but the CLIENT
    /// still wants to route through RegisterCompleteView first, and
    /// only ITS success should flip global state (see
    /// completeRegistration). Flipping state here too would let RootView
    /// jump straight to the main app mid-registration.
    @discardableResult
    func verifyEmailCode(email: String, code: String) async throws -> Bool {
        let body = try JSONEncoder().encode(VerifyEmailCodeRequest(
            email: email, code: code, access_token: tokenStorage.accessToken ?? ""
        ))
        let request = APIRequest(path: "/auth/email/verify-code", method: .post, body: body)
        let response: VerifyEmailCodeResponse = try await apiClient.send(request, skipAuthRetry: true)

        try tokenStorage.saveSession(accessToken: response.access_token, refreshToken: response.refresh_token)
        try tokenStorage.saveEmail(email)
        self.email = email
        updateHasCompletedOnboarding(response.has_completed_onboarding)

        if !response.is_new_registration {
            state = .authenticated
        }
        return response.is_new_registration
    }

    private struct CompleteRegistrationRequest: Encodable { let display_name: String }

    /// Completes registration for the account just linked via
    /// `verifyEmailCode`. Only after this succeeds does `state` become
    /// `.authenticated`.
    func completeRegistration(displayName: String) async throws {
        let body = try JSONEncoder().encode(CompleteRegistrationRequest(display_name: displayName))
        let request = APIRequest(path: "/auth/register/complete", method: .post, body: body)
        try await apiClient.send(request)
        state = .authenticated
    }

    // MARK: - Refresh (called by APIClient on a 401)

    private struct RefreshRequest: Encodable { let refresh_token: String }

    /// Attempts to rotate the refresh token. Returns `true` if a new
    /// access token is now stored and the request that triggered this
    /// may safely be retried with it; `false` if refreshing failed and
    /// the session fell back to a brand-new anonymous one — in which
    /// case the caller must NOT retry the original request (see
    /// APIClient.performWithRefreshRetry's doc comment for why).
    func refreshOrFallbackToAnonymous() async -> Bool {
        if let existing = refreshTask {
            return await existing.value
        }

        let task = Task { await self.performRefreshOrFallback() }
        refreshTask = task
        let result = await task.value
        refreshTask = nil
        return result
    }

    private func performRefreshOrFallback() async -> Bool {
        guard let refreshToken = tokenStorage.refreshToken else {
            await fallBackToAnonymous()
            return false
        }

        do {
            let body = try JSONEncoder().encode(RefreshRequest(refresh_token: refreshToken))
            let request = APIRequest(path: "/auth/refresh", method: .post, body: body)
            // skipAuthRetry: true is required here, not just tidy — this
            // IS the refresh call. If it 401s and the interceptor tried
            // to "refresh and retry" again, it would call back into this
            // exact function while it's already running, awaiting its
            // own result forever.
            let response: SessionResponse = try await apiClient.send(request, skipAuthRetry: true)
            try tokenStorage.saveSession(accessToken: response.access_token, refreshToken: response.refresh_token)
            updateHasCompletedOnboarding(response.has_completed_onboarding)
            state = state(forAccessToken: response.access_token)
            return true
        } catch {
            await fallBackToAnonymous()
            return false
        }
    }

    private func fallBackToAnonymous() async {
        tokenStorage.clearSession()
        email = nil
        await createAnonymousSession()
    }

    // MARK: - Onboarding status

    private func updateHasCompletedOnboarding(_ value: Bool) {
        hasCompletedOnboarding = value
        onboardingStatusStore.hasCompletedOnboarding = value
    }

    /// Called by OnboardingCoordinator the instant its local flow
    /// finishes — regardless of whether syncing the profile to the
    /// backend succeeded (see OnboardingRetryQueue) — so RootView stops
    /// showing onboarding immediately, without waiting on any network
    /// call.
    func markOnboardingCompleted() {
        updateHasCompletedOnboarding(true)
    }

    // MARK: - Debug: reset local data

    private struct LogoutRequest: Encodable { let refresh_token: String }

    /// Wipes every locally stored auth artifact — session tokens AND the
    /// device id — and bootstraps a completely fresh anonymous session,
    /// as if the app had just been installed on a new device. Backing
    /// this is `TokenStorage.clearAll()`, distinct from the normal
    /// fallback path's `clearSession()` (which deliberately keeps the
    /// device id).
    ///
    /// This exists for local development: verifying Durum A/B, the
    /// refresh interceptor, etc. otherwise requires erasing the whole
    /// Simulator, since a plain uninstall does NOT clear Keychain data
    /// on iOS. See ProfileView, where this is wired to a `#if DEBUG`-only
    /// button — never exposed to real users, since it silently
    /// disconnects whatever email is currently linked with no undo.
    ///
    /// Known limitation: this resets `hasCompletedOnboarding` to
    /// `false`, but RootView won't actually show onboarding again
    /// within the SAME app run — its `needsOnboarding` check also looks
    /// at the onboarding coordinator's own step, which stays wherever
    /// it was left (e.g. `.completed`, if onboarding already finished
    /// once this run). A real relaunch re-evaluates both fresh and
    /// behaves correctly; for immediate re-testing, relaunch after
    /// resetting rather than expecting onboarding to reappear in place.
    func debugResetLocalSession() async {
        if let refreshToken = tokenStorage.refreshToken {
            // Best-effort: revoke server-side too, but a reset must
            // proceed locally regardless of whether this succeeds.
            let body = try? JSONEncoder().encode(LogoutRequest(refresh_token: refreshToken))
            if let body {
                try? await apiClient.send(APIRequest(path: "/auth/logout", method: .post, body: body), skipAuthRetry: true)
            }
        }

        tokenStorage.clearAll()
        email = nil
        // Deliberately not `state = .loading` first: that would make
        // RootView switch to AnonymousBootstrapView, whose own `.task`
        // calls `bootstrap()` — racing with the `createAnonymousSession()`
        // call right below and firing two concurrent
        // POST /auth/anonymous requests. createAnonymousSession() sets
        // `state` itself once it's done; nothing needs the bridge state.
        await createAnonymousSession()
    }
}

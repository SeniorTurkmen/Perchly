import XCTest
@testable import Perchly

/// Exercises APIClient's 401 -> refresh -> retry interceptor and
/// AuthManager's fallback-to-anonymous path against a real, locally
/// running backend — the trickiest piece of the auth rework to get
/// right (in particular, avoiding the interceptor calling back into
/// itself when the /auth/refresh call it makes fails). Skips itself if
/// no backend is reachable.
final class AuthInterceptorIntegrationTests: XCTestCase {
    private let tokenStorage = TokenStorage()

    override func setUp() async throws {
        do {
            try await APIClient.shared.send(APIRequest(path: "/health"))
        } catch {
            throw XCTSkip("No local backend reachable at localhost:8080 (see backend/scripts/dev.sh); skipping.")
        }
    }

    /// A corrupted access token but a still-valid refresh token: the
    /// interceptor should transparently refresh and retry, succeeding
    /// with a brand new access token.
    func testExpiredAccessToken_RefreshesAndRetriesSuccessfully() async throws {
        await AuthManager.shared.bootstrap()
        guard let validRefreshToken = tokenStorage.refreshToken else {
            XCTFail("expected a refresh token after bootstrap")
            return
        }

        try tokenStorage.saveSession(accessToken: "garbage.garbage.garbage", refreshToken: validRefreshToken)

        // Any authenticated endpoint will do; personas listing doesn't
        // need auth, so use something that does.
        let personaID = try await firstPersonaID()
        struct CreateConversationRequest: Encodable { let persona_id: String }
        let body = try JSONEncoder().encode(CreateConversationRequest(persona_id: personaID))

        struct ConversationResponse: Decodable { let id: String }
        let conversation: ConversationResponse = try await APIClient.shared.send(
            APIRequest(path: "/conversations", method: .post, body: body)
        )
        XCTAssertFalse(conversation.id.isEmpty)

        // The interceptor must have replaced the corrupted access token.
        XCTAssertNotEqual(tokenStorage.accessToken, "garbage.garbage.garbage")
        XCTAssertNotNil(tokenStorage.accessToken)
    }

    /// Both tokens invalid: refreshing must fail, AuthManager must fall
    /// back to a fresh anonymous session (state == .anonymous, new valid
    /// tokens stored) — and, critically, this must complete at all
    /// rather than hang, which is exactly what would happen if the
    /// interceptor's own /auth/refresh call re-triggered itself on its
    /// own 401.
    func testInvalidRefreshToken_FallsBackToFreshAnonymousSession() async throws {
        await AuthManager.shared.bootstrap()
        let deviceIDBefore = tokenStorage.deviceID

        try tokenStorage.saveSession(accessToken: "garbage.garbage.garbage", refreshToken: "not-a-real-refresh-token")

        let personaID = try await firstPersonaID()
        struct CreateConversationRequest: Encodable { let persona_id: String }
        let body = try JSONEncoder().encode(CreateConversationRequest(persona_id: personaID))

        do {
            try await APIClient.shared.send(APIRequest(path: "/conversations", method: .post, body: body))
            XCTFail("expected the original request to fail — falling back to a new identity should not silently retry it")
        } catch {
            // Expected: the original 401 surfaces rather than being retried.
        }

        let stateAfterFallback = await AuthManager.shared.state
        XCTAssertEqual(stateAfterFallback, .anonymous)
        XCTAssertNotNil(tokenStorage.accessToken)
        XCTAssertNotEqual(tokenStorage.accessToken, "garbage.garbage.garbage")
        // Same device id -> the new anonymous session resolves back to
        // the same backend user, per AuthManager.resolveAnonymousUser.
        XCTAssertEqual(tokenStorage.deviceID, deviceIDBefore)

        // And the fallback session must actually be usable.
        struct ConversationResponse: Decodable { let id: String }
        let conversation: ConversationResponse = try await APIClient.shared.send(
            APIRequest(path: "/conversations", method: .post, body: body)
        )
        XCTAssertFalse(conversation.id.isEmpty)
    }

    private func firstPersonaID() async throws -> String {
        let personas: [Persona] = try await APIClient.shared.send(APIRequest(path: "/personas"))
        guard let id = personas.first?.id else {
            XCTFail("no personas seeded")
            throw APIError.invalidResponse
        }
        return id
    }
}

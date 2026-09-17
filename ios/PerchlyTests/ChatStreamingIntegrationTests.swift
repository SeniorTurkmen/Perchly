import XCTest
@testable import Perchly

/// Exercises the real APIClient.streamEvents SSE client against a locally
/// running backend (see backend/scripts/dev.sh) — the one piece of this
/// app that can't be verified by SwiftUI previews or a type-checker, since
/// it depends on exactly how AsyncBytes/AsyncLineSequence behave. Skips
/// itself if no backend is reachable, so it never breaks a normal build.
final class ChatStreamingIntegrationTests: XCTestCase {
    func testStreamEventsAgainstRunningBackend() async throws {
        let apiClient = APIClient.shared

        do {
            try await apiClient.send(APIRequest(path: "/health"))
        } catch {
            throw XCTSkip("No local backend reachable at localhost:8080 (see backend/scripts/dev.sh); skipping.")
        }

        // /conversations now requires auth — bootstrap an anonymous
        // session so APIClient has an access token to attach.
        await AuthManager.shared.bootstrap()

        let personas: [Persona] = try await apiClient.send(APIRequest(path: "/personas"))
        guard let personaID = personas.first?.id else {
            XCTFail("no personas seeded")
            return
        }

        struct CreateConversationRequest: Encodable { let persona_id: String }
        struct ConversationResponse: Decodable { let id: String }

        let body = try JSONEncoder().encode(CreateConversationRequest(persona_id: personaID))
        let conversation: ConversationResponse = try await apiClient.send(
            APIRequest(path: "/conversations", method: .post, body: body)
        )

        struct SendMessageRequest: Encodable { let content: String }
        let messageBody = try JSONEncoder().encode(SendMessageRequest(content: "test mesajı, parça parça gelmeli"))
        let request = APIRequest(path: "/conversations/\(conversation.id)/messages", method: .post, body: messageBody)

        var chunkCount = 0
        var receivedText = ""
        var sawDone = false

        for try await event in apiClient.streamEvents(request) {
            switch event.event {
            case "message":
                chunkCount += 1
                receivedText += try event.decodedText()
            case "reaction":
                // The persona may lead its reply with a reaction tag
                // instead of, or before, replying with text (see the
                // backend's ChatService.SendMessage) — a legitimate
                // event this test doesn't otherwise care about.
                break
            case "done":
                sawDone = true
            default:
                XCTFail("unexpected event: \(event.event)")
            }
        }

        XCTAssertTrue(sawDone, "expected a done event")
        XCTAssertGreaterThan(chunkCount, 1, "expected multiple chunks, got \(chunkCount)")
        XCTAssertEqual(receivedText, "test mesajı, parça parça gelmeli")
    }
}

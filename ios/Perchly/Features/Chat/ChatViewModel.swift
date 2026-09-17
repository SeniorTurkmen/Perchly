import Foundation

@MainActor
final class ChatViewModel: ObservableObject {
    @Published private(set) var messages: [ChatMessage] = []
    @Published var draft: String = ""
    @Published private(set) var isSending: Bool = false
    @Published var errorMessage: String?
    /// From `POST /conversations` `created_at` — the real session start,
    /// not a client-invented clock.
    @Published private(set) var sessionStartedAt: Date?

    let persona: Persona

    private let apiClient: APIClient
    private var conversationID: String?
    private var streamTask: Task<Void, Never>?
    private var hasLoadedHistory = false

    init(persona: Persona, conversationID: String? = nil, apiClient: APIClient = .shared) {
        self.persona = persona
        self.conversationID = conversationID
        self.apiClient = apiClient
    }

    /// Loads persisted history when this screen was opened from an
    /// existing thread. A new conversation is not created until the
    /// first send, so the inbox never fills with empty shells.
    func startConversationIfNeeded() async {
        if let conversationID {
            await loadHistoryIfNeeded(conversationID)
        }
    }

    private func loadHistoryIfNeeded(_ id: String) async {
        guard !hasLoadedHistory else { return }
        hasLoadedHistory = true

        do {
            let persisted: [PersistedMessage] = try await apiClient.send(
                APIRequest(path: "/conversations/\(id)/messages")
            )
            messages = persisted.map { ChatMessage(persisted: $0) }
            sessionStartedAt = persisted.first?.createdAt
        } catch {
            hasLoadedHistory = false
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }

    private func createConversation() async {
        guard conversationID == nil else { return }

        struct CreateConversationRequest: Encodable { let persona_id: String }
        struct ConversationResponse: Decodable {
            let id: String
            let personaID: String
            let createdAt: Date

            enum CodingKeys: String, CodingKey {
                case id
                case personaID = "persona_id"
                case createdAt = "created_at"
            }
        }

        do {
            let body = try JSONEncoder().encode(CreateConversationRequest(persona_id: persona.id))
            let request = APIRequest(path: "/conversations", method: .post, body: body)
            let conversation: ConversationResponse = try await apiClient.send(request)
            guard conversation.personaID == persona.id else {
                errorMessage = "konuşma beklenen persona ile eşleşmedi"
                return
            }
            conversationID = conversation.id
            sessionStartedAt = conversation.createdAt
        } catch {
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }

    func sendDraft() {
        let text = draft.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty, !isSending else { return }
        draft = ""
        send(text)
    }

    func applyQuickReply(_ text: String) {
        guard !isSending else { return }
        draft = text
    }

    private func send(_ text: String) {
        let userID = UUID()
        messages.append(ChatMessage(id: userID, role: .user, content: text))

        let assistantID = UUID()
        messages.append(ChatMessage(id: assistantID, role: .assistant, content: "", isStreaming: true))

        isSending = true
        errorMessage = nil

        streamTask?.cancel()
        streamTask = Task { [weak self] in
            await self?.stream(text, userMessageID: userID, assistantMessageID: assistantID)
        }
    }

    private func stream(_ text: String, userMessageID: UUID, assistantMessageID: UUID) async {
        defer {
            setStreaming(assistantMessageID, streaming: false)
            isSending = false
        }

        await startConversationIfNeeded()
        if conversationID == nil {
            await createConversation()
        }
        guard let conversationID else { return }

        struct SendMessageRequest: Encodable { let content: String }
        struct DoneEventPayload: Decodable {
            let userMessageID: String
            let assistantMessageID: String?

            enum CodingKeys: String, CodingKey {
                case userMessageID = "user_message_id"
                case assistantMessageID = "assistant_message_id"
            }
        }

        do {
            let body = try JSONEncoder().encode(SendMessageRequest(content: text))
            let request = APIRequest(path: "/conversations/\(conversationID)/messages", method: .post, body: body)

            for try await event in apiClient.streamEvents(request) {
                switch event.event {
                case "message":
                    appendDelta(try event.decodedText(), to: assistantMessageID)
                case "reaction":
                    setReactionEmoji(try event.decodedText(), on: userMessageID)
                case "error":
                    errorMessage = (try? event.decodedText()) ?? "Bilinmeyen hata"
                case "done":
                    if let data = event.data.data(using: .utf8),
                       let payload = try? JSONDecoder().decode(DoneEventPayload.self, from: data) {
                        setBackendID(payload.userMessageID, on: userMessageID)
                        if let assistantBackendID = payload.assistantMessageID, !assistantBackendID.isEmpty {
                            setBackendID(assistantBackendID, on: assistantMessageID)
                        }
                    }
                default:
                    break
                }
            }
        } catch is CancellationError {
            // Superseded by a newer message; leave whatever streamed in as-is.
        } catch {
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }

    private func appendDelta(_ delta: String, to id: UUID) {
        guard let index = messages.firstIndex(where: { $0.id == id }) else { return }
        messages[index].content += delta
    }

    private func setStreaming(_ id: UUID, streaming: Bool) {
        guard let index = messages.firstIndex(where: { $0.id == id }) else { return }
        messages[index].isStreaming = streaming
    }

    private func setBackendID(_ backendID: String, on id: UUID) {
        guard let index = messages.firstIndex(where: { $0.id == id }) else { return }
        messages[index].backendID = backendID
    }

    private func setReactionEmoji(_ emoji: String, on id: UUID) {
        guard let index = messages.firstIndex(where: { $0.id == id }) else { return }
        messages[index].reactionEmoji = emoji
    }

    // MARK: - Reactions

    private struct SetReactionRequest: Encodable { let emoji: String }

    /// Sets (emoji != nil) or clears (nil) the caller's reaction on an
    /// assistant message — see ChatBubble's long-press picker. Applies
    /// optimistically so the badge appears instantly, and reverts if the
    /// backend call fails. A message whose real id hasn't arrived yet
    /// (still mid-stream — see ChatMessage.backendID) can't be reacted
    /// to; the picker isn't offered in that case.
    func setReaction(on message: ChatMessage, emoji: String?) async {
        guard let conversationID, let backendID = message.backendID else { return }
        guard let index = messages.firstIndex(where: { $0.id == message.id }) else { return }

        let previous = messages[index].reactionEmoji
        messages[index].reactionEmoji = emoji

        do {
            let path = "/conversations/\(conversationID)/messages/\(backendID)/reaction"
            if let emoji {
                let body = try JSONEncoder().encode(SetReactionRequest(emoji: emoji))
                try await apiClient.send(APIRequest(path: path, method: .post, body: body))
            } else {
                try await apiClient.send(APIRequest(path: path, method: .delete))
            }
        } catch {
            if let index = messages.firstIndex(where: { $0.id == message.id }) {
                messages[index].reactionEmoji = previous
            }
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }
}

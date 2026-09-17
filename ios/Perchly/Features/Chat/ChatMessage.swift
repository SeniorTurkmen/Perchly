import Foundation

/// A single bubble in the chat transcript. Built either from persisted
/// `GET /conversations/{id}/messages` rows or from the live SSE stream.
struct ChatMessage: Identifiable, Equatable {
    enum Role: Equatable {
        case user
        case assistant
    }

    let id: UUID
    let role: Role
    var content: String
    var isStreaming: Bool
    let createdAt: Date
    /// The single emoji reaction left on this message by whichever side
    /// didn't send it, or nil if none. Always one of
    /// MessageReaction.allowedEmojis.
    var reactionEmoji: String?
    /// This message's real, persisted backend id — nil until known.
    /// Known immediately for messages loaded from history (`id` below);
    /// for a message from the live SSE stream, `id` starts as a
    /// locally-generated placeholder and backendID stays nil until the
    /// stream's "done" event reports the real one (see
    /// ChatViewModel.stream) — reacting to a message requires this, not
    /// the local placeholder id.
    var backendID: String?

    init(
        id: UUID = UUID(),
        role: Role,
        content: String,
        isStreaming: Bool = false,
        createdAt: Date = .now,
        reactionEmoji: String? = nil,
        backendID: String? = nil
    ) {
        self.id = id
        self.role = role
        self.content = content
        self.isStreaming = isStreaming
        self.createdAt = createdAt
        self.reactionEmoji = reactionEmoji
        self.backendID = backendID
    }

    init(persisted: PersistedMessage) {
        self.init(
            id: UUID(uuidString: persisted.id) ?? UUID(),
            role: persisted.role == "user" ? .user : .assistant,
            content: persisted.content,
            createdAt: persisted.createdAt,
            reactionEmoji: persisted.reactionEmoji,
            backendID: persisted.id
        )
    }
}

/// Mirrors the backend message JSON (`GET /conversations/{id}/messages`
/// and `last_message` on the inbox preview).
struct PersistedMessage: Identifiable, Codable, Equatable, Hashable {
    let id: String
    let conversationID: String
    let role: String
    let content: String
    let reactionEmoji: String?
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id, role, content
        case conversationID = "conversation_id"
        case reactionEmoji = "reaction_emoji"
        case createdAt = "created_at"
    }
}

/// Mirrors `GET /conversations` — one inbox row.
struct ConversationSummary: Identifiable, Codable, Equatable, Hashable {
    let id: String
    let userID: String
    let personaID: String
    let createdAt: Date
    let updatedAt: Date
    let persona: Persona
    let lastMessage: PersistedMessage?

    enum CodingKeys: String, CodingKey {
        case id
        case userID = "user_id"
        case personaID = "persona_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case persona
        case lastMessage = "last_message"
    }
}

import Foundation

/// The fixed, iMessage-tapback-style set of emoji reactions the product
/// supports, in display order — mirrors the backend's
/// `model.OrderedAllowedReactionEmojis` exactly. Kept as a single shared
/// list so the long-press picker (see ChatBubble) and the server-side
/// validator can never drift apart.
enum MessageReaction {
    static let allowedEmojis = ["❤️", "😂", "👍", "👎", "‼️", "❓"]
}

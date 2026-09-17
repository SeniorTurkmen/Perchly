import Foundation

/// `GET /users/chat-energy` — Keşfet's "Günün sohbet enerjisi" card.
/// `remaining` is min(remaining) across active personas, matching the
/// persona that will 429 first.
struct ChatEnergy: Decodable, Equatable {
    let remaining: Int
    let limit: Int
    let used: Int
    let moodLabel: String

    enum CodingKeys: String, CodingKey {
        case remaining, limit, used
        case moodLabel = "mood_label"
    }

    var progress: CGFloat {
        guard limit > 0 else { return 0 }
        return min(1, max(0, CGFloat(remaining) / CGFloat(limit)))
    }

    var shareCopy: String {
        "Acele yok, \(remaining) derin sohbet payı seninle"
    }

    var fractionLabel: String {
        "\(remaining)/\(limit)"
    }
}

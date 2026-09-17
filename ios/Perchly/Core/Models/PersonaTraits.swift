import Foundation

/// The five personality dials a user can tune for a persona, each
/// 0-100 — mirrors the backend's `model.PersonaTraits` exactly. Read
/// fresh by the backend on every message send, so a change here takes
/// effect starting with the user's very next message.
struct PersonaTraits: Codable, Equatable {
    var warmth: Int
    var humor: Int
    var wisdom: Int
    var directness: Int
    var energy: Int
}

/// Mirrors `GET/PUT/DELETE /personas/{id}/traits`.
struct PersonaTraitsResponse: Decodable {
    let traits: PersonaTraits
    let isCustomized: Bool

    enum CodingKeys: String, CodingKey {
        case traits
        case isCustomized = "is_customized"
    }
}

import Foundation
import SwiftUI

/// Mirrors the backend's `GET /personas` payload. `system_prompt` is
/// intentionally never sent to clients by the backend, so there is no
/// field for it here.
struct Persona: Identifiable, Codable, Equatable, Hashable {
    let id: String
    let slug: String
    let name: String
    let category: String
    let shortDescription: String
    let toneDescription: String
    let avatarURL: String?
    let accentColor: String
    /// Whether this persona is safe to show/offer to a user flagged as
    /// a minor (see OnboardingProfile.isMinor) — the backend enforces
    /// this server-side too when starting a conversation, but the
    /// persona-pick screen filters by it as well so a minor never sees
    /// (and then gets rejected from) a persona that isn't for them.
    let isMinorAppropriate: Bool
    let isActive: Bool
    let sortOrder: Int
    let createdAt: Date
    let updatedAt: Date
    /// Set only on `GET /personas?recommend=true`. Plain `GET /personas`
    /// omits it (decodes as nil).
    let recommended: Bool?
    let matchReason: String?

    enum CodingKeys: String, CodingKey {
        case id, slug, name, category
        case shortDescription = "short_description"
        case toneDescription = "tone_description"
        case avatarURL = "avatar_url"
        case accentColor = "accent_color"
        case isMinorAppropriate = "is_minor_appropriate"
        case isActive = "is_active"
        case sortOrder = "sort_order"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case recommended
        case matchReason = "match_reason"
    }
}

/// Visual mapping from a backend persona `category` onto the Keşfet
/// Stitch screen (filter pill, badge, glow, talk-button tint). Unknown
/// categories fall back to a neutral slate treatment so new seeds still
/// render without a client update.
struct PersonaCategoryStyle: Equatable {
    let id: String
    let filterTitle: String
    let badgeTitle: String
    let roleTitle: String
    let skillTitle: String
    let skillSymbol: String
    let accent: Color
    let glow: Color
    let badgeFill: Color
    let badgeText: Color
    let ringStart: Color
    let ringEnd: Color
    let talkFill: Color
    let filterDot: Color

    static func style(for category: String) -> PersonaCategoryStyle {
        switch category {
        case "daily_companion":
            PersonaCategoryStyle(
                id: category,
                filterTitle: String(localized: "Günlük Sohbet"),
                badgeTitle: String(localized: "Günlük Sohbet"),
                roleTitle: String(localized: "Sohbet Arkadaşı"),
                skillTitle: String(localized: "Yargılamadan dinler"),
                skillSymbol: "bubble.left.fill",
                accent: PerchlyPalette.Discover.primary,
                glow: PerchlyPalette.Discover.primaryFixedDim.opacity(0.4),
                badgeFill: PerchlyPalette.Discover.primaryFixed.opacity(0.6),
                badgeText: PerchlyPalette.Discover.onPrimaryFixed,
                ringStart: PerchlyPalette.Discover.primaryFixed,
                ringEnd: PerchlyPalette.Discover.primaryContainer,
                talkFill: PerchlyPalette.Discover.primary,
                filterDot: PerchlyPalette.Discover.primaryContainer
            )
        case "motivational_coach":
            PersonaCategoryStyle(
                id: category,
                filterTitle: String(localized: "Motivasyon & Koç"),
                badgeTitle: String(localized: "Motivasyon & Koç"),
                roleTitle: String(localized: "Motivasyon Koçu"),
                skillTitle: String(localized: "Küçük adımlar, somut öneriler"),
                skillSymbol: "flag.fill",
                accent: PerchlyPalette.Discover.tertiary,
                glow: PerchlyPalette.Discover.tertiaryFixed.opacity(0.5),
                badgeFill: PerchlyPalette.Discover.tertiaryFixed.opacity(0.6),
                badgeText: PerchlyPalette.Discover.onTertiaryFixed,
                ringStart: PerchlyPalette.Discover.tertiaryFixed,
                ringEnd: PerchlyPalette.Discover.tertiaryContainer,
                talkFill: PerchlyPalette.Discover.tertiary,
                filterDot: PerchlyPalette.Discover.tertiaryContainer
            )
        case "hobby_book_club":
            PersonaCategoryStyle(
                id: category,
                filterTitle: String(localized: "Hobi & Kitap"),
                badgeTitle: String(localized: "Hobi & Kitap"),
                roleTitle: String(localized: "Kitap & Hobi Partneri"),
                skillTitle: String(localized: "Kitap, film ve hobiler"),
                skillSymbol: "book.fill",
                accent: PerchlyPalette.Discover.onSurfaceVariant,
                glow: PerchlyPalette.Discover.surfaceContainerHigh.opacity(0.7),
                badgeFill: PerchlyPalette.Discover.surfaceContainerHigh,
                badgeText: PerchlyPalette.Discover.onSurface,
                ringStart: PerchlyPalette.Discover.surfaceVariant,
                ringEnd: PerchlyPalette.Discover.primaryFixed,
                talkFill: PerchlyPalette.Discover.onSurface,
                filterDot: PerchlyPalette.Discover.surfaceDim
            )
        default:
            PersonaCategoryStyle(
                id: category,
                filterTitle: String(localized: "Duygusal Destek"),
                badgeTitle: String(localized: "Duygusal Destek"),
                roleTitle: String(localized: "Samimi Dinleyici"),
                skillTitle: String(localized: "Empatik & Derin Dinleme"),
                skillSymbol: "brain.head.profile",
                accent: PerchlyPalette.Discover.secondary,
                glow: PerchlyPalette.Discover.secondaryFixedDim.opacity(0.45),
                badgeFill: PerchlyPalette.Discover.secondaryFixed.opacity(0.5),
                badgeText: PerchlyPalette.Discover.onSecondaryFixed,
                ringStart: PerchlyPalette.Discover.secondaryFixed,
                ringEnd: PerchlyPalette.Discover.secondaryContainer,
                talkFill: PerchlyPalette.Discover.secondary,
                filterDot: PerchlyPalette.Discover.secondaryContainer
            )
        }
    }
}

extension Persona {
    var categoryStyle: PersonaCategoryStyle { .style(for: category) }

    /// Backend `accent_color` (e.g. Ada `#FF6B35`). Used for avatars and
    /// session chrome so each seeded persona keeps its own identity.
    var accent: Color { Color(hex: accentColor) }

    /// Starter chips mapped from `category` — the only persona field that
    /// tells us *what this persona is for*. Copy is not a fake transcript;
    /// tapping still just fills the composer.
    var quickReplies: [(title: String, symbol: String)] {
        switch category {
        case "motivational_coach":
            [
                (String(localized: "Bir hedefe odaklanmak istiyorum"), "flag.fill"),
                (String(localized: "Bugün küçük bir adım atmak istiyorum"), "figure.walk"),
                (String(localized: "Biraz cesaretlendirilmeye ihtiyacım var"), "sparkles"),
            ]
        case "daily_companion":
            [
                (String(localized: "Günümü anlatmak istiyorum"), "text.bubble"),
                (String(localized: "Sadece sohbet etmek istiyorum"), "cup.and.saucer.fill"),
                (String(localized: "Biraz dinlenmek istiyorum"), "leaf"),
            ]
        case "hobby_book_club":
            [
                (String(localized: "Bir kitaptan bahsetmek istiyorum"), "book.fill"),
                (String(localized: "Film veya dizi konuşalım"), "film"),
                (String(localized: "Hobimden konuşalım"), "paintpalette.fill"),
            ]
        default:
            [
                (String(localized: "Biraz anlatayım..."), "square.and.pencil"),
                (String(localized: "Sadece sohbet etmek istiyorum"), "bubble.left"),
                (String(localized: "Tavsiye değil, dinlenmek istiyorum"), "leaf"),
            ]
        }
    }

    var listeningCue: String {
        switch category {
        case "motivational_coach":
            String(localized: "\(name) bir sonraki adımı düşünüyor...")
        case "hobby_book_club":
            String(localized: "\(name) ne diyeceğini düşünüyor...")
        default:
            String(localized: "\(name) dinliyor...")
        }
    }

    /// Four chips on Persona Detayı, mapped from `category` — the backend
    /// has no per-persona trait list, so these stay category-level.
    var detailTraits: [(title: String, symbol: String, tint: Color)] {
        switch category {
        case "motivational_coach":
            [
                (String(localized: "Somut küçük adımlar"), "flag.fill", PerchlyPalette.Discover.tertiary),
                (String(localized: "Cesaretlendirici"), "sparkles", PerchlyPalette.Discover.primary),
                (String(localized: "Disiplinli ama sıcak"), "flame.fill", PerchlyPalette.Discover.secondary),
                (String(localized: "Gizlilik Öncelikli"), "lock.fill", PerchlyPalette.Discover.primary),
            ]
        case "hobby_book_club":
            [
                (String(localized: "Meraklı sorular"), "questionmark.circle.fill", PerchlyPalette.Discover.secondary),
                (String(localized: "Kitap & film"), "book.fill", PerchlyPalette.Discover.primary),
                (String(localized: "Alçakgönüllü öneriler"), "leaf.fill", PerchlyPalette.Discover.tertiary),
                (String(localized: "Gizlilik Öncelikli"), "lock.fill", PerchlyPalette.Discover.primary),
            ]
        case "daily_companion":
            [
                (String(localized: "Aktif dinleme"), "ear.fill", PerchlyPalette.Discover.secondary),
                (String(localized: "Yargılamadan eşlik"), "heart.fill", PerchlyPalette.Discover.primary),
                (String(localized: "Gündelik dil"), "text.bubble.fill", PerchlyPalette.Discover.tertiary),
                (String(localized: "Gizlilik Öncelikli"), "lock.fill", PerchlyPalette.Discover.primary),
            ]
        default:
            [
                (String(localized: "Empatik dinleme"), "ear.fill", PerchlyPalette.Discover.secondary),
                (String(localized: "Tavsiye vermeden eşlik"), "brain.head.profile", PerchlyPalette.Discover.primary),
                (String(localized: "Düşük enerjili günler"), "bolt.fill", PerchlyPalette.Discover.tertiary),
                (String(localized: "Gizlilik Öncelikli"), "lock.fill", PerchlyPalette.Discover.primary),
            ]
        }
    }

    var detailAtmosphereLine: String {
        switch category {
        case "motivational_coach":
            String(localized: "Küçük bir adım atmak için hazır.")
        case "hobby_book_club":
            String(localized: "Kitap, film ve hobiler için hazır.")
        default:
            String(localized: "Sakin bir sohbet molası için hazır.")
        }
    }

    var detailZeroJudgmentLine: String {
        switch category {
        case "motivational_coach":
            String(localized: "Küçük adımları kutlar, asla küçümsemez.")
        case "hobby_book_club":
            String(localized: "Dayatmadan önerir, meraklı kalır.")
        default:
            String(localized: "Yargılamadan, olduğun gibi dinler.")
        }
    }
}

// Not gated behind #if DEBUG: it's referenced from #Preview blocks
// across the codebase, and those are still type-checked (so still need
// this to exist) in every build configuration, not just Debug.
extension Persona {
    static let preview = Persona(
        id: "8a870ee3-e627-4871-8764-f310cb3251c4",
        slug: "motivational-coach",
        name: "Ada",
        category: "motivational_coach",
        shortDescription: "Hedeflerine ulaşman için seni içten içe iten, disiplinli ama sıcak bir koç.",
        toneDescription: "Enerjik, doğrudan, cesaretlendirici; asla küçümsemez, küçük adımları kutlar.",
        avatarURL: nil,
        accentColor: "#FF6B35",
        isMinorAppropriate: true,
        isActive: true,
        sortOrder: 1,
        createdAt: .now,
        updatedAt: .now,
        recommended: true,
        matchReason: "Motive olmak istediğin için önerdik."
    )
}

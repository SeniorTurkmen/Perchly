import Foundation

/// What OnboardingCoordinator collects across its four screens.
/// `Codable` so `OnboardingRetryQueue` can persist an unsent one to
/// disk and replay it later.
struct OnboardingProfile: Codable, Equatable {
    enum AgeRange: String, CaseIterable, Codable {
        case under18, age18to24, age25to34, age35plus
    }

    enum MoodPreference: String, CaseIterable, Codable {
        case motivation, dailyChat, hobbyTalk, skipped
    }

    var ageRange: AgeRange?
    var moodPreference: MoodPreference?
    var notificationsGranted: Bool = false
    var selectedPersonaID: String?

    /// Automatically true for `.under18` — never set independently, so
    /// it can never drift out of sync with `ageRange`. The backend
    /// re-derives the same value server-side from age_range and never
    /// trusts whatever a client sends for it either (see
    /// AuthService/OnboardingService) — this mirrors that on purpose.
    var isMinor: Bool { ageRange == .under18 }
}

extension OnboardingProfile.MoodPreference {
    /// The persona `category` (as the backend's seeded personas define
    /// it) this mood preference should prioritize on the persona-pick
    /// screen. `nil` for `.skipped`, which leaves the default order.
    var personaCategory: String? {
        switch self {
        case .motivation: return "motivational_coach"
        case .dailyChat: return "daily_companion"
        case .hobbyTalk: return "hobby_book_club"
        case .skipped: return nil
        }
    }
}

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

    /// How the user wants personas to address them. Synced on
    /// `POST /users/onboarding-profile` as `preferred_name`.
    /// `nil` when `skipHitap` is true — never a generated nickname.
    var preferredName: String?
    /// True when the user chose anonymous continue: store no name and
    /// do not address them by any handle in chat (`skip_hitap`).
    var skipHitap: Bool = false
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

/// Client-side cache of `preferred_name` / `skip_hitap`. Personas read
/// the hitap from the backend prompt; this is for reinstall sync and UI.
enum PreferredNameRules {
    static let maxLength = 40

    static func isValid(_ name: String) -> Bool {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard (1...maxLength).contains(trimmed.count) else { return false }
        return trimmed.unicodeScalars.allSatisfy { !CharacterSet.controlCharacters.contains($0) }
    }
}

enum LocalPreferredNameStore {
    private static let nameKey = "perchly.local_preferred_name"
    private static let skipKey = "perchly.local_skip_hitap"
    /// Previous keys from the short-lived generated-nickname experiment.
    private static let legacyNameKey = "perchly.local_display_name"
    private static let legacyAnonymousKey = "perchly.local_display_name_is_anonymous"

    static func save(name: String) {
        UserDefaults.standard.set(name, forKey: nameKey)
        UserDefaults.standard.set(false, forKey: skipKey)
        UserDefaults.standard.removeObject(forKey: legacyNameKey)
        UserDefaults.standard.removeObject(forKey: legacyAnonymousKey)
    }

    static func clear() {
        UserDefaults.standard.removeObject(forKey: nameKey)
        UserDefaults.standard.set(true, forKey: skipKey)
        UserDefaults.standard.removeObject(forKey: legacyNameKey)
        UserDefaults.standard.removeObject(forKey: legacyAnonymousKey)
    }

    static var name: String? {
        if UserDefaults.standard.bool(forKey: skipKey) { return nil }
        if let name = UserDefaults.standard.string(forKey: nameKey), !name.isEmpty {
            return name
        }
        return nil
    }

    static var skipHitap: Bool {
        UserDefaults.standard.bool(forKey: skipKey)
    }
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

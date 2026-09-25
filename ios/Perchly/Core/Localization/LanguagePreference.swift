import Foundation

/// Persists an explicit in-app language override (see the language
/// picker in ProfileView) — distinct from the device's own system
/// language. `nil` (the default) means "follow the system", which is
/// what the overwhelming majority of users never need to change.
@MainActor
final class LanguagePreference: ObservableObject {
    static let shared = LanguagePreference()

    private nonisolated static let userDefaultsKey = "perchly.languageOverride"

    @Published private(set) var override: AppLanguage?

    private let defaults: UserDefaults

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
        if let raw = defaults.string(forKey: Self.userDefaultsKey) {
            override = AppLanguage(rawValue: raw)
        }
    }

    func setOverride(_ language: AppLanguage?) {
        override = language
        if let language {
            defaults.set(language.rawValue, forKey: Self.userDefaultsKey)
        } else {
            defaults.removeObject(forKey: Self.userDefaultsKey)
        }
    }

    /// The language actually in effect: the explicit override if one is
    /// set, otherwise the system's own preferred language reduced to
    /// one of Perchly's 8 supported ones, otherwise Turkish (the source
    /// language every string is originally written in, and — like the
    /// backend's apierror.DefaultLocale and admin's defaultLocale — the
    /// one fallback target used everywhere in this app, so all three
    /// degrade the same way for an unsupported language).
    var effective: AppLanguage { Self.resolveEffective(override: override) }

    var effectiveLocale: Locale { Locale(identifier: effective.rawValue) }

    /// A nonisolated snapshot for contexts that can't hop to the main
    /// actor just to read a language preference — APIClient's
    /// synchronous makeURLRequest, called from arbitrary Task contexts,
    /// is the only caller. Re-reads UserDefaults directly rather than
    /// touching the main-actor-isolated `override` property above.
    nonisolated static var currentLanguageCode: String {
        let raw = UserDefaults.standard.string(forKey: userDefaultsKey)
        let override = raw.flatMap(AppLanguage.init(rawValue:))
        return resolveEffective(override: override).languageCode
    }

    /// preferredLanguages is injectable (defaulting to the real
    /// Locale.preferredLanguages) purely so LanguagePreferenceTests can
    /// exercise the system-language-matching branch deterministically,
    /// independent of whatever language the test machine happens to run.
    nonisolated static func resolveEffective(override: AppLanguage?, preferredLanguages: [String] = Locale.preferredLanguages) -> AppLanguage {
        if let override { return override }
        for preferred in preferredLanguages {
            guard let code = Locale(identifier: preferred).language.languageCode?.identifier.lowercased() else {
                continue
            }
            if let match = AppLanguage.allCases.first(where: { $0.languageCode == code }) {
                return match
            }
        }
        return .tr
    }
}

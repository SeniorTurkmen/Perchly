import Foundation

/// The 8 languages Perchly is translated into — mirrors
/// backend/internal/apierror/locale.go's Locale type and
/// admin/src/i18n/locales.ts by hand (small, stable list, no
/// cross-language codegen in this repo).
enum AppLanguage: String, CaseIterable, Identifiable {
    case tr, en, de, ar, es, fr, ru
    case zh = "zh-Hans"

    var id: String { rawValue }

    var displayName: String {
        switch self {
        case .tr: "Türkçe"
        case .en: "English"
        case .de: "Deutsch"
        case .ar: "العربية"
        case .es: "Español"
        case .fr: "Français"
        case .ru: "Русский"
        case .zh: "中文"
        }
    }

    var isRTL: Bool { self == .ar }

    /// The bare primary language subtag ("zh", not "zh-Hans") — what
    /// the backend's Accept-Language parser (apierror.ParseAcceptLanguage)
    /// and admin's mirror of it both match against. Only "zh" differs
    /// from rawValue; every other case's subtag already equals it.
    var languageCode: String {
        self == .zh ? "zh" : rawValue
    }
}

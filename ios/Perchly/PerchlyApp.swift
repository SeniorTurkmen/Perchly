import SwiftUI

@main
struct PerchlyApp: App {
    @ObservedObject private var languagePreference = LanguagePreference.shared

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(\.locale, languagePreference.effectiveLocale)
                .environment(\.layoutDirection, languagePreference.effective.isRTL ? .rightToLeft : .leftToRight)
                // Forces a full view-identity reset when the language
                // changes — SwiftUI's Text views can otherwise hold
                // onto an already-resolved string from before the
                // switch instead of re-resolving against the new locale.
                .id(languagePreference.effective)
        }
    }
}

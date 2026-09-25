import XCTest
@testable import Perchly

@MainActor
final class LanguagePreferenceTests: XCTestCase {
    private func makePreference() -> LanguagePreference {
        // A fresh, uniquely-named suite per test — never touches the
        // real app's UserDefaults.standard, and never leaks state
        // between tests.
        let suiteName = "LanguagePreferenceTests.\(UUID().uuidString)"
        let defaults = UserDefaults(suiteName: suiteName)!
        addTeardownBlock { defaults.removePersistentDomain(forName: suiteName) }
        return LanguagePreference(defaults: defaults)
    }

    func testNoOverridePersistedInitially() {
        XCTAssertNil(makePreference().override)
    }

    func testSetOverridePersistsAndIsEffective() {
        let preference = makePreference()
        preference.setOverride(.de)
        XCTAssertEqual(preference.override, .de)
        XCTAssertEqual(preference.effective, .de)
        XCTAssertEqual(preference.effectiveLocale.identifier, "de")
    }

    func testClearingOverrideFallsBackToSystemResolution() {
        let preference = makePreference()
        preference.setOverride(.fr)
        preference.setOverride(nil)
        XCTAssertNil(preference.override)
    }

    func testOverrideSurvivesAcrossInstancesOfTheSameDefaults() {
        let suiteName = "LanguagePreferenceTests.\(UUID().uuidString)"
        let defaults = UserDefaults(suiteName: suiteName)!
        addTeardownBlock { defaults.removePersistentDomain(forName: suiteName) }

        LanguagePreference(defaults: defaults).setOverride(.ru)
        // A brand-new instance reading the same UserDefaults suite
        // should pick the persisted override back up — this is what
        // makes the choice survive an app relaunch.
        XCTAssertEqual(LanguagePreference(defaults: defaults).override, .ru)
    }

    // MARK: - resolveEffective (the pure function APIClient's nonisolated
    // currentLanguageCode also goes through)

    func testResolveEffective_ExplicitOverrideAlwaysWins() {
        XCTAssertEqual(
            LanguagePreference.resolveEffective(override: .es, preferredLanguages: ["en-US"]),
            .es
        )
    }

    func testResolveEffective_MatchesSystemPreferredLanguage() {
        XCTAssertEqual(
            LanguagePreference.resolveEffective(override: nil, preferredLanguages: ["de-DE", "en-US"]),
            .de
        )
    }

    func testResolveEffective_ReducesRegionAndScriptSubtags() {
        XCTAssertEqual(
            LanguagePreference.resolveEffective(override: nil, preferredLanguages: ["zh-Hans-CN"]),
            .zh
        )
    }

    func testResolveEffective_SkipsUnsupportedLanguagesInOrder() {
        XCTAssertEqual(
            LanguagePreference.resolveEffective(override: nil, preferredLanguages: ["fi-FI", "ru-RU"]),
            .ru
        )
    }

    func testResolveEffective_FallsBackToTurkishWhenNothingMatches() {
        XCTAssertEqual(
            LanguagePreference.resolveEffective(override: nil, preferredLanguages: ["fi-FI", "sv-SE"]),
            .tr
        )
    }

    func testResolveEffective_FallsBackToTurkishForEmptyPreferences() {
        XCTAssertEqual(
            LanguagePreference.resolveEffective(override: nil, preferredLanguages: []),
            .tr
        )
    }

    // MARK: - AppLanguage.languageCode

    func testLanguageCode_MatchesRawValueForMostCases() {
        for language: AppLanguage in [.tr, .en, .de, .ar, .es, .fr, .ru] {
            XCTAssertEqual(language.languageCode, language.rawValue)
        }
    }

    func testLanguageCode_ChineseReducesToBarePrimarySubtag() {
        XCTAssertEqual(AppLanguage.zh.rawValue, "zh-Hans")
        XCTAssertEqual(AppLanguage.zh.languageCode, "zh")
    }

    func testIsRTL_OnlyArabic() {
        for language in AppLanguage.allCases {
            XCTAssertEqual(language.isRTL, language == .ar)
        }
    }
}

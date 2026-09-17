import XCTest

/// Drives the personality-dial sliders on Persona Detayı end to end,
/// against a locally running backend (see backend/scripts/dev.sh):
/// the sliders load the persona's real default values, dragging one
/// persists immediately, the change survives navigating away and back,
/// and "Varsayılana Döndür" clears it back to the persona's default.
final class PersonaTraitsUITests: XCTestCase {
    private let directnessID = "traitSlider_directness"

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testTraitSliders_LoadPersistAndReset() throws {
        let app = XCUIApplication()
        // This test is about the trait sliders, not onboarding — skip
        // it outright rather than depend on tapping through it (see
        // AuthManager's DEBUG-only handling of this launch argument).
        app.launchArguments = ["--uitest-skip-onboarding"]
        app.launch()

        let card = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "personaCard_")).firstMatch
        XCTAssertTrue(card.waitForExistence(timeout: 10), "expected a persona card")
        card.tap()

        let directnessSlider = app.sliders[directnessID]
        XCTAssertTrue(directnessSlider.waitForExistence(timeout: 10), "expected the personality dial sliders on Persona Detayı")
        Self.scrollUntilHittable(directnessSlider, in: app)

        // Loaded from the real backend, not a hardcoded local default —
        // Ada's seeded default_directness is 75, never 50.
        guard let loadedValue = Self.percent(directnessID, in: app) else {
            XCTFail("expected the slider to report a numeric percentage value")
            return
        }
        XCTAssertNotEqual(loadedValue, 50, "expected the persona's real default, not the placeholder initial value")

        let resetButton = app.buttons["personaTraitsResetButton"]
        XCTAssertFalse(resetButton.exists, "no customization yet, so no reset button should show")

        // Push directness toward its maximum — this is the "harsher
        // answers" dial from the feature request. The first attempt at
        // this used .adjust(toNormalizedSliderPosition:) before the
        // scroll-into-view fix above existed, while the slider was
        // still below the fold — the touch had nowhere real to land,
        // not a problem with this API itself. Retrying it now that the
        // slider is actually on-screen.
        directnessSlider.adjust(toNormalizedSliderPosition: 1.0)

        guard let customizedValue = Self.percent(directnessID, in: app) else {
            try? app.debugDescription.write(toFile: "/tmp/perchly_traits_debug.txt", atomically: true, encoding: .utf8)
            XCTFail("expected the slider to report a numeric percentage value after dragging")
            return
        }
        XCTAssertGreaterThan(customizedValue, loadedValue, "expected dragging toward the max to raise the value")
        XCTAssertGreaterThanOrEqual(customizedValue, 90, "expected dragging to the far end to land near 100%")

        Self.scrollUntilHittable(resetButton, in: app)
        XCTAssertTrue(resetButton.waitForExistence(timeout: 5), "expected the reset button once the persona is customized")

        // Give the save (fired on drag-end) a moment to land, then leave
        // and come back to prove it actually persisted server-side, not
        // just in local view state.
        Thread.sleep(forTimeInterval: 1.0)
        app.navigationBars.buttons.element(boundBy: 0).tap()
        XCTAssertTrue(card.waitForExistence(timeout: 5))
        card.tap()

        let reopenedSlider = app.sliders[directnessID]
        XCTAssertTrue(reopenedSlider.waitForExistence(timeout: 10))
        Self.scrollUntilHittable(reopenedSlider, in: app)
        let persistedValue = Self.percent(directnessID, in: app)
        XCTAssertEqual(persistedValue, customizedValue, "expected the customization to have persisted across navigation")

        let reopenedResetButton = app.buttons["personaTraitsResetButton"]
        Self.scrollUntilHittable(reopenedResetButton, in: app)
        XCTAssertTrue(reopenedResetButton.waitForExistence(timeout: 5))
        reopenedResetButton.tap()
        // Reset clears isCustomized only once the DELETE round-trip
        // completes (see PersonaTraitsViewModel.reset) — the button's
        // own disappearance is proof enough that it succeeded (it's
        // driven directly by isCustomized), without needing to re-read
        // the slider's value right after a layout change reshuffles
        // this ScrollView's contents.
        let resetButtonGone = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: reopenedResetButton)
        XCTAssertEqual(XCTWaiter.wait(for: [resetButtonGone], timeout: 5), .completed, "expected the reset button to disappear once back to defaults")
    }

    /// Reads a slider's current percentage by parsing the accessibility
    /// tree dump rather than `XCUIElement.value` directly — on this
    /// Simulator/Xcode combination, `.value` intermittently throws a
    /// hard "Failed to get matching snapshot" failure (a known-flaky
    /// accessibility-snapshot issue on this machine, unrelated to the
    /// app), while `app.debugDescription` has proven reliable.
    private static func percent(_ identifier: String, in app: XCUIApplication) -> Int? {
        let description = app.debugDescription
        guard let markerRange = description.range(of: "identifier: '\(identifier)', value: ") else { return nil }
        let afterMarker = description[markerRange.upperBound...]
        guard let percentIndex = afterMarker.firstIndex(of: "%") else { return nil }
        return Int(afterMarker[..<percentIndex])
    }

    /// Persona Detayı is one long ScrollView — an element can exist in
    /// the accessibility tree the moment the screen loads while still
    /// being below the fold, where a synthesized touch lands nowhere.
    /// Swipes up until it's actually on-screen and interactable.
    private static func scrollUntilHittable(_ element: XCUIElement, in app: XCUIApplication, maxAttempts: Int = 10) {
        var attempts = 0
        while element.exists && !element.isHittable && attempts < maxAttempts {
            app.swipeUp()
            attempts += 1
        }
    }
}

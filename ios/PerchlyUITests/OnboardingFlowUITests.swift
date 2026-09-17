import XCTest

/// Drives the full onboarding data-collection flow end to end, on a
/// fresh install (no prior session), against a locally running backend
/// (see backend/scripts/dev.sh): age range -> mood preference -> the
/// real native notification permission dialog -> persona pick -> chat.
final class OnboardingFlowUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testOnboardingFlow_CollectsAnswersAndReachesChat() throws {
        let app = XCUIApplication()

        // Auto-allow the native notification permission dialog the
        // moment it appears — it's presented by springboard, not this
        // app, so it needs its own interruption handling.
        addUIInterruptionMonitor(withDescription: "Notification permission") { alert in
            let allow = alert.buttons["Allow"]
            if allow.exists {
                allow.tap()
                return true
            }
            return false
        }

        app.launch()

        // --- Screen 1: age range (mandatory) ---
        let continueButton = app.buttons["ageContinueButton"]
        XCTAssertTrue(continueButton.waitForExistence(timeout: 10))
        XCTAssertFalse(continueButton.isEnabled, "Continue should be disabled with nothing selected yet")

        let ageOption = app.buttons["ageOption_age25to34"]
        XCTAssertTrue(ageOption.waitForExistence(timeout: 5))
        ageOption.tap()
        XCTAssertTrue(continueButton.isEnabled, "Continue should enable once an age range is selected")
        continueButton.tap()

        // --- Screen 2: mood preference (optional; pick one this run) ---
        let motivationCard = app.buttons["moodOption_motivation"]
        XCTAssertTrue(motivationCard.waitForExistence(timeout: 5))
        motivationCard.tap()

        // --- Screen 3: notification permission ---
        let requestButton = app.buttons["requestNotificationButton"]
        XCTAssertTrue(requestButton.waitForExistence(timeout: 5))
        requestButton.tap()
        // Nudge the interruption monitor to fire, then wait for the
        // flow to move past the alert on its own.
        app.tap()

        // --- Screen 4: persona pick, prioritized by mood -> should
        // show the motivational-coach persona first (Ada, per the
        // seed data) since we picked "Motive Olmak" above. ---
        let personaButtons = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "personaOption_"))
        let firstPersona = personaButtons.firstMatch
        XCTAssertTrue(firstPersona.waitForExistence(timeout: 10), "expected at least one persona option")
        XCTAssertTrue(firstPersona.label.contains("Ada"), "expected the motivational coach first for a 'motivation' mood preference, got: \(firstPersona.label)")
        firstPersona.tap()

        // --- Completion: straight to chat, no extra screens ---
        let chatField = app.textFields["chatMessageField"]
        if !chatField.waitForExistence(timeout: 10) {
            try? app.debugDescription.write(toFile: "/tmp/perchly_ui_debug.txt", atomically: true, encoding: .utf8)
        }
        XCTAssertTrue(chatField.exists, "expected to land directly in ChatView after picking a persona")

        // Onboarding screens must be gone for good.
        XCTAssertFalse(app.buttons["ageContinueButton"].exists)
        XCTAssertFalse(personaButtons.firstMatch.exists)

        // The user must be able to navigate back to the persona list —
        // not stuck looking at an isolated ChatView with no way out
        // until a full relaunch (a real bug this test caught once).
        app.navigationBars.buttons.element(boundBy: 0).tap()
        if !app.buttons["profileButton"].waitForExistence(timeout: 5) {
            try? app.debugDescription.write(toFile: "/tmp/perchly_ui_debug2.txt", atomically: true, encoding: .utf8)
        }
        XCTAssertTrue(app.buttons["profileButton"].exists, "expected a working back button into the persona list after onboarding")
    }
}

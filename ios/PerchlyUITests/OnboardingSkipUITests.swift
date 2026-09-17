import XCTest

/// Verifies the "don't show onboarding again" rule: once
/// `has_completed_onboarding` is true (persisted from a prior run — see
/// OnboardingFlowUITests, which must run first in the same simulator
/// install), a fresh launch goes straight to the persona list, never
/// back through onboarding.
final class OnboardingSkipUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testRelaunch_SkipsOnboardingWhenAlreadyCompleted() throws {
        let app = XCUIApplication()
        app.launch()

        let profileButton = app.buttons["profileButton"]
        XCTAssertTrue(profileButton.waitForExistence(timeout: 10), "expected to land directly on the persona list, not onboarding")
        XCTAssertFalse(app.buttons["ageContinueButton"].exists, "onboarding must not reappear once completed")
    }

    func testPersonaCard_OpensDetailThenChat() throws {
        let app = XCUIApplication()
        addUIInterruptionMonitor(withDescription: "Notification permission") { alert in
            let allow = alert.buttons["Allow"]
            if allow.exists {
                allow.tap()
                return true
            }
            return false
        }
        app.launch()
        fastForwardThroughOnboardingIfPresented(app)

        XCTAssertTrue(app.buttons["profileButton"].waitForExistence(timeout: 10))

        let card = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "personaCard_")).firstMatch
        XCTAssertTrue(card.waitForExistence(timeout: 10), "expected a persona card on Keşfet")
        card.tap()

        let startChat = app.buttons["personaDetailStartChat"]
        XCTAssertTrue(startChat.waitForExistence(timeout: 5), "card tap should open Persona Detayı")
        startChat.tap()

        XCTAssertTrue(app.textFields["chatMessageField"].waitForExistence(timeout: 10), "detail CTA should open chat")

        app.navigationBars.buttons.element(boundBy: 0).tap()
        XCTAssertTrue(startChat.waitForExistence(timeout: 5), "back from chat should return to Persona Detayı")

        app.navigationBars.buttons.element(boundBy: 0).tap()
        XCTAssertTrue(app.buttons["profileButton"].waitForExistence(timeout: 5), "back from detail should return to Keşfet")
    }
}

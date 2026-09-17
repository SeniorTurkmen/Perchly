import XCTest

/// Verifies the debug-only "reset local data" button (see ProfileView,
/// AuthManager.debugResetLocalSession) actually wipes the session and
/// bootstraps a fresh anonymous one — the whole point of this button is
/// to avoid needing a full Simulator erase during development (`simctl
/// uninstall` alone does NOT clear Keychain data on iOS).
final class DebugResetUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testResetButton_ReturnsToFreshAnonymousState() throws {
        let app = XCUIApplication()
        addUIInterruptionMonitor(withDescription: "Notification permission") { alert in
            guard alert.buttons["Allow"].exists else { return false }
            alert.buttons["Allow"].tap()
            return true
        }
        app.launch()
        fastForwardThroughOnboardingIfPresented(app)

        let profileButton = app.buttons["profileButton"]
        XCTAssertTrue(profileButton.waitForExistence(timeout: 10))
        profileButton.tap()

        let resetButton = app.buttons["debugResetButton"]
        XCTAssertTrue(resetButton.waitForExistence(timeout: 5), "expected the debug-only reset button in a Debug build")
        resetButton.tap()

        // Confirm the destructive alert.
        let confirmButton = app.buttons["Sıfırla"]
        XCTAssertTrue(confirmButton.waitForExistence(timeout: 5))
        confirmButton.tap()

        // Whatever state we were in before (anonymous or authenticated),
        // a reset must land back on the anonymous "link account" card.
        let linkCard = app.buttons["linkAccountCard"]
        XCTAssertTrue(linkCard.waitForExistence(timeout: 10), "expected a fresh anonymous session after reset")
        XCTAssertFalse(app.staticTexts["accountEmailLabel"].exists, "no email should remain linked after a reset")
    }
}

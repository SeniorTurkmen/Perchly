import XCTest

extension XCTestCase {
    /// A fresh install always lands on onboarding now (it gates the main
    /// app — see RootView). Tests that care about something else
    /// (auth, profile, ...) call this first to get through it as fast
    /// as possible; it no-ops if onboarding isn't showing (e.g. this
    /// install already completed it in an earlier test).
    func fastForwardThroughOnboardingIfPresented(_ app: XCUIApplication) {
        if app.buttons["skipNameAnonymousButton"].waitForExistence(timeout: 3) {
            app.buttons["skipNameAnonymousButton"].tap()
        }

        let ageOption = app.buttons["ageOption_age25to34"]
        guard ageOption.waitForExistence(timeout: 5) else { return }

        ageOption.tap()
        app.buttons["ageContinueButton"].tap()

        let skipMood = app.buttons["skipMoodButton"]
        XCTAssertTrue(skipMood.waitForExistence(timeout: 5))
        skipMood.tap()

        let requestButton = app.buttons["requestNotificationButton"]
        XCTAssertTrue(requestButton.waitForExistence(timeout: 5))
        requestButton.tap()
        app.tap() // nudges the (auto-handled, via addUIInterruptionMonitor) system permission alert along

        let personaButtons = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "personaOption_"))
        let firstPersona = personaButtons.firstMatch
        XCTAssertTrue(firstPersona.waitForExistence(timeout: 10))
        firstPersona.tap()

        // Onboarding hands off into chat with the picked persona,
        // inside the real persona-list navigation stack (see
        // RootView/PersonaListView) — go back to the list so the
        // caller's own test starts from a predictable place.
        XCTAssertTrue(app.textFields["chatMessageField"].waitForExistence(timeout: 10))
        app.navigationBars.buttons.element(boundBy: 0).tap()
    }
}

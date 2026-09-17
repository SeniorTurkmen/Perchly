import XCTest

/// Drives the real account-linking flow end to end against a locally
/// running backend (see backend/scripts/dev.sh), reading the emailed
/// verification code straight out of the backend's log file — the
/// backend logs it via ConsoleSender when SMTP isn't configured, which
/// is exactly the local dev setup this test expects.
///
/// Requires the backend log to be redirected to `logPath` below (e.g.
/// `go run ./cmd/api > /tmp/perchly_api_iosauth.log 2>&1`). Not meant
/// for CI — a manual verification harness for this flow.
final class AuthFlowUITests: XCTestCase {
    private let logPath = "/tmp/perchly_api_iosauth.log"

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testEmailLinkingFlow_NewRegistration() throws {
        let app = XCUIApplication()
        addUIInterruptionMonitor(withDescription: "Notification permission") { alert in
            guard alert.buttons["Allow"].exists else { return false }
            alert.buttons["Allow"].tap()
            return true
        }
        app.launch()
        fastForwardThroughOnboardingIfPresented(app)

        let profileButton = app.buttons["profileButton"]
        XCTAssertTrue(profileButton.waitForExistence(timeout: 10), "profile button never appeared")
        profileButton.tap()

        let linkCard = app.buttons["linkAccountCard"]
        XCTAssertTrue(linkCard.waitForExistence(timeout: 5), "expected the anonymous 'link account' card")
        linkCard.tap()

        let emailField = app.textFields["emailField"]
        XCTAssertTrue(emailField.waitForExistence(timeout: 5))
        emailField.tap()

        let uniqueEmail = "uitest-\(Int(Date().timeIntervalSince1970))@example.com"
        emailField.typeText(uniqueEmail)

        let sendButton = app.buttons["sendCodeButton"]
        XCTAssertTrue(sendButton.waitForExistence(timeout: 5))
        sendButton.tap()

        let codeField = app.textFields["codeEntryField"]
        XCTAssertTrue(codeField.waitForExistence(timeout: 10), "never navigated to CodeVerificationView")

        let code = try waitForCode(forEmail: uniqueEmail, timeout: 10)
        codeField.typeText(code)

        // A never-before-seen email must be Durum B / new registration.
        let displayNameField = app.textFields["displayNameField"]
        XCTAssertTrue(displayNameField.waitForExistence(timeout: 10), "expected RegisterCompleteView for a brand-new email")
        displayNameField.tap()
        displayNameField.typeText("Test Kullanıcı")

        let completeButton = app.buttons["completeRegistrationButton"]
        completeButton.tap()

        // Success dismisses the whole linking sheet back to the
        // already-open ProfileView underneath, which should now show
        // the linked email instead of the "link your account" card.
        let emailLabel = app.staticTexts["accountEmailLabel"]
        XCTAssertTrue(emailLabel.waitForExistence(timeout: 10), "ProfileView never showed the linked email")
        XCTAssertEqual(emailLabel.label, uniqueEmail)
    }

    private func waitForCode(forEmail email: String, timeout: TimeInterval) throws -> String {
        let marker = "verification code for \(email): "
        let deadline = Date().addingTimeInterval(timeout)

        while Date() < deadline {
            if let contents = try? String(contentsOfFile: logPath, encoding: .utf8),
               let range = contents.range(of: marker) {
                let code = contents[range.upperBound...].prefix { $0.isNumber }
                if code.count == 6 {
                    return String(code)
                }
            }
            Thread.sleep(forTimeInterval: 0.5)
        }

        XCTFail("timed out waiting for the verification code to appear in \(logPath)")
        return ""
    }
}

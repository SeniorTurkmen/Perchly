import XCTest

/// Verifies Durum A on the client: linking an email that's already a
/// verified account on a DIFFERENT (fresh) anonymous session must skip
/// RegisterCompleteView entirely and land straight on the authenticated
/// ProfileView. Requires an email already verified by a prior run of
/// AuthFlowUITests (or any prior successful registration) — see
/// `existingVerifiedEmail`.
final class AuthFlowDurumAUITests: XCTestCase {
    private let logPath = "/tmp/perchly_api_iosauth.log"

    /// Must be an email that has ALREADY completed registration once
    /// (e.g. from a prior AuthFlowUITests run) so this run hits Durum A
    /// (existing verified account) rather than Durum B (new account).
    private var existingVerifiedEmail: String {
        AuthFlowDurumAUITests.readLastRegisteredEmail(logPath: logPath) ?? "uitest-durum-a-seed@example.com"
    }

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testEmailLinkingFlow_ExistingAccountSkipsRegistration() throws {
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
        emailField.typeText(existingVerifiedEmail)

        app.buttons["sendCodeButton"].tap()

        let codeField = app.textFields["codeEntryField"]
        XCTAssertTrue(codeField.waitForExistence(timeout: 10))

        let code = try waitForCode(forEmail: existingVerifiedEmail, timeout: 10)
        codeField.typeText(code)

        // Durum A: must go straight to the authenticated profile — never
        // RegisterCompleteView.
        let emailLabel = app.staticTexts["accountEmailLabel"]
        XCTAssertTrue(emailLabel.waitForExistence(timeout: 10), "Durum A should skip RegisterCompleteView entirely")
        XCTAssertEqual(emailLabel.label, existingVerifiedEmail)
        XCTAssertFalse(app.textFields["displayNameField"].exists, "RegisterCompleteView must not appear for an existing verified account")
    }

    private func waitForCode(forEmail email: String, timeout: TimeInterval) throws -> String {
        let marker = "verification code for \(email): "
        let deadline = Date().addingTimeInterval(timeout)

        while Date() < deadline {
            if let contents = try? String(contentsOfFile: logPath, encoding: .utf8),
               let range = contents.range(of: marker, options: .backwards) {
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

    private static func readLastRegisteredEmail(logPath: String) -> String? {
        guard let contents = try? String(contentsOfFile: logPath, encoding: .utf8) else { return nil }
        // AuthFlowUITests's emails look like "uitest-1234567890@example.com".
        let pattern = #"uitest-\d+@example\.com"#
        guard let regex = try? NSRegularExpression(pattern: pattern) else { return nil }
        let matches = regex.matches(in: contents, range: NSRange(contents.startIndex..., in: contents))
        guard let last = matches.last, let range = Range(last.range, in: contents) else { return nil }
        return String(contents[range])
    }
}

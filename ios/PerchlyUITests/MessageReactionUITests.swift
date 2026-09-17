import XCTest

/// Drives the iMessage-style emoji reaction end to end, on a real chat
/// turn against a locally running backend (see backend/scripts/dev.sh):
/// send a message, wait for the assistant's real reply bubble, long-press
/// it to reveal the reaction picker, pick an emoji, and verify the
/// reaction badge appears on the bubble.
final class MessageReactionUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testUserCanReactToAssistantMessageViaLongPress() throws {
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

        // fastForwardThroughOnboardingIfPresented leaves the app on the
        // persona list; open a persona and start (or resume) a chat.
        let card = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "personaCard_")).firstMatch
        XCTAssertTrue(card.waitForExistence(timeout: 10), "expected a persona card")
        card.tap()

        let startChat = app.buttons["personaDetailStartChat"]
        if startChat.waitForExistence(timeout: 3) {
            startChat.tap()
        }

        let field = app.textFields["chatMessageField"]
        XCTAssertTrue(field.waitForExistence(timeout: 10), "expected the chat composer")
        field.tap()
        field.typeText("Merhaba, bugün harika bir gün!\n")

        // Wait for the assistant's real reply bubble — its long-press
        // target carries a stable "assistantBubble_" prefixed id (see
        // ChatBubble). Streamed text arrives progressively, so give this
        // plenty of time. This install's conversation may already carry
        // history from earlier test runs, so target the LAST match
        // (this turn's reply), not the first (some older message).
        let assistantBubbleQuery = app.descendants(matching: .any)
            .matching(NSPredicate(format: "identifier BEGINSWITH %@", "assistantBubble_"))
        let waitExpectation = XCTNSPredicateExpectation(
            predicate: NSPredicate(format: "count > 0"),
            object: assistantBubbleQuery
        )
        XCTAssertEqual(XCTWaiter.wait(for: [waitExpectation], timeout: 30), .completed, "expected an assistant reply bubble")
        let assistantBubble = assistantBubbleQuery.element(boundBy: assistantBubbleQuery.count - 1)

        // The bubble's accessibility id exists from the moment the empty
        // streaming placeholder is appended — well before the reply
        // finishes arriving. The long-press picker is deliberately
        // disabled while a message is still streaming (see
        // ChatBubble.beginReacting), so retry the long press until the
        // turn has fully settled instead of guessing a fixed delay.
        let firstReactionOption = app.buttons["reactionOption_0"]
        var pickerAppeared = false
        for _ in 0..<10 {
            assistantBubble.press(forDuration: 0.6)
            if firstReactionOption.waitForExistence(timeout: 2) {
                pickerAppeared = true
                break
            }
        }
        if !pickerAppeared {
            try? app.debugDescription.write(toFile: "/tmp/perchly_reaction_debug.txt", atomically: true, encoding: .utf8)
        }
        XCTAssertTrue(pickerAppeared, "expected the reaction picker to appear on long press once the reply finished streaming")
        firstReactionOption.tap()

        let badge = app.descendants(matching: .any)
            .matching(NSPredicate(format: "identifier BEGINSWITH %@", "reactionBadge_"))
            .firstMatch
        XCTAssertTrue(badge.waitForExistence(timeout: 5), "expected a reaction badge on the bubble after picking one")

        // Picking the same emoji again clears it (toggle-off).
        assistantBubble.press(forDuration: 0.6)
        XCTAssertTrue(app.buttons["reactionOption_0"].waitForExistence(timeout: 5))
        app.buttons["reactionOption_0"].tap()
        XCTAssertFalse(badge.waitForExistence(timeout: 3), "expected the badge to clear after re-picking the same emoji")
    }
}

import XCTest
@testable import Perchly

/// Exercises the client-side half of the backend's global error model
/// (internal/apierror): decoding `{"error", "code"}` bodies into
/// `APIErrorCode`, and a real call site (CodeVerificationViewModel)
/// branching on the decoded code instead of matching message text.
final class APIErrorCodeTests: XCTestCase {
    private struct Payload: Decodable {
        let error: String
        let code: APIErrorCode?
    }

    func testDecodesAKnownCode() throws {
        let json = #"{"error":"kodun süresi doldu, yeni kod isteyin","code":"verification_code_expired"}"#
        let payload = try JSONDecoder().decode(Payload.self, from: Data(json.utf8))
        XCTAssertEqual(payload.code, .verificationCodeExpired)
    }

    /// A code this build doesn't recognize yet (e.g. the backend shipped
    /// a new one) must decode to `.unknown`, not fail the whole response.
    func testUnrecognizedCodeDecodesToUnknown() throws {
        let json = #"{"error":"bir şey oldu","code":"some_future_code_this_app_has_never_heard_of"}"#
        let payload = try JSONDecoder().decode(Payload.self, from: Data(json.utf8))
        XCTAssertEqual(payload.code, .unknown)
    }

    /// A body from an older backend deploy, without `code` at all, must
    /// still decode — `code` is optional precisely for this.
    func testMissingCodeFieldDecodesToNil() throws {
        let json = #"{"error":"giriş gerekli"}"#
        let payload = try JSONDecoder().decode(Payload.self, from: Data(json.utf8))
        XCTAssertNil(payload.code)
    }

    @MainActor
    func testCodeVerificationViewModel_ClassifiesByCodeNotMessageText() {
        let viewModel = CodeVerificationViewModel(email: "test@example.com")

        let expired = APIError.message(statusCode: 401, code: .verificationCodeExpired, text: "kodun süresi doldu, yeni kod isteyin")
        XCTAssertEqual(viewModel.classify(expired).0, .expired)

        let wrongCode = APIError.message(statusCode: 401, code: .invalidVerificationCode, text: "kod geçersiz")
        XCTAssertEqual(viewModel.classify(wrongCode).0, .entering)

        let tooMany = APIError.message(statusCode: 429, code: .tooManyAttempts, text: "çok fazla hatalı deneme, yeni kod isteyin")
        XCTAssertEqual(viewModel.classify(tooMany).0, .expired)

        // A recognized status but an APIErrorCode this classify doesn't
        // special-case: must fall back to the calm generic message, not
        // crash or misclassify as expired.
        let somethingElse = APIError.message(statusCode: 401, code: .unknown, text: "?")
        XCTAssertEqual(viewModel.classify(somethingElse).0, .entering)
    }
}

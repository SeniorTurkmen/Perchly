import Foundation

enum APIError: Error, LocalizedError {
    case invalidURL
    case invalidResponse
    case server(statusCode: Int, data: Data?)
    case decoding(Error)
    case transport(Error)
    /// A server-provided error, i.e. the `{"error": "...", "code": "..."}`
    /// body our handlers return on 4xx/5xx responses — see
    /// `internal/apierror` on the backend. Carries the status code and
    /// the machine-readable `code` alongside the display text, so
    /// callers can branch on either (e.g. `code == .verificationCodeExpired`)
    /// without re-parsing or matching on the Turkish text.
    case message(statusCode: Int, code: APIErrorCode?, text: String)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return String(localized: "Geçersiz URL.")
        case .invalidResponse:
            return String(localized: "Sunucudan geçersiz yanıt alındı.")
        case .server(let statusCode, _):
            return String(localized: "Sunucu hatası (\(String(statusCode))).")
        case .decoding:
            return String(localized: "Yanıt çözümlenemedi.")
        case .transport(let error):
            return error.localizedDescription
        case .message(_, _, let text):
            return text
        }
    }

    /// The HTTP status code, when this error came from a non-2xx
    /// response — nil for transport/decoding/local errors.
    var statusCode: Int? {
        switch self {
        case .server(let statusCode, _): return statusCode
        case .message(let statusCode, _, _): return statusCode
        default: return nil
        }
    }

    /// The backend's machine-readable failure reason, when this error
    /// came from a JSON `{"error", "code"}` body — nil for
    /// transport/decoding/local errors, or a body an older backend
    /// sent without a `code` field.
    var code: APIErrorCode? {
        switch self {
        case .message(_, let code, _): return code
        default: return nil
        }
    }
}

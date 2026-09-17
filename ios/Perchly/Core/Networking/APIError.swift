import Foundation

enum APIError: Error, LocalizedError {
    case invalidURL
    case invalidResponse
    case server(statusCode: Int, data: Data?)
    case decoding(Error)
    case transport(Error)
    /// A server-provided error message, e.g. the `{"error": "..."}` body
    /// our handlers return on 4xx/5xx responses. Carries the status code
    /// alongside the text so callers can branch on it (e.g. 429 vs 401)
    /// without re-parsing anything.
    case message(statusCode: Int, text: String)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Geçersiz URL."
        case .invalidResponse:
            return "Sunucudan geçersiz yanıt alındı."
        case .server(let statusCode, _):
            return "Sunucu hatası (\(statusCode))."
        case .decoding:
            return "Yanıt çözümlenemedi."
        case .transport(let error):
            return error.localizedDescription
        case .message(_, let text):
            return text
        }
    }

    /// The HTTP status code, when this error came from a non-2xx
    /// response — nil for transport/decoding/local errors.
    var statusCode: Int? {
        switch self {
        case .server(let statusCode, _): return statusCode
        case .message(let statusCode, _): return statusCode
        default: return nil
        }
    }
}

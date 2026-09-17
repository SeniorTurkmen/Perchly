import Foundation

/// One parsed Server-Sent Event: an `event:` name (defaults to "message"
/// per the SSE spec when absent) and its `data:` payload, joined back
/// together if the payload spanned multiple `data:` lines.
struct SSEEvent {
    let event: String
    let data: String

    /// The backend always sends `data:` as a JSON-encoded string (e.g.
    /// `data: "Merhaba, "`), so it can carry arbitrary text (quotes,
    /// newlines) safely inside the single-line SSE format. This decodes
    /// that JSON string back to plain text.
    func decodedText() throws -> String {
        try JSONDecoder().decode(String.self, from: Data(data.utf8))
    }
}

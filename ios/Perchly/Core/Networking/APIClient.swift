import Foundation

/// Minimal async/await HTTP client that automatically attaches the
/// current access token (if any) from `TokenStorage` to every outgoing
/// request, and transparently recovers from an expired one.
///
/// All stored properties are immutable and their types are themselves
/// thread-safe, so it is safe to share a single instance across tasks.
final class APIClient: @unchecked Sendable {
    static let shared = APIClient()

    private let baseURL: URL
    private let session: URLSession
    private let tokenStorage: TokenStorage

    init(
        baseURL: URL = APIClient.defaultBaseURL,
        session: URLSession = .shared,
        tokenStorage: TokenStorage = TokenStorage()
    ) {
        self.baseURL = baseURL
        self.session = session
        self.tokenStorage = tokenStorage
    }

    private static var defaultBaseURL: URL {
        URL(string: "http://localhost:8080")!
    }

    /// Decodes JSON responses. Go's `time.Time` marshals to RFC 3339 with
    /// fractional seconds (e.g. "2026-09-14T14:17:59.890198+03:00"), which
    /// the plain `.iso8601` strategy can't parse, so dates are decoded with
    /// formatters that try fractional seconds first.
    private static let jsonDecoder: JSONDecoder = {
        let decoder = JSONDecoder()

        // Configured once and only ever read from afterwards, so sharing
        // these across the @Sendable decoding closure's invocations is safe
        // even though ISO8601DateFormatter itself isn't Sendable.
        nonisolated(unsafe) let withFractionalSeconds = ISO8601DateFormatter()
        withFractionalSeconds.formatOptions = [.withInternetDateTime, .withFractionalSeconds]

        nonisolated(unsafe) let withoutFractionalSeconds = ISO8601DateFormatter()
        withoutFractionalSeconds.formatOptions = [.withInternetDateTime]

        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            if let date = withFractionalSeconds.date(from: dateString) {
                return date
            }
            if let date = withoutFractionalSeconds.date(from: dateString) {
                return date
            }
            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Beklenmeyen tarih formatı: \(dateString)"
            )
        }

        return decoder
    }()

    /// Sends a request and decodes the response body as `Response`.
    ///
    /// - Parameter skipAuthRetry: Pass `true` for calls that must never
    ///   trigger the 401-refresh interceptor — specifically
    ///   `AuthManager`'s own `/auth/refresh` call. Retrying a failed
    ///   refresh by... refreshing again would call back into the very
    ///   function this request is already running inside of. Every
    ///   other request should leave this `false`.
    func send<Response: Decodable>(_ request: APIRequest, skipAuthRetry: Bool = false) async throws -> Response {
        let data = try await performWithRefreshRetry(request, skipAuthRetry: skipAuthRetry)
        do {
            return try Self.jsonDecoder.decode(Response.self, from: data)
        } catch {
            throw APIError.decoding(error)
        }
    }

    /// Sends a request that has no meaningful response body (e.g. 204s).
    func send(_ request: APIRequest, skipAuthRetry: Bool = false) async throws {
        _ = try await performWithRefreshRetry(request, skipAuthRetry: skipAuthRetry)
    }

    /// Performs `request`, and if the first attempt comes back 401,
    /// asks `AuthManager` to refresh the session and retries exactly
    /// once with the new access token. If refreshing instead falls back
    /// to a brand-new anonymous session, the original 401 is what gets
    /// thrown — replaying the same request under a different identity
    /// wouldn't be meaningful (e.g. it may no longer own the resource it
    /// was addressing), so that's left to the caller to handle.
    private func performWithRefreshRetry(_ request: APIRequest, skipAuthRetry: Bool) async throws -> Data {
        let (data, httpResponse) = try await performOnce(request)

        guard httpResponse.statusCode == 401, !skipAuthRetry else {
            return try Self.unwrapSuccess(data: data, statusCode: httpResponse.statusCode)
        }

        guard await AuthManager.shared.refreshOrFallbackToAnonymous() else {
            return try Self.unwrapSuccess(data: data, statusCode: httpResponse.statusCode)
        }

        let (retryData, retryResponse) = try await performOnce(request)
        return try Self.unwrapSuccess(data: retryData, statusCode: retryResponse.statusCode)
    }

    private func performOnce(_ request: APIRequest) async throws -> (Data, HTTPURLResponse) {
        let urlRequest = try makeURLRequest(for: request)
        let (data, response) = try await perform(urlRequest)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        return (data, httpResponse)
    }

    private static func unwrapSuccess(data: Data, statusCode: Int) throws -> Data {
        guard (200..<300).contains(statusCode) else {
            throw apiError(forStatus: statusCode, data: data)
        }
        return data
    }

    /// `code` is optional so a body from an older backend deploy
    /// without it still decodes — see APIError.message's doc.
    private struct ErrorPayload: Decodable {
        let error: String
        let code: APIErrorCode?
    }

    private static func apiError(forStatus statusCode: Int, data: Data) -> APIError {
        if let payload = try? JSONDecoder().decode(ErrorPayload.self, from: data) {
            return .message(statusCode: statusCode, code: payload.code, text: payload.error)
        }
        return .server(statusCode: statusCode, data: data)
    }

    /// Streams Server-Sent Events for a request, e.g. the chat message
    /// endpoint. Retries once on an initial 401 exactly like `send`
    /// (before any bytes are streamed, so it's just as safe to replay).
    func streamEvents(_ request: APIRequest) -> AsyncThrowingStream<SSEEvent, Error> {
        AsyncThrowingStream { continuation in
            let task = Task {
                do {
                    func open() async throws -> (URLSession.AsyncBytes, HTTPURLResponse) {
                        let urlRequest = try makeURLRequest(for: request)
                        let (byteStream, response) = try await session.bytes(for: urlRequest)
                        guard let httpResponse = response as? HTTPURLResponse else {
                            throw APIError.invalidResponse
                        }
                        return (byteStream, httpResponse)
                    }

                    var (byteStream, httpResponse) = try await open()
                    if httpResponse.statusCode == 401, await AuthManager.shared.refreshOrFallbackToAnonymous() {
                        (byteStream, httpResponse) = try await open()
                    }

                    guard (200..<300).contains(httpResponse.statusCode) else {
                        throw try await Self.readError(from: byteStream, statusCode: httpResponse.statusCode)
                    }

                    // Deliberately not using `byteStream.lines`: AsyncLineSequence
                    // collapses the blank line SSE uses to separate events, which
                    // merges every event in a response into one. Buffering raw
                    // bytes and splitting on the literal "\n\n" separator avoids
                    // that, and matching on bytes (not decoded text) is UTF-8-safe
                    // since 0x0A only ever appears as a real newline, never inside
                    // a multi-byte UTF-8 sequence.
                    let eventSeparator = Data([0x0A, 0x0A])
                    var buffer = Data()

                    for try await byte in byteStream {
                        try Task.checkCancellation()
                        buffer.append(byte)
                        while let range = buffer.range(of: eventSeparator) {
                            let rawEvent = buffer.subdata(in: buffer.startIndex..<range.lowerBound)
                            buffer.removeSubrange(buffer.startIndex..<range.upperBound)
                            if let event = Self.parseSSEEvent(rawEvent) {
                                continuation.yield(event)
                            }
                        }
                    }
                    if let event = Self.parseSSEEvent(buffer) {
                        continuation.yield(event)
                    }
                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
            }

            continuation.onTermination = { _ in
                task.cancel()
            }
        }
    }

    private static func parseSSEEvent(_ raw: Data) -> SSEEvent? {
        guard let text = String(data: raw, encoding: .utf8), !text.isEmpty else { return nil }

        var eventType = "message"
        var dataLines: [String] = []
        for line in text.split(separator: "\n", omittingEmptySubsequences: false) {
            if line.hasPrefix("event:") {
                eventType = line.dropFirst(6).trimmingCharacters(in: .whitespaces)
            } else if line.hasPrefix("data:") {
                dataLines.append(String(line.dropFirst(5)).trimmingCharacters(in: .whitespaces))
            }
        }
        guard !dataLines.isEmpty else { return nil }
        return SSEEvent(event: eventType, data: dataLines.joined(separator: "\n"))
    }

    private static func readError(from byteStream: URLSession.AsyncBytes, statusCode: Int) async throws -> APIError {
        var data = Data()
        for try await byte in byteStream { data.append(byte) }
        return apiError(forStatus: statusCode, data: data)
    }

    private func perform(_ request: URLRequest) async throws -> (Data, URLResponse) {
        do {
            return try await session.data(for: request)
        } catch {
            throw APIError.transport(error)
        }
    }

    private func makeURLRequest(for request: APIRequest) throws -> URLRequest {
        guard let url = URL(string: request.path, relativeTo: baseURL) else {
            throw APIError.invalidURL
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = request.method.rawValue
        urlRequest.httpBody = request.body

        for (key, value) in request.headers {
            urlRequest.setValue(value, forHTTPHeaderField: key)
        }

        if let token = tokenStorage.accessToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        return urlRequest
    }
}

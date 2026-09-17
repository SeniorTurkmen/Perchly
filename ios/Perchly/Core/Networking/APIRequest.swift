import Foundation

/// Describes a single request to the backend, independent of `URLRequest`
/// so it stays easy to construct and unit test.
struct APIRequest {
    let path: String
    let method: HTTPMethod
    let headers: [String: String]
    let body: Data?

    init(
        path: String,
        method: HTTPMethod = .get,
        headers: [String: String] = ["Content-Type": "application/json"],
        body: Data? = nil
    ) {
        self.path = path
        self.method = method
        self.headers = headers
        self.body = body
    }
}

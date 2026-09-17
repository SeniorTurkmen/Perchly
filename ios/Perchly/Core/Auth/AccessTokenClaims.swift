import Foundation

/// The claims Perchly's backend puts in an access token (see the
/// backend's `internal/auth.AccessClaims`).
struct AccessTokenClaims: Decodable {
    let subject: String
    let isAnonymous: Bool

    enum CodingKeys: String, CodingKey {
        case subject = "sub"
        case isAnonymous = "is_anonymous"
    }

    /// Decodes the claims from a JWT's payload segment, without
    /// verifying its signature. That's deliberate, not an oversight:
    /// this is only ever used to read a claim for local UI/state
    /// purposes (which screen to show), never to make an authorization
    /// decision — the backend independently re-validates the token's
    /// signature on every request regardless of what the client
    /// believes about it.
    static func decode(from accessToken: String) -> AccessTokenClaims? {
        let segments = accessToken.split(separator: ".")
        guard segments.count >= 2 else { return nil }

        var base64 = String(segments[1])
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        while base64.count % 4 != 0 {
            base64 += "="
        }

        guard let data = Data(base64Encoded: base64) else { return nil }
        return try? JSONDecoder().decode(AccessTokenClaims.self, from: data)
    }
}

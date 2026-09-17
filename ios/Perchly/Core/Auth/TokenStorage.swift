import Foundation
import Security

enum KeychainError: Error {
    case unhandled(OSStatus)
}

/// Keychain-backed storage for the session state that must survive app
/// restarts and never belongs in UserDefaults: the access token, the
/// refresh token, the device id an anonymous session is pinned to, and
/// the linked account's email (for display). Every property here is a
/// plain Keychain read — no caching layer — so storage is always the
/// single source of truth.
struct TokenStorage {
    private let service: String

    init(service: String = Bundle.main.bundleIdentifier ?? "com.perchly.app") {
        self.service = service
    }

    var accessToken: String? { read(account: .accessToken) }
    var refreshToken: String? { read(account: .refreshToken) }

    /// The email of the currently linked account, for display (e.g. on
    /// the profile screen) across app restarts. Not part of the token
    /// pair itself, but stored the same way (Keychain, not UserDefaults)
    /// since it identifies the account rather than being disposable UI
    /// state. Nil for a purely anonymous session.
    var email: String? { read(account: .email) }

    func saveEmail(_ email: String) throws {
        try write(email, account: .email)
    }

    /// A stable per-install identifier an anonymous session is pinned
    /// to. Generated once, on first access, and never changes
    /// afterward — including across `clearSession()` calls, since a new
    /// anonymous session must still resolve back to the same backend
    /// user via this id (see AuthManager).
    var deviceID: String {
        if let existing = read(account: .deviceID) {
            return existing
        }
        let generated = UUID().uuidString
        try? write(generated, account: .deviceID)
        return generated
    }

    func saveSession(accessToken: String, refreshToken: String) throws {
        try write(accessToken, account: .accessToken)
        try write(refreshToken, account: .refreshToken)
    }

    /// Removes the current session's tokens and linked email — falling
    /// back to a new anonymous session means starting a fresh identity,
    /// so the old one's email no longer applies. Deliberately does not
    /// touch the device id — see its doc comment.
    func clearSession() {
        delete(account: .accessToken)
        delete(account: .refreshToken)
        delete(account: .email)
    }

    /// Wipes everything, including the device id — the next `deviceID`
    /// access generates a brand new one, so the following anonymous
    /// session reads as a completely unrelated install rather than a
    /// fallback for the current one. This is what "reset local data"
    /// (a debug-only affordance — see ProfileView) needs, distinct from
    /// `clearSession()`'s normal fallback-to-anonymous behavior.
    func clearAll() {
        clearSession()
        delete(account: .deviceID)
    }

    private enum Account: String {
        case accessToken = "perchly.access_token"
        case refreshToken = "perchly.refresh_token"
        case deviceID = "perchly.device_id"
        case email = "perchly.email"
    }

    private func read(account: Account) -> String? {
        var query = baseQuery(account: account)
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess, let data = result as? Data else {
            return nil
        }
        return String(data: data, encoding: .utf8)
    }

    private func write(_ value: String, account: Account) throws {
        let data = Data(value.utf8)
        let query = baseQuery(account: account)

        if SecItemCopyMatching(query as CFDictionary, nil) == errSecSuccess {
            let attributes: [String: Any] = [kSecValueData as String: data]
            let status = SecItemUpdate(query as CFDictionary, attributes as CFDictionary)
            guard status == errSecSuccess else { throw KeychainError.unhandled(status) }
        } else {
            var newItem = query
            newItem[kSecValueData as String] = data
            let status = SecItemAdd(newItem as CFDictionary, nil)
            guard status == errSecSuccess else { throw KeychainError.unhandled(status) }
        }
    }

    private func delete(account: Account) {
        SecItemDelete(baseQuery(account: account) as CFDictionary)
    }

    private func baseQuery(account: Account) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account.rawValue,
        ]
    }
}

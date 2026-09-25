import Foundation

@MainActor
final class EmailEntryViewModel: ObservableObject {
    @Published var email: String = ""
    @Published private(set) var isSending: Bool = false
    @Published var errorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }

    var normalizedEmail: String {
        email.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
    }

    var canSend: Bool {
        isValidEmail(normalizedEmail) && !isSending
    }

    /// Requests a code. Returns whether the caller should move on to
    /// CodeVerificationView.
    func sendCode() async -> Bool {
        guard canSend else { return false }
        isSending = true
        errorMessage = nil
        defer { isSending = false }

        do {
            try await authManager.requestEmailCode(email: normalizedEmail)
            return true
        } catch {
            errorMessage = calmMessage(for: error)
            return false
        }
    }

    private func isValidEmail(_ value: String) -> Bool {
        guard let atIndex = value.firstIndex(of: "@") else { return false }
        return atIndex != value.startIndex && value.index(after: atIndex) != value.endIndex
    }

    private func calmMessage(for error: Error) -> String {
        // The backend currently never actually returns this from this
        // endpoint — rate-limiting there is silent by design, specifically
        // so throttling can't be used to probe whether an email is
        // registered. This branch is kept ready in case that policy ever
        // changes, but today it's effectively unreachable.
        if let apiError = error as? APIError, apiError.code == .tooManyAttempts {
            return String(localized: "Çok fazla deneme yaptın, biraz sonra tekrar dene.")
        }
        return String(localized: "Bir şeyler ters gitti, birazdan tekrar dene.")
    }
}

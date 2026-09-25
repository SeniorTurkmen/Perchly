import Foundation

@MainActor
final class RegisterCompleteViewModel: ObservableObject {
    @Published var displayName: String = ""
    @Published private(set) var isSubmitting = false
    @Published var errorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }

    var canSubmit: Bool {
        !displayName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !isSubmitting
    }

    func submit() async {
        guard canSubmit else { return }
        isSubmitting = true
        errorMessage = nil
        defer { isSubmitting = false }

        do {
            try await authManager.completeRegistration(
                displayName: displayName.trimmingCharacters(in: .whitespacesAndNewlines)
            )
        } catch {
            if let apiError = error as? APIError, apiError.code == .invalidDisplayName {
                errorMessage = String(localized: "Bu görünen ad kullanılamıyor. Farklı bir ad dener misin?")
            } else {
                errorMessage = String(localized: "Bir şeyler ters gitti, birazdan tekrar dene.")
            }
        }
    }
}

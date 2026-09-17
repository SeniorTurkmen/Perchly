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
            errorMessage = "Bir şeyler ters gitti, birazdan tekrar dene."
        }
    }
}

import Foundation

@MainActor
final class CodeVerificationViewModel: ObservableObject {
    enum Phase: Equatable {
        case entering
        case verifying
        /// The code (or all of its attempts) is dead — either its 10
        /// minute window ran out, or the backend's attempt cap was hit,
        /// which the backend treats identically to expiry (rejects even
        /// the correct code until a new one is requested).
        case expired
    }

    let email: String

    @Published var code: String = ""
    @Published private(set) var phase: Phase = .entering
    @Published var errorMessage: String?
    @Published private(set) var isRequestingNewCode = false
    @Published private(set) var remainingSeconds: Int

    /// Set once verification succeeds; the view observes this to
    /// navigate onward (see didComplete/isNewRegistration below).
    @Published private(set) var didComplete = false
    @Published private(set) var isNewRegistration = false

    private let authManager: AuthManager
    private let validityDuration: TimeInterval = 10 * 60
    private var expiresAt: Date
    private var countdownTask: Task<Void, Never>?

    init(email: String, authManager: AuthManager = .shared) {
        self.email = email
        self.authManager = authManager
        self.expiresAt = Date().addingTimeInterval(validityDuration)
        self.remainingSeconds = Int(validityDuration)
        startCountdown()
    }

    deinit {
        countdownTask?.cancel()
    }

    var remainingFraction: Double {
        Double(remainingSeconds) / validityDuration
    }

    var formattedRemaining: String {
        String(format: "%d:%02d", remainingSeconds / 60, remainingSeconds % 60)
    }

    /// Called by CodeEntryField as soon as its 6th digit is entered.
    func verifyIfComplete(_ enteredCode: String) {
        guard enteredCode.count == 6, phase != .verifying else { return }
        Task { await verify(code: enteredCode) }
    }

    private func verify(code: String) async {
        phase = .verifying
        errorMessage = nil

        do {
            let isNew = try await authManager.verifyEmailCode(email: email, code: code)
            isNewRegistration = isNew
            didComplete = true
        } catch {
            let (nextPhase, message) = classify(error)
            phase = nextPhase
            self.code = ""
            errorMessage = message
        }
    }

    func requestNewCode() async {
        isRequestingNewCode = true
        errorMessage = nil
        defer { isRequestingNewCode = false }

        do {
            try await authManager.requestEmailCode(email: email)
            code = ""
            expiresAt = Date().addingTimeInterval(validityDuration)
            phase = .entering
            startCountdown()
        } catch {
            errorMessage = "Bir şeyler ters gitti, birazdan tekrar dene."
        }
    }

    private func startCountdown() {
        countdownTask?.cancel()
        countdownTask = Task { [weak self] in
            while let self, !Task.isCancelled {
                let remaining = Int(self.expiresAt.timeIntervalSinceNow.rounded(.up))
                self.remainingSeconds = max(0, remaining)
                if remaining <= 0 {
                    self.phase = .expired
                    break
                }
                try? await Task.sleep(for: .seconds(1))
            }
        }
    }

    /// Maps a failed verify attempt to (next phase, calm message).
    ///
    /// "Expired" vs. "wrong code" are both a plain 401 — they're told
    /// apart by the backend's machine-readable `code`
    /// (`.verificationCodeExpired` vs. `.invalidVerificationCode`), not
    /// by matching on the Turkish message text.
    ///
    /// Internal (not private) so APIErrorCodeTests can exercise it
    /// directly without a real network round trip.
    func classify(_ error: Error) -> (Phase, String) {
        guard let apiError = error as? APIError else {
            return (.entering, "Bir şeyler ters gitti, birazdan tekrar dene.")
        }

        switch apiError.code {
        case .tooManyAttempts:
            return (.expired, "Çok fazla hatalı deneme yaptın. Yeni bir kod istemen gerekiyor.")
        case .verificationCodeExpired:
            return (.expired, "Kodun süresi doldu. Yeni bir kod isteyebilirsin.")
        case .invalidVerificationCode:
            return (.entering, "Kod geçersiz. Tekrar dener misin?")
        default:
            return (.entering, "Bir şeyler ters gitti, birazdan tekrar dene.")
        }
    }
}

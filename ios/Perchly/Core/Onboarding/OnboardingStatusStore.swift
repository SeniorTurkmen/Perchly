import Foundation

/// Whether the user has completed onboarding, cached in UserDefaults —
/// not sensitive data, just a UI-routing flag — so RootView can decide
/// what to show at launch without a network round trip. Kept in sync
/// from two directions (see AuthManager): every session response
/// carries the backend's own answer (self-healing after a reinstall,
/// new device, or a linked email that already onboarded elsewhere), and
/// OnboardingCoordinator sets it directly the instant its local flow
/// finishes, without waiting on that network call either.
struct OnboardingStatusStore {
    private let key = "perchly.has_completed_onboarding"

    var hasCompletedOnboarding: Bool {
        get { UserDefaults.standard.bool(forKey: key) }
        nonmutating set { UserDefaults.standard.set(newValue, forKey: key) }
    }
}

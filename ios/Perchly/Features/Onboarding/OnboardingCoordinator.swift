import Foundation

/// The onboarding flow's steps, in order. A future welcome/intro
/// sequence ("Perchlemek" — karşılama, kelime tanıtımı, platonik
/// konumlandırma) is meant to live under this same coordinator ahead of
/// `.ageRange`, once those screens exist; for now this covers only the
/// data-collection steps that are actually specified.
enum OnboardingStep: Equatable {
    case ageRange
    case moodPreference
    case notificationPermission
    case personaPick
    case completed
}

/// The single source of truth for onboarding: which step is current,
/// and the answers collected so far. Screens never navigate to each
/// other directly — each one only reports "I'm done" back to the
/// coordinator via one of the methods below, and the coordinator alone
/// decides what comes next (see OnboardingFlowView, which renders
/// whichever step this points at).
@MainActor
final class OnboardingCoordinator: ObservableObject {
    @Published private(set) var step: OnboardingStep = .ageRange
    @Published private(set) var profile = OnboardingProfile()
    /// Set the instant a persona is picked, so OnboardingFlowView can
    /// render ChatView for `.completed` without waiting on anything.
    @Published private(set) var chatPersona: Persona?

    private let authManager: AuthManager
    private let retryQueue: OnboardingRetryQueue

    init(authManager: AuthManager = .shared, retryQueue: OnboardingRetryQueue = OnboardingRetryQueue()) {
        self.authManager = authManager
        self.retryQueue = retryQueue
    }

    // MARK: Screen 1 — age range (mandatory)

    func selectAgeRange(_ ageRange: OnboardingProfile.AgeRange) {
        profile.ageRange = ageRange
        step = .moodPreference
    }

    // MARK: Screen 2 — mood preference (optional)

    func selectMood(_ mood: OnboardingProfile.MoodPreference) {
        profile.moodPreference = mood
        step = .notificationPermission
    }

    func skipMood() {
        profile.moodPreference = .skipped
        step = .notificationPermission
    }

    // MARK: Screen 3 — notification permission

    func setNotificationsGranted(_ granted: Bool) {
        profile.notificationsGranted = granted
        step = .personaPick
    }

    // MARK: Screen 4 — persona pick, and completion

    /// The user tapped a persona. Local completion is immediate and
    /// unconditional — `step` flips to `.completed` right away, and
    /// `AuthManager` is told onboarding is done before any network call
    /// even starts, so the UI never waits on the backend. Syncing the
    /// profile happens in the background afterward; if it fails,
    /// `OnboardingRetryQueue` holds onto it for the next opportunity.
    func selectPersona(_ persona: Persona) {
        profile.selectedPersonaID = persona.id
        chatPersona = persona
        authManager.markOnboardingCompleted()
        step = .completed

        let profileToSync = profile
        Task { [retryQueue] in
            await retryQueue.submitOrEnqueue(profileToSync)
        }
    }
}

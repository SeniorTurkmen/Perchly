import Foundation

/// The onboarding flow's steps, in order: name/hitap, age, mood,
/// notifications, then a persona match preview.
enum OnboardingStep: Equatable {
    case nameHitap
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
    @Published private(set) var step: OnboardingStep = .nameHitap
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

    // MARK: Screen 1 — preferred name, or no hitap at all

    func selectPreferredName(_ name: String) {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard PreferredNameRules.isValid(trimmed) else { return }
        profile.preferredName = trimmed
        profile.skipHitap = false
        LocalPreferredNameStore.save(name: trimmed)
        step = .ageRange
    }

    /// User opted out of being addressed by name. Nothing is stored and
    /// personas must not invent a nickname either.
    func continueAnonymously() {
        profile.preferredName = nil
        profile.skipHitap = true
        LocalPreferredNameStore.clear()
        step = .ageRange
    }

    func goBack() {
        switch step {
        case .ageRange:
            step = .nameHitap
        case .moodPreference:
            step = .ageRange
        case .notificationPermission:
            step = .moodPreference
        case .personaPick:
            step = .notificationPermission
        case .nameHitap, .completed:
            break
        }
    }

    // MARK: Screen 2 — age range (mandatory)

    func selectAgeRange(_ ageRange: OnboardingProfile.AgeRange) {
        profile.ageRange = ageRange
        step = .moodPreference
    }

    // MARK: Screen 3 — mood preference (optional)

    func selectMood(_ mood: OnboardingProfile.MoodPreference) {
        profile.moodPreference = mood
        step = .notificationPermission
    }

    func skipMood() {
        profile.moodPreference = .skipped
        step = .notificationPermission
    }

    // MARK: Screen 4 — notification permission

    func setNotificationsGranted(_ granted: Bool) {
        profile.notificationsGranted = granted
        step = .personaPick
    }

    // MARK: Screen 5 — persona match preview, and completion

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

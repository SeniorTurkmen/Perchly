import Foundation

/// Syncs a completed OnboardingProfile to the backend
/// (`POST /users/onboarding-profile`), and if that fails, persists it to
/// disk so it isn't lost — retried on the next opportunity (see
/// `retryPending()`, called at app launch from AnonymousBootstrapView).
///
/// The onboarding flow itself never waits on any of this: it treats
/// onboarding as locally "done" the instant the user finishes,
/// regardless of whether the network call succeeds (see
/// OnboardingCoordinator.selectPersona).
struct OnboardingRetryQueue {
    private let apiClient: APIClient
    private let defaultsKey = "perchly.onboarding_retry_queue"

    init(apiClient: APIClient = .shared) {
        self.apiClient = apiClient
    }

    /// Tries to submit `profile` right away; if that fails, queues it
    /// instead of losing it. Never throws — callers proceed with their
    /// own flow regardless of the outcome.
    func submitOrEnqueue(_ profile: OnboardingProfile) async {
        do {
            try await submit(profile)
        } catch {
            enqueue(profile)
        }
    }

    /// Retries every profile still pending from a previous failed
    /// attempt. Call at app launch, after a session is established.
    func retryPending() async {
        let pending = loadPending()
        guard !pending.isEmpty else { return }

        var stillPending: [OnboardingProfile] = []
        for profile in pending {
            do {
                try await submit(profile)
            } catch {
                stillPending.append(profile)
            }
        }
        savePending(stillPending)
    }

    private struct SaveProfileRequest: Encodable {
        let age_range: String
        let is_minor: Bool
        let mood_preference: String?
        let notifications_granted: Bool
        let selected_persona_id: String?
    }

    private func submit(_ profile: OnboardingProfile) async throws {
        // Screen 1 (age range) is mandatory, so a profile reaching this
        // point should always have one — but there's nothing meaningful
        // to submit without it, so just skip rather than send garbage.
        guard let ageRange = profile.ageRange else { return }

        let body = try JSONEncoder().encode(SaveProfileRequest(
            age_range: ageRange.rawValue,
            is_minor: profile.isMinor,
            mood_preference: profile.moodPreference?.rawValue,
            notifications_granted: profile.notificationsGranted,
            selected_persona_id: profile.selectedPersonaID
        ))
        try await apiClient.send(APIRequest(path: "/users/onboarding-profile", method: .post, body: body))
    }

    private func enqueue(_ profile: OnboardingProfile) {
        var pending = loadPending()
        pending.append(profile)
        savePending(pending)
    }

    private func loadPending() -> [OnboardingProfile] {
        guard let data = UserDefaults.standard.data(forKey: defaultsKey) else { return [] }
        return (try? JSONDecoder().decode([OnboardingProfile].self, from: data)) ?? []
    }

    private func savePending(_ profiles: [OnboardingProfile]) {
        guard let data = try? JSONEncoder().encode(profiles) else { return }
        UserDefaults.standard.set(data, forKey: defaultsKey)
    }
}

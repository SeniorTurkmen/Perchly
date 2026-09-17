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

    /// Tries to submit `profile` right away; if that fails for a reason
    /// a retry could plausibly fix (network blip, backend hiccup),
    /// queues it instead of losing it. Never throws — callers proceed
    /// with their own flow regardless of the outcome.
    func submitOrEnqueue(_ profile: OnboardingProfile) async {
        do {
            try await submit(profile)
        } catch {
            if Self.isRetryable(error) {
                enqueue(profile)
            }
            // A permanent validation failure (see isRetryable) would
            // fail identically on every future launch — queuing it
            // would just grow the retry list forever with an entry
            // that can never succeed.
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
                if Self.isRetryable(error) {
                    stillPending.append(profile)
                }
            }
        }
        savePending(stillPending)
    }

    /// A 4xx body-validation failure (e.g. this build's local checks
    /// missed something the backend now rejects) will fail the exact
    /// same way on every retry — pointless to keep queuing. Anything
    /// else (network failure, decoding, a 5xx) is worth retrying later.
    /// Defaults to retryable for a non-APIError or a body without a
    /// `code`, since dropping something we can't classify risks losing
    /// data for no reason.
    ///
    /// Internal (not private) so OnboardingRetryQueueTests can exercise
    /// it directly.
    static func isRetryable(_ error: Error) -> Bool {
        guard let apiError = error as? APIError, let code = apiError.code else { return true }
        switch code {
        case .invalidAgeRange, .invalidMoodPreference, .invalidSelectedPersonaID,
             .preferredNameRequired, .preferredNameInvalid:
            return false
        default:
            return true
        }
    }

    private struct SaveProfileRequest: Encodable {
        let age_range: String
        let is_minor: Bool
        let mood_preference: String?
        let notifications_granted: Bool
        let selected_persona_id: String?
        /// Omitted (not null) when this payload predates hitap, so the
        /// server preserves whatever is already on file — see
        /// OnboardingService.SaveProfile's "neither field sent" path.
        let preferred_name: String?
        let skip_hitap: Bool?

        enum CodingKeys: String, CodingKey {
            case age_range, is_minor, mood_preference
            case notifications_granted, selected_persona_id
            case preferred_name, skip_hitap
        }

        func encode(to encoder: Encoder) throws {
            var container = encoder.container(keyedBy: CodingKeys.self)
            try container.encode(age_range, forKey: .age_range)
            try container.encode(is_minor, forKey: .is_minor)
            try container.encodeIfPresent(mood_preference, forKey: .mood_preference)
            try container.encode(notifications_granted, forKey: .notifications_granted)
            try container.encodeIfPresent(selected_persona_id, forKey: .selected_persona_id)
            try container.encodeIfPresent(preferred_name, forKey: .preferred_name)
            try container.encodeIfPresent(skip_hitap, forKey: .skip_hitap)
        }
    }

    private func submit(_ profile: OnboardingProfile) async throws {
        // Screen 2 (age range) is mandatory, so a profile reaching this
        // point should always have one — but there's nothing meaningful
        // to submit without it, so just skip rather than send garbage.
        guard let ageRange = profile.ageRange else { return }

        let preferredName: String?
        let skipHitap: Bool?
        if profile.skipHitap {
            preferredName = nil
            skipHitap = true
        } else if let name = profile.preferredName?.trimmingCharacters(in: .whitespacesAndNewlines),
                  PreferredNameRules.isValid(name) {
            preferredName = name
            skipHitap = nil
        } else {
            preferredName = nil
            skipHitap = nil
        }

        let body = try JSONEncoder().encode(SaveProfileRequest(
            age_range: ageRange.rawValue,
            is_minor: profile.isMinor,
            mood_preference: profile.moodPreference?.rawValue,
            notifications_granted: profile.notificationsGranted,
            selected_persona_id: profile.selectedPersonaID,
            preferred_name: preferredName,
            skip_hitap: skipHitap
        ))
        try await apiClient.send(APIRequest(path: "/users/onboarding-profile", method: .post, body: body))
    }

    /// Pulls `preferred_name` / `skip_hitap` from the server so a
    /// reinstall (same anonymous device id) doesn't lose hitap. 404
    /// means onboarding hasn't been saved yet — leave the local store.
    func syncPreferredName() async {
        struct Remote: Decodable {
            let preferredName: String?
            let skipHitap: Bool

            enum CodingKeys: String, CodingKey {
                case preferredName = "preferred_name"
                case skipHitap = "skip_hitap"
            }
        }

        do {
            let remote: Remote = try await apiClient.send(APIRequest(path: "/users/onboarding-profile"))
            if remote.skipHitap {
                LocalPreferredNameStore.clear()
            } else if let name = remote.preferredName, PreferredNameRules.isValid(name) {
                LocalPreferredNameStore.save(name: name)
            }
        } catch {
            return
        }
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

import SwiftUI

/// Branches on `AuthManager.state` and, once a session exists, on
/// whether onboarding is still needed. Both `.anonymous` and
/// `.authenticated` show the same main app — an anonymous session is
/// fully usable on its own; the only visible difference between the two
/// states is what ProfileView shows (an invitation to link an email, vs.
/// the linked email itself). Onboarding is likewise collected regardless
/// of which of those two states the user is in.
struct RootView: View {
    @ObservedObject private var authManager = AuthManager.shared

    /// Owned here, not by OnboardingFlowView, specifically so this view
    /// can read `chatPersona` the instant onboarding finishes and hand
    /// it straight to PersonaListView (see below) — that's what gives
    /// the user a real, working back button into the persona list
    /// instead of being stuck looking at an isolated ChatView with no
    /// way out until the app is relaunched.
    @StateObject private var onboardingCoordinator = OnboardingCoordinator()

    var body: some View {
        switch authManager.state {
        case .loading:
            AnonymousBootstrapView()
        case .anonymous, .authenticated:
            content
        }
    }

    @ViewBuilder
    private var content: some View {
        if needsOnboarding {
            OnboardingFlowView(coordinator: onboardingCoordinator)
        } else {
            PersonaListView(initialPersona: onboardingCoordinator.chatPersona)
        }
    }

    /// True until onboarding both (a) hasn't been marked complete yet
    /// AND (b) hasn't just finished in THIS app run either. Checking
    /// both — not just `hasCompletedOnboarding` — matters because
    /// `OnboardingCoordinator.selectPersona` sets `chatPersona` and
    /// flips `hasCompletedOnboarding` together, synchronously, before
    /// either published change is even dispatched: by the time this
    /// re-evaluates, `chatPersona` is already there to hand off, in the
    /// very same render pass, rather than switching over one beat too
    /// early with nothing to show yet.
    private var needsOnboarding: Bool {
        !authManager.hasCompletedOnboarding && onboardingCoordinator.step != .completed
    }
}

#Preview {
    RootView()
}

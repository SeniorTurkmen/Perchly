import SwiftUI

/// Renders whichever onboarding step `coordinator` currently points to.
/// Steps never navigate to each other via NavigationLink — this view is
/// the only thing that switches what's on screen, driven entirely by
/// `coordinator.step`.
///
/// Takes `coordinator` from the outside (owned by RootView) rather than
/// creating its own: once `.completed` is reached, RootView is what
/// actually shows the persona list with the picked persona already
/// pushed (see PersonaListView's `initialPersona`) — this view's own
/// `.completed` case is just a brief, invisible bridge frame while that
/// hand-off happens, never really shown to the user.
///
/// Transitions are a soft crossfade+slide, on a slow, calm timing curve
/// — Liquid Glass reads best with gentle motion, never a hard cut.
struct OnboardingFlowView: View {
    @ObservedObject var coordinator: OnboardingCoordinator

    var body: some View {
        ZStack {
            switch coordinator.step {
            case .ageRange:
                AgeRangeStepView(coordinator: coordinator)
                    .transition(stepTransition)
            case .moodPreference:
                MoodPreferenceStepView(coordinator: coordinator)
                    .transition(stepTransition)
            case .notificationPermission:
                NotificationPermissionStepView(coordinator: coordinator)
                    .transition(stepTransition)
            case .personaPick:
                PersonaPickStepView(coordinator: coordinator)
                    .transition(stepTransition)
            case .completed:
                Color.clear
            }
        }
        .animation(.easeInOut(duration: 0.45), value: coordinator.step)
    }

    private var stepTransition: AnyTransition {
        .asymmetric(
            insertion: .opacity.combined(with: .move(edge: .trailing)),
            removal: .opacity.combined(with: .move(edge: .leading))
        )
    }
}

#Preview {
    OnboardingFlowView(coordinator: OnboardingCoordinator())
}

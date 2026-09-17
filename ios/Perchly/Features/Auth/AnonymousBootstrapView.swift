import SwiftUI

/// Shown only for `AuthManager.state == .loading`, i.e. very briefly at
/// launch while an existing session is checked or a first anonymous one
/// is created. Deliberately blank — no spinner, no logo animation — so
/// it reads as a continuation of the launch screen rather than a
/// separate loading moment.
struct AnonymousBootstrapView: View {
    var body: some View {
        PerchlyPalette.background
            .ignoresSafeArea()
            .task {
                await AuthManager.shared.bootstrap()
                // Best-effort, after a session exists: flush any
                // onboarding profile that failed to sync last time.
                await OnboardingRetryQueue().retryPending()
            }
    }
}

#Preview {
    AnonymousBootstrapView()
}

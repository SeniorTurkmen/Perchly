import SwiftUI
import UserNotifications

/// Screen 3: a calm explainer shown before the native permission
/// dialog — tapping "İzin Ver" triggers the real
/// `UNUserNotificationCenter` prompt. Either outcome (granted or
/// denied) just records the result and moves on; the flow never stalls
/// on a decline.
struct NotificationPermissionStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @State private var isRequesting = false

    var body: some View {
        VStack(spacing: 28) {
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                Text("Seni haberdar edelim mi?")
                    .font(PerchlyTypography.largeTitle)
                Text("Biri sana yanıt verdiğinde haber verelim.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            GlassButton(title: "İzin Ver", isLoading: isRequesting) {
                Task { await requestPermission() }
            }
            .accessibilityIdentifier("requestNotificationButton")

            Spacer()
        }
        .padding()
    }

    private func requestPermission() async {
        isRequesting = true
        defer { isRequesting = false }

        let granted = await requestNotificationAuthorization()
        coordinator.setNotificationsGranted(granted)
    }

    private func requestNotificationAuthorization() async -> Bool {
        await withCheckedContinuation { continuation in
            UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge]) { granted, _ in
                continuation.resume(returning: granted)
            }
        }
    }
}

#Preview {
    NotificationPermissionStepView(coordinator: OnboardingCoordinator())
}

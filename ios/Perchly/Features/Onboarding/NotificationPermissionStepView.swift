import SwiftUI
import UserNotifications

/// Stitch "Onboarding - Bildirim İzni". Still drives the real
/// `UNUserNotificationCenter` prompt; decline just records false and
/// continues, same as before.
struct NotificationPermissionStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @State private var isRequesting = false

    var body: some View {
        OnboardingScaffold(stepIndex: 4, stepLabel: "Bildirim", onBack: coordinator.goBack) {
            VStack(spacing: 20) {
                ZStack {
                    Circle()
                        .fill(PerchlyPalette.Discover.primaryFixed.opacity(0.7))
                        .frame(width: 120, height: 120)
                        .blur(radius: 12)
                    ZStack {
                        Circle().fill(PerchlyPalette.Discover.primary)
                        Image(systemName: "bell.fill")
                            .font(.system(size: 32))
                            .foregroundStyle(PerchlyPalette.Discover.onPrimary)
                    }
                    .frame(width: 88, height: 88)
                    .shadow(color: PerchlyPalette.Discover.primary.opacity(0.25), radius: 12, y: 6)
                }
                .padding(.top, 12)

                VStack(spacing: 8) {
                    Text("Seni haberdar edelim mi?")
                        .font(PerchlyTypography.Discover.headlineLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                        .multilineTextAlignment(.center)
                    Text("Biri sana yanıt verdiğinde haber verelim. İzni daha sonra ayarlardan değiştirebilirsin.")
                        .font(PerchlyTypography.Discover.bodyMD)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                        .multilineTextAlignment(.center)
                }
            }
            .frame(maxWidth: .infinity)
        } footer: {
            VStack(spacing: 10) {
                OnboardingPrimaryButton(title: "İzin Ver", isLoading: isRequesting) {
                    Task { await requestPermission() }
                }
                .accessibilityIdentifier("requestNotificationButton")

                Button {
                    coordinator.setNotificationsGranted(false)
                } label: {
                    Text("Şimdi değil")
                        .font(PerchlyTypography.Discover.labelLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("skipNotificationButton")
            }
        }
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

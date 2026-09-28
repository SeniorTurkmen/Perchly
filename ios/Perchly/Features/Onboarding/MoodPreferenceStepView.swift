import SwiftUI

/// Stitch "Onboarding - Ruh Hali Seçimi". Mood values still match the
/// backend enum; presentation is local until a dedicated mood API exists.
struct MoodPreferenceStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator

    private let options: [(mood: OnboardingProfile.MoodPreference, icon: String, title: String, subtitle: String)] = [
        (.motivation, "flame.fill", "Motive Olmak", "Küçük adımlar, somut cesaret"),
        (.dailyChat, "bubble.left.and.bubble.right.fill", "Gündelik Sohbet", "Günün nasıl geçtiğini paylaş"),
        (.hobbyTalk, "book.fill", "Hobi Paylaşmak", "Kitap, film ve tutkular"),
    ]

    var body: some View {
        OnboardingScaffold(stepIndex: 3, stepLabel: "Ruh Hali", onBack: coordinator.goBack) {
            VStack(alignment: .leading, spacing: 20) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Şu an nasıl bir sohbet arıyorsun?")
                        .font(PerchlyTypography.Discover.headlineLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Text("İstersen bunu atlayabilirsin; istediğin zaman değiştirebilirsin.")
                        .font(PerchlyTypography.Discover.bodyMD)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }

                VStack(spacing: 12) {
                    ForEach(options, id: \.mood) { option in
                        Button {
                            coordinator.selectMood(option.mood)
                        } label: {
                            HStack(spacing: 14) {
                                ZStack {
                                    Circle().fill(PerchlyPalette.Discover.primaryFixed)
                                    Image(systemName: option.icon)
                                        .font(.system(size: 18))
                                        .foregroundStyle(PerchlyPalette.Discover.primary)
                                }
                                .frame(width: 48, height: 48)

                                VStack(alignment: .leading, spacing: 4) {
                                    Text(option.title)
                                        .font(PerchlyTypography.Discover.headlineSM)
                                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                                    Text(option.subtitle)
                                        .font(PerchlyTypography.Discover.bodySM)
                                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                                }
                                Spacer(minLength: 0)
                                Image(systemName: "chevron.forward")
                                    .font(.system(size: 12, weight: .semibold))
                                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                            }
                            .padding(16)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.7)), in: .rect(cornerRadius: 18))
                        }
                        .buttonStyle(.plain)
                        .accessibilityIdentifier("moodOption_\(option.mood.rawValue)")
                    }
                }
            }
        } footer: {
            Button {
                coordinator.skipMood()
            } label: {
                Text("Şimdilik Atla")
                    .font(PerchlyTypography.Discover.labelLG)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    .frame(maxWidth: .infinity)
                    .frame(height: 44)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("skipMoodButton")
        }
    }
}

#Preview {
    MoodPreferenceStepView(coordinator: OnboardingCoordinator())
}

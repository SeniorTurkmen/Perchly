import SwiftUI

/// Screen 2: optional, skippable. Three GlassCard options (icon +
/// title); tapping one, or skipping, both advance the same way.
struct MoodPreferenceStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator

    private let options: [(mood: OnboardingProfile.MoodPreference, icon: String, title: String)] = [
        (.motivation, "flame.fill", "Motive Olmak"),
        (.dailyChat, "bubble.left.and.bubble.right.fill", "Gündelik Sohbet"),
        (.hobbyTalk, "sparkles", "Hobi Paylaşmak"),
    ]

    var body: some View {
        VStack(spacing: 28) {
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                Text("Şu an nasıl bir sohbet arıyorsun?")
                    .font(PerchlyTypography.largeTitle)
                Text("İstersen bunu atlayabilirsin, istediğin zaman değiştirebilirsin.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            VStack(spacing: 12) {
                ForEach(options, id: \.mood) { option in
                    Button {
                        coordinator.selectMood(option.mood)
                    } label: {
                        GlassCard {
                            HStack(spacing: 14) {
                                Image(systemName: option.icon)
                                    .font(.title2)
                                    .foregroundStyle(PerchlyPalette.accent)
                                    .frame(width: 28)
                                Text(option.title)
                                    .font(PerchlyTypography.body.weight(.semibold))
                                    .foregroundStyle(PerchlyPalette.textPrimary)
                                Spacer(minLength: 0)
                            }
                        }
                    }
                    .buttonStyle(.plain)
                    .accessibilityIdentifier("moodOption_\(option.mood.rawValue)")
                }
            }

            Button("Şimdilik Atla") {
                coordinator.skipMood()
            }
            .font(PerchlyTypography.body)
            .foregroundStyle(PerchlyPalette.textSecondary)
            .accessibilityIdentifier("skipMoodButton")

            Spacer()
        }
        .padding()
    }
}

#Preview {
    MoodPreferenceStepView(coordinator: OnboardingCoordinator())
}

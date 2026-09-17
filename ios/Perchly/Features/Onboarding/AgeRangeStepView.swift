import SwiftUI

/// Screen 1: mandatory, no skip. Single-select age range via GlassButton.
struct AgeRangeStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @State private var selected: OnboardingProfile.AgeRange?

    private let options: [(range: OnboardingProfile.AgeRange, label: String)] = [
        (.under18, "13-17"),
        (.age18to24, "18-24"),
        (.age25to34, "25-34"),
        (.age35plus, "35+"),
    ]

    var body: some View {
        VStack(spacing: 28) {
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                Text("Yaşını hangi aralıkta belirtebiliriz?")
                    .font(PerchlyTypography.largeTitle)
                Text("Bu bilgiyi sana daha uygun bir deneyim sunmak için kullanıyoruz.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            VStack(spacing: 12) {
                ForEach(options, id: \.range) { option in
                    GlassButton(title: option.label, isSelected: selected == option.range) {
                        selected = option.range
                    }
                    .frame(maxWidth: .infinity)
                    .accessibilityIdentifier("ageOption_\(option.range.rawValue)")
                }
            }

            GlassButton(title: "Devam Et", isDisabled: selected == nil) {
                if let selected {
                    coordinator.selectAgeRange(selected)
                }
            }
            .accessibilityIdentifier("ageContinueButton")

            Spacer()
        }
        .padding()
    }
}

#Preview {
    AgeRangeStepView(coordinator: OnboardingCoordinator())
}

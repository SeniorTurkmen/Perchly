import SwiftUI

/// Stitch "Onboarding - Yaş Seçimi". Options still map 1:1 onto the
/// backend `age_range` enum; copy/icons are presentation-only.
struct AgeRangeStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @State private var selected: OnboardingProfile.AgeRange?

    private let options: [(range: OnboardingProfile.AgeRange, label: String, subtitle: String, badge: String?, symbol: String)] = [
        (.under18, "13 – 17", "Genç & Samimi Bakış", "İlk Adım", "graduationcap.fill"),
        (.age18to24, "18 – 24", "Dinamik & Paylaşımcı", "Popüler", "bolt.fill"),
        (.age25to34, "25 – 34", "Denge & Günlük Dinlenme", nil, "cup.and.saucer.fill"),
        (.age35plus, "35+", "Dingin & Olgun Sohbetler", nil, "leaf.fill"),
    ]

    var body: some View {
        OnboardingScaffold(stepIndex: 2, stepLabel: "Yaş Seçimi", onBack: coordinator.goBack) {
            VStack(spacing: 20) {
                VStack(spacing: 8) {
                    Text("Sana en uygun Perchmate'i bulalım.")
                        .font(PerchlyTypography.Discover.headlineLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                        .multilineTextAlignment(.center)
                    Text("Deneyimi yaşına göre kişiselleştirebilmemiz için yaş aralığını seçebilirsin.")
                        .font(PerchlyTypography.Discover.bodySM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                        .multilineTextAlignment(.center)
                }
                .padding(.horizontal, 8)

                VStack(spacing: 10) {
                    ForEach(options, id: \.range) { option in
                        ageCard(option)
                    }
                }

                HStack(spacing: 10) {
                    ZStack {
                        Circle().fill(PerchlyPalette.Discover.tertiaryFixed)
                        Image(systemName: "checkmark.shield.fill")
                            .font(.system(size: 13))
                            .foregroundStyle(PerchlyPalette.Discover.tertiary)
                    }
                    .frame(width: 36, height: 36)
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Güvenli ve Bireysel Alan")
                            .font(PerchlyTypography.Discover.labelMD.weight(.semibold))
                            .foregroundStyle(PerchlyPalette.Discover.onSurface)
                        Text("Sohbet tonu bu seçime göre şekillenecektir.")
                            .font(PerchlyTypography.Discover.bodySM)
                            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    }
                    Spacer(minLength: 0)
                }
                .padding(12)
                .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.6)), in: .rect(cornerRadius: 16))
            }
        } footer: {
            VStack(spacing: 10) {
                OnboardingPrimaryButton(title: "Devam Et", isDisabled: selected == nil) {
                    if let selected {
                        coordinator.selectAgeRange(selected)
                    }
                }
                .accessibilityIdentifier("ageContinueButton")

                HStack(alignment: .top, spacing: 6) {
                    Image(systemName: "lock.fill")
                        .font(.system(size: 11))
                        .padding(.top, 2)
                    Text("Yaş bilgin gizli tutulur ve yalnızca eşleşme uyumunu artırmak için kullanılır.")
                        .font(PerchlyTypography.Discover.labelSM)
                }
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
            }
        }
    }

    private func ageCard(_ option: (range: OnboardingProfile.AgeRange, label: String, subtitle: String, badge: String?, symbol: String)) -> some View {
        let isSelected = selected == option.range
        return Button {
            selected = option.range
        } label: {
            HStack(spacing: 14) {
                ZStack {
                    Circle().fill(isSelected ? PerchlyPalette.Discover.primary : PerchlyPalette.Discover.surfaceContainer)
                    Image(systemName: option.symbol)
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundStyle(isSelected ? PerchlyPalette.Discover.onPrimary : PerchlyPalette.Discover.primary)
                }
                .frame(width: 48, height: 48)

                VStack(alignment: .leading, spacing: 4) {
                    HStack(spacing: 8) {
                        Text(option.label)
                            .font(PerchlyTypography.Discover.labelLG)
                            .foregroundStyle(PerchlyPalette.Discover.onSurface)
                        if let badge = option.badge {
                            Text(badge)
                                .font(PerchlyTypography.Discover.labelSM)
                                .foregroundStyle(isSelected ? PerchlyPalette.Discover.onPrimaryContainer : PerchlyPalette.Discover.secondary)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 3)
                                .background(
                                    (isSelected ? PerchlyPalette.Discover.primaryContainer : PerchlyPalette.Discover.secondaryFixed).opacity(0.55),
                                    in: Capsule()
                                )
                        }
                    }
                    Text(option.subtitle)
                        .font(PerchlyTypography.Discover.bodySM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }

                Spacer(minLength: 8)

                ZStack {
                    Circle()
                        .fill(isSelected ? PerchlyPalette.Discover.primary : PerchlyPalette.Discover.surfaceContainerHigh)
                    if isSelected {
                        Image(systemName: "checkmark")
                            .font(.system(size: 11, weight: .bold))
                            .foregroundStyle(PerchlyPalette.Discover.onPrimary)
                    }
                }
                .frame(width: 28, height: 28)
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                isSelected
                    ? PerchlyPalette.Discover.primaryFixed.opacity(0.45)
                    : PerchlyPalette.Discover.surfaceLowest.opacity(0.72),
                in: RoundedRectangle(cornerRadius: 16, style: .continuous)
            )
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("ageOption_\(option.range.rawValue)")
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}

#Preview {
    AgeRangeStepView(coordinator: OnboardingCoordinator())
}

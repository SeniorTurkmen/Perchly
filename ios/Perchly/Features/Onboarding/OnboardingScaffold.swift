import SwiftUI

/// Shared Stitch chrome for the five onboarding steps: ambient blobs,
/// PERCHLY + progress dots, a step pill, and a pinned primary footer.
struct OnboardingScaffold<Content: View, Footer: View>: View {
    let stepIndex: Int
    let stepLabel: String
    var onBack: (() -> Void)?
    @ViewBuilder var content: Content
    @ViewBuilder var footer: Footer

    private let totalSteps = 5

    var body: some View {
        VStack(spacing: 0) {
            header
            ScrollView {
                content
                    .padding(.horizontal, 20)
                    .padding(.top, 16)
                    .padding(.bottom, 24)
            }
            .scrollIndicators(.hidden)
            footer
                .padding(.horizontal, 20)
                .padding(.top, 8)
                .padding(.bottom, 16)
                .background(
                    LinearGradient(
                        colors: [
                            PerchlyPalette.Discover.background.opacity(0),
                            PerchlyPalette.Discover.background,
                        ],
                        startPoint: .top,
                        endPoint: .bottom
                    )
                )
        }
        .background {
            DiscoverAmbientBackground()
        }
        .background(PerchlyPalette.Discover.background)
        .preferredColorScheme(.light)
    }

    private var header: some View {
        VStack(spacing: 8) {
            HStack {
                if let onBack {
                    Button(action: onBack) {
                        Image(systemName: "chevron.backward")
                            .font(.system(size: 15, weight: .semibold))
                            .foregroundStyle(PerchlyPalette.Discover.onSurface)
                            .frame(width: 40, height: 40)
                            .background(PerchlyPalette.Discover.surfaceLowest.opacity(0.8), in: Circle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("Geri")
                } else {
                    Color.clear.frame(width: 40, height: 40)
                }

                Spacer(minLength: 8)

                VStack(spacing: 6) {
                    Text("PERCHLY")
                        .font(PerchlyTypography.Discover.labelSM.weight(.semibold))
                        .tracking(1.2)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    HStack(spacing: 4) {
                        ForEach(1...totalSteps, id: \.self) { index in
                            Capsule()
                                .fill(index == stepIndex ? PerchlyPalette.Discover.primary : PerchlyPalette.Discover.surfaceContainerHigh)
                                .frame(width: index == stepIndex ? 22 : 8, height: 4)
                        }
                    }
                    .accessibilityElement(children: .ignore)
                    .accessibilityLabel("Adım \(String(stepIndex)) / \(String(totalSteps))")
                }

                Spacer(minLength: 8)
                Color.clear.frame(width: 40, height: 40)
            }
            .padding(.horizontal, 20)
            .padding(.top, 8)

            HStack(spacing: 6) {
                Circle()
                    .fill(PerchlyPalette.Discover.primary)
                    .frame(width: 6, height: 6)
                Text("Adım \(String(stepIndex)) / \(String(totalSteps))")
                Text("•")
                    .opacity(0.4)
                Text(stepLabel)
                    .foregroundStyle(PerchlyPalette.Discover.primary)
                    .fontWeight(.semibold)
            }
            .font(PerchlyTypography.Discover.labelSM)
            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
            .padding(.horizontal, 12)
            .padding(.vertical, 5)
            .background(PerchlyPalette.Discover.surfaceLow.opacity(0.9), in: Capsule())
            .padding(.bottom, 8)
        }
        .background(PerchlyPalette.Discover.surface.opacity(0.72))
        .background(.ultraThinMaterial)
    }
}

struct OnboardingPrimaryButton: View {
    // LocalizedStringKey, not String: Text(String) is verbatim and
    // skips the string catalog.
    let title: LocalizedStringKey
    var isDisabled: Bool = false
    var isLoading: Bool = false
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                if isLoading {
                    ProgressView()
                        .tint(PerchlyPalette.Discover.onPrimary)
                } else {
                    Text(title)
                    Image(systemName: "arrow.forward")
                        .font(.system(size: 13, weight: .semibold))
                }
            }
            .font(PerchlyTypography.Discover.labelLG)
            .foregroundStyle(PerchlyPalette.Discover.onPrimary)
            .frame(maxWidth: .infinity)
            .frame(height: 52)
            .background(PerchlyPalette.Discover.primary, in: Capsule())
        }
        .buttonStyle(.plain)
        .disabled(isDisabled || isLoading)
        .opacity(isDisabled ? 0.45 : 1)
        .shadow(color: PerchlyPalette.Discover.primary.opacity(0.28), radius: 16, y: 8)
    }
}

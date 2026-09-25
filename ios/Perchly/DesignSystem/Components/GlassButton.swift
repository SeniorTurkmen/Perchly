import SwiftUI

/// A reusable pill-shaped action button rendered with the Liquid Glass
/// material. `isLoading` swaps the title for a spinner (e.g. while an
/// async action is in flight) and implies disabled. `isSelected` fills
/// the glass with `tint` and switches to white text — for single-select
/// option groups (e.g. onboarding's age-range picker) where a button
/// doubles as a toggle rather than a one-shot action.
struct GlassButton: View {
    // LocalizedStringKey, not String: Text(String) is verbatim and
    // skips the string catalog — only Text(LocalizedStringKey) resolves
    // through Localizable.xcstrings.
    let title: LocalizedStringKey
    var isDisabled: Bool = false
    var isLoading: Bool = false
    var isSelected: Bool = false
    var tint: Color = PerchlyPalette.accent
    let action: () -> Void

    private var isInteractionDisabled: Bool { isDisabled || isLoading }

    var body: some View {
        Button(action: action) {
            ZStack {
                Text(title).opacity(isLoading ? 0 : 1)
                if isLoading {
                    ProgressView()
                        .progressViewStyle(.circular)
                        .tint(PerchlyPalette.textPrimary)
                }
            }
            .font(PerchlyTypography.body.weight(.semibold))
            .foregroundStyle(isSelected ? .white : PerchlyPalette.textPrimary)
            .padding(.horizontal, 24)
            .padding(.vertical, 12)
        }
        .buttonStyle(.plain)
        .disabled(isInteractionDisabled)
        .opacity(isDisabled ? 0.5 : 1)
        .glassEffect(glass, in: .capsule)
    }

    private var glass: Glass {
        isSelected ? .regular.tint(tint).interactive() : .regular.interactive()
    }
}

#Preview {
    VStack(spacing: 16) {
        GlassButton(title: "Devam Et") {}
        GlassButton(title: "18-24", isSelected: true) {}
    }
    .padding()
}

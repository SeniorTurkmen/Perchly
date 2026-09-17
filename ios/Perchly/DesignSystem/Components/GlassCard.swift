import SwiftUI

/// A reusable content surface rendered with the Liquid Glass material.
///
/// Wrap any content in a `GlassCard` to get consistent padding and glass
/// styling across features instead of applying `.glassEffect()` ad hoc.
/// Pass `tint` to let the glass pick up a brand or persona color.
struct GlassCard<Content: View>: View {
    var tint: Color? = nil
    @ViewBuilder let content: () -> Content

    var body: some View {
        content()
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .glassEffect(glass, in: .rect(cornerRadius: 24))
    }

    private var glass: Glass {
        guard let tint else { return .regular }
        return .regular.tint(tint.opacity(0.35))
    }
}

#Preview("Untinted") {
    GlassCard {
        VStack(alignment: .leading, spacing: 8) {
            Text("Glass Card")
                .font(PerchlyTypography.title)
            Text("Liquid Glass materyaliyle örnek bir kart bileşeni.")
                .font(PerchlyTypography.body)
                .foregroundStyle(PerchlyPalette.textSecondary)
        }
    }
    .padding()
}

#Preview("Tinted") {
    GlassCard(tint: Color(hex: "#FF6B35")) {
        VStack(alignment: .leading, spacing: 8) {
            Text("Glass Card")
                .font(PerchlyTypography.title)
            Text("accent_color camın altından yansır.")
                .font(PerchlyTypography.body)
                .foregroundStyle(PerchlyPalette.textSecondary)
        }
    }
    .padding()
}

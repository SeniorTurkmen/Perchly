import SwiftUI

/// Floating listening-session composer: a glass capsule with the Stitch
/// placeholder and a circular send disc that turns primary when the
/// draft can be sent.
struct ChatInputBar: View {
    @Binding var text: String
    let isSending: Bool
    let onSend: () -> Void

    var body: some View {
        HStack(spacing: 6) {
            TextField("Dilediğin gibi yaz, burası güvenli bir alan...", text: $text)
                .font(PerchlyTypography.Discover.bodySM)
                .foregroundStyle(PerchlyPalette.Discover.onSurface)
                .submitLabel(.send)
                .onSubmit {
                    if canSend { onSend() }
                }
                .padding(.leading, 18)
                .accessibilityIdentifier("chatMessageField")

            Button(action: onSend) {
                Image(systemName: "arrow.up")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(canSend ? PerchlyPalette.Discover.onPrimary : PerchlyPalette.Discover.onSecondaryFixed)
                    .frame(width: 40, height: 40)
                    .background(
                        Circle().fill(canSend ? PerchlyPalette.Discover.primary : PerchlyPalette.Discover.secondaryFixed)
                    )
            }
            .disabled(!canSend)
            .padding(.trailing, 8)
            .padding(.vertical, 8)
        }
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.85)), in: .capsule)
        .shadow(color: .black.opacity(0.08), radius: 16, y: 6)
    }

    private var canSend: Bool {
        !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !isSending
    }
}

#Preview {
    VStack {
        Spacer()
        ChatInputBar(text: .constant("Merhaba"), isSending: false, onSend: {})
            .padding(.horizontal, 20)
            .padding(.bottom, 12)
    }
    .background(PerchlyPalette.Discover.background)
}

import SwiftUI

/// One chat row matching the Stitch listening-session bubbles: glass
/// persona messages with a small avatar and timestamp, solid
/// primary-container user messages with a read receipt.
///
/// Assistant bubbles support an iMessage-style long-press reaction: hold
/// the bubble to reveal a floating row of MessageReaction.allowedEmojis,
/// tap one to react (tap the same one again to clear). Only assistant
/// messages can be reacted to this way — the persona's own reactions to
/// the user's messages are set by the backend and shown here purely as
/// a read-only badge (see reactionEmoji on either row).
struct ChatBubble: View {
    let message: ChatMessage
    let persona: Persona
    /// Called with the newly chosen emoji (or nil to clear) when the
    /// user picks a reaction from the long-press picker. Left nil (the
    /// default) disables the picker entirely — used by previews and any
    /// read-only rendering of this bubble.
    var onReact: ((String?) -> Void)?

    @State private var isShowingReactionPicker = false

    var body: some View {
        Group {
            switch message.role {
            case .assistant:
                assistantRow
            case .user:
                userRow
            }
        }
    }

    private var assistantRow: some View {
        HStack(alignment: .bottom, spacing: 10) {
            ChatPersonaAvatar(persona: persona, size: 28)

            VStack(alignment: .leading, spacing: 4) {
                bubbleText(color: PerchlyPalette.Discover.onSurface)
                    .padding(16)
                    .frame(maxWidth: 320, alignment: .leading)
                    .glassEffect(
                        .regular.tint(persona.accent.opacity(0.14)),
                        in: UnevenRoundedRectangle(
                            topLeadingRadius: 22,
                            bottomLeadingRadius: 6,
                            bottomTrailingRadius: 22,
                            topTrailingRadius: 22,
                            style: .continuous
                        )
                    )
                    .contentShape(.rect)
                    .onLongPressGesture(minimumDuration: 0.4) {
                        beginReacting()
                    }
                    .onTapGesture {
                        dismissReactionPickerIfShowing()
                    }
                    .accessibilityIdentifier("assistantBubble_\(message.id.uuidString)")
                    // Applied AFTER the identifier above, deliberately:
                    // an .overlay added before .accessibilityIdentifier
                    // gets its own accessibility elements' ids silently
                    // overridden by the ancestor's (observed via a UI
                    // test dump — the picker's buttons all inherited
                    // "assistantBubble_…" instead of "reactionOption_N").
                    .overlay(alignment: .topTrailing) {
                        reactionBadge
                    }
                    .overlay(alignment: .topLeading) {
                        if isShowingReactionPicker {
                            reactionPicker
                                .offset(y: -46)
                                .transition(.scale(scale: 0.85, anchor: .bottomLeading).combined(with: .opacity))
                                .zIndex(1)
                        }
                    }

                Text("\(Self.timeString(from: message.createdAt)) • \(persona.name)")
                    .font(PerchlyTypography.Discover.labelSM)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant.opacity(0.7))
                    .padding(.leading, 8)
            }

            Spacer(minLength: 24)
        }
    }

    private var userRow: some View {
        HStack(alignment: .bottom, spacing: 0) {
            Spacer(minLength: 36)

            VStack(alignment: .trailing, spacing: 4) {
                bubbleText(color: PerchlyPalette.Discover.onPrimaryContainer)
                    .padding(16)
                    .frame(maxWidth: 300, alignment: .leading)
                    .background(
                        UnevenRoundedRectangle(
                            topLeadingRadius: 22,
                            bottomLeadingRadius: 22,
                            bottomTrailingRadius: 6,
                            topTrailingRadius: 22,
                            style: .continuous
                        )
                        .fill(PerchlyPalette.Discover.primaryContainer)
                    )
                    .overlay(alignment: .topLeading) {
                        reactionBadge
                    }

                HStack(spacing: 4) {
                    Text(Self.timeString(from: message.createdAt))
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 11))
                        .foregroundStyle(PerchlyPalette.Discover.primary)
                }
                .font(PerchlyTypography.Discover.labelSM)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant.opacity(0.7))
                .padding(.trailing, 8)
            }
        }
    }

    @ViewBuilder
    private var reactionBadge: some View {
        if let emoji = message.reactionEmoji {
            Text(emoji)
                .font(.system(size: 14))
                .padding(4)
                .background(Circle().fill(PerchlyPalette.Discover.background))
                .overlay(Circle().strokeBorder(.black.opacity(0.06)))
                .shadow(color: .black.opacity(0.12), radius: 3, y: 1)
                .offset(x: 6, y: -6)
                .accessibilityIdentifier("reactionBadge_\(message.id.uuidString)")
                .accessibilityLabel("Tepki: \(emoji)")
        }
    }

    private var reactionPicker: some View {
        HStack(spacing: 2) {
            ForEach(Array(MessageReaction.allowedEmojis.enumerated()), id: \.offset) { index, emoji in
                Button {
                    pickReaction(emoji)
                } label: {
                    Text(emoji)
                        .font(.system(size: 22))
                        .frame(width: 38, height: 38)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("reactionOption_\(index)")
            }
        }
        .padding(.horizontal, 4)
        .padding(.vertical, 3)
        .glassEffect(.regular, in: .capsule)
        .shadow(color: .black.opacity(0.18), radius: 10, y: 4)
    }

    private func beginReacting() {
        guard onReact != nil, !message.isStreaming else { return }
        withAnimation(.spring(response: 0.3, dampingFraction: 0.75)) {
            isShowingReactionPicker = true
        }
    }

    private func dismissReactionPickerIfShowing() {
        guard isShowingReactionPicker else { return }
        withAnimation(.spring(response: 0.25, dampingFraction: 0.8)) {
            isShowingReactionPicker = false
        }
    }

    private func pickReaction(_ emoji: String) {
        withAnimation(.spring(response: 0.25, dampingFraction: 0.8)) {
            isShowingReactionPicker = false
        }
        let newValue = message.reactionEmoji == emoji ? nil : emoji
        onReact?(newValue)
    }

    @ViewBuilder
    private func bubbleText(color: Color) -> some View {
        if message.content.isEmpty && message.isStreaming {
            TypingIndicator(tint: color)
                .padding(.vertical, 4)
        } else {
            Text(message.content)
                .font(PerchlyTypography.Discover.bodyMD)
                .foregroundStyle(color)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    static func timeString(from date: Date) -> String {
        formatter.string(from: date)
    }

    private static let formatter: DateFormatter = {
        let formatter = DateFormatter()
        // Was hardcoded to Locale(identifier: "tr_TR") — forced Turkish
        // formatting regardless of the device's actual language.
        formatter.locale = .autoupdatingCurrent
        formatter.dateFormat = "HH:mm"
        return formatter
    }()
}

struct ChatPersonaAvatar: View {
    let persona: Persona
    var size: CGFloat = 48

    var body: some View {
        Group {
            if let avatarURL = persona.avatarURL, let url = URL(string: avatarURL) {
                AsyncImage(url: url) { phase in
                    switch phase {
                    case .success(let image):
                        image.resizable().scaledToFill()
                    default:
                        initials
                    }
                }
            } else {
                initials
            }
        }
        .frame(width: size, height: size)
        .clipShape(Circle())
        .shadow(color: .black.opacity(0.04), radius: 3, y: 1)
    }

    private var initials: some View {
        Text(persona.name.prefix(1))
            .font(.system(size: size * 0.4, weight: .semibold, design: .rounded))
            .foregroundStyle(persona.accent)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(persona.accent.opacity(0.22))
    }
}

#Preview {
    VStack(spacing: 16) {
        ChatBubble(
            message: ChatMessage(role: .assistant, content: "Merhaba, nasılsın bugün?"),
            persona: .preview
        )
        ChatBubble(
            message: ChatMessage(role: .user, content: "Bugün işte her şey biraz üstüme geldi.", reactionEmoji: "❤️"),
            persona: .preview
        )
        ChatBubble(
            message: ChatMessage(role: .assistant, content: "", isStreaming: true),
            persona: .preview
        )
    }
    .padding()
    .background(PerchlyPalette.Discover.background)
}

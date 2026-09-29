import SwiftUI

/// A persona card for Keşfet and onboarding. Matches the Stitch layout:
/// gradient avatar ring, category badge, quote, skill line, and a Konuş
/// pill. On Keşfet, `onTalk` makes Konuş a real button (opens chat) while
/// the parent’s tap opens Persona Detayı. Onboarding still uses
/// `showsTalkButton: false` so the whole card is the pick control.
struct PersonaCard: View {
    let persona: Persona
    var isSelected: Bool = false
    var showsTalkButton: Bool = true
    /// When set, the Konuş pill starts a chat instead of following the
    /// card's own tap (which opens Persona Detayı on Keşfet).
    var onTalk: (() -> Void)? = nil

    private var style: PersonaCategoryStyle { persona.categoryStyle }

    var body: some View {
        ZStack(alignment: .topTrailing) {
            Circle()
                .fill(style.glow)
                .frame(width: 160, height: 160)
                .blur(radius: 32)
                .offset(x: 32, y: -32)
                .allowsHitTesting(false)

            VStack(alignment: .leading, spacing: 14) {
                header
                Text("“\(persona.shortDescription)”")
                    .font(PerchlyTypography.Discover.bodyMD)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    .fixedSize(horizontal: false, vertical: true)

                HStack(alignment: .center) {
                    Label(persona.toneDescription, systemImage: style.skillSymbol)
                        .font(PerchlyTypography.Discover.labelSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                        .labelStyle(.titleAndIcon)
                        .lineLimit(1)

                    Spacer(minLength: 8)

                    if let onTalk {
                        talkPill
                            .highPriorityGesture(TapGesture().onEnded(onTalk))
                            .accessibilityAddTraits(.isButton)
                            .accessibilityLabel("\(persona.name) ile konuş")
                            .accessibilityIdentifier("personaTalk_\(persona.id)")
                    } else if showsTalkButton {
                        talkPill
                            .accessibilityHidden(true)
                    } else if isSelected {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundStyle(persona.accent)
                    }
                }
            }
        }
        .padding(24)
        .frame(maxWidth: .infinity, alignment: .leading)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.55)), in: .rect(cornerRadius: 16))
    }

    private var header: some View {
        HStack(alignment: .top, spacing: 12) {
            HStack(alignment: .center, spacing: 12) {
                avatar

                VStack(alignment: .leading, spacing: 2) {
                    Text(persona.name)
                        .font(PerchlyTypography.Discover.headlineSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Text(style.roleTitle)
                        .font(PerchlyTypography.Discover.labelSM)
                        .foregroundStyle(persona.accent)
                    HStack(spacing: 4) {
                        Circle()
                            .fill(persona.isActive ? persona.accent : PerchlyPalette.Discover.onSurfaceVariant)
                            .frame(width: 6, height: 6)
                        Text(persona.isActive ? "Aktif" : "Pasif")
                            .font(PerchlyTypography.Discover.labelSM)
                            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    }
                    .padding(.top, 2)
                }
            }

            Spacer(minLength: 8)

            VStack(alignment: .trailing, spacing: 6) {
                if persona.recommended == true, showsTalkButton {
                    Text("Önerilen")
                        .font(PerchlyTypography.Discover.labelSM.weight(.semibold))
                        .foregroundStyle(PerchlyPalette.Discover.primary)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 4)
                        .background(PerchlyPalette.Discover.primaryFixed.opacity(0.7), in: Capsule())
                }
                Text(style.badgeTitle)
                    .font(PerchlyTypography.Discover.labelSM)
                    .foregroundStyle(style.badgeText)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 4)
                    .background(style.badgeFill, in: Capsule())
            }
        }
    }

    private var talkPill: some View {
        HStack(spacing: 6) {
            Text("Konuş")
            Image(systemName: "arrow.forward")
                .font(.system(size: 12, weight: .semibold))
        }
        .font(PerchlyTypography.Discover.labelMD.weight(.semibold))
        .foregroundStyle(PerchlyPalette.Discover.onPrimary)
        .padding(.horizontal, 20)
        .frame(height: 40)
        .background(style.talkFill, in: Capsule())
    }

    private var avatar: some View {
        ZStack(alignment: .bottomTrailing) {
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
            .frame(width: 52, height: 52)
            .clipShape(Circle())
            .padding(2)
            .background(
                LinearGradient(colors: [style.ringStart, style.ringEnd], startPoint: .bottomLeading, endPoint: .topTrailing),
                in: Circle()
            )

            Circle()
                .fill(persona.accent)
                .frame(width: 14, height: 14)
                .overlay {
                    Circle().stroke(PerchlyPalette.Discover.surfaceLowest, lineWidth: 2)
                }
        }
        .shadow(color: .black.opacity(0.04), radius: 4, y: 2)
    }

    private var initials: some View {
        Text(persona.name.prefix(1))
            .font(PerchlyTypography.Discover.headlineSM)
            .foregroundStyle(persona.accent)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(persona.accent.opacity(0.22))
    }
}

#Preview {
    ScrollView {
        VStack(spacing: 16) {
            PersonaCard(persona: .preview, isSelected: true)
            PersonaCard(
                persona: Persona(
                    id: "696a17eb-78b8-46d4-9ce4-22532fb71fdb",
                    slug: "daily-companion",
                    name: "Mira",
                    category: "daily_companion",
                    shortDescription: "Günün nasıl geçtiğini soran, sohbeti uzatmayı seven, samimi bir arkadaş.",
                    toneDescription: "Sıcak, meraklı, gündelik dilde konuşan, yargılamayan bir dinleyici.",
                    avatarURL: nil,
                    accentColor: "#4A90D9",
                    isMinorAppropriate: true,
                    isActive: true,
                    sortOrder: 2,
                    createdAt: .now,
                    updatedAt: .now,
                    recommended: nil,
                    matchReason: nil
                ),
                showsTalkButton: true
            )
        }
        .padding()
    }
    .background(PerchlyPalette.Discover.background)
}

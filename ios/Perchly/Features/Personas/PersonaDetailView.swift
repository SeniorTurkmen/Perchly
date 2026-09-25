import SwiftUI

/// Stitch "Persona Detayı" chrome, bound to a real `GET /personas` row.
/// Skips mock-only widgets (uyum %, voice sliders, lifestyle photo).
struct PersonaDetailView: View {
    let persona: Persona
    let onStartChat: () -> Void
    @StateObject private var traitsViewModel: PersonaTraitsViewModel

    init(persona: Persona, onStartChat: @escaping () -> Void) {
        self.persona = persona
        self.onStartChat = onStartChat
        _traitsViewModel = StateObject(wrappedValue: PersonaTraitsViewModel(personaID: persona.id))
    }

    private var style: PersonaCategoryStyle { persona.categoryStyle }

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                auraCard
                startChatButton
                personalityDialsSection
                traitsSection
                aboutCard
                privacyRow
                atmosphereCard
                disclaimer
            }
            .padding(.horizontal, 20)
            .padding(.top, 12)
            .padding(.bottom, 32)
        }
        .scrollIndicators(.hidden)
        .background {
            DiscoverAmbientBackground()
        }
        .background(PerchlyPalette.Discover.background)
        .navigationTitle("Persona Detay & Uyum")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar(.visible, for: .navigationBar)
        .toolbarBackground(PerchlyPalette.Discover.surface.opacity(0.82), for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .principal) {
                VStack(spacing: 2) {
                    Text("Persona Detay & Uyum")
                        .font(PerchlyTypography.Discover.headlineSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    HStack(spacing: 4) {
                        Circle()
                            .fill(persona.isActive ? persona.accent : PerchlyPalette.Discover.onSurfaceVariant)
                            .frame(width: 6, height: 6)
                        Text(persona.isActive ? "Aktif Seans" : "\(persona.name)")
                            .font(PerchlyTypography.Discover.labelSM)
                            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    }
                }
            }
        }
        .preferredColorScheme(.light)
    }

    private var auraCard: some View {
        VStack(spacing: 16) {
            avatar

            VStack(spacing: 6) {
                HStack(spacing: 6) {
                    Image(systemName: "heart.fill")
                        .font(.system(size: 12))
                    Text(style.badgeTitle)
                        .font(PerchlyTypography.Discover.labelSM.weight(.semibold))
                }
                .foregroundStyle(style.badgeText)
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(style.badgeFill, in: Capsule())

                Text(persona.name)
                    .font(PerchlyTypography.Discover.headlineLG)
                    .foregroundStyle(PerchlyPalette.Discover.onSurface)

                Text(style.roleTitle)
                    .font(PerchlyTypography.Discover.bodyMD)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
            }

            HStack(spacing: 12) {
                HStack(spacing: 10) {
                    ZStack {
                        Circle().fill(persona.accent.opacity(0.18))
                        Image(systemName: style.skillSymbol)
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundStyle(persona.accent)
                    }
                    .frame(width: 36, height: 36)

                    VStack(alignment: .leading, spacing: 2) {
                        Text("İletişim")
                            .font(PerchlyTypography.Discover.labelSM)
                            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                        Text(style.skillTitle)
                            .font(PerchlyTypography.Discover.labelLG)
                            .foregroundStyle(PerchlyPalette.Discover.onSurface)
                            .lineLimit(2)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }

                Spacer(minLength: 8)

                HStack(spacing: 4) {
                    Image(systemName: persona.isActive ? "checkmark.shield.fill" : "pause.circle.fill")
                        .font(.system(size: 14))
                    Text(persona.isActive ? "Aktif" : "Pasif")
                        .font(PerchlyTypography.Discover.labelSM.weight(.semibold))
                }
                .foregroundStyle(persona.accent)
            }
            .padding(10)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PerchlyPalette.Discover.surfaceLow.opacity(0.7), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        }
        .padding(24)
        .frame(maxWidth: .infinity)
        .background {
            ZStack {
                Circle()
                    .fill(PerchlyPalette.Discover.secondaryFixedDim.opacity(0.35))
                    .frame(width: 180, height: 180)
                    .blur(radius: 28)
                    .offset(x: 90, y: -70)
                Circle()
                    .fill(persona.accent.opacity(0.22))
                    .frame(width: 200, height: 200)
                    .blur(radius: 32)
                    .offset(x: -80, y: 90)
            }
            .allowsHitTesting(false)
        }
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.7)), in: .rect(cornerRadius: 20))
    }

    private var avatar: some View {
        ZStack {
            Circle()
                .fill(
                    LinearGradient(
                        colors: [style.ringStart, style.ringEnd, PerchlyPalette.Discover.secondaryContainer],
                        startPoint: .bottomLeading,
                        endPoint: .topTrailing
                    )
                )
                .frame(width: 148, height: 148)
                .blur(radius: 10)
                .opacity(0.7)

            ChatPersonaAvatar(persona: persona, size: 128)
                .padding(4)
                .background(PerchlyPalette.Discover.surfaceLowest.opacity(0.9), in: Circle())
                .overlay(alignment: .bottomTrailing) {
                    ZStack {
                        Circle()
                            .fill(PerchlyPalette.Discover.surfaceLowest)
                            .frame(width: 22, height: 22)
                        Circle()
                            .fill(persona.isActive ? persona.accent : PerchlyPalette.Discover.onSurfaceVariant)
                            .frame(width: 12, height: 12)
                    }
                    .offset(x: -8, y: -8)
                }
        }
    }

    private var startChatButton: some View {
        Button(action: onStartChat) {
            HStack(spacing: 10) {
                Image(systemName: "bubble.left.fill")
                    .font(.system(size: 16, weight: .semibold))
                Text("\(persona.name) ile Sohbete Başla")
                    .font(PerchlyTypography.Discover.labelLG)
                Image(systemName: "arrow.right")
                    .font(.system(size: 13, weight: .semibold))
            }
            .foregroundStyle(PerchlyPalette.Discover.onPrimary)
            .frame(maxWidth: .infinity)
            .frame(height: 54)
            .background(PerchlyPalette.Discover.primary, in: Capsule())
        }
        .buttonStyle(.plain)
        .shadow(color: PerchlyPalette.Discover.primary.opacity(0.28), radius: 16, y: 8)
        .background {
            Capsule()
                .fill(
                    LinearGradient(
                        colors: [
                            PerchlyPalette.Discover.primaryContainer,
                            PerchlyPalette.Discover.secondaryContainer,
                            PerchlyPalette.Discover.primaryFixedDim,
                        ],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                )
                .blur(radius: 12)
                .opacity(0.55)
                .padding(-4)
                .allowsHitTesting(false)
        }
        .accessibilityIdentifier("personaDetailStartChat")
        .accessibilityLabel("\(persona.name) ile sohbete başla")
    }

    /// The user's own personality-dial customization for this persona
    /// (Perchmate) — separate from `traitsSection` below, which is just
    /// decorative, category-level copy. Each slider saves the instant
    /// its drag ends, so "changed" and "applied" are the same moment;
    /// there's nothing to remember to hit save on. Changes take effect
    /// starting with the user's very next message to this persona.
    private var personalityDialsSection: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack {
                Text("Kişilik Ayarları")
                    .font(PerchlyTypography.Discover.labelLG)
                    .foregroundStyle(PerchlyPalette.Discover.onSurface)
                Spacer()
                if traitsViewModel.isCustomized {
                    Button("Varsayılana Döndür") {
                        Task { await traitsViewModel.reset() }
                    }
                    .font(PerchlyTypography.Discover.labelSM.weight(.semibold))
                    .foregroundStyle(persona.accent)
                    .accessibilityIdentifier("personaTraitsResetButton")
                }
            }

            Text("\(persona.name)'ın sana nasıl karşılık verdiğini kendine göre ayarla. Örneğin sıcaklığı yükseltip doğrudanlığı düşürerek daha yumuşak, ya da tam tersini yaparak daha sert yanıtlar alabilirsin.")
                .font(PerchlyTypography.Discover.bodySM)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .fixedSize(horizontal: false, vertical: true)

            VStack(spacing: 18) {
                traitSlider(key: "warmth", title: "Sıcaklık", systemImage: "heart.fill", value: $traitsViewModel.traits.warmth)
                traitSlider(key: "humor", title: "Espri", systemImage: "face.smiling.fill", value: $traitsViewModel.traits.humor)
                traitSlider(key: "wisdom", title: "Bilgelik", systemImage: "brain.head.profile", value: $traitsViewModel.traits.wisdom)
                traitSlider(key: "directness", title: "Doğrudanlık", systemImage: "bolt.fill", value: $traitsViewModel.traits.directness)
                traitSlider(key: "energy", title: "Enerji", systemImage: "sparkles", value: $traitsViewModel.traits.energy)
            }

            if let errorMessage = traitsViewModel.errorMessage {
                Text(errorMessage)
                    .font(PerchlyTypography.Discover.bodySM)
                    .foregroundStyle(PerchlyPalette.Discover.secondary)
            }
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.65)), in: .rect(cornerRadius: 16))
        .task { await traitsViewModel.loadIfNeeded() }
    }

    private func traitSlider(key: String, title: LocalizedStringKey, systemImage: String, value: Binding<Int>) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Label(title, systemImage: systemImage)
                    .font(PerchlyTypography.Discover.labelMD)
                    .foregroundStyle(PerchlyPalette.Discover.onSurface)
                Spacer()
                Text(verbatim: "%\(value.wrappedValue)")
                    .font(PerchlyTypography.Discover.labelMD.weight(.semibold))
                    .foregroundStyle(persona.accent)
                    .monospacedDigit()
            }

            Slider(
                value: Binding(
                    get: { Double(value.wrappedValue) },
                    set: { value.wrappedValue = Int($0.rounded()) }
                ),
                in: 0...100,
                step: 1,
                onEditingChanged: { isEditing in
                    if !isEditing {
                        Task { await traitsViewModel.save() }
                    }
                }
            )
            .tint(persona.accent)
            .accessibilityIdentifier("traitSlider_\(key)")
        }
    }

    private var traitsSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("İletişim Tarzı & Nitelikler")
                    .font(PerchlyTypography.Discover.labelLG)
                    .foregroundStyle(PerchlyPalette.Discover.onSurface)
                Spacer()
                Text("\(String(persona.detailTraits.count)) Aktif Nitelik")
                    .font(PerchlyTypography.Discover.labelSM)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
            }

            FlowChips(items: persona.detailTraits)
        }
    }

    private var aboutCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                Image(systemName: "sparkles")
                    .foregroundStyle(PerchlyPalette.Discover.secondary)
                Text("\(persona.name) Hakkında")
                    .font(PerchlyTypography.Discover.headlineSM.weight(.semibold))
                    .foregroundStyle(PerchlyPalette.Discover.onSurface)
            }

            Text(persona.shortDescription)
                .font(PerchlyTypography.Discover.bodyMD)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .fixedSize(horizontal: false, vertical: true)

            Text(persona.toneDescription)
                .font(PerchlyTypography.Discover.bodySM)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background {
            Circle()
                .fill(PerchlyPalette.Discover.secondaryFixed.opacity(0.35))
                .frame(width: 112, height: 112)
                .blur(radius: 24)
                .offset(x: 80, y: -40)
                .allowsHitTesting(false)
        }
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.65)), in: .rect(cornerRadius: 16))
    }

    private var privacyRow: some View {
        HStack(alignment: .top, spacing: 8) {
            privacyCard(
                symbol: "sparkles",
                tint: PerchlyPalette.Discover.primary,
                title: "Sohbet kaydı",
                body: String(localized: "Aynı konuşmada kaldığın sürece mesajların burada durur.")
            )
            privacyCard(
                symbol: "moon.stars.fill",
                tint: PerchlyPalette.Discover.secondary,
                title: "Sıfır yargı",
                body: persona.detailZeroJudgmentLine
            )
        }
    }

    private func privacyCard(symbol: String, tint: Color, title: LocalizedStringKey, body: String) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Image(systemName: symbol)
                .font(.system(size: 16))
                .foregroundStyle(tint)
            Text(title)
                .font(PerchlyTypography.Discover.labelMD.weight(.semibold))
                .foregroundStyle(PerchlyPalette.Discover.onSurface)
            Text(body)
                .font(PerchlyTypography.Discover.bodySM)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PerchlyPalette.Discover.surfaceLow.opacity(0.75), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    private var atmosphereCard: some View {
        ZStack(alignment: .bottomLeading) {
            LinearGradient(
                colors: [
                    persona.accent.opacity(0.85),
                    PerchlyPalette.Discover.onSurface.opacity(0.72),
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )

            HStack(spacing: 10) {
                Image(systemName: style.skillSymbol)
                    .font(.system(size: 18))
                Text(persona.detailAtmosphereLine)
                    .font(PerchlyTypography.Discover.labelMD)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .foregroundStyle(PerchlyPalette.Discover.surfaceLowest)
            .padding(16)
        }
        .frame(maxWidth: .infinity, minHeight: 120, alignment: .bottomLeading)
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    private var disclaimer: some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "info.circle")
                .font(.system(size: 16))
                .padding(.top, 1)
            Text("\(persona.name), gerçek insan bağlarını ikame etmez; duygu dünyanı düzenlemene yardımcı olan destekleyici bir yoldaştır.")
                .font(PerchlyTypography.Discover.bodySM)
                .fixedSize(horizontal: false, vertical: true)
        }
        .foregroundStyle(PerchlyPalette.Discover.secondary)
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PerchlyPalette.Discover.secondaryFixed.opacity(0.35), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
    }
}

/// Lightweight wrap layout for the trait chips. Enough for four items
/// without pulling in a flow-layout package.
private struct FlowChips: View {
    let items: [(title: String, symbol: String, tint: Color)]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            ForEach(rows, id: \.0) { _, row in
                HStack(spacing: 8) {
                    ForEach(row, id: \.title) { item in
                        chip(item)
                    }
                }
            }
        }
    }

    private var rows: [(Int, [(title: String, symbol: String, tint: Color)])] {
        stride(from: 0, to: items.count, by: 2).map { start in
            let end = min(start + 2, items.count)
            return (start, Array(items[start..<end]))
        }
    }

    private func chip(_ item: (title: String, symbol: String, tint: Color)) -> some View {
        HStack(spacing: 6) {
            Image(systemName: item.symbol)
                .font(.system(size: 13))
                .foregroundStyle(item.tint)
            Text(item.title)
                .font(PerchlyTypography.Discover.labelSM)
                .foregroundStyle(PerchlyPalette.Discover.onSurface)
                .lineLimit(2)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .frame(maxWidth: .infinity, alignment: .leading)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.85)), in: .capsule)
    }
}

#Preview {
    NavigationStack {
        PersonaDetailView(persona: .preview) {}
    }
}

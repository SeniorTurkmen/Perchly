import SwiftUI

/// Stitch "Onboarding - Persona Eşleşme Önizlemesi". Mood is still
/// local here (the onboarding profile isn't POSTed until a persona is
/// picked), so the list is reordered client-side. After onboarding,
/// Keşfet uses `GET /personas?recommend=true` for the same mapping.
struct PersonaPickStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @StateObject private var viewModel = PersonaListViewModel()

    var body: some View {
        OnboardingScaffold(stepIndex: 5, stepLabel: "Eşleşme", onBack: coordinator.goBack) {
            VStack(alignment: .leading, spacing: 20) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Sana bir Perchmate önerdik")
                        .font(PerchlyTypography.Discover.headlineLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Text(matchCopy)
                        .font(PerchlyTypography.Discover.bodyMD)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }

                content
            }
        } footer: {
            EmptyView()
        }
        .task {
            await viewModel.load(prioritizedCategory: coordinator.profile.moodPreference?.personaCategory)
        }
    }

    private var matchCopy: String {
        if let reason = featuredPersona?.matchReason, !reason.isEmpty {
            return "\(reason) İstersen başka birini seçebilirsin."
        }
        switch coordinator.profile.moodPreference {
        case .motivation:
            return "Motive olmak istediğin için bu eşleşmeyi öne çıkardık. İstersen başka birini seçebilirsin."
        case .dailyChat:
            return "Gündelik sohbet aradığın için bu eşleşmeyi öne çıkardık. İstersen başka birini seçebilirsin."
        case .hobbyTalk:
            return "Hobi paylaşmak istediğin için bu eşleşmeyi öne çıkardık. İstersen başka birini seçebilirsin."
        default:
            return "Seçimini istediğin zaman değiştirebilirsin."
        }
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.state {
        case .idle, .loading:
            ProgressView()
                .frame(maxWidth: .infinity, minHeight: 220)
        case .failed(let message):
            errorView(message)
        case .loaded:
            personaList
        }
    }

    private var personaList: some View {
        VStack(spacing: 16) {
            ForEach(visiblePersonas) { persona in
                Button {
                    coordinator.selectPersona(persona)
                } label: {
                    VStack(alignment: .leading, spacing: 12) {
                        if persona.id == featuredPersona?.id {
                            Text("Önerilen eşleşme")
                                .font(PerchlyTypography.Discover.labelSM.weight(.semibold))
                                .foregroundStyle(PerchlyPalette.Discover.primary)
                                .padding(.horizontal, 10)
                                .padding(.vertical, 4)
                                .background(PerchlyPalette.Discover.primaryFixed.opacity(0.7), in: Capsule())
                        }
                        PersonaCard(persona: persona, showsTalkButton: false)
                    }
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("personaOption_\(persona.id)")
            }
        }
    }

    /// Minors never see a persona the backend wouldn't let them start a
    /// conversation with anyway (see Persona.isMinorAppropriate).
    private var visiblePersonas: [Persona] {
        guard coordinator.profile.isMinor else { return viewModel.personas }
        return viewModel.personas.filter(\.isMinorAppropriate)
    }

    private var featuredPersona: Persona? {
        visiblePersonas.first(where: { $0.recommended == true }) ?? visiblePersonas.first
    }

    private func errorView(_ message: String) -> some View {
        VStack(spacing: 12) {
            Text("Bir şeyler ters gitti")
                .font(PerchlyTypography.title)
            Text(message)
                .font(PerchlyTypography.body)
                .foregroundStyle(PerchlyPalette.textSecondary)
                .multilineTextAlignment(.center)
            GlassButton(title: "Tekrar dene") {
                Task { await viewModel.load(prioritizedCategory: coordinator.profile.moodPreference?.personaCategory) }
            }
        }
        .padding()
        .frame(maxWidth: .infinity, minHeight: 220)
    }
}

#Preview {
    PersonaPickStepView(coordinator: OnboardingCoordinator())
}

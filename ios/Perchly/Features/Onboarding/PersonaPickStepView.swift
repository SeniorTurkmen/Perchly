import SwiftUI

/// Screen 4, the last link in the chain: personalized persona
/// suggestions (reordered by `PersonaListViewModel`, using the mood
/// preference from screen 2 — see
/// `OnboardingProfile.MoodPreference.personaCategory`), filtered for a
/// minor's own onboarding answer so a restricted persona is never even
/// shown (the backend also enforces this server-side when starting a
/// conversation — this is belt-and-suspenders on the client). Tapping a
/// persona hands off straight to the coordinator; this screen never
/// navigates anywhere itself.
struct PersonaPickStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @StateObject private var viewModel = PersonaListViewModel()

    var body: some View {
        VStack(spacing: 16) {
            VStack(alignment: .leading, spacing: 8) {
                Text("Kiminle konuşmak istersin?")
                    .font(PerchlyTypography.largeTitle)
                Text("Seçimini istediğin zaman değiştirebilirsin.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding([.horizontal, .top])

            content
        }
        .task {
            await viewModel.load(prioritizedCategory: coordinator.profile.moodPreference?.personaCategory)
        }
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.state {
        case .idle, .loading:
            ProgressView()
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .failed(let message):
            errorView(message)
        case .loaded:
            personaList
        }
    }

    private var personaList: some View {
        ScrollView {
            VStack(spacing: 16) {
                ForEach(visiblePersonas) { persona in
                    Button {
                        coordinator.selectPersona(persona)
                    } label: {
                        PersonaCard(persona: persona, showsTalkButton: false)
                    }
                    .buttonStyle(.plain)
                    .accessibilityIdentifier("personaOption_\(persona.id)")
                }
            }
            .padding()
        }
    }

    /// Minors never see a persona the backend wouldn't let them start a
    /// conversation with anyway (see Persona.isMinorAppropriate).
    private var visiblePersonas: [Persona] {
        guard coordinator.profile.isMinor else { return viewModel.personas }
        return viewModel.personas.filter(\.isMinorAppropriate)
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
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

#Preview {
    PersonaPickStepView(coordinator: OnboardingCoordinator())
}

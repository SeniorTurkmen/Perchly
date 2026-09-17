import SwiftUI

struct PersonaListView: View {
    @StateObject private var viewModel: PersonaListViewModel
    @State private var isShowingProfile = false
    @State private var path = NavigationPath()
    @State private var selectedTab: DiscoverTab = .personas
    /// Guards `initialPersona` from being pushed more than once. Without
    /// this, checking `path.isEmpty` alone would re-push it every time
    /// `.task` happens to re-run while the user has legitimately
    /// navigated back to an empty path — silently undoing their own
    /// back button tap.
    @State private var hasAppliedInitialPersona = false

    /// When set (coming straight from onboarding's persona-pick step —
    /// see RootView), pushed onto the navigation stack as soon as the
    /// list loads, so the user lands directly in chat with the persona
    /// they just picked — but with a real, working back button into
    /// this list, unlike showing ChatView in total isolation would give
    /// them.
    private let initialPersona: Persona?

    init(viewModel: PersonaListViewModel = PersonaListViewModel(), initialPersona: Persona? = nil) {
        _viewModel = StateObject(wrappedValue: viewModel)
        self.initialPersona = initialPersona
    }

    var body: some View {
        NavigationStack(path: $path) {
            ZStack {
                content
            }
            .background {
                DiscoverAmbientBackground()
            }
            .background(PerchlyPalette.Discover.background)
            .navigationTitle("Keşfet Personalar")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar(path.isEmpty ? .hidden : .automatic, for: .navigationBar)
            .navigationDestination(for: PersonaDetailRoute.self) { route in
                PersonaDetailView(persona: route.persona) {
                    openChat(persona: route.persona)
                }
            }
            .navigationDestination(for: ChatRoute.self) { route in
                ChatView(persona: route.persona, conversationID: route.conversationID)
            }
            .sheet(isPresented: $isShowingProfile) {
                ProfileView()
            }
            .task {
                let shouldOpenInitialChat = !hasAppliedInitialPersona && initialPersona != nil
                if shouldOpenInitialChat {
                    hasAppliedInitialPersona = true
                }
                await viewModel.load()
                await viewModel.loadConversations()
                await viewModel.loadChatEnergy()
                if shouldOpenInitialChat, let initialPersona {
                    openChat(persona: initialPersona)
                }
            }
            .onChange(of: path.count) {
                if path.isEmpty {
                    Task {
                        await viewModel.loadConversations()
                        await viewModel.loadChatEnergy()
                    }
                }
            }
            .preferredColorScheme(.light)
        }
    }

    @ViewBuilder
    private var content: some View {
        VStack(spacing: 0) {
            if selectedTab == .personas {
                discoverHeader
                switch viewModel.state {
                case .idle, .loading:
                    ProgressView()
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                case .failed(let message):
                    errorView(message) {
                        await viewModel.load()
                        await viewModel.loadChatEnergy()
                    }
                case .loaded:
                    personaList
                }
            } else {
                loadedBody
            }
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            if path.isEmpty {
                DiscoverTabBar(selectedTab: $selectedTab, onProfile: { isShowingProfile = true })
            }
        }
    }

    @ViewBuilder
    private var loadedBody: some View {
        switch selectedTab {
        case .personas:
            personaList
        case .chats:
            conversationInbox
        case .journal:
            comingSoon(
                title: "Günlük",
                systemImage: "book.fill",
                message: "Günlük düşüncelerin için sakin bir yer hazırlıyoruz."
            )
        }
    }

    private var discoverHeader: some View {
        HStack(spacing: 8) {
            HStack(spacing: 8) {
                ZStack {
                    Circle()
                        .fill(
                            LinearGradient(
                                colors: [PerchlyPalette.Discover.primaryContainer, PerchlyPalette.Discover.secondaryFixed],
                                startPoint: .bottomLeading,
                                endPoint: .topTrailing
                            )
                        )
                    Image(systemName: "sparkles")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(PerchlyPalette.Discover.primary)
                }
                .frame(width: 40, height: 40)
                .padding(2)
                .background(PerchlyPalette.Discover.surfaceLowest.opacity(0.8), in: Circle())
                .shadow(color: .black.opacity(0.04), radius: 4, y: 2)

                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 6) {
                        Circle()
                            .fill(PerchlyPalette.Discover.primary)
                            .frame(width: 8, height: 8)
                        Text("Keşfet Personalar")
                            .font(PerchlyTypography.Discover.headlineSM)
                            .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    }
                    Text("Seni dinlemeye hazır")
                        .font(PerchlyTypography.Discover.labelSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }
            }

            Spacer(minLength: 8)

            Button {
                isShowingProfile = true
            } label: {
                ZStack {
                    Circle().fill(PerchlyPalette.Discover.primary)
                    Image(systemName: "person.fill")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(PerchlyPalette.Discover.onPrimary)
                }
                .frame(width: 32, height: 32)
                .shadow(color: PerchlyPalette.Discover.primary.opacity(0.2), radius: 4, y: 1)
            }
            .accessibilityIdentifier("profileButton")
            .accessibilityLabel("Profil")
        }
        .padding(.horizontal, 20)
        .padding(.top, 8)
        .padding(.bottom, 12)
        .background(PerchlyPalette.Discover.surface.opacity(0.7))
        .background(.ultraThinMaterial)
    }

    private var personaList: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                greeting
                if let energy = viewModel.chatEnergy {
                    energyCapsule(energy)
                }
                categoryFilters

                VStack(spacing: 16) {
                    ForEach(viewModel.filteredPersonas) { persona in
                        Button {
                            openDetail(persona)
                        } label: {
                            PersonaCard(
                                persona: persona,
                                isSelected: viewModel.selectedPersonaID == persona.id,
                                onTalk: { openChat(persona: persona) }
                            )
                        }
                        .buttonStyle(.plain)
                        .accessibilityHint("Detayı aç")
                        .accessibilityIdentifier("personaCard_\(persona.id)")
                    }
                }

                trustFooter
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
            .padding(.bottom, 24)
        }
        .scrollIndicators(.hidden)
    }

    private var conversationInbox: some View {
        Group {
            switch viewModel.conversationsState {
            case .idle, .loading:
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            case .failed(let message):
                errorView(message) {
                    await viewModel.loadConversations()
                }
            case .loaded:
                if viewModel.conversations.isEmpty {
                    comingSoon(
                        title: "Sohbetler",
                        systemImage: "bubble.left.and.bubble.right.fill",
                        message: "Henüz sohbet yok. Personalar’dan biriyle konuşmaya başla."
                    )
                } else {
                    conversationList
                }
            }
        }
        .task {
            await viewModel.loadConversations()
        }
    }

    private var conversationList: some View {
        ScrollView {
            LazyVStack(spacing: 10) {
                ForEach(viewModel.conversations) { conversation in
                    Button {
                        openChat(persona: conversation.persona, conversationID: conversation.id)
                    } label: {
                        conversationRow(conversation)
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
            .padding(.bottom, 24)
        }
        .scrollIndicators(.hidden)
    }

    private func conversationRow(_ conversation: ConversationSummary) -> some View {
        HStack(spacing: 12) {
            ChatPersonaAvatar(persona: conversation.persona, size: 48)

            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text(conversation.persona.name)
                        .font(PerchlyTypography.Discover.headlineSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Spacer(minLength: 8)
                    Text(ChatBubble.timeString(from: conversation.lastMessage?.createdAt ?? conversation.updatedAt))
                        .font(PerchlyTypography.Discover.labelSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }
                Text(conversation.lastMessage?.content ?? "Henüz mesaj yok")
                    .font(PerchlyTypography.Discover.bodySM)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    .lineLimit(2)
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.55)), in: .rect(cornerRadius: 18))
    }

    private var greeting: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                ZStack {
                    Circle().fill(PerchlyPalette.Discover.secondaryFixed)
                    Image(systemName: "heart.fill")
                        .font(.system(size: 10))
                        .foregroundStyle(PerchlyPalette.Discover.onSecondaryFixed)
                }
                .frame(width: 24, height: 24)
                Text("Yargısız, Şefkatli Alan")
                    .font(PerchlyTypography.Discover.labelMD)
                    .foregroundStyle(PerchlyPalette.Discover.secondary)
            }

            Text("Bugün kiminle dertleşmek istersin?")
                .font(PerchlyTypography.Discover.headlineLG)
                .foregroundStyle(PerchlyPalette.Discover.onSurface)
            Text("Seni dinlemeye her zaman hazır, sakin bir dost var.")
                .font(PerchlyTypography.Discover.bodyMD)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
        }
    }

    private func energyCapsule(_ energy: ChatEnergy) -> some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .stroke(PerchlyPalette.Discover.surfaceContainerHigh, lineWidth: 3.5)
                Circle()
                    .trim(from: 0, to: energy.progress)
                    .stroke(PerchlyPalette.Discover.primaryContainer, style: StrokeStyle(lineWidth: 3.5, lineCap: .round))
                    .rotationEffect(.degrees(-90))
                Text(energy.fractionLabel)
                    .font(PerchlyTypography.Discover.labelSM.weight(.bold))
                    .foregroundStyle(PerchlyPalette.Discover.primary)
            }
            .frame(width: 40, height: 40)

            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(PerchlyPalette.Discover.primary)
                        .frame(width: 6, height: 6)
                    Text("Günün Sohbet Enerjisi")
                        .font(PerchlyTypography.Discover.labelLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                        .lineLimit(1)
                }
                Text(energy.shareCopy)
                    .font(PerchlyTypography.Discover.bodySM)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .layoutPriority(1)

            Spacer(minLength: 8)

            HStack(spacing: 4) {
                Image(systemName: "drop.fill")
                    .font(.system(size: 11))
                Text(energy.moodLabel)
            }
            .font(PerchlyTypography.Discover.labelSM)
            .foregroundStyle(PerchlyPalette.Discover.primary)
            .padding(.horizontal, 10)
            .padding(.vertical, 4)
            .background(PerchlyPalette.Discover.surfaceLow, in: Capsule())
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.5)), in: .rect(cornerRadius: 16))
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Günün sohbet enerjisi \(energy.fractionLabel), \(energy.moodLabel)")
    }

    private var categoryFilters: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                filterPill(title: "Tümü", isSelected: viewModel.selectedCategory == nil, leading: .chevron) {
                    viewModel.selectCategory(nil)
                }
                ForEach(viewModel.availableCategories, id: \.self) { category in
                    let style = PersonaCategoryStyle.style(for: category)
                    filterPill(title: style.filterTitle, isSelected: viewModel.selectedCategory == category, leading: .dot(style.filterDot)) {
                        viewModel.selectCategory(category)
                    }
                }
            }
            .padding(.vertical, 4)
        }
    }

    private enum FilterLeading {
        case chevron
        case dot(Color)
    }

    private func filterPill(title: String, isSelected: Bool, leading: FilterLeading, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack(spacing: 6) {
                switch leading {
                case .chevron:
                    Image(systemName: "chevron.left")
                        .font(.system(size: 10, weight: .semibold))
                case .dot(let color):
                    Circle().fill(color).frame(width: 8, height: 8)
                }
                Text(title)
                    .lineLimit(1)
            }
            .font(PerchlyTypography.Discover.labelMD)
            .foregroundStyle(isSelected ? PerchlyPalette.Discover.onPrimary : PerchlyPalette.Discover.onSurfaceVariant)
            .padding(.horizontal, 16)
            .frame(height: 36)
            .background {
                if isSelected {
                    Capsule().fill(PerchlyPalette.Discover.primary)
                } else {
                    Capsule().fill(PerchlyPalette.Discover.surfaceLowest.opacity(0.8))
                }
            }
            .shadow(color: isSelected ? PerchlyPalette.Discover.primary.opacity(0.18) : .clear, radius: 4, y: 1)
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }

    private var trustFooter: some View {
        VStack(spacing: 8) {
            ZStack {
                Circle().fill(PerchlyPalette.Discover.primaryFixed)
                Image(systemName: "checkmark.shield.fill")
                    .font(.system(size: 14))
                    .foregroundStyle(PerchlyPalette.Discover.primary)
            }
            .frame(width: 28, height: 28)
            Text("Güvenli ve Gizli Liman")
                .font(PerchlyTypography.Discover.labelMD.weight(.semibold))
                .foregroundStyle(PerchlyPalette.Discover.onSurface)
            Text("İnsan ilişkilerinin tamamlayıcısı, her an yanında sakin bir sığınak.")
                .font(PerchlyTypography.Discover.bodySM)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .multilineTextAlignment(.center)
                .frame(maxWidth: 280)
        }
        .padding(16)
        .frame(maxWidth: .infinity)
        .background(PerchlyPalette.Discover.surfaceLow.opacity(0.7), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    private func comingSoon(title: String, systemImage: String, message: String) -> some View {
        VStack(spacing: 12) {
            Image(systemName: systemImage)
                .font(.system(size: 28))
                .foregroundStyle(PerchlyPalette.Discover.primary)
            Text(title)
                .font(PerchlyTypography.Discover.headlineSM)
                .foregroundStyle(PerchlyPalette.Discover.onSurface)
            Text(message)
                .font(PerchlyTypography.Discover.bodyMD)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .multilineTextAlignment(.center)
        }
        .padding(32)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func openDetail(_ persona: Persona) {
        viewModel.select(persona)
        path.append(PersonaDetailRoute(persona: persona))
    }

    private func openChat(persona: Persona, conversationID: String? = nil) {
        viewModel.select(persona)
        let threadID = conversationID ?? viewModel.existingConversation(for: persona)?.id
        path.append(ChatRoute(persona: persona, conversationID: threadID))
    }

    private func errorView(_ message: String, retry: @escaping () async -> Void) -> some View {
        VStack(spacing: 12) {
            Text("Bir şeyler ters gitti")
                .font(PerchlyTypography.title)
            Text(message)
                .font(PerchlyTypography.body)
                .foregroundStyle(PerchlyPalette.textSecondary)
                .multilineTextAlignment(.center)
            GlassButton(title: "Tekrar dene") {
                Task { await retry() }
            }
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

private struct PersonaDetailRoute: Hashable {
    let persona: Persona

    static func == (lhs: PersonaDetailRoute, rhs: PersonaDetailRoute) -> Bool {
        lhs.persona.id == rhs.persona.id
    }

    func hash(into hasher: inout Hasher) {
        hasher.combine(persona.id)
    }
}

private struct ChatRoute: Hashable {
    let persona: Persona
    var conversationID: String?

    static func == (lhs: ChatRoute, rhs: ChatRoute) -> Bool {
        lhs.persona.id == rhs.persona.id && lhs.conversationID == rhs.conversationID
    }

    func hash(into hasher: inout Hasher) {
        hasher.combine(persona.id)
        hasher.combine(conversationID)
    }
}

private enum DiscoverTab {
    case personas, chats, journal
}

private struct DiscoverTabBar: View {
    @Binding var selectedTab: DiscoverTab
    let onProfile: () -> Void

    var body: some View {
        HStack(spacing: 0) {
            tabButton("Personalar", systemImage: "circle.grid.2x2.fill", tab: .personas)
            tabButton("Sohbetler", systemImage: "bubble.left.and.bubble.right.fill", tab: .chats)
            tabButton("Günlük", systemImage: "book.fill", tab: .journal)
            Button(action: onProfile) {
                tabLabel("Samimi Kulüp", systemImage: "heart.fill", isSelected: false)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("clubTabButton")
        }
        .padding(6)
        .background(PerchlyPalette.Discover.surface.opacity(0.75))
        .background(.ultraThinMaterial)
        .clipShape(Capsule())
        .shadow(color: PerchlyPalette.Discover.primary.opacity(0.08), radius: 18, y: 8)
        .padding(.horizontal, 20)
        .padding(.bottom, 12)
    }

    private func tabButton(_ title: String, systemImage: String, tab: DiscoverTab) -> some View {
        Button {
            selectedTab = tab
        } label: {
            tabLabel(title, systemImage: systemImage, isSelected: selectedTab == tab)
        }
        .buttonStyle(.plain)
    }

    private func tabLabel(_ title: String, systemImage: String, isSelected: Bool) -> some View {
        VStack(spacing: 2) {
            Image(systemName: systemImage)
                .font(.system(size: 18))
            Text(title)
                .font(PerchlyTypography.Discover.labelSM)
                .lineLimit(1)
                .minimumScaleFactor(0.8)
        }
        .foregroundStyle(isSelected ? PerchlyPalette.Discover.primary : PerchlyPalette.Discover.onSurfaceVariant)
        .frame(maxWidth: .infinity)
        .frame(height: 52)
        .background {
            if isSelected {
                Capsule().fill(PerchlyPalette.Discover.surfaceLowest.opacity(0.9))
            }
        }
    }
}

struct DiscoverAmbientBackground: View {
    var body: some View {
        ZStack {
            PerchlyPalette.Discover.background
            Circle()
                .fill(PerchlyPalette.Discover.primaryContainer.opacity(0.25))
                .frame(width: 320, height: 320)
                .blur(radius: 60)
                .offset(x: -120, y: -180)
            Circle()
                .fill(PerchlyPalette.Discover.secondaryFixedDim.opacity(0.2))
                .frame(width: 280, height: 280)
                .blur(radius: 60)
                .offset(x: 140, y: 80)
            Circle()
                .fill(PerchlyPalette.Discover.tertiaryFixed.opacity(0.25))
                .frame(width: 360, height: 360)
                .blur(radius: 70)
                .offset(y: 320)
        }
        .ignoresSafeArea()
        .allowsHitTesting(false)
    }
}

#Preview {
    PersonaListView()
}

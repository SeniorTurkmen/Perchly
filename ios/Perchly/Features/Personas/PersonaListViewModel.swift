import Foundation

@MainActor
final class PersonaListViewModel: ObservableObject {
    enum LoadState: Equatable {
        case idle
        case loading
        case loaded
        case failed(String)
    }

    @Published private(set) var personas: [Persona] = []
    @Published private(set) var state: LoadState = .idle
    @Published private(set) var selectedPersonaID: String?
    /// `nil` means the "Tümü" filter — every loaded persona is shown.
    @Published private(set) var selectedCategory: String?
    @Published private(set) var conversations: [ConversationSummary] = []
    @Published private(set) var conversationsState: LoadState = .idle
    @Published private(set) var chatEnergy: ChatEnergy?

    private let apiClient: APIClient
    private static let categoryOrder = [
        "daily_companion",
        "motivational_coach",
        "hobby_book_club",
    ]

    var filteredPersonas: [Persona] {
        guard let selectedCategory else { return personas }
        return personas.filter { $0.category == selectedCategory }
    }

    /// Distinct categories present in the loaded list, in a stable
    /// Keşfet-pill order, so we never show a filter that would empty
    /// the list.
    var availableCategories: [String] {
        let present = Set(personas.map(\.category))
        let known = Self.categoryOrder.filter(present.contains)
        let unknown = present.subtracting(Self.categoryOrder).sorted()
        return known + unknown
    }

    init(apiClient: APIClient = .shared) {
        self.apiClient = apiClient
    }

    /// Loads personas. Keşfet uses `GET /personas?recommend=true` so the
    /// backend can flag exactly one recommended match from the saved
    /// onboarding mood. Onboarding persona-pick still uses plain
    /// `GET /personas` plus a local category reorder — the profile
    /// (and therefore the mood) hasn't been POSTed yet, so recommend
    /// would have nothing to go on.
    func load(prioritizedCategory: String? = nil) async {
        state = .loading
        do {
            if let prioritizedCategory {
                var loaded: [Persona] = try await apiClient.send(APIRequest(path: "/personas"))
                loaded.sort { lhs, rhs in
                    let lhsMatches = lhs.category == prioritizedCategory
                    let rhsMatches = rhs.category == prioritizedCategory
                    if lhsMatches != rhsMatches { return lhsMatches }
                    return lhs.sortOrder < rhs.sortOrder
                }
                personas = loaded
            } else {
                personas = try await loadRecommended()
            }
            state = .loaded
        } catch {
            state = .failed((error as? LocalizedError)?.errorDescription ?? error.localizedDescription)
        }
    }

    private func loadRecommended() async throws -> [Persona] {
        do {
            var loaded: [Persona] = try await apiClient.send(APIRequest(path: "/personas?recommend=true"))
            loaded.sort { lhs, rhs in
                let lhsRec = lhs.recommended == true
                let rhsRec = rhs.recommended == true
                if lhsRec != rhsRec { return lhsRec }
                return lhs.sortOrder < rhs.sortOrder
            }
            return loaded
        } catch {
            return try await apiClient.send(APIRequest(path: "/personas"))
        }
    }

    func loadChatEnergy() async {
        do {
            chatEnergy = try await apiClient.send(APIRequest(path: "/users/chat-energy"))
        } catch {
            chatEnergy = nil
        }
    }

    func select(_ persona: Persona) {
        selectedPersonaID = persona.id
    }

    func selectCategory(_ category: String?) {
        selectedCategory = category
    }

    func loadConversations() async {
        conversationsState = .loading
        do {
            conversations = try await apiClient.send(APIRequest(path: "/conversations"))
            conversationsState = .loaded
        } catch {
            conversationsState = .failed((error as? LocalizedError)?.errorDescription ?? error.localizedDescription)
        }
    }

    func existingConversation(for persona: Persona) -> ConversationSummary? {
        conversations.first { $0.personaID == persona.id }
    }
}

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

    /// Loads personas. When `prioritizedCategory` is given (see
    /// OnboardingProfile.MoodPreference.personaCategory), personas in
    /// that category are moved to the front, each group keeping its own
    /// relative `sortOrder` — used by the onboarding persona-pick screen
    /// to surface e.g. the motivational-coach category first when the
    /// user said they're looking for motivation. `nil` (the default, and
    /// what every other caller uses) keeps the backend's own order.
    func load(prioritizedCategory: String? = nil) async {
        state = .loading
        do {
            var loaded: [Persona] = try await apiClient.send(APIRequest(path: "/personas"))
            if let prioritizedCategory {
                loaded.sort { lhs, rhs in
                    let lhsMatches = lhs.category == prioritizedCategory
                    let rhsMatches = rhs.category == prioritizedCategory
                    if lhsMatches != rhsMatches { return lhsMatches }
                    return lhs.sortOrder < rhs.sortOrder
                }
            }
            personas = loaded
            state = .loaded
        } catch {
            state = .failed((error as? LocalizedError)?.errorDescription ?? error.localizedDescription)
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

import Foundation

/// Backs the personality-dial sliders on Persona Detayı. Loads the
/// user's current effective values for one persona (their own
/// customization if any, else the persona's defaults), and saves
/// immediately whenever a slider settles — no separate "save" step the
/// user has to remember to hit.
@MainActor
final class PersonaTraitsViewModel: ObservableObject {
    @Published var traits = PersonaTraits(warmth: 50, humor: 50, wisdom: 50, directness: 50, energy: 50)
    @Published private(set) var isCustomized = false
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?

    let personaID: String

    private let apiClient: APIClient
    private var hasLoaded = false

    init(personaID: String, apiClient: APIClient = .shared) {
        self.personaID = personaID
        self.apiClient = apiClient
    }

    private var path: String { "/personas/\(personaID)/traits" }

    func loadIfNeeded() async {
        guard !hasLoaded else { return }
        hasLoaded = true
        isLoading = true
        defer { isLoading = false }

        do {
            let response: PersonaTraitsResponse = try await apiClient.send(APIRequest(path: path))
            traits = response.traits
            isCustomized = response.isCustomized
        } catch {
            hasLoaded = false
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }

    /// Persists the current slider positions. Called once a slider
    /// drag ends (see PersonaDetailView) — not on every intermediate
    /// tick while dragging.
    func save() async {
        do {
            let body = try JSONEncoder().encode(traits)
            let request = APIRequest(path: path, method: .put, body: body)
            let response: PersonaTraitsResponse = try await apiClient.send(request)
            traits = response.traits
            isCustomized = response.isCustomized
        } catch {
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }

    /// Clears the user's customization — the persona's own defaults
    /// apply again from their next message on.
    func reset() async {
        do {
            let response: PersonaTraitsResponse = try await apiClient.send(APIRequest(path: path, method: .delete))
            traits = response.traits
            isCustomized = response.isCustomized
        } catch {
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }
}

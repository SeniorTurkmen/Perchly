import Foundation

@MainActor
final class ProfileViewModel: ObservableObject {
    @Published var isShowingAccountLinking = false
}

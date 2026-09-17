/// The app's top-level session state, driving RootView's branching.
enum AuthState: Equatable {
    /// Bootstrapping: checking Keychain / creating the initial
    /// anonymous session. Shown only very briefly, once, at launch.
    case loading
    /// A usable session tied to a device id, no linked email yet. The
    /// main app is fully usable in this state — email-linking is
    /// optional, offered from the profile screen.
    case anonymous
    /// A session backed by a verified email.
    case authenticated
}

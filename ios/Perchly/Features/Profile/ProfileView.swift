import SwiftUI

struct ProfileView: View {
    @ObservedObject private var authManager = AuthManager.shared
    @StateObject private var viewModel = ProfileViewModel()
    #if DEBUG
    @State private var isShowingResetConfirmation = false
    @State private var isResetting = false
    #endif

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    accountSection
                    #if DEBUG
                    debugSection
                    #endif
                }
                .padding()
            }
            .navigationTitle("Profil")
            .sheet(isPresented: $viewModel.isShowingAccountLinking) {
                NavigationStack {
                    EmailEntryView()
                }
            }
            .onChange(of: authManager.state) {
                if authManager.state == .authenticated {
                    viewModel.isShowingAccountLinking = false
                }
            }
        }
    }

    @ViewBuilder
    private var accountSection: some View {
        if authManager.state == .authenticated, let email = authManager.email {
            GlassCard {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Hesabın")
                        .font(PerchlyTypography.title)
                    Text(email)
                        .font(PerchlyTypography.body)
                        .foregroundStyle(PerchlyPalette.textSecondary)
                        .accessibilityIdentifier("accountEmailLabel")
                }
            }
        } else {
            Button {
                viewModel.isShowingAccountLinking = true
            } label: {
                GlassCard(tint: PerchlyPalette.accent) {
                    VStack(alignment: .leading, spacing: 6) {
                        Text("Hesabını e-posta ile bağla")
                            .font(PerchlyTypography.title)
                        Text("Verilerini kaybetme, cihaz değiştirdiğinde de seninle kalsın.")
                            .font(PerchlyTypography.body)
                            .foregroundStyle(PerchlyPalette.textSecondary)
                    }
                }
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("linkAccountCard")
        }
    }

    #if DEBUG
    /// Debug-only: local development needs a way to fully reset the
    /// stored session (including the device id) without erasing the
    /// whole Simulator — see AuthManager.debugResetLocalSession. Never
    /// compiled into a release build, so real users never see this.
    private var debugSection: some View {
        VStack(spacing: 8) {
            GlassButton(title: "Yerel Verileri Sıfırla (Debug)", isLoading: isResetting) {
                isShowingResetConfirmation = true
            }
            .accessibilityIdentifier("debugResetButton")

            Text("Sadece geliştirme içindir. Cihazdaki oturumu, bağlı e-postayı ve cihaz kimliğini siler.")
                .font(PerchlyTypography.caption)
                .foregroundStyle(PerchlyPalette.textSecondary)
                .multilineTextAlignment(.center)
        }
        .alert("Yerel verileri sıfırla?", isPresented: $isShowingResetConfirmation) {
            Button("Sıfırla", role: .destructive) {
                Task {
                    isResetting = true
                    await authManager.debugResetLocalSession()
                    isResetting = false
                }
            }
            Button("Vazgeç", role: .cancel) {}
        } message: {
            Text("Bu cihazdaki oturum, bağlı e-posta ve cihaz kimliği silinir; sıfırdan anonim bir oturum açılır. Sadece geliştirme için kullan.")
        }
    }
    #endif
}

#Preview {
    ProfileView()
}

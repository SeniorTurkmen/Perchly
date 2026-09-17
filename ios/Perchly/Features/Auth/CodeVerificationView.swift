import SwiftUI

struct CodeVerificationView: View {
    @StateObject private var viewModel: CodeVerificationViewModel
    @State private var showRegisterComplete = false

    init(email: String) {
        _viewModel = StateObject(wrappedValue: CodeVerificationViewModel(email: email))
    }

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                Text("Kodu doğrula")
                    .font(PerchlyTypography.largeTitle)
                Text("\(viewModel.email) adresine gönderdiğimiz 6 haneli kodu gir.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            CodeEntryField(code: $viewModel.code, onComplete: viewModel.verifyIfComplete)
                .disabled(viewModel.phase == .verifying || viewModel.phase == .expired)

            if viewModel.phase == .verifying {
                ProgressView()
            }

            if let errorMessage = viewModel.errorMessage {
                Text(errorMessage)
                    .font(PerchlyTypography.caption)
                    .foregroundStyle(PerchlyPalette.textSecondary)
                    .multilineTextAlignment(.center)
            }

            if viewModel.phase == .expired {
                GlassButton(title: "Yeni Kod İste", isLoading: viewModel.isRequestingNewCode) {
                    Task { await viewModel.requestNewCode() }
                }
            } else {
                VStack(spacing: 6) {
                    CountdownBar(remainingFraction: viewModel.remainingFraction)
                    Text("Kod \(viewModel.formattedRemaining) içinde geçersiz olacak")
                        .font(PerchlyTypography.caption)
                        .foregroundStyle(PerchlyPalette.textSecondary)
                }
            }

            Spacer()
        }
        .padding()
        .navigationTitle("Kod Doğrulama")
        .navigationBarTitleDisplayMode(.inline)
        .navigationDestination(isPresented: $showRegisterComplete) {
            RegisterCompleteView()
        }
        .onChange(of: viewModel.didComplete) {
            guard viewModel.didComplete else { return }
            if viewModel.isNewRegistration {
                showRegisterComplete = true
            }
            // Durum A/B without a new registration: AuthManager.state
            // already flipped to .authenticated inside verifyEmailCode.
            // The sheet presenting this whole flow (see ProfileView)
            // dismisses itself in response — nothing to do here.
        }
    }
}

#Preview {
    NavigationStack {
        CodeVerificationView(email: "ada@example.com")
    }
}

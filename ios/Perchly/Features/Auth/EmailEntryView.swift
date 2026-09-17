import SwiftUI

struct EmailEntryView: View {
    @StateObject private var viewModel = EmailEntryViewModel()
    @State private var verifyingEmail: String?

    var body: some View {
        VStack(spacing: 20) {
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                Text("Hesabını bağla")
                    .font(PerchlyTypography.largeTitle)
                Text("E-postanı doğrula, cihaz değiştirsen de verilerin seninle kalsın.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            GlassTextField(
                placeholder: "E-posta adresin",
                text: $viewModel.email,
                keyboardType: .emailAddress,
                textContentType: .emailAddress,
                submitLabel: .send,
                onSubmit: { Task { await send() } }
            )
            .accessibilityIdentifier("emailField")

            if let errorMessage = viewModel.errorMessage {
                Text(errorMessage)
                    .font(PerchlyTypography.caption)
                    .foregroundStyle(PerchlyPalette.textSecondary)
                    .multilineTextAlignment(.center)
            }

            GlassButton(title: "Kod Gönder", isDisabled: !viewModel.canSend, isLoading: viewModel.isSending) {
                Task { await send() }
            }
            .accessibilityIdentifier("sendCodeButton")

            Spacer()
        }
        .padding()
        .navigationTitle("E-posta ile Bağlan")
        .navigationBarTitleDisplayMode(.inline)
        .navigationDestination(item: $verifyingEmail) { email in
            CodeVerificationView(email: email)
        }
    }

    private func send() async {
        if await viewModel.sendCode() {
            verifyingEmail = viewModel.normalizedEmail
        }
    }
}

#Preview {
    NavigationStack {
        EmailEntryView()
    }
}

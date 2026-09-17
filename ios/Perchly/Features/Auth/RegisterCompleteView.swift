import SwiftUI

/// Shown only when email verification reported `is_new_registration`.
/// Success flips `AuthManager.state` to `.authenticated`, which the
/// sheet presenting this whole flow (see ProfileView) reacts to by
/// dismissing — there's no onboarding flow to hand off to yet, so this
/// just returns to the main app.
struct RegisterCompleteView: View {
    @StateObject private var viewModel = RegisterCompleteViewModel()

    var body: some View {
        VStack(spacing: 20) {
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                Text("Seni nasıl çağıralım?")
                    .font(PerchlyTypography.largeTitle)
                Text("Bu isim sohbetlerinde ve profilinde görünecek.")
                    .font(PerchlyTypography.body)
                    .foregroundStyle(PerchlyPalette.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            GlassTextField(
                placeholder: "Adın",
                text: $viewModel.displayName,
                autocapitalization: .words,
                autocorrectionDisabled: false,
                submitLabel: .done,
                onSubmit: { Task { await viewModel.submit() } }
            )
            .accessibilityIdentifier("displayNameField")

            if let errorMessage = viewModel.errorMessage {
                Text(errorMessage)
                    .font(PerchlyTypography.caption)
                    .foregroundStyle(PerchlyPalette.textSecondary)
                    .multilineTextAlignment(.center)
            }

            GlassButton(title: "Devam Et", isDisabled: !viewModel.canSubmit, isLoading: viewModel.isSubmitting) {
                Task { await viewModel.submit() }
            }
            .accessibilityIdentifier("completeRegistrationButton")

            Spacer()
        }
        .padding()
        .navigationTitle("Kayıt Tamamla")
        .navigationBarTitleDisplayMode(.inline)
        .navigationBarBackButtonHidden(true)
    }
}

#Preview {
    NavigationStack {
        RegisterCompleteView()
    }
}

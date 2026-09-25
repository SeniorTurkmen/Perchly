import SwiftUI

/// A single-line Liquid Glass text field, for auth screens (email,
/// display name) and anywhere else a plain glass input is needed.
struct GlassTextField: View {
    // LocalizedStringKey, not String: TextField(String, text:) is
    // verbatim and skips the string catalog.
    let placeholder: LocalizedStringKey
    @Binding var text: String
    var keyboardType: UIKeyboardType = .default
    var textContentType: UITextContentType?
    var autocapitalization: TextInputAutocapitalization = .never
    var autocorrectionDisabled: Bool = true
    var submitLabel: SubmitLabel = .done
    var onSubmit: () -> Void = {}

    var body: some View {
        TextField(placeholder, text: $text)
            .font(PerchlyTypography.body)
            .keyboardType(keyboardType)
            .textContentType(textContentType)
            .textInputAutocapitalization(autocapitalization)
            .autocorrectionDisabled(autocorrectionDisabled)
            .submitLabel(submitLabel)
            .onSubmit(onSubmit)
            .padding(.horizontal, 20)
            .padding(.vertical, 16)
            .glassEffect(.regular, in: .rect(cornerRadius: 18))
    }
}

#Preview {
    VStack(spacing: 16) {
        GlassTextField(
            placeholder: "E-posta adresin",
            text: .constant(""),
            keyboardType: .emailAddress,
            textContentType: .emailAddress
        )
        GlassTextField(
            placeholder: "Adın",
            text: .constant("Ada"),
            autocapitalization: .words,
            autocorrectionDisabled: false
        )
    }
    .padding()
}

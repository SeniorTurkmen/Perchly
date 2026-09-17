import SwiftUI

/// A 6-digit code entry field, shown as separate Liquid Glass boxes with
/// an auto-advancing "current box" highlight, backed by one hidden
/// `TextField` rather than 6 chained ones — this is what gives it
/// reliable paste-a-whole-code and SMS-autofill (`.oneTimeCode`) support
/// without hand-rolling per-box focus transfer.
struct CodeEntryField: View {
    @Binding var code: String
    var length: Int = 6
    var accessibilityIdentifier: String = "codeEntryField"
    var onComplete: (String) -> Void = { _ in }

    @FocusState private var isFocused: Bool

    var body: some View {
        ZStack {
            TextField("", text: $code)
                .keyboardType(.numberPad)
                .textContentType(.oneTimeCode)
                .focused($isFocused)
                .accessibilityIdentifier(accessibilityIdentifier)
                .opacity(0.01)
                .onChange(of: code) { _, newValue in
                    let filtered = String(newValue.filter(\.isNumber).prefix(length))
                    if filtered != newValue {
                        code = filtered
                    }
                    if filtered.count == length {
                        onComplete(filtered)
                    }
                }

            HStack(spacing: 10) {
                ForEach(0..<length, id: \.self) { index in
                    digitBox(at: index)
                }
            }
            .allowsHitTesting(false)
        }
        .contentShape(Rectangle())
        .onTapGesture { isFocused = true }
        .task { isFocused = true }
    }

    private func digitBox(at index: Int) -> some View {
        let characters = Array(code)
        let character = index < characters.count ? String(characters[index]) : ""
        let isCurrent = isFocused && index == characters.count

        return Text(character)
            .font(PerchlyTypography.title)
            .frame(width: 44, height: 56)
            .glassEffect(glass(isCurrent: isCurrent), in: .rect(cornerRadius: 14))
    }

    private func glass(isCurrent: Bool) -> Glass {
        isCurrent ? .regular.tint(PerchlyPalette.accent.opacity(0.25)) : .regular
    }
}

#Preview {
    CodeEntryField(code: .constant("482"))
        .padding()
}

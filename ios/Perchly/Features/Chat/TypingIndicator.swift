import SwiftUI

/// Three dots that pulse in sequence, shown while an assistant bubble is
/// waiting for its first chunk of text.
struct TypingIndicator: View {
    let tint: Color

    @State private var activeIndex = 0

    private let timer = Timer.publish(every: 0.35, on: .main, in: .common).autoconnect()

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<3, id: \.self) { index in
                Circle()
                    .fill(tint.opacity(index == activeIndex ? 1 : 0.3))
                    .frame(width: 6, height: 6)
            }
        }
        .onReceive(timer) { _ in
            activeIndex = (activeIndex + 1) % 3
        }
    }
}

#Preview {
    TypingIndicator(tint: .accentColor)
        .padding()
}

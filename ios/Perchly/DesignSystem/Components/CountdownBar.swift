import SwiftUI

/// A calm, thin Liquid Glass progress indicator — the glass capsule is
/// the track, a plain accent-colored capsule inset inside it shrinks as
/// `remainingFraction` drops. Used for the verification code's 10-minute
/// validity window.
struct CountdownBar: View {
    /// 0...1, where 1 is "just started" and 0 is "expired".
    let remainingFraction: Double

    var body: some View {
        GeometryReader { proxy in
            ZStack(alignment: .leading) {
                Capsule()
                    .glassEffect(.regular, in: .capsule)
                Capsule()
                    .fill(PerchlyPalette.accent)
                    .frame(width: max(4, proxy.size.width * remainingFraction.clamped(to: 0...1)))
                    .padding(2)
            }
        }
        .frame(height: 8)
        .animation(.linear(duration: 1), value: remainingFraction)
    }
}

private extension Double {
    func clamped(to range: ClosedRange<Double>) -> Double {
        min(max(self, range.lowerBound), range.upperBound)
    }
}

#Preview {
    VStack(spacing: 20) {
        CountdownBar(remainingFraction: 1.0)
        CountdownBar(remainingFraction: 0.5)
        CountdownBar(remainingFraction: 0.05)
    }
    .padding()
}

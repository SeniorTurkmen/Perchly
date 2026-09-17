import SwiftUI

extension Color {
    /// Creates a color from a "#RRGGBB" or "#RRGGBBAA" hex string, such as
    /// a persona's `accent_color` from the backend. Falls back to
    /// `PerchlyPalette.accent` if the string can't be parsed.
    init(hex: String) {
        let hexString = hex.trimmingCharacters(in: CharacterSet(charactersIn: "#"))

        var value: UInt64 = 0
        guard Scanner(string: hexString).scanHexInt64(&value),
              hexString.count == 6 || hexString.count == 8 else {
            self = PerchlyPalette.accent
            return
        }

        let r, g, b, a: Double
        if hexString.count == 8 {
            r = Double((value & 0xFF00_0000) >> 24) / 255
            g = Double((value & 0x00FF_0000) >> 16) / 255
            b = Double((value & 0x0000_FF00) >> 8) / 255
            a = Double(value & 0x0000_00FF) / 255
        } else {
            r = Double((value & 0xFF0000) >> 16) / 255
            g = Double((value & 0x00FF00) >> 8) / 255
            b = Double(value & 0x0000FF) / 255
            a = 1
        }

        self.init(.sRGB, red: r, green: g, blue: b, opacity: a)
    }
}

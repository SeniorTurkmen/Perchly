import SwiftUI

/// Central color palette for the app. Prefer these tokens over raw
/// `Color` literals so the design language can evolve from one place.
enum PerchlyPalette {
    static let background = Color(.systemBackground)
    static let surface = Color.primary.opacity(0.06)
    static let accent = Discover.primary
    static let textPrimary = Color.primary
    static let textSecondary = Color.secondary

    /// Tokens from the Keşfet Personalar Stitch screen. Used by the
    /// discover list; other features keep the system-adaptive colors
    /// above so dark mode elsewhere is not forced into this light theme.
    enum Discover {
        static let background = Color(hex: "#fbf8ff")
        static let onSurface = Color(hex: "#1a1b22")
        static let onSurfaceVariant = Color(hex: "#424751")
        static let primary = Color(hex: "#115eaf")
        static let onPrimary = Color.white
        static let primaryContainer = Color(hex: "#6ea8fe")
        static let onPrimaryContainer = Color(hex: "#003c76")
        static let primaryFixed = Color(hex: "#d5e3ff")
        static let primaryFixedDim = Color(hex: "#a8c8ff")
        static let onPrimaryFixed = Color(hex: "#001b3c")
        static let secondary = Color(hex: "#99415a")
        static let onSecondary = Color.white
        static let secondaryContainer = Color(hex: "#ff93ad")
        static let secondaryFixed = Color(hex: "#ffd9e0")
        static let secondaryFixedDim = Color(hex: "#ffb1c2")
        static let onSecondaryFixed = Color(hex: "#3f0018")
        static let tertiary = Color(hex: "#855316")
        static let onTertiary = Color.white
        static let tertiaryContainer = Color(hex: "#d89a57")
        static let tertiaryFixed = Color(hex: "#ffdcbd")
        static let onTertiaryFixed = Color(hex: "#2c1600")
        static let surface = Color(hex: "#fbf8ff")
        static let surfaceLowest = Color.white
        static let surfaceLow = Color(hex: "#f4f2fd")
        static let surfaceContainer = Color(hex: "#eeedf7")
        static let surfaceContainerHigh = Color(hex: "#e8e7f1")
        static let surfaceDim = Color(hex: "#dad9e3")
        static let surfaceVariant = Color(hex: "#e3e1ec")
    }
}

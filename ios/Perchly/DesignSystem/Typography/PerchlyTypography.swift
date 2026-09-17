import SwiftUI

/// Central type scale for the app. Prefer these tokens over raw
/// `.font(...)` calls so text styling stays consistent across features.
enum PerchlyTypography {
    static let largeTitle = Font.system(.largeTitle, weight: .bold)
    static let title = Font.system(.title2, weight: .semibold)
    static let body = Font.system(.body)
    static let caption = Font.system(.caption)

    /// Discover (Keşfet) type scale — Plus Jakarta–like rounded SF.
    enum Discover {
        static let headlineLG = Font.system(size: 22, weight: .semibold, design: .rounded)
        static let headlineSM = Font.system(size: 18, weight: .semibold, design: .rounded)
        static let bodyMD = Font.system(size: 15, weight: .regular, design: .rounded)
        static let bodySM = Font.system(size: 13, weight: .regular, design: .rounded)
        static let labelLG = Font.system(size: 14, weight: .semibold, design: .rounded)
        static let labelMD = Font.system(size: 12, weight: .medium, design: .rounded)
        static let labelSM = Font.system(size: 11, weight: .medium, design: .rounded)
    }
}

import SwiftUI

/// Stitch "Onboarding - İsim & Hitap Tercihi". The name is POSTed as
/// `preferred_name` with the rest of the onboarding profile. There is
/// no generated nickname: either the user types how they want to be
/// addressed, or they continue anonymously and personas use no name.
struct NameHitapStepView: View {
    @ObservedObject var coordinator: OnboardingCoordinator
    @State private var name = ""

    private var trimmedName: String {
        name.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private var isNameValid: Bool { PreferredNameRules.isValid(trimmedName) }

    var body: some View {
        OnboardingScaffold(stepIndex: 1, stepLabel: "Hitap") {
            VStack(alignment: .leading, spacing: 20) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Sana nasıl seslenelim?")
                        .font(PerchlyTypography.Discover.headlineLG)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Text("Perchmate’inin sana seçtiğin isimle hitap etmesini isteyebilirsin. Anonim devam edersen sohbette hiçbir isimle seslenilmez.")
                        .font(PerchlyTypography.Discover.bodyMD)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }

                namedCard

                privacyCard
            }
        } footer: {
            VStack(spacing: 10) {
                OnboardingPrimaryButton(
                    title: trimmedName.isEmpty ? "Devam Et" : "\(trimmedName) Olarak Devam Et",
                    isDisabled: !isNameValid
                ) {
                    coordinator.selectPreferredName(trimmedName)
                }
                .accessibilityIdentifier("nameContinueButton")

                Button {
                    coordinator.continueAnonymously()
                } label: {
                    HStack(spacing: 6) {
                        Image(systemName: "eye.slash")
                        Text("Anonim Olarak Devam Et")
                    }
                    .font(PerchlyTypography.Discover.labelLG)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("skipNameAnonymousButton")
            }
        }
    }

    private var namedCard: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .top, spacing: 12) {
                ZStack {
                    Circle().fill(PerchlyPalette.Discover.primaryFixed)
                    Image(systemName: "face.smiling.fill")
                        .foregroundStyle(PerchlyPalette.Discover.primary)
                }
                .frame(width: 40, height: 40)

                VStack(alignment: .leading, spacing: 4) {
                    Text("İsminle Hitap Edilsin")
                        .font(PerchlyTypography.Discover.headlineSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Text("Daha samimi bir bağ için dilediğin hitabı yaz. Takma ad üretilmez; yalnızca senin yazdığın kullanılır.")
                        .font(PerchlyTypography.Discover.bodySM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                }
            }

            HStack(spacing: 8) {
                Image(systemName: "face.smiling")
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                TextField("Örn: Deniz, Ece, Can...", text: $name)
                    .font(PerchlyTypography.Discover.bodyMD)
                    .textInputAutocapitalization(.words)
                    .autocorrectionDisabled()
                    .accessibilityIdentifier("displayNameField")
                if !name.isEmpty {
                    Button {
                        name = ""
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("Temizle")
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(PerchlyPalette.Discover.surfaceLow, in: Capsule())

            if trimmedName.count > PreferredNameRules.maxLength {
                Text("Hitap en fazla \(PreferredNameRules.maxLength) karakter olabilir.")
                    .font(PerchlyTypography.Discover.bodySM)
                    .foregroundStyle(PerchlyPalette.Discover.secondary)
            }
        }
        .padding(20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.7)), in: .rect(cornerRadius: 20))
    }

    private var privacyCard: some View {
        HStack(alignment: .top, spacing: 12) {
            ZStack {
                Circle().fill(PerchlyPalette.Discover.tertiaryFixed)
                Image(systemName: "lock.open.fill")
                    .font(.system(size: 13))
                    .foregroundStyle(PerchlyPalette.Discover.tertiary)
            }
            .frame(width: 32, height: 32)

            VStack(alignment: .leading, spacing: 4) {
                Text("Gizlilik Sözü")
                    .font(PerchlyTypography.Discover.labelLG)
                    .foregroundStyle(PerchlyPalette.Discover.onSurface)
                Text("Yazdığın hitap yalnızca sohbette sana seslenmek için kullanılır; satılmaz ve üçüncü partilerle paylaşılmaz. Anonim devam edersen hiçbir isim kaydedilmez.")
                    .font(PerchlyTypography.Discover.bodySM)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
            }
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PerchlyPalette.Discover.surfaceLow.opacity(0.75), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
    }
}

#Preview {
    NameHitapStepView(coordinator: OnboardingCoordinator())
}

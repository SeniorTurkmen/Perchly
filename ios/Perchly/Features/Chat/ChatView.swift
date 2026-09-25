import SwiftUI

struct ChatView: View {
    @StateObject private var viewModel: ChatViewModel

    init(persona: Persona, conversationID: String? = nil) {
        _viewModel = StateObject(wrappedValue: ChatViewModel(persona: persona, conversationID: conversationID))
    }

    var body: some View {
        VStack(spacing: 0) {
            presenceCard
                .padding(.horizontal, 20)
                .padding(.top, 8)
                .padding(.bottom, 8)

            messageList
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            composerDock
        }
        .background {
            ChatAmbientBackground(accent: viewModel.persona.accent)
        }
        .background(PerchlyPalette.Discover.background)
        .navigationTitle("Persona Sohbet Seansı")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar(.visible, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .principal) {
                VStack(spacing: 2) {
                    Text("Persona Sohbet Seansı")
                        .font(PerchlyTypography.Discover.headlineSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    HStack(spacing: 4) {
                        Circle()
                            .fill(viewModel.sessionStartedAt == nil ? PerchlyPalette.Discover.onSurfaceVariant : viewModel.persona.accent)
                            .frame(width: 6, height: 6)
                        Text(sessionSubtitle)
                            .font(PerchlyTypography.Discover.labelSM)
                            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    }
                }
            }
        }
        .preferredColorScheme(.light)
        .task { await viewModel.startConversationIfNeeded() }
        .alert("Bir şeyler ters gitti", isPresented: isShowingError) {
            Button("Tamam", role: .cancel) {}
        } message: {
            Text(viewModel.errorMessage ?? "")
        }
    }

    private var sessionSubtitle: String {
        if viewModel.sessionStartedAt == nil {
            return "\(viewModel.persona.name) - Bağlanıyor"
        }
        return "\(viewModel.persona.name) - Aktif Seans"
    }

    private var presenceCard: some View {
        let persona = viewModel.persona
        let style = persona.categoryStyle
        return HStack(spacing: 12) {
            ZStack(alignment: .bottomTrailing) {
                ChatPersonaAvatar(persona: persona, size: 48)
                Circle()
                    .fill(persona.accent.opacity(0.35))
                    .frame(width: 14, height: 14)
                    .overlay {
                        Circle()
                            .fill(persona.accent)
                            .frame(width: 8, height: 8)
                    }
            }

            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 6) {
                    Text(persona.name)
                        .font(PerchlyTypography.Discover.headlineSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                    Text("(\(style.roleTitle))")
                        .font(PerchlyTypography.Discover.labelSM)
                        .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                        .lineLimit(1)
                }
                HStack(spacing: 4) {
                    Image(systemName: style.skillSymbol)
                        .font(.system(size: 10))
                    Text(style.badgeTitle)
                }
                .font(PerchlyTypography.Discover.labelSM)
                .foregroundStyle(style.badgeText)
                .padding(.horizontal, 8)
                .padding(.vertical, 3)
                .background(style.badgeFill, in: Capsule())

                Text(persona.toneDescription)
                    .font(PerchlyTypography.Discover.labelSM)
                    .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                    .lineLimit(2)
                    .fixedSize(horizontal: false, vertical: true)
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.5)), in: .rect(cornerRadius: 20))
    }

    private var messageList: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(spacing: 16) {
                    if viewModel.messages.isEmpty {
                        emptySessionPrompt
                    } else {
                        if let startedAt = viewModel.sessionStartedAt ?? viewModel.messages.first?.createdAt {
                            dayDivider(for: startedAt)
                        }

                        ForEach(viewModel.messages) { message in
                            ChatBubble(message: message, persona: viewModel.persona) { emoji in
                                Task { await viewModel.setReaction(on: message, emoji: emoji) }
                            }
                            .padding(.bottom, message.id == viewModel.messages.last?.id ? Self.composerFadeClearance : 0)
                            .id(message.id)
                        }
                    }
                }
                .padding(.horizontal, 20)
                .padding(.top, 4)
                .padding(.bottom, viewModel.messages.isEmpty ? Self.composerFadeClearance : 0)
            }
            .scrollIndicators(.hidden)
            .onAppear { scrollToBottom(proxy) }
            .onChange(of: viewModel.messages.last?.content) {
                scrollToBottom(proxy)
            }
            .onChange(of: viewModel.messages.count) {
                scrollToBottom(proxy)
            }
            .onChange(of: viewModel.isSending) {
                scrollToBottom(proxy)
            }
        }
    }

    private var emptySessionPrompt: some View {
        VStack(spacing: 8) {
            Text("“\(viewModel.persona.shortDescription)”")
                .font(PerchlyTypography.Discover.bodyMD)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .multilineTextAlignment(.center)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(.top, 24)
        .padding(.horizontal, 12)
    }

    private func dayDivider(for date: Date) -> some View {
        Text(Self.sessionStamp(for: date))
            .font(PerchlyTypography.Discover.labelSM)
            .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
            .padding(.horizontal, 14)
            .padding(.vertical, 4)
            .background(PerchlyPalette.Discover.surfaceContainerHigh.opacity(0.6), in: Capsule())
            .padding(.vertical, 4)
    }

    private var composerDock: some View {
        VStack(spacing: 8) {
            if viewModel.isSending {
                listeningCue
            } else if !hasStartedTalking {
                quickReplyRow
            }

            ChatInputBar(text: $viewModel.draft, isSending: viewModel.isSending, onSend: viewModel.sendDraft)
        }
        .padding(.horizontal, 20)
        .padding(.top, 10)
        .padding(.bottom, 10)
        .background {
            VStack(spacing: 0) {
                LinearGradient(
                    colors: [
                        PerchlyPalette.Discover.background.opacity(0),
                        PerchlyPalette.Discover.background.opacity(0.72),
                        PerchlyPalette.Discover.background,
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )
                .frame(height: Self.composerFadeHeight)
                Rectangle().fill(PerchlyPalette.Discover.background)
            }
            .padding(.top, -Self.composerFadeHeight)
            .ignoresSafeArea(edges: .bottom)
            .allowsHitTesting(false)
        }
    }

    /// Starter chips belong to an empty session. Once the user has sent
    /// something they get in the way of the last bubble.
    private var hasStartedTalking: Bool {
        viewModel.messages.contains { $0.role == .user }
    }

    private var listeningCue: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(viewModel.persona.accent)
                .frame(width: 6, height: 6)
            Text(viewModel.persona.listeningCue)
                .font(PerchlyTypography.Discover.labelSM)
                .foregroundStyle(PerchlyPalette.Discover.onSurfaceVariant)
                .lineLimit(1)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 5)
        .background(viewModel.persona.accent.opacity(0.14), in: Capsule())
    }

    private var quickReplyRow: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(viewModel.persona.quickReplies, id: \.title) { reply in
                    Button {
                        viewModel.applyQuickReply(reply.title)
                    } label: {
                        HStack(spacing: 5) {
                            Text(reply.title)
                            Image(systemName: reply.symbol)
                                .font(.system(size: 11))
                                .opacity(0.55)
                        }
                        .font(PerchlyTypography.Discover.labelMD)
                        .foregroundStyle(PerchlyPalette.Discover.onSurface)
                        .padding(.horizontal, 12)
                        .frame(height: 32)
                        .glassEffect(.regular.tint(PerchlyPalette.Discover.surfaceLowest.opacity(0.8)), in: .capsule)
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.vertical, 1)
        }
    }

    private func scrollToBottom(_ proxy: ScrollViewProxy) {
        guard let lastID = viewModel.messages.last?.id else { return }
        withAnimation(.easeOut(duration: 0.2)) {
            proxy.scrollTo(lastID, anchor: .bottom)
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { viewModel.errorMessage != nil },
            set: { isPresented in if !isPresented { viewModel.errorMessage = nil } }
        )
    }

    private static let composerFadeHeight: CGFloat = 28
    /// Extra space below the last bubble so it sits above the dock fade
    /// when scrolled to the bottom — `scrollTo(..., .bottom)` includes
    /// this padding as part of the last row.
    private static let composerFadeClearance: CGFloat = 44

    private static func sessionStamp(for date: Date) -> String {
        let time = ChatBubble.timeString(from: date)
        if Calendar.current.isDateInToday(date) {
            return String(localized: "Bugün \(time)")
        }
        // Was hardcoded to Locale(identifier: "tr_TR") — forced Turkish
        // day/month word order (and month names) regardless of the
        // device's actual language.
        let day = date.formatted(.dateTime.day().month(.wide).locale(.autoupdatingCurrent))
        return "\(day) \(time)"
    }
}

private struct ChatAmbientBackground: View {
    let accent: Color

    var body: some View {
        ZStack {
            PerchlyPalette.Discover.background
            Circle()
                .fill(accent.opacity(0.22))
                .frame(width: 320, height: 320)
                .blur(radius: 60)
                .offset(x: 140, y: -180)
            Circle()
                .fill(PerchlyPalette.Discover.primaryContainer.opacity(0.2))
                .frame(width: 280, height: 280)
                .blur(radius: 60)
                .offset(x: -140, y: 80)
            Circle()
                .fill(accent.opacity(0.16))
                .frame(width: 240, height: 240)
                .blur(radius: 50)
                .offset(x: -80, y: 280)
        }
        .ignoresSafeArea()
        .allowsHitTesting(false)
    }
}

#Preview {
    NavigationStack {
        ChatView(persona: .preview)
    }
}

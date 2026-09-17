import Foundation

/// Mirrors the backend's `internal/apierror.Code` — the stable,
/// machine-readable identifier every failed response now carries
/// alongside its (Turkish, display-only) `error` message.
///
/// `unknown` absorbs any raw value this enum doesn't list yet, so an
/// older build of the app never fails to decode an error body just
/// because the backend shipped a new code — see `init(from:)`.
enum APIErrorCode: String, Decodable {
    // Generic.
    case invalidRequestBody = "invalid_request_body"
    case unauthorized = "unauthorized"
    case internalError = "internal_error"

    // Auth.
    case sessionCreateFailed = "session_create_failed"
    case invalidEmail = "invalid_email"
    case emailCodeSendFailed = "email_code_send_failed"
    case tooManyAttempts = "too_many_attempts"
    case verificationCodeExpired = "verification_code_expired"
    case invalidVerificationCode = "invalid_verification_code"
    case verificationFailed = "verification_failed"
    case invalidDisplayName = "invalid_display_name"
    case userNotFound = "user_not_found"
    case profileUpdateFailed = "profile_update_failed"
    case invalidRefreshToken = "invalid_refresh_token"
    case sessionRefreshFailed = "session_refresh_failed"
    case logoutFailed = "logout_failed"

    // Personas.
    case invalidPersonaID = "invalid_persona_id"
    case personaNotFound = "persona_not_found"
    case personaFetchFailed = "persona_fetch_failed"
    case personasListFailed = "personas_list_failed"
    case personaNotAgeAppropriate = "persona_not_age_appropriate"
    case personaTraitOutOfRange = "persona_trait_out_of_range"
    case personaTraitsFailed = "persona_traits_failed"

    // Conversations & messages.
    case invalidConversationID = "invalid_conversation_id"
    case conversationNotFound = "conversation_not_found"
    case conversationForbidden = "conversation_forbidden"
    case conversationCreateFailed = "conversation_create_failed"
    case conversationsListFailed = "conversations_list_failed"
    case conversationFetchFailed = "conversation_fetch_failed"
    case conversationContextMissing = "conversation_context_missing"
    case messagesListFailed = "messages_list_failed"
    case invalidMessageID = "invalid_message_id"
    case messageNotFound = "message_not_found"
    case messageContentRequired = "message_content_required"
    case messageSendFailed = "message_send_failed"
    case streamingUnsupported = "streaming_unsupported"
    case reactionEmojiRequired = "reaction_emoji_required"
    case invalidReactionEmoji = "invalid_reaction_emoji"
    case cannotReactToOwnMessage = "cannot_react_to_own_message"
    case reactionFailed = "reaction_failed"

    // Onboarding.
    case invalidAgeRange = "invalid_age_range"
    case invalidMoodPreference = "invalid_mood_preference"
    case invalidSelectedPersonaID = "invalid_selected_persona_id"
    case preferredNameRequired = "preferred_name_required"
    case preferredNameInvalid = "preferred_name_invalid"
    case onboardingSaveFailed = "onboarding_save_failed"
    case onboardingFetchFailed = "onboarding_fetch_failed"
    case onboardingProfileNotFound = "onboarding_profile_not_found"

    // Quota / chat energy.
    case quotaCheckFailed = "quota_check_failed"
    case quotaExceeded = "quota_exceeded"
    case chatEnergyFailed = "chat_energy_failed"

    /// A code this build doesn't recognize yet — never thrown by the
    /// backend itself, only produced by decoding.
    case unknown

    init(from decoder: Decoder) throws {
        let raw = try decoder.singleValueContainer().decode(String.self)
        self = APIErrorCode(rawValue: raw) ?? .unknown
    }
}

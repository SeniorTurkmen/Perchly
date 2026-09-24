// Package apierror is the single, shared home for the API's error
// envelope — used by internal/handler and internal/auth alike, so
// every failed response across the whole server (not just handlers
// that remember to opt in) carries the same {error, code} shape.
package apierror

// Code is a stable, machine-readable identifier for a failure reason.
// Unlike the free-text message (which is for display, in Turkish, and
// may be reworded at any time), a client can safely switch on this to
// drive its own logic (e.g. show a specific screen for
// CodePersonaNotAgeAppropriate vs. a generic error banner for
// anything else).
type Code string

const (
	// Generic — apply across many endpoints.
	CodeInvalidRequestBody Code = "invalid_request_body"
	CodeUnauthorized       Code = "unauthorized"
	CodeInternal           Code = "internal_error"

	// Auth.
	CodeSessionCreateFailed  Code = "session_create_failed"
	CodeInvalidEmail         Code = "invalid_email"
	CodeEmailCodeSendFailed  Code = "email_code_send_failed"
	CodeTooManyAttempts      Code = "too_many_attempts"
	CodeVerificationExpired  Code = "verification_code_expired"
	CodeInvalidCode          Code = "invalid_verification_code"
	CodeVerificationFailed   Code = "verification_failed"
	CodeInvalidDisplayName   Code = "invalid_display_name"
	CodeUserNotFound         Code = "user_not_found"
	CodeProfileUpdateFailed  Code = "profile_update_failed"
	CodeInvalidRefreshToken  Code = "invalid_refresh_token"
	CodeSessionRefreshFailed Code = "session_refresh_failed"
	CodeLogoutFailed         Code = "logout_failed"

	// Personas.
	CodeInvalidPersonaID         Code = "invalid_persona_id"
	CodePersonaNotFound          Code = "persona_not_found"
	CodePersonaFetchFailed       Code = "persona_fetch_failed"
	CodePersonasListFailed       Code = "personas_list_failed"
	CodePersonaNotAgeAppropriate Code = "persona_not_age_appropriate"
	CodePersonaTraitOutOfRange   Code = "persona_trait_out_of_range"
	CodePersonaTraitsFailed      Code = "persona_traits_failed"

	// Conversations & messages.
	CodeInvalidConversationID      Code = "invalid_conversation_id"
	CodeConversationNotFound       Code = "conversation_not_found"
	CodeConversationForbidden      Code = "conversation_forbidden"
	CodeConversationCreateFailed   Code = "conversation_create_failed"
	CodeConversationsListFailed    Code = "conversations_list_failed"
	CodeConversationFetchFailed    Code = "conversation_fetch_failed"
	CodeConversationContextMissing Code = "conversation_context_missing"
	CodeMessagesListFailed         Code = "messages_list_failed"
	CodeInvalidMessageID           Code = "invalid_message_id"
	CodeMessageNotFound            Code = "message_not_found"
	CodeMessageContentRequired     Code = "message_content_required"
	CodeMessageSendFailed          Code = "message_send_failed"
	CodeStreamingUnsupported       Code = "streaming_unsupported"
	CodeReactionEmojiRequired      Code = "reaction_emoji_required"
	CodeInvalidReactionEmoji       Code = "invalid_reaction_emoji"
	CodeCannotReactToOwnMessage    Code = "cannot_react_to_own_message"
	CodeReactionFailed             Code = "reaction_failed"

	// Onboarding.
	CodeInvalidAgeRange           Code = "invalid_age_range"
	CodeInvalidMoodPreference     Code = "invalid_mood_preference"
	CodeInvalidSelectedPersonaID  Code = "invalid_selected_persona_id"
	CodePreferredNameRequired     Code = "preferred_name_required"
	CodePreferredNameInvalid      Code = "preferred_name_invalid"
	CodeOnboardingSaveFailed      Code = "onboarding_save_failed"
	CodeOnboardingFetchFailed     Code = "onboarding_fetch_failed"
	CodeOnboardingProfileNotFound Code = "onboarding_profile_not_found"

	// Quota / chat energy.
	CodeQuotaCheckFailed Code = "quota_check_failed"
	CodeQuotaExceeded    Code = "quota_exceeded"
	CodeChatEnergyFailed Code = "chat_energy_failed"

	// Admin dashboard auth — entirely separate from the app-user codes
	// above so a client can never confuse the two.
	CodeAdminInvalidCredentials Code = "admin_invalid_credentials"
	CodeAdminLoginFailed        Code = "admin_login_failed"
	CodeAdminUnauthorized       Code = "admin_unauthorized"
	CodeAdminLogoutFailed       Code = "admin_logout_failed"
	CodeAdminMeFailed           Code = "admin_me_failed"

	// Admin dashboard — users.
	CodeInvalidUserID            Code = "invalid_user_id"
	CodeAdminUsersListFailed     Code = "admin_users_list_failed"
	CodeAdminUserFetchFailed     Code = "admin_user_fetch_failed"
	CodeInvalidDailyLimit        Code = "invalid_daily_limit"
	CodeAdminQuotaUpdateFailed   Code = "admin_quota_update_failed"
	CodeInvalidCreditAmount      Code = "invalid_credit_amount"
	CodeAdminCreditsUpdateFailed Code = "admin_credits_update_failed"

	// Admin dashboard — personas.
	CodeInvalidPersonaInput      Code = "invalid_persona_input"
	CodeAdminPersonasListFailed  Code = "admin_personas_list_failed"
	CodeAdminPersonaCreateFailed Code = "admin_persona_create_failed"
	CodeAdminPersonaUpdateFailed Code = "admin_persona_update_failed"
	CodeInvalidPersonaLLMModelID Code = "invalid_persona_llm_model_id"
	CodePersonaLLMModelNotFound  Code = "persona_llm_model_not_found"
	CodePersonaLLMModelInactive  Code = "persona_llm_model_inactive"

	// Admin dashboard — home metrics.
	CodeAdminMetricsFailed Code = "admin_metrics_failed"

	// Admin dashboard — conversations & moderation.
	CodeAdminConversationsListFailed Code = "admin_conversations_list_failed"
	CodeAdminConversationFetchFailed Code = "admin_conversation_fetch_failed"
	CodeAdminMessageDeleteFailed     Code = "admin_message_delete_failed"

	// Admin dashboard — request logs.
	CodeAdminLogsListFailed Code = "admin_logs_list_failed"

	// Admin dashboard — own activity log.
	CodeAdminActivityListFailed Code = "admin_activity_list_failed"

	// Admin dashboard — onboarding insights.
	CodeAdminOnboardingInsightsFailed Code = "admin_onboarding_insights_failed"

	// Admin dashboard — LLM provider credentials & models.
	CodeInvalidLLMCredentialInput      Code = "invalid_llm_credential_input"
	CodeLLMCredentialNotFound          Code = "llm_credential_not_found"
	CodeLLMCredentialInUse             Code = "llm_credential_in_use"
	CodeAdminLLMCredentialsListFailed  Code = "admin_llm_credentials_list_failed"
	CodeAdminLLMCredentialFetchFailed  Code = "admin_llm_credential_fetch_failed"
	CodeAdminLLMCredentialCreateFailed Code = "admin_llm_credential_create_failed"
	CodeAdminLLMCredentialUpdateFailed Code = "admin_llm_credential_update_failed"
	CodeAdminLLMCredentialDeleteFailed Code = "admin_llm_credential_delete_failed"
	CodeInvalidLLMModelInput           Code = "invalid_llm_model_input"
	CodeLLMModelNotFound               Code = "llm_model_not_found"
	CodeLLMCredentialInactive          Code = "llm_credential_inactive"
	CodeAdminLLMModelsListFailed       Code = "admin_llm_models_list_failed"
	CodeAdminLLMModelCreateFailed      Code = "admin_llm_model_create_failed"
	CodeAdminLLMModelUpdateFailed      Code = "admin_llm_model_update_failed"
	CodeAdminLLMModelDeleteFailed      Code = "admin_llm_model_delete_failed"
)

// Response is the standard JSON error envelope every failed endpoint
// returns, from any layer (auth middleware, quota middleware, or a
// handler). `error` is the plain display string clients have always
// gotten; `code` is additive — safe for older clients to ignore, and
// the only field a newer client needs to switch on programmatically.
type Response struct {
	Error string `json:"error"`
	Code  Code   `json:"code"`
}

func New(code Code, message string) Response {
	return Response{Error: message, Code: code}
}

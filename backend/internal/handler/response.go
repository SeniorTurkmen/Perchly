package handler

import (
	"encoding/json"
	"net/http"

	"perchly-backend/internal/apierror"
)

// ErrorCode and errorResponse are local aliases for the shared
// apierror types — see that package's doc for why the envelope lives
// there and not here (internal/auth needs the exact same shape for
// 401s it rejects before a handler ever runs).
type ErrorCode = apierror.Code

type errorResponse = apierror.Response

const (
	ErrCodeInvalidRequestBody = apierror.CodeInvalidRequestBody
	ErrCodeUnauthorized       = apierror.CodeUnauthorized
	ErrCodeInternal           = apierror.CodeInternal

	ErrCodeSessionCreateFailed  = apierror.CodeSessionCreateFailed
	ErrCodeInvalidEmail         = apierror.CodeInvalidEmail
	ErrCodeEmailCodeSendFailed  = apierror.CodeEmailCodeSendFailed
	ErrCodeTooManyAttempts      = apierror.CodeTooManyAttempts
	ErrCodeVerificationExpired  = apierror.CodeVerificationExpired
	ErrCodeInvalidCode          = apierror.CodeInvalidCode
	ErrCodeVerificationFailed   = apierror.CodeVerificationFailed
	ErrCodeInvalidDisplayName   = apierror.CodeInvalidDisplayName
	ErrCodeUserNotFound         = apierror.CodeUserNotFound
	ErrCodeProfileUpdateFailed  = apierror.CodeProfileUpdateFailed
	ErrCodeInvalidRefreshToken  = apierror.CodeInvalidRefreshToken
	ErrCodeSessionRefreshFailed = apierror.CodeSessionRefreshFailed
	ErrCodeLogoutFailed         = apierror.CodeLogoutFailed

	ErrCodeInvalidPersonaID         = apierror.CodeInvalidPersonaID
	ErrCodePersonaNotFound          = apierror.CodePersonaNotFound
	ErrCodePersonaFetchFailed       = apierror.CodePersonaFetchFailed
	ErrCodePersonasListFailed       = apierror.CodePersonasListFailed
	ErrCodePersonaNotAgeAppropriate = apierror.CodePersonaNotAgeAppropriate
	ErrCodePersonaTraitOutOfRange   = apierror.CodePersonaTraitOutOfRange
	ErrCodePersonaTraitsFailed      = apierror.CodePersonaTraitsFailed

	ErrCodeInvalidConversationID      = apierror.CodeInvalidConversationID
	ErrCodeConversationNotFound       = apierror.CodeConversationNotFound
	ErrCodeConversationForbidden      = apierror.CodeConversationForbidden
	ErrCodeConversationCreateFailed   = apierror.CodeConversationCreateFailed
	ErrCodeConversationsListFailed    = apierror.CodeConversationsListFailed
	ErrCodeConversationFetchFailed    = apierror.CodeConversationFetchFailed
	ErrCodeConversationContextMissing = apierror.CodeConversationContextMissing
	ErrCodeMessagesListFailed         = apierror.CodeMessagesListFailed
	ErrCodeInvalidMessageID           = apierror.CodeInvalidMessageID
	ErrCodeMessageNotFound            = apierror.CodeMessageNotFound
	ErrCodeMessageContentRequired     = apierror.CodeMessageContentRequired
	ErrCodeMessageSendFailed          = apierror.CodeMessageSendFailed
	ErrCodeStreamingUnsupported       = apierror.CodeStreamingUnsupported
	ErrCodeReactionEmojiRequired      = apierror.CodeReactionEmojiRequired
	ErrCodeInvalidReactionEmoji       = apierror.CodeInvalidReactionEmoji
	ErrCodeCannotReactToOwnMessage    = apierror.CodeCannotReactToOwnMessage
	ErrCodeReactionFailed             = apierror.CodeReactionFailed

	ErrCodeInvalidAgeRange           = apierror.CodeInvalidAgeRange
	ErrCodeInvalidMoodPreference     = apierror.CodeInvalidMoodPreference
	ErrCodeInvalidSelectedPersonaID  = apierror.CodeInvalidSelectedPersonaID
	ErrCodePreferredNameRequired     = apierror.CodePreferredNameRequired
	ErrCodePreferredNameInvalid      = apierror.CodePreferredNameInvalid
	ErrCodeOnboardingSaveFailed      = apierror.CodeOnboardingSaveFailed
	ErrCodeOnboardingFetchFailed     = apierror.CodeOnboardingFetchFailed
	ErrCodeOnboardingProfileNotFound = apierror.CodeOnboardingProfileNotFound

	ErrCodeQuotaCheckFailed = apierror.CodeQuotaCheckFailed
	ErrCodeQuotaExceeded    = apierror.CodeQuotaExceeded
	ErrCodeChatEnergyFailed = apierror.CodeChatEnergyFailed

	ErrCodeAdminInvalidCredentials = apierror.CodeAdminInvalidCredentials
	ErrCodeAdminLoginFailed        = apierror.CodeAdminLoginFailed
	ErrCodeAdminUnauthorized       = apierror.CodeAdminUnauthorized
	ErrCodeAdminLogoutFailed       = apierror.CodeAdminLogoutFailed
	ErrCodeAdminMeFailed           = apierror.CodeAdminMeFailed

	ErrCodeInvalidUserID            = apierror.CodeInvalidUserID
	ErrCodeAdminUsersListFailed     = apierror.CodeAdminUsersListFailed
	ErrCodeAdminUserFetchFailed     = apierror.CodeAdminUserFetchFailed
	ErrCodeInvalidDailyLimit        = apierror.CodeInvalidDailyLimit
	ErrCodeAdminQuotaUpdateFailed   = apierror.CodeAdminQuotaUpdateFailed
	ErrCodeInvalidCreditAmount      = apierror.CodeInvalidCreditAmount
	ErrCodeAdminCreditsUpdateFailed = apierror.CodeAdminCreditsUpdateFailed

	ErrCodeInvalidPersonaInput      = apierror.CodeInvalidPersonaInput
	ErrCodeAdminPersonasListFailed  = apierror.CodeAdminPersonasListFailed
	ErrCodeAdminPersonaCreateFailed = apierror.CodeAdminPersonaCreateFailed
	ErrCodeAdminPersonaUpdateFailed = apierror.CodeAdminPersonaUpdateFailed
	ErrCodeInvalidPersonaLLMModelID = apierror.CodeInvalidPersonaLLMModelID
	ErrCodePersonaLLMModelNotFound  = apierror.CodePersonaLLMModelNotFound
	ErrCodePersonaLLMModelInactive  = apierror.CodePersonaLLMModelInactive

	ErrCodeAdminMetricsFailed = apierror.CodeAdminMetricsFailed

	ErrCodeAdminConversationsListFailed = apierror.CodeAdminConversationsListFailed
	ErrCodeAdminConversationFetchFailed = apierror.CodeAdminConversationFetchFailed
	ErrCodeAdminMessageDeleteFailed     = apierror.CodeAdminMessageDeleteFailed

	ErrCodeAdminLogsListFailed = apierror.CodeAdminLogsListFailed

	ErrCodeAdminActivityListFailed = apierror.CodeAdminActivityListFailed

	ErrCodeAdminOnboardingInsightsFailed = apierror.CodeAdminOnboardingInsightsFailed

	ErrCodeInvalidLLMCredentialInput      = apierror.CodeInvalidLLMCredentialInput
	ErrCodeLLMCredentialNotFound          = apierror.CodeLLMCredentialNotFound
	ErrCodeLLMCredentialInUse             = apierror.CodeLLMCredentialInUse
	ErrCodeAdminLLMCredentialsListFailed  = apierror.CodeAdminLLMCredentialsListFailed
	ErrCodeAdminLLMCredentialFetchFailed  = apierror.CodeAdminLLMCredentialFetchFailed
	ErrCodeAdminLLMCredentialCreateFailed = apierror.CodeAdminLLMCredentialCreateFailed
	ErrCodeAdminLLMCredentialUpdateFailed = apierror.CodeAdminLLMCredentialUpdateFailed
	ErrCodeAdminLLMCredentialDeleteFailed = apierror.CodeAdminLLMCredentialDeleteFailed
	ErrCodeInvalidLLMModelInput           = apierror.CodeInvalidLLMModelInput
	ErrCodeLLMModelNotFound               = apierror.CodeLLMModelNotFound
	ErrCodeLLMCredentialInactive          = apierror.CodeLLMCredentialInactive
	ErrCodeAdminLLMModelsListFailed       = apierror.CodeAdminLLMModelsListFailed
	ErrCodeAdminLLMModelCreateFailed      = apierror.CodeAdminLLMModelCreateFailed
	ErrCodeAdminLLMModelUpdateFailed      = apierror.CodeAdminLLMModelUpdateFailed
	ErrCodeAdminLLMModelDeleteFailed      = apierror.CodeAdminLLMModelDeleteFailed
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	writeJSON(w, status, apierror.New(code, message))
}

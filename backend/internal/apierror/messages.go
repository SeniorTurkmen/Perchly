package apierror

// messages holds a translated display string per (Code, Locale), used
// only when the resolved request locale isn't LocaleTR — see Message.
// Turkish itself is never in here: every handler call site already
// carries its own Turkish string inline (that's still what ships for
// LocaleTR and for any locale with no entry below), so duplicating it
// here would be a second source of truth for the same text.
//
// Only "en" is filled in for now (i18n Faz 0: infrastructure + one
// validation language). "de"/"ar"/"es"/"fr"/"ru"/"zh" entries are added
// in the translation passes (Faz 1/2) — an unfilled locale simply falls
// back to the Turkish inline string, same as today, never a missing- or
// empty-string response.
var messages = map[Code]map[Locale]string{
	// Generic.
	CodeInvalidRequestBody: {LocaleEN: "invalid request body"},
	CodeUnauthorized:       {LocaleEN: "login required"},
	CodeInternal:           {LocaleEN: "internal error"},

	// Auth.
	CodeSessionCreateFailed:  {LocaleEN: "could not create session"},
	CodeInvalidEmail:         {LocaleEN: "invalid email"},
	CodeEmailCodeSendFailed:  {LocaleEN: "could not send verification code"},
	CodeTooManyAttempts:      {LocaleEN: "too many attempts, please try again later"},
	CodeVerificationExpired:  {LocaleEN: "verification code expired"},
	CodeInvalidCode:          {LocaleEN: "invalid verification code"},
	CodeVerificationFailed:   {LocaleEN: "verification failed"},
	CodeInvalidDisplayName:   {LocaleEN: "invalid display name"},
	CodeUserNotFound:         {LocaleEN: "user not found"},
	CodeProfileUpdateFailed:  {LocaleEN: "could not update profile"},
	CodeInvalidRefreshToken:  {LocaleEN: "invalid refresh token"},
	CodeSessionRefreshFailed: {LocaleEN: "could not refresh session"},
	CodeLogoutFailed:         {LocaleEN: "logout failed"},

	// Personas.
	CodeInvalidPersonaID:         {LocaleEN: "invalid persona id"},
	CodePersonaNotFound:          {LocaleEN: "persona not found"},
	CodePersonaFetchFailed:       {LocaleEN: "could not fetch persona"},
	CodePersonasListFailed:       {LocaleEN: "could not list personas"},
	CodePersonaNotAgeAppropriate: {LocaleEN: "this persona isn't available for your age range"},
	CodePersonaTraitOutOfRange:   {LocaleEN: "persona trait value out of range"},
	CodePersonaTraitsFailed:      {LocaleEN: "could not update persona traits"},

	// Conversations & messages.
	CodeInvalidConversationID:      {LocaleEN: "invalid conversation id"},
	CodeConversationNotFound:       {LocaleEN: "conversation not found"},
	CodeConversationForbidden:      {LocaleEN: "you don't have access to this conversation"},
	CodeConversationCreateFailed:   {LocaleEN: "could not create conversation"},
	CodeConversationsListFailed:    {LocaleEN: "could not list conversations"},
	CodeConversationFetchFailed:    {LocaleEN: "could not fetch conversation"},
	CodeConversationContextMissing: {LocaleEN: "conversation context missing"},
	CodeMessagesListFailed:         {LocaleEN: "could not list messages"},
	CodeInvalidMessageID:           {LocaleEN: "invalid message id"},
	CodeMessageNotFound:            {LocaleEN: "message not found"},
	CodeMessageContentRequired:     {LocaleEN: "content is required"},
	CodeMessageSendFailed:          {LocaleEN: "could not send message"},
	CodeStreamingUnsupported:       {LocaleEN: "streaming isn't supported"},
	CodeReactionEmojiRequired:      {LocaleEN: "an emoji is required"},
	CodeInvalidReactionEmoji:       {LocaleEN: "invalid reaction emoji"},
	CodeCannotReactToOwnMessage:    {LocaleEN: "you can't react to your own message"},
	CodeReactionFailed:             {LocaleEN: "could not save reaction"},

	// Onboarding.
	CodeInvalidAgeRange:           {LocaleEN: "invalid age range"},
	CodeInvalidMoodPreference:     {LocaleEN: "invalid mood preference"},
	CodeInvalidSelectedPersonaID:  {LocaleEN: "invalid selected persona id"},
	CodePreferredNameRequired:     {LocaleEN: "preferred name is required"},
	CodePreferredNameInvalid:      {LocaleEN: "invalid preferred name"},
	CodeOnboardingSaveFailed:      {LocaleEN: "could not save onboarding profile"},
	CodeOnboardingFetchFailed:     {LocaleEN: "could not fetch onboarding profile"},
	CodeOnboardingProfileNotFound: {LocaleEN: "onboarding profile not found"},

	// Quota / chat energy.
	CodeQuotaCheckFailed: {LocaleEN: "could not check quota"},
	CodeQuotaExceeded:    {LocaleEN: "you've reached today's message limit"},
	CodeChatEnergyFailed: {LocaleEN: "could not load chat energy"},

	// Admin dashboard auth.
	CodeAdminInvalidCredentials: {LocaleEN: "invalid email or password"},
	CodeAdminLoginFailed:        {LocaleEN: "login failed"},
	CodeAdminUnauthorized:       {LocaleEN: "admin login required"},
	CodeAdminLogoutFailed:       {LocaleEN: "logout failed"},
	CodeAdminMeFailed:           {LocaleEN: "could not fetch admin profile"},

	// Admin dashboard — users.
	CodeInvalidUserID:            {LocaleEN: "invalid user id"},
	CodeAdminUsersListFailed:     {LocaleEN: "could not list users"},
	CodeAdminUserFetchFailed:     {LocaleEN: "could not fetch user"},
	CodeInvalidDailyLimit:        {LocaleEN: "daily limit must be 0 or greater"},
	CodeAdminQuotaUpdateFailed:   {LocaleEN: "could not update quota"},
	CodeInvalidCreditAmount:      {LocaleEN: "credit amount must be 0 or greater"},
	CodeAdminCreditsUpdateFailed: {LocaleEN: "could not update credits"},

	// Admin dashboard — personas.
	CodeInvalidPersonaInput:      {LocaleEN: "invalid persona input"},
	CodeAdminPersonasListFailed:  {LocaleEN: "could not list personas"},
	CodeAdminPersonaCreateFailed: {LocaleEN: "could not create persona"},
	CodeAdminPersonaUpdateFailed: {LocaleEN: "could not update persona"},
	CodeInvalidPersonaLLMModelID: {LocaleEN: "invalid model id"},
	CodePersonaLLMModelNotFound:  {LocaleEN: "the selected model wasn't found"},
	CodePersonaLLMModelInactive:  {LocaleEN: "the selected model or its credential is inactive"},

	// Admin dashboard — persona translations.
	CodeInvalidPersonaTranslationLocale:     {LocaleEN: "invalid or unsupported translation locale"},
	CodeInvalidPersonaTranslationInput:      {LocaleEN: "invalid translation input"},
	CodeAdminPersonaTranslationsListFailed:  {LocaleEN: "could not list persona translations"},
	CodeAdminPersonaTranslationUpsertFailed: {LocaleEN: "could not save persona translation"},
	CodeAdminPersonaTranslationDeleteFailed: {LocaleEN: "could not delete persona translation"},

	// Admin dashboard — home metrics.
	CodeAdminMetricsFailed: {LocaleEN: "could not load dashboard metrics"},

	// Admin dashboard — conversations & moderation.
	CodeAdminConversationsListFailed: {LocaleEN: "could not list conversations"},
	CodeAdminConversationFetchFailed: {LocaleEN: "could not fetch conversation"},
	CodeAdminMessageDeleteFailed:     {LocaleEN: "could not delete message"},

	// Admin dashboard — request logs.
	CodeAdminLogsListFailed: {LocaleEN: "could not list request logs"},

	// Admin dashboard — own activity log.
	CodeAdminActivityListFailed: {LocaleEN: "could not list admin activity"},

	// Admin dashboard — onboarding insights.
	CodeAdminOnboardingInsightsFailed: {LocaleEN: "could not load onboarding insights"},

	// Admin dashboard — LLM provider credentials & models.
	CodeInvalidLLMCredentialInput:      {LocaleEN: "invalid credential input"},
	CodeLLMCredentialNotFound:          {LocaleEN: "credential not found"},
	CodeLLMCredentialInUse:             {LocaleEN: "remove this credential's models first"},
	CodeAdminLLMCredentialsListFailed:  {LocaleEN: "could not list credentials"},
	CodeAdminLLMCredentialFetchFailed:  {LocaleEN: "could not fetch credential"},
	CodeAdminLLMCredentialCreateFailed: {LocaleEN: "could not create credential"},
	CodeAdminLLMCredentialUpdateFailed: {LocaleEN: "could not update credential"},
	CodeAdminLLMCredentialDeleteFailed: {LocaleEN: "could not delete credential"},
	CodeInvalidLLMModelInput:           {LocaleEN: "invalid model input"},
	CodeLLMModelNotFound:               {LocaleEN: "model not found"},
	CodeLLMCredentialInactive:          {LocaleEN: "credential is inactive"},
	CodeAdminLLMModelsListFailed:       {LocaleEN: "could not list models"},
	CodeAdminLLMModelCreateFailed:      {LocaleEN: "could not create model"},
	CodeAdminLLMModelUpdateFailed:      {LocaleEN: "could not update model"},
	CodeAdminLLMModelDeleteFailed:      {LocaleEN: "could not delete model"},
}

// Message resolves the display text for code in locale: the Turkish
// fallback itself when locale is LocaleTR (or unrecognized) or when no
// translation has been added yet, otherwise the translated string.
// fallback is always what the call site would have shown before i18n
// existed, so a code with no entry in messages — including every
// locale that isn't "en" yet — degrades to exactly today's behavior,
// never an empty or missing string.
func Message(code Code, locale Locale, fallback string) string {
	if locale == LocaleTR {
		return fallback
	}
	if translated, ok := messages[code][locale]; ok {
		return translated
	}
	return fallback
}

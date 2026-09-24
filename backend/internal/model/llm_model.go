package model

import "time"

// LLMModel is one model made callable through a specific LLMCredential
// (a single provider API key can usually call several models). Personas
// will later reference one of these (see the planned personas.llm_model_id
// column) instead of the process-wide LLM_PROVIDER/LLM_MODEL env vars.
type LLMModel struct {
	ID           string
	CredentialID string
	ModelName    string
	DisplayName  string
	IsDefault    bool
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

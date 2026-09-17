package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"perchly-backend/internal/embedding"
	"perchly-backend/internal/model"
)

type fakeSummaryRepoForContextBuilder struct{}

func (fakeSummaryRepoForContextBuilder) GetByConversationID(context.Context, string) (model.ConversationSummary, error) {
	return model.ConversationSummary{}, errors.New("no summary")
}

func (fakeSummaryRepoForContextBuilder) Upsert(context.Context, string, string, int) error {
	return nil
}

type fakeMessageEmbeddingSearcher struct{}

func (fakeMessageEmbeddingSearcher) NearestByConversation(context.Context, string, []float32, int, []string) ([]model.Message, error) {
	return nil, nil
}

func TestContextBuilder_Build_IncludesTraitsInstruction(t *testing.T) {
	builder := NewContextBuilder(
		fakeSummaryRepoForContextBuilder{},
		fakeMessageEmbeddingSearcher{},
		embedding.NewUnconfiguredClient(errors.New("no embedding client in this test")),
	)

	traits := model.PersonaTraits{Warmth: 90, Humor: 10, Wisdom: 55, Directness: 95, Energy: 20}
	messages := builder.Build(context.Background(), "conv-1", "sen Ada'sın", nil, "merhaba", traits)

	if got, want := messages[0].Content, "sen Ada'sın"; got != want {
		t.Fatalf("first system message = %q, want the persona system prompt %q", got, want)
	}

	var traitsMessage string
	for _, m := range messages {
		if strings.Contains(m.Content, "Sıcaklık") {
			traitsMessage = m.Content
			break
		}
	}
	if traitsMessage == "" {
		t.Fatalf("expected a system message carrying the traits instruction, got: %+v", messages)
	}

	for _, want := range []string{"Sıcaklık %90", "Espri %10", "Bilgelik %55", "Doğrudanlık %95", "Enerji %20"} {
		if !strings.Contains(traitsMessage, want) {
			t.Errorf("traits instruction missing %q: %s", want, traitsMessage)
		}
	}
	// High directness should read as blunt, not softened.
	if !strings.Contains(traitsMessage, "sert ve doğrudan") {
		t.Errorf("expected the top directness tier's phrase for value 95, got: %s", traitsMessage)
	}
	// Low humor should read as fully serious.
	if !strings.Contains(traitsMessage, "tamamen ciddi") {
		t.Errorf("expected the lowest humor tier's phrase for value 10, got: %s", traitsMessage)
	}
}

func TestTraitTier(t *testing.T) {
	tests := []struct {
		value    int
		wantTier int
	}{
		{0, 0}, {19, 0}, {20, 1}, {39, 1}, {40, 2}, {59, 2}, {60, 3}, {79, 3}, {80, 4}, {100, 4},
	}
	for _, tt := range tests {
		if got := traitTier(tt.value); got != tt.wantTier {
			t.Errorf("traitTier(%d) = %d, want %d", tt.value, got, tt.wantTier)
		}
	}
}

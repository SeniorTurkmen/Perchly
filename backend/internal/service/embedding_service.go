package service

import (
	"context"
	"log"
	"time"

	"perchly-backend/internal/embedding"
)

// MessageEmbeddingWriter is the persistence dependency EmbeddingService
// needs.
type MessageEmbeddingWriter interface {
	Create(ctx context.Context, messageID, conversationID string, vector []float32) error
}

// EmbeddingService generates and stores an embedding for each message in
// the background, off the request path, so chat latency never depends on
// the embedding provider.
type EmbeddingService struct {
	client embedding.Client
	repo   MessageEmbeddingWriter
}

func NewEmbeddingService(client embedding.Client, repo MessageEmbeddingWriter) *EmbeddingService {
	return &EmbeddingService{client: client, repo: repo}
}

// EmbedMessageAsync fires off a goroutine to embed and store one
// message's content. It intentionally uses its own background context
// with a fixed timeout rather than the caller's request context, which
// will already be canceled (the HTTP response finished) by the time this
// goroutine actually runs.
func (s *EmbeddingService) EmbedMessageAsync(conversationID, messageID, content string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		vector, err := s.client.Embed(ctx, content)
		if err != nil {
			log.Printf("embedding generation failed (message=%s): %v", messageID, err)
			return
		}
		if err := s.repo.Create(ctx, messageID, conversationID, vector); err != nil {
			log.Printf("embedding save failed (message=%s): %v", messageID, err)
		}
	}()
}

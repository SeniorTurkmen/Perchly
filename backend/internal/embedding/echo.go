package embedding

import (
	"context"
	"math"
	"strings"
)

const echoDimensions = 1536

// EchoClient produces a deterministic embedding via feature hashing (the
// "hashing trick"): each word hashes into one of 1536 buckets, which are
// then L2-normalized. No API key needed, and unlike random noise, text
// sharing words ends up genuinely closer in cosine distance than
// unrelated text — enough to exercise nearest-neighbor search
// meaningfully for local dev/tests. Select it with EMBEDDING_PROVIDER=echo.
type EchoClient struct{}

func NewEchoClient() *EchoClient {
	return &EchoClient{}
}

func (c *EchoClient) Embed(_ context.Context, text string) ([]float32, error) {
	vector := make([]float32, echoDimensions)
	for _, word := range strings.Fields(strings.ToLower(text)) {
		bucket := fnv1a(word) % uint32(echoDimensions)
		vector[bucket]++
	}
	normalize(vector)
	return vector, nil
}

func fnv1a(s string) uint32 {
	const (
		offsetBasis uint32 = 2166136261
		prime       uint32 = 16777619
	)
	h := offsetBasis
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime
	}
	return h
}

func normalize(v []float32) {
	var sumSquares float64
	for _, x := range v {
		sumSquares += float64(x) * float64(x)
	}
	if sumSquares == 0 {
		return
	}
	norm := float32(math.Sqrt(sumSquares))
	for i := range v {
		v[i] /= norm
	}
}

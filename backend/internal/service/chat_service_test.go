package service

import (
	"strings"
	"testing"
)

// TestScanForReactionTag_PlainTextTailSafetyWindow checks the no-bracket
// case (the vast majority of turns) by invariant rather than a hand-
// counted expected string: exactly reactionTagTailSafetyLen runes are
// held back — no more, no less — and flush+pending always reassembles
// to the original text.
func TestScanForReactionTag_PlainTextTailSafetyWindow(t *testing.T) {
	tests := []string{
		"Merhaba, nasılsın",
		"Sela",
		"",
		"❤️😂 emoji dolu bir mesaj örneği",
	}

	for _, buffer := range tests {
		t.Run(buffer, func(t *testing.T) {
			result := scanForReactionTag(buffer)
			if result.resolved {
				t.Fatalf("scanForReactionTag(%q) resolved = true, want false (no tag present)", buffer)
			}
			if got, want := result.flush+result.pending, buffer; got != want {
				t.Fatalf("flush+pending = %q, want original %q", got, want)
			}
			wantPendingRunes := min(len([]rune(buffer)), reactionTagTailSafetyLen)
			if got := len([]rune(result.pending)); got != wantPendingRunes {
				t.Fatalf("pending has %d runes, want exactly %d (the tail safety window)", got, wantPendingRunes)
			}
		})
	}
}

func TestScanForReactionTag(t *testing.T) {
	tests := []struct {
		name         string
		buffer       string
		wantFlush    string
		wantPending  string
		wantResolved bool
		wantEmoji    string
	}{
		{
			name:        "an open tag with no close yet flushes everything before it and holds the rest",
			buffer:      "Merhaba [[REACT:",
			wantFlush:   "Merhaba ",
			wantPending: "[[REACT:",
		},
		{
			name:         "a complete valid tag with no trailing text resolves with the emoji",
			buffer:       "[[REACT:❤️]]",
			wantFlush:    "",
			wantResolved: true,
			wantEmoji:    "❤️",
		},
		{
			name:         "a complete valid tag followed by a written reply resolves and flushes the reply",
			buffer:       "[[REACT:👍]] Harika haber!",
			wantFlush:    " Harika haber!",
			wantResolved: true,
			wantEmoji:    "👍",
		},
		{
			name:         "a complete tag naming an emoji outside the allowed set is stripped but yields no reaction",
			buffer:       "[[REACT:🐸]] merhaba",
			wantFlush:    " merhaba",
			wantResolved: true,
			wantEmoji:    "",
		},
		{
			name:         "a tag preceded by a stray model preamble is still found and stripped",
			buffer:       "Assign(read([[REACT:❤️]]))\n\nMerhaba!",
			wantFlush:    "Assign(read())\n\nMerhaba!",
			wantResolved: true,
			wantEmoji:    "❤️",
		},
		{
			name:         "an open tag that never closes within the scan limit gives up and flushes everything",
			buffer:       "[[REACT:" + strings.Repeat("x", maxReactionTagScanLen),
			wantFlush:    "[[REACT:" + strings.Repeat("x", maxReactionTagScanLen),
			wantResolved: true,
			wantEmoji:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanForReactionTag(tt.buffer)
			if result.flush != tt.wantFlush {
				t.Errorf("scanForReactionTag(%q).flush = %q, want %q", tt.buffer, result.flush, tt.wantFlush)
			}
			if result.pending != tt.wantPending {
				t.Errorf("scanForReactionTag(%q).pending = %q, want %q", tt.buffer, result.pending, tt.wantPending)
			}
			if result.resolved != tt.wantResolved {
				t.Errorf("scanForReactionTag(%q).resolved = %v, want %v", tt.buffer, result.resolved, tt.wantResolved)
			}
			if result.emoji != tt.wantEmoji {
				t.Errorf("scanForReactionTag(%q).emoji = %q, want %q", tt.buffer, result.emoji, tt.wantEmoji)
			}
		})
	}
}

// TestScanForReactionTag_IncrementalStreamingPreservesGranularity guards
// against the fix's whole point: plain text (the common case, no
// reaction) must keep streaming to the client roughly as it arrives, not
// get buffered in one lump until the reply ends or the scan limit is hit.
func TestScanForReactionTag_IncrementalStreamingPreservesGranularity(t *testing.T) {
	deltas := []string{"Merhaba! ", "Bugün ", "nasılsın?"}
	var pending strings.Builder
	var flushCount int
	var flushed strings.Builder

	for _, delta := range deltas {
		pending.WriteString(delta)
		result := scanForReactionTag(pending.String())
		pending.Reset()
		pending.WriteString(result.pending)
		if result.flush != "" {
			flushCount++
			flushed.WriteString(result.flush)
		}
	}
	flushed.WriteString(pending.String()) // final drain, as SendMessage does once the stream ends

	if flushCount < 2 {
		t.Fatalf("expected multiple incremental flushes for plain streamed text, got %d", flushCount)
	}
	want := strings.Join(deltas, "")
	if got := flushed.String(); got != want {
		t.Fatalf("reassembled flushed text = %q, want %q", got, want)
	}
}

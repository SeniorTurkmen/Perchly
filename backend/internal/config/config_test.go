package config

import "testing"

func TestNormalizeSMTPPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "already compact", in: "abcdabcdabcdabcd", want: "abcdabcdabcdabcd"},
		{name: "google display spaces", in: "abcd efgh ijkl mnop", want: "abcdefghijklmnop"},
		{name: "quoted google display", in: `"abcd efgh ijkl mnop"`, want: "abcdefghijklmnop"},
		{name: "nbsp from copy-paste", in: "abcd\u00a0efgh\u00a0ijkl\u00a0mnop", want: "abcdefghijklmnop"},
		{name: "quoted nbsp", in: "\"abcd\u00a0efgh\u00a0ijkl\u00a0mnop\"", want: "abcdefghijklmnop"},
		{name: "empty", in: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeSMTPPassword(tt.in); got != tt.want {
				t.Fatalf("normalizeSMTPPassword(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

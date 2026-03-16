package tts

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name string
		text string
		lang string
		want map[string]string // expected query params
	}{
		{
			name: "basic english",
			text: "hello world",
			lang: "en",
			want: map[string]string{"q": "hello world", "tl": "en", "client": "tw-ob", "ie": "UTF-8"},
		},
		{
			name: "french with accents",
			text: "café résumé",
			lang: "fr",
			want: map[string]string{"q": "café résumé", "tl": "fr"},
		},
		{
			name: "special characters get encoded",
			text: "hello & goodbye",
			lang: "en",
			want: map[string]string{"q": "hello & goodbye"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildURL(tt.text, tt.lang)

			if !strings.HasPrefix(got, ttsBaseURL+"?") {
				t.Errorf("URL should start with base URL, got %s", got)
			}

			u, err := url.Parse(got)
			if err != nil {
				t.Fatalf("invalid URL: %v", err)
			}

			for key, want := range tt.want {
				if got := u.Query().Get(key); got != want {
					t.Errorf("param %s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestChunkText_Basic(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantChunks int
		wantErr    bool
	}{
		{
			name:       "short text is one chunk",
			text:       "Hello world",
			wantChunks: 1,
		},
		{
			name:       "empty text returns error",
			text:       "",
			wantChunks: 0,
			wantErr:    true,
		},
		{
			name:       "whitespace only returns error",
			text:       "   \t\n  ",
			wantChunks: 0,
			wantErr:    true,
		},
		{
			name:       "exactly maxChunkLen chars",
			text:       strings.Repeat("a", maxChunkLen),
			wantChunks: 1,
		},
		{
			name:       "one char over maxChunkLen splits",
			text:       strings.Repeat("a", maxChunkLen) + " b",
			wantChunks: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := ChunkText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ChunkText() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(chunks) != tt.wantChunks {
				t.Errorf("got %d chunks, want %d", len(chunks), tt.wantChunks)
			}
		})
	}
}

func TestChunkText_MaxLength(t *testing.T) {
	// Text at exactly the limit should work.
	text := strings.Repeat("a", maxTextLen)
	_, err := ChunkText(text)
	if err != nil {
		t.Errorf("text at max length should not error, got: %v", err)
	}

	// One character over should fail.
	text = strings.Repeat("a", maxTextLen+1)
	_, err = ChunkText(text)
	if err == nil {
		t.Error("text over max length should return error")
	}
}

func TestChunkText_SentenceBoundary(t *testing.T) {
	// Two sentences that together exceed maxChunkLen but individually fit.
	s1 := strings.Repeat("a", 100) + "."
	s2 := strings.Repeat("b", 100) + "."
	text := s1 + " " + s2

	chunks, err := ChunkText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %v", len(chunks), chunks)
	}
	if chunks[0] != s1 {
		t.Errorf("first chunk = %q, want %q", chunks[0], s1)
	}
	if chunks[1] != s2 {
		t.Errorf("second chunk = %q, want %q", chunks[1], s2)
	}
}

func TestChunkText_WordBoundary(t *testing.T) {
	// No sentence boundary, should split at word boundary.
	word := strings.Repeat("x", 50)
	text := strings.Join([]string{word, word, word, word, word}, " ")
	// 5 words of 50 chars + 4 spaces = 254 chars, should split into 2 chunks

	chunks, err := ChunkText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	// Each chunk should not exceed maxChunkLen runes.
	for i, c := range chunks {
		if len([]rune(c)) > maxChunkLen {
			t.Errorf("chunk %d exceeds maxChunkLen: %d runes", i, len([]rune(c)))
		}
	}
}

func TestChunkText_NoSpaces(t *testing.T) {
	// Long text with no spaces or sentence boundaries — force hard split.
	text := strings.Repeat("x", maxChunkLen+50)
	chunks, err := ChunkText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if len([]rune(chunks[0])) != maxChunkLen {
		t.Errorf("first chunk should be %d runes, got %d", maxChunkLen, len([]rune(chunks[0])))
	}
}

func TestChunkText_Unicode(t *testing.T) {
	// Each emoji is one rune but multiple bytes.
	emoji := "😀"
	text := strings.Repeat(emoji, maxChunkLen+10)
	chunks, err := ChunkText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if len([]rune(chunks[0])) != maxChunkLen {
		t.Errorf("first chunk should be %d runes, got %d", maxChunkLen, len([]rune(chunks[0])))
	}
}

func TestChunkText_PreservesContent(t *testing.T) {
	text := "First sentence. Second sentence. Third sentence."
	chunks, err := ChunkText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Rejoin should match original (modulo whitespace normalization).
	rejoined := strings.Join(chunks, " ")
	if rejoined != text {
		t.Errorf("rejoined chunks = %q, want %q", rejoined, text)
	}
}

func TestChunkText_SentenceBoundaryTypes(t *testing.T) {
	tests := []struct {
		name string
		sep  string
	}{
		{"period", "."},
		{"exclamation", "!"},
		{"question", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s1 := strings.Repeat("a", 100) + tt.sep
			s2 := strings.Repeat("b", 100)
			text := s1 + " " + s2

			chunks, err := ChunkText(text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(chunks) < 2 {
				t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
			}
			if chunks[0] != s1 {
				t.Errorf("first chunk = %q, want %q", chunks[0], s1)
			}
		})
	}
}

func TestBuildURLs(t *testing.T) {
	text := "Hello world"
	urls, err := BuildURLs(text, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	if !strings.Contains(urls[0], "q=Hello+world") || !strings.Contains(urls[0], "tl=en") {
		t.Errorf("URL missing expected params: %s", urls[0])
	}
}

func TestBuildURLs_EmptyText(t *testing.T) {
	_, err := BuildURLs("", "en")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestBuildURLs_MultipleChunks(t *testing.T) {
	s1 := strings.Repeat("a", 150) + "."
	s2 := strings.Repeat("b", 100)
	text := s1 + " " + s2

	urls, err := BuildURLs(text, "de")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	for _, u := range urls {
		if !strings.Contains(u, "tl=de") {
			t.Errorf("URL missing language param: %s", u)
		}
	}
}

func TestLastSentenceBoundary(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello world.", 12},
		{"hello! world", 6},
		{"what? ok", 5},
		{"no boundary here", -1},
		{"first. second.", 14},
		{"", -1},
	}

	for _, tt := range tests {
		got := lastSentenceBoundary(tt.input)
		if got != tt.want {
			t.Errorf("lastSentenceBoundary(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestLastWordBoundary(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello world", 5},
		{"one two three", 7},
		{"nospaceshere", -1},
		{"trailing ", 8},
		{"", -1},
	}

	for _, tt := range tests {
		got := lastWordBoundary(tt.input)
		if got != tt.want {
			t.Errorf("lastWordBoundary(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestSubstringRunes(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 3, "hel"},
		{"hello", 10, "hello"},
		{"hello", 0, ""},
		{"café", 3, "caf"},
		{"😀😀😀", 2, "😀😀"},
	}

	for _, tt := range tests {
		got := substringRunes(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("substringRunes(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}

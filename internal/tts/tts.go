package tts

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	maxChunkLen = 200
	maxTextLen  = 5000
	ttsBaseURL  = "https://translate.google.com/translate_tts"
)

// BuildURL returns a Google Translate TTS URL for the given text and language.
func BuildURL(text, lang string) string {
	v := url.Values{}
	v.Set("ie", "UTF-8")
	v.Set("client", "tw-ob")
	v.Set("tl", lang)
	v.Set("q", text)
	return ttsBaseURL + "?" + v.Encode()
}

// ChunkText splits text into chunks of at most maxChunkLen characters,
// breaking at sentence boundaries (.!?) first, then at word boundaries.
func ChunkText(text string) ([]string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}
	if utf8.RuneCountInString(text) > maxTextLen {
		return nil, fmt.Errorf("text exceeds maximum length of %d characters", maxTextLen)
	}

	var chunks []string
	for len(text) > 0 {
		if utf8.RuneCountInString(text) <= maxChunkLen {
			chunks = append(chunks, strings.TrimSpace(text))
			break
		}

		chunk := substringRunes(text, maxChunkLen)
		cut := lastSentenceBoundary(chunk)
		if cut < 0 {
			cut = lastWordBoundary(chunk)
		}
		if cut < 0 {
			cut = maxChunkLen
		}

		raw := substringRunes(text, cut)
		piece := strings.TrimSpace(raw)
		if piece != "" {
			chunks = append(chunks, piece)
		}
		text = strings.TrimSpace(text[len(raw):])
	}
	return chunks, nil
}

// BuildURLs splits text into chunks and returns a TTS URL for each chunk.
func BuildURLs(text, lang string) ([]string, error) {
	chunks, err := ChunkText(text)
	if err != nil {
		return nil, err
	}
	urls := make([]string, len(chunks))
	for i, chunk := range chunks {
		urls[i] = BuildURL(chunk, lang)
	}
	return urls, nil
}

func lastSentenceBoundary(s string) int {
	idx := strings.LastIndexAny(s, ".!?")
	if idx < 0 {
		return -1
	}
	return utf8.RuneCountInString(s[:idx]) + 1
}

func lastWordBoundary(s string) int {
	idx := strings.LastIndexByte(s, ' ')
	if idx < 0 {
		return -1
	}
	return utf8.RuneCountInString(s[:idx])
}

func substringRunes(s string, n int) string {
	i := 0
	for j := 0; j < n && i < len(s); j++ {
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
	}
	return s[:i]
}

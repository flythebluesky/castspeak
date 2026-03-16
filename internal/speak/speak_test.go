package speak

import (
	"context"
	"testing"
	"time"
)

func TestSpeak_EmptyText(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, _, err := Speak(ctx, "", "device", "", "en")
	if err == nil {
		t.Error("expected error for empty text")
	}
	if err.Error() != "text is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSpeak_MissingDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, _, err := Speak(ctx, "hello", "", "", "en")
	if err == nil {
		t.Error("expected error for missing device")
	}
	if err.Error() != "device_name or device_uuid is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSpeak_DefaultLanguage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// This will fail at discovery (no network), but we're testing that
	// empty language doesn't cause a panic — the error should be from discovery.
	_, _, err := Speak(ctx, "hello", "device", "", "")
	if err == nil {
		t.Error("expected error (discovery should fail)")
	}
}

func TestSpeak_TextTooLong(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	longText := make([]byte, 5001)
	for i := range longText {
		longText[i] = 'a'
	}

	_, _, err := Speak(ctx, string(longText), "device", "", "en")
	if err == nil {
		t.Error("expected error for text exceeding max length")
	}
}

func TestFindDevice_MissingIdentifier(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := findDevice(ctx, "", "")
	if err == nil {
		t.Error("expected error for missing device identifier")
	}
	if err.Error() != "device_name or device_uuid is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetVolume_MissingDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := SetVolume(ctx, "", "", 0.5)
	if err == nil {
		t.Error("expected error for missing device")
	}
}

func TestSetMuted_MissingDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := SetMuted(ctx, "", "", true)
	if err == nil {
		t.Error("expected error for missing device")
	}
}

func TestStop_MissingDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := Stop(ctx, "", "")
	if err == nil {
		t.Error("expected error for missing device")
	}
}

func TestStatus_MissingDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, _, err := Status(ctx, "", "")
	if err == nil {
		t.Error("expected error for missing device")
	}
}

func TestPlayURL_MissingDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := PlayURL(ctx, "", "", "http://example.com/audio.mp3")
	if err == nil {
		t.Error("expected error for missing device")
	}
}

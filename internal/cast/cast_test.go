package cast

import (
	"testing"
)

func TestDeviceStatus_ZeroValue(t *testing.T) {
	var ds DeviceStatus

	if ds.App != "" {
		t.Errorf("App = %q, want empty", ds.App)
	}
	if ds.Idle != false {
		t.Error("Idle should default to false")
	}
	if ds.Volume != 0 {
		t.Errorf("Volume = %f, want 0", ds.Volume)
	}
	if ds.Muted != false {
		t.Error("Muted should default to false")
	}
	if ds.PlayerState != "" {
		t.Errorf("PlayerState = %q, want empty", ds.PlayerState)
	}
	if ds.MediaID != "" {
		t.Errorf("MediaID = %q, want empty", ds.MediaID)
	}
}

func TestPlayURLs_EmptySlice(t *testing.T) {
	// Playing zero URLs on a nonexistent device should fail at connection,
	// not panic on an empty slice.
	err := PlayURLs("192.168.255.255", 8009, nil)
	if err == nil {
		t.Error("expected connection error for unreachable address")
	}
}

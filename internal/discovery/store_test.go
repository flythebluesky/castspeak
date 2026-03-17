package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempConfigDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	configDirFunc = func() (string, error) { return tmp, nil }
	t.Cleanup(func() { configDirFunc = os.UserConfigDir })
}

func TestSaveAndLoadDevices(t *testing.T) {
	withTempConfigDir(t)

	devices := []Device{
		{Name: "Kitchen", UUID: "uuid-1", Addr: "192.168.1.10", Port: 8009, Model: "Nest Mini"},
		{Name: "Office", UUID: "uuid-2", Addr: "192.168.1.11", Port: 8009},
	}

	if err := SaveDevices(devices); err != nil {
		t.Fatalf("SaveDevices: %v", err)
	}

	loaded, err := LoadDevices()
	if err != nil {
		t.Fatalf("LoadDevices: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("got %d devices, want 2", len(loaded))
	}
	if loaded[0].Name != "Kitchen" || loaded[0].UUID != "uuid-1" {
		t.Errorf("device[0] = %+v, want Kitchen/uuid-1", loaded[0])
	}
	if loaded[1].Name != "Office" || loaded[1].UUID != "uuid-2" {
		t.Errorf("device[1] = %+v, want Office/uuid-2", loaded[1])
	}
}

func TestLoadDevices_MissingFile(t *testing.T) {
	withTempConfigDir(t)

	devices, err := LoadDevices()
	if err != nil {
		t.Fatalf("LoadDevices: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("got %d devices, want 0", len(devices))
	}
}

func TestFindSavedDevice_ByName(t *testing.T) {
	withTempConfigDir(t)

	SaveDevices([]Device{
		{Name: "Kitchen", UUID: "uuid-1", Addr: "192.168.1.10", Port: 8009},
	})

	dev, err := FindSavedDevice("Kitchen", "")
	if err != nil {
		t.Fatalf("FindSavedDevice: %v", err)
	}
	if dev.UUID != "uuid-1" {
		t.Errorf("UUID = %q, want uuid-1", dev.UUID)
	}
}

func TestFindSavedDevice_ByUUID(t *testing.T) {
	withTempConfigDir(t)

	SaveDevices([]Device{
		{Name: "Kitchen", UUID: "uuid-1", Addr: "192.168.1.10", Port: 8009},
	})

	dev, err := FindSavedDevice("", "uuid-1")
	if err != nil {
		t.Fatalf("FindSavedDevice: %v", err)
	}
	if dev.Name != "Kitchen" {
		t.Errorf("Name = %q, want Kitchen", dev.Name)
	}
}

func TestFindSavedDevice_CaseInsensitive(t *testing.T) {
	withTempConfigDir(t)

	SaveDevices([]Device{
		{Name: "Kitchen Speaker", UUID: "UUID-ABC", Addr: "192.168.1.10", Port: 8009},
	})

	dev, err := FindSavedDevice("kitchen speaker", "")
	if err != nil {
		t.Fatalf("FindSavedDevice: %v", err)
	}
	if dev.Name != "Kitchen Speaker" {
		t.Errorf("Name = %q, want Kitchen Speaker", dev.Name)
	}
}

func TestFindSavedDevice_NotFound(t *testing.T) {
	withTempConfigDir(t)

	SaveDevices([]Device{
		{Name: "Kitchen", UUID: "uuid-1", Addr: "192.168.1.10", Port: 8009},
	})

	_, err := FindSavedDevice("Bedroom", "")
	if err == nil {
		t.Error("expected error for missing device")
	}
}

func TestRemoveSavedDevices(t *testing.T) {
	withTempConfigDir(t)

	SaveDevices([]Device{
		{Name: "Kitchen", UUID: "uuid-1", Addr: "192.168.1.10", Port: 8009},
	})

	if err := RemoveSavedDevices(); err != nil {
		t.Fatalf("RemoveSavedDevices: %v", err)
	}

	// Verify file is gone
	path, _ := ConfigPath()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("devices file should be removed")
	}
}

func TestRemoveSavedDevices_NoFile(t *testing.T) {
	withTempConfigDir(t)

	if err := RemoveSavedDevices(); err != nil {
		t.Fatalf("RemoveSavedDevices should be no-op: %v", err)
	}
}

func TestConfigPath(t *testing.T) {
	withTempConfigDir(t)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if filepath.Base(path) != "devices.json" {
		t.Errorf("expected devices.json, got %s", filepath.Base(path))
	}
}

func TestConfigPath_Error(t *testing.T) {
	configDirFunc = func() (string, error) { return "", fmt.Errorf("no home dir") }
	t.Cleanup(func() { configDirFunc = os.UserConfigDir })

	_, err := ConfigPath()
	if err == nil {
		t.Error("expected error when configDirFunc fails")
	}

	// Verify error propagates through all functions
	if err := SaveDevices([]Device{{Name: "x"}}); err == nil {
		t.Error("SaveDevices should propagate ConfigPath error")
	}
	if _, err := LoadDevices(); err == nil {
		t.Error("LoadDevices should propagate ConfigPath error")
	}
	if err := RemoveSavedDevices(); err == nil {
		t.Error("RemoveSavedDevices should propagate ConfigPath error")
	}
	if _, err := FindSavedDevice("x", ""); err == nil {
		t.Error("FindSavedDevice should propagate ConfigPath error")
	}
}

func TestLoadDevices_CorruptJSON(t *testing.T) {
	withTempConfigDir(t)

	path, _ := ConfigPath()
	os.MkdirAll(filepath.Dir(path), 0o700)
	os.WriteFile(path, []byte("{not valid json!!!"), 0o600)

	_, err := LoadDevices()
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	if !strings.Contains(err.Error(), "parse devices") {
		t.Errorf("error should mention 'parse devices': %v", err)
	}
}

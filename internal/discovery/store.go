package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// configDirFunc is overridable in tests.
var configDirFunc = os.UserConfigDir

// ConfigPath returns the path to the saved devices file.
func ConfigPath() (string, error) {
	dir, err := configDirFunc()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, "castspeak", "devices.json"), nil
}

// SaveDevices writes devices to the config file, creating directories as needed.
func SaveDevices(devices []Device) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal devices: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write devices file: %w", err)
	}
	return nil
}

// LoadDevices reads saved devices from the config file.
// Returns an empty slice if the file does not exist.
func LoadDevices() ([]Device, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Device{}, nil
		}
		return nil, fmt.Errorf("read devices: %w", err)
	}
	var devices []Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}
	return devices, nil
}

// RemoveSavedDevices deletes the saved devices file. No-op if missing.
func RemoveSavedDevices() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove devices file: %w", err)
	}
	return nil
}

// FindSavedDevice searches saved devices by name or UUID.
func FindSavedDevice(name, uuid string) (Device, error) {
	devices, err := LoadDevices()
	if err != nil {
		return Device{}, err
	}
	for _, d := range devices {
		if (name != "" && strings.EqualFold(d.Name, name)) ||
			(uuid != "" && strings.EqualFold(d.UUID, uuid)) {
			return d, nil
		}
	}
	identifier := name
	if identifier == "" {
		identifier = uuid
	}
	return Device{}, fmt.Errorf("device not found in saved devices: %s", identifier)
}

package speak

import (
	"context"
	"fmt"
	"time"

	"castspeak/internal/cast"
	"castspeak/internal/discovery"
	"castspeak/internal/tts"
)

const DefaultDiscoveryTimeout = 5 * time.Second

// ListDevices returns all Cast devices found on the local network.
func ListDevices(ctx context.Context) ([]discovery.Device, error) {
	return discovery.Discover(ctx)
}

// Speak discovers a device by name or UUID, builds TTS URLs, and casts them.
// Returns the device name used and the number of chunks played.
func Speak(ctx context.Context, text, deviceName, deviceUUID, language string) (string, int, error) {
	if text == "" {
		return "", 0, fmt.Errorf("text is required")
	}
	if language == "" {
		language = "en"
	}

	urls, err := tts.BuildURLs(text, language)
	if err != nil {
		return "", 0, fmt.Errorf("tts: %w", err)
	}

	device, err := findDevice(ctx, deviceName, deviceUUID)
	if err != nil {
		return "", 0, err
	}

	if err := cast.PlayURLs(device.Addr, device.Port, urls); err != nil {
		return "", 0, fmt.Errorf("cast: %w", err)
	}

	return device.Name, len(urls), nil
}

// findDevice is a helper that resolves a device by name or UUID.
func findDevice(ctx context.Context, deviceName, deviceUUID string) (discovery.Device, error) {
	if deviceName == "" && deviceUUID == "" {
		return discovery.Device{}, fmt.Errorf("device_name or device_uuid is required")
	}
	return discovery.FindDevice(ctx, deviceName, deviceUUID)
}

// SetVolume discovers a device and sets its volume (0.0–1.0).
func SetVolume(ctx context.Context, deviceName, deviceUUID string, level float32) error {
	device, err := findDevice(ctx, deviceName, deviceUUID)
	if err != nil {
		return err
	}
	return cast.SetVolume(device.Addr, device.Port, level)
}

// SetMuted discovers a device and sets its mute state.
func SetMuted(ctx context.Context, deviceName, deviceUUID string, muted bool) error {
	device, err := findDevice(ctx, deviceName, deviceUUID)
	if err != nil {
		return err
	}
	return cast.SetMuted(device.Addr, device.Port, muted)
}

// Stop discovers a device and stops any active media.
func Stop(ctx context.Context, deviceName, deviceUUID string) error {
	device, err := findDevice(ctx, deviceName, deviceUUID)
	if err != nil {
		return err
	}
	return cast.Stop(device.Addr, device.Port)
}

// Status discovers a device and returns its current status.
func Status(ctx context.Context, deviceName, deviceUUID string) (string, cast.DeviceStatus, error) {
	device, err := findDevice(ctx, deviceName, deviceUUID)
	if err != nil {
		return "", cast.DeviceStatus{}, err
	}
	status, err := cast.Status(device.Addr, device.Port)
	if err != nil {
		return "", cast.DeviceStatus{}, err
	}
	return device.Name, status, nil
}

// PlayURL discovers a device and plays an arbitrary media URL.
func PlayURL(ctx context.Context, deviceName, deviceUUID, url string) error {
	device, err := findDevice(ctx, deviceName, deviceUUID)
	if err != nil {
		return err
	}
	return cast.PlayURLs(device.Addr, device.Port, []string{url})
}

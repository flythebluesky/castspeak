package speak

import (
	"context"
	"fmt"
	"log"
	"time"

	"castspeak/internal/cast"
	"castspeak/internal/discovery"
	"castspeak/internal/scan"
	"castspeak/internal/tts"
)

// Device is an alias for discovery.Device, exposed so callers don't import discovery directly.
type Device = discovery.Device

const DefaultDiscoveryTimeout = 5 * time.Second

// ListDevices returns all Cast devices found on the local network.
// Falls back to saved devices if mDNS discovery fails or returns nothing.
func ListDevices(ctx context.Context) ([]discovery.Device, error) {
	devices, err := discovery.Discover(ctx)
	if err == nil && len(devices) > 0 {
		return devices, nil
	}

	saved, savedErr := discovery.LoadDevices()
	if savedErr != nil {
		log.Printf("warning: failed to load saved devices: %v", savedErr)
	} else if len(saved) > 0 {
		if err != nil {
			log.Printf("mDNS discovery failed (%v); falling back to %d saved device(s)", err, len(saved))
		} else {
			log.Printf("mDNS returned no devices; falling back to %d saved device(s)", len(saved))
		}
		return saved, nil
	}

	if err != nil {
		return nil, err
	}
	return devices, nil
}

// Speak discovers a device by name, UUID, or host address, builds TTS URLs, and casts them.
// Returns the device name used and the number of chunks played.
func Speak(ctx context.Context, text, deviceName, deviceUUID, host, language string) (string, int, error) {
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

	device, err := findDevice(ctx, deviceName, deviceUUID, host)
	if err != nil {
		return "", 0, err
	}

	if err := cast.PlayURLs(device.Addr, device.Port, urls); err != nil {
		return "", 0, fmt.Errorf("cast: %w", err)
	}

	return device.Name, len(urls), nil
}

// findDevice resolves a device by host, name, or UUID.
// If host is provided, mDNS discovery is skipped entirely.
func findDevice(ctx context.Context, deviceName, deviceUUID, host string) (discovery.Device, error) {
	if host != "" {
		return discovery.DeviceFromHost(host)
	}
	if deviceName == "" && deviceUUID == "" {
		return discovery.Device{}, fmt.Errorf("device_name, device_uuid, or host is required")
	}
	dev, err := discovery.FindDevice(ctx, deviceName, deviceUUID)
	if err == nil {
		return dev, nil
	}

	// Fallback to saved devices when mDNS fails
	saved, savedErr := discovery.FindSavedDevice(deviceName, deviceUUID)
	if savedErr == nil {
		log.Printf("WARNING: mDNS discovery failed (%v); using saved device %q at %s — address may be stale", err, saved.Name, saved.Addr)
		return saved, nil
	}
	return discovery.Device{}, err
}

// SetVolume discovers a device and sets its volume (0.0–1.0).
func SetVolume(ctx context.Context, deviceName, deviceUUID, host string, level float32) error {
	device, err := findDevice(ctx, deviceName, deviceUUID, host)
	if err != nil {
		return err
	}
	return cast.SetVolume(device.Addr, device.Port, level)
}

// SetMuted discovers a device and sets its mute state.
func SetMuted(ctx context.Context, deviceName, deviceUUID, host string, muted bool) error {
	device, err := findDevice(ctx, deviceName, deviceUUID, host)
	if err != nil {
		return err
	}
	return cast.SetMuted(device.Addr, device.Port, muted)
}

// Stop discovers a device and stops any active media.
func Stop(ctx context.Context, deviceName, deviceUUID, host string) error {
	device, err := findDevice(ctx, deviceName, deviceUUID, host)
	if err != nil {
		return err
	}
	return cast.Stop(device.Addr, device.Port)
}

// Status discovers a device and returns its current status.
func Status(ctx context.Context, deviceName, deviceUUID, host string) (string, cast.DeviceStatus, error) {
	device, err := findDevice(ctx, deviceName, deviceUUID, host)
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
func PlayURL(ctx context.Context, deviceName, deviceUUID, host, url string) error {
	device, err := findDevice(ctx, deviceName, deviceUUID, host)
	if err != nil {
		return err
	}
	return cast.PlayURLs(device.Addr, device.Port, []string{url})
}

// ScanDevices scans local subnets for Cast devices via TCP port scan + HTTP metadata.
func ScanDevices(ctx context.Context) ([]discovery.Device, error) {
	return scan.ScanAndIdentify(ctx)
}

// SaveDevices writes devices to the config file.
func SaveDevices(devices []discovery.Device) error {
	return discovery.SaveDevices(devices)
}

// LoadSavedDevices reads saved devices from the config file.
func LoadSavedDevices() ([]discovery.Device, error) {
	return discovery.LoadDevices()
}

// ForgetDevices deletes the saved devices file.
func ForgetDevices() error {
	return discovery.RemoveSavedDevices()
}

// SavedDevicesPath returns the path to the saved devices file.
func SavedDevicesPath() (string, error) {
	return discovery.ConfigPath()
}

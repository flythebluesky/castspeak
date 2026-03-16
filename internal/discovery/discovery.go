package discovery

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/vishen/go-chromecast/dns"
)

// Device holds discovered Cast device information.
type Device struct {
	Name  string
	UUID  string
	Addr  string
	Port  int
	Model string
}

// Discover returns all Cast devices found on the local network within the
// given context's deadline.
func Discover(ctx context.Context) ([]Device, error) {
	entryChan, err := dns.DiscoverCastDNSEntries(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("mDNS discovery failed: %w", err)
	}

	var devices []Device
	for entry := range entryChan {
		devices = append(devices, deviceFromEntry(entry))
	}
	return devices, nil
}

// FindDevice discovers devices and returns the first one matching by name or UUID.
func FindDevice(ctx context.Context, name, uuid string) (Device, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	entryChan, err := dns.DiscoverCastDNSEntries(ctx, nil)
	if err != nil {
		return Device{}, fmt.Errorf("mDNS discovery failed: %w", err)
	}

	for entry := range entryChan {
		if (name != "" && entry.GetName() == name) ||
			(uuid != "" && entry.GetUUID() == uuid) {
			return deviceFromEntry(entry), nil
		}
	}

	identifier := name
	if identifier == "" {
		identifier = uuid
	}
	if ctx.Err() != nil {
		return Device{}, fmt.Errorf("discovery timed out before finding device %q: %w", identifier, ctx.Err())
	}
	return Device{}, fmt.Errorf("device not found: %s", identifier)
}

// HostPort returns the device address as host:port for dialing.
// Handles IPv6 addresses that may already be bracketed by go-chromecast.
func (d Device) HostPort() string {
	addr := strings.TrimRight(strings.TrimLeft(d.Addr, "["), "]")
	return net.JoinHostPort(addr, strconv.Itoa(d.Port))
}

func deviceFromEntry(e dns.CastEntry) Device {
	return Device{
		Name:  e.GetName(),
		UUID:  e.GetUUID(),
		Addr:  e.GetAddr(),
		Port:  e.GetPort(),
		Model: e.Device,
	}
}

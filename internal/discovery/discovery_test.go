package discovery

import (
	"net"
	"testing"

	"github.com/vishen/go-chromecast/dns"
)

func TestDevice_HostPort(t *testing.T) {
	tests := []struct {
		name string
		dev  Device
		want string
	}{
		{
			name: "ipv4",
			dev:  Device{Addr: "192.168.1.42", Port: 8009},
			want: "192.168.1.42:8009",
		},
		{
			name: "ipv6 already bracketed by go-chromecast",
			dev:  Device{Addr: "[::1]", Port: 8009},
			want: "[::1]:8009",
		},
		{
			name: "different port",
			dev:  Device{Addr: "10.0.0.1", Port: 9000},
			want: "10.0.0.1:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dev.HostPort()
			if got != tt.want {
				t.Errorf("HostPort() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeviceFromHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		wantAddr string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "ip with port",
			host:     "192.168.86.248:8009",
			wantAddr: "192.168.86.248",
			wantPort: 8009,
		},
		{
			name:     "ip without port defaults to 8009",
			host:     "192.168.86.248",
			wantAddr: "192.168.86.248",
			wantPort: 8009,
		},
		{
			name:     "custom port",
			host:     "10.0.0.1:9000",
			wantAddr: "10.0.0.1",
			wantPort: 9000,
		},
		{
			name:     "hostname with port",
			host:     "mydevice.local:8009",
			wantAddr: "mydevice.local",
			wantPort: 8009,
		},
		{
			name:     "hostname without port",
			host:     "mydevice.local",
			wantAddr: "mydevice.local",
			wantPort: 8009,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev, err := DeviceFromHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeviceFromHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if dev.Addr != tt.wantAddr {
				t.Errorf("Addr = %q, want %q", dev.Addr, tt.wantAddr)
			}
			if dev.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", dev.Port, tt.wantPort)
			}
		})
	}
}

func TestDeviceFromEntry(t *testing.T) {
	entry := dns.CastEntry{
		AddrV4:     net.ParseIP("192.168.1.100"),
		Port:       8009,
		DeviceName: "Kitchen speaker",
		UUID:       "uuid-123",
		Device:     "Google Nest Mini",
	}

	dev := deviceFromEntry(entry)

	if dev.Name != "Kitchen speaker" {
		t.Errorf("Name = %q, want %q", dev.Name, "Kitchen speaker")
	}
	if dev.UUID != "uuid-123" {
		t.Errorf("UUID = %q, want %q", dev.UUID, "uuid-123")
	}
	if dev.Addr != "192.168.1.100" {
		t.Errorf("Addr = %q, want %q", dev.Addr, "192.168.1.100")
	}
	if dev.Port != 8009 {
		t.Errorf("Port = %d, want %d", dev.Port, 8009)
	}
	if dev.Model != "Google Nest Mini" {
		t.Errorf("Model = %q, want %q", dev.Model, "Google Nest Mini")
	}
}

func TestDeviceFromEntry_IPv6(t *testing.T) {
	entry := dns.CastEntry{
		AddrV6:     net.ParseIP("::1"),
		Port:       8009,
		DeviceName: "Office",
		UUID:       "uuid-456",
		Device:     "Chromecast",
	}

	dev := deviceFromEntry(entry)

	// With no IPv4, GetAddr() returns the IPv6 wrapped in brackets.
	if dev.Addr != "[::1]" {
		t.Errorf("Addr = %q, want %q", dev.Addr, "[::1]")
	}
}

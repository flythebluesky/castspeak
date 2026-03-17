package scan

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestGetLocalSubnets(t *testing.T) {
	subnets, err := GetLocalSubnets()
	if err != nil {
		t.Fatalf("GetLocalSubnets: %v", err)
	}
	if len(subnets) == 0 {
		t.Fatal("expected at least one subnet")
	}
	for _, s := range subnets {
		ip4 := s.IP.To4()
		if ip4 == nil {
			t.Errorf("expected IPv4 subnet, got %s", s.IP)
		}
		if ip4[0] == 127 {
			t.Errorf("should exclude loopback, got %s", s.IP)
		}
		if ip4[0] == 169 && ip4[1] == 254 {
			t.Errorf("should exclude link-local, got %s", s.IP)
		}
	}
}

func TestSubnetIPs(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("192.168.1.0/24")
	ips := SubnetIPs(cidr)
	if len(ips) != 254 {
		t.Errorf("got %d IPs, want 254", len(ips))
	}
	if ips[0].String() != "192.168.1.1" {
		t.Errorf("first IP = %s, want 192.168.1.1", ips[0])
	}
	if ips[len(ips)-1].String() != "192.168.1.254" {
		t.Errorf("last IP = %s, want 192.168.1.254", ips[len(ips)-1])
	}
}

func TestSubnetIPs_Small(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("10.0.0.0/30")
	ips := SubnetIPs(cidr)
	if len(ips) != 2 {
		t.Errorf("got %d IPs, want 2", len(ips))
	}
}

func TestFetchDeviceInfo(t *testing.T) {
	info := map[string]interface{}{
		"name":     "Kitchen speaker",
		"ssdp_udn": "uuid:abc-123",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/setup/eureka_info" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(info)
	}))
	defer srv.Close()

	// Point infoPort at the test server's port
	_, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	testPort, _ := strconv.Atoi(portStr)
	origPort := infoPort
	infoPort = testPort
	defer func() { infoPort = origPort }()

	dev, err := FetchDeviceInfo("127.0.0.1")
	if err != nil {
		t.Fatalf("FetchDeviceInfo: %v", err)
	}
	if dev.Name != "Kitchen speaker" {
		t.Errorf("Name = %q, want Kitchen speaker", dev.Name)
	}
	if dev.UUID != "abc-123" {
		t.Errorf("UUID = %q, want abc-123 (uuid: prefix should be stripped)", dev.UUID)
	}
	if dev.Port != castPort {
		t.Errorf("Port = %d, want %d", dev.Port, castPort)
	}
}

func TestFetchDeviceInfo_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	testPort, _ := strconv.Atoi(portStr)
	origPort := infoPort
	infoPort = testPort
	defer func() { infoPort = origPort }()

	_, err := FetchDeviceInfo("127.0.0.1")
	if err == nil {
		t.Error("expected error for non-200 response")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error should mention status code: %v", err)
	}
}

func TestScanSubnet_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Already cancelled

	// Use TEST-NET-1 (RFC 5737) to avoid real network connections
	_, cidr, _ := net.ParseCIDR("192.0.2.0/30")
	_, err := ScanSubnet(ctx, cidr, 8009)
	if err == nil {
		t.Error("expected context error")
	}
}

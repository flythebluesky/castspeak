package scan

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"castspeak/internal/discovery"
)

const (
	castPort    = 8009
	dialTimeout = 500 * time.Millisecond
	httpTimeout = 3 * time.Second
	workers     = 64
)

// infoPort is the HTTP port for eureka_info. Variable for test overrides.
var infoPort = 8008

// GetLocalSubnets returns IPv4 subnets from non-loopback, up interfaces.
func GetLocalSubnets() ([]net.IPNet, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	var subnets []net.IPNet
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("warning: skipping interface %s: %v", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}
			// Skip link-local (169.254.x.x)
			if ip4[0] == 169 && ip4[1] == 254 {
				continue
			}
			subnets = append(subnets, *ipNet)
		}
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("no usable IPv4 subnets found")
	}
	return subnets, nil
}

// SubnetIPs returns all host IPs in a subnet (excluding network and broadcast).
func SubnetIPs(subnet *net.IPNet) []net.IP {
	ip4 := subnet.IP.To4()
	mask := subnet.Mask
	if len(mask) == 16 {
		mask = mask[12:]
	}
	if ip4 == nil || len(mask) != 4 {
		return nil
	}

	networkInt := binary.BigEndian.Uint32(ip4)
	maskInt := binary.BigEndian.Uint32(mask)
	broadcastInt := networkInt | ^maskInt
	networkAddr := networkInt & maskInt

	var ips []net.IP
	for i := networkAddr + 1; i < broadcastInt; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		ips = append(ips, ip)
	}
	return ips
}

// ScanSubnet probes all IPs in a subnet on the given TCP port.
// Returns IPs that accepted a connection.
func ScanSubnet(ctx context.Context, subnet *net.IPNet, port int) ([]string, error) {
	ips := SubnetIPs(subnet)
	if len(ips) == 0 {
		return nil, nil
	}

	ch := make(chan net.IP, len(ips))
	for _, ip := range ips {
		ch <- ip
	}
	close(ch)

	var mu sync.Mutex
	var found []string
	var wg sync.WaitGroup

	for i := 0; i < workers && i < len(ips); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ch {
				if ctx.Err() != nil {
					return
				}
				addr := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
				conn, err := net.DialTimeout("tcp", addr, dialTimeout)
				if err != nil {
					continue
				}
				conn.Close()
				mu.Lock()
				found = append(found, ip.String())
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return found, ctx.Err()
}

// eurekaInfo is the subset of fields we need from /setup/eureka_info.
type eurekaInfo struct {
	Name    string `json:"name"`
	SsdpUdn string `json:"ssdp_udn"`
}

// FetchDeviceInfo retrieves Cast device metadata from the eureka_info endpoint.
func FetchDeviceInfo(ip string) (discovery.Device, error) {
	url := fmt.Sprintf("http://%s:%d/setup/eureka_info", ip, infoPort)
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return discovery.Device{}, fmt.Errorf("fetch eureka_info from %s: %w", ip, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return discovery.Device{}, fmt.Errorf("eureka_info %s returned %d", ip, resp.StatusCode)
	}

	var info eurekaInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return discovery.Device{}, fmt.Errorf("parse eureka_info from %s: %w", ip, err)
	}

	uuid := strings.TrimPrefix(info.SsdpUdn, "uuid:")

	return discovery.Device{
		Name: info.Name,
		UUID: uuid,
		Addr: ip,
		Port: castPort,
	}, nil
}

// ScanAndIdentify scans local subnets for Cast devices and fetches their info.
func ScanAndIdentify(ctx context.Context) ([]discovery.Device, error) {
	subnets, err := GetLocalSubnets()
	if err != nil {
		return nil, err
	}

	var allIPs []string
	for i := range subnets {
		found, err := ScanSubnet(ctx, &subnets[i], castPort)
		allIPs = append(allIPs, found...)
		if err != nil {
			break // context cancelled; proceed with whatever was found
		}
	}

	var devices []discovery.Device
	for _, ip := range allIPs {
		if ctx.Err() != nil {
			return devices, ctx.Err()
		}
		dev, err := FetchDeviceInfo(ip)
		if err != nil {
			log.Printf("skipping %s (port %d open but info fetch failed: %v)", ip, castPort, err)
			continue
		}
		devices = append(devices, dev)
	}
	return devices, nil
}

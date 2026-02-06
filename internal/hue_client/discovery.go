package hueclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/brutella/dnssd"
	log "github.com/sirupsen/logrus"
)

type DiscoveredBridge struct {
	IP   string
	ID   string
	Name string
}

type BridgeConfig struct {
	Name             string  `json:"name"`
	SwVersion        string  `json:"swversion"`
	APIVersion       string  `json:"apiversion"`
	MAC              string  `json:"mac"`
	BridgeID         string  `json:"bridgeid"`
	FactoryNew       bool    `json:"factorynew"`
	ReplacesBridgeID *string `json:"replacesbridgeid"`
	ModelID          string  `json:"modelid"`
}

type DiscoverBridgeResult struct {
	ID                string `json:"id"`
	InternalIPAddress string `json:"internalipaddress"`
	MACAddress        string `json:"macaddress"`
	Name              string `json:"name"`
}

type BridgeDiscoveryService struct {
	logger *log.Entry
}

func NewBridgeDiscoveryService(logger *log.Entry) *BridgeDiscoveryService {
	return &BridgeDiscoveryService{
		logger: logger.WithField("component", "BridgeDiscoveryService"),
	}
}

// DiscoverFirstBridge tries to discover a single Hue Bridge on the local network.
func (d *BridgeDiscoveryService) DiscoverFirstBridge(logger *log.Entry) (*DiscoveredBridge, error) {
	bridges, err := d.DiscoverBridges()
	if err != nil {
		return nil, fmt.Errorf("failed to discover bridge: %w", err)
	}

	if len(bridges) == 0 {
		return nil, fmt.Errorf("no Hue Bridges found")
	}

	return bridges[0], nil
}

func (d *BridgeDiscoveryService) DiscoverBridges() ([]*DiscoveredBridge, error) {
	bridgeIp, err := d.FindHueBridgeBymDNS()
	if err != nil {
		// Falling back to discover.meethue.com endpoint
		return d.fetchBridgesFromDiscoverEndpoint()
	}

	if bridgeIp == "" {
		return nil, fmt.Errorf("failed to discover bridge with mDNS discovery: %w", err)
	}

	config, err := d.fetchBridgeConfigByIP(bridgeIp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config for discovered bridge \"%s\": %w", bridgeIp, err)
	}

	return []*DiscoveredBridge{{
		IP:   bridgeIp,
		ID:   config.BridgeID,
		Name: config.Name,
	}}, nil
}

func (d *BridgeDiscoveryService) FindHueBridgeBymDNS() (string, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*15))
	defer cancel()

	addrChan := make(chan []net.IP)

	addFn := func(e dnssd.BrowseEntry) {
		addrChan <- e.IPs
	}

	rmvFn := func(e dnssd.BrowseEntry) {
	}

	// Discover Hue bridges via mDNS/DNS-SD

	service := "_hue._tcp.local."
	go func() {
		if err := dnssd.LookupType(ctx, service, addFn, rmvFn); err != nil {
			if err != context.Canceled {
				fmt.Println("Error during mDNS lookup:", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Discovery timeout")
		return "", fmt.Errorf("discovery timeout")
	case ips := <-addrChan:
		if len(ips) > 0 {
			for _, ip := range ips {
				if ip.To4() != nil {
					return ip.String(), nil
				}
			}
		}
	}

	return "", nil
}

func (d *BridgeDiscoveryService) fetchBridgesFromDiscoverEndpoint() ([]*DiscoveredBridge, error) {
	bridges, err := d.fetchBridgesByDiscoverEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to discover bridges via discover endpoint: %w", err)
	}
	var discoveredBridges []*DiscoveredBridge
	for _, b := range bridges {
		discoveredBridges = append(discoveredBridges, &DiscoveredBridge{
			IP:   b.InternalIPAddress,
			ID:   b.ID,
			Name: b.Name,
		})
	}
	return discoveredBridges, nil
}

func (d *BridgeDiscoveryService) fetchBridgesByDiscoverEndpoint() ([]*DiscoverBridgeResult, error) {

	resp, err := http.Get("https://discovery.meethue.com")
	if err != nil {
		return nil, fmt.Errorf("failed to discover bridge: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request failed with status code: %v", resp.StatusCode)
	}

	var result []*DiscoverBridgeResult

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode discovery response: %w", err)
	}

	return result, nil
}

func (d *BridgeDiscoveryService) fetchBridgeConfigByIP(bridgeIP string) (*BridgeConfig, error) {
	url := fmt.Sprintf("http://%s/api/0/config", bridgeIP)
	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to get bridge config: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bridge config request failed with status code: %d", resp.StatusCode)
	}

	var config BridgeConfig

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode bridge config response: %v", err)
	}

	return &config, nil
}

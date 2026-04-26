package cloud

import "strings"

// DefaultBeaconInfraAPIURL is the production BeaconInfra API base URL.
// It can be overridden at build time for forks/self-hosted builds:
//
//	go build -ldflags "-X github.com/Bajusz15/beacon-vpn/internal/cloud.DefaultBeaconInfraAPIURL=https://example.com/api"
var DefaultBeaconInfraAPIURL = "https://beaconinfra.dev/api"

func BeaconInfraAPIBase() string {
	s := strings.TrimSpace(DefaultBeaconInfraAPIURL)
	if s == "" {
		return "https://beaconinfra.dev/api"
	}
	return strings.TrimSuffix(s, "/")
}

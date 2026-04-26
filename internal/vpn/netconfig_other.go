//go:build !linux && !darwin && !windows

package vpn

import "errors"

func applyClientNetwork(ifName, vpnAddress string) error {
	return errors.New("VPN is only supported on Linux, macOS, and Windows")
}

func teardownNetwork(ifName string) {}

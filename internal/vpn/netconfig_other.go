//go:build !linux && !darwin

package vpn

import "errors"

func applyClientNetwork(ifName, vpnAddress string) error {
	return errors.New("VPN is only supported on Linux and macOS")
}

func teardownNetwork(ifName string) {}

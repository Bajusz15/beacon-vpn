//go:build windows

package vpn

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func applyClientNetwork(ifName, vpnAddress string) error {
	ip := net.ParseIP(strings.TrimSpace(vpnAddress))
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("invalid IPv4 VPN address %q", vpnAddress)
	}
	if !commandExists("netsh") {
		return fmt.Errorf("netsh command not found")
	}
	if err := runCmd("netsh", "interface", "ipv4", "set", "address", "name="+ifName, "static", ip.String(), "255.255.255.0"); err != nil {
		return fmt.Errorf("netsh set address: %w", err)
	}
	if err := runCmd("netsh", "interface", "ipv4", "set", "subinterface", ifName, "mtu=1420", "store=active"); err != nil {
		return fmt.Errorf("netsh set mtu: %w", err)
	}
	return nil
}

func teardownNetwork(ifName string) {
	_ = runCmd("netsh", "interface", "ipv4", "delete", "address", "name="+ifName, "addr=all")
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

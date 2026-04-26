//go:build darwin

package vpn

import (
	"fmt"
	"os/exec"
	"strings"
)

func applyClientNetwork(ifName, vpnAddress string) error {
	if !commandExists("ifconfig") {
		return fmt.Errorf("ifconfig not found")
	}
	if err := runCmd("ifconfig", ifName, "inet", vpnAddress, vpnAddress, "netmask", "255.255.255.0", "up"); err != nil {
		return fmt.Errorf("ifconfig: %w", err)
	}
	return nil
}

func teardownNetwork(ifName string) {
	_ = runCmd("ifconfig", ifName, "down")
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

//go:build linux

package vpn

import (
	"fmt"
	"os/exec"
	"strings"
)

func applyClientNetwork(ifName, vpnAddress string) error {
	if !commandExists("ip") {
		return fmt.Errorf("`ip` command not found; install iproute2")
	}
	if err := runCmd("ip", "addr", "add", vpnAddress+"/24", "dev", ifName); err != nil {
		if !strings.Contains(err.Error(), "File exists") {
			return fmt.Errorf("ip addr add: %w", err)
		}
	}
	if err := runCmd("ip", "link", "set", "dev", ifName, "up"); err != nil {
		return fmt.Errorf("ip link up: %w", err)
	}
	return nil
}

func teardownNetwork(ifName string) {
	_ = runCmd("ip", "link", "set", "dev", ifName, "down")
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

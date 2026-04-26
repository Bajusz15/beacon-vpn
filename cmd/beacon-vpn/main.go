package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Bajusz15/beacon-vpn/internal/cloud"
	"github.com/Bajusz15/beacon-vpn/internal/config"
	"github.com/Bajusz15/beacon-vpn/internal/vpn"
	"github.com/spf13/cobra"
)

var (
	apiBase    string
	apiKeyFlag string
	nameFlag   string
)

func main() {
	root := &cobra.Command{
		Use:   "beacon-vpn",
		Short: "Lightweight Beacon VPN client",
		Long:  "Beacon VPN connects this computer to a Beacon-managed exit node using WireGuard.",
	}
	root.PersistentFlags().StringVar(&apiBase, "api-base", cloud.BeaconInfraAPIBase(), "BeaconInfra API base URL")
	root.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Beacon API key (or BEACON_API_KEY)")
	root.PersistentFlags().StringVar(&nameFlag, "name", "", "local client device name (defaults to hostname)")
	_ = root.PersistentFlags().MarkHidden("api-base")

	root.AddCommand(loginCommand(), logoutCommand(), connectCommand(), enableExitNodeCommand(), statusCommand(), disconnectCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Save Beacon API credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.EnsureAuth(apiKeyFlag, nameFlag)
			if err != nil {
				return err
			}
			path, _ := config.Path()
			fmt.Printf("Saved credentials for %q in %s\n", cfg.DeviceName, path)
			return nil
		},
	}
}

func logoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved Beacon API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ClearAuth(); err != nil {
				return err
			}
			fmt.Println("Removed saved Beacon API key.")
			return nil
		},
	}
}

func connectCommand() *cobra.Command {
	var listenPort int
	cmd := &cobra.Command{
		Use:     "connect <exit-device>",
		Aliases: []string{"use"},
		Short:   "Connect to a Beacon exit node",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.EnsureAuth(apiKeyFlag, nameFlag)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			client := cloud.NewVPNClient(apiBase, cfg.APIKey, cfg.DeviceName)
			session, err := vpn.StartClient(ctx, client, args[0], listenPort)
			if err != nil {
				return err
			}
			defer func() {
				closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := session.Close(closeCtx); err != nil {
					fmt.Fprintf(os.Stderr, "disconnect cleanup failed: %v\n", err)
				}
			}()

			st := session.Status()
			fmt.Printf("Connected to %q via %s\n", st.PeerDevice, st.InterfaceName)
			fmt.Printf("VPN address: %s\n", st.VPNAddress)
			fmt.Printf("Peer endpoint: %s\n", st.PeerEndpoint)
			fmt.Println("Press Ctrl+C to disconnect.")

			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					fmt.Println("Disconnecting...")
					return nil
				case <-ticker.C:
					printStatusLine(session.Status())
				}
			}
		},
	}
	cmd.Flags().IntVar(&listenPort, "listen-port", vpn.DefaultListenPort, "local WireGuard listen port")
	return cmd
}

func enableExitNodeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "enable-exit-node",
		Aliases: []string{"enable"},
		Short:   "Show how to enable a Beacon VPN exit node",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Exit-node mode is intentionally handled by the full beacon agent.")
			fmt.Println()
			fmt.Println("On the hosting machine, run:")
			fmt.Println("  beacon cloud login")
			fmt.Println("  beacon vpn enable")
			fmt.Println("  beacon master --foreground")
			fmt.Println()
			fmt.Println("Then connect from this client with:")
			fmt.Println("  beacon-vpn connect <exit-device>")
		},
	}
}

func statusCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show saved Beacon VPN client configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			kp, keyExists, keyErr := vpn.LoadPrivateKeyIfExists()
			if jsonOut {
				out := map[string]any{
					"configured":   strings.TrimSpace(cfg.APIKey) != "",
					"device_name":  cfg.DeviceName,
					"has_api_key":  strings.TrimSpace(cfg.APIKey) != "",
					"public_key":   "",
					"config_note":  "connect is foreground-only; live status is printed by the running connect process",
					"key_load_err": "",
				}
				if keyErr == nil && keyExists {
					out["public_key"] = kp.PublicKey
				} else if keyErr != nil {
					out["key_load_err"] = keyErr.Error()
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			path, _ := config.Path()
			fmt.Printf("Config: %s\n", path)
			fmt.Printf("Device: %s\n", valueOr(cfg.DeviceName, "(not set)"))
			fmt.Printf("API key: %s\n", yesNo(strings.TrimSpace(cfg.APIKey) != ""))
			if keyErr == nil && keyExists {
				fmt.Printf("Public key: %s\n", kp.PublicKey)
			} else if keyErr == nil {
				fmt.Println("Public key: not generated yet")
			} else {
				fmt.Printf("Public key: unavailable (%v)\n", keyErr)
			}
			fmt.Println("Live tunnel status is shown by `beacon-vpn connect` while it is running.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "print JSON")
	return cmd
}

func disconnectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect",
		Short: "Deregister this client from BeaconInfra and remove the local VPN key",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.EnsureAuth(apiKeyFlag, nameFlag)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client := cloud.NewVPNClient(apiBase, cfg.APIKey, cfg.DeviceName)
			if err := client.DeregisterVPN(ctx); err != nil {
				return err
			}
			if err := vpn.DeletePrivateKey(); err != nil {
				return err
			}
			fmt.Println("Disconnected and removed local VPN key.")
			return nil
		},
	}
}

func printStatusLine(st vpn.Status) {
	connected := "no handshake yet"
	if st.Connected {
		connected = "connected"
	}
	fmt.Printf("status: %s, rx=%s, tx=%s\n", connected, humanBytes(st.BytesRx), humanBytes(st.BytesTx))
}

func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for n/div >= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

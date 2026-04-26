package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Config is the tiny subset of ~/.beacon/config.yaml that beacon-vpn needs.
// The writer preserves unrelated top-level fields so it can coexist with beacon.
type Config struct {
	APIKey     string `yaml:"api_key,omitempty"`
	DeviceName string `yaml:"device_name,omitempty"`
}

func BeaconHomeDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("BEACON_HOME")); v != "" {
		return filepath.Abs(os.ExpandEnv(v))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".beacon"), nil
}

func Path() (string, error) {
	home, err := BeaconHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

func SaveAuth(apiKey, deviceName string) error {
	apiKey = strings.TrimSpace(apiKey)
	deviceName = strings.TrimSpace(deviceName)
	if apiKey == "" {
		return errors.New("api key is required")
	}
	if deviceName == "" {
		deviceName = DetectHostname()
	}
	if deviceName == "" {
		return errors.New("device name is required")
	}

	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("mkdir beacon home: %w", err)
	}

	raw := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse existing %s: %w", path, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing %s: %w", path, err)
	}

	raw["api_key"] = apiKey
	raw["device_name"] = deviceName

	data, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

func ClearAuth() error {
	path, err := Path()
	if err != nil {
		return err
	}
	raw := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse existing %s: %w", path, err)
		}
	} else if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read existing %s: %w", path, err)
	}

	delete(raw, "api_key")
	data, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("mkdir beacon home: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

func EnsureAuth(apiKeyFlag, nameFlag string) (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	apiKey := strings.TrimSpace(apiKeyFlag)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("BEACON_API_KEY"))
	}
	if apiKey == "" {
		apiKey = strings.TrimSpace(cfg.APIKey)
	}
	name := strings.TrimSpace(nameFlag)
	if name == "" {
		name = strings.TrimSpace(cfg.DeviceName)
	}
	if name == "" {
		name = DetectHostname()
	}
	if apiKey == "" {
		apiKey, err = PromptSecret("Beacon API key")
		if err != nil {
			return nil, err
		}
	}
	if err := SaveAuth(apiKey, name); err != nil {
		return nil, err
	}
	return &Config{APIKey: apiKey, DeviceName: name}, nil
}

func PromptSecret(label string) (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintf(os.Stderr, "%s: ", label)
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		v := strings.TrimSpace(string(raw))
		if v == "" {
			return "", fmt.Errorf("%s is required", strings.ToLower(label))
		}
		return v, nil
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "%s: ", label)
	v, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", fmt.Errorf("%s is required", strings.ToLower(label))
	}
	return v, nil
}

func DetectHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(h)
}

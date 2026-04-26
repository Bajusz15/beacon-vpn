package vpn

import "time"

type Role string

const (
	RoleClient Role = "client"
)

type PeerInfo struct {
	DeviceName string `json:"device_name"`
	PublicKey  string `json:"public_key"`
	Endpoint   string `json:"endpoint"`
	VPNAddress string `json:"vpn_address"`
	AllowedIPs string `json:"allowed_ips"`
}

type Status struct {
	Enabled       bool      `json:"enabled"`
	Role          Role      `json:"role"`
	InterfaceName string    `json:"interface_name,omitempty"`
	VPNAddress    string    `json:"vpn_address,omitempty"`
	ListenPort    int       `json:"listen_port,omitempty"`
	PublicKey     string    `json:"public_key,omitempty"`
	PeerDevice    string    `json:"peer_device,omitempty"`
	PeerEndpoint  string    `json:"peer_endpoint,omitempty"`
	Connected     bool      `json:"connected"`
	LastHandshake time.Time `json:"last_handshake,omitempty"`
	BytesRx       uint64    `json:"bytes_rx"`
	BytesTx       uint64    `json:"bytes_tx"`
	Error         string    `json:"error,omitempty"`
}

const DefaultListenPort = 51820

const InterfaceName = "beacon0"

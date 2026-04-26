package vpn

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type PeerResolver interface {
	RegisterVPN(ctx context.Context, publicKey, role string, listenPort int, endpoint string) (string, error)
	GetPeer(ctx context.Context, deviceName string) (*PeerInfo, error)
	DeregisterVPN(ctx context.Context) error
}

type Session struct {
	resolver PeerResolver
	wg       *wgDevice
	status   Status
}

func StartClient(ctx context.Context, resolver PeerResolver, peerDevice string, listenPort int) (*Session, error) {
	peerDevice = strings.TrimSpace(peerDevice)
	if peerDevice == "" {
		return nil, errors.New("peer device is required")
	}
	if resolver == nil {
		return nil, errors.New("peer resolver is required")
	}
	if listenPort <= 0 {
		listenPort = DefaultListenPort
	}

	kp, err := LoadOrCreatePrivateKey()
	if err != nil {
		return nil, err
	}
	addr, err := resolver.RegisterVPN(ctx, kp.PublicKey, string(RoleClient), listenPort, "")
	if err != nil {
		return nil, fmt.Errorf("register vpn with cloud: %w", err)
	}
	registered := true
	defer func() {
		if registered {
			_ = resolver.DeregisterVPN(context.Background())
		}
	}()

	peer, err := resolver.GetPeer(ctx, peerDevice)
	if err != nil {
		return nil, fmt.Errorf("fetch peer info: %w", err)
	}
	if peer.Endpoint == "" {
		return nil, fmt.Errorf("peer %q has no public endpoint yet; make sure the exit node is running `beacon vpn enable`", peerDevice)
	}
	allowedIPs := strings.TrimSpace(peer.AllowedIPs)
	if allowedIPs == "" {
		allowedIPs = strings.TrimSpace(peer.VPNAddress) + "/32"
	}
	if strings.TrimSpace(peer.PublicKey) == "" || strings.TrimSpace(peer.VPNAddress) == "" {
		return nil, fmt.Errorf("peer %q returned incomplete vpn metadata", peerDevice)
	}

	wg, err := createWGDevice(InterfaceName, kp.PrivateKey, listenPort)
	if err != nil {
		return nil, err
	}
	if err := applyClientNetwork(wg.name, addr); err != nil {
		wg.close()
		return nil, err
	}
	if err := wg.configurePeer(peer.PublicKey, peer.Endpoint, allowedIPs, 25*time.Second); err != nil {
		wg.close()
		teardownNetwork(wg.name)
		return nil, err
	}

	registered = false
	return &Session{
		resolver: resolver,
		wg:       wg,
		status: Status{
			Enabled:       true,
			Role:          RoleClient,
			InterfaceName: wg.name,
			VPNAddress:    addr,
			ListenPort:    listenPort,
			PublicKey:     kp.PublicKey,
			PeerDevice:    peerDevice,
			PeerEndpoint:  peer.Endpoint,
		},
	}, nil
}

func (s *Session) Status() Status {
	if s == nil {
		return Status{}
	}
	st := s.status
	if s.wg != nil {
		if rx, tx, lh, err := s.wg.stats(); err == nil {
			st.BytesRx = rx
			st.BytesTx = tx
			st.LastHandshake = lh
			st.Connected = !lh.IsZero() && time.Since(lh) < 3*time.Minute
		} else {
			st.Error = err.Error()
		}
	}
	return st
}

func (s *Session) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.wg != nil {
		teardownNetwork(s.wg.name)
		s.wg.close()
		s.wg = nil
	}
	if s.resolver != nil {
		if ctx == nil {
			ctx = context.Background()
		}
		return s.resolver.DeregisterVPN(ctx)
	}
	return nil
}

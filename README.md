# Beacon VPN

Beacon VPN client for laptops and desktops to connect to Beacon agent managed computers.

`beacon-vpn` connects this computer to a Beacon-managed exit node using
WireGuard. The exit node still runs the full `beacon` agent; this repo is only
for client machines that do not need the full Beacon binary.

VPN traffic flows directly between your devices. BeaconInfra is only used for
API-key auth, public-key registration, and peer endpoint lookup.

## Install

```bash
go install github.com/Bajusz15/beacon-vpn/cmd/beacon-vpn@latest
```

Tagged releases build standalone binaries for macOS and Linux on AMD64/ARM64.

## Usage

Save your Beacon API key once:

```bash
beacon-vpn login
```

Remove the saved API key:

```bash
beacon-vpn logout
```

Connect to an exit node by its Beacon device name:

```bash
beacon-vpn connect my-home-pi
```

`connect` runs in the foreground and keeps the tunnel alive until you press
Ctrl+C. It deregisters the client on shutdown.

Show saved local configuration:

```bash
beacon-vpn status
```

Force deregistration and remove the local VPN key:

```bash
beacon-vpn disconnect
```

`beacon-vpn enable-exit-node` prints the full-agent commands for hosting
machines. Exit-node mode intentionally remains in the main `beacon` binary.

## Config

The client reuses Beacon's config path:

```yaml
api_key: usr_xxxxxxxxxxxxxxxx
device_name: mate-macbook
```

The file lives at `~/.beacon/config.yaml` by default. Set `BEACON_HOME` to use a
different directory.

## Exit Node

On the hosting machine, keep using the full Beacon agent:

```bash
beacon cloud login
beacon vpn enable
beacon master --foreground
```

The exit node must have a reachable WireGuard UDP port in the current direct
connection phase.

# Architecture

`beacon-vpn` is a client-only companion to the main Beacon agent.

## Boundary

Hosting machines run the full `beacon` binary. They own exit-node behavior,
heartbeats, project monitoring, remote command dispatch, and local dashboards.

Client machines run `beacon-vpn`. This repo owns only:

- API-key login/logout for `~/.beacon/config.yaml`
- BeaconInfra VPN register, peer lookup, and deregister calls
- local WireGuard key generation and encrypted storage
- client-side userspace WireGuard tunnel setup

## Code Reuse Policy

The first version intentionally copies the small VPN client contracts from
`github.com/Bajusz15/beacon` instead of importing the main repo. Most source
packages in `beacon` are under `internal/`, and exposing them now would create a
public API before the client UX has stabilized.

If duplication becomes painful, the likely extraction point is a tiny stable
package for BeaconInfra VPN API types and client calls. The WireGuard lifecycle
should stay local to this repo unless both binaries need the exact same runtime
behavior.

# Phase 0: Xray gRPC API Audit

**Date:** 2026-05-24
**Server:** 31.56.208.169 (Debian 12, x86_64, Xray v26.3.27)

## Test Setup

Xray v26.3.27 installed via official XTLS install script. Config:

```json
{
  "api": {
    "services": ["HandlerService", "RoutingService"],
    "tag": "api"
  },
  "inbounds": [{
    "tag": "api",
    "listen": "127.0.0.1",
    "port": 10085,
    "protocol": "dokodemo-door",
    "settings": { "address": "127.0.0.1" }
  }],
  "outbounds": [{ "protocol": "freedom", "tag": "direct" }]
}
```

## Results

| Test | Method | Result | Details |
|------|--------|--------|---------|
| Port binding | `ss -tlnp` | ✅ | Port 10085 listening on 127.0.0.1 |
| TCP acceptance | Xray logs | ✅ | Connections accepted: `[api >> direct]` |
| gRPC handshake | `grpcurl -plaintext localhost:10085 list` | ❌ | context deadline exceeded |
| gRPC (Go client) | Go integration test via SSH tunnel | ❌ | context deadline exceeded |
| gRPC AddInbound | Protobuf HandlerService | ❌ | Not reachable |
| gRPC RoutingService | Protobuf RoutingService | ❌ | Not reachable |

## Root Cause

Xray v26.3.27 routes API inbound traffic to the `direct` outbound (`[api >> direct]` in logs) instead of handling it internally via the API service dispatcher. The `dokodemo-door` protocol forwards TCP connections to the configured address rather than passing them to Xray's internal gRPC handler.

This behavior differs from earlier Xray versions where the same config was documented to work.

## Config File Approach: VERIFIED

| Test | Method | Result | Details |
|------|--------|--------|---------|
| Config write + test | `xray run -test` | ✅ | Configuration OK |
| VLESS TCP inbound | config.json → systemctl start | ✅ | Port 443 active |
| Reality config | config.json with Reality settings | ❌ | v26 requires new format (`serverName` instead of `serverNames`, additional fields) |
| SIGHUP reload | `pkill -HUP xray` | ❌ | Xray stopped instead of reloading with Reality config error |

Config file approach WORKS for simple configs. Reality format needs updating for v26.

## v26 Config Format Changes

| Old (v1.8) | New (v26) |
|------------|-----------|
| `realitySettings.serverNames: ["discord.com"]` | `realitySettings.serverName: "discord.com"` |
| `realitySettings.shortIds: ["6ba85179"]` | `realitySettings.shortId: "6ba85179"` |
| No `password` required | `realitySettings.password: ""` (new required field) |

## Decision

**gRPC API: NOT USABLE for v1.**
**Config file: PRIMARY method, with v26 format adjustments needed for Reality.**

LucX will use **config.json read/write** as the PRIMARY and ONLY configuration method for Xray v1:
- `AddInbound` → read config.json → append inbound → write config.json → SIGHUP Xray
- `RemoveInbound` → read → filter → write → SIGHUP
- `AddOutbound` / `RemoveOutbound` → same pattern
- `SetRouting` → read → modify routing section → write → SIGHUP

## Implications

1. **Xray restart required** — config changes require SIGHUP (graceful reload). Active connections may be interrupted.
2. **No atomic operations** — cannot add inbound + outbound + routing in a single gRPC call. Each change is a separate config write.
3. **Transaction rollback works differently** — instead of RemoveInbound gRPC call, rollback writes the pre-change config back to disk.
4. **Config snapshots become critical** — before every chain apply, the config.json must be backed up so rollback can restore it.

## Future

- Monitor Xray releases for gRPC API fixes
- If gRPC becomes stable, add as a parallel strategy (try gRPC first, fallback to config)
- The `ProxyBackend` interface already supports this — `grpc.go` methods use config.json, future `grpc_v2.go` can use actual gRPC

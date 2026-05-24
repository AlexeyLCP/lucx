# LucX

**Personal Cross-Platform Multi-Hop Proxy Orchestrator**

LucX is a personal tool for visually constructing multi-hop proxy chains and generating client configurations. It installs and configures Xray-core directly on servers via SSH.

---

## ⚠️ Licensing — READ CAREFULLY

**LucX is intended exclusively for personal, non-commercial, educational, and research use.**

**Any commercial use (including VPN resale, VPN-as-a-service, paid proxy/VPN hosting, managed services, bundling with commercial products) is STRICTLY PROHIBITED without explicit written permission from the author.**

This is not negotiable. If you want to use LucX commercially — ask first.

Full license: [LICENSE](LICENSE) · PolyForm Noncommercial 1.0.0

---

## What LucX Does

- **Visual Chain Builder** — drag & drop servers into multi-hop proxy chains (Entry → Hop → Exit)
- **Automatic Xray Installation** — SSH into servers, auto-install Xray-core, configure systemd
- **Client Config Generation** — one click to get a ready-to-use `vless://` link
- **Topology Map** — visual overview of all servers and their connections
- **Transactional Chain Apply** — all-or-nothing with automatic rollback on failure
- **Pre-flight Safety Scan** — detects existing services before install, never overwrites without confirmation

## Supported Targets (v1)

| Target | Core Deployment | Proxy Backend |
|--------|----------------|---------------|
| Desktop (Linux/macOS/Windows) | ✅ | Xray-core |
| Server (VPS/Dedicated) | ✅ | Xray-core |
| Router (OpenWrt) | ✅ | Xray-core |
| Router (Keenetic) | ✅ | Xray-core |

Future: AmneziaWG, Sing-box, Hysteria2, TUIC

## Architecture

```
Flutter Client (Desktop / Mobile / Web)
        │ REST + WebSocket
LucX Core (Go binary, ~12MB)
        │ SSH + gRPC (or config.json)
Xray-core on each server
```

## Quick Start

```bash
# Build Core
make build

# Run Core
./build/lucx-core --listen :8744 --db ./lucx.db

# Cross-compile for routers
make cross
```

## Development

**Phase 0 (current):** Verify Xray gRPC API compatibility → `docs/api-audit.md`

```bash
make test       # Run tests (CGO_ENABLED=0)
make vet        # Lint
make cross      # Cross-compile all targets
```

## Disclaimer

This project is a personal research and educational tool. The author assumes no responsibility for any use made of this software by third parties. Users are solely responsible for complying with all applicable laws.

---

© 2026 LucX Project · [PolyForm Noncommercial 1.0.0](LICENSE)

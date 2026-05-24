# LucX — Design Specification

**Status:** Approved
**Date:** 2026-05-24
**Project:** LucX — Personal Cross-Platform Multi-Hop Proxy Orchestrator

## 1. Overview

LucX is a personal tool for visually constructing multi-hop proxy chains and generating client configurations. It installs and configures Xray-core directly on servers via SSH. No multi-user system, no billing, no traffic accounting — just chain building and config export for the owner.

### Key Principles

- **Personal tool** — single owner, no user/account management.
- **Zero damage** — never overwrite existing services without explicit confirmation. Detect and import existing Xray installations.
- **Protocol-agnostic** — architecture abstracts proxy backends behind a single interface. v1 = Xray, v2 = AWG, Hysteria2, Sing-box, TUIC.
- **Transactions** — chain application is atomic. Pre-flight → execute → commit or full rollback.
- **gRPC-first** — Xray configuration via gRPC HandlerService. Config file as fallback only.

## 2. Architecture Decision

**Chosen: Go Backend Core + Flutter Client**

### Why Go Core (even for a personal tool)

A 3-hop chain requires 6-8 sequential operations across 3 servers (inbounds, outbounds, routing on each). If executed from a Flutter client and the network drops mid-chain, servers are left in an inconsistent state — some configured, some not, no way to rollback.

Go Core provides:
- **Transactional orchestration** — all-or-nothing chain application with rollback.
- **Single binary** — deploy anywhere: laptop, server, OpenWrt router, Keenetic.
- **On-demand SSH** — connect only during operations or manual refresh. No persistent connections.
- **Pre-flight safety checks** — detect existing services, prevent damage, import existing configs.

### Deployment Modes

LucX Core runs in three modes, selected by CLI flag and resource auto-detection:

| Mode | Target | RAM | CPU | Storage | SSH | Monitor |
|------|--------|-----|-----|---------|-----|---------|
| **Desktop** | Laptop/PC | 50+ MB | Any | Any | On-demand | Every 2h |
| **Server** | VPS/Dedicated | 30+ MB | Any | 20 MB | On-demand | Every 2h |
| **Router** | OpenWrt / Keenetic | 15 MB target | MIPS/ARM, slow | 8 MB binary | On-demand | Disabled |

#### Resource Strategy

1. **On-demand SSH** — SSH connections are NEVER persistent. Connect → execute operation → disconnect. No idle connections consuming memory.

2. **Rare background refresh** — server status is checked every 2-4 hours (configurable). No constant polling. Default: 3h.

3. **Full refresh on client connect** — when Flutter client opens, Core does a full scan of all servers (status, inbounds, outbounds). This is the primary refresh trigger.

4. **No traffic monitoring** — Core does not continuously poll traffic stats. Traffic data is fetched only when user explicitly requests it in the UI.

#### Router Mode Optimizations

1. **Pure-Go SQLite** — `modernc.org/sqlite` instead of `mattn/go-sqlite3`. No CGo = trivial cross-compilation to MIPS/ARM. No libc dependency = works on musl (OpenWrt).

2. **Reduced binary size** — `-ldflags="-s -w"` + UPX compression. Target: <8 MB.

3. **Memory cap** — configurable Go GC (`GOGC=50`, `GOMEMLIMIT=32MiB`).

4. **Headless by design** — Core is a REST+WebSocket server. No GUI dependency.

5. **No background monitor** — in Router mode, monitor is completely disabled. Status only on explicit refresh.

6. **Config file preferred over gRPC** — on routers, avoid gRPC dependency if Xray gRPC port is not accessible. Fallback to config.json write + SIGHUP.

7. **Comfortable on 128-512 MB RAM** — Core process targets <20 MB RSS. Remaining RAM for OS and Xray.

#### Build Targets

```bash
GOOS=linux GOARCH=amd64   # Servers, desktops, WSL
GOOS=linux GOARCH=arm64   # ARM routers (RPi, newer OpenWrt)
GOOS=linux GOARCH=mipsle  # MIPS routers (older OpenWrt, Keenetic)
GOOS=linux GOARCH=arm GOARM=7  # ARMv7 routers
GOOS=darwin GOARCH=amd64  # macOS (dev/testing)
GOOS=windows GOARCH=amd64 # Windows (dev/testing)
```

CI produces 5 Linux binaries per release. All compiled with CGO_ENABLED=0 (pure Go SQLite).

## 3. System Architecture

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Desktop  │  │  Mobile  │  │   Web    │   Flutter (single codebase)
│(Win/Mac/ │  │(iOS/And) │  │  (PWA)   │   + local SQLite cache
│  Linux)  │  │          │  │          │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │
     └─────────────┼─────────────┘
                   │ REST (JWT) + WebSocket
              ┌────┴────┐
              │ LucX    │   Go Binary (~12MB desktop, ~8MB router)
              │ Core    │   Pure-Go SQLite (modernc.org/sqlite)
              └────┬────┘
                   │ SSH + gRPC (or config file)
     ┌─────────────┼─────────────────────┐
     │             │                     │
┌────┴────┐  ┌────┴────┐  ┌──────┐  ┌──────────┐
│ Server  │  │ Server  │  │Router│  │  Router   │
│(Entry)  │  │ (Exit)  │  │OpenWrt│  │ Keenetic  │
│ Xray    │  │ Xray    │  │ Xray  │  │  Xray     │
└─────────┘  └─────────┘  └──────┘  └──────────┘
```

**Core itself can run on any of these nodes** — including the router. Typically: Core on laptop (Desktop mode) or Core on router (Router mode, headless, Flutter connects remotely).

## 4. Protocol Abstraction Layer

### ProxyBackend Interface

```go
// internal/backend/interface.go
type ProxyBackend interface {
    // Identity
    Type() BackendType  // "xray", "awg", "sing-box", "hysteria2", "tuic"

    // Lifecycle (via SSH)
    Install(ctx, ssh) (string, error)       // download binary, create systemd unit, return binary path
    Start(ctx, ssh) error
    Stop(ctx, ssh) error
    Status(ctx, ssh) (BackendStatus, error) // running/stopped/error + version + PID

    // Configuration — backend chooses gRPC, REST, or config-file internally
    AddInbound(ctx, ssh, InboundSpec) (InboundResult, error)
    RemoveInbound(ctx, ssh, tag string) error
    AddOutbound(ctx, ssh, OutboundSpec) (OutboundResult, error)
    RemoveOutbound(ctx, ssh, tag string) error
    SetRouting(ctx, ssh, rules []RoutingRule) error
    GetConfig(ctx, ssh) (RawConfig, error)  // full current config for import/scan

    // Client config export
    BuildClientConfig(ctx, ssh, inboundTag string) (string, error)
    // Returns vless://... or equivalent for the owner
}
```

### XrayBackend (v1): gRPC-first with fallback

```
┌─────────────────────────────┐
│        XrayBackend          │
│                             │
│  ┌───────────────────────┐  │
│  │ XrayGRPC (priority)   │  │  HandlerService.AddInbound
│  │ gRPC HandlerService   │──┤  .AddOutbound .RemoveInbound
│  │ (no restart needed)   │  │
│  └───────────────────────┘  │
│             │               │
│             │ fallback       │
│  ┌───────────────────────┐  │
│  │ XrayConfigFile         │  │  read config.json → modify →
│  │ (requires restart)     │  │  write → restart Xray
│  └───────────────────────┘  │
└─────────────────────────────┘
```

Auto-detect at Install: try gRPC → if unavailable, use config file. gRPC port configured in systemd unit (--format=json on :10085 by default).

### Plugin Registry

```go
// internal/backend/registry.go
var Backends = map[BackendType]func() ProxyBackend{
    "xray":      func() ProxyBackend { return &xray.XrayBackend{} },       // v1
    "awg":       func() ProxyBackend { return &awg.AWGBackend{} },         // v2
    "sing-box":  func() ProxyBackend { return &singbox.SingBoxBackend{} }, // v2
    "hysteria2": func() ProxyBackend { return &hysteria2.H2Backend{} },    // v2
    "tuic":      func() ProxyBackend { return &tuic.TUICBackend{} },       // v2
}
```

Chain Engine never imports a concrete backend. It only uses `ProxyBackend` interface.

## 5. Server Safety — Pre-Install Scan & Import

### Pre-Install Safety Check

Before installing Xray on a new server, LucX MUST scan for existing services:

```
PreInstallCheck(ssh):
  1. Port scan: check common proxy ports (443, 8443, 10085, etc.)
  2. Service scan: detect systemd units matching *xray*, *3x-ui*, *x-ui*, *sing-box*, *awg*
  3. Process scan: detect running binaries (xray, sing-box, amneziawg, hysteria)
  4. Binary scan: check /usr/local/bin/, /opt/ for known proxy binaries
  5. Config scan: check /usr/local/etc/xray/, /etc/sing-box/, /etc/amnezia/
```

### Decision Matrix

| Detection | Action |
|-----------|--------|
| Nothing found | Safe to install Xray |
| Known: standalone Xray (no 3x-UI) | **Import mode** — read config.json, import inbounds/outbounds/routing into LucX model. Do NOT overwrite. |
| Known: Xray behind 3x-UI | Warn user. LucX cannot manage 3x-UI-managed Xray. Suggest manual migration. |
| Known: Other proxy (Sing-box, AWG) | Warn user. Mark for future import (v2). Do NOT install over it. |
| Unknown service on port 443/8443 | Warn user. Require explicit confirmation to proceed. |

### Import Flow (standalone Xray detected)

```
1. Detect: systemd unit "xray.service" exists, no "x-ui" service
2. Read: /usr/local/etc/xray/config.json via SSH
3. Parse: extract all inbounds, outbounds, routing rules
4. Present: show user what was found, ask to import
5. Import: create server entry in LucX DB, populate inbounds/outbounds as read-only snapshot
6. Mark: server status = 'imported', backend = 'xray', config_managed = true
```

After import, the server is "adopted" — LucX can now manage it (add/remove inbounds, etc.) without reinstalling.

## 6. Go Backend Core Structure

```
github.com/alexeylcp/lucx-core/
├── cmd/lucx-core/main.go           # entry point, flags, graceful shutdown
├── internal/
│   ├── api/                        # HTTP handlers (chi router) + WebSocket
│   │   ├── router.go               # chi setup, middleware chain
│   │   ├── auth.go                 # simple JWT (single user, local auth)
│   │   ├── servers.go              # CRUD servers, SSH install, import
│   │   ├── chains.go               # Chain CRUD + apply/validate/rollback
│   │   ├── map.go                  # Topology map endpoint
│   │   └── ws.go                   # WebSocket events (status, logs)
│   │
│   ├── backend/                    # Protocol abstraction layer
│   │   ├── interface.go            # ProxyBackend interface + shared types
│   │   ├── registry.go             # Backend registry map
│   │   ├── types.go                # InboundSpec, OutboundSpec, RoutingRule, BackendStatus
│   │   ├── xray/                   # Xray backend (v1)
│   │   │   ├── backend.go          # XrayBackend struct, implements ProxyBackend
│   │   │   ├── grpc.go             # gRPC HandlerService client
│   │   │   ├── configfile.go       # config.json reader/writer (fallback — used in Router mode)
│   │   │   ├── installer.go        # Download binary, create systemd/init.d unit
│   │   │   └── config_gen.go       # BuildClientConfig: vless:// link generation
│   │   ├── awg/                    # AWG backend (v2, stub in v1)
│   │   ├── singbox/                # Sing-box backend (v2, stub in v1)
│   │   ├── hysteria2/              # Hysteria2 backend (v2, stub in v1)
│   │   └── tuic/                   # TUIC backend (v2, stub in v1)
│   │
│   ├── chain/                      # Transactional chain engine
│   │   ├── engine.go               # Apply(chain) — orchestrates full flow
│   │   ├── planner.go              # Builds execution plan from chain graph
│   │   ├── executor.go             # Executes plan, collects created resources
│   │   ├── validator.go            # Pre-flight: SSH + gRPC + tag/port conflicts
│   │   └── rollback.go             # Undo created resources in reverse order
│   │
│   ├── scanner/                    # Server safety scanner
│   │   ├── preflight.go            # PreInstallCheck — ports, services, processes
│   │   ├── importer.go             # Import existing Xray config into LucX model
│   │   └── detector.go             # Detect: what proxy software is running?
│   │
│   ├── ssh/                        # SSH client pool (multi-distro)
│   │   ├── pool.go                 # Connection pool — size varies by mode
│   │   ├── client.go               # SSH client wrapper (run commands, read/write files)
│   │   ├── keys.go                 # SSH key management (encrypted storage + ssh-agent)
│   │   └── distro.go               # OS detection: Debian/Ubuntu/Alma/Arch/OpenWrt/Keenetic
│   │
│   ├── store/                      # Pure-Go SQLite data layer (no CGo)
│   │   ├── db.go                   # modernc.org/sqlite, WAL mode, migrations
│   │   ├── servers.go              # Server CRUD
│   │   └── chains.go               # Chain + chain_nodes CRUD
│   │
│   ├── mode/                       # Deployment mode management
│   │   └── mode.go                 # Desktop/Server/Router mode, resource limits, feature flags
│   │
│   └── config/                     # Core configuration
│       └── config.go               # CLI flags, env vars, config file
```

### Mode Package

```go
// internal/mode/mode.go
type RunMode string
const (
    ModeDesktop RunMode = "desktop" // full features, default
    ModeServer  RunMode = "server"  // headless, medium resources
    ModeRouter  RunMode = "router"  // headless, minimal resources
)

type ModeConfig struct {
    RunMode        RunMode
    MaxSSHPool     int   // desktop: 10, server: 5, router: 1
    MonitorEnabled bool  // desktop: true, server: true, router: false
    GOMEMLIMIT     string // desktop: "", server: "64MiB", router: "32MiB"
    GOGC           int    // desktop: 100, server: 50, router: 50
}

func DetectMode() RunMode {
    // Auto-detect: check /proc/cpuinfo, memory, etc.
    // Override via --mode flag
}
```

## 7. Data Models

### servers
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT PK | UUID |
| name | TEXT NOT NULL | "Finland-Helsinki" |
| host | TEXT NOT NULL | SSH host/IP |
| port | INTEGER DEFAULT 22 | SSH port |
| username | TEXT NOT NULL | SSH user (root or sudoer) |
| auth_method | TEXT NOT NULL | 'password' or 'key' |
| credential | TEXT NOT NULL | Encrypted: password or private key |
| os | TEXT | Detected: 'debian12', 'ubuntu24', etc. |
| arch | TEXT | 'amd64', 'arm64' |
| status | TEXT DEFAULT 'unknown' | 'online', 'offline', 'imported' |
| source | TEXT DEFAULT 'fresh' | 'fresh' (LucX installed), 'imported' (pre-existing Xray adopted) |
| tags | TEXT DEFAULT '[]' | JSON: ["eu", "1gbit", "reality"] |
| last_seen | DATETIME | |
| created_at | DATETIME DEFAULT NOW() | |

### server_backends
| Column | Type | Description |
|--------|------|-------------|
| server_id | TEXT FK→servers | |
| backend_type | TEXT NOT NULL | 'xray', 'awg', etc. |
| version | TEXT | "1.8.23" |
| status | TEXT DEFAULT 'stopped' | 'running', 'stopped', 'error' |
| config_path | TEXT | /usr/local/etc/xray/config.json |
| api_endpoint | TEXT | gRPC: "localhost:10085" |
| config_managed | BOOLEAN DEFAULT true | false = imported, don't auto-modify |
| installed_at | DATETIME | |
| PK | (server_id, backend_type) | |

### chains
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT PK | UUID |
| name | TEXT NOT NULL | "FI → NL → DE" |
| status | TEXT DEFAULT 'draft' | 'draft', 'active', 'broken', 'rolling_back' |
| applied_at | DATETIME | |
| created_at | DATETIME DEFAULT NOW() | |

### chain_nodes
| Column | Type | Description |
|--------|------|-------------|
| chain_id | TEXT FK→chains | |
| server_id | TEXT FK→servers | |
| backend_type | TEXT NOT NULL | 'xray' (v1), mixed per node in v2 (xray→awg→xray) |
| protocol | TEXT NOT NULL | 'vless', 'vmess', 'trojan', 'wireguard', etc. — per node |
| position | INTEGER NOT NULL | 0=entry, N=intermediate, last=exit |
| role | TEXT NOT NULL | 'entry', 'hop', 'exit' |
| inbound_spec | TEXT | JSON: InboundSpec (what to create) |
| outbound_spec | TEXT | JSON: OutboundSpec (route to next node, nil for exit) |
| inbound_result | TEXT | JSON: created inbound details (for rollback) |
| outbound_result | TEXT | JSON: created outbound details (for rollback) |
| PK | (chain_id, position) | |

### InboundSpec / OutboundSpec (Go types, serialized to JSON in chain_nodes)

```go
type InboundSpec struct {
    Tag      string // unique tag, e.g. "chain-xxx-entry"
    Protocol string // "vless", "vmess", "trojan"
    Port     int    // 443
    Listen   string // "0.0.0.0"
    Settings json.RawMessage
    StreamSettings *StreamSettings
}

type OutboundSpec struct {
    Tag      string // e.g. "chain-xxx-to-nl"
    Protocol string
    Settings json.RawMessage
    StreamSettings *StreamSettings
    SendThrough *string // outbound IP (optional)
}

type RoutingRule struct {
    Type        string // "field"
    InboundTag  []string
    OutboundTag string
}

type StreamSettings struct {
    Network  string // "tcp", "ws", "grpc", "h2", "quic"
    Security string // "none", "tls", "reality"
    // ... protocol-specific fields
}
```

## 8. Core REST API

Base: `/api/v1/`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/login | Simple password auth (single user) |
| CRUD | /servers | Manage servers |
| POST | /servers/:id/install | Install Xray via SSH (with pre-flight scan) |
| POST | /servers/:id/scan | Pre-flight scan: detect existing services |
| POST | /servers/:id/import | Import existing Xray config |
| GET | /servers/:id/status | Server + backend status |
| CRUD | /chains | Create/edit/delete chains |
| POST | /chains/:id/validate | Pre-flight validation only |
| POST | /chains/:id/apply | Apply chain (transactional) |
| POST | /chains/:id/rollback | Rollback an active chain |
| GET | /chains/:id/config | Get generated client config (vless:// link) |
| GET | /map | Topology map: servers + chains as graph JSON |
| WS | /ws/events | Real-time: apply progress, status changes |

## 9. Chain Transaction Flow

### Apply Chain: "Finland → Netherlands → Germany (Exit)"

```
1. PRE-FLIGHT (validator.go)
   ├─ Check SSH connectivity to all 3 servers
   ├─ Check backend Status() on all 3 (Xray running?)
   ├─ Check no existing inbounds with conflicting tags/ports
   └─ Build execution plan: ordered list of operations

2. EXECUTE (executor.go, sequential, stops on first error)
   ├─ Finland: AddInbound (VLESS+Reality, :443, tag="chain-xxx-entry")
   ├─ Finland: AddOutbound (VLESS, → Netherlands, tag="chain-xxx-to-nl")
   ├─ Finland: SetRouting (inbound:"chain-xxx-entry" → outbound:"chain-xxx-to-nl")
   ├─ Netherlands: AddInbound (VLESS, tag="chain-xxx-hop1")
   ├─ Netherlands: AddOutbound (VLESS, → Germany, tag="chain-xxx-to-de")
   ├─ Netherlands: SetRouting (inbound:"chain-xxx-hop1" → outbound:"chain-xxx-to-de")
   └─ Germany: AddInbound (VLESS+Reality, tag="chain-xxx-exit")

3. COMMIT
   ├─ Save chain + chain_nodes to SQLite
   ├─ Store created inbound/outbound details in chain_nodes (for rollback)
   └─ chain.status = 'active'

4. GENERATE CLIENT CONFIG
   └─ BuildClientConfig on Germany → vless://uuid@germany-ip:443?...

ON ERROR (e.g., Netherlands step fails):
   → ROLLBACK (reverse order):
     ├─ Finland: RemoveInbound("chain-xxx-entry")
     ├─ Finland: RemoveOutbound("chain-xxx-to-nl")
     └─ chain.status = 'draft'
   → All servers returned to pre-apply state
```

### Rollback Guarantee

Every Add* operation stores its result before the next operation. On failure, rollback iterates the collected results in reverse and calls the corresponding Remove*.

## 10. Key Screens

### 10.1 Spider Web — Dashboard

Live topology graph. Servers as nodes, chains as edges.
- Node size = number of chains through server
- Animated dots on edges = live traffic indication
- Colors: green=online, red=offline, orange=imported, blue=entry, gray=hop
- Stats: total servers, active chains
- Click node → server details. Click edge → chain details.

### 10.2 Chain Builder (Desktop)

Three-panel layout:
- **Left:** Server palette — drag sources, grouped by region/tag, status indicator
- **Center:** Canvas — drag servers, connect lines, ENTRY→HOP→EXIT labels, + button between nodes
- **Right:** Inspector — chain name, protocol per hop, transport, security. Validate + Apply buttons.

### 10.3 Chain Builder (Mobile)

Wizard mode (step-by-step):
1. Select Entry server → configure inbound (protocol, port, security)
2. Add Hop? (optional, repeatable) → select server, configure
3. Select Exit server → configure
4. Review summary → Apply

Drafts: chain auto-saved as `draft` on every wizard step. Resume from dashboard. Drafts persist in local SQLite.

### 10.4 Server Management

Card/table view. Status, SSH info, installed backends, tags.
Quick actions: Install Xray, Scan (pre-flight), Import config, Terminal (raw SSH).

### 10.5 Pre-Install Scan Dialog

Before installing Xray on a server, shows:
- Detected services/ports
- Safety assessment: green (safe), yellow (known, importable), red (conflict, needs user decision)
- Import button if standalone Xray detected

## 11. Security Model

- **SSH credentials:** encrypted at rest (AES-256-GCM, key derived from user password via Argon2id). Advanced users can opt into ssh-agent (no credential storage — LucX delegates auth to the system agent).
- **No credentials in logs:** SSH passwords/keys never written to logs.
- **JWT:** Simple token auth for Flutter ↔ Core communication. Single user, no roles.
- **TLS:** Core serves HTTPS. Self-signed cert auto-generated, user can provide custom cert.
- **SSH pool:** One persistent connection per server, reused. Timeout: 5min idle.

## 12. Platform Support

### Client Platforms (Flutter)

| Platform | v1 |
|----------|-----|
| Linux Desktop | ✓ Primary |
| macOS Desktop | ✓ |
| Windows Desktop | ✓ |
| Android | ✓ (wizard mode for chain builder) |
| iOS | ✓ (wizard mode for chain builder) |
| Web (PWA) | ✓ |

### Core Deployment Targets (Go — Linux only)

| Target | Arch | Use |
|--------|------|-----|
| linux/amd64 | x86_64 | Servers, desktops, WSL |
| linux/arm64 | ARM64 | ARM routers (RPi 4/5), newer OpenWrt, ARM VPS |
| linux/mipsle | MIPS32LE | Older OpenWrt routers, some Keenetic models |
| linux/arm (GOARM=7) | ARMv7 | Older ARM routers, older Keenetic models |

All builds: `CGO_ENABLED=0` (pure-Go SQLite, no libc dependency — compatible with musl/OpenWrt).

## 13. Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| gRPC HandlerService fails with Reality + uTLS + fallback configs | HIGH | Phase 0 tests this first. Fallback: config-file strategy for Reality chains. |
| gRPC requires Xray restart for certain settings (streamSettings changes) | MEDIUM | Phase 0 identifies which settings need restart. Document as known limitation. |
| Multi-distro SSH installer (apt/dnf/pacman, systemd/openrc, paths differ) | MEDIUM | Dedicated `distro.go` — detect OS/init/pkg-manager before install. Test on top 5 distros in MVP. |
| Xray version incompatibility (config format changes between 1.8.x versions) | LOW | Pin tested Xray version. Installer downloads specific version, not "latest". Auto-update in v2. |
| User has existing production Xray — LucX overwrites config | HIGH | Pre-flight scanner MUST detect before install. Import mode for standalone Xray. Never overwrite without explicit user approval. |
| SSH connection drops during chain apply | MEDIUM | Transaction engine: each step must succeed before next. Rollback on any failure. SSH pool with retry (3 attempts). |

## 14. Development Phases

### Phase 0 — API Verification (1-2 days)

**Critical path — must verify before any code:**

1. Deploy test Xray 1.8.x+ instance, enable gRPC HandlerService.
2. Test AddInbound with VLESS+Reality+uTLS fingerprint+fallback — verify it works without Xray restart.
3. Test AddOutbound with same parameters.
4. Test AddRoutingRule — verify routing is applied correctly.
5. Test RemoveInbound/RemoveOutbound — verify cleanup.
6. Test config.json fallback path: write complex config, restart Xray, verify handshake works.
7. Document findings in `docs/api-audit.md`.
8. If gRPC is flaky with Reality+uTLS: document exact limitations, plan fallback strategy.

**Go/No-Go:** If gRPC API cannot handle Reality+uTLS configurations, v1 will use config-file strategy exclusively for Reality chains.

### Phase 1 — MVP (4-6 weeks)

**Scope:**
- Go Core: basic HTTP server, SQLite, server CRUD, SSH client pool
- XrayBackend: gRPC config API, installer (download binary + systemd unit)
- Chain engine: single-protocol (VLESS+Reality), 2-hop max, transaction with rollback
- Pre-flight scanner: detect existing services, import standalone Xray config
- Flutter Desktop: server list, simple chain builder (2 nodes), config export as vless:// link
- No mobile wizard yet, no spider web, no routing rule management (implicit routing only)

**Deliverable:** Install Xray on 2 servers, create chain between them, get working vless:// link.

### Phase 2 — v1 (6-8 weeks)

**Scope:**
- Full chain builder: N hops, all Xray protocols (VLESS, VMess, Trojan, Shadowsocks)
- All transports (TCP, WS, gRPC, H2, QUIC) and security (Reality, TLS)
- Spider web topology dashboard
- Mobile wizard mode
- Routing rule editor (explicit routing between hops)
- Chain status monitoring (broken chain detection)
- Flutter Web (PWA)
- **Router Mode:** pure-Go SQLite, MIPS/ARM builds, UPX compression, resource limits
- **Headless mode:** Core runs without Flutter, manageable via REST API
- **OpenWrt/Keenetic:** init.d scripts, opkg/entware packaging
- **Multi-distro SSH installer:** OpenWrt (opkg), Keenetic (entware), in addition to apt/dnf/pacman

**Deliverable:** Full-featured personal orchestrator for Xray multi-hop chains, deployable on routers.

### Phase 3 — v2 (future)

**Scope:**
- AWG backend (AmneziaWG)
- Sing-box backend
- Hysteria2 backend
- TUIC backend
- Mixed-protocol chains (Xray → AWG → Xray)
- Subscription link import

## 15. Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend Core | Go 1.23+, chi router, modernc.org/sqlite (pure-Go, no CGo), golang-jwt, golang.org/x/crypto |
| Frontend (all platforms) | Flutter 3.x, drift (SQLite cache), riverpod |
| Xray config API | gRPC (HandlerService) → fallback: config.json write + SIGHUP |
| SSH | golang.org/x/crypto/ssh |
| Communication | REST (JSON) + WebSocket |
| Encryption | AES-256-GCM, Argon2id |
| Cross-compilation | CGO_ENABLED=0, GOOS=linux, GOARCH=amd64/arm64/mipsle/arm |
| Binary compression | UPX (router builds), -ldflags="-s -w" |

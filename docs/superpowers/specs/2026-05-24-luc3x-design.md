# Luc3X — Design Specification

**Status:** Approved
**Date:** 2026-05-24
**Project:** Luc3X — Cross-platform 3x-UI Orchestrator

## 1. Overview

Luc3X is a cross-platform application for orchestrating multiple 3x-UI (Xray-core) panels. It provides a visual chain builder for multi-hop proxy configurations, centralized monitoring, and batch user management — all via the official 3x-UI REST API.

### Key Principles

- **Zero Trust:** Panel credentials encrypted (AES-256-GCM). Master key held by user, never stored on backend.
- **Transactions:** Every chain operation is a transaction: pre-flight → execute → commit or rollback.
- **Offline-First:** Clients cache state locally. Work without Core connectivity, sync when reconnected.
- **API-Only:** All operations (except initial SSH install) via 3x-UI REST API. No direct config file manipulation.

## 2. Architecture Decision

**Chosen: Flutter Client + Go Backend Core**

### Why Go Backend

- **Transactional chains** — multi-hop setup requires 5+ sequential API calls across panels; Go server guarantees consistency with rollback on failure.
- **Single binary** — compiles to one ~10MB file, deploy alongside 3x-UI on any server.
- **24/7 monitoring** — background traffic collection, alerting (webhook/push) even when clients are offline.
- **Single source of truth** — all clients (mobile, desktop, web) stay synchronized.

### Why Flutter Client

- True cross-platform: Android, iOS, Windows, macOS, Linux, Web (PWA) from single codebase.
- Mature canvas/drawing support for topology graph and chain builder.
- SQLite via `drift` for local cache.

## 3. System Architecture

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Mobile  │  │ Desktop  │  │   Web    │   Flutter Clients
│ (iOS/And)│  │(Win/Mac/ │  │  (PWA)   │   + local SQLite cache
│          │  │  Linux)  │  │          │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │
     └─────────────┼─────────────┘
                   │ REST (JWT) + WebSocket
              ┌────┴────┐
              │ Luc3X   │   Go Binary (~10MB)
              │ Core    │   SQLite (WAL) + Encrypted Vault
              └────┬────┘
                   │ REST API (3x-UI panels)
     ┌─────────────┼─────────────┐
┌────┴────┐  ┌────┴────┐  ┌────┴────┐
│ Panel A │  │ Panel B │  │ Panel C │   3x-UI Instances
│ (Main)  │  │ (Hop)   │  │ (Exit)  │   Xray-core inside
└─────────┘  └─────────┘  └─────────┘
```

## 4. Go Backend Core Structure

```
github.com/alexeylcp/luc3x-core/
├── cmd/luc3x-core/main.go       # entry point, flags, graceful shutdown
├── internal/
│   ├── api/                     # HTTP handlers (chi) + WebSocket hub
│   │   ├── auth.go              # JWT issue/refresh/revoke
│   │   ├── panels.go            # CRUD panels, test connection
│   │   ├── chains.go            # Chain CRUD + apply/validate
│   │   ├── inbounds.go          # Proxy to panel APIs
│   │   ├── users.go             # Batch user operations
│   │   └── ws.go                # WebSocket events (traffic, alerts)
│   ├── panel/                   # 3x-UI API client
│   │   ├── client.go            # HTTP client, auth, retry, circuit breaker
│   │   ├── inbound.go           # /panel/api/inbound/*
│   │   ├── outbound.go          # /panel/api/outbound/* (when available)
│   │   ├── user.go              # /panel/api/inbound/client*
│   │   └── xray.go              # /panel/api/xray/* (status)
│   ├── chain/                   # Chain engine (core feature)
│   │   ├── engine.go            # Orchestrator: pre-flight→execute→commit/rollback
│   │   ├── planner.go           # Builds execution plan from chain graph
│   │   ├── executor.go          # Executes plan sequentially
│   │   ├── validator.go         # Pre-flight: connectivity, tag conflicts
│   │   └── rollback.go          # Undo created resources on failure
│   ├── sync/                    # Panel state sync (in-memory, not persisted)
│   │   ├── scanner.go           # Scans inbounds/outbounds/routing from panels
│   │   ├── diff.go              # Diffs current vs previous snapshot
│   │   ├── mapper.go            # Builds connection graph
│   │   └── cache.go             # In-memory TTL cache for users + snapshots
│   ├── monitor/                 # Background monitoring
│   │   ├── collector.go         # Periodic traffic collection
│   │   ├── alerter.go           # Rules: limit, expiration, offline
│   │   └── push.go              # Webhook / FCM push notifications
│   ├── vault/                   # Encrypted credential storage
│   │   └── vault.go             # AES-256-GCM, key derivation
│   ├── store/                   # Data layer (SQLite — only own state)
│   │   ├── db.go                # Connection, migrations, WAL mode
│   │   ├── panels.go            # Panel CRUD queries
│   │   ├── chains.go            # Chain + chain_nodes queries
│   │   └── audit.go             # Audit log queries
│   └── ssh/                     # Auto-install 3x-UI
│       └── installer.go         # SSH + curl install.sh + verify API
```

## 5. Data Models

**Persisted in SQLite (4 tables):** panels, chains, chain_nodes, audit_log — это собственные данные оркестратора.
**In-memory cache (TTL):** users, snapshots — источник правды это панели, перезапрашиваются при каждом подключении.

### panels
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT PK | UUID |
| name | TEXT NOT NULL | "Amsterdam-Main" |
| url | TEXT NOT NULL | "https://1.2.3.4:2053" |
| username | TEXT NOT NULL | Encrypted (vault) |
| password | TEXT NOT NULL | Encrypted (vault) |
| role | TEXT DEFAULT 'slave' | 'master' or 'slave' |
| status | TEXT DEFAULT 'unknown' | 'online', 'offline', 'degraded', 'unknown' |
| tags | TEXT DEFAULT '[]' | JSON array: ["eu", "reality", "10gbit"] |
| xray_version | TEXT | Xray version string |
| last_seen | DATETIME | Last successful API contact |
| created_at | DATETIME | DEFAULT NOW() |

### chains
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT PK | UUID |
| name | TEXT NOT NULL | "EU-Hop → Finland-Exit" |
| protocol | TEXT NOT NULL | 'vless', 'vmess', 'trojan', 'shadowsocks', 'hysteria', 'tuic' |
| transport | TEXT DEFAULT 'tcp' | 'tcp', 'ws', 'grpc', 'h2', 'quic' |
| security | TEXT DEFAULT 'reality' | 'none', 'tls', 'reality' |
| status | TEXT DEFAULT 'draft' | 'draft', 'active', 'broken', 'rolling_back' |
| applied_at | DATETIME | When the chain was last applied |
| created_at | DATETIME | DEFAULT NOW() |

### chain_nodes
| Column | Type | Description |
|--------|------|-------------|
| chain_id | TEXT FK→chains | |
| panel_id | TEXT FK→panels | |
| position | INTEGER | 0=entry, N=intermediate, last=exit |
| inbound_tag | TEXT | Tag of inbound created on this node |
| outbound_tag | TEXT | Tag of outbound routing to next node |
| config_snapshot | TEXT | JSON — what was created (for rollback) |
| PK | (chain_id, panel_id, position) | |

### users — in-memory cache (NOT persisted)

Fetched fresh from panels on every connection. Held in memory with TTL. Panels are the source of truth.

| Field | Type | Description |
|-------|------|-------------|
| id | string | "panel_id:inbound_id:email" |
| panel_id | string | |
| inbound_id | int | Inbound ID on the panel |
| email | string | User identifier |
| uuid | string | Xray UUID |
| traffic_up | int64 | |
| traffic_down | int64 | |
| traffic_limit | int64 | Bytes, 0 = unlimited |
| expire_at | time.Time | Zero = never |
| status | string | 'active', 'expired', 'disabled', 'over_limit' |
| chain_id | string | Bound chain, empty if unassigned |

### snapshots — in-memory cache (NOT persisted)

Full panel state (inbounds + outbounds + routing) fetched on-demand for map view. Re-fetched when user opens dashboard or triggers refresh. Used to diff and detect changes.

| Field | Type | Description |
|-------|------|-------------|
| panel_id | string | |
| data | json.RawMessage | { inbounds, outbounds, routing } |
| taken_at | time.Time | |

### audit_log
| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER PK | AUTOINCREMENT |
| action | TEXT NOT NULL | 'chain.apply', 'user.batch_create', etc. |
| target_type | TEXT | 'chain', 'user', 'panel' |
| target_id | TEXT | |
| details | TEXT | JSON with operation details |
| status | TEXT NOT NULL | 'success', 'failed', 'rolled_back' |
| error_msg | TEXT | Error if status=failed |
| created_at | DATETIME | DEFAULT NOW() |

## 6. Core REST API

Base: `/api/v1/`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/login | Local login to Core |
| POST | /auth/unlock | Unlock Vault with master key |
| CRUD | /panels | Manage panels (add/remove/test) |
| GET | /panels/:id/inbounds | Proxy: list panel inbounds |
| GET | /panels/:id/outbounds | Proxy: list panel outbounds |
| GET | /panels/:id/traffic | Panel traffic (up/down/total) |
| CRUD | /chains | Create/edit/delete chains |
| POST | /chains/:id/apply | Apply chain (transactional) |
| POST | /chains/:id/validate | Pre-flight validation only |
| POST | /users/batch | Batch user operations |
| GET | /map | Full connection map (graph JSON) |
| POST | /ssh/install | Auto-install 3x-UI via SSH |
| WS | /ws/events | Real-time event stream |

## 7. Key Screens

### 7.1 Spider Web — Dashboard (Main Screen)

Live network topology graph. All servers as nodes, all chains as edges.

- **Node size** = number of active chains through this server
- **Line thickness** = traffic volume
- **Animated dots** on lines = live traffic flow
- **Colors:** green glow = online, red glow = offline, orange border = master, dashed border = pending
- **Stats overlay:** total servers, chains, users, offline count
- **Legend:** active chain, intermediate hop, broken
- Client nodes shown on periphery

### 7.2 Chain Builder

Three-panel layout (desktop):

- **Left sidebar:** Server list with status indicators, drag source
- **Center canvas:** Graph editor. Drag servers onto canvas, connect with lines. ENTRY → HOP(s) → EXIT. Grid background, + buttons between nodes to insert hops.
- **Right inspector:** Chain name, protocol selector, transport, security. Validate + Apply buttons.

Mobile: card-based chain list with status indicators, FAB to create new chain. Chain editor is fullscreen wizard.

### 7.3 Connection Map

Auto-scanned graph of actual inbounds/outbounds/routing across all panels. Shows real (not designed) connections. Highlights:
- Broken chains (offline intermediate node)
- Orphaned inbounds/outbounds
- Conflicts (duplicate tags, port conflicts)

### 7.4 Server List / Management

Table/card view of all panels with:
- Status (online/offline/degraded/last seen)
- Role (master/slave)
- Tags for filtering
- Quick actions: test connection, view inbounds, open web panel

### 7.5 User Management

Batch operations across multiple chains/panels:
- Mass create/edit/delete users
- Assign users to chains
- Traffic monitoring, limits, expiration
- Bulk export/import (subscription links)

## 8. Chain Transaction Flow

When user clicks "Apply Chain" on a chain like `Finland → Netherlands → Germany (Exit)`:

```
1. PRE-FLIGHT
   - Check all 3 panels online
   - Validate no tag conflicts
   - Verify protocol/transport compatibility

2. EXECUTE (sequential, stops on first error)
   2a. Finland:   create Outbound → Netherlands (tag: "hop-to-nl")
   2b. Netherlands: create Inbound (tag: "from-fi")
                   create Outbound → Germany (tag: "hop-to-de")
                   create Routing rule: from-fi → hop-to-de
   2c. Germany:   create Inbound (tag: "exit-de")
                   create Client (user)

3. COMMIT
   - Save config_snapshot JSON to each chain_node (for rollback)
   - Update chain status = 'active'
   - Write audit log

ON ERROR (e.g., step 2b fails):
   → ROLLBACK: delete Outbound created in 2a from Finland
   → Update chain status = 'draft'
   → Write audit log with error_msg
```

## 9. Security Model

- **Vault:** Panel credentials encrypted with AES-256-GCM. Master key derived from user password via Argon2id.
- **Unlock flow:** Client sends master key → Core unlocks Vault → key held in memory only, never persisted.
- **JWT:** Short-lived access tokens (15min) + refresh tokens (7d). Revocable.
- **TLS:** Core serves HTTPS only (auto-generate self-signed or use provided cert).
- **Audit:** Every mutation logged to audit_log.
- **Minimal logging:** No credentials in logs, no request bodies in logs.

## 10. Platform Support

| Platform | Status | Notes |
|----------|--------|-------|
| Linux Desktop | v1 | Primary |
| macOS Desktop | v1 | |
| Windows Desktop | v1 | |
| Android | v1 | |
| iOS | v1 | |
| Web (PWA) | v1 | For server-side access |
| CLI | v2 | Automation/scripting companion |

## 11. Out of Scope for v1

- **AmneziaWG (AWG) support** — vanilla 3x-UI API only. AWG panel support in v2.
- **Telemt (MTProto)** — v2.
- **DPI obfuscation presets** — v2 (depends on AWG).
- **Subscription link import** — v1 supports URL+login+password; subscription parsing in v2.
- **Embedded Xray-core** — Luc3X orchestrates, does not proxy traffic itself.

## 12. Tech Stack Summary

| Layer | Technology |
|-------|-----------|
| Backend Core | Go 1.23+, chi router, mattn/go-sqlite3, golang-jwt |
| Frontend (all platforms) | Flutter 3.x, drift (SQLite), riverpod |
| Communication | REST (JSON) + WebSocket |
| Encryption | AES-256-GCM, Argon2id |
| SSH | golang.org/x/crypto/ssh |

# ForgeMate MVP Master Plan

## 1. Goal

Build a personal AI assistant service ("ForgeMate") inspired by OpenClaw, with:

- Gateway Daemon
- Telegram Channel Runtime Adapter (MVP channel)
- Plugin Runtime (full-surface, secure by default)
- Agent Runtime via Node sidecar (mariozechner stack)
- Routing + Session Keys
- WS-RPC + small HTTP control interface
- Admin UI + Admin CLI parity (for MVP action set)
- Vector memory in MVP

## 2. Locked Architecture Decisions

1. Core runtime: Go
2. Agent runtime: Node sidecar (mariozechner-based)
3. Core <-> sidecar transport: gRPC bidirectional stream over UDS
4. State source-of-truth: Go core
5. Tool execution owner: Go core
6. Data layer: file-backed state for MVP
7. Vector memory backend: SQLite + sqlite-vec
8. Provider auth:
   - OpenAI: BYOK + OAuth
   - Google: BYOK + OAuth
   - Anthropic: BYOK only
9. Applications scope: first-party/internal only
10. Connection model: WS-RPC first + small HTTP

## 3. Operational Policy Locks

### 3.1 Gateway lifecycle/reload

- Hybrid hot-apply + targeted restart
- Graceful drain 30s for in-flight runs
- Sidecar restart with exponential backoff + circuit breaker

### 3.2 OAuth lifecycle

- Token sink model (centralized auth profiles)
- Pre-refresh at 10 minutes before expiry
- Refresh failure -> degraded + cooldown + auto-failover

### 3.3 Plugin security boundary

- Manifest capabilities + runtime deny-by-default
- In-process + capability sandbox
- Global + per-plugin kill switch

### 3.4 Telegram operations

- Polling + webhook supported, single active mode per account
- Explicit mode switching only
- Webhook security: path secret + source validation + replay window
- Replay/dedupe: replay 5m + dedupe TTL 24h
- Outbound retry: exponential backoff + idempotency key + max 5 retries

### 3.5 Vector memory policy

- Embedding fallback: configured primary -> OpenAI -> Google -> disable
- Full reindex when provider/model/endpoint/chunking fingerprint changes
- Cost controls: per-agent cap + LRU eviction + monthly embedding budget

## 4. Public Interfaces / API Contracts

### 4.1 Sidecar gRPC service

- Health
- Run (bi-di stream)
- Abort
- Wait
- AuthProbe

Run stream frames include:

- Core -> Sidecar: RunStart, ToolResult, AbortSignal, Ping
- Sidecar -> Core: RunAccepted, AssistantDelta, ToolCallRequest, ToolProgress, RunCompleted, RunFailed, Pong

### 4.2 Gateway WS protocol

- connect-first handshake
- req/res/event/error frame model
- role-based auth: operator/app/node
- scope-based authorization (read/write/config/channels/sessions/plugins/cron/events)

### 4.3 Plugin manifest v1

- Required: id, version, entry, configSchema, capabilities
- Capability gates enforced at registration and call-time

## 5. Data Layout (MVP)

- ~/.forgemate/config/forgemate.json5
- ~/.forgemate/agents/<agentId>/agent/auth-profiles.json
- ~/.forgemate/agents/<agentId>/sessions/sessions.json
- ~/.forgemate/agents/<agentId>/sessions/*.jsonl
- ~/.forgemate/agents/<agentId>/memory/memory.sqlite

Security defaults:

- Sensitive files chmod 0600
- UDS socket owner-restricted

## 6. Epic Breakdown (Implementation)

### Epic A: Gateway Core Skeleton

Deliverables:

- Go daemon bootstrap
- config load/validate
- WS connect/health/status
- process lifecycle hooks

### Epic B: Sidecar Contract + Runtime Bridge

Deliverables:

- proto definitions
- UDS transport
- run stream, abort, wait
- sidecar supervision and breaker

### Epic C: Routing + Session Persistence

Deliverables:

- route precedence implementation
- deterministic session keys
- session store + transcript persistence + locking

### Epic D: Provider/Auth Lifecycle

Deliverables:

- OpenAI/Google/Anthropic adapters
- BYOK + OAuth flows (per matrix)
- token sink refresh/cooldown/failover

### Epic E: Telegram Adapter

Deliverables:

- polling + webhook mode support
- explicit mode switch orchestration
- dedupe/replay/retry policies

### Epic F: Plugin Runtime + Security Gates

Deliverables:

- discovery/load/register flow
- capability manifest enforcement
- kill switches + audit logging

### Epic G: Vector Memory MVP

Deliverables:

- SQLite + sqlite-vec indexing
- memory store/search APIs
- fallback/reindex/cost controls

### Epic H: Admin UI + CLI + App Access

Deliverables:

- MVP operations UI
- CLI parity for MVP action set
- role/scope token support for internal apps

### Epic I: Reliability + Verification

Deliverables:

- reload/drain behavior
- chaos tests (sidecar crash, token expiry, webhook replay)
- observability dashboards and alerts

## 7. Test Scenarios

1. Gateway:
   - connect-first protocol rejects invalid frames
   - hot-apply vs restart decisions are deterministic
   - graceful drain completes within timeout behavior

2. Sidecar:
   - stream ordering and seq dedupe
   - tool call/result idempotency
   - abort race safety (single terminal state)

3. Auth:
   - token pre-refresh path
   - refresh fail -> degraded/cooldown/failover
   - profile switching correctness

4. Telegram:
   - mode switch does not double-consume updates
   - replay rejection and 24h dedupe
   - retry respects idempotency

5. Plugins:
   - undeclared capability invocation denied
   - kill switch immediate effect
   - plugin crash containment behavior

6. Memory:
   - indexing works under caps
   - reindex trigger on fingerprint change
   - fallback provider behavior and disable path

7. End-to-end:
   - inbound telegram -> route -> sidecar run -> tool callback -> outbound reply
   - session continuity across restarts

## 8. Delivery Strategy and Worktree Plan

Recommended branches/worktrees:

1. worktree `forgemate-core`
   - Epic A + C + initial I

2. worktree `forgemate-sidecar`
   - Epic B + D

3. worktree `forgemate-integrations`
   - Epic E + F + G + H

Integration gates:

- Gate 1: Core + sidecar stream stable
- Gate 2: Telegram E2E with auth lifecycle
- Gate 3: Plugin + memory + UI/CLI parity
- Gate 4: Reliability hardening complete

## 9. Scope Feasibility

One Codex session:

- Feasible only for scaffold + one thin vertical slice.
- Not feasible for full MVP with required quality gates.

Recommended:

- Multi-session, worktree-based parallel implementation.
- Practical target: 8-12 focused implementation sessions, depending on test depth.

## 10. Assumptions and Defaults

1. Deployment target: single region, 1-3 nodes initially
2. Personal-first product, internal apps only in MVP
3. Go core owns control/state truth
4. Node sidecar remains execution engine for mariozechner runtime
5. Official auth mechanisms only; no unofficial token scraping flows

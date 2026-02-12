# OpenClaw Fork Design Reference (Personal Version)

Generated on: 2026-02-11  
Source repo restudied: `/Users/gregho/GitHub/AI/openclaw` @ `baa1e95b9`  
Base study read: `/Users/gregho/GitHub/AI/ForgeMate/studys/openclaw-codebase-deep-study.md`

## 1. Goal And Scope

This document is a **fork implementation reference** for building your own OpenClaw-inspired version with:
- high-level architecture clarity
- critical runtime details you should not accidentally break
- a practical phased implementation path

This is not a full clone plan. It is a "minimal stable core first" blueprint.

## 2. Critical Invariants You Should Keep

### 2.1 Single Gateway Ownership

Keep one long-lived Gateway process as the control plane and channel owner.

Why:
- avoids duplicate channel sessions and state drift
- centralizes auth, routing, events, approvals, and observability

References:
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/server.impl.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/server-runtime-state.ts`
- `/Users/gregho/GitHub/AI/openclaw/docs/gateway/index.md`

### 2.2 WS Handshake + Typed Protocol Discipline

Protocol constraints to preserve:
- first frame must be `connect`
- protocol version negotiation
- request/response/event frame separation
- strict frame/param validation (AJV/JSON schema)

References:
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/server/ws-connection/message-handler.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/protocol/index.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/protocol/schema.ts`

### 2.3 Auth And Scope Separation

Keep these layers separate:
- transport auth (`token` / `password` / tailscale identity)
- role (`operator` vs `node`)
- scope authorization per RPC method

References:
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/auth.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/server-methods.ts`

### 2.4 Deterministic Routing + Session Keys

Preserve deterministic route resolution and session key building.

Must-have behaviors:
- binding precedence (peer > parent peer > guild/team > account > channel default > global default)
- stable session key format for persistence/concurrency
- explicit DM scope policy (`main`, `per-peer`, etc.)

References:
- `/Users/gregho/GitHub/AI/openclaw/src/routing/resolve-route.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/routing/session-key.ts`

### 2.5 Config Pipeline Order

Config load order matters and should remain explicit:
1. read JSON5
2. resolve includes
3. apply `config.env` into process env
4. `${VAR}` substitution
5. schema + plugin-aware validation
6. defaults + path normalization
7. snapshot/hash/cache/hot-reload decisions

References:
- `/Users/gregho/GitHub/AI/openclaw/src/config/io.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/config/validation.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/config-reload.ts`

### 2.6 Plugin Runtime Safety Model

Keep manifest-first plugin loading discipline:
- discover candidates by precedence
- load manifest/schema first
- validate plugin config before register
- register into typed runtime surface

References:
- `/Users/gregho/GitHub/AI/openclaw/src/plugins/discovery.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/plugins/loader.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/plugins/runtime/index.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/plugin-sdk/index.ts`

### 2.7 Reply Dispatch Ordering

Preserve queued delivery ordering and separate tool/block/final channels.

Why:
- avoids out-of-order streamed responses
- keeps typing/human-delay behavior deterministic enough
- supports TTS/block-stream fallback logic cleanly

References:
- `/Users/gregho/GitHub/AI/openclaw/src/auto-reply/reply/reply-dispatcher.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/auto-reply/reply/dispatch-from-config.ts`

## 3. Fork Architecture (Recommended)

## 3.1 Minimal Core (MVP)

- `core/config`: config IO + validation + defaults
- `core/routing`: route resolver + session key
- `core/agent`: model selection + fallback + execution adapter
- `gateway/ws`: protocol, auth, method dispatch
- `gateway/runtime`: startup lifecycle + reload + shutdown
- `channels/<one channel>`: start with one provider only (Telegram or Web)
- `plugins`: minimal loader + register API (channel/tool/method only)

## 3.2 Deferred Layers (After MVP)

- multi-platform native apps (`apps/android`, `apps/ios`, `apps/macos` style)
- full extension ecosystem parity
- advanced discovery/tailscale/bonjour
- voice wake / sidecars / heavy media pipelines

## 4. Suggested Fork Folder Layout

```text
myclaw/
  src/
    entry.ts
    cli/
    config/
    gateway/
      protocol/
      ws/
      methods/
    routing/
    agents/
    channels/
      telegram/
    plugins/
    plugin-sdk/
    infra/
    logging/
  extensions/
    telegram/
  docs/
  test/
```

## 5. Startup Sequence Blueprint

1. normalize env + runtime guard
2. fast-route cheap CLI commands (optional)
3. load + validate config snapshot
4. auto-enable/resolve plugin set
5. build gateway runtime config (bind/auth/http flags)
6. create HTTP+WS runtime state
7. register methods + WS handlers
8. start channels/cron/hooks/plugin services
9. enable config watcher (hot-reload/restart)
10. install graceful shutdown handlers

References:
- `/Users/gregho/GitHub/AI/openclaw/src/entry.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/cli/run-main.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/cli/route.ts`
- `/Users/gregho/GitHub/AI/openclaw/src/gateway/server.impl.ts`

## 6. Implementation Phases

## Phase 1: Control Plane Skeleton

Deliver:
- CLI entry
- config load/validate
- WS connect + health/status RPC
- one in-memory session map

Exit criteria:
- `gateway start`
- `connect` handshake passes
- `health` returns stable JSON

## Phase 2: Routing + Agent Loop

Deliver:
- route resolver + session key
- model selection/fallback abstraction
- `agent` method with ack + final response

Exit criteria:
- deterministic session persistence key
- fallback attempt chain works

## Phase 3: One Real Channel

Deliver:
- one channel monitor/outbound adapter
- inbound normalization into shared context
- auto-reply dispatch with block/final queueing

Exit criteria:
- end-to-end inbound -> reply path works on real channel

## Phase 4: Plugin-First Extension Point

Deliver:
- plugin discovery + manifest validation
- register gateway methods/tools/channels
- plugin config schema validation

Exit criteria:
- install one sample extension and call its RPC

## Phase 5: Reliability + Security Hardening

Deliver:
- scope-based auth checks
- idempotency/dedupe for side effects
- reload/shutdown behavior + health metrics

Exit criteria:
- restart-safe operation
- no unauthorized method execution

## 7. Testing Strategy You Should Keep

Minimum gates per phase:
- unit: routing/session key/config validation
- integration: gateway handshake + method authorization + one channel path
- e2e: full inbound->agent->outbound flow

OpenClaw reference testing model:
- `/Users/gregho/GitHub/AI/openclaw/vitest.config.ts`
- `/Users/gregho/GitHub/AI/openclaw/scripts/test-parallel.mjs`
- `/Users/gregho/GitHub/AI/openclaw/docs/testing.md`

## 8. Practical Cut List (What To Not Fork Initially)

Do not copy initially:
- full channel matrix (Discord/Slack/Signal/iMessage/etc.)
- full native apps stack
- all bundled skills/plugins
- complete release automation/mac notarization pipeline

Focus instead on:
- a clean core with one production-quality channel
- strict protocol/auth/config correctness
- plugin boundary that lets you grow later

## 9. Final Recommendation

Build your fork as a **Gateway-first modular monolith**:
- keep protocol + config + routing + plugin boundaries strict
- aggressively reduce surface area in v1
- add channels/features only after core reliability is proven

This gives you OpenClaw-level extensibility without inheriting full repository complexity on day one.

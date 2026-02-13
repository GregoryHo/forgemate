# ForgeMate

ForgeMate is a personal AI assistant service inspired by OpenClaw.

## Current Scaffold (MVP Foundation)

- Go gateway daemon skeleton:
  - config loading (`FORGEMATE_*`)
  - file-backed state layout bootstrap
  - health/readiness HTTP endpoints
  - connect-first protocol validation endpoint
  - sidecar supervisor with backoff + breaker model
- Sidecar contract:
  - proto at `/proto/runtime/v1/agent_runtime.proto`
  - gRPC methods: `Health`, `Run`, `Abort`, `Wait`, `AuthProbe`
- Node sidecar skeleton:
  - gRPC runtime in `/sidecar/src/server.mjs`
  - minimal Run bi-di stream behavior for smoke verification
  - smoke client in `/sidecar/scripts/smoke-run.mjs`

## Quick Run

1. Start sidecar:
```bash
cd /Users/gregho/GitHub/AI/ForgeMate/sidecar
npm install
node src/server.mjs
```

2. Run sidecar smoke test (new terminal):
```bash
cd /Users/gregho/GitHub/AI/ForgeMate/sidecar
node scripts/smoke-run.mjs
```

3. Start gateway (requires Go toolchain):
```bash
cd /Users/gregho/GitHub/AI/ForgeMate
go run ./cmd/forgemate-gateway
```

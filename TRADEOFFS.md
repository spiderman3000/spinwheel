# Tradeoffs

Every deliberate shortcut in this repo, with what we gave up and when to
revisit. Ticket links point at the follow-up work. Updated 2026-09-13.

## Platform

### Neon Free + Cloud Run + Cloudflare Pages (not Fly.io / Render / Hetzner)
- **Decision:** Neon Free (`0.5GB`, `100 CU-h/mo`, scale-to-zero, `5GB` egress)
  + Cloud Run free tier (`2M req`, `180k vCPU-sec`, `360k GiB-sec/mo`) + Pages.
  Validated for ~1000 DAU (~150k req/mo ≈ `15k` vCPU-sec — far inside free).
- **Gave up:** Fly's multi-region story; Render's simplicity; Hetzner's fixed cost.
- **Revisit when:** CU-h exhausts (overflow is Launch `$0.106/CU-h`, ~`$5–10/mo`)
  or p95 latency needs regions closer to users.

### Cloud Run sizing: CPU 1 / 512MB / concurrency 80 / min 0 / max 3
- **Decision:** scale-to-zero (cost $0 idle), cap at 3 instances.
- **Gave up:** `1–2s` cold starts on idle wakeups; headroom past ~240 concurrent.
- **Revisit when:** cold starts hurt UX or concurrency saturates.

### Deploy via `gcloud run deploy --source` (Cloud Build)
- **Decision:** push source, let Cloud Build compile (same Dockerfile validated
  locally via `docker build`).
- **Gave up:** reproducible prebuilt-image pins per deploy; first builds are slow.
- **Revisit when:** build times or reproducibility demand Artifact Registry images.

## API surface

### HTTP-only public API; gRPC built but unmounted (→ SPI-17)
- **Decision:** browsers get JSON only. `internal/handler` stays compiled and
  unit-tested with a mount NOTE in `main.go`.
- **Gave up:** typed clients, streaming, grpcurl-driven ops.
- **Revisit when:** a non-browser client needs a contract, or ops wants grpcurl.

### Stdlib-only gateway, no chi / grpc-gateway
- **Decision:** 3 routes don't justify a framework or codegen churn; subtree
  patterns compatible with the `go 1.21` language version.
- **Gave up:** expressive routing, middleware ecosystem.
- **Revisit when:** routes grow past ~10 or need path-parameter ergonomics.

## Data & migrations

### Migrations run out-of-band; no boot-migrate (→ SPI-11)
- **Decision:** `migrate` binary + `DIRECT_URL` from a machine, before deploy.
  Distroless image has no shell. Safe while the schema is stable (v2).
- **Gave up:** fully hands-off deploys.
- **Revisit when:** schema churn picks up. Failure mode without it: code
  deployed against an old schema 500s until someone runs migrate.

### UpdateWheel = full item-set replace
- **Decision:** Postgres deletes + re-inserts items (positions rewritten);
  spins keep `winner_idx`, `winner_item_id` goes NULL on replace.
- **Gave up:** item-row stability across edits (history shows index fallback).
- **Revisit when:** audit requirements need immutable item rows (then version
  item sets instead of replacing).

### Retention: prune events/spins older than 90 days
- **Decision:** keeps the `0.5GB` Neon project from filling (a full disk
  fails writes).
- **Gave up:** indefinite raw history (aggregates can still be kept).
- **Revisit when:** storage grows or analytics needs longer lookback. Pruning
  itself is not yet automated — oldest TODO in this file.

## Spin integrity

### Strict `items_hash` (fail-closed)
- **Decision:** empty or stale hash → `wheel_changed` (HTTP 400 / gRPC
  FailedPrecondition). A spin against an unknown list could never verify.
- **Gave up:** lenient spins that "just work" across FE/BE version skew.
- **Revisit when:** never for integrity; only the FE refresh UX around it.

### HMAC fail-fast at boot, fail-closed in service
- **Decision:** server refuses to start without `HMAC_SECRET`; spins error
  without it. No unsigned code path exists.
- **Gave up:** running the server in a degraded demo mode.
- **Revisit when:** never — this one is load-bearing.

### No deterministic seed mode (→ SPI-16)
- **Decision:** always `crypto/rand`, even when the client sends a seed.
- **Gave up:** scripted exact-winner assertions in CI/FE tests.
- **Revisit when:** test determinism matters more than the extra code path
  (must stay impossible in production).

### Browser checks `expiry`, never `sig`
- **Decision:** the HMAC secret can't ship to the browser, so FE tamper
  checking is limited to expiry + server trust. Real verification is
  server-side (unit-covered) and FE-displayed.
- **Gave up:** client-side fraud proofs.
- **Revisit when:** a public-verifiability story is needed (e.g. signed
  transparency log, asymmetric sigs).

## History & events

### `total_count` is page-length approximate (→ SPI-15)
- **Decision:** ship pagination now, add `CountSpins` later. Marked in code.
- **Gave up:** correct "showing X of Y" UI.
- **Revisit when:** history UI exists (no reader yet — pure backlog).

### Event idempotency key `(session_id, client_ts, type, spin_id)`
- **Decision:** redelivery is a no-op, enforced by a unique index.
- **Gave up:** counting raw deliveries (dedupe hides retries).
- **Revisit when:** delivery diagnostics are needed (log redeliveries).

### Spin lifecycle events live in `Wheel.tsx`, not `Playground`
- **Decision:** the component that owns `wheel_id`/`spin_id` emits
  `spin_start`/`spin_end` — equivalent funnel coverage, no prop drilling.
- **Gave up:** literal ticket-task placement.
- **Revisit when:** never; architecture over ticket literalism.

## Identity & privacy

### Anon-first: `sw_sid` cookie, `x-user-id` optional fallback
- **Decision:** no logins; UUIDv7 `HttpOnly`/`Lax`/1yr cookie is the identity,
  header survives only for grpcurl/tests.
- **Gave up:** cross-device identity, abuse attribution beyond IP+session.
- **Revisit when:** accounts arrive (`optional user_id` + link table, no
  contract break per v1 freeze).

### Cross-site cookie gap covered by `localStorage` echo (→ SPI-14)
- **Decision:** `Lax` cookies don't travel on cross-site fetch (Pages × Cloud
  Run), so FE sends a `session_id` echo the server adopts. No custom domain
  needed today.
- **Gave up:** pure-cookie sessions; slight client complexity.
- **Revisit when:** `api.yourdomain.com` first-party hostname lands (then the
  echo becomes a fallback, not the mechanism).

### Raw IPs never persisted or logged
- **Decision:** SHA-256 `ip_hash` in rows and logs; `CF-Connecting-IP` →
  leftmost `X-Forwarded-For` → connection address.
- **Gave up:** IP-based forensics and geo without extra work.
- **Revisit when:** abuse response needs more (then hash+salt rotation plan).

## Limits & hardening

### Rate limit 100/min per IP+session, in-memory per instance
- **Decision:** no Redis; each Cloud Run instance tracks its own windows.
- **Gave up:** exact global limits — effective limit scales with instance
  count (fine at max 3).
- **Revisit when:** distributed enforcement matters (shared store) or limits
  need per-route tuning.

### 1MB body cap, CORS allowlist, credentialed preflights
- **Decision:** `MaxBytesReader` + `ContentLength` fast-path (413 over);
  explicit origin echo (no wildcard with credentials); OPTIONS short-circuits
  before limit/logging.
- **Gave up:** large payloads (nothing needs them), open CORS.
- **Revisit when:** an endpoint needs bigger bodies (raise per-route).

### CSP: `connect-src` only, via meta tag
- **Decision:** build-injected meta tag (API origin from `VITE_API_URL`,
  localhost fallback). Covers the exfiltration vector that matters.
- **Gave up:** full policy (script/style/img) and headers-only directives.
- **Revisit when:** hardening pass — move to hosting headers (`_headers`).

## FE sync & dev experience

### Added `POST`/`PUT /v1/wheels` (not in the original ticket)
- **Decision:** FE lists are local-only, so server-driven spins needed a sync
  route. Full-list replace with stable IDs.
- **Gave up:** ticket literalism; added a small write surface (abuse-capped
  by the same rate limiter).
- **Revisit when:** multi-user wheels need authZ per wheel.

### Weight `<= 0` → `1.0`, color `→ ''` default in the HTTP layer
- **Decision:** FE sends option-only items; the same defaults are baked into
  the client hash recipe so hashes agree byte-for-byte.
- **Gave up:** distinguishing "unset" from "1.0" downstream.
- **Revisit when:** FE exposes weight/color pickers (then send them truly).

### Sync-on-fingerprint-change at spin time
- **Decision:** no effect-driven syncing; the list is pushed only when it
  changed since the last spin. One extra round trip on change, zero races
  with animation.
- **Gave up:** sub-second spin latency on edited lists.
- **Revisit when:** latency budget demands optimistic sync.

### InMemory dev seeding (demo wheel + logged id/hash)
- **Decision:** `DATABASE_URL` unset → seeded wheel so the API is usable with
  zero setup. Never seeds on Postgres.
- **Gave up:** dev/prod parity for first-run state.
- **Revisit when:** dev needs Postgres parity (then seed via migrate data).

## Toolchain

### Dropped dead vendored `google/api` protos
- **Decision:** deleted, not fixed — corrupt since day one, referenced by
  nothing, excluded from buf. Unblocked `bazel test //...`.
- **Gave up:** nothing (dead code).
- **Revisit when:** HTTP annotations (transcoding) are ever wanted — re-vendor
  properly instead.

### Pinned buf toolchain; codegen at Docker build time
- **Decision:** `buf v1.32.2` + pinned plugins in the Dockerfile; generated
  code is gitignored build output.
- **Gave up:** checking in generated code (simpler diffs, slower builds).
- **Revisit when:** builds get slow or offline builds are needed.

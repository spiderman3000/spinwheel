# Spinwheel Contract v1 — frozen (2026-09-12, SPI-7 step 0)

No login. Anon identity = `sw_sid` cookie (HttpOnly, SameSite=Lax, 1yr, UUIDv7).
Old `x-user-id` gRPC metadata is dropped (fallback for grpcurl/tests only).
Future login = `optional string user_id` + `sessions(session_id → user_id)` link table. No breaking change.

## Flows

### Spin (server-authoritative)
1. FE renders wheel, computes `items_hash = sha256(sorted(id|option|weight|color))`, generates `client_seed` (16B hex).
2. Optional `POST /v1/events {type: SPIN_START, wheel_id, session_id}` (funnel).
3. `POST /v1/wheels/{id}/spin {client_seed, items_hash}` → `{spin_id, winner_index, winner_item, nonce, sig, expires_at: +60s}`.
4. FE verifies `expires_at > now`, animates to `winner_index` (6×2π + offset), shows `winner_item.option`.
5. `POST /v1/events {type: SPIN_END, wheel_id, spin_id, sig}`. Server re-verifies `sig`, else rejects win claim.

### Pageview (deterministic analytics)
Every load → `POST /v1/events {type: PAGEVIEW, path, session_id}` to **same host** as spins (blocking it breaks spins). Backup sources: BE access log (every HTTP hit, even JS-off) + Cloudflare edge Logpush. Reconcile `DB pageviews vs edge vs app logs`.

## Field rationale

| Field | Why present |
|---|---|
| `wheel_id` | N wheels exist; no implicit current wheel. |
| `client_seed` | Client entropy; `seed_server(crypto/rand 32B, secret) + client_seed → winner`. Stops server pre-farming, enables fairness audit. |
| `items_hash` | Detects stale/forged item lists → `FAILED_PRECONDITION(wheel_changed)`. |
| `seed (int64)` | Deterministic `bazel test` only; ignored in prod. |
| `spin_id` | Idempotency + join key (events, history, logs). |
| `winner_index` | Unambiguous animation target; options can duplicate. |
| `winner_item` | Avoids refetch; preserves `weight` (previously dropped in handler). |
| `nonce` | Per-spin uniqueness for `sig` on repeat winners. |
| `sig = HMAC_SHA256(HMAC_SECRET, spin_id\|wheel_id\|winner_idx\|items_hash\|nonce)` | Opaque to FE; server verifies on `spin_end`/redeem. HMAC = 1 env var; upgrade to Ed25519 if FE offline verify needed. |
| `expires_at (+60s)` | Anti-replay; FE pre-checks, server enforces. |
| `EventType PAGEVIEW/SPIN_START/SPIN_END` | Conversion funnel without fingerprinting. |
| `session_id` | The anon identity. Joins pageviews→spins. |
| `client_ts` (+ server `received_at`) | Ordering + skew detection. |
| `ua/cf_ray/ip_hash` | **Server-filled** from headers, never client body. `ip_hash = SHA256(ip+salt)`, no raw PII. |

## HTTP mapping (gRPC = internal, HTTP = public)

| gRPC | HTTP |
|---|---|
| `SpinWheel` | `POST /v1/wheels/{id}/spin → 200 SpinWheelResponse` |
| `RecordEvent` | `POST /v1/events → 204` (idempotent on session_id+client_ts+type+spin_id) |
| — | `GET /healthz → 200 {ok:true}` |
| `GetSpinHistory` | internal only (dashboard reads DB directly) |

CORS: `yourdomain.com` only. Body limit 1MB. Rate limit `100/min` by `IP+session_id`.

## Canonical examples

`items_hash`: `sha256("id1|Red|1|#f00\nid2|Blue|2|#00f")` (sorted lines, `\n`-joined, UTF-8).
`sig`: `hex(HMAC_SHA256("secret", "spin_01|wheel_01|2|<items_hash>|n_01"))`.
`expires_at`: `now + 60s`; server allows 10s clock skew.

## Versioning

v1 frozen. Changes = new fields only (never renumber). Breaking change → `v2` package + 3-month dual-serve. Regenerate via `buf generate` from the repo root (toolchain pinned in `Dockerfile`, config in `buf.yaml`/`buf.gen.yaml`). Check compat with `buf breaking --against '.git#branch=main'`.

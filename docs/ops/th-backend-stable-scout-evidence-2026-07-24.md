# TH Backend Stable Scout · Evidence (2026-07-24)

> **Module:** M-TH-backend-stable-scout  
> **Product:** TransitHub · worktree `C:\Users\yuanjia\orca\workspaces\src\th-backend-stable-scout`  
> **Branch:** `xvyimu/th-backend-stable-scout` · HEAD at start `f7a8b9bd`  
> **Scope:** read-only map of `migrations/` · DB pool/timeouts · Redis hot paths → evidence only  
> **Non-goals:** push · production `FRONTEND_MODE`/DSN · delete `web/default` · go.mod major bumps · D7  
> **D7:** NOT EXECUTED

## Executive summary

| Area | Finding |
|------|---------|
| **A · migrations** | Only portable-ish file track is `migrations/main/000001_baseline.{up,down}.sql` — **SQLite-shaped export**. Empty SQLite `up`+`version=1` is green. MySQL/PG file migrate **not validated**; keep `SQL_AUTO_MIGRATE` default **true** for server dialects until dialect baselines exist (TARGET §4 / `migrations/README.md`). |
| **B · pools/timeouts** | DB pool via `SQL_MAX_*` with SQLite-safe defaults (2/4). Relay outbound timeouts via `RELAY_*`. HTTP server `ReadHeaderTimeout`/`IdleTimeout` only (no global `WriteTimeout` — intentional for SSE). |
| **C · Redis** | Optional: unset `REDIS_CONN_STRING` → `RedisEnabled=false`, memory fallbacks for rate limits/caches. **Startup fail-closed** if conn string set but ping fails (`FatalLog`). Runtime rate-limit Redis errors often **HTTP 500** (not silent open). Adaptive metrics sync is best-effort (silent). |
| **Code delta** | Docs-only; no business/code change in this module run. |

---

## A) Migrations · three-dialect audit

### A.1 File inventory

| Path | Role | Notes |
|------|------|-------|
| `migrations/README.md` | Policy SSOT | Tool = golang-migrate / `cmd/dbmigrate`; three-dialect rules; baseline gate |
| `migrations/main/000001_baseline.up.sql` | Main schema baseline | ~21 KiB · **31** `CREATE TABLE` · **106** `CREATE INDEX` (incl. UNIQUE) · SQLite export |
| `migrations/main/000001_baseline.down.sql` | Destructive down | Empty/dev only; drop all baseline tables; **not** for live rollback |
| `migrations/archive/migration_v0.2-v0.3.sql` | Historical data patch | Not in auto chain |
| `migrations/archive/migration_v0.3-v0.4.sql` | Historical data patch | Not in auto chain |
| `migrations/archive/README.md` | Archive policy | Do not re-run on already-patched DBs |
| `migrations/clickhouse/README.md` | Log DB track | CH still `model.migrateClickHouseLogDB`; do not merge into `main/` |

**No** `*.mysql.*` / `*.postgres.*` / `main/{sqlite,mysql,postgres}/` dialect split files yet.

### A.2 Portability matrix (current baseline)

| Dialect | Empty-DB file migrate | App runtime support | Risk / gap |
|---------|----------------------|---------------------|------------|
| **SQLite** | **PASS** (CI + local runner) | Yes (default when `SQL_DSN` unset) | Baseline is the source export shape; CI-required |
| **MySQL** ≥5.7.8 | **Not validated** / expected fail on same SQL | Yes via GORM AutoMigrate | Backticks + SQLite types (`numeric`, untyped `json`, `datetime`, `real`); no ENGINE/charset |
| **PostgreSQL** ≥9.6 | **Not validated** / expected fail | Yes via GORM AutoMigrate | `"group"`/`"key"` quoting differs; bool/json/serial mapping; reserved words |
| **ClickHouse** (log only) | Separate track | `LOG_SQL_DSN=clickhouse://…` | Not on main migrate track |

### A.3 Baseline SQL portability flags (examples)

| Pattern in `000001_baseline.up.sql` | SQLite | MySQL | PG | Risk |
|-------------------------------------|--------|-------|----|------|
| Backtick identifiers `` `group` `` | OK | OK | Fail / needs `"` | High for raw PG apply |
| Type mix `integer`/`numeric`/`text`/`json`/`datetime`/`real`/`decimal(10,6)` | OK (loose) | Partial | Partial | Medium — GORM maps better than raw SQL |
| `DEFAULT false` on `numeric` columns | OK-ish | Dialect-dependent | Prefer boolean | Medium |
| Double-quoted string defaults e.g. `DEFAULT "default"` | SQLite | May need single quotes | May need single quotes | Medium |
| `PRIMARY KEY (`id`)` without AUTOINCREMENT/SERIAL | SQLite rowid-style | Needs AI for GORM insert path | Needs serial/identity for inserts | High if file-migrate cutover without GORM |
| **Missing table `refund_intents`** | N/A in file | Created by AutoMigrate today | same | **Schema drift:** model `RefundIntent` is in `migrateDB()` AutoMigrate list but **absent** from `000001` baseline |

### A.4 Ops policy (unchanged; reconfirmed)

1. Prefer portable subset; dialect split only with documented mechanism (`migrations/README.md`).  
2. Before `SQL_AUTO_MIGRATE=false` on MySQL/PG: dialect baselines + empty-DB CI + force procedure (`docs/ops/migrate-three-dialect-strategy.md`).  
3. Production: do not `down` past baseline; restore backup.  
4. Hard constraint: do not drop SQLite or MySQL support without product decision (AGENTS / TARGET §5).

### A.5 Related TARGET §4

`docs/ARCHITECTURE_TARGET.md` §4: `migrations/` is SSOT for **new** evolution; current `000001` only proven on empty SQLite; server dialects stay AutoMigrate until cutover evidence exists.

---

## B) Timeouts / connection pools

### B.1 Database pool

| Item | Source | Default | Hot path |
|------|--------|---------|----------|
| Max idle conns | env `SQL_MAX_IDLE_CONNS` | MySQL/PG **100** · SQLite **2** | `model.connectionPoolConfig` → `configureConnectionPool` on `InitDB` / `InitLogDB` |
| Max open conns | env `SQL_MAX_OPEN_CONNS` | MySQL/PG **1000** · SQLite **4** | same |
| Conn max lifetime | env `SQL_MAX_LIFETIME` (seconds) | **60s** | `sqlDB.SetConnMaxLifetime` |
| Main DSN | `SQL_DSN` | empty → SQLite `common.SQLitePath` | `model.chooseDB` |
| Log DSN | `LOG_SQL_DSN` | empty → share main DB | `InitLogDB` |
| SQLite path | `SQLITE_PATH` | `one-api.db?_busy_timeout=30000` | `common.InitEnv` re-attaches busy_timeout if missing |
| SQLite PRAGMA | code | WAL · `busy_timeout=5000` · `synchronous=NORMAL` | `applySQLitePragmas` (note: DSN busy_timeout 30s vs PRAGMA 5s — both present) |
| AutoMigrate gate | `SQL_AUTO_MIGRATE` | **true** if unset | `shouldAutoMigrate`; false/0/no/off skips GORM migrate |

**Risk:** MySQL/PG default `max_open=1000` is aggressive for small managed DB tiers — operators should override. SQLite defaults intentionally tiny (OOM / file-handle protection; see comments in `model/main.go`).

**Tests:** `model.TestConnectionPoolConfig*` — PASS (exit 0).

### B.2 Relay / outbound HTTP

| Env | Default (s / count) | Applied in |
|-----|---------------------|------------|
| `RELAY_TIMEOUT` | **0** (unbounded client timeout) | Documented: must **not** set `http.Client.Timeout` (kills SSE); see `service.InitHttpClient` |
| `RELAY_IDLE_CONN_TIMEOUT` | 90 | `common.NewOutboundHTTPTransport` |
| `RELAY_DIAL_TIMEOUT` | 10 | `net.Dialer.Timeout` |
| `RELAY_TLS_HANDSHAKE_TIMEOUT` | 10 | transport |
| `RELAY_RESPONSE_HEADER_TIMEOUT` | 120 | transport (stall protection) |
| `RELAY_EXPECT_CONTINUE_TIMEOUT` | 1 | transport |
| `RELAY_MAX_IDLE_CONNS` | 500 | transport |
| `RELAY_MAX_IDLE_CONNS_PER_HOST` | 100 | transport |
| `STREAMING_TIMEOUT` | 300 | `constant.StreamingTimeout` (stream idle cutoff) |
| `READINESS_TIMEOUT_SECONDS` | 3 | `controller.GetReadiness` |

### B.3 Inbound HTTP server

| Env | Default | Applied in |
|-----|---------|------------|
| `HTTP_READ_HEADER_TIMEOUT_SECONDS` | 10 | `main.newHTTPServer` `ReadHeaderTimeout` |
| `HTTP_IDLE_TIMEOUT_SECONDS` | 120 | `IdleTimeout` |
| `HTTP_MAX_HEADER_BYTES` | 1 MiB | `MaxHeaderBytes` |
| WriteTimeout | **unset** | Intentional — stream responses (`.env.example`) |

**Tests:** `TestNewHTTPServer*` live in package `main`; `go test .` **fails setup** in this worktree because `frontend_assets_embedded.go` requires `web/classic/dist` (missing embed assets). Package-level `./common` and `./model` tests still run.

### B.4 Init order (hot path)

`main.InitResources`: `.env` → `InitEnv` → `InitHttpClient` → `model.InitDB` → authz → `InitLogDB` → **`common.InitRedisClient`** → perf metrics …

---

## C) Redis usage · hot paths

### C.1 Enablement / pool

| Item | Behavior |
|------|----------|
| Enable | `REDIS_CONN_STRING` non-empty → parse URL, `PoolSize` from `REDIS_POOL_SIZE` (**default 10**), Ping 5s |
| Disable | empty conn string → `RedisEnabled=false`, log, continue (**optional Redis**) |
| Bad URL / ping fail | `FatalLog` — process **does not start** (startup fail-closed when Redis configured) |
| Sync TTL for hash caches | `SYNC_FREQUENCY` seconds (default 60) via `RedisKeyCacheSeconds()` |

### C.2 Hot path table

| Package / function | Purpose | Failure / degrade | Risk |
|--------------------|---------|-------------------|------|
| `common.InitRedisClient` | Client lifecycle | Fatal on parse/ping if configured | Mis-set URL blocks boot |
| `common.Redis{Get,Set,Del,HSetObj,HGetObj,Incr,HIncrBy,HSetField}` | Shared primitives | Errors returned to callers; **no** auto memory fallback at this layer | Callers must branch on `RedisEnabled` |
| `common.RedisIncr` / `HIncrBy` / `HSetField` | Atomic field ops | **No-op success if key missing or TTL≤0** (only mutates when `ttl > 0`) | Stale/missing cache may skip delta; quota paths must not rely solely on this |
| `middleware.redisRateLimiter` / `userRedisRateLimiter` | Global/API/critical/search rate limits | Redis error → **500 + Abort** (fail-closed request) | Redis blip → user-visible 500; multi-instance needs Redis |
| `middleware.redisRateLimitHandler` (model request) | Model RPM limits | Redis error → OpenAI-style error **500** | same |
| `middleware.redisEmailVerificationRateLimiter` | Email send limit | Redis error → **fallback memory** | Fail-open to single-node memory under Redis error |
| `model.token_cache` / `token.go` cache paths | Token HMAC hash cache | Miss → DB; populate async if `shouldUpdateRedis` | Multi-instance consistency depends on Redis + TTL |
| `model.user_cache` | UserBase hash cache | Miss → DB; invalidate on status change | Role/status demotion relies on invalidate paths |
| `model.subscription` + `cachex.HybridCache` | Plan / plan-info cache | Redis off → in-memory hot cache | Multi-instance without Redis → per-process drift |
| `service.channel_affinity` + HybridCache + Lua ZSET LRU | Sticky channel affinity | Redis off → memory; set-limit uses Eval when Redis on | Affinity break under Redis loss |
| `service.SyncAdaptiveMetricsToRedis` | Multi-instance metrics snapshot | Best-effort `_ = RedisSet`; silent | Metrics only |
| `pkg/perf_metrics` `recordRedis` | Relay perf counters in Redis | Best-effort when Redis on (local buckets always) | Metrics accuracy |
| `service.CheckNotificationLimit` | Notify flood control | Redis path errors return error; memory path if disabled | Operational |
| `controller.GetReadiness` | k8s readiness | If Redis enabled, Ping required or **503** | Correct: ready means Redis up when configured |
| `main` (Redis enabled) | Forces `MemoryCacheEnabled=true`; 30s adaptive sync ticker | N/A | Side effect of enabling Redis |

### C.3 Fail-open vs fail-closed (summary)

| Class | Mode |
|-------|------|
| Redis **not configured** | Degraded single-node (memory rate limit / HybridCache mem) — **by design** |
| Redis **configured, boot** | Fail-closed (fatal on ping) |
| Redis **configured, runtime rate limit** | Mostly fail-closed (500) |
| Email verification rate limit | Fail-open to memory on Redis error |
| Cache populate / adaptive metrics / perf | Best-effort / soft |

---

## Verification (commands + exit codes)

Run from worktree root `th-backend-stable-scout` on 2026-07-24:

| Command | Exit | Notes |
|---------|------|-------|
| `go build -o NUL .` | **1** | `frontend_assets_embedded.go:20: pattern web/classic/dist: no matching files found` — embed asset gap in this worktree; not introduced by scout |
| `go test ./model/ -run 'ConnectionPool' -count=1` | **0** | 3 pool config tests PASS |
| `go test ./common/ -run 'HTTP\|Timeout\|Relay' -count=1` | **0** | package ok |
| `go test . -run 'TestNewHTTPServer' -count=1` | **1** | same classic `dist` embed setup failure as `go build` |
| `pwsh -NoProfile -File scripts/migrate-three-dialect.ps1` | **0** | **PASS sqlite** version=1; **SKIP** mysql/pg (no `MIGRATE_*_URL`) |

No production DSN touched. No migrate against non-empty or remote DBs.

---

## Residual risks / follow-ups (docs-only recommendations)

1. **Dialect baselines** for MySQL/PG before any `SQL_AUTO_MIGRATE=false` production cutover (already in strategy docs).  
2. **Baseline drift:** add `refund_intents` (and any other post-export AutoMigrate-only tables) to a new `00000N` migration **or** regenerate SQLite baseline with explicit review.  
3. **Embed assets:** ensure `web/classic/dist` (or build tags) present for `go build .` in CI/agents.  
4. **RedisIncr TTL gate:** document that quota deltas only apply when key has positive TTL; primary authority remains DB.  
5. Rate-limit Redis errors returning 500 — acceptable fail-closed; ops should alert on Redis availability when `REDIS_CONN_STRING` is set.

---

## Sign-off

| Field | Value |
|-------|--------|
| Status | **DONE** · **in-review** |
| Coord | th-coord — scout complete; docs commit only; **no push** |
| D7 | **NOT EXECUTED** |
| Code changes | evidence doc only |

### Required reads (completed)

- `docs/PROJECT.md`  
- `docs/ARCHITECTURE_TARGET.md` §4  
- `migrations/README.md`  
- `docs/ops/migrate-three-dialect-strategy.md`  
- AGENTS.md DB / three-dialect rules  

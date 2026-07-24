# M-TH-be-migrate-3db · evidence · 2026-07-24

## D7 FLIP: NOT EXECUTED

No production migrate. No `SQL_AUTO_MIGRATE=false` on any live DSN. No production DSN used. No `git push`. No MySQL/PG “green” claim without empty-DB apply.

## Worktree identity

| Field | Value |
|-------|--------|
| Module ID | **M-TH-be-migrate-3db** |
| Worktree (absolute) | `C:\Users\yuanjia\orca\workspaces\src\th-be-migrate-3db` |
| Branch | `xvyimu/th-be-migrate-3db` |
| Tip (start) | `f7a8b9bde34ff8c2a9b9683b1d1ad59970b6c3b0` (`docs(ops): TH E2E operator card`) |
| Tip (post evidence) | this commit |
| Agent | claude |
| Scope | `migrations/` audit · `docs/ops/` evidence · empty SQLite migrate runner only |
| Date | 2026-07-24 |
| Status | **DONE** · **in-review** (th-coord) |

## Boundary

| In | Out |
|----|-----|
| Diff `000001_baseline` vs `model.migrateDB` AutoMigrate export | Production migrate / force baseline |
| `scripts/migrate-three-dialect.ps1` (no `MIGRATE_*_URL`) | Claiming MySQL/PG file-migrate green |
| Next migration file **suggestion** (not implemented) | go.mod major · business code · push default branch |
| | **D7** · `SQL_AUTO_MIGRATE=false` production |

## Pre-read

| Path | Result |
|------|--------|
| `docs/ops/th-backend-stable-scout-evidence-2026-07-24.md` | **ABSENT** in this worktree; read from coord copy `../th-coord/docs/ops/…` (scout already flagged missing `refund_intents`) |
| `migrations/README.md` | Read — SQLite CI only; MySQL/PG keep AutoMigrate until dialect baseline gate |
| `docs/ops/migrate-three-dialect-strategy.md` | Read — runner policy; no prod DSN |
| `model/main.go` `migrateDB()` | Read — AutoMigrate list includes `&RefundIntent{}` |
| `model/refund_intent.go` | Read — table `refund_intents` |
| `scripts/export-sqlite-schema` | Used offline to regenerate current AutoMigrate DDL |

## 1) Baseline vs AutoMigrate drift

Method:

1. Parse `migrations/main/000001_baseline.up.sql` tables/indexes.
2. `go run ./scripts/export-sqlite-schema/` → `.tmp/current-automigrate-schema.sql` (same path as `ExportMigrateForSchema` / production AutoMigrate).
3. Compare table set, index set, and per-table column DDL.

### 1.1 Table inventory

| Source | Table count | Notes |
|--------|-------------|--------|
| `000001_baseline.up.sql` | **31** | No `refund_intents` |
| Current AutoMigrate export | **32** | + `refund_intents` only |

**Only in current (missing from baseline):** `refund_intents`  
**Only in baseline:** *(none)*  
**Shared 31 tables:** column-level DDL **identical** (no extra/missing columns on shared tables).

### 1.2 Index inventory

| Source | Index count (named CREATE INDEX / UNIQUE INDEX) |
|--------|--------------------------------------------------|
| Baseline | **125** |
| Current | **129** |

**Only in current:**

| Index | Table |
|-------|--------|
| `idx_refund_intents_idempotency_key` (UNIQUE) | `refund_intents` |
| `idx_refund_intents_status` | `refund_intents` |
| `idx_refund_intents_token_id` | `refund_intents` |
| `idx_refund_intents_user_id` | `refund_intents` |

### 1.3 Missing table DDL (AutoMigrate export, SQLite shape)

```sql
CREATE TABLE `refund_intents` (
  `id` integer,
  `idempotency_key` varchar(128) NOT NULL,
  `token_id` integer NOT NULL,
  `user_id` integer,
  `token_quota` integer NOT NULL DEFAULT 0,
  `extra_reserved` integer NOT NULL DEFAULT 0,
  `subscription_id` integer,
  `funding_source` varchar(32),
  `funding_request_id` varchar(128),
  `wallet_consumed` integer NOT NULL DEFAULT 0,
  `token_key` varchar(128),
  `is_playground` numeric NOT NULL DEFAULT false,
  `wallet_done` numeric NOT NULL DEFAULT false,
  `subscription_done` numeric NOT NULL DEFAULT false,
  `token_done` numeric NOT NULL DEFAULT false,
  `status` varchar(16) NOT NULL,
  `attempts` integer NOT NULL DEFAULT 0,
  `last_error` text,
  `created_at` integer,
  `updated_at` integer,
  PRIMARY KEY (`id`)
);

CREATE UNIQUE INDEX `idx_refund_intents_idempotency_key` ON `refund_intents`(`idempotency_key`);
CREATE INDEX `idx_refund_intents_status` ON `refund_intents`(`status`);
CREATE INDEX `idx_refund_intents_token_id` ON `refund_intents`(`token_id`);
CREATE INDEX `idx_refund_intents_user_id` ON `refund_intents`(`user_id`);
```

Source model: `model.RefundIntent` · `TableName() = "refund_intents"` · registered in `migrateDB()` and `migrateDBFast()`.

### 1.4 Down baseline gap (related)

`000001_baseline.down.sql` drops the 31 baseline tables only (plus comment policy). No `refund_intents` drop needed until a later version creates it. Down remains **empty/dev only**.

### 1.5 Dialect note (unchanged from scout/W2)

Baseline remains **SQLite-exported** (backticks, `numeric`/`json`/`datetime` mix). Empty MySQL/PG apply of the same files is **not** validated and must not be claimed green.

## 2) Three-dialect runner

Command (no `MIGRATE_MYSQL_URL` / `MIGRATE_PG_URL`):

```powershell
pwsh -NoProfile -File scripts/migrate-three-dialect.ps1
```

| Dialect | Result | Detail |
|---------|--------|--------|
| **SQLite** | **PASS** | empty file under `.tmp/migrate-w2-sqlite-*.db` · `up` · `version=1` |
| **MySQL** | **SKIP** | no URL env (expected) |
| **PostgreSQL** | **SKIP** | no URL env (expected) |
| **Script exit** | **0** | `PASS migrate-three-dialect (sqlite required green; others optional/skip)` |

Honest matrix: SQLite file-migrate CI path remains green; MySQL/PG file-migrate still **not** proven.

## 3) Next migration suggestion (not implemented this knife)

Prefer **additive `000002`** over rewriting `000001` (live installs may already `force 1` on old baseline).

| Item | Recommendation |
|------|----------------|
| Version | `migrations/main/000002_add_refund_intents.up.sql` + `.down.sql` |
| Content | SQLite-portable `CREATE TABLE` + 4 indexes from §1.3 (optionally `IF NOT EXISTS` only if dialect policy allows; otherwise plain CREATE for empty→v2 path) |
| Down | `DROP TABLE IF EXISTS refund_intents;` (indexes drop with table on SQLite) |
| Scope | **SQLite CI first** — same as `000001`; do not claim MySQL/PG until dialect selection lands |
| Do **not** | Fold into re-export of `000001` without force/version plan for existing `schema_migrations` |
| Later | Dialect split (`*.mysql.*` / `main/mysql/`) + empty MySQL/PG containers — still blocked by baseline gate in `migrations/README.md` |
| Runtime today | Keep default `SQL_AUTO_MIGRATE=true` so AutoMigrate creates `refund_intents` on MySQL/PG/SQLite until file track catches up |

**Why not implement here:** module boundary is audit + evidence; adding SQL without extending CI version assertion (`version == 1` today) risks false confidence. Implement in a follow-up that also bumps `scripts/migrate-three-dialect.ps1` / quality job expected version to **2**.

## 4) Risk (one line)

**If production/SQLite turns `SQL_AUTO_MIGRATE=false` while only `000001` is applied, `refund_intents` never appears → refund outbox / WP-C paths fail at runtime until AutoMigrate or `000002` lands.**

## Intentionally not done

- No `000002` SQL committed (suggestion only).
- No MySQL/PG empty-DB apply (no non-prod URL; would red on current SQLite-shaped SQL).
- No rewrite of `000001` baseline.
- No push · no D7 · no production env flips.

## Related

| Path | Role |
|------|------|
| `docs/ops/th-backend-stable-scout-evidence-2026-07-24.md` | Prior scout (coord tree) — drift already noted |
| `docs/ops/migrate-three-dialect-strategy.md` | W2 runner / dialect policy |
| `migrations/README.md` | Baseline gate before MySQL/PG cutover |
| `docs/operations/db-migrations.md` | Force baseline + publish contract |
| `model/refund_intent.go` | Domain model |
| `scripts/migrate-three-dialect.ps1` | Empty-DB smoke |

## Closeout

- **DONE** · **in-review** (for th-coord)
- **D7 NOT EXECUTED**
- Evidence path: `docs/ops/th-be-migrate-3db-evidence-2026-07-24.md`

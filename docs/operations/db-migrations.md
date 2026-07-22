# Database migration operations (Phase1 WP-S)

## Goals

- Schema evolution is **reviewable SQL** under `migrations/`.
- App replicas can run with **zero DDL** (`SQL_AUTO_MIGRATE=false`).
- Migrate process is **single-flight** (one Job / master), not every replica.

## Environment

| Variable | Meaning |
|----------|---------|
| `SQL_DSN` | Main DB. Empty / unset → SQLite file (`SQLITE_PATH` / default). MySQL DSN, `postgres://` / `postgresql://` as today. |
| `LOG_SQL_DSN` | Optional separate log DB; ClickHouse stays on GORM/CH path until `migrations/clickhouse` is populated. |
| `SQL_AUTO_MIGRATE` | Default **true** (safe for existing installs). Set **false** when file migrations own schema. |
| `NODE_TYPE` | `slave` skips AutoMigrate (unchanged). |
| `RUN_MODE=migrate` | Existing process mode: init resources then exit after migration path (still uses GORM when AutoMigrate on). Prefer `scripts/db-migrate.ps1` for file migrations. |

## Publish contract (S5)

Order for production cutover:

1. **Backup** main DB (and log DB if separate).
2. **Single-flight migrate**: `go run ./cmd/dbmigrate … up` or `scripts/db-migrate.ps1 -Direction up` (or CI/CD migrate Job) against the target DSN.
3. Deploy app with `SQL_AUTO_MIGRATE=false` (and normal `RUN_MODE`).
4. Smoke: `/healthz`, login, one read path.
5. Only then shift traffic / scale replicas.

Do **not** run file migrate concurrently from multiple pods without a lock/leader.

## Existing databases (force baseline)

Empty DB: run `up` from version 0.

Already-live DB that matches AutoMigrate shape:

1. Backup.
2. Diff schema vs `000001` (or run export tool on a clone).
3. `migrate force 1` (golang-migrate) so version = 1 **without** re-executing baseline DDL.
4. Set `SQL_AUTO_MIGRATE=false`.
5. Future changes only as `000002+`.

## Rollback (S7)

| Environment | Policy |
|-------------|--------|
| Production | Prefer **restore backup**. Do not `down` past baseline on live data. |
| Staging/dev | `db-migrate.ps1 -Direction down` allowed for empty/throwaway DBs. |
| Breaking change | Expand/contract two-step; mark irreversible downs in SQL comments. |

Down of `000001_baseline` drops all baseline tables — **dev only**.

## CI

Quality workflow includes a **SQLite migrate up** smoke job (see `.github/workflows/quality.yml`). MySQL/PostgreSQL matrix is phased later; dialects remain supported via AutoMigrate until dedicated baselines exist.

## Related

- `migrations/README.md` — layout and dialect rules
- `docs/phase1-bid-sql-data-architect.md` — design bid
- `docs/operations/build-and-release.md` — release boundary (binary build; schema is this doc)

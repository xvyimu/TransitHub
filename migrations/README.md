# Database migrations (Phase1 WP-S)

## Tool

- **Default**: [golang-migrate](https://github.com/golang-migrate/migrate) CLI (`migrate`).
- **Source of truth for new schema changes**: SQL files under this tree (not GORM AutoMigrate alone).
- **App default**: `SQL_AUTO_MIGRATE` defaults to **enabled** (conservative). Production cutover sets `SQL_AUTO_MIGRATE=false` and applies files before traffic.

## Layout

```text
migrations/
  main/                 # SQL_DSN (SQLite / MySQL / PostgreSQL)
    sqlite/             # SQLite-shaped DDL
      NNNNNN_name.up.sql
      NNNNNN_name.down.sql
    mysql/              # MySQL-shaped DDL (InnoDB/utf8mb4, prefix-length indexes)
      NNNNNN_name.up.sql
      NNNNNN_name.down.sql
    postgres/           # PostgreSQL-shaped DDL (bigserial, boolean, jsonb, timestamptz)
      NNNNNN_name.up.sql
      NNNNNN_name.down.sql
  archive/              # Historical bin/ data patches (not auto-run)
  clickhouse/           # Optional LOG_SQL_DSN=clickhouse (separate track)
  README.md             # This file
```

**Dialect selection (the one chosen mechanism):** `cmd/dbmigrate` picks the
per-dialect subdirectory from the database URL scheme — `sqlite://` → `sqlite/`,
`mysql://` → `mysql/`, `postgres://` → `postgres/`. Point `-path` at the parent
(`migrations/main`) and the tool descends automatically; point it at a leaf dir
(e.g. `migrations/main/sqlite`) to override. This is the single unambiguous
mechanism required by the baseline gate below — do not also use file suffixes.

Version table: `schema_migrations` (managed by golang-migrate).

## Naming

- `NNNNNN_snake_case.up.sql` / `.down.sql` (six-digit zero-padded version).
- Every version number must exist in **all three** dialect subdirs
  (`sqlite/`, `mysql/`, `postgres/`) with matching up/down files, so any
  configured database applies the same logical schema step.
- The chosen divergence mechanism is the **per-dialect subdirectory** (above).
  Do **not** also introduce `*.mysql.up.sql`-style filename suffixes; the two
  styles must not be mixed (baseline gate rule 1).

## Three-dialect policy

| Dialect | Role | Baseline status |
|---------|------|-----------------|
| **SQLite** | Dev / edge / CI required | `000001_baseline` empty-DB `up` verified (CI job `sqlite-migrate`) |
| **MySQL** | Common production | `000001_baseline` empty-DB `up` verified (CI job `mysql-migrate`, service container) |
| **PostgreSQL** | Preferred production | `000001_baseline` empty-DB `up` verified (CI job `pg-migrate`, service container) |

CI proves an empty-database `up` + `schema_migrations` version `1` for all three
dialects using GitHub Actions **service containers** (throwaway DBs, never a
production DSN). Ops note: `docs/ops/migrate-three-dialect-strategy.md` + runner
`scripts/migrate-three-dialect.ps1` (SQLite required locally; MySQL/PG opt-in env).

Hard constraint (AGENTS.md): **do not remove SQLite or MySQL** without a product decision.

Rules:

1. Prefer standard subset: `CREATE TABLE`, `ADD COLUMN`, indexes.
2. No MySQL-only / PG-only / SQLite-unsupported `ALTER COLUMN` without a fallback branch.
3. Expand/contract for breaking changes; never silent column drop in up migrations.
4. ClickHouse log schema is **not** on the main track.

### Baseline gate before any file-migration cutover

The four gate conditions are now **satisfied for the empty-database case**:

1. **One selection mechanism** — per-dialect subdirectory chosen from the URL
   scheme (see *Dialect selection* above); no filename-suffix style is mixed in.
2. **Empty-DB baseline + version assertion for all three dialects** — `sqlite/`,
   `mysql/`, and `postgres/` each ship `000001_baseline`, and CI asserts
   `schema_migrations` version `1` after `up` on a fresh DB.
3. **Existing-install baseline/force + irreversible-down policy** — documented
   below (*Existing-install baseline*) and in `docs/ops/migrate-three-dialect-strategy.md`.
4. **CI checks without a production database** — jobs `sqlite-migrate`,
   `mysql-migrate`, `pg-migrate` in `.github/workflows/quality.yml` run against a
   pure-Go SQLite file and MySQL/PostgreSQL **service containers** (throwaway).

What is **still not** proven: applying `000001_baseline` on top of a **live,
already-populated** MySQL/PostgreSQL database created by GORM AutoMigrate. Empty
`up` parity is necessary but not sufficient for that. Keep `SQL_AUTO_MIGRATE`
enabled for MySQL/PostgreSQL deployments and treat a production migration as an
explicit, separate, human-gated operation (cutover G-series + D7). This
repository change does not run migrations or change deployment environment values.

### Existing-install baseline (force) and down policy

For a database that already has the schema (GORM AutoMigrate created it), do
**not** run `000001_baseline up` — it would try to re-create existing tables.
Instead mark the baseline as already applied:

```powershell
# Backup first. Point -database at the target DSN (never in CI).
go run ./cmd/dbmigrate -path migrations/main -database "<dsn>" force 1
go run ./cmd/dbmigrate -path migrations/main -database "<dsn>" version   # expect: 1
```

`down` is **destructive** and intended for empty/dev databases only: the baseline
down drops every table. Never migrate `down` past baseline on live data; restore
from backup instead. See `docs/operations/db-migrations.md` § rollback.

## Developer workflow (model ↔ SQL)

1. Change `model/*.go` structs/tags if needed.
2. Same PR: add `migrations/main/{sqlite,mysql,postgres}/NNNNNN_*.up.sql` (+ down
   or mark irreversible) — the same version in all three subdirs, dialect-shaped.
3. Local: `pwsh -File scripts/db-migrate.ps1 -Direction up` (SQLite).
4. CI: `sqlite-migrate`, `mysql-migrate`, `pg-migrate` jobs must all pass.
5. **Forbidden**: rely only on startup AutoMigrate for production schema evolution.

Export helper (draft baseline refresh):

```powershell
go run ./scripts/export-sqlite-schema/ > tmp_schema.sql
```

## Commands

**Preferred runner**: in-repo `cmd/dbmigrate` (pure-Go SQLite driver, no CGO; works on Windows CI).

`-path` points at the parent `migrations/main`; the dialect subdir is chosen
from the URL scheme (see "Dialect selection" above).

```powershell
# Empty SQLite demo (descends into migrations/main/sqlite)
go run ./cmd/dbmigrate -path migrations/main -database "sqlite://.tmp/migrate-demo.db" up
go run ./cmd/dbmigrate -path migrations/main -database "sqlite://.tmp/migrate-demo.db" version

# Empty MySQL / PostgreSQL (throwaway DBs only — never a production DSN).
# MySQL needs multiStatements=true (golang-migrate execs the whole file at once).
go run ./cmd/dbmigrate -path migrations/main -database "mysql://root:root@tcp(127.0.0.1:3306)/th_migrate_empty?multiStatements=true" up
go run ./cmd/dbmigrate -path migrations/main -database "postgres://postgres:postgres@127.0.0.1:5432/th_migrate_empty?sslmode=disable" up

# Or wrapper (SQLite by default)
pwsh -File scripts/db-migrate.ps1 -Direction up
```

Optional external CLI (needs CGO for sqlite3 tag): `go install -tags sqlite3 github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3`.

See also `docs/operations/db-migrations.md`.

## Historical files

`bin/migration_v0.2-v0.3.sql` and `v0.3-v0.4.sql` are archived under `migrations/archive/`. Do not re-run them automatically.

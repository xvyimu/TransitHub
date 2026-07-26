# Migrations · three-dialect strategy (W2)

> Empty-DB evidence + policy. **No production migrate. No D7.**  
> Companion runner: `scripts/migrate-three-dialect.ps1`  
> Tooling: `cmd/dbmigrate` (golang-migrate · pure-Go SQLite).

## Worktree identity (W2 evidence)

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\w2-th-claude` |
| Branch | `xvyimu/w2-th-claude` |
| Date | 2026-07-23 |
| Agent | claude (solo) |

## Dialect matrix

| Dialect | Role | Empty-DB file migrate | Production default | CI evidence |
|---------|------|-----------------------|--------------------|-------------|
| **SQLite** | Dev / edge / **CI required** | **Green** — `000001_baseline` → version `1` | AutoMigrate or file migrate OK | `sqlite-migrate` (pure-Go driver) |
| **MySQL** ≥5.7.8 | Common production | **Green** — dialect-shaped `000001_baseline` → version `1` | Keep **`SQL_AUTO_MIGRATE=true`** until cutover | `mysql-migrate` (service container `mysql:8.0`) |
| **PostgreSQL** ≥9.6 | Preferred production | **Green** — dialect-shaped `000001_baseline` → version `1` | Keep **`SQL_AUTO_MIGRATE=true`** until cutover | `pg-migrate` (service container `postgres:16`) |

Hard constraint (AGENTS.md): do not drop SQLite or MySQL support.

## How the three dialects are kept separate

Each dialect has its own baseline under `migrations/main/{sqlite,mysql,postgres}/`,
and `cmd/dbmigrate` selects the subdir from the database URL scheme (single
unambiguous mechanism — no filename suffixes). Type mapping applied when the
MySQL/PostgreSQL baselines were derived from the SQLite export:

| SQLite | MySQL | PostgreSQL |
|--------|-------|------------|
| autoincrement PK `integer` | `bigint AUTO_INCREMENT` | `bigserial` |
| `numeric` (bool) | `tinyint(1)` | `boolean` |
| `real` | `double` | `double precision` |
| `json` | `json` | `jsonb` |
| `datetime` | `datetime` | `timestamptz` |
| backtick quoting | backticks + prefix-length text indexes (191) | double-quoted reserved words |

MySQL DSN must carry `multiStatements=true` — golang-migrate execs each file as
one statement and go-sql-driver otherwise rejects multi-statement DDL. lib/pq
(Postgres) and the pure-Go SQLite driver handle multi-statement natively.

## Remaining before MySQL/PG file-migrate cutover

Empty-DB parity is proven. Still human-gated:

1. Validate `force 1` baseline on a **live, AutoMigrate-created** MySQL/PG copy (backup first).
2. Only then set `SQL_AUTO_MIGRATE=false` for that dialect in ops runbooks (cutover G-series + D7).

Until then: server dialects stay on GORM AutoMigrate in production.

## Local commands (non-prod)

### SQLite (required)

```powershell
pwsh -NoProfile -File scripts/migrate-three-dialect.ps1
# or:
pwsh -NoProfile -File scripts/sql-migrate-dry-run.ps1
# or CI-parity:
mkdir -Force .tmp | Out-Null
go run ./cmd/dbmigrate -path migrations/main -database "sqlite://.tmp/ci-migrate.db" up
go run ./cmd/dbmigrate -path migrations/main -database "sqlite://.tmp/ci-migrate.db" version
# expect: 1
```

### MySQL / PostgreSQL (optional · empty DB only)

```powershell
# Create empty local DBs first — never point at production DSN.
# MySQL needs multiStatements=true: golang-migrate execs each file in one call.
$env:MIGRATE_MYSQL_URL = 'mysql://user:pass@tcp(127.0.0.1:3306)/th_migrate_empty?multiStatements=true'
$env:MIGRATE_PG_URL    = 'postgres://user:pass@127.0.0.1:5432/th_migrate_empty?sslmode=disable'
pwsh -NoProfile -File scripts/migrate-three-dialect.ps1
```

Against an empty DB the baseline now applies cleanly for all three dialects
(the runner picks the matching `migrations/main/<dialect>/` subdir automatically).

`-RequireMySQL` / `-RequirePostgres` flips SKIP into FAIL when env missing.

## CI today

| Job | File | Behavior |
|-----|------|----------|
| `sqlite-migrate` | `.github/workflows/quality.yml` | empty SQLite `up` + `version == 1` |
| `mysql-migrate` | `.github/workflows/quality.yml` | empty MySQL (service container) `up` + `version == 1` |
| `pg-migrate` | `.github/workflows/quality.yml` | empty PostgreSQL (service container) `up` + `version == 1` |

All three gate `image-reproducibility`; service containers are throwaway (never a production DSN).

## Safety

| Do | Do not |
|----|--------|
| Empty / throwaway DBs for experiments | Point `MIGRATE_*_URL` at production |
| Keep AutoMigrate on for MySQL/PG prod | Flip `SQL_AUTO_MIGRATE=false` without force/up |
| Single-flight migrate Job | Concurrent migrate from every replica |
| Backup before any live force | `down` past baseline on live data |

## Related

| Path | Role |
|------|------|
| [migrations/README.md](../../migrations/README.md) | Layout + three-dialect policy |
| [db-migrations.md](../operations/db-migrations.md) | Ops publish contract |
| [sql-migrate-dry-run-2026-07-22.md](../operations/sql-migrate-dry-run-2026-07-22.md) | Prior SQLite dry-run |
| `scripts/migrate-three-dialect.ps1` | W2 runner |
| `scripts/db-migrate.ps1` | Generic wrapper |
| `cmd/dbmigrate` | In-repo migrate CLI |

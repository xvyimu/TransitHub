# SQL migration dry-run report · 2026-07-22

## Command

```powershell
cd D:\TransitHub\src
$tmp = Join-Path $PWD ".tmp\migrate-dry-$(Get-Date -Format yyyyMMddHHmmss).db"
New-Item -ItemType Directory -Force -Path (Split-Path $tmp) | Out-Null
$url = "sqlite:///" + ($tmp -replace '\\','/')
go run ./cmd/dbmigrate -path migrations/main -database $url up
go run ./cmd/dbmigrate -path migrations/main -database $url version
```

## Expected

| Step | Result |
|------|--------|
| `up` on empty SQLite | applies `000001_baseline` |
| `version` | `1` (or tool-specific “1 dirty” only if interrupted — clean run should be clean) |
| App default | `SQL_AUTO_MIGRATE` still **true** until ops force baseline on live DBs |

## Live / staging recommendation

1. **Backup** main (and log) DB.  
2. Diff live schema vs `000001` (or export clone).  
3. If already at AutoMigrate shape: `force 1` **without** re-running baseline DDL.  
4. Deploy app with `SQL_AUTO_MIGRATE=false` only after force/up success.  
5. Smoke `/healthz` + login.

## Do not

- Run concurrent migrate Jobs without leader election.  
- Set `SQL_AUTO_MIGRATE=false` on live before force/up.  
- Drop SQLite/MySQL dialects.

## Script

`scripts/sql-migrate-dry-run.ps1` automates empty-SQLite up + version for CI/local.

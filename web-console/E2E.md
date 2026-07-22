# web-console login e2e (cookie session)

Isolated smoke for Phase1 console API subset. **Does not** cut over production React.

## Prerequisites

- Backend on `http://127.0.0.1:3000` (or set `TH_API_BASE`)
- Optional Vite on `http://127.0.0.1:5173` for proxy check
- Test account env:

```powershell
$env:TH_E2E_USER = 'root'      # fresh DB default from createRootAccountIfNeed
$env:TH_E2E_PASS = '123456'
# Prefer a dedicated non-prod account on shared DBs
```

> Production / shared SQLite: **change root password**; do not rely on `123456` beyond empty local DB.

## Run

```powershell
cd D:\TransitHub\src
pwsh -File scripts/e2e-web-console-login.ps1
```

## Notes

- Auth is **session cookie + `New-Api-User` header** (same as legacy React). Fresh DB: call `POST /api/setup` first (see script comments / isolated e2e).
- Default empty-DB root `root/123456` only applies when `createRootAccountIfNeed` path runs; current product usually requires **setup wizard** first.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | login + self + (optional) proxy health pass |
| 1 | login failed |
| 2 | self failed after login |
| 3 | backend unreachable |

## Production cutover

**Deferred** (user 2026-07-22). Keep `FRONTEND_MODE` default / embedded React until explicit cutover gate. See `docs/operations/web-console-cutover-rollback.md`.

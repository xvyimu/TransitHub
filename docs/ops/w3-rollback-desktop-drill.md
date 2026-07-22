# W3 · Rollback desktop drill (non-production)

> **D7 FLIP: NOT EXECUTED.**  
> This is a **procedure** for operators on **staging / local** only.  
> SSOT runbook: [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md).  
> Goal: prove **≤5 minute** config/image rollback to React without SQL down-migration.

## Preconditions

| # | Requirement |
|---|-------------|
| 1 | Non-prod stack only (no production DSN / no production edge) |
| 2 | Known-good **integrated** React image digest **or** React separated image available |
| 3 | Backend health known (`/healthz` / `/readyz`) before drill |
| 4 | Stopwatch / wall-clock note for G7 evidence |

## Option A — Integrated binary/image (fastest · preferred)

| Step | Action | Expected | Est. time |
|------|--------|----------|-----------|
| A1 | Note current Vue edge image digest + backend env (`FRONTEND_MODE`) | Snapshot for re-cut if needed | 30s |
| A2 | Deploy previous **integrated** image/binary (embed `web/default` + `web/classic`) | Process/container healthy | 1–2 min |
| A3 | Set `FRONTEND_MODE` **unset** or `auto` (**not** `disabled`) | Backend serves embedded React | 30s |
| A4 | Remove/bypass separated frontend container as public entry if it was sole edge | Single public origin | 30s |
| A5 | Verify | See checklist below | 1 min |

**Expected total:** **≤5 min** when images pre-pulled.

## Option B — Separated stack, swap SPA only

| Step | Action | Expected | Est. time |
|------|--------|----------|-----------|
| B1 | Keep backend `FRONTEND_MODE=disabled` | API-only backend stays | 0 |
| B2 | `docker build -f deploy/separated/Dockerfile.frontend -t new-api-frontend:local .` **or** redeploy known React frontend image | Image ready | 2–4 min (build) / &lt;1 min (redeploy) |
| B3 | Redeploy frontend service only; Nginx contract unchanged (`nginx.conf.template`) | Edge serves React shell | 30–60s |
| B4 | Verify | Checklist | 1 min |

**Expected total:** **≤5 min** if React image already built; build adds time (pre-build for real G7).

## Verification checklist (after rollback)

| # | Check | Pass |
|---|-------|------|
| 1 | Public origin serves **React** console shell (not Vue `web-console` assets only) | ☐ |
| 2 | `POST /api/user/login` works (session cookie) | ☐ |
| 3 | `GET /api/status` → 200 | ☐ |
| 4 | No SQL migration reverse required | ☐ (always true for UI rollback) |
| 5 | Wall-clock from A2/B2 start → login OK ≤ **5 min** | ☐ |

Optional smoke (separated edge still up):

```powershell
pwsh -NoProfile -File deploy/separated/smoke.ps1 -FrontendBase http://127.0.0.1:8080
```

## W3 execution status

| Item | Result |
|------|--------|
| Desktop drill **executed** on this agent host? | **No** |
| Reason | `docker` not on PATH; no operator-owned staging compose; W3 bans production flip |
| Artifact produced | This step table + pointer to runbook |
| Next | Operator runs Option A or B on **non-prod**; paste wall-clock + checklist into [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G7 |

## What not to do

- Do not run this against production without human D7 authorization path (rollback after failed flip is separate emergency).  
- Do not delete `web/default` as part of drill.  
- Do not reverse SQL migrations for UI-only rollback.  
- Do not leave dual public React+Vue origins.

## Related

| Path | Role |
|------|------|
| [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md) | Full operator runbook |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G7 status |
| [deploy/separated/README.md](../../deploy/separated/README.md) | Separated topology |

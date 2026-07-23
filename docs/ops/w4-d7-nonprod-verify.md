# W4 · D7 non-production verify pack · **D7 FLIP: NOT EXECUTED**

> One-shot **non-prod** gate check for cutover readiness.  
> **Does not** change production `FRONTEND_MODE`, delete React, migrate live DB, or flip traffic.  
> Orchestrator: [`scripts/w4-d7-nonprod-verify.ps1`](../../scripts/w4-d7-nonprod-verify.ps1).

## Quick start

```powershell
# From repo root (this worktree or D:\TransitHub\src)
$env:TH_API_BASE = 'http://127.0.0.1:3000'   # optional; default same
$env:TH_E2E_USER = '<non-prod-admin>'        # required for G2/G3 — never commit
$env:TH_E2E_PASS = '<non-prod-secret>'

pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1
```

Faster local re-check (skip heavy builds):

```powershell
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
```

Contract + probes only (honest **exit 10** — not flip-ready):

```powershell
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipAuth -SkipConsoleBuild -SkipBackendBuild
```

## What it runs

| Step | Maps to gate | How |
|------|--------------|-----|
| `GET /healthz` | G5/runtime | Fail → exit **3** |
| `GET /api/status` | smoke | Fail → exit **6** |
| `python scripts/validate-console-contract.py` | G3 contract | Fail → exit **5** |
| Login + `GET /api/user/self` | **G2** | Needs `TH_E2E_*`; fail → **1** / **2** |
| `GET /api/channel/` key-omission check | **G3** live | Fail / key leak → **4** |
| `pnpm build` in `web-console/` | console quality | Fail → **5** |
| `go build -tags frontend_external` | **G5** | Fail → **5** |

Related standalone scripts (still valid):

| Script | Role |
|--------|------|
| `scripts/e2e-web-console-login.ps1` | Login-only e2e (defaults root/123456 if env unset — **prefer** verify pack which refuses silent default for G2) |
| `deploy/separated/smoke.ps1` | Separated edge smoke (needs Vue/React frontend base) |
| `scripts/migrate-three-dialect.ps1` | Empty-DB migrate (not part of D7 traffic flip) |

## Exit codes (no fake green)

| Code | Meaning | Operator action |
|------|---------|-----------------|
| **0** | All selected steps green | Still need G4 image · G6 soak · G7 drill · **G8 human** before production flip |
| **1** | Login failed | Mint non-prod admin; see [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) |
| **2** | Self after login failed | Cookie / `New-Api-User` mismatch |
| **3** | Backend unreachable | Start non-prod API; check `TH_API_BASE` |
| **4** | Channels RO failed or key material on list | Admin role + ChannelRead; file bug if keys leak |
| **5** | Contract or build failed | Fix OpenAPI / `web-console` / Go build |
| **6** | `/api/status` failed | Backend routing / readiness |
| **10** | **Credentials not set** or `-SkipAuth` | Export `TH_E2E_USER` + `TH_E2E_PASS` (non-prod only). **Not** exit 0 |

Credentials checklist (names only): [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md).

## W4 agent recorded run (this message)

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\w4-th-claude` |
| Branch | `xvyimu/w4-th-claude` |
| Tip | `97516c0f` (start) |
| Date | 2026-07-23 |
| Host | healthz **200** · `TH_E2E_*` **unset** · docker **absent** |

| Command | Exit |
|---------|-----:|
| `pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild` | **10** (`healthz=0 status=0 contract=0 login=10 channels=10`) |
| `python scripts/validate-console-contract.py` | **0** |
| `pwsh -NoProfile -File scripts/e2e-web-console-login.ps1 -SkipVite` | **1** |
| `go build -tags frontend_external …` | **0** |
| `web-console` pnpm typecheck / test / build | **0** / **0** / **0** |

Full table: [w4-arch-upgrade-transithub-claude.md](./w4-arch-upgrade-transithub-claude.md).

**Honest outcome without creds:** exit **10** (or **1** if calling legacy e2e with default root). **Not** treated as G2/G3 green.

## Rollback non-prod (minimal command sequence)

Desktop full table: [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) · SSOT runbook: [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md).

### Integrated (preferred · ≤5 min if image pre-pulled)

```powershell
# NON-PROD ONLY — snapshot digests first
# 1) Redeploy previous integrated image/binary (embed React)
# 2) Unset FRONTEND_MODE or set FRONTEND_MODE=auto  (NOT disabled)
# 3) Point public origin at integrated process (remove Vue-only edge if sole entry)
# 4) Verify:
#    Invoke-WebRequest http://127.0.0.1:3000/healthz
#    # login via UI or:
#    pwsh -NoProfile -File scripts/e2e-web-console-login.ps1 -SkipVite
```

### Separated SPA swap

```powershell
# NON-PROD · backend stays FRONTEND_MODE=disabled
docker build -f deploy/separated/Dockerfile.frontend -t new-api-frontend:local .
# redeploy frontend service only → Nginx template unchanged
pwsh -NoProfile -File deploy/separated/smoke.ps1 -FrontendBase http://127.0.0.1:8080
```

W4 agent: **not executed** (no docker / no staging compose). Sequence documented for operator timing.

## Explicit non-goals

| Item | Status |
|------|--------|
| Production D7 flip | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched by this pack |
| Fake green when creds missing | **Forbidden** (exit 10) |
| Secrets in git / report | **Forbidden** |

## Related

| Path | Role |
|------|------|
| [th-day-e2e-harness-2026-07-24.md](./th-day-e2e-harness-2026-07-24.md) | Full E2E entry map + failure-mode catalog |
| [w4-arch-upgrade-transithub-claude.md](./w4-arch-upgrade-transithub-claude.md) | W4 report + exit table |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G1–G8 dossier (W4 evidence appended) |
| [../operations/web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | Gate definitions |
| [../../web-console/E2E.md](../../web-console/E2E.md) | Login e2e notes |

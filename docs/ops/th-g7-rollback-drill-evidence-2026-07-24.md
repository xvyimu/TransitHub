# M-TH-g7-rollback-drill · evidence · 2026-07-24

## D7 FLIP: NOT EXECUTED

Production `FRONTEND_MODE` **not** changed. No production migrate. No `git push`. No React delete (`web/default` kept). No dual React+Vue public origin. **No timed ≤5 min rollback** on this host.

## Worktree identity

| Field | Value |
|-------|--------|
| Module ID | **M-TH-g7-rollback-drill** |
| Worktree (absolute) | `C:\Users\yuanjia\orca\workspaces\src\th-g7-rollback-drill` |
| Branch | `xvyimu/th-g7-rollback-drill` |
| Tip (start / evidence base) | `f7a8b9bde34ff8c2a9b9683b1d1ad59970b6c3b0` (`docs(ops): TH E2E operator card`) |
| Tip (post evidence) | this commit |
| Agent | claude |
| Scope | G7 rollback **command sequence packaging + readonly dry-run only** · `docs/ops/` · **no** claim of timed drill green |
| Date | **2026-07-24** |
| Status | **DONE** · **in-review** (th-coord) |

## Boundary

| In | Out |
|----|-----|
| ≤5 min rollback cmd seq (Option A integrated / Option B SPA swap) | Real flip Vue → React with wall-clock ≤5 min |
| Readonly dry-run: path exists · script `-?` · doc anchors · local probes | `docker build` / compose redeploy / image pull |
| Honest G7 status + operator unblock | Production traffic · secrets in git · D7 |
| Docs under `docs/ops/` | `git push` · production `FRONTEND_MODE` · delete `web/default` |

## Pre-read

| Path | Result |
|------|--------|
| [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md) | Read — SSOT runbook Option 1/2 · ≤5 min · no DB down-migration |
| [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) | Read — Option A/B step table · W3/W4 **not executed** |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) § Rollback | Read — min cmd seq clipboard |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) § G7 | Read — status already **blocked (doc + min cmd seq)** |

---

## 1 · G7 status (authoritative for this module)

| Field | Value |
|-------|--------|
| **G7 status** | **blocked (doc + dry-run only)** |
| Timed desktop drill executed? | **No** |
| Fake green / invented wall-clock? | **Forbidden** |
| Dossier alignment | Matches [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G7 · still blocked until operator times non-prod flip |
| Unblock | Operator on **non-prod** runs Option A or B below; paste wall-clock + verification exits into dossier G7 |

> Dry-run below proves **artifacts + scripts parse** and records local listener state. It does **not** prove ≤5 min rollback.

---

## 2 · ≤5 min rollback command sequence (operator clipboard)

SSOT narrative: [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md) §B · step table: [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md).  
**NON-PROD ONLY.** Config/image only — **no SQL down-migration.** Pre-pull images for real G7 timing.

### Option A — Integrated binary/image (preferred · embed React)

| Step | Action | Est. |
|------|--------|------|
| A1 | Snapshot current digests + `FRONTEND_MODE` (Vue edge if any) | 30s |
| A2 | Redeploy previous **integrated** image/binary (embed `web/default` + `web/classic`) | 1–2 min |
| A3 | `FRONTEND_MODE` **unset** or `auto` (**not** `disabled`) | 30s |
| A4 | Single public origin → integrated process (remove/bypass separated frontend if sole edge) | 30s |
| A5 | Verify (checklist §3) | 1 min |

```powershell
# NON-PROD ONLY · Option A integrated (preferred)
# 1) Note digests + FRONTEND_MODE
# 2) Redeploy previous integrated image/binary (operator deploy path)
# 3) FRONTEND_MODE unset or auto  (NOT disabled)
# 4) Point public origin at integrated process
# 5) Verify:
Invoke-WebRequest http://127.0.0.1:3000/healthz -UseBasicParsing
Invoke-WebRequest http://127.0.0.1:3000/api/status -UseBasicParsing
# login (needs non-prod TH_E2E_* — never commit):
#   pwsh -NoProfile -File scripts/e2e-web-console-login.ps1 -SkipVite
# Prefer G2 honesty pack (exit 10 if creds missing — not fake green):
#   pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
```

**Expected total:** **≤5 min** when integrated image pre-pulled.

### Option B — Separated stack · swap SPA to React only

| Step | Action | Est. |
|------|--------|------|
| B1 | Keep backend `FRONTEND_MODE=disabled` | 0 |
| B2 | Redeploy **known-good React** frontend image **or** build: `docker build -f deploy/separated/Dockerfile.frontend -t new-api-frontend:local .` | &lt;1 min redeploy / 2–4 min build |
| B3 | Redeploy frontend service only; Nginx contract unchanged (`deploy/separated/nginx.conf.template`) | 30–60s |
| B4 | Verify | 1 min |

```powershell
# NON-PROD · Option B separated SPA swap
# Keep FRONTEND_MODE=disabled on backend
docker build -f deploy/separated/Dockerfile.frontend -t new-api-frontend:local .
# redeploy frontend service only (compose/k8s — operator path; nginx template unchanged)
pwsh -NoProfile -File deploy/separated/smoke.ps1 -FrontendBase http://127.0.0.1:8080
# optional login via edge origin after SPA is up
```

**Expected total:** **≤5 min** if React image already built; **pre-build** for real G7 (build time does not count against the 5 min flip budget if image is pre-staged).

### Post-rollback checklist (must all ☑ for G7 green)

| # | Check | Pass (this module) |
|---|-------|--------------------|
| 1 | Public origin serves **React** console shell | ☐ not timed |
| 2 | `POST /api/user/login` works (session cookie) | ☐ not timed |
| 3 | `GET /api/status` → 200 | ☐ not timed as rollback; local instant probe only (§4) |
| 4 | No SQL migration reverse required | ☑ always true for UI rollback (design invariant) |
| 5 | Wall-clock A2/B2 start → login OK ≤ **5 min** | ☐ **not measured** |

---

## 3 · Readonly dry-run (this agent host · 2026-07-24 ~14:13 +08:00)

| Field | Value |
|-------|--------|
| Host | Agent worktree machine · **not** operator staging |
| `docker` on PATH | **ABSENT** |
| `FRONTEND_MODE` (session) | **unset** |
| Stack for Option B | No separated edge on `:8080` |
| Claim | **Doc + dry-run only** · **≠** timed G7 |

### 3.1 Path presence (`Test-Path` · exit **0** all required)

| Path | Present | Kind |
|------|---------|------|
| `docs/operations/web-console-cutover-rollback.md` | **True** | file |
| `docs/ops/w3-rollback-desktop-drill.md` | **True** | file |
| `docs/ops/w4-d7-nonprod-verify.md` | **True** | file |
| `docs/ops/w3-d7-gate-dossier.md` | **True** | file |
| `deploy/separated/Dockerfile.frontend` | **True** | file |
| `deploy/separated/Dockerfile.frontend.vue` | **True** | file |
| `deploy/separated/nginx.conf.template` | **True** | file |
| `deploy/separated/smoke.ps1` | **True** | file |
| `deploy/separated/docker-compose.yml` | **True** | file |
| `scripts/e2e-web-console-login.ps1` | **True** | file |
| `scripts/w4-d7-nonprod-verify.ps1` | **True** | file |
| `web/default` | **True** | dir (React LEGACY kept — no delete) |
| `web-console` | **True** | dir |
| **path_batch** | **ALL_OK** | exit **0** |

### 3.2 Script help / param surface (parse only · no live flip)

| Command | Exit | Synopsis |
|---------|-----:|----------|
| `Get-Help ./scripts/e2e-web-console-login.ps1` | **0** | `[[-ApiBase] <string>] [[-ViteBase] <string>] [-SkipVite]` |
| `Get-Help ./deploy/separated/smoke.ps1` | **0** | `[[-FrontendBase] <string>]` |
| `Get-Help ./scripts/w4-d7-nonprod-verify.ps1` | **0** | `[[-ApiBase] <string>] [-SkipAuth] [-SkipConsoleBuild] [-SkipBackendBuild] [-SkipContract]` |

### 3.3 Doc anchors (string contains)

| File | Anchors | Result |
|------|---------|--------|
| `web-console-cutover-rollback.md` | Option 1 · Option 2 · FRONTEND_MODE · Dockerfile.frontend · ≤5 minute · no DB down-migration | **OK** all |
| `w3-rollback-desktop-drill.md` | Option A · Option B · ≤5 min · D7 FLIP: NOT EXECUTED · smoke.ps1 | **OK** all |
| `w4-d7-nonprod-verify.md` | Rollback non-prod · Integrated · Separated SPA · D7 FLIP: NOT EXECUTED | **OK** all |
| `w3-d7-gate-dossier.md` | G7 · Rollback drill · blocked | **OK** all |

### 3.4 Local probes (instant · not rollback)

| Probe | Result | Exit / note |
|-------|--------|-------------|
| `GET http://127.0.0.1:3000/healthz` | **200** `{"plane":"all","status":"ok"}` | exit **0** · integrated-ish listener present |
| `GET http://127.0.0.1:8080/frontend-healthz` | connection refused | no separated edge |
| `pwsh -NoProfile -File deploy/separated/smoke.ps1 -FrontendBase http://127.0.0.1:8080` | **passed=0 failed=6** | exit **1** (expected without edge) |
| `docker --version` / `Get-Command docker` | **ABSENT** | n/a · **blocks timed Option B build/redeploy on this host** |

### 3.5 What dry-run does **not** prove

| Claim | Status |
|-------|--------|
| ≤5 min wall-clock Vue → React | **Not proven** |
| Option A integrated redeploy | **Not run** (no operator deploy path / image digest swap) |
| Option B `docker build` + frontend redeploy | **Not run** (`docker` absent) |
| Login after rollback | **Not run** |
| G7 green | **No** |

---

## 4 · Commands + exits (recorded this session)

| # | Command | Exit | Notes |
|---|---------|-----:|-------|
| 1 | Path batch `Test-Path` (13 required artifacts) | **0** | ALL_OK |
| 2 | `Get-Help scripts/e2e-web-console-login.ps1` | **0** | param surface |
| 3 | `Get-Help deploy/separated/smoke.ps1` | **0** | param surface |
| 4 | `Get-Help scripts/w4-d7-nonprod-verify.ps1` | **0** | param surface |
| 5 | Doc anchor scan (4 files) | **0** | all needles present |
| 6 | `Invoke-WebRequest http://127.0.0.1:3000/healthz` | **0** | HTTP 200 |
| 7 | `Invoke-WebRequest http://127.0.0.1:8080/frontend-healthz` | fail | refused |
| 8 | `pwsh -NoProfile -File deploy/separated/smoke.ps1 -FrontendBase http://127.0.0.1:8080` | **1** | 0/6 pass · no edge |
| 9 | `Get-Command docker` | fail | ABSENT |
| 10 | Timed Option A/B flip | **n/a** | **blocked timed drill** |

---

## 5 · Operator unblock (staging ownership)

Owner: **operator / SRE who owns non-prod** (not this agent worktree).

1. Pre-stage known-good **integrated** React image **or** React separated image (Option A preferred).  
2. Confirm non-prod only (no production DSN / edge).  
3. Run Option A or B from §2 with stopwatch.  
4. Tick post-rollback checklist; record wall-clock + login/status exits.  
5. Paste into [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G7 Evidence row.  
6. Still require **G8** human phrase `D7 flip 现在` before any production flip — this module does **not** flip.

### Explicit out of scope

| Item | Status |
|------|--------|
| Production traffic / FRONTEND_MODE | Forbidden until G8 |
| Delete `web/default` | Forbidden |
| SQL reverse for UI rollback | Not required · do not invent |
| Secrets in git / this evidence | Forbidden |

---

## 6 · Intentionally not done

| Item | Status |
|------|--------|
| Timed ≤5 min rollback | **Not run** |
| Marking G7 **green** | **Not done** (would be fake green) |
| `docker build` React/Vue images | **Not run** (docker absent) |
| **D7 production flip** | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched |
| `git push` / merge default branch | Not done |

## 7 · Outcome

| Claim | Evidence |
|-------|----------|
| ≤5 min Option A + B cmd seq packaged | §2 |
| Readonly dry-run exits recorded | §3 · §4 |
| Explicit **blocked (doc + dry-run only)** | §1 |
| Timed drill **not** claimed | §3.5 · §6 |
| D7 / push / FRONTEND_MODE / React delete | **NOT EXECUTED** / no push / untouched / `web/default` present |

## Related

| Path | Role |
|------|------|
| [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md) | Operator rollback SSOT |
| [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) | Desktop step table |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) | Min cmd seq + nonprod pack |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G7 dossier status |
| [deploy/separated/smoke.ps1](../../deploy/separated/smoke.ps1) | Separated edge smoke |
| [scripts/e2e-web-console-login.ps1](../../scripts/e2e-web-console-login.ps1) | Login probe (not G2 sole authority) |
| [scripts/w4-d7-nonprod-verify.ps1](../../scripts/w4-d7-nonprod-verify.ps1) | Honest G2/G3 pack |

## Handoff · th-coord

- **Status:** DONE + **in-review**
- **Ask:** accept G7 evidence as **doc + dry-run packaging**; keep G7 **blocked (doc + dry-run only)** until operator times non-prod Option A/B
- **Do not:** D7 · push · production FRONTEND_MODE · treat this as timed-drill green · delete `web/default`

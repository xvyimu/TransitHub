# TH-DAY · E2E harness map + failure modes · **D7 FLIP: NOT EXECUTED**

> 目标：插上**非生产** `TH_E2E_*` 后，按本文直接绿 W4 verify（G2/G3 live）。  
> **禁止**：生产 flip、`FRONTEND_MODE` 改生产、push 本分支到远程、把真实密码写进 git/本文件。  
> 继承：[`th-w1-d7-prereq-2026-07-23.md`](./th-w1-d7-prereq-2026-07-23.md) · 编排器：`scripts/w4-d7-nonprod-verify.ps1`。

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-day-e2e-harness`（本机路径，可移植性无保证） |
| Branch | `xvyimu/th-day-e2e-harness` |
| Base tip | `4da227b4` (TH-W1 D7 prereq runbook) |
| Date | **2026-07-24** |
| Agent host (this run) | healthz **200** · `TH_E2E_*` **unset** · docker **absent** |
| Gate card | [`th-e2e-gate-card.md`](./th-e2e-gate-card.md) |

---

## 1. E2E entry map

| Entry | Path | Role | Auth | Skip / dry flags | Prefer for |
|-------|------|------|------|------------------|------------|
| **W4 verify pack (orchestrator)** | `scripts/w4-d7-nonprod-verify.ps1` | healthz · status · contract · **login** · **channels RO** · optional builds | Requires `TH_E2E_USER` **and** `TH_E2E_PASS` (no silent root) | `-SkipAuth` · `-SkipContract` · `-SkipConsoleBuild` · `-SkipBackendBuild` | **G2/G3 gates** · default path |
| Login-only e2e (legacy) | `scripts/e2e-web-console-login.ps1` | healthz · login · self · optional Vite proxy | Defaults `root`/`123456` if env unset (**false-green risk** on empty DB only) | `-SkipVite` | Cross-check after W4 green |
| OpenAPI contract | `scripts/validate-console-contract.py` | Static `console-subset.yaml` ops/schemas | None | n/a | G3 **contract** (no live) |
| Logs live smoke | `scripts/smoke-logs.ps1` | `/api/log/` + self (T-TH-003) | `TH_ACCESS_TOKEN`+`TH_USER_ID` **or** `TH_E2E_*` | n/a | Optional; **not** G2 |
| Separated edge smoke | `deploy/separated/smoke.ps1` | frontend-healthz · SPA · proxy `/api/status` · 401/metrics | None (edge) | `-FrontendBase` | G4/separated stack only |
| Three-dialect migrate | `scripts/migrate-three-dialect.ps1` | Empty-DB migrate matrix | DSN envs (see script) | dry-run variants in sibling scripts | **Not** D7 traffic |

**Prefer W4 pack** over legacy login e2e: missing creds → exit **10** (actionable block), never silent default root.

---

## 2. Env names (no secrets)

| Variable | Used by | Required when | Default if unset | Notes |
|----------|---------|---------------|------------------|-------|
| `TH_API_BASE` | W4 · login e2e · smoke-logs | always (probes) | `http://127.0.0.1:3000` | Non-prod only |
| `TH_E2E_USER` | W4 · login e2e · smoke-logs (password path) | G2/G3 live | *(none in W4)* / `root` in legacy | Whitespace trimmed; empty → treated unset |
| `TH_E2E_PASS` | same | G2/G3 live | *(none in W4)* / `123456` in legacy | **Never** commit; never log value |
| `TH_VITE_BASE` | login e2e only | optional | `http://127.0.0.1:5173` | WARN+continue if down |
| `TH_ACCESS_TOKEN` | smoke-logs | token path | — | Alt: `TH_SMOKE_TOKEN` |
| `TH_USER_ID` | smoke-logs | with access token | — | Alt: `TH_SMOKE_USER_ID` · `New-Api-User` |
| `TH_SMOKE_USER` / `TH_SMOKE_PASS` | smoke-logs | alt to `TH_E2E_*` | — | Same non-prod rules |

Mint checklist (operators): [`w2-cutover-e2e-credentials.md`](./w2-cutover-e2e-credentials.md).

---

## 3. Dependency services

| Dependency | Why | How to know it's up | If down |
|------------|-----|---------------------|---------|
| **Backend API** (non-prod) | All probes | `GET {TH_API_BASE}/healthz` → 200 | W4 exit **3** |
| **DB** (SQLite/MySQL/PG behind API) | login / channels | `/api/status` JSON `setup` field; login success | login **1** or status **6** |
| **Redis** (if configured) | sessions / rate limit on some deploys | product-specific; not asserted by W4 pack | may surface as login/self failures |
| **Vite** `5173` | optional proxy smoke | only login e2e without `-SkipVite` | WARN, not fail |
| **pnpm + node_modules** | console build step | `-SkipConsoleBuild` to skip | exit **5** if build selected and fails |
| **Go toolchain** | `frontend_external` build | `-SkipBackendBuild` to skip | exit **5** |
| **Python 3** | contract validator | `python scripts/validate-console-contract.py` | exit **5** if selected |
| **Docker** | G4 Vue image local | `docker version` | local G4 blocked; CI SSOT |

This agent host (2026-07-24): API **up** · docker **absent** · `TH_E2E_*` **unset**.

---

## 4. Flags (mock / dry-run / skip) — honest semantics

| Flag / mode | Script | What it skips | Exit impact | Fake green? |
|-------------|--------|---------------|-------------|-------------|
| `-SkipAuth` | W4 | login + channels | forces **10** if nothing else failed | **No** — exit 10 by design |
| `-SkipContract` | W4 | OpenAPI validator | no exit 5 from contract | only if auth green + rest green |
| `-SkipConsoleBuild` | W4 | `pnpm build` in `web-console/` | no exit 5 from console | use after prior green build |
| `-SkipBackendBuild` | W4 | `go build -tags frontend_external` | no exit 5 from Go | same |
| `-SkipVite` | login e2e | Vite proxy check | none (was WARN only) | n/a |
| *(no dry-run for G2)* | — | — | — | **Do not** invent a “pass without login” mode |

**Rule:** skip flags never turn missing production-grade checks into exit **0**. Auth absence / `-SkipAuth` → **10**.

---

## 5. Failure mode catalog

### 5.1 W4 pack (`w4-d7-nonprod-verify.ps1`)

| Failure mode | How to reproduce (non-prod) | Expected exit | Log / SUMMARY features | Operator fix |
|--------------|----------------------------|--------------:|------------------------|--------------|
| **Missing both creds** | unset `TH_E2E_USER` + `TH_E2E_PASS` | **10** | `BLOCK credentials incomplete — missing: TH_E2E_USER + TH_E2E_PASS` · `login=10 channels=10` · `SUMMARY exit=10` | Export both env vars (non-prod) |
| **Partial creds (user only)** | set USER, unset PASS | **10** | `missing: TH_E2E_PASS` · “USER is set but PASS empty” | Set `TH_E2E_PASS` |
| **Partial creds (pass only)** | set PASS, unset USER | **10** | `missing: TH_E2E_USER` | Set `TH_E2E_USER` |
| **Whitespace-only creds** | `TH_E2E_USER='   '` | **10** | treated as unset (trimmed) | Real non-empty values |
| **`-SkipAuth`** | any host | **10** | `login=skip channels=skip` · “Exit 10 by design” | Remove flag + set creds for G2/G3 |
| **Wrong password / banned** | wrong `TH_E2E_PASS` | **1** | `FAIL login rejected (success=false)` · `login=1 channels=skip` · “Exit 1 = wrong password… NOT exit 10” | Mint/reset non-prod admin |
| **Login HTTP error** | bad path / proxy 502 | **1** | `FAIL login request:` | Check `TH_API_BASE`, reverse proxy |
| **Self after login fails** | cookie stripped / missing `New-Api-User` | **2** | `self=2` · `channels=skip` | Session cookie same-origin; header from login `data.id` |
| **Backend down (refused)** | `TH_API_BASE=http://127.0.0.1:3999` | **3** | `FAIL backend unreachable` · CONNECTION REFUSED hint · `SUMMARY exit=3 healthz=3` | Start API or fix base URL |
| **Backend timeout** | firewalled / hung listener | **3** | TIMEOUT symptom hint | Process health, firewall |
| **Channels RO fail / no role** | user lacks Admin/ChannelRead | **4** | `channels=4` · need AdminAuth + ChannelRead | Elevate non-prod e2e user |
| **Key material on list** | regression leaking `key` | **4** | `channels list appears to include key material` | **Bug** — fix API; do not weaken check |
| **Contract / build fail** | broken OpenAPI or build | **5** | `contract=…` / `console_build=5` / `backend_build=5` | Fix sources; re-run |
| **`/api/status` fail** | routing broken while healthz up | **6** | `status=6` | Backend routing / readiness |
| **All selected green** | full pack + valid non-prod creds | **0** | `login=0 channels=0 …` · still “NOT a production D7 flip” | Proceed to G4/G6/G7 + **human G8** only |

### 5.2 Legacy login e2e (`e2e-web-console-login.ps1`)

| Failure mode | Expected exit | Log feature | Notes |
|--------------|--------------:|-------------|-------|
| Default root on shared DB | **1** | `ERR login failed` · `success:false` · message like 用户名或密码错误… | **Expected** when env unset on seeded DB |
| Env unset (empty DB only) | **0** possible | `WARN using default root/123456` | Prefer W4 pack |
| Backend down | **3** | `ERR backend unreachable` | |
| Self fail | **2** | `ERR self` | |

### 5.3 Logs smoke (`smoke-logs.ps1`)

| Failure mode | Expected exit | Log feature |
|--------------|--------------:|-------------|
| No token and no `TH_E2E_*` | **4** | `RESULT blocked-auth` · `WARN no TH_ACCESS_TOKEN…` |
| Token without `TH_USER_ID` | **4** | missing New-Api-User |
| Auth fail | **1** | login / self fail |
| Backend down | **3** | status/healthz |

### 5.4 Separated smoke (`deploy/separated/smoke.ps1`)

| Failure mode | Expected exit | Notes |
|--------------|--------------:|-------|
| Any check fail | **1** | `passed=N failed=M` · needs frontend base (default `:8080`) |
| All pass | **0** | Edge only — not G2 |

---

## 6. Recorded run (this session · 2026-07-24)

| Command | Exit | Redacted summary |
|---------|-----:|------------------|
| `pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild` | **10** | `healthz=0 status=0 contract=0 login=10 channels=10 console_build=skip backend_build=skip` |
| same + `-SkipAuth` | **10** | `login=skip channels=skip` |
| `python scripts/validate-console-contract.py` | **0** | 8 ops / 8 schemas |
| `pwsh -NoProfile -File scripts/e2e-web-console-login.ps1 -SkipVite` | **1** | default root rejected (`success:false`) |
| `pwsh -NoProfile -File scripts/smoke-logs.ps1` | **4** | blocked-auth |
| Partial USER only → W4 | **10** | missing `TH_E2E_PASS` (after harness improve) |
| `TH_API_BASE=http://127.0.0.1:3999` → W4 | **3** | connection refused |

**Honest outcome without creds:** exit **10**. G2/G3 live **not** green. **D7 FLIP: NOT EXECUTED.** **No push.**

---

## 7. Next session — inject non-prod creds → green path

```powershell
# From repo root (this worktree or D:\TransitHub\src)
$env:TH_API_BASE = 'http://127.0.0.1:3000'   # or staging — NON-PROD only
$env:TH_E2E_USER = '<non-prod-admin>'
$env:TH_E2E_PASS = '<non-prod-secret>'       # never commit

# Fast re-check (builds already trusted):
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
# Need: SUMMARY exit=0  and login=0 channels=0

# Full pack (first time / after console or Go changes):
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1
```

Green W4 pack still **≠** D7. Remaining: G4 image · G6 soak · G7 rollback timed · **G8 human** phrase `D7 flip 现在`.

---

## 8. Harness improvements landed this wave

| Change | File | Why |
|--------|------|-----|
| Distinguish missing USER vs PASS vs both; trim whitespace | `scripts/w4-d7-nonprod-verify.ps1` | Partial env was opaque “not both set” |
| Clearer exit-10 vs exit-1 messaging; pointer to this doc | same | Operators confuse “no creds” with “wrong password” |
| healthz refused vs timeout hints | same | Faster ops triage |
| Legacy e2e WARN when defaulting root | `scripts/e2e-web-console-login.ps1` | Reduce silent false path |
| This catalog | `docs/ops/th-day-e2e-harness-2026-07-24.md` | Single map for next session |

**Not changed:** product D7 behavior, production config, remote push.

---

## 9. Explicit non-goals

| Item | Status |
|------|--------|
| Production D7 flip | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched |
| `git push` | Not done |
| Secrets in repo / this file | **Forbidden** |
| Fake green when creds missing | **Forbidden** (exit 10) |
| Auto-skip G2 on agent hosts | **Forbidden** |

---

## 10. Related

| Path | Role |
|------|------|
| [th-e2e-gate-card.md](./th-e2e-gate-card.md) | **一页纸** exit · 入口 · G2/G3≠D7 |
| [th-w1-d7-prereq-2026-07-23.md](./th-w1-d7-prereq-2026-07-23.md) | Prior prereq runbook |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) | Pack how-to |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | Mint non-prod account |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G1–G8 dossier |
| [../operations/web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | Gate definitions |
| [../../web-console/E2E.md](../../web-console/E2E.md) | Login e2e notes |

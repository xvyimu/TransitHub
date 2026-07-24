# M-TH-g2-e2e-nonprod · evidence · 2026-07-24

> **Module:** M-TH-g2-e2e-nonprod  
> **Product:** TransitHub · non-prod G2 gate only  
> **D7 FLIP: NOT EXECUTED** · no production `FRONTEND_MODE` change · no push

## Env (names only — no values)

| Variable | State |
|----------|--------|
| `TH_E2E_USER` | **unset** |
| `TH_E2E_PASS` | **unset** |
| `TH_API_BASE` | default `http://127.0.0.1:3000` (not overridden in session) |

**Secrets:** none printed · none written to this file · none committed.

## Command + exit

```powershell
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
```

| Field | Value |
|-------|--------|
| **Exit code** | **10** |
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-g2-e2e-nonprod` |
| Branch | `xvyimu/th-g2-e2e-nonprod` |
| Tip at run | `f7a8b9bd` (pre-evidence commit) |
| Date | 2026-07-24 |

### Step exits (from SUMMARY)

| Step | Exit / result |
|------|----------------|
| healthz | **0** (HTTP 200) |
| `/api/status` | **0** (HTTP 200) |
| console-subset contract | **0** (`validate-console-contract.py`) |
| login (G2) | **10** (credentials incomplete) |
| channels RO (G3 live) | **10** (blocked with login) |
| web-console build | **skip** (`-SkipConsoleBuild`) |
| backend build | **skip** (`-SkipBackendBuild`) |

SUMMARY line:

```
SUMMARY exit=10  healthz=0 status=0 contract=0 login=10 channels=10 console_build=skip backend_build=skip
```

Script banner: `D7 FLIP: NOT EXECUTED (this script never changes production FRONTEND_MODE)`.

## blocked · G2 (no fake green)

| Gate | Status | Notes |
|------|--------|-------|
| **G2** Login e2e (non-prod) | **blocked** | `TH_E2E_USER` + `TH_E2E_PASS` unset; orchestrator refuses silent default root |
| **G3** Channels RO live | **blocked** | depends on G2 session |
| G3 contract (static) | green this run | contract step exit 0 only — **not** live G3 |
| **D7** production flip | **NOT EXECUTED** | exit 10 / docs commit ≠ D7 |

**Honest outcome:** exit **10** = actionable block, **not** pass. Do **not** treat healthz/status/contract green as G2/G3 green.

**Not used for G2 judgment:** `scripts/e2e-web-console-login.ps1` default `root`/`123456` (legacy; can false-green or fail on shared DBs). G2/G3 sole green path: `scripts/w4-d7-nonprod-verify.ps1` with real non-prod `TH_E2E_*`.

## Unblock (operator · non-prod only)

1. Mint non-prod admin — `docs/ops/w2-cutover-e2e-credentials.md`
2. Session only:
   ```powershell
   $env:TH_E2E_USER = '<non-prod-admin>'
   $env:TH_E2E_PASS = '<non-prod-secret>'   # never commit / never log
   ```
3. Re-run:
   ```powershell
   pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
   ```
4. Expect login=0 channels=0 overall **0** only when authed steps green — still **≠ D7** (need G4/G6/G7/G8 human).

## Explicit non-goals (this module)

| Item | Status |
|------|--------|
| Production `FRONTEND_MODE` / traffic flip | **not touched** |
| Push to remote | **not done** |
| Write `TH_E2E_PASS` to any file | **forbidden / not done** |
| D7 flip | **NOT EXECUTED** |
| Claim G2 green without `TH_E2E_*` | **forbidden** (exit 10) |

## Status line

```
TH-E2E: blocked · D7 NOT flipped · no push
```

Detail: `W4 exit 10 · healthz=0 status=0 contract=0 login=10 channels=10 · G2 blocked (no TH_E2E_*)`

## Related

| Path | Role |
|------|------|
| [th-e2e-gate-card.md](./th-e2e-gate-card.md) | One-pager exit / G2≠D7 |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) | Pack how-to |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | Cred mint (names only) |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G1–G8 dossier (GATE-MATRIX absent → this file) |
| `scripts/w4-d7-nonprod-verify.ps1` | Sole G2/G3 orchestrator |

## Agent close

- **DONE** · **in-review** (for th-coord)
- Evidence path: `docs/ops/th-g2-e2e-nonprod-evidence-2026-07-24.md`
- Stack lock: no prod FRONTEND_MODE · no delete `web/default` · no dual-write

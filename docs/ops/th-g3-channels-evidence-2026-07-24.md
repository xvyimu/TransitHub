# M-TH-g3-channels · evidence · 2026-07-24

> **Module:** M-TH-g3-channels  
> **Product:** TransitHub · cutover **G3** (Channels RO · key-omission) only  
> **D7 FLIP: NOT EXECUTED** · no production `FRONTEND_MODE` change · no push · no secrets written

## Worktree identity

| Field | Value |
|-------|--------|
| Module ID | **M-TH-g3-channels** |
| Worktree (absolute) | `C:\Users\yuanjia\orca\workspaces\src\th-g3-channels` |
| Branch | `xvyimu/th-g3-channels` |
| Tip (start / pre-evidence) | `f7a8b9bd` (`docs(ops): TH E2E operator card`) |
| Tip (post evidence) | this commit |
| Agent | claude |
| Date | **2026-07-24** (~13:40 +0800) |
| Status | **DONE** · **in-review** (th-coord) |
| Scope | `docs/ops/` evidence + contract / W4 pack re-run · **no** production · **no** secrets |

## Boundary

| In | Out |
|----|-----|
| `python scripts/validate-console-contract.py` | Fake-green with default `root`/`123456` |
| W4 pack channels RO step (needs auth) | Live claim of G3 without `TH_E2E_*` |
| Cutover-plan G3对照（列表可用、keys 不出现） | D7 / `FRONTEND_MODE` / delete `web/default` |
| Docs under `docs/ops/` | `git push` · production traffic · dual-write |

## Pre-read

| Path | Result |
|------|--------|
| `docs/ops/th-g2-e2e-nonprod-evidence-2026-07-24.md` | **ABSENT** in this worktree; read sibling copy at `th-coord/docs/ops/th-g2-e2e-nonprod-evidence-2026-07-24.md` — G2 **blocked** (no `TH_E2E_*`), W4 exit **10** |
| `docs/gateway/CONSOLE_API_CONTRACT.md` | Read — §3 `GET /api/channel/` AdminAuth + ChannelRead · **Omit("key")** + `clearChannelInfo` · keys not on list |
| `docs/ops/th-e2e-gate-card.md` | Read — G3 live ≠ D7; exit **10** when creds missing |
| `docs/operations/web-console-cutover-plan.md` | G3 = **Channels RO usable** · `/channels` lists **without keys** |

## Env (names only — no values)

| Variable | State |
|----------|--------|
| `TH_E2E_USER` | **unset** |
| `TH_E2E_PASS` | **unset** |
| `TH_API_BASE` | default `http://127.0.0.1:3000` (not overridden) |

**Secrets:** none printed · none written to this file · none committed.  
**Legacy default root:** **not used** for G3 judgment (would risk false green / wrong DB).

## Commands + exits (this message)

### 1) OpenAPI contract (static · G3 contract only)

```powershell
python scripts/validate-console-contract.py
```

| Field | Value |
|-------|--------|
| **Exit code** | **0** |
| ops | 8 (incl. `GET /api/channel/`) |
| schemas | 8 |

```text
file: docs\openapi\console-subset.yaml
ops_found=8 schemas_found=8
  OK GET /api/channel/
  OK GET /api/status
  OK GET /api/user/logout
  OK GET /api/user/self
  OK GET /healthz
  OK GET /livez
  OK GET /readyz
  OK POST /api/user/login
PASS validate-console-contract
```

### 2) W4 pack (sole G2/G3 live orchestrator)

```powershell
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
```

| Field | Value |
|-------|--------|
| **Exit code** | **10** |
| Banner | `D7 FLIP: NOT EXECUTED` |

| Step | Exit / result |
|------|----------------|
| healthz | **0** (HTTP 200) |
| `/api/status` | **0** (HTTP 200) |
| console-subset contract | **0** |
| login (G2) | **10** (credentials incomplete) |
| **channels RO (G3 live · key-omission)** | **10** (blocked with login — **not** exercised live) |
| web-console build | **skip** |
| backend build | **skip** |

SUMMARY:

```text
SUMMARY exit=10  healthz=0 status=0 contract=0 login=10 channels=10 console_build=skip backend_build=skip
```

Script message (honest block, not pass):

```text
BLOCK credentials incomplete — missing: TH_E2E_USER + TH_E2E_PASS (login + channels RO blocked)
...
Exit code 10 = actionable block, NOT pass. Do not treat as G2/G3 green.
```

## Cutover G3对照

| Cutover criterion (`web-console-cutover-plan.md`) | This run |
|---------------------------------------------------|----------|
| G3 Channels RO usable — list without keys | **Live blocked** — depends on G2 session (`TH_E2E_*`) |
| Contract documents list + key omission | **green** — `CONSOLE_API_CONTRACT.md` §3 + `console-subset.yaml` + validator exit **0** |
| Implementation still omits key on list path | Code path present: `controller/channel_list.go` uses `.Omit("key")` on list queries (not re-executed live) |
| W4 pack key-leak detector | Present in `scripts/w4-d7-nonprod-verify.ps1` (items.key + `"key":"..."` heuristic → exit **4**); **not reached** without auth |

## Gate matrix (this module)

| Gate | Status | Notes |
|------|--------|-------|
| **G3 contract** (static OpenAPI / docs) | **green** | validator exit **0** |
| **G3 live** (authenticated `GET /api/channel/` + key-omission) | **blocked · G3 live blocked (depends on G2)** | `TH_E2E_*` unset → W4 exit **10** · `channels=10` |
| **G2** | blocked (upstream) | same exit 10 · see G2 evidence pattern |
| **D7** | **NOT EXECUTED** | docs + exit 10 ≠ production flip |

**Honest outcome:** do **not** treat healthz / status / contract green as G3 live green.  
**G3 live blocked** until operator sets non-prod `TH_E2E_USER` + `TH_E2E_PASS` and re-runs W4 pack to `login=0 channels=0` (still **≠ D7**).

## Unblock (operator · non-prod only)

1. Mint non-prod admin with AdminAuth + ChannelRead — `docs/ops/w2-cutover-e2e-credentials.md`
2. Session only (never commit):
   ```powershell
   $env:TH_E2E_USER = '<non-prod-admin>'
   $env:TH_E2E_PASS = '<non-prod-secret>'
   ```
3. Re-run:
   ```powershell
   pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
   ```
4. Expect `channels=0` only when list succeeds **and** no key material — overall **0** still needs G4/G6/G7/G8 before any D7 human gate.

## Explicit non-goals (this module)

| Item | Status |
|------|--------|
| Production `FRONTEND_MODE` / traffic flip | **not touched** |
| Push to remote | **not done** |
| Write `TH_E2E_PASS` / secrets to any file | **forbidden / not done** |
| Default root login for G3 green | **not used** |
| D7 flip | **NOT EXECUTED** |
| Claim G3 live green without auth | **forbidden** (exit 10) |

## Status line

```
TH-E2E: blocked · G3 live blocked (depends on G2) · contract green · D7 NOT flipped · no push
```

Detail: `W4 exit 10 · healthz=0 status=0 contract=0 login=10 channels=10 · G3 contract=0 · G3 live=blocked`

## Related

| Path | Role |
|------|------|
| [th-e2e-gate-card.md](./th-e2e-gate-card.md) | One-pager exit / G2≠D7 |
| [th-e2e-operator-card-2026-07-25.md](./th-e2e-operator-card-2026-07-25.md) | Operator mint + re-run |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) | Pack how-to |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | Cred mint (names only) |
| [../gateway/CONSOLE_API_CONTRACT.md](../gateway/CONSOLE_API_CONTRACT.md) | §3 channels RO + key omission |
| [../operations/web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | G3 gate definition |
| `scripts/w4-d7-nonprod-verify.ps1` | Sole G2/G3 orchestrator |
| `scripts/validate-console-contract.py` | G3 contract only |

## Agent close

- **DONE** · **in-review** (for th-coord)
- Evidence path: `docs/ops/th-g3-channels-evidence-2026-07-24.md`
- Stack lock: no prod FRONTEND_MODE · no delete `web/default` · no dual-write
- **D7 NOT EXECUTED**

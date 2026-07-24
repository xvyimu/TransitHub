# G8 · Human flip checklist（**仅提示 · 本总控不执行 flip**）

> 本文件是 **等人 gate** 清单。  
> **没有**人类原话 `D7 flip 现在`（或 cutover-plan 等价授权）时，**禁止**改生产 `FRONTEND_MODE`、切流量、删 `web/default`。  
> 文档齐 / PR 合 / 非生产 exit 0 **都不是** D7。

## Before you say the flip phrase

| # | Prerequisite | Status to confirm yourself |
|---|--------------|----------------------------|
| 1 | **G1** Module2 on production tip | □ green on the commit you will deploy |
| 2 | **G2** Non-prod login e2e | □ `w4-d7-nonprod-verify.ps1` exit **0** with real non-prod `TH_E2E_*` (not default root) |
| 3 | **G3** Channels RO live | □ pack channels=0 · keys omitted |
| 4 | **G4** Vue image | □ CI digest **or** local docker build recorded |
| 5 | **G5** `frontend_external` binary/image | □ exit 0 on deploy artifact |
| 6 | **G6** Staging soak ≥24h | □ checklist filled · no dual public React+Vue |
| 7 | **G7** Rollback drill | □ timed ≤5 min on non-prod · runbook open |
| 8 | DB backup + current image digests noted | □ |
| 9 | On-call / rollback owner named | □ |
| 10 | You will say exactly: **`D7 flip 现在`** (or agreed equivalent) | □ |

## After you authorize (operators — not this coord session)

1. Follow `docs/operations/web-console-cutover-plan.md` cutover steps.  
2. Keep `docs/operations/web-console-cutover-rollback.md` ready (≤5 min).  
3. Smoke: `/frontend-healthz`, `/healthz`, login, health, channels RO.  
4. Do **not** delete React in the same window as first flip.

## Coord stance this wave

| Item | Value |
|------|--------|
| Flip readiness | **NO** until rows 1–9 green **and** row 10 said by human |
| Auto-flip by agent | **Never** |
| This long wave default | Evidence only · **D7 NOT EXECUTED** |

Updated: 2026-07-24 · th-long-wave G0=D

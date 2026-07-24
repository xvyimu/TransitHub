# TransitHub · Long Wave · Progress

> **Coord session:** th-coord (`xvyimu/th-coord`) · agent Claude · 2026-07-24  
> **Product root:** `D:\TransitHub\src` · this worktree: `C:\Users\yuanjia\orca\workspaces\src\th-coord`  
> **D7 FLIP: NOT EXECUTED** · no production `FRONTEND_MODE` · no push · no delete `web/default`

## Status

| Field | Value |
|-------|--------|
| Phase | **dispatched** (G0=D authorized · Phase2 workers) |
| G0 | **D = A+C non-prod** (human 2026-07-24) |
| North star | 网关稳定 + web-console 质量；新能力只进 console；生产默认 React 直至 D7 |
| Flip readiness | **NO** — G8 checklist only: [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md) |
| Fleet | 2 child wt dispatched · no Agent×N |

## Worktree inventory (TransitHub)

| displayName | path | branch | role | action |
|-------------|------|--------|------|--------|
| **main** | `D:\TransitHub\src` | `main` | product root | keep |
| **th-coord** | `…\src\th-coord` | `xvyimu/th-coord` | **本总控** | active |
| **th-coord-d7** | `…\src\th-coord-d7` | `xvyimu/th-coord-d7` | sibling D7 coord | freeze stack; do not duplicate agents |
| **th-console-quality** | (Orca create) | `xvyimu/th-console-quality` | M-TH-console-quality | **dispatched** |
| **th-backend-stable-scout** | (Orca create) | `xvyimu/th-backend-stable-scout` | M-TH-backend-stable-scout | **dispatched** |

Historical (not live unless recreated): w1–w4 th-claude, wave8-th-codex, th-d7-scout (branch `xvyimu/th-d7-scout` @ `35194ff3` on origin — inherit docs only).

**Hygiene:** create → list → keep agent only; close non-agent; never stop `name:orca` / `path:D:\orca`. Child DONE → receipt → dirty commit → close.

## Inherited evidence (no invent)

| Pack | Path | Key exits / note |
|------|------|------------------|
| W3 dossier | `docs/ops/w3-d7-gate-dossier.md` | G1/G5 green; rest blocked written |
| W4 pack | `scripts/w4-d7-nonprod-verify.ps1` · `docs/ops/w4-*` | exit **10** no creds; contract 0; frontend_external 0; console quality 0 |
| Scout | origin `xvyimu/th-d7-scout` @ `35194ff3` · `docs/ops/th-d7-scout-2026-07-24.md` | 2026-07-24 re-probe same blockers |
| E2E card | `docs/ops/th-e2e-gate-card.md` | exit 10 ≠ pass |
| Cutover plan | `docs/operations/web-console-cutover-plan.md` | G1–G8 SSOT |

### GATE snapshot (post-refresh · G0=D)

| Gate | Status |
|------|--------|
| G1 / G5 | **green** |
| G2 | **blocked** (no `TH_E2E_*`) |
| G3 | **blocked** live · contract **green** |
| G4 | **blocked** local · CI SSOT |
| G6 / G7 / G8 | **blocked** |

Detail: [GATE-MATRIX.md](./GATE-MATRIX.md).

## Stack lock

| Layer | Lock |
|-------|------|
| Backend | Go · Gin · GORM · 3-DB · JSON `common/json.go` · AGPL |
| New UI | only `web-console/` |
| Prod default | React `web/default` until D7 |
| This wave | **no** go.mod major bump · **no** D7 · **no** dual-write |

## Dispatched workers

### M-TH-console-quality → `th-console-quality`

- Boundary: `web-console/` · evidence under `docs/ops/` · may re-run W4 pack (honest exits)
- Do: `pnpm install --frozen-lockfile` · typecheck · test · build · debt list · evidence md + commit
- Don't: FRONTEND_MODE · delete React · dual-write · push · secrets

### M-TH-backend-stable-scout → `th-backend-stable-scout`

- Boundary: read-only audit of `migrations/` · DB pool/timeout · Redis hot paths
- Do: tables + findings evidence · optional `go test`/`go build` on touch surface · commit docs only preferred
- Don't: go.mod major bump · production DSN · flip · push

## G8

Only [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md). Coord **will not** flip.

## Log

| Time | Event |
|------|--------|
| 2026-07-24 | Phase0 scout + G0 wait |
| 2026-07-24 | **G0=D** · GATE refresh · G8 checklist · dispatch th-console-quality + th-backend-stable-scout |

# TransitHub · Long Wave · Progress

> **Coord:** th-coord · Claude · 2026-07-24  
> **D7 FLIP: NOT EXECUTED** · no production FRONTEND_MODE · no push · no delete web/default

## Status

| Field | Value |
|-------|--------|
| Phase | **week cadence · W1 reviewed · W2+W5 live · W1 wt rm blocked (PTY)** |
| G0 | **D = A+C non-prod** |
| Horizon | ~7 days · [WEEK-BACKLOG.md](./WEEK-BACKLOG.md) |
| Flip readiness | **NO** · [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md) |
| Live cap | ≤3 **active agents** · W1 shells FROZEN (orca rm PTY fail) |

## Worktree inventory (TH)

| displayName | path | role | action |
|-------------|------|------|--------|
| main | `D:\TransitHub\src` | product root | keep |
| th-coord | `…\th-coord` | **总控** | active |
| th-console-quality | `…\th-console-quality` | W1a DONE | **FROZEN** · stop ok · `worktree rm --force` fails PTY — no new agent |
| th-backend-stable-scout | `…\th-backend-stable-scout` | W1b DONE | **FROZEN** · same PTY rm debt |
| th-g2-e2e-nonprod | `…\th-g2-e2e-nonprod` | W2 | **live** agent |
| th-g5-backend-regression | `…\th-g5-backend-regression` | W5 | **live** agent |

## GATE snapshot

| Gate | Status |
|------|--------|
| G1 | **green** |
| G2 | **blocked** (no TH_E2E_*) |
| G3 | **blocked** live · contract **green** |
| G4 | **blocked** local · CI SSOT |
| G5 | **green** inherit · W5 refresh |
| G6 / G7 / G8 | **blocked** |

## W1 harvest (reviewed)

| Module | Commit | Exits (key) |
|--------|--------|-------------|
| console-quality | `4afcf5b3` → coord `907eaa6b` | pnpm ×4 **0** · W4 pack **10** |
| backend-stable-scout | `d1dd3278` → coord `0973f5d3` | model/common tests **0** · migrate sqlite **0** · bare go build **1** (classic dist) |

## Stack lock

Go · Gin · GORM · 3-DB · JSON common/json.go · AGPL · new UI only web-console · React LEGACY until D7 · **no** go.mod major this week.

## Log

| Time | Event |
|------|--------|
| 2026-07-24 | Phase0 · G0 wait |
| 2026-07-24 | G0=D · dispatch W1a/W1b |
| 2026-07-24 | **Week mode** · W1 review PASS · WEEK-BACKLOG · open W2+W5 · W1 rm PTY debt (FROZEN) |

# TransitHub · Long Wave · Progress

> **D7 NOT EXECUTED** · 2026-07-24

## Status

| Field | Value |
|-------|--------|
| Phase | **W3/W4/W6 harvested · W7/W8/W9 dispatch** |
| G0 | D = A+C non-prod |
| Flip | **NO** |
| Live agents | ≤3 (W7 W8 W9) |

## Harvest (latest)

| Module | Tip | Key |
|--------|-----|-----|
| W3 G3 | `4daf0ba9` → coord `600e5ed3` | contract **0** · live **blocked** |
| W4 G4 | `4c2560bf` → coord `f022701e` | docker **absent** · CI SSOT |
| W6 G6 | `f4669be9` → coord `4fed4626` | full soak **not run** |

## GATE

G1/G5 green · G2/G3 live/G4/G6/G7/G8 blocked (honest) · G3 contract green

## Fleet now

| wt | action |
|----|--------|
| th-coord | active |
| th-g3/g4/g6 | **closing** |
| th-g7-rollback-drill · th-legacy-gate-scan · th-be-migrate-3db | **open** |

## Log

| Time | Event |
|------|--------|
| 2026-07-24 | Force continue: harvest W3–W6 · open W7–W9 · feature push allowed |

# TransitHub · Long Wave · GATE Matrix

> **D7 FLIP: NOT EXECUTED** · G0=D · updated **2026-07-24** (W3/W4/W6 harvest)

| Gate | Status | Evidence | Unblock |
|------|--------|----------|---------|
| **G1** | **green** | Module2 + console quality | — |
| **G2** | **blocked** | `th-g2-e2e-nonprod-evidence` · pack exit **10** | `TH_E2E_*` |
| **G3** | **blocked** live · contract **green** | `th-g3-channels-evidence` · contract **0** · live **10** | after G2 |
| **G4** | **blocked** local · **CI SSOT** | `th-g4-image-repro-evidence` · docker absent | Docker or CI digest |
| **G5** | **green** | `th-g5-backend-regression-evidence` · frontend_external **0** | — |
| **G6** | **blocked** (not run full soak) | `th-g6-soak-checklist-evidence` · half-probe only | staging 24h |
| **G7** | **blocked** | W7 opening | timed drill |
| **G8** | **blocked** | no `D7 flip 现在` | [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md) |

## Workers

| ID | Status |
|----|--------|
| W1–W6 | **DONE · reviewed** (closing W3/W4/W6 wt) |
| W7 / W8 / W9 | **opening** |

## Flip

Docs / exit 10 / feature push **≠ D7**. No production FRONTEND_MODE.

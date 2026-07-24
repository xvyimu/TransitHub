# TransitHub · Long Wave · GATE Matrix

> **D7 FLIP: NOT EXECUTED** · G0=D · **2026-07-24** 7m patrol harvest W7–W9

| Gate | Status | Evidence | Unblock |
|------|--------|----------|---------|
| **G1** | **green** | Module2 + console quality | — |
| **G2** | **blocked** | pack exit **10** · no `TH_E2E_*` | non-prod creds |
| **G3** | **blocked** live · contract **green** | channels evidence · contract **0** | after G2 |
| **G4** | **blocked** local · **CI SSOT** | docker absent | Docker / CI digest |
| **G5** | **green** | frontend_external + common/model tests **0** | — |
| **G6** | **blocked** | soak not run | staging 24h |
| **G7** | **blocked** (doc + dry-run) | `th-g7-rollback-drill-evidence` · timed n/a · docker absent | operator timed drill |
| **G8** | **blocked** | no `D7 flip 现在` | [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md) |

## Backend / legacy (not cutover gates)

| Item | Status | Evidence |
|------|--------|----------|
| LEGACY scan W8 | **DONE** · no branch delta vs main · historical feat(web) 可疑 pre-gate | `th-legacy-gate-scan-evidence` @ `98ddd6bd` |
| 3DB migrate W9 | **DONE** · only `refund_intents` missing from baseline · migrate-three-dialect **0** | `th-be-migrate-3db-evidence` @ `44ab1b5e` |
| Timeouts/Redis W10 | **live** | `th-be-timeouts-redis` |

## Flip

**Forbidden** without human phrase. Feature push ≠ D7. No default-branch push from coord loop.

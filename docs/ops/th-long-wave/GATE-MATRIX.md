# TransitHub · Long Wave · GATE Matrix

> SSOT for cutover gates this long wave. **Inherits W3/W4 + th-d7-scout `35194ff3` + W1 long-wave evidence** — no greenwash.  
> Plan: `docs/operations/web-console-cutover-plan.md` · Week queue: [WEEK-BACKLOG.md](./WEEK-BACKLOG.md)  
> **D7 FLIP: NOT EXECUTED** · G0=**D** (A+C non-prod) · week horizon 2026-07-24→31

Updated: **2026-07-24** (daily) · coord tip after W1 cherry-picks

| Gate | Meaning | Status | Evidence pointer | Unblock |
|------|---------|--------|------------------|---------|
| **G1** | Module2 on tip | **green** | W3/W4 + scout + W1 tree still present | — |
| **G2** | Non-prod login e2e | **blocked** | W4 + W1a pack exit **10** (`TH_E2E_*` unset); healthz/status/contract 0 | Operator non-prod creds → W2 wt |
| **G3** | Channels RO live | **blocked** live · contract **green** | contract exit **0** (W1a); live needs G2 | W3 after W2 |
| **G4** | Vue image build | **blocked** local · **CI SSOT** | docker absent; CI `image-reproducibility` | W4 · Docker or CI URL/digest |
| **G5** | `frontend_external` build | **green** (inherit) · **refresh via W5** | W4/scout exit **0**; bare `go build .` needs classic dist or tags | W5 re-run tagged build + tests |
| **G6** | Staging soak ≥24h | **blocked** | checklist only | W6 half/full evidence |
| **G7** | Rollback drill timed | **blocked** | doc + min seq only | W7 dry-run + wall-clock if env |
| **G8** | Owner sign-off | **blocked** | no `D7 flip 现在` | [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md) only |

## G0 + week policy

| Field | Value |
|-------|--------|
| Choice | **D = A + C non-prod** |
| Week | G1–G7 signable or explicit blocked; backend path reports; live ≤3 |
| Forbidden | production FRONTEND_MODE · D7 · delete web/default · dual-write · go.mod major · push unless ordered |

## Workers

| Module | wt | Status |
|--------|-----|--------|
| W1a M-TH-console-quality | `th-console-quality` | **DONE · reviewed** · closing |
| W1b M-TH-backend-stable-scout | `th-backend-stable-scout` | **DONE · reviewed** · closing |
| W2 M-TH-g2-e2e-nonprod | `th-g2-e2e-nonprod` | opening |
| W5 M-TH-g5-backend-regression | `th-g5-backend-regression` | opening |

## Backend findings feed (from W1b · not gates)

| Finding | Severity | Follow-up |
|---------|----------|-----------|
| MySQL/PG file migrate unvalidated | high for AUTO_MIGRATE=false | W9 |
| `refund_intents` missing from 000001 baseline | medium drift | W9 |
| max_open default 1000 MySQL/PG | ops risk | W10 |
| Redis rate-limit fail-closed 500 | expected | W10 document/ops |
| Email RL fail-open to memory | note | W10 |
| embed classic dist missing → bare `go build .` fail | agent env | W5 use `-tags frontend_external` |

## Related

- [progress.md](./progress.md) · [WEEK-BACKLOG.md](./WEEK-BACKLOG.md) · [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md)
- `docs/ops/th-console-quality-evidence-2026-07-24.md`
- `docs/ops/th-backend-stable-scout-evidence-2026-07-24.md`

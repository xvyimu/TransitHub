# TransitHub · Long Wave · WEEK BACKLOG (7-day)

> **Horizon:** 2026-07-24 → 2026-07-31 (approx)  
> **G0:** D = A+C non-prod  
> **D7 FLIP: NOT EXECUTED** · never without human `D7 flip 现在`  
> **Cadence:** DONE → review → commit on coord → `terminal stop` → `worktree rm --force` → open next · **live ≤ 3**  
> **Progress SSOT:** [progress.md](./progress.md) · gates: [GATE-MATRIX.md](./GATE-MATRIX.md)

## North star (week)

1. G1–G7 non-prod evidence = **signable green** or **explicit blocked** (no fake green).  
2. Backend stability = path-level reports + small safe fixes only (**no** go.mod major bump).  
3. G8 = human checklist only — [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md).

## Queue

| ID | wt name | Scope | Gate link | Status | Evidence |
|----|---------|-------|-----------|--------|----------|
| **W1a** | `th-console-quality` | pnpm typecheck/test/build + debt + W4 pack | console quality · G2/G3 honest | **DONE · reviewed** · wt rm (human) | `th-console-quality-evidence-2026-07-24.md` @ `4afcf5b3` |
| **W1b** | `th-backend-stable-scout` | migrations 3DB · pool/timeout · Redis map | backend stable | **DONE · reviewed** · wt rm (human) | `th-backend-stable-scout-evidence-2026-07-24.md` @ `d1dd3278` |
| **W2** | `th-g2-e2e-nonprod` | W4 login; no `TH_E2E_*` → **blocked** file | G2 | **DONE · reviewed · closing** | `th-g2-e2e-nonprod-evidence-2026-07-24.md` @ `d1957b64` · exit **10** |
| **W3** | `th-g3-channels` | Channels RO path (contract + live if G2) | G3 | **DONE · reviewed · closing** | `th-g3-channels-evidence` @ `4daf0ba9` · contract **0** · live blocked |
| **W4** | `th-g4-image-repro` | Vue image; Docker else **CI SSOT** | G4 | **DONE · reviewed · closing** | `th-g4-image-repro-evidence` @ `4c2560bf` · docker absent |
| **W5** | `th-g5-backend-regression` | `go test` + `go build -tags frontend_external` | G5 | **DONE · reviewed** · wt rm · tip `d6e3dfae` | exits **0** |
| **W6** | `th-g6-soak-checklist` | Soak checklist half-or-full · no fake 24h | G6 | **DONE · reviewed · closing** | `th-g6-soak-checklist-evidence` @ `f4669be9` · blocked not run |
| **W7** | `th-g7-rollback-drill` | Rollback doc + command dry-run (no prod) | G7 | **live / opening** | — |
| **W8** | `th-legacy-gate-scan` | Scan `web/default` for non-hotfix feature drift | legacy gate | **live / opening** | — |
| **W9** | `th-be-migrate-3db` | 3DB migrate quality audit RO + small fix if approved | backend | **live / opening** | refund_intents drift |
| **W10** | `th-be-timeouts-redis` | Timeout/pool/Redis follow-ups from W1b | backend | queued (W1b feeds) | — |
| **W11** | `th-console-a11y-debt` | Small console UX/a11y debt | console | queued (W1a feeds) | — |
| **W12** | coord-only | GATE pack refresh + INTEGRATE · **G8 human table only** | G1–G8 | queued end | — |

## Daily loop (coord)

1. `orca worktree list` (TH only) · count live ≤ 3.  
2. Harvest DONE evidence · cherry-pick/merge to `th-coord` · update GATE + this file.  
3. `orca terminal stop --worktree name:<wt>` → `orca worktree rm --worktree name:<wt> --force` (never orca / D:\orca).  
4. Open next 1–2 from queue with Orca `--agent claude`.  
5. End of day: commit progress on `th-coord` · **no push** unless ordered.

## Env blockers (honest)

| Need | For | Current |
|------|-----|---------|
| `TH_E2E_USER` + `TH_E2E_PASS` (non-prod) | W2/W3 live green | **unset** → blocked files OK |
| Docker CLI | W4 local image | **absent** → CI SSOT only |
| Staging ownership | W6 ≥24h | agent cannot fake |

## Red lines

- No `D7 flip 现在` → **no** production `FRONTEND_MODE`  
- No delete `web/default` · no React+Vue dual feature write · no second backend language  
- No go.mod Gin/redis major bump this week  
- No secrets in git · no fake green exit 0

## Review notes · W1 (2026-07-24)

### W1a console-quality — **PASS review**

| Check | Result |
|-------|--------|
| Boundary | `web-console/` + docs/ops evidence only |
| Exits | install/typecheck/test/build **0** · W4 pack **10** (honest) |
| Secrets | none |
| Dual-write / flip | none |
| Residual | unit tests thin (logQuery only); placeholders majority of nav |

### W1b backend-stable-scout — **PASS review**

| Check | Result |
|-------|--------|
| Boundary | docs-only scout |
| Exits | pool/common tests **0** · migrate-three-dialect **0** (sqlite) · bare `go build .` **1** (missing `web/classic/dist` embed — env, not scout bug) |
| Key findings | MySQL/PG file migrate not validated; `refund_intents` AutoMigrate vs baseline drift; Redis fail-closed on boot/rate-limit; email RL fail-open to memory |
| go.mod bump | none |

## Log

| Date | Event |
|------|--------|
| 2026-07-24 | WEEK-BACKLOG created · W1a/W1b DONE reviewed · open W2+W5 |

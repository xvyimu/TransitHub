# M-TH-g5-backend-regression · evidence · 2026-07-24

## D7 FLIP: NOT EXECUTED

Production `FRONTEND_MODE` **not** changed. No production migrate. No `git push`. No React delete. No business-semantic change.

## Worktree identity

| Field | Value |
|-------|--------|
| Module ID | **M-TH-g5-backend-regression** |
| Worktree (absolute) | `C:\Users\yuanjia\orca\workspaces\src\th-g5-backend-regression` |
| Branch | `xvyimu/th-g5-backend-regression` |
| Tip (start / evidence base) | `f7a8b9bde34ff8c2a9b9683b1d1ad59970b6c3b0` (`docs(ops): TH E2E operator card`) |
| Tip (post first evidence) | `75f3a6c0f8987ae84f82eb80a547aeda389f668f` |
| Tip (re-verify after agent rebuild) | this commit (post re-run exits + tip fields) |
| Agent | claude |
| Scope | Go build/test regression evidence only · `docs/ops/` · no go.mod major bump · no other-repo edits |
| Date | 2026-07-24 |
| Status | **DONE** · **in-review** (th-coord) |

## Boundary

| In | Out |
|----|-----|
| `go build -tags frontend_external` evidence | Bare `go build .` (may fail without `web/classic/dist` — **not** hard-fixed) |
| `go test ./common/ ./model/` | Full-repo `go test ./...` (out of knife) |
| Optional three-dialect migrate (empty SQLite) | MySQL/PG live URLs (skipped without env) |
| Docs under `docs/ops/` | Business semantic changes unless a single-line fix is **proven** necessary by unit tests (none needed) |
| | `git push` · production `FRONTEND_MODE` · delete `web/default` · **D7** |

## Pre-read

| Path | Result |
|------|--------|
| `docs/PROJECT.md` | Read — stack SSOT (Go/Gin/GORM · three DB · Vue strangler / React prod until D7) |
| `AGENTS.md` | Read (via project conventions) |
| `docs/ops/th-backend-stable-scout-evidence-2026-07-24.md` | **ABSENT** in this worktree (no matching path under `docs/ops/`) — noted; proceeded with PROJECT + AGENTS + prior W4 G5 pattern |

## Intentionally not done

| Item | Status |
|------|--------|
| **D7 production flip** | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched |
| Delete / replace React `web/default` | Not done |
| Bare `go build .` / embed classic dist repair | Not attempted (G5 authority = `frontend_external`) |
| `go.mod` major version bump | Not done |
| Business code / 1-line “necessary” fix | **None required** — build + targeted tests green |
| `git push` / merge default branch | Not done |
| Full `./...` test suite | Out of module boundary |

## Verification (this message · recorded exits)

Agent host: Go **1.26.5** windows/amd64 · toolchain `go1.26.5` · pwsh 7.x · CWD = worktree root.

| # | Command | Exit | Notes |
|---|---------|-----:|-------|
| 1 | `go build -trimpath -buildvcs=true -tags frontend_external -o new-api-backend-g5-verify.exe .` | **0** | artifact ~**84.9 MB** · **deleted after verify** (not committed) |
| 2 | `go test ./common/ ./model/ -count=1 -timeout 10m` | **0** | see package results |
| 3 | `pwsh -NoProfile -File scripts/migrate-three-dialect.ps1` | **0** | SQLite empty up **version=1 PASS** · mysql/postgres **SKIP** (no URL env) |

### Re-run (agent window rebuild · 2026-07-24 ~13:22 +0800)

Same three commands re-executed after session rebuild; exits re-confirmed.

| # | Command | Exit | Notes |
|---|---------|-----:|-------|
| 1 | `go build -trimpath -buildvcs=true -tags frontend_external -o new-api-backend-g5-verify.exe .` | **0** | ~**84.9 MB** · deleted after verify |
| 2 | `go test ./common/ ./model/ -count=1 -timeout 10m` | **0** | `common` **3.301s** · `model` **7.050s** |
| 3 | `pwsh -NoProfile -File scripts/migrate-three-dialect.ps1` | **0** | sqlite PASS version=1 · mysql/pg SKIP |

### Package results (test #2 · latest re-run)

```text
ok  	github.com/xvyimu/TransitHub/common	3.301s
ok  	github.com/xvyimu/TransitHub/model	7.050s
```

### G5 note vs bare build

- **Authoritative for G5:** `-tags frontend_external` → `frontend_assets_external.go` empty theme assets; no embed of `web/*/dist`.
- **Bare `go build .`** may fail if `web/classic/dist` (or other embed paths) are missing in this worktree. That is **not** a G5 failure and was **not** fixed by inventing dist.

## Outcome

| Claim | Evidence |
|-------|----------|
| Backend pure-Go path compiles | build #1 exit **0** |
| Core shared + model packages pass unit tests | test #2 exit **0** |
| Empty SQLite migrate still green | migrate #3 exit **0** (optional dialects skip) |
| No fake green | All exits recorded; no `-run` narrowing needed |
| No binary committed | `new-api-backend-g5-verify.exe` removed |
| D7 / FRONTEND_MODE / push | **NOT EXECUTED** / untouched / no push |

## Related

| Path | Role |
|------|------|
| `frontend_assets_external.go` | `//go:build frontend_external` empty assets |
| `docs/ops/stack-matrix-2026-07.md` | Prior G5 re-green history (W3/W4) |
| `docs/ops/w4-arch-upgrade-transithub-claude.md` | Prior `frontend_external` verify pattern |
| `docs/ops/migrate-three-dialect-strategy.md` | MySQL/PG skip policy |
| `docs/PROJECT.md` | Morph + stack SSOT |

## Handoff · th-coord

- **Status:** DONE + **in-review**
- **Ask:** accept G5 regression evidence; no further code knife unless coord expands scope
- **Do not:** D7 · push · production FRONTEND_MODE

# W3 · TransitHub · Claude · architecture upgrade

## D7 FLIP: NOT EXECUTED

Production `FRONTEND_MODE` **not** changed. No production migrate. No `git push`. No React delete.

## Worktree identity

| Field | Value |
|-------|--------|
| Worktree (absolute) | `C:\Users\yuanjia\orca\workspaces\src\w3-th-claude` |
| Branch | `xvyimu/w3-th-claude` |
| Tip (start/end docs) | `b2fff447` (`merge(main): integrate form-stack SSOT before W2 land`) |
| Portfolio baseline | ~`b2fff447` per `prompts/w3-th.md` |
| Agent | claude (solo) |
| Scope | W3 only: D7 gate dossier · rollback desktop drill doc · staging soak checklist · Gin/redis optional (defer) |
| Date | 2026-07-23 |

## Delivered

1. **`docs/ops/w3-d7-gate-dossier.md`** — G1–G8 each with status · evidence · owner · gap. No blank statuses.  
2. **`docs/ops/w3-rollback-desktop-drill.md`** — non-prod step table + expected ≤5 min · **not executed** (no docker/staging).  
3. **`docs/ops/w3-staging-soak-checklist.md`** — 24h observation checklist (4xx/5xx, login, channels RO, probes).  
4. **Gin / redis** — **defer** both (no `go.mod` bump); spike reaffirmed.  
5. **stack-matrix** W3 column · cutover-plan W3 evidence pack · this report.

## Gate snapshot (W3)

| Gate | Status | One-line |
|------|--------|----------|
| G1 | **green** | Module2 tree + contract/cutover docs present |
| G2 | **blocked** | healthz 200; login exit **1**; need `TH_E2E_*` |
| G3 | **blocked** live / contract green | needs G2 session; OpenAPI validates |
| G4 | **blocked** local / **CI SSOT** | docker not on PATH; `image-reproducibility` builds Vue image |
| G5 | **green** | `go build -tags frontend_external` exit **0** |
| G6 | **blocked** | soak checklist only — not run |
| G7 | **blocked** | drill doc only — not run |
| G8 | **blocked** | human “D7 flip 现在” not given |

**Flip readiness: NO.** Written **延期/阻断证据** lives in dossier (not silent DEFER).

## Intentionally not done (W3 bans)

| Item | Status |
|------|--------|
| **D7 production flip** | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched |
| Delete / replace React `web/default` | Not done |
| Production migrate / live DSN | Not done |
| `git push` / merge default branch | Not done (总控) |
| Gin 1.10 **or** redis v9 bump | Deferred (optional; tests would be required — not attempted) |
| Real staging soak / docker image build on agent | Environment block |
| publish-runtime / asar / ISS | N/A |
| Concurrent CP production flip | Forbidden by portfolio |

## Verification (this message · recorded exits)

Agent: Go **1.26.5** · Node **v24.16.0** (console CI pins 22) · pnpm **11.5.0** · Python **3.14.5**.

### Contract / migrate / G2 / G5

| # | Command | Exit | Notes |
|---|---------|-----:|-------|
| 1 | `python scripts/validate-console-contract.py` | **0** | 8 ops + schemas PASS |
| 2 | `pwsh -NoProfile -File scripts/migrate-three-dialect.ps1` | **0** | SQLite version=1 PASS; mysql/postgres SKIP |
| 3 | `pwsh -NoProfile -File scripts/e2e-web-console-login.ps1 -SkipVite` | **1** | healthz 200; login fail (no `TH_E2E_*`) |
| 4 | `go build -trimpath -buildvcs=true -tags frontend_external -o new-api-backend-w3.exe .` | **0** | ~89 MB; removed after verify |
| 5 | docker on PATH | **1** (absent) | G4 local blocked; CI SSOT |

### web-console (CWD `web-console/`)

| # | Command | Exit | Notes |
|---|---------|-----:|-------|
| 6 | `pnpm install --frozen-lockfile` | **0** | pnpm 11.5.0 |
| 7 | `pnpm typecheck` | **0** | clean |
| 8 | `pnpm test` | **0** | Vitest 5/5 |
| 9 | `pnpm build` | **0** | dist written · chunk-size warning only |

### NOTICE

| # | Check | Count |
|---|-------|------:|
| 10 | QuantumNous new-api URL in `ConsoleLayout.vue` | **1** |
| 11 | NOTICE English string in `en.ts` | **1** |

### Pins (no bump)

| Module | go.mod | Latest listed (proxy) |
|--------|--------|------------------------|
| gin | v1.9.1 | v1.12.0 |
| go-redis/v8 | v8.11.5 | (v9 module: v9.21.0) |

## Acceptance checklist (W3 prompt)

| Criterion | Met? |
|-----------|------|
| Dossier exists · G1–G8 no blank status | **Yes** |
| Rollback desktop drill step table | **Yes** (not executed) |
| Staging soak checklist | **Yes** (not executed) |
| Gin **or** redis **or** explicit defer | **Yes** — defer |
| Report: NOT EXECUTED · exits · no push | **Yes** |

## Related

| Path | Role |
|------|------|
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G1–G8 dossier |
| [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) | Non-prod rollback drill |
| [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) | 24h soak |
| [stack-matrix-2026-07.md](./stack-matrix-2026-07.md) | Stack card (W3 column) |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | G2 env names |
| [w1-gin-redis-spike.md](./w1-gin-redis-spike.md) | Gin/redis still defer |
| [../operations/web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | G1–G8 + W3 pack |
| [../PROJECT.md](../PROJECT.md) §2.2 | Frontend transition SSOT |

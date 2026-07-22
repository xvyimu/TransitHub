# TransitHub · stack matrix · 2026-07

> Stack card: **current → H2 2026 target → wave status**.  
> Companion: portfolio plan `D:\orca\.planning\portfolio-arch-upgrade-2026h2\repos\th.md`.  
> Does **not** authorize production cutover, dependency bumps that fail CI, or React deletion.

## Worktree identity

| Field | W1 | W2 | W3 |
|-------|----|----|----|
| Worktree | `...\w1-th-claude` | `C:\Users\yuanjia\orca\workspaces\src\w2-th-claude` | `C:\Users\yuanjia\orca\workspaces\src\w3-th-claude` |
| Branch | `xvyimu/w1-th-claude` | `xvyimu/w2-th-claude` | `xvyimu/w3-th-claude` |
| Date | 2026-07-23 | 2026-07-23 | 2026-07-23 |
| Agent | claude | claude | claude |

## Matrix

| Layer | Current (repo tip) | Target (H2 ~2027-01) | W1 | W2 | W3 this wave |
|-------|--------------------|----------------------|----|----|--------------|
| **Go** | `go 1.25.1` · `toolchain go1.26.5` | Pin CI to toolchain; security patches | CI pins `1.26.5` | Unchanged | Unchanged · G5 re-green |
| **Gin** | `github.com/gin-gonic/gin v1.9.1` | Maintain line **≥1.10** after soak | Spike only | Still defer | **Still defer** (no go.mod) |
| **Redis client** | `github.com/go-redis/redis/v8 v8.11.5` | **v9** dedicated wt | Spike only | Still defer | **Still defer** (no go.mod) |
| **GORM** | `gorm.io/gorm v1.25.2` | Maintain + three-dialect empty-DB | Doc only | SQLite green + strategy | SQLite empty migrate re-green |
| **web-console** | Vue 3.5 · Vite 8 · pnpm 11.5 · Node 22 CI | Patch follow; e2e main CI later | Quality green | Quality re-run | Quality re-run exit 0 |
| **Console API contract** | Markdown + partial yaml | OpenAPI subset → TS client path | — | **`console-subset.yaml` 1.1.0-w2** + validator | Validator re-green (no schema bump) |
| **D7 / cutover** | React prod default until gate | D7 after G1–G8 + human | Evidence pack start | Cred checklist | **Dossier G1–G8** · **NOT EXECUTED** |
| **React default** | `web/default` LEGACY freeze | Shrink after D7 | Unchanged | Unchanged | Unchanged (no delete) |
| **Node (console CI)** | Node 22 · pnpm 11.5 | Crosscut align | Match | Match | Match |
| **Bun (legacy web)** | Bun 1.3.14 | Until D7 | Unchanged | Unchanged | Unchanged |

## CI pins (authoritative)

| Job / path | Pin |
|------------|-----|
| `go-quality` / `sqlite-migrate` | Go **1.26.5** |
| `web-quality` | Bun **1.3.14** |
| `web-console-quality` | Node **22** · pnpm **11.5.0** |
| Vue image builder | `node:22-bookworm` · pnpm 11.5.0 |

## Architecture boundary (unchanged)

- TransitHub = **AI gateway** (protocol / billing / channel routing). Not a desktop Chat workbench.
- Console = same-origin management plane; **no** relay/billing logic in Vue.
- Production default frontend remains **embedded React** until D7 human gate (`docs/operations/web-console-cutover-plan.md`).

## Next waves (pointer only)

| Wave | Stack / arch |
|------|----------------|
| W3 | **D7 dossier + drill/soak docs** (flip still human) — **this wave** · Gin/redis still dedicated wt |
| W4 | React shrink strategy or schedule · matrix close-out · MySQL/PG file baselines if scheduled · D7 only if G1–G8 + human |

## Related

| Path | Role |
|------|------|
| [w1-arch-upgrade-transithub-claude.md](./w1-arch-upgrade-transithub-claude.md) | W1 report |
| [w2-arch-upgrade-transithub-claude.md](./w2-arch-upgrade-transithub-claude.md) | W2 report |
| [w3-arch-upgrade-transithub-claude.md](./w3-arch-upgrade-transithub-claude.md) | W3 report · **D7 NOT EXECUTED** |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G1–G8 executable dossier |
| [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) | Non-prod rollback drill |
| [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) | 24h soak checklist |
| [w1-gin-redis-spike.md](./w1-gin-redis-spike.md) | Gin/redis — W3 still defer |
| [migrate-three-dialect-strategy.md](./migrate-three-dialect-strategy.md) | Empty-DB dialect policy |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | G2 non-prod env list |
| [../openapi/console-subset.yaml](../openapi/console-subset.yaml) | Machine contract |
| [../operations/web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | Cutover G1–G8 |
| [../ARCHITECTURE_TARGET.md](../ARCHITECTURE_TARGET.md) | Phase1 target |

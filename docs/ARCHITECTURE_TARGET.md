# TransitHub · Architecture Target

> **Status:** Phase1 target contract. This document defines implementation and verification boundaries; it does **not** authorize a production cutover, database migration, or a change to the default frontend.

## 1. Target in one page

TransitHub remains a Go AI gateway. The management console is being moved incrementally to `web-console/` (Vue 3 + Naive UI) behind the existing same-origin delivery seam. The existing React consoles remain buildable rollback artifacts until every gate below is closed and an operator explicitly authorizes cutover.

The architecture follows a one-way runtime path:

```text
Browser / SDK
    -> Nginx or embedded frontend
    -> Gin router -> middleware -> controller -> service -> model
                                      \-> relay/channel -> upstream AI
    -> SQLite | MySQL | PostgreSQL, optional Redis, optional log database
```

`router -> middleware -> controller -> service -> model` is the management-plane dependency direction. Provider adapters, relay protocols, quota arithmetic, pre-consume/settle behavior, authentication, and database compatibility stay in Go.

## 2. Frontend target and boundaries

| Area | Target | Current boundary |
|------|--------|------------------|
| New console work | `web-console/` Vue 3 + TypeScript + antdv-next | New console capability belongs here; health, channels, models, and logs are presently read-only slices. |
| Existing default UI | `web/default/` React | **LEGACY**: security, severe-regression, embed/rollback build, and existing-production typo fixes only. |
| Classic UI | `web/classic/` | **L2 frozen**: no new screens or feature-parity work. |
| Delivery | `FRONTEND_MODE` plus `deploy/separated/` | Default delivery remains React until an explicit cutover decision. Vue is served by `Dockerfile.frontend.vue` through the same Nginx API proxy. |

The console consumes same-origin management endpoints under `/api/*`; it does not connect directly to a database, a relay provider, or a future AI-Core process. Long-lived React/Vue dual implementation of the same screen is out of scope.

## 3. Cutover gates and rollback

All gates are required before a production-default Vue switch:

1. `web-console` locked-install, `vue-tsc`, unit tests, Vite build, and its separated image/Nginx configuration are green in CI.  
   **Wave6 Dual-B (main):** `.github/workflows/quality.yml` job `web-console-quality` (+ image job dependency) and `web-console/pnpm-workspace.yaml` `allowBuilds.esbuild` land this gate for install/typecheck/test/build; **does not** flip production traffic.
2. The Vue shared layout presents the NOTICE attribution and a visible link to the original new-api project.  
   **Wave6 Dual-B:** footer NOTICE link present in `web-console` layout (still not a cutover).
3. On a non-production environment, same-origin login, `/api/status`, and the approved read-only console pages behave as expected.
4. The SQL migration path has an empty-database validation strategy for SQLite, MySQL, and PostgreSQL; the current SQLite-only baseline is not sufficient.  
   See `migrations/README.md` three-dialect policy + `docs/operations/db-migrations.md` force-baseline notes.
5. An operator approves a documented cutover plan (`docs/operations/web-console-cutover-plan.md` G1–G8). Until then, do not change production `FRONTEND_MODE`, traffic routing, or images. **D7 remains human gate.**  
   **W1 (2026-07-23):** pre-flip evidence pack recorded — G5 `frontend_external` build exit 0; web-console quality re-green; G2 blocked on non-prod credentials; G4 blocked without Docker CLI on agent host (CI image job remains SSOT). See `docs/ops/w1-arch-upgrade-transithub-claude.md` and the W1 table in the cutover plan. **Still no production flip.**  
   **W3 (2026-07-23):** executable G1–G8 dossier `docs/ops/w3-d7-gate-dossier.md` + rollback desktop drill + staging soak checklist. **D7 FLIP: NOT EXECUTED.** G2/G3 live / G4 local / G6 / G7 / G8 still blocked with written evidence (not silent DEFER). See `docs/ops/w3-arch-upgrade-transithub-claude.md`.

Rollback is configuration/image selection, not a source rewrite: restore the embedded React frontend or use `deploy/separated/Dockerfile.frontend` with the same Nginx proxy contract. See [the rollback runbook](operations/web-console-cutover-rollback.md).

## 4. Database migration target

`migrations/` is the source of truth for new schema evolution. Every main-database change must remain valid for SQLite, MySQL, and PostgreSQL, using portable SQL where possible and an explicit, documented dialect split where it is not. Expand/contract changes are required for destructive schema evolution.

The current `000001` baseline is only exercised against an empty SQLite database. MySQL and PostgreSQL remain supported by the application, but they are not yet a file-migration cutover baseline. Do not set `SQL_AUTO_MIGRATE=false` for a deployment until the target dialect has a reviewed baseline, CI evidence, and an approved operation plan. See [migrations/README.md](../migrations/README.md) and [database migration operations](operations/db-migrations.md).

## 5. Explicit non-goals

- No D7/production cutover, traffic switch, or production environment change.
- No relay, provider, billing, quota, authentication, SSE, or WebSocket rewrite.
- No removal or replacement of new-api, QuantumNous, or other required attribution.
- No Python AI-Core on the synchronous request, billing, or console data path. A future AI-Core may only be an asynchronous, versioned Go outbound seam after a separate decision.
- No removal of SQLite or MySQL support without a product decision.

## 6. Evidence and companion documents

| Topic | Repository source |
|-------|-------------------|
| Current facts and legacy constraints | [ARCHITECTURE_ASIS.md](ARCHITECTURE_ASIS.md) |
| Module boundaries and management/relay separation | [gateway/MODULE_BOUNDARIES.md](gateway/MODULE_BOUNDARIES.md) |
| Console API contract | [gateway/CONSOLE_API_CONTRACT.md](gateway/CONSOLE_API_CONTRACT.md) |
| Legacy frontend gate | [legacy-frontend-gate.md](legacy-frontend-gate.md) |
| Frontend/backend delivery seam | [adr/0001-frontend-backend-delivery-seam.md](adr/0001-frontend-backend-delivery-seam.md) |
| Build and separated-delivery checks | [operations/build-and-release.md](operations/build-and-release.md), [../.github/workflows/quality.yml](../.github/workflows/quality.yml) |

The source tree, tests, and these repository-relative documents are the evidence for this target. External planning documents are context only and are not required to apply or verify it.

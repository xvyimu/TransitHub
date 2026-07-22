# Phase1 · TransitHub `web-console` (Vue3 + Naive UI)

Strangler-target admin console. **Phase1 MVP**: password login + health/status home.  
Does **not** replace `web/default` (React LEGACY) until cutover gate.

## Stack

| Piece | Choice |
|-------|--------|
| Build | Vite 8 · TypeScript strict |
| UI | Vue 3 · Naive UI · vue-router · Pinia · vue-i18n |
| HTTP | axios · `withCredentials: true` · same-origin `baseURL=''` |
| Package manager | **pnpm 11.5.0** (independent lockfile; not under `web/` bun workspace). `pnpm-workspace.yaml` allows only the `esbuild` install hook because Vite requires its platform binary. |

## Local dev

```bash
# terminal A — Go backend (repo root)
# FRONTEND_MODE=disabled optional; embed also works if you only need /api
go run .   # default :3000

# terminal B
cd web-console
pnpm install
pnpm dev   # :5173, proxies /api /healthz /livez /readyz → :3000
```

Open http://127.0.0.1:5173/login

## Scripts

| Command | Purpose |
|---------|---------|
| `pnpm install` / `pnpm install --frozen-lockfile` | Install deps (CI uses frozen lockfile) |
| `pnpm dev` | Vite dev server + API proxy |
| `pnpm typecheck` | `vue-tsc -b --pretty false` |
| `pnpm test` | Vitest unit tests (once) |
| `pnpm build` | `vue-tsc -b && vite build` → `dist/` |
| `pnpm preview` | Preview production build |

## Non-production acceptance (smoke)

**Scope:** local / CI quality only. Does **not** flip production traffic, change `FRONTEND_MODE`, or replace embedded React. Aligns with TARGET cutover gates 1–2 (`docs/ARCHITECTURE_TARGET.md` §3) and CI job `web-console-quality`.

Run from `web-console/` (package manager **pnpm 11.5.0** via Corepack recommended):

```bash
# 1) locked install (same as CI)
pnpm install --frozen-lockfile

# 2) static quality
pnpm typecheck
pnpm test
pnpm build
```

NOTICE attribution must stay **user-visible** in the shared layout footer (not only in source comments). Static check (from repo root; same strings as CI):

```bash
# expect exit 0; strings must match locale + layout
grep -F -- 'https://github.com/QuantumNous/new-api' web-console/src/layouts/ConsoleLayout.vue
grep -F -- 'Frontend design and development by New API contributors.' web-console/src/i18n/locales/en.ts
```

| Step | Command | Expect |
|------|---------|--------|
| Install | `pnpm install --frozen-lockfile` | exit **0** |
| Typecheck | `pnpm typecheck` | exit **0** · no `vue-tsc` errors |
| Unit tests | `pnpm test` | exit **0** · Vitest all green |
| Build | `pnpm build` | exit **0** · `dist/` written |
| NOTICE link | `grep` QuantumNous URL in `ConsoleLayout.vue` | exit **0** · match |
| NOTICE text | `grep` attribution string in `en.ts` | exit **0** · match |

Optional live API smokes (need a **non-prod** backend; credentials via env only — never commit):

| Smoke | Command | Notes |
|-------|---------|-------|
| Login e2e | `pwsh -File scripts/e2e-web-console-login.ps1` | see `E2E.md` |
| Logs RO | `pwsh -File scripts/smoke-logs.ps1` | see `docs/ops/T-TH-003-logs-live-smoke.md` |

Dev server check (manual): `pnpm dev` → open `/login`; after auth, footer shows NOTICE attribution + original new-api link.

## Console API subset (Phase1)

| Method | Path |
|--------|------|
| POST | `/api/user/login` |
| GET | `/api/user/logout` |
| GET | `/api/user/self` |
| GET | `/api/status` |
| GET | `/healthz` `/livez` `/readyz` |
| GET | `/frontend-healthz` (Nginx edge only) |

## Production / separated

Prefer same-origin Nginx (ADR-0001):

```text
browser → Nginx (this dist) → Go FRONTEND_MODE=disabled (-tags frontend_external)
```

See:

- `deploy/separated/Dockerfile.frontend.vue`
- `deploy/separated/README.md` (Vue section)
- `docs/operations/web-console-cutover-rollback.md`

Default integrated image **still embeds React** until organizational cutover.

## Non-goals (Phase1)

- Full feature rewrite (channels CRUD, wallet, playground SSE, …)
- OAuth / Passkey / 2FA complete flows (2FA login returns a clear error)
- Long-term React+Vue dual-write — new UI features go **here** only

## Spec

- Execution SSOT: `D:\orca\docs\phase1-execution-spec-transithub-2026-07-22.md` §3 WP-V  
- Bid: `docs/phase1-bid-vue-console.md`

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
| `pnpm dev` | Vite dev server + API proxy |
| `pnpm build` | `vue-tsc -b && vite build` → `dist/` |
| `pnpm preview` | Preview production build |
| `pnpm typecheck` | `vue-tsc -b --pretty false` |

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

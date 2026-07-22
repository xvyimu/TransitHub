# Wave8 · TransitHub · Claude

## Worktree identity

| Field | Value |
|-------|-------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\wave8-th-claude` |
| Branch | `xvyimu/wave8-th-claude` |
| Baseline tip | `89f65b78f4b2c23c3c85cd56679f687291396b71` (`docs(arch): wave7 note Dual-B gates on TARGET cutover list`) |
| Agent | claude |
| Scope | Non-prod web-console smoke / acceptance command table + verify |

## Tip check

- `git rev-parse HEAD` at start: `89f65b78f4b2c23c3c85cd56679f687291396b71`
- Matches shared main tip used by peer worktree `xvyimu/wave8-th-codex` at wave start
- Working tree was clean before this knife

## Delivered (one knife)

1. **`web-console/README.md`** — Non-production acceptance (smoke) table:
   - `pnpm install --frozen-lockfile` / `typecheck` / `test` / `build`
   - NOTICE static checks (same strings as CI `web-console-quality`)
   - Optional live smokes pointers (`e2e-web-console-login.ps1`, `smoke-logs.ps1`)
   - Explicit non-prod boundary (no `FRONTEND_MODE` flip, no React delete)
2. **This report** — commands + exit codes + intentional non-goals

No production env, no D7 cutover, no ISS feature work.

## Intentionally not done

- No production `FRONTEND_MODE` change, traffic cutover, deploy, push to default branch, or release publish
- No D7 human gate, production CSP/RLS, or ISS features
- No delete/replace of React `web/default`
- No relay / billing / auth / DB migration execution
- Optional “thin error-copy TARGET alignment” skipped (README + CI strings already match TARGET gates 1–2)

## Verification

Run from worktree root unless noted. Recorded **2026-07-22** on this agent.

| # | Command | CWD | Exit | Notes |
|---|---------|-----|-----:|-------|
| 1 | `pnpm install --frozen-lockfile` | `web-console/` | **0** | pnpm 11.5.0 · lockfile up to date · 145 packages |
| 2 | `pnpm typecheck` | `web-console/` | **0** | `vue-tsc -b --pretty false` clean |
| 3 | `pnpm test` | `web-console/` | **0** | Vitest 5/5 passed (`logQuery.test.ts`) |
| 4 | `pnpm build` | `web-console/` | **0** | `vue-tsc -b && vite build` · dist written (chunk-size warning only) |
| 5 | NOTICE link (`Select-String` / grep equivalent) | repo root | **0** | match in `ConsoleLayout.vue` |
| 6 | NOTICE text | repo root | **0** | match in `en.ts` |

### Exact NOTICE checks (CI parity)

```text
grep -F -- 'https://github.com/QuantumNous/new-api' web-console/src/layouts/ConsoleLayout.vue
grep -F -- 'Frontend design and development by New API contributors.' web-console/src/i18n/locales/en.ts
```

## Related

| Path | Role |
|------|------|
| `web-console/README.md` | Non-prod smoke table (this wave) |
| `docs/ARCHITECTURE_TARGET.md` §3 | Cutover gates (no flip) |
| `.github/workflows/quality.yml` | job `web-console-quality` |
| `docs/ops/wave6-dual-b-transithub-codex.md` | Prior Dual-B Vue CI + NOTICE |
| `web-console/E2E.md` | Optional login e2e |

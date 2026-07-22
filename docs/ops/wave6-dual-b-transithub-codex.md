# Wave6 Dual-B · TransitHub · Codex

## Worktree identity

| Field | Value |
|-------|-------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\dual-b-th-codex` |
| Branch | `xvyimu/dual-b-th-codex` |
| Baseline tip | `57d0891c1b7ada5200a3b5ea5a564af3911fde4e` |
| Scope | Phase1 architecture/quality/attribution gates only |

## Delivered

1. Added `docs/ARCHITECTURE_TARGET.md` as the repository-relative Phase1 target contract. It freezes Go relay, billing, authentication, and three-dialect compatibility boundaries; defines the Vue strangler and rollback path; and lists the gates that still block a default Vue cutover.
2. Corrected stale facts in `docs/ARCHITECTURE_ASIS.md`: `web-console/` and `migrations/` now appear as existing Phase1 work, while their remaining production gates are explicit. `migrations/README.md` now states that SQLite is the only current file-migration baseline and specifies the evidence required before MySQL/PostgreSQL file-migration cutover.
3. Added a `web-console-quality` CI job with locked pnpm 11.5.0 installation, `vue-tsc`, unit tests, production build, and a NOTICE-link check. The image job waits for it, builds `Dockerfile.frontend.vue`, validates the shared Nginx template in both frontend images, and records the Vue image digest.
4. Added a visible, localized Vue layout footer containing the required New API contributor attribution, original new-api link, and supplementary TransitHub source link. `web-console/pnpm-workspace.yaml` explicitly permits only Vite's required `esbuild` install hook; the Dockerfile copies that policy before locked install.

## Intentionally not done

- No production `FRONTEND_MODE` change, traffic cutover, deployment, push, merge, or release publishing.
- No database migration was executed and no MySQL/PostgreSQL baseline was invented.
- No relay, provider, quota, billing, authentication, or legacy React feature change.

## Verification

| Command | Result | Exit code |
|---------|--------|----------:|
| `pnpm install --frozen-lockfile` (`web-console/`) | locked install and explicit esbuild approval accepted | 0 |
| `pnpm typecheck` (`web-console/`) | `vue-tsc` reports no errors | 0 |
| `pnpm test` (`web-console/`) | Vitest: 5/5 tests passed | 0 |
| `pnpm build` (`web-console/`) | `vue-tsc -b && vite build` completed | 0 |
| `python -c "import pathlib, yaml; yaml.safe_load(pathlib.Path('.github/workflows/quality.yml').read_text(encoding='utf-8')); print('quality.yml parsed')"` | workflow YAML parsed | 0 |
| scoped `rg -n "sk-[A-Za-z0-9]|api[_-]?key\\s*[:=]|BEGIN (RSA |OPENSSH )?PRIVATE" <changed paths>` | no new secret-like match (`rg` uses exit 1 for no match) | 1 |
| `docker version --format '{{.Server.Version}}'` | not available on this worker; Vue image/Nginx runtime check is left to the new CI job | 1 |

The Vite production build emits a chunk-size warning; this wave does not alter code-splitting scope.

## Assumed items the opposite worktree can absorb

1. Pnpm 11 no longer reads `package.json#pnpm.onlyBuiltDependencies`; use `pnpm-workspace.yaml` with `allowBuilds: { esbuild: true }`, and copy that file into the Vue Docker build stage.
2. A Vue quality job should be a dependency of image reproducibility so an image cannot be the first place a `vue-tsc` or unit-test failure appears.
3. The NOTICE gate should check both the fixed upstream URL in `ConsoleLayout.vue` and the required attribution text in the locale source, preserving a user-visible location without a cutover.

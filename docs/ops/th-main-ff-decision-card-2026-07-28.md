# Decision card — `origin/main` ← `governance-cleanup-2026-07-28` fast-forward

**Date:** 2026-07-28  
**Author context:** M-TH-json-ctrl agent run (docs only)  
**Range:** `origin/main` (`7ee1f197`) .. tip `cb61e983` (branch `governance-cleanup-2026-07-28` / this worktree base)  
**Action taken here:** **NONE — FF not executed.** This is a human gate card only.

```text
git rev-list --count origin/main..cb61e983   # → 5
git merge-base origin/main cb61e983          # → 7ee1f197 (origin/main tip)
```

Linear ancestry (graph):

```text
*   cb61e983 merge: th-migrate-3db (three-dialect empty-DB baseline + CI)
|\
| * 76c7557c feat(migrations): three-dialect empty-DB baseline + CI service containers
*   c6b2e5ca Merge branch 'xvyimu/th-console-views'
|\
| * 6485f548 feat(web-console): add Tokens CRUD write view (contract-driven)
* 4d012b43 ci(legacy-guard): add PR gate blocking new features in legacy web/default
  (base) 7ee1f197 docs(ops): D-G4 image-reproducibility SSOT + D-LEG legacy-guard  ← origin/main
```

Because merge-base == `origin/main`, a **fast-forward** of `main` to `cb61e983` is git-clean (no merge commit required). Whether it *should* land is a product/ops decision below.

---

## The 5 commits (newest first)

| # | SHA | Subject | What lands on main if FF |
|---|-----|---------|--------------------------|
| 1 | `cb61e983` | **merge:** th-migrate-3db | Merge node integrating #3 into the console/legacy line. |
| 2 | `c6b2e5ca` | **Merge** branch `xvyimu/th-console-views` | Merge node integrating #4 into the legacy-guard line. |
| 3 | `76c7557c` | **feat(migrations):** three-dialect empty-DB baseline + CI service containers | Per-dialect `migrations/main/{sqlite,mysql,postgres}/`; `cmd/dbmigrate` scheme→subdir; MySQL/PG drivers; CI `mysql-migrate` / `pg-migrate` service jobs; docs. **Empty-DB only**; no production DSN; no `SQL_AUTO_MIGRATE` flip; existing-install force-baseline still human-gated. |
| 4 | `6485f548` | **feat(web-console):** Tokens CRUD write view (contract-driven) | Vue `web-console` Tokens list/create/edit/enable/delete/reveal; form helpers + unit tests; `CONSOLE_API_CONTRACT` §3.3. **No** `FRONTEND_MODE` / D7 change; React `web/default` untouched for this feature. |
| 5 | `4d012b43` | **ci(legacy-guard):** PR gate blocking new features in legacy `web/default` | New `.github/workflows/legacy-guard.yml` + `scripts/check-legacy-guard.ps1` (G-LEG-1/2/4). Does **not** delete `web/default`. |

### Risk notes per theme

| Theme | Risk if FF’d | Mitigations already in commits |
|-------|--------------|--------------------------------|
| **Migrations 3-dialect** | Wrong path selection on non-empty / AutoMigrate-era installs; CI service cost; accidental operator use of force-baseline | Empty-DB scoped; irreversible-down documented; production migrate not auto-run; D7 still required for live force-baseline |
| **web-console Tokens** | Contract drift vs backend token API; i18n gaps | Contract doc + pure form unit tests; UserAuth path only; no dual-write to React |
| **legacy-guard CI** | False-positive blocks on legitimate LEGACY-HOTFIX; workflow noise | Explicit HOTFIX marker + allowlist; independent workflow so quality.yml can evolve separately |

---

## Suggested FF conditions (human gate)

Do **not** FF until a human checks:

1. **CI green on the tip** for Quality Gate (and Legacy Guard if exercised via PR). Note: recent `main` Quality Gate runs were already **failing** on docs/json merges before this range — treat tip CI as a **fresh** signal, not “main was green”.
2. **No production migrate** will be triggered solely by merging file baselines (confirm ops runbook: empty-DB / new install only).
3. **D7 still closed** — confirm no commit in range flips `FRONTEND_MODE` or production default UI (inspected: none do).
4. **Operator awareness** of legacy-guard: future PRs touching `web/default` need LEGACY-HOTFIX body markers.
5. **Explicit human command** to update `main` (local or `origin`). Prefer PR + review over bare `git push` even for FF.

### Example commands (**do not run in this task**)

```bash
# inspect only
git log --oneline origin/main..cb61e983
git diff --stat origin/main...cb61e983

# only after human approval
git checkout main && git merge --ff-only cb61e983
# push origin main is a separate, higher bar — still human-gated
```

---

## Decision status

| Field | Value |
|-------|--------|
| **FF executed?** | **No** |
| **Push `main`?** | **No** |
| **Recommendation** | Eligible for FF *technically* (linear, merge-base clean). Defer until tip CI reviewed + operator ack on migrations empty-DB scope + legacy-guard semantics. |
| **Related evidence** | `docs/ops/th-json-controller-evidence-2026-07-28.md` (orthogonal debt work on this branch; **not** part of the 5-commit FF range) |

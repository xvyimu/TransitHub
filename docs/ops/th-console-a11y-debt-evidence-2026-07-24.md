# M-TH-console-a11y-debt · evidence · 2026-07-24

> **Module:** M-TH-console-a11y-debt (worktree `th-console-a11y-debt-2`)  
> **D7 FLIP: NOT EXECUTED** · no production `FRONTEND_MODE` · no `web/default` dual-write · no default-branch push  
> **Agent note:** prior child agent stalled; **coord completed scan + evidence** on this wt after `terminal stop`.

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-console-a11y-debt-2` |
| Branch | `xvyimu/th-console-a11y-debt-2` |
| Tip (pre-evidence) | `f7a8b9bd` |
| Date | **2026-07-24** |
| Scope | `web-console/src` static a11y debt inventory · optional typecheck · **docs only** (no product code change this knife) |
| Status | **DONE** · **in-review** (th-coord) |

## Boundary

| In | Out |
|----|-----|
| Scan Vue SFCs under `web-console/src` | Large a11y rewrite / design system |
| Debt table + typecheck exit | `web/default` / React |
| Evidence + commit | D7 · secrets · push main |

## Commands + exits

| # | Command | CWD | Exit | Notes |
|---|---------|-----|-----:|-------|
| 1 | `pnpm install --frozen-lockfile` | `web-console/` | **0** | pnpm 11.5.0 · lock clean |
| 2 | `pnpm typecheck` (`vue-tsc -b --pretty false`) | `web-console/` | **0** | clean |
| 3 | Content scan (ripgrep-style) for `aria-` / `alt=` / `NButton` / `PlaceholderView` / `NDataTable` | `web-console/src` | n/a | **zero** explicit `aria-label` / `alt=` in app SFCs |

## Surface inventory

| Area | Files | Role |
|------|-------|------|
| Layout | `layouts/ConsoleLayout.vue` | sider menu + header logout + NOTICE footer |
| Auth | `views/LoginView.vue` | username/password form |
| Implemented RO | `HealthView` · `ChannelsView` · `ModelsView` · `LogsView` | tables + search |
| Placeholder | `PlaceholderView.vue` via router | keys · users · billing · settings · system · playground · profile |
| 404 | `NotFoundView.vue` | result + back button |

## Debt table

| Sev | Path | Issue | Notes / fix hint |
|-----|------|-------|------------------|
| **P1** | `views/ChannelsView.vue` · `ModelsView.vue` · `LogsView.vue` | Search `NInput` / filters rely on **placeholder only** — no visible `<label>` / `aria-label` / `NFormItem` | Add `aria-label` bound to i18n or wrap `NFormItem :label` |
| **P1** | `views/ChannelsView.vue` (`NSelect` status) | Status filter control has **no accessible name** | `aria-label` or form item label |
| **P1** | `layouts/ConsoleLayout.vue` (`NLayoutSider` `show-trigger`) | Collapse control is Naive **icon trigger** — no custom `aria-label` in app code | Pass trigger slot with labeled button or document Naive default SR name in QA |
| **P2** | `views/PlaceholderView.vue` + 7 router entries | Same `NResult` title/body for **all** placeholder routes — SR/history may not distinguish “Keys” vs “Billing” | Pass `route.meta.title` into result title |
| **P2** | App shell | **No skip-to-content** link; main content is `NLayoutContent` without explicit `main`/landmark override | Add skip link → `#main` on content root |
| **P2** | Tables (`NDataTable` in channels/models/logs) | Column titles present via script columns (OK baseline); **no** empty-state / sort button a11y audit beyond Naive defaults | Manual keyboard pass on sort/pagination |
| **P2** | Status presentation (tags/alerts) | Errors use `NAlert type="error"` **with text** (good); channel/model **status** may still be color-primary for quick scan | Ensure status column always has text, not color-only |
| **OK** | `LoginView.vue` | `NFormItem` labels for user/pass · autocomplete set · submit button has text | Keep; consider `type="submit"` native association if form grows |
| **OK** | Buttons with visible text | Logout / Refresh / Search / Submit — not icon-only | — |
| **OK** | Footer links | Text visible · `rel="noopener noreferrer"` | — |
| **OK** | Images | **No** `<img>` in scanned SFCs | — |
| **OK** | Explicit `aria-*` in app | **None** — relies on Naive defaults | Document as residual risk for custom icon controls |

## Explicit non-goals this knife

- No code fixes (scan-only · keep diff small for coord harvest)  
- No axe/playwright browser run (no e2e creds / G2 blocked)  
- No React LEGACY touch  
- **D7 NOT EXECUTED**

## Risk (one line)

**Risk:** console is usable for sighted keyboard users on labeled buttons, but **search/filter controls lack accessible names** and **placeholder routes share one SR title**, so keyboard/SR admin ops on list pages and nav stubs will under-communicate without a small label pass.

## Related

| Path | Role |
|------|------|
| `docs/ops/th-console-quality-evidence-2026-07-24.md` | prior quality · thin unit tests |
| `docs/legacy-frontend-gate.md` | new UI only `web-console/` |
| findings digest | stack review · no a11y P0 |

## Agent close

```
DONE + in-review · D7 NOT EXECUTED · no push main
```

Coord: **th-coord** · G2 remains **blocked** without `TH_E2E_*` (unrelated to this module).

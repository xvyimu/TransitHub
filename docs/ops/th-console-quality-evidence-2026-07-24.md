# TH console quality · evidence · **2026-07-24** · **D7 FLIP: NOT EXECUTED**

> Module: **M-TH-console-quality**  
> Scope: `web-console/` quality gates + debt inventory · **no** production `FRONTEND_MODE` · **no** React dual-write · **no** push.  
> Gate stack: [`docs/PROJECT.md`](../PROJECT.md) · [`docs/legacy-frontend-gate.md`](../legacy-frontend-gate.md) · [`w3-d7-gate-dossier.md`](./w3-d7-gate-dossier.md) · [`w4-d7-nonprod-verify.md`](./w4-d7-nonprod-verify.md).

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-console-quality`（本机路径，可移植性无保证） |
| Branch | `xvyimu/th-console-quality` |
| Tip (start) | `f7a8b9bd` |
| Date | **2026-07-24** |
| Agent | claude |
| Secrets in git | **none** (`TH_E2E_*` unset this session) |

---

## 1. Exit code table (this message)

Run from worktree root unless noted.

| # | Command | CWD | Exit | Notes |
|---|---------|-----|-----:|-------|
| 1 | `pnpm install --frozen-lockfile` | `web-console/` | **0** | pnpm 11.5.0 · lockfile up to date · 145 packages |
| 2 | `pnpm typecheck` | `web-console/` | **0** | `vue-tsc -b --pretty false` clean |
| 3 | `pnpm test` | `web-console/` | **0** | Vitest 3.2.7 · **1** file · **5** tests · `src/api/logQuery.test.ts` |
| 4 | `pnpm build` | `web-console/` | **0** | `vue-tsc -b && vite build` · dist written · chunk-size warning only (`index-*.js` ~1.14 MB / gzip ~300 kB) |
| 5 | `pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild` | repo root | **10** | honest block · no fake green |

### W4 SUMMARY (redacted)

```text
SUMMARY exit=10  healthz=0 status=0 contract=0 login=10 channels=10 console_build=skip backend_build=skip
BLOCK credentials incomplete — missing: TH_E2E_USER + TH_E2E_PASS
```

| Step | Exit / result |
|------|----------------|
| healthz | **0** (HTTP 200) |
| `/api/status` | **0** (HTTP 200) |
| `validate-console-contract.py` | **0** (`ops_found=8 schemas_found=8`) |
| login (G2) | **10** (creds unset) |
| channels RO (G3 live) | **10** (blocked on login) |
| console / backend build | **skip** (flags) |

**Interpretation:** console static quality **green**; G2/G3 live **blocked** (operator creds). **D7 NOT EXECUTED.**

---

## 2. Console 债单（quality debt）

### 2.1 Typecheck

| Item | Status |
|------|--------|
| `vue-tsc -b` on current tree | **green** (exit 0) |
| Strict gaps | Index signatures on `UserSelf` / `ChannelItem` / `LogItem` / `StatusData` / `ModelItem` (`[key: string]: unknown`) weaken exhaustiveness — intentional for backend shell drift, but hides field typos |
| View list normalizers | `ChannelsView.normalizeListBody` uses loose `as` casts; no shared typed unwrapper for list shells |

**No typecheck errors this run.** Debt is structural looseness, not failing `tsc`.

### 2.2 Test 缺口

| Area | Coverage now | Gap |
|------|--------------|-----|
| `logQuery` pure helpers | **5** unit tests | Only success-shell / type-param helpers — not full `listLogs` HTTP path |
| `safeRedirect` (open-redirect guard) | **none** in vitest | Logic is pure in `router/index.ts` — easy unit target, untested |
| `http` interceptors (`New-Api-User`, 401 handler) | **none** | Relies on live e2e / manual |
| `auth` store (bootstrap / login / 2FA branch) | **none** | 2FA fail-path only messaged, not tested |
| `listChannels` keyword → `/search` vs list | **none** | Branch untested |
| Vue SFCs / Naive tables | **none** | Vitest env is `node` only — no component mount suite |
| Client key-omission invariant (G3) | **none** in unit | W4 pack checks live response when authed; no offline fixture test |
| E2E | PowerShell scripts only | Needs `TH_E2E_*`; W4 exit **10** this session |

**Summary:** test suite is a **single pure-helper file**. Static gates pass; behavioral regression surface is thin outside live harness.

### 2.3 Feature / route placeholders (strangler debt)

| Route | Component | Notes |
|-------|-----------|--------|
| `/keys` | `PlaceholderView` | nav present · no CRUD |
| `/users` | `PlaceholderView` | |
| `/billing` | `PlaceholderView` | **must not** dual-write React |
| `/settings` | `PlaceholderView` | |
| `/system` | `PlaceholderView` | |
| `/playground` | `PlaceholderView` | |
| `/profile` | `PlaceholderView` | |

Implemented RO / MVP surfaces: **login · health · channels (RO) · models · logs**. Auth store documents **2FA not in Phase1 MVP** (redirect to legacy console message).

### 2.4 Fragile points

1. **Session coupling:** cookie + `localStorage` `uid` + `New-Api-User` header must stay aligned with backend `UserAuth` (legacy React contract). Mismatch → silent 401 loops.
2. **Channels RO trust model:** UI never renders a `key` column; **key omission is backend-enforced**. Client does not assert absence of key-like fields on list items (W4 live check does when G2 green).
3. **Bundle weight:** main chunk >500 kB (Naive + i18n + axios). Vite warns; not a gate fail — risk for first-load on constrained admin networks.
4. **axios pin `^1.18.1`:** resolves to 1.18.x tree this install; track CVE/advisory separately from this quality pass.
5. **Live G2/G3 blocked without `TH_E2E_*`:** static green ≠ flip-ready (see dossier G2/G3/G8).
6. **No component tests + Placeholder majority of nav:** cutover G3/G8 still depend on operator live smoke, not CI unit depth.

### 2.5 Explicit non-goals this knife

- No production `FRONTEND_MODE` / D7 flip  
- No delete or feature work on `web/default` / `web/classic`  
- No dual-write of same screen on React + Vue  
- No backend Go changes  
- No secrets written to repo  
- No `git push`

---

## 3. Risk (one line)

**Risk:** console CI gates are green, but unit coverage is almost only `logQuery` helpers and G2/G3 live remain blocked without non-prod `TH_E2E_*` — a cutover decision based on typecheck/build alone would still ship unauthenticated e2e debt and large placeholder surface.

---

## 4. Related

| Path | Role |
|------|------|
| [`w4-d7-nonprod-verify.md`](./w4-d7-nonprod-verify.md) | Operator pack + exit codes |
| [`w3-d7-gate-dossier.md`](./w3-d7-gate-dossier.md) | G1–G8 status |
| [`th-e2e-gate-2026-07-24.md`](./th-e2e-gate-2026-07-24.md) | Prior exit-10 baseline |
| [`wave8-th-claude.md`](./wave8-th-claude.md) | Earlier console smoke table (2026-07-22) |
| `web-console/README.md` | Dev / acceptance commands |

**DONE · in-review · D7 NOT EXECUTED · no push**

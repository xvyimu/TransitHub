# TH · LEGACY gate scan evidence · **D7 NOT EXECUTED**

| Field | Value |
|-------|--------|
| Module | `M-TH-legacy-gate-scan` |
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-legacy-gate-scan` |
| Branch | `xvyimu/th-legacy-gate-scan` |
| Tip (scan) | `f7a8b9bd` |
| Date | **2026-07-24** |
| Gate SSOT | [`docs/legacy-frontend-gate.md`](../legacy-frontend-gate.md) (Resolution **B**, 2026-07-22 · intro commit `3fbaf691`) |
| Stack SSOT | [`docs/PROJECT.md`](../PROJECT.md) §2.2 — React `web/default` = **LEGACY-HOTFIX only**; new admin UI → `web-console/` |
| D7 flip | **NOT EXECUTED** |
| `git push` | **not done** |
| Code change this module | **none** (scan + evidence only) |

---

## 1. Scope

| In | Out |
|----|-----|
| `web/default/src/features/**` inventory | Implementing features on React |
| `web/default/src/routes/**` surface | Deleting / rewriting `web/default` |
| `git log` on `web/default/` (recent + post-gate) | Production cutover / `FRONTEND_MODE` |
| Compliance vs LEGACY gate + classic freeze | Dual-write into Vue (not claimed here) |

---

## 2. Features directory inventory (`web/default/src/features/`)

Top-level feature packages (23) and approximate file counts:

| Feature dir | ~files | Role (route-backed?) |
|-------------|-------:|----------------------|
| `about` | 3 | yes · `/about` |
| `auth` | 37 | yes · sign-in/up, oauth, otp, passkey, reset |
| `channels` | 63 | yes · `/channels` |
| `chat` | 4 | yes · `/chat/$chatId`, chat2link |
| `dashboard` | 38 | yes · `/dashboard` |
| `errors` | 5 | yes · error pages |
| `home` | 21 | yes · public home |
| `keys` | 20 | yes · `/keys` |
| `legal` | 6 | yes · privacy / user-agreement |
| `models` | 40 | yes · `/models` |
| `performance-metrics` | 3 | **no dedicated route** — API/types used by dashboard + pricing panels |
| `playground` | 47 | yes · `/playground` |
| `pricing` | 39 | yes · `/pricing` |
| `profile` | 29 | yes · `/profile` |
| `rankings` | 14 | yes · `/rankings` |
| `redemption-codes` | 17 | yes · `/redemption-codes` |
| `setup` | 9 | yes · `/setup` |
| `subscriptions` | 18 | yes · `/subscriptions` |
| `system-info` | 5 | yes · `/system-info` |
| `system-settings` | 129 | yes · multi-section settings |
| `usage-logs` | 36 | yes · `/usage-logs` |
| `users` | 18 | yes · `/users` |
| `wallet` | 28 | yes · `/wallet` |

**Scan note:** No *new* top-level feature package appears after gate intro (`3fbaf691`, 2026-07-22). Tree is the existing production React admin surface, not a fresh parallel console.

---

## 3. Routes surface (`web/default/src/routes/`)

**61** route module files. High-level map:

| Area | Paths (relative to `routes/`) |
|------|-------------------------------|
| Root / public | `__root.tsx`, `index.tsx`, `about/`, `pricing/`, `rankings/`, `setup/`, `privacy-policy`, `user-agreement`, `oauth/$provider` |
| Auth | `(auth)/sign-in`, `sign-up`, `register`, `otp`, `oauth`, `forgot-password`, `reset`, `user/reset` |
| Errors | `(errors)/401|403|404|500|503` |
| Authenticated core | `channels`, `keys`, `users`, `wallet`, `profile`, `subscriptions`, `redemption-codes`, `playground`, `system-info`, `chat/*`, `chat2link` |
| Dashboard / models / logs | `dashboard/`, `models/`, `usage-logs/` (+ `$section` variants) |
| System settings | `system-settings/{auth,billing,content,models,operations,security,site}/` |
| Console aliases | `console/log`, `console/topup` |

**vs classic freeze rule:** `web/classic` remains L2 frozen (no feature-parity work). This scan did **not** find a *new* route tree under `web/default` that only exists to chase classic parity after the gate date — routes are the long-standing default admin graph.

**Post-2026-07-01 route *adds*:** `git log --diff-filter=A -- web/default/src/routes/` returned **empty** in this worktree (no new route files added in that window).

---

## 4. `git log --oneline -20 -- web/default/`

```
a72c558c feat(web/default): V2 Atelier A0/A1 structural chrome
869d200d chore: independent TransitHub stack (module path + remotes + identity)
81cb7063 fix(web): type sanitize options as DOMPurify Config
a423a161 feat(web): refund intents read-only card on channels page
ca0788a8 feat(pricing): probe price source, sub2api adapter, ratio snapshots
c7aef084 feat(web): localize channel health UI and show call failure reasons
aa59d987 feat(web): show channel failure signals with error-log deep links
11eac8e8 feat(web): ops strip cold-start empty state and shadow badge
da2ad493 fix(web): route default sanitize through DOMPurify; hard-fail vitest on release
4e23f53c feat: refund outbox, channel health metrics, and backup restore drill
bfd7186f feat: land P1 UI permissions, error visibility, vitest, and circuit/proxy hardening
c9883a6a feat(web): ops-oriented channel and dashboard UX shortcuts
20cbb7d9 refactor(channel): simplify merge plan path and reuse helpers
7e644381 feat(channel): one-click merge of duplicate channels into multi-key
ba0744ca fix: suppress session-expired noise during intentional sign-out
f8a1b8ef fix: clear session cookie and hard-redirect after sign-out
415e1ef8 fix: unwrap nested translation namespace for default frontend i18n
a85e4300 fix: force Chinese UI default past browser English detection
30a8e835 fix: default UI and API locale fallback to Simplified Chinese
7c41ddf6 fix: sanitize legacy frontend HTML
```

### 4.1 This branch vs `main`

```text
git diff --stat main...HEAD -- web/default/
# (empty)
```

`merge-base(HEAD, main) == HEAD` (`f7a8b9bd`). **No local `web/default` delta** on `xvyimu/th-legacy-gate-scan` relative to `main`.

### 4.2 Commits after gate intro (`3fbaf691` → `HEAD` on `web/default/`)

| Commit | Date | Touch | Gate read |
|--------|------|-------|-----------|
| `a72c558c` | 2026-07-23 | layout chrome + `theme.css` + card tokens | **Cosmetic / structural chrome** — not a new product screen; still **not** labeled `LEGACY-HOTFIX` |
| `869d200d` | 2026-07-22 | `features/about` identity blurb | **Branding/identity** (TransitHub stack independence) — low product-feature risk |
| `81cb7063` | 2026-07-21* | DOMPurify typing | **Security/typing** — allowed class |
| `a423a161` … `7e644381` | 2026-07-20–21 | channel ops UI, refund card, dashboard ops strip, merge dialog, … | **Pre-gate product features** on LEGACY (see §5) |

\*Ordering: gate commit date 2026-07-22; several `feat(web)` land immediately before the written gate.

---

## 5. Conclusion table

| Verdict | Path / item | Reason |
|---------|-------------|--------|
| **合规** | Branch `xvyimu/th-legacy-gate-scan` vs `main` · `web/default` | Empty diff — this module did not extend LEGACY |
| **合规** | `web/default/src/routes/**` (post-2026-07-01) | No new route files added |
| **合规** | Security/i18n/auth session fixes (`7c41ddf6`, `da2ad493`, `81cb7063`, `f8a1b8ef`, `ba0744ca`, locale defaults) | Match **Allowed**: security / production-hygiene / typo-class |
| **合规（结构）** | Feature package set (23 dirs) | Existing prod admin map; no post-gate *new* feature package |
| **可疑** | `web/default/src/features/channels/components/refund-intents-card.tsx` · `a423a161` | **New product UI surface** (read-only card) on LEGACY channels page; gate forbids new tables/forms/features without cutover exception / `LEGACY-HOTFIX` |
| **可疑** | `…/dashboard/components/overview/ops-health-strip.tsx` · `c9883a6a` / `11eac8e8` / `4e23f53c` | Ops-oriented **dashboard UX feature** growth on React after strangler direction was known |
| **可疑** | `…/channels/.../channel-merge-dialog.tsx` · `7e644381` | **New dialog/workflow** on LEGACY channels (merge multi-key) — product feature, not security hotfix |
| **可疑** | `…/channels/.../channel-failure-strip.tsx` · `aa59d987` / `c7aef084` | New failure-signal UI + deep links on LEGACY |
| **可疑（轻）** | `a72c558c` Atelier A0/A1 chrome | Theme/header/footer polish **after** gate without `LEGACY-HOTFIX` label; low product risk but process drift |
| **观察** | `features/performance-metrics/` | Support module only (no route); not a new screen by itself |
| **观察** | `web/classic/**` | L2 frozen; not expanded in this scan’s `web/default` delta |
| **D7** | Production Vue flip / `FRONTEND_MODE` | **NOT EXECUTED** (out of module scope; not authorized) |

### 5.1 Overall

| Question | Answer |
|----------|--------|
| Does *this branch* violate the gate by shipping new LEGACY product UI? | **No** — no `web/default` changes vs `main`. |
| Does *current tree history* show LEGACY product growth around gate week? | **Yes** — several `feat(web)` items (channels/dashboard) land **just before** gate formalization; post-gate only chrome/identity/security-class touches. |
| New screens / routes vs classic freeze? | **No new routes** found in the scanned window; growth is **in-page components** on existing screens, not a new SPA area. |
| Dual-write React+Vue same feature? | Not proven by this scan (would need feature-by-feature `web-console` parity audit). Gate rule remains: **new** work only in `web-console/`. |

**Process recommendation (ops only · not implemented here):** future `web/default` PRs require `LEGACY-HOTFIX` + incident link per gate checklist; prefer porting ops cards (refund intents, health strip) to `web-console` rather than deepening React.

---

## 6. Verification commands (re-runnable)

```powershell
# CWD: repo root (this worktree)
pwsh -NoProfile -Command @'
Set-Location "C:\Users\yuanjia\orca\workspaces\src\th-legacy-gate-scan"

# 1) features top-level
Get-ChildItem web/default/src/features -Directory |
  ForEach-Object { "{0}`t{1}" -f $_.Name, (Get-ChildItem $_.FullName -Recurse -File).Count }

# 2) routes inventory
Get-ChildItem web/default/src/routes -Recurse -File -Filter *.tsx |
  ForEach-Object { $_.FullName.Substring((Resolve-Path web/default/src/routes).Path.Length+1) }

# 3) recent history
git log --oneline -20 -- web/default/

# 4) branch vs main (expect empty)
git diff --stat main...HEAD -- web/default/

# 5) post-gate LEGACY touches
git log 3fbaf691..HEAD --oneline -- web/default/

# 6) new route files since 2026-07-01 (expect empty)
git log --since=2026-07-01 --diff-filter=A --name-only --pretty=format: -- web/default/src/routes/

# 7) gate + stack docs present
Test-Path docs/legacy-frontend-gate.md
Test-Path docs/PROJECT.md
'@
```

Recorded outcomes this session:

| Check | Result |
|-------|--------|
| Features dirs | **23** listed §2 |
| Route files | **61** |
| `main...HEAD` `web/default` | **empty** |
| New routes since 2026-07-01 | **none** |
| Post-gate `web/default` commits | `a72c558c`, `869d200d`, (+ earlier listed in log) |

---

## 7. Document gap

Gate text already forbids “new product features / tables / forms” on LEGACY. No mandatory one-line code/doc patch required for this scan: historical `feat(web)` drift is **evidence**, not a missing rule.

Optional (not applied): PR template already referenced in gate; enforce label `legacy-ui` in CI remains “proposed”.

---

## 8. Hand-off

| Item | Value |
|------|--------|
| Status | **DONE** · **in-review** |
| Evidence | this file |
| Commit | `938208ae` |
| Coord | **th-coord** |
| D7 | **NOT EXECUTED** |
| Push | **not done** |

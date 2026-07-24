# M-TH-g6-soak-checklist · evidence · 2026-07-24

## D7 FLIP: NOT EXECUTED

Production `FRONTEND_MODE` **not** changed. No production migrate. No `git push`. No React delete. No dual React+Vue public origin.

## Worktree identity

| Field | Value |
|-------|--------|
| Module ID | **M-TH-g6-soak-checklist** |
| Worktree (absolute) | `C:\Users\yuanjia\orca\workspaces\src\th-g6-soak-checklist` |
| Branch | `xvyimu/th-g6-soak-checklist` |
| Tip (start / evidence base) | `f7a8b9bde34ff8c2a9b9683b1d1ad59970b6c3b0` (`docs(ops): TH E2E operator card`) |
| Tip (post evidence) | this commit |
| Agent | claude |
| Scope | G6 soak **checklist packaging + half-probe only** · `docs/ops/` · **no** claim of ≥24h soak green |
| Date | **2026-07-24** |
| Status | **DONE** · **in-review** (th-coord) |

## Boundary

| In | Out |
|----|-----|
| Cite / package [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) critical rows | Running a real **≥24h** staging soak |
| Instant local probes: `GET /healthz` · `GET /api/status` (+ related probes) with timestamp | Filling T+6 / T+12 / T+24 as green without observation |
| Honest G6 status + operator unblock | Production traffic · dual public React+Vue · secrets in git |
| Docs under `docs/ops/` | `git push` · production `FRONTEND_MODE` · **D7** |

## Pre-read

| Path | Result |
|------|--------|
| [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) | Read — SSOT for T0 / 24h observation / end-of-soak |
| [web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) § G6 | Read — G6 = staging soak ≥24h; login + health + channels RO; no dual public URLs |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) § G6 | Read — status already **blocked (not run)** W3/W4 |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) | Read — pre-soak smoke / exit codes |

---

## 1 · G6 status (authoritative for this module)

| Field | Value |
|-------|--------|
| **G6 status** | **blocked (not run full soak)** |
| Operator filled checklist? | **No** — all soak boxes remain ☐ in SSOT; no operator summary attached |
| Fake green? | **Forbidden** — this doc does **not** promote G6 to green |
| Dossier alignment | Matches [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G6 · GATE-MATRIX long-wave |

> Half-probe below is **instantaneous** (seconds). It is **not** a ≥24h continuous soak and must not be used as G6 pass evidence.

---

## 2 · Soak checklist — critical rows (copy from SSOT)

Source of truth: [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md).  
**Pass column left unchecked** — agent did not own staging for 24h.

### Start-of-soak (T0)

| # | Item | Pass (this module) |
|---|------|--------------------|
| 1 | Image digests recorded (backend + Vue frontend) | ☐ not done |
| 2 | `FRONTEND_MODE=disabled` on backend | ☐ not verified on staging |
| 3 | `GET /frontend-healthz` → ok | ☐ (local integrated process → **404** — see §3) |
| 4 | `GET /healthz` · `/livez` · `/readyz` acceptable | ☑ local instant only · **not** soak T0 |
| 5 | Login e2e exit 0 (`scripts/e2e-web-console-login.ps1` / W4 pack) | ☐ `TH_E2E_*` unset · G2 blocked |
| 6 | Channels RO page lists; **keys absent** | ☐ needs G2 session |
| 7 | Public `/metrics` → **404** | ☑ local instant only · **not** soak T0 |
| 8 | Single public origin (no dual React+Vue URLs) | ☐ staging topology not owned by agent |

### 24h observation (T+6h · T+12h · T+24h)

| # | Signal | Pass criteria | T+6h | T+12h | T+24h |
|---|--------|---------------|------|-------|-------|
| 1 | **5xx** rate | No sustained spike vs pre-soak baseline | ☐ | ☐ | ☐ |
| 2 | **4xx** (auth) | No unexpected surge on login/self | ☐ | ☐ | ☐ |
| 3 | **Login** | Manual login + optional e2e still green | ☐ | ☐ | ☐ |
| 4 | **Channels RO** | List usable; no key leakage in UI/network | ☐ | ☐ | ☐ |
| 5 | **Probes** | healthz/livez/readyz/frontend-healthz healthy | ☐ | ☐ | ☐ |
| 6 | **Session** | Cookie session survives normal browser use | ☐ | ☐ | ☐ |
| 7 | **SSE/stream** (if used on staging) | No new disconnect pattern vs baseline | ☐ / n/a | ☐ / n/a | ☐ / n/a |
| 8 | **Logs** | No continuous panic / fatal; quota saturation not flooding | ☐ | ☐ | ☐ |
| 9 | **Edge metrics** | `/metrics` still blocked on public origin | ☐ | ☐ | ☐ |

### End-of-soak (before requesting D7)

| # | Item | Pass |
|---|------|------|
| 1 | All critical rows above green for full window | ☐ |
| 2 | G7 rollback drill recorded (or scheduled immediately after) | ☐ |
| 3 | Soak notes linked from [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G6 | ☐ (this evidence links checklist; **not** end-of-soak) |
| 4 | No open P0/P1 on console login / channels RO | ☐ |

---

## 3 · Half-probe (agent host · **not** ≥24h soak)

| Field | Value |
|-------|--------|
| Host | Agent worktree machine · **not** dedicated staging edge |
| Base | `http://127.0.0.1:3000` (default `TH_API_BASE`) |
| Listener | TCP **3000** LISTEN pid **12188** |
| `TH_E2E_*` | **unset** |
| Topology note | `/frontend-healthz` → **404** → process looks like **integrated** backend, **not** Vue-edge `frontend_external` separated stack required for G6 soak |
| Claim | **Instant smoke only** · seconds · **≠** continuous ≥24h observation |

### Sample A · 2026-07-24 13:38:20 +08:00

| URL | HTTP | Body (truncated / non-secret) |
|-----|-----:|--------------------------------|
| `GET /healthz` | **200** | `{"plane":"all","status":"ok"}` |
| `GET /livez` | **200** | (ok) |
| `GET /readyz` | **200** | (ok) |
| `GET /api/status` | **200** | JSON envelope `data` present · len≈1792 |
| `GET /frontend-healthz` | **404** | not served on this process |

### Sample B · 2026-07-24 13:38:47 +08:00

| URL | HTTP | Notes |
|-----|-----:|-------|
| `GET /livez` | **200** | `{"plane":"all","status":"ok"}` |
| `GET /readyz` | **200** | `{"status":"ok"}` |
| `GET /metrics` | **404** | matches public-edge expectation when metrics blocked |

### Sample C · 2026-07-24 13:39:05 +08:00 · `/api/status` shape (no secrets)

| Field | Value |
|-------|--------|
| HTTP | **200** |
| `data` key count | **68** |
| `version` (public) | `v1.0.0-rc.21-81-gd1397d6e` |
| `system_name` | `New API` |
| `demo_site_enabled` | `false` |
| `self_use_mode_enabled` | `true` |
| `start_time` | `1784860977` (epoch as returned) |

No passwords, tokens, channel keys, or cookie values recorded.

### What this half-probe does **not** prove

| Claim | Status |
|-------|--------|
| ≥24h continuous staging | **Not proven** |
| Vue edge + `FRONTEND_MODE=disabled` topology | **Not proven** (`/frontend-healthz` 404) |
| Login e2e / channels RO | **Not run** (G2 blocked) |
| 5xx/4xx baseline comparison | **Not run** |
| G6 green | **No** |

---

## 4 · Operator unblock (staging ownership)

Owner: **operator / SRE who owns non-prod staging** (not this agent worktree).

### Preconditions (before T0)

1. Deploy **non-prod only**: same-origin Nginx Vue SPA → Go `:3000` with **`FRONTEND_MODE=disabled`** / `frontend_external` binary.  
2. **Do not** dual-publish React + Vue on the same public origin during soak.  
3. Record image digests (backend + Vue frontend).  
4. Export non-prod `TH_E2E_USER` / `TH_E2E_PASS` (never commit) — see [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md).  
5. Prefer pre-soak pack:  
   `pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild`  
   Expect overall **0** only when G2/G3 live steps green; **exit 10** without creds is honest (not pass).

### T0 checklist (must all ☑)

Copy rows from §2 Start-of-soak; run login e2e + channels RO key-omission; confirm `/frontend-healthz` ok on **edge**; public `/metrics` **404**.

### Observation window

- Duration: **≥ 24 hours continuous** on that staging stack.  
- Checkpoints: at least **T+6h · T+12h · T+24h** against §2 24h table (5xx/4xx/login/channels/probes/session/SSE/logs/metrics).  
- Keep notes/metrics exports outside git if they contain secrets; paste **summary only** into dossier.

### Close soak → request D7 path

1. End-of-soak rows §2 all green.  
2. Schedule / record **G7** rollback drill ([w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md)).  
3. Link soak notes into [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G6.  
4. Still require **G8** human phrase `D7 flip 现在` — this module does **not** flip.

### Explicit out of scope (unchanged)

| Item | Status |
|------|--------|
| Production traffic | Forbidden until G8 |
| Production DB migrate | Forbidden |
| Production passwords in agent reports | Forbidden |

---

## 5 · Intentionally not done

| Item | Status |
|------|--------|
| **Full ≥24h soak** | **Not run** |
| Marking G6 **green** | **Not done** (would be fake green) |
| **D7 production flip** | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched |
| `git push` / merge default branch | Not done |
| Starting a long-running soak from agent host | Out of boundary (no staging ownership) |

## 6 · Outcome

| Claim | Evidence |
|-------|----------|
| Checklist critical rows packaged | §2 tables = SSOT copy with honest ☐ |
| Instant healthz + status sampled | §3 timestamps + HTTP codes |
| Explicitly **not** 24h soak | §1 · §3 · this row |
| G6 remains **blocked (not run full soak)** | §1 |
| Operator unblock written | §4 |
| D7 / push / FRONTEND_MODE | **NOT EXECUTED** / no push / untouched |

## Related

| Path | Role |
|------|------|
| [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) | Soak SSOT |
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G6 dossier status |
| [web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | G1–G8 definitions |
| [w4-d7-nonprod-verify.md](./w4-d7-nonprod-verify.md) | Pre-soak nonprod pack |
| [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) | G7 desktop drill |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | Non-prod env names |

## Handoff · th-coord

- **Status:** DONE + **in-review**
- **Ask:** accept G6 evidence as **checklist + half-probe packaging**; keep G6 **blocked** until staging owner runs full soak
- **Do not:** D7 · push · production FRONTEND_MODE · treat this as soak green

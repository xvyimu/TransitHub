# W3 · D7 gate dossier (TransitHub) · **D7 FLIP: NOT EXECUTED**

> **Purpose:** Executable pre-flip package for G1–G8.  
> **Does not:** change production `FRONTEND_MODE`, delete React, migrate live DB, or flip traffic.  
> **Requires for live D7:** this dossier green **and** explicit human phrase `D7 flip 现在` (or equivalent).  
> Recorded: **2026-07-23** · worktree `C:\Users\yuanjia\orca\workspaces\src\w3-th-claude` · branch `xvyimu/w3-th-claude` · tip `b2fff447`.

## Summary

| Gate | Status | Owner | Gap |
|------|--------|-------|-----|
| **G1** Module2 on tip | **green** | platform / TH maintainers | — |
| **G2** Login e2e (non-prod) | **blocked** | operator (creds) | `TH_E2E_USER` / `TH_E2E_PASS` unset; default `root/123456` rejected on live `:3000` |
| **G3** Channels RO | **blocked** (depends G2) | operator + console | live list needs G2 session; **contract green** |
| **G4** Vue image | **blocked** (agent host) / **CI SSOT green path** | CI + operator with Docker | local `docker` not on PATH; CI job `image-reproducibility` builds `Dockerfile.frontend.vue` |
| **G5** Backend external | **green** | platform | re-verified W3 exit 0 |
| **G6** Staging soak ≥24h | **blocked** (no staging soak run) | operator / SRE | checklist ready — not executed this wave |
| **G7** Rollback drill | **blocked** (desktop drill doc only) | operator | non-prod drill steps ready; **not** executed (no docker / no staging stack) |
| **G8** Owner sign-off | **blocked** | **human** | no “cutover now” / “D7 flip 现在” in W3 |

**Flip readiness:** **NO** — G2/G3/G4(local)/G6/G7/G8 open. Production flip **forbidden** until green + human gate.

---

## G1 · Module2 on tip · **green**

| Field | Value |
|-------|--------|
| Status | **green** |
| Owner | TH maintainers |
| How to prove | Tree presence on default-branch tip / this worktree |
| Evidence (W3) | All present: `web-console/` · `migrations/` · `docs/gateway/` · `docs/openapi/console-subset.yaml` · `docs/gateway/CONSOLE_API_CONTRACT.md` · `docs/operations/web-console-cutover-plan.md` · `docs/operations/web-console-cutover-rollback.md` · `deploy/separated/Dockerfile.frontend.vue` · `deploy/separated/Dockerfile.frontend` · `deploy/separated/nginx.conf.template` |
| Gap | None |
| Commands | `Test-Path` list (W3 report) — all OK |

---

## G2 · Login e2e (non-prod) · **blocked**

| Field | Value |
|-------|--------|
| Status | **blocked** (credentials) |
| Owner | Operator who controls **non-prod** admin password |
| How to prove | `pwsh -NoProfile -File scripts/e2e-web-console-login.ps1 -SkipVite` → exit **0** |
| Credentials | Env names only — see [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md). **Never** commit secrets. |
| Evidence (W3) | Backend `http://127.0.0.1:3000/healthz` → **200**. Login as default `root` → `用户名或密码错误…`. Exit **1**. `TH_E2E_USER`/`TH_E2E_PASS` **unset**. **Not production.** |
| Gap | Operator must export non-prod `TH_E2E_USER` / `TH_E2E_PASS` (or seed empty local DB via setup wizard). |
| Unblock | Follow w2 credentials checklist § mint non-prod account; re-run e2e; attach exit 0 to this dossier. |

---

## G3 · Channels RO · **blocked** (live) / contract **green**

| Field | Value |
|-------|--------|
| Status | **blocked** live · **green** machine contract |
| Owner | Operator (session) + console owners |
| How to prove | Authenticated `GET /api/channel/` same-origin; response lists channels; **keys omitted** from list items |
| Evidence | OpenAPI `docs/openapi/console-subset.yaml` op `getChannelsList` · human `CONSOLE_API_CONTRACT.md` §3 · `python scripts/validate-console-contract.py` exit **0** (W3) |
| Gap | Live list requires G2 cookie / `New-Api-User` session |
| Unblock | After G2 exit 0: manual or scripted `GET /api/channel/`; confirm no key material in body |

---

## G4 · Vue image builds · **CI SSOT** / local **blocked**

| Field | Value |
|-------|--------|
| Status | **blocked** on agent host · **CI is SSOT** |
| Owner | CI (`image-reproducibility`) · local operator with Docker Desktop |
| How to prove (CI) | `.github/workflows/quality.yml` job `image-reproducibility` step **Build separated Vue console image** (`-f deploy/separated/Dockerfile.frontend.vue`) + nginx `-t` on both React and Vue images |
| How to prove (local) | `docker build -f deploy/separated/Dockerfile.frontend.vue -t new-api-frontend-vue:local .` exit 0 |
| Evidence (W3) | `docker` **not on PATH** → local exit n/a (`EXIT_DOCKER=1`). Dockerfile present. CI job definition present on tip. |
| Gap | This agent cannot re-run docker build. Operators with Docker should record digest on non-prod. |
| Unblock | Run CI on branch/PR **or** local docker build; paste job URL / image id into this dossier |

---

## G5 · Backend `frontend_external` · **green**

| Field | Value |
|-------|--------|
| Status | **green** |
| Owner | platform |
| How to prove | `go build -trimpath -buildvcs=true -tags frontend_external -o <bin> .` exit 0; CI `go-quality` same tag |
| Evidence (W3) | Exit **0** · binary ~89 MB · removed after verify (gitignored `*.exe`) |
| Gap | None for compile gate |

---

## G6 · Staging soak ≥24h · **blocked** (not run)

| Field | Value |
|-------|--------|
| Status | **blocked** — soak **not executed** this wave |
| Owner | Operator / SRE owning staging |
| How to prove | Staging on Vue edge + `FRONTEND_MODE=disabled` backend for **≥24h**; checklist in [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) all critical rows checked |
| Evidence (W3) | Checklist authored only |
| Gap | No dedicated staging stack controlled by this agent; no dual public React+Vue |
| Unblock | Run soak on **non-prod**; attach log/metrics summary + filled checklist |

### Staging soak — what to watch (summary)

| Signal | Pass criteria (suggest) |
|--------|-------------------------|
| 5xx rate | No sustained spike vs baseline; investigate any burst |
| 4xx auth | No unexpected surge on `/api/user/login` / `/api/user/self` |
| Login | Manual + e2e login remain green |
| Channels RO | `/channels` lists; keys absent |
| Probes | `/healthz` `/livez` `/readyz` `/frontend-healthz` OK |
| Metrics edge | Public origin `/metrics` stays **404** |

Full table: [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md).

---

## G7 · Rollback drill · **blocked** (doc ready · not executed)

| Field | Value |
|-------|--------|
| Status | **blocked** — **desktop drill documented**, not run on live staging |
| Owner | Operator |
| How to prove | On **non-prod** only: flip Vue → React (integrated or separated React image) in ≤5 min; login works |
| Evidence (W3) | Step table: [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) · runbook SSOT [web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md) |
| Gap | No docker / no staging compose on this host → cannot time a real flip |
| Unblock | Execute desktop drill on non-prod; record wall-clock and verification exits |

---

## G8 · Owner sign-off · **blocked**

| Field | Value |
|-------|--------|
| Status | **blocked** |
| Owner | **Human product/ops owner** |
| How to prove | Explicit written authorization: `D7 flip 现在` (or cutover-plan “cutover now”) **after** G1–G7 green |
| Evidence (W3) | User W3 prompt: **D7 FLIP NOT EXECUTED** · no production `FRONTEND_MODE` change |
| Gap | Human gate not given |
| Unblock | G1–G7 closed + owner phrase |

---

## Gin / redis (optional W3) · **defer**

| Choice | Decision |
|--------|----------|
| Gin 1.10+ **or** redis v9 | **Neither bumped** in W3 |
| Rationale | D7 dossier is the W3 main knife; Gin ~300+ imports; redis v9 runtime-critical. Portfolio allows dedicated wt later. |
| Spike | [w1-gin-redis-spike.md](./w1-gin-redis-spike.md) (W3 reaffirm) |
| Pins | Gin `v1.9.1` · redis/v8 `v8.11.5` · module proxy latest Gin **v1.12.0** · go-redis/v9 **v9.21.0** |

---

## Explicit non-goals (this dossier)

| Item | Status |
|------|--------|
| Production D7 flip | **NOT EXECUTED** |
| Production `FRONTEND_MODE` | Untouched |
| Delete `web/default` | Not done |
| Production migrate / live DSN | Not done |
| `git push` / merge main | Not done (总控) |
| TH+CP simultaneous production flip | Forbidden by portfolio protocol |

---

## Related

| Path | Role |
|------|------|
| [w3-arch-upgrade-transithub-claude.md](./w3-arch-upgrade-transithub-claude.md) | W3 report + verification exits |
| [w3-rollback-desktop-drill.md](./w3-rollback-desktop-drill.md) | Non-prod rollback step table |
| [w3-staging-soak-checklist.md](./w3-staging-soak-checklist.md) | 24h soak checklist |
| [w2-cutover-e2e-credentials.md](./w2-cutover-e2e-credentials.md) | G2 env names |
| [../operations/web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | G1–G8 definition |
| [../operations/web-console-cutover-rollback.md](../operations/web-console-cutover-rollback.md) | Operator rollback runbook |
| [../PROJECT.md](../PROJECT.md) §2.2 | Frontend transition SSOT |

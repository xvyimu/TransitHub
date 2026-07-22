# W3 · Staging soak checklist (≥24h · non-production)

> **D7 FLIP: NOT EXECUTED.**  
> Use on **staging** after Vue edge + `FRONTEND_MODE=disabled` backend are deployed **non-prod**.  
> Do **not** dual-publish React and Vue on the same public origin during soak.

## Soak window

| Field | Value |
|-------|--------|
| Target duration | **≥ 24 hours** continuous |
| Environment | staging / non-prod only |
| Topology | same-origin Nginx Vue SPA → Go `:3000` (`frontend_external`) |
| W3 agent execution | **Not run** — checklist only |

## Start-of-soak (T0)

| # | Item | Pass |
|---|------|------|
| 1 | Image digests recorded (backend + Vue frontend) | ☐ |
| 2 | `FRONTEND_MODE=disabled` on backend | ☐ |
| 3 | `GET /frontend-healthz` → ok | ☐ |
| 4 | `GET /healthz` · `/livez` · `/readyz` acceptable | ☐ |
| 5 | Login e2e exit 0 (`scripts/e2e-web-console-login.ps1`) | ☐ |
| 6 | Channels RO page lists; **keys absent** | ☐ |
| 7 | Public `/metrics` → **404** | ☐ |
| 8 | Single public origin (no dual React+Vue URLs) | ☐ |

## 24h observation (check at least T+6h · T+12h · T+24h)

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

## End-of-soak (before requesting D7)

| # | Item | Pass |
|---|------|------|
| 1 | All critical rows above green for full window | ☐ |
| 2 | G7 rollback drill recorded (or scheduled immediately after) | ☐ |
| 3 | Soak notes linked from [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) G6 | ☐ |
| 4 | No open P0/P1 on console login / channels RO | ☐ |

## Explicit out of scope

| Item | Status |
|------|--------|
| Production traffic | Forbidden until G8 |
| Production DB migrate | Forbidden |
| Using production passwords in agent reports | Forbidden |

## Related

| Path | Role |
|------|------|
| [w3-d7-gate-dossier.md](./w3-d7-gate-dossier.md) | G6 status |
| [web-console-cutover-plan.md](../operations/web-console-cutover-plan.md) | Gate definitions |
| [deploy/separated/smoke.ps1](../../deploy/separated/smoke.ps1) | Quick edge smoke |

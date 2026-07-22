# Production cutover plan — Vue web-console (NO traffic flip yet)

**Status**: Plan only · **2026-07-22**  
**Decision**: User deferred live cutover; this document is the gate package.  
**Does not**: switch production traffic, delete React, change DNS.

## Goal

Same-origin public console served from Vue `web-console`, backend `FRONTEND_MODE=disabled` (`-tags frontend_external`), with **≤5 min** rollback to embedded React.

## Preconditions (all required before flip)

| # | Gate | How to prove |
|---|------|----------------|
| G1 | Module2 on `main` | `web-console/`, `migrations/`, gateway docs present |
| G2 | Login e2e green | `scripts/e2e-web-console-login.ps1` on non-prod |
| G3 | Channels RO usable | `/channels` lists without keys |
| G4 | Vue image builds | `docker build -f deploy/separated/Dockerfile.frontend.vue` |
| G5 | Backend external build | `go build -tags frontend_external` |
| G6 | Staging soak ≥ 24h | Login + health + channels RO; no dual public URLs |
| G7 | Rollback drill | Flip back to React image/binary once on staging |
| G8 | Owner sign-off | Explicit “cutover now” from you |

## Topology (target)

```text
browser → Nginx (Vue dist)
            ├─ static SPA
            └─ /api /v1 /healthz… → Go :3000
                                    FRONTEND_MODE=disabled
```

Cross-origin SPA + open CORS = **not** the default (cookie / CSRF).

## Cutover steps (when G1–G8 pass)

1. Backup DB; note current image digests.  
2. Deploy backend image/binary with `frontend_external` + `FRONTEND_MODE=disabled`.  
3. Deploy frontend image from `Dockerfile.frontend.vue`.  
4. Smoke: `/frontend-healthz`, `/healthz`, login, `/health`, `/channels`.  
5. Confirm `/metrics` not on public origin.  
6. Soak; watch 4xx/5xx and auth errors.

## Rollback (≤5 min)

See `docs/operations/web-console-cutover-rollback.md`:

- **Fastest**: previous **integrated** React embed image; unset/`auto` `FRONTEND_MODE`.  
- **Alt**: keep external backend; swap frontend image to React `Dockerfile.frontend`.

No SQL down-migration required for UI rollback.

## Explicit non-goals (this plan)

- Deleting `web/default`  
- Long-term dual public React+Vue  
- Turning off AutoMigrate without migration force on live DB  

## Related

| Path | Role |
|------|------|
| `docs/operations/web-console-cutover-rollback.md` | Operator runbook |
| `docs/legacy-frontend-gate.md` | React feature freeze |
| `deploy/separated/Dockerfile.frontend.vue` | Vue image |
| `web-console/E2E.md` | Login e2e |

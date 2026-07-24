# TransitHub · Long Wave · GATE Matrix

> SSOT for cutover gates this long wave. **Inherits W3/W4 + th-d7-scout `35194ff3`** — statuses not greenwashed.  
> Full narrative: `docs/ops/w3-d7-gate-dossier.md` · scout: `docs/ops/th-d7-scout-2026-07-24.md` (branch `xvyimu/th-d7-scout`) · plan: `docs/operations/web-console-cutover-plan.md`  
> **D7 FLIP: NOT EXECUTED** · G0 authorized **D = A+C non-prod** (2026-07-24)

Updated: **2026-07-24** · coord tip `f7a8b9bd` (th-coord) · scout tip `35194ff3`

| Gate | Meaning | Status | Evidence pointer | Unblock |
|------|---------|--------|------------------|---------|
| **G1** | Module2 on tip | **green** | Tree re-check W3/W4 + scout 2026-07-24 (`Test-Path` all True) | — |
| **G2** | Non-prod login e2e | **blocked** | W4/scout pack exit **10** (no `TH_E2E_*`); healthz/status 200 ≠ login green; card: `th-e2e-gate-card.md` | Operator mint non-prod `TH_E2E_*`; re-run `scripts/w4-d7-nonprod-verify.ps1` → login=0 (**≠ D7**) |
| **G3** | Channels RO live | **blocked** live · contract **green** | `validate-console-contract.py` exit **0** (W4+scout); live needs G2 session + key-omission | After G2: pack channels=0 |
| **G4** | Vue image build | **blocked** local · **CI SSOT** | docker absent (scout re-confirm); CI job `image-reproducibility` + `Dockerfile.frontend.vue` | Docker Desktop **or** paste CI job URL/digest into dossier |
| **G5** | `go build -tags frontend_external` | **green** | W4 + scout re-green exit **0** ~85MB | — |
| **G6** | Staging soak ≥24h | **blocked** | checklist only: `w3-staging-soak-checklist.md` | Operator staging + filled checklist (agent cannot fake 24h) |
| **G7** | Rollback drill timed | **blocked** | doc + min seq: `w4-d7-nonprod-verify.md` · `w3-rollback-desktop-drill.md` | Operator Option A/B + wall-clock ≤5m on non-prod |
| **G8** | Owner sign-off | **blocked** | no `D7 flip 现在` | Human only after G1–G7 green — see [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md) |

## G0 authorization (this wave)

| Field | Value |
|-------|--------|
| Choice | **D = A + C non-prod** |
| A | Non-prod cutover evidence pack refresh (no flip) |
| C | Backend stability scout (migrations 3-DB audit · timeout/pool/Redis hot paths) — **no go.mod major bump** |
| Forbidden | production `FRONTEND_MODE` · D7 flip · delete `web/default` · dual-write · push unless ordered |

## Flip policy (coord)

| Action | Allowed? |
|--------|----------|
| Refresh evidence / docs / non-prod scripts | Yes |
| W4 pack exit 0 non-prod | Yes — **≠ D7** |
| Backend read-only audit + small safe fixes if approved later | Scout first (this wave) |
| Change production `FRONTEND_MODE` / delete `web/default` | **No** without G8 phrase |
| `go.mod` Gin/redis major bump | **No** this wave |
| Push / merge default | Only if human orders |

## Workers (dispatched)

| Module | wt name | Scope | Status |
|--------|---------|-------|--------|
| M-TH-console-quality | `th-console-quality` | web-console pnpm typecheck/test/build + debt list + non-prod evidence notes | dispatched |
| M-TH-backend-stable-scout | `th-backend-stable-scout` | migrations 3-DB audit · timeout/pool/Redis hotspot table · **read-only** | dispatched |

## Module backlog (post-worker)

| ID | Area | Notes |
|----|------|-------|
| M-TH-e2e | G2/G3 live | Needs human `TH_E2E_*` |
| M-TH-g4 | G4 digest paste | Needs Docker or CI URL |
| Gin/redis bump | go.mod | **defer** — dedicated wt + full test if ever |

## Related

- [progress.md](./progress.md)
- [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md)
- `docs/ops/w4-arch-upgrade-transithub-claude.md`
- `docs/ops/th-e2e-gate-card.md`
- origin `xvyimu/th-d7-scout` @ `35194ff3`

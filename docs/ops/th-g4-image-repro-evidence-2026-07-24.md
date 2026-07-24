# TH G4 · Vue image reproducibility evidence · **2026-07-24**

> **Status:** **blocked local · CI SSOT**  
> **D7 FLIP: NOT EXECUTED** · no production `FRONTEND_MODE` · no traffic flip · no `git push`  
> Module: **M-TH-g4-image-repro** · Gate: **G4** (Vue separated image)  
> Inherit: [`w3-d7-gate-dossier.md`](./w3-d7-gate-dossier.md) §G4 · [`web-console-cutover-plan.md`](../operations/web-console-cutover-plan.md) G4

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-g4-image-repro`（本机路径，可移植性无保证） |
| Branch | `xvyimu/th-g4-image-repro` |
| Tip (pre-commit) | `f7a8b9bd` (docs(ops): TH E2E operator card) |
| Date | **2026-07-24** |
| Agent host | Windows · `docker` **absent** from PATH · Docker Desktop default path **not installed** |
| Local `docker build` | **not run** (no CLI) — **not** claimed green |

---

## 1. Verdict (honest)

| Check | Result |
|-------|--------|
| `docker` on PATH | **NO** (`Get-Command docker` fails; `where.exe docker` empty; `C:\Program Files\Docker\Docker\resources\bin\docker.exe` missing) |
| Local `docker build -f deploy/separated/Dockerfile.frontend.vue -t new-api-frontend-vue:local .` | **n/a** · exit **not recorded** (would be false-green to invent) |
| Local image id / digest | **n/a** |
| Tree artifacts (Dockerfile + nginx template) | **present** (see §2) |
| CI authority for G4 | **`.github/workflows/quality.yml` job `image-reproducibility`** (see §3) |
| Overall G4 on this host | **blocked local · CI SSOT** |
| D7 | **NOT EXECUTED** |

**Rule (module bound):** 无 docker **不得**写 local green。G4 本地通过必须以本机 `docker build` exit **0** + image id 为准；否则只记 blocked + CI 步骤摘录。

---

## 2. Tree presence (Test-Path · this worktree)

| Path | Present | Notes |
|------|---------|--------|
| `deploy/separated/Dockerfile.frontend.vue` | **True** | Vue3 `web-console` multi-stage → unprivileged nginx; copies `nginx.conf.template` + `docker-entrypoint.sh` |
| `deploy/separated/nginx.conf.template` | **True** | Shared template for React + Vue separated images |
| `deploy/separated/Dockerfile.frontend` | **True** | React legacy image (CI also builds) |
| `deploy/separated/docker-entrypoint.sh` | **True** | (directory listing) |
| `.github/workflows/quality.yml` | **True** | job `image-reproducibility` |

Dockerfile summary (no rebuild):

- Builder: `node:22-bookworm` · `pnpm@11.5.0` · `web-console` `pnpm install --frozen-lockfile` · `pnpm run build`
- Runtime: `ARG NGINX_IMAGE=nginxinc/nginx-unprivileged@sha256:65e3e85d…` (CI overrides with pulled digest)
- Serves dist on **8080** · healthcheck `GET /frontend-healthz` · same-origin API proxy via envsubst template

---

## 3. CI SSOT · `image-reproducibility` (quality.yml)

**File:** `.github/workflows/quality.yml`  
**Job id / name:** `image-reproducibility`  
**Needs:** `go-quality`, `web-quality`, `web-console-quality`, `sqlite-migrate`  
**Runner:** `ubuntu-latest`

| Step name (exact) | What it does | Path / tag |
|-------------------|--------------|------------|
| Check out | `actions/checkout@9c091bb…` | repo root |
| Build monolith image | `docker build --tag new-api:quality .` | root Dockerfile |
| Build pure backend image | `docker build -f Dockerfile.backend --tag new-api-backend:quality .` | `Dockerfile.backend` |
| Resolve nginx unprivileged base digest | `docker pull nginxinc/nginx-unprivileged:1.27-alpine` → `RepoDigests[0]` → `steps.nginx_base.outputs.image` | pin floating tag |
| **Build separated frontend image** | `docker build -f deploy/separated/Dockerfile.frontend --build-arg NGINX_IMAGE=… --tag new-api-frontend:quality .` | React |
| **Build separated Vue console image** | `docker build -f deploy/separated/Dockerfile.frontend.vue --build-arg NGINX_IMAGE=… --tag new-api-frontend-vue:quality .` | **G4 primary** |
| Validate separated frontend Nginx configurations | `docker run … envsubst` template → `nginx -t` for **both** `new-api-frontend:quality` and `new-api-frontend-vue:quality` | template `/etc/nginx/templates/nginx.conf.template` |
| Record separated image digests | `docker image inspect` Id for four tags + nginx base → `$GITHUB_STEP_SUMMARY` | always() |

G4 local equivalent (when Docker available):

```powershell
docker build -f deploy/separated/Dockerfile.frontend.vue -t new-api-frontend-vue:local .
# need: exit 0; then: docker image inspect --format='{{.Id}}' new-api-frontend-vue:local
```

---

## 4. Operator paste table (CI job URL / digests)

Fill after a green CI run on this branch/PR or after a local docker build. **Do not invent values.**

| Field | Value (paste) |
|-------|----------------|
| GitHub Actions run URL | _empty_ |
| Workflow | `quality.yml` |
| Job | `image-reproducibility` |
| Commit SHA | _empty_ |
| Branch / PR | _empty_ |
| `new-api-frontend-vue:quality` image Id | _empty_ |
| `new-api-frontend:quality` image Id | _empty_ |
| nginx base digests (`steps.nginx_base`) | _empty_ |
| Local (optional) `new-api-frontend-vue:local` Id | _empty_ (host has no docker this session) |
| Recorded by | _empty_ |
| Date | _empty_ |

After paste: update dossier G4 row in [`w3-d7-gate-dossier.md`](./w3-d7-gate-dossier.md) only with real exit 0 + digest; still **not** D7.

---

## 5. Unblock paths

| Path | Owner | Done when |
|------|-------|-----------|
| A · CI | PR/push triggers `quality` → job `image-reproducibility` green | paste run URL + image Ids into §4 |
| B · Local Docker | Operator installs Docker Desktop / CLI on PATH | `docker build -f deploy/separated/Dockerfile.frontend.vue -t new-api-frontend-vue:local .` exit **0** + Id in §4 |
| C · Not unblock | Inventing local green without docker | **forbidden** (module rule) |

G4 green (CI or local) is **necessary but not sufficient** for D7. Still need G2/G3 live, G6 soak, G7 rollback timed, **G8** human phrase `D7 flip 现在`.

---

## 6. Explicit non-claims

- **Did not** run `docker build` (CLI absent).
- **Did not** mark G4 local green.
- **Did not** change production config, flip `FRONTEND_MODE`, or push.
- **Did not** execute D7.

---

## 7. Receipt · M-TH-g4-image-repro

| Item | Value |
|------|--------|
| Module | M-TH-g4-image-repro |
| Outcome | **DONE** · **in-review** |
| G4 local | **blocked** (no docker) |
| G4 SSOT | CI `image-reproducibility` · steps above |
| Evidence doc | `docs/ops/th-g4-image-repro-evidence-2026-07-24.md` |
| D7 | **NOT EXECUTED** |
| Coord | **th-coord** — G4 host blocked; CI remains authority; no flip |

---

*End of G4 evidence pack · 2026-07-24*

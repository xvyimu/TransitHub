# D-G4 · Image 复现 SSOT（本机 Docker 缺席下的权威来源）

**波次：** wave-debt-close-2026-07-27 · wt `th-d7-gate`
**债单：** D-G4-DOCKER（P1，open）
**结论：** 本机 Docker **不在 PATH**，不强跑本机 docker；**CI job `image-reproducibility`（`.github/workflows/quality.yml`）为唯一 SSOT**。

---

## 1. 本机现状（实测）

- `command -v docker` → **DOCKER_NOT_ON_PATH**（本机无 docker CLI）。
- 决策：按任务书红线「不强跑本机 docker」，不安装、不旁路。镜像复现证据完全依赖 CI。

## 2. SSOT = CI job `image-reproducibility`

位置：`.github/workflows/quality.yml`，job 名 `image-reproducibility`。

**依赖门（needs）：** `go-quality` · `web-quality` · `web-console-quality` · `sqlite-migrate` —— 四个上游质量门全绿后才构建镜像，保证镜像基于已验证的源。

**构建矩阵（4 image + 1 base）：**

| 构建产物 | Dockerfile | 说明 |
|----------|-----------|------|
| `new-api:quality` | `Dockerfile` | 一体化集成镜像（embed 前端） |
| `new-api-backend:quality` | `Dockerfile.backend` | 纯后端镜像（`frontend_external`） |
| `new-api-frontend:quality` | `deploy/separated/Dockerfile.frontend` | 分离式 React 前端（nginx） |
| `new-api-frontend-vue:quality` | `deploy/separated/Dockerfile.frontend.vue` | 分离式 Vue console 前端（nginx） |
| nginx base | `nginxinc/nginx-unprivileged:1.27-alpine` | **按 digest pin**，防浮动 tag 漂移 |

**防漂移机制：**

- `Resolve nginx unprivileged base digest`：`docker pull` 后用 `docker inspect --format='{{index .RepoDigests 0}}'` 解析 digest，经 `--build-arg NGINX_IMAGE=<digest>` 注入分离式前端构建，浮动 tag 不能静默变基。
- `Validate separated frontend Nginx configurations`：对两个分离前端镜像用 `envsubst` 渲染 `nginx.conf.template` 后 `nginx -t`，用 loopback upstream 避免依赖活的后端主机名。
- `Record separated image digests`（`if: always()`）：把 4 个镜像的 `.Id` 与 nginx base digest 写入 `$GITHUB_STEP_SUMMARY`，作为可追溯的复现记录。

## 3. 发布镜像的签名/复现（旁证，非本 job）

`docker-build.yml`（tag 触发的多架构发布）具备：`provenance: mode=max` · `sbom: true` · `cosign sign`（单架构 digest + 多架构 manifest）。这是发布面的供应链证据，与 `image-reproducibility` 的 PR/质量面互补。

> 注意：发布面镜像名 `calciumion/new-api` 为受保护身份（AGPL 谱系），不改。

## 4. 验收口径（本波）

- **不判 G4 为「本机实测绿」**：本机无 docker，无法出本机 digest。
- **G4 SSOT = CI**：以最近一次 `quality.yml` 的 `image-reproducibility` job 结论为准；该 job 绿 = 四镜像可构建 + nginx base 已 pin + nginx 配置 `nginx -t` 通过。
- 复现校验方式：读 job 的 `Record separated image digests` summary 比对 `.Id`。

## 5. 剩余 / 后续

- 若需本机复现，需先授权安装 docker（本波不做）。
- D-G4 状态维持 **open → 由 CI 托管**；非 BLOCKED（不依赖 E2E 凭证）。

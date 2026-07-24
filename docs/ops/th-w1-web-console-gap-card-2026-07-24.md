# TH-W1: Vue3 web-console 上线 gap 卡

> **更新：** 2026-07-24 · main `bcb44ad4` · live `d1397d6e` · main ahead 58  
> **范围：** `web-console/`（Vue3+Pinia+i18n+vitest）从代码就绪到本地可构建/可部署的差距。  
> **不覆盖：** 生产 D7 flip（G8 门闩）· 非生产 staging soak · React 前端删除 · 数据库迁移。

## 背景

ARCHITECTURE_TARGET.md 定义了 cutover 八门 G1–G8，已在 `docs/operations/web-console-cutover-plan.md` 逐波追踪。但本地构建管线（`build-release.ps1`、`orca.yaml` setup、Go embed 编译）尚未适配 `web-console/`。本卡列出可闭可修的缺口，按 **构建→部署→验证** 排序。

---

## 缺口清单

### A. `orca.yaml` setup 缺 web-console 依赖

| 现状 | 要求 |
|------|------|
| `scripts.setup: "go mod download && bun install --cwd web"` | 加 `pnpm install --dir web-console --frozen-lockfile` |

创建 orca worktree 后先跑 `pnpm install` 安装 Vue 依赖。`pnpm@11.5.0` 已装。

**修法：** `orca.yaml` 一行追加。

---

### B. `build-release.ps1` 缺 web-console 构建步骤

现状：`build-release.ps1` 只构建 `web/default`（React）和 `web/classic`（React）。

| 缺口 | 说明 |
|------|------|
| B1 安装 | 无 `pnpm install --cwd web-console --frozen-lockfile` |
| B2 typecheck | 无 `pnpm --dir web-console typecheck`（`vue-tsc`，~10s） |
| B3 测试 | 无 `pnpm --dir web-console test`（vitest，~5s） |
| B4 构建 | 无 `pnpm --dir web-console build`（Vite，~30s） |
| B5 产物检查 | `build-release.ps1` 只检查 `web/default/dist/index.html` 和 `web/classic/dist/index.html`，不检查 `web-console/dist/index.html` |
| B6 构建顺序 | web-console 构建应并**行**或跟在 React 构建之后，不阻塞 React 构建 |

**影响：** 当前 `build-release.ps1` 产出不含 Vue console 产物。即便 `web-console/dist/` 被本地构建，`gitignore` 不追踪，Go embed 时找不到。

**修法：** 在 `build-release.ps1` 的 `$SkipWebBuild` 块内补入。注意用 `pnpm` 不是 `bun`。

---

### C. Go embed 不含 `web-console/dist`

`frontend_assets_embedded.go`：

```go
//go:embed web/default/dist
var buildFS embed.FS

//go:embed web/classic/dist
var classicBuildFS embed.FS
```

`web-console/dist` 的 embed 不存在。

| 选项 | 方案 | 工作量 |
|------|------|--------|
| C1 同 binary 嵌入 | 新增 `//go:embed web-console/dist` + `FRONTEND_MODE=vue` 路由选择 | ~30 行 `frontend_assets_embedded.go` + ~20 行 `router/web-router.go` |
| C2 分离容器 | 沿用 `deploy/separated/Dockerfile.frontend.vue` + 现有 Nginx 代理 | 已有 Dockerfile，但 agent host 无 Docker CLI（G4） |
| C3 先选 C2 上线，C1 后续补 | 分离部署先走，embed 做兜底 | 两阶段 |

**推荐：** C1 先做（本地单 binary 部署方便），C2 作为生产分离部署路径。但 **C1 需要 `web-console/dist` 在编译时存在**，也就是 B 必须先修。

---

### D. 路由层缺 Vue console 选择机制

`router/web-router.go` 的 `SetWebRouter` 只处理 `DefaultBuildFS`（React）和 `ClassicBuildFS`（React）。`FRONTEND_MODE` 支持 `auto/embedded/redirect/disabled`，没有 `vue` 或 `web-console` 值。

| 缺口 | 说明 |
|------|------|
| D1 `parseFrontendMode` | 需加 `vue` 枚举值（`router/main.go:60`） |
| D2 `SetWebRouter` | 需接受 `VueBuildFS` + `VueIndexPage`，`FRONTEND_MODE=vue` 时用 Vue 作为 SPA 回退 |
| D3 `ThemeAssets` | 需加 `VueBuildFS embed.FS` + `VueIndexPage []byte` |

**修法：** 与 C1 配套。`FRONTEND_MODE=vue` 时路由 `/` 到 `web-console/dist`，`NoRoute` 回退到 `VueIndexPage`。

---

### E. E2E 凭证缺失（G2 阻塞）

| 缺口 | 说明 |
|------|------|
| E1 非生产环境 | 无 `TH_E2E_USER`/`TH_E2E_PASS`，`scripts/e2e-web-console-login.ps1` exit 10 |
| E2 验证路径 | 首次登录、`/health`、`/channels` 只读页面的 E2E 合约 |

**状态：** 从 W1 到 W4 一直阻塞。`docs/ops/w2-cutover-e2e-credentials.md` 有清单。需要你手工设环境变量。

---

### F. Docker CLI 不可用（G4 阻塞）

| 缺口 | 说明 |
|------|------|
| F1 本地 | agent host 无 `docker` 命令，无法本地构建 `Dockerfile.frontend.vue` |
| F2 CI | CI `image-reproducibility` job 是 SSOT，但本地无法复现 |

**影响：** 分离部署方案（C2）的本地验证不可行。CI 通过的 image 可直接拉取，但本地调试困难。

---

### G. CI 未集成到本地构建

| 缺口 | 说明 |
|------|------|
| G1 `.github/workflows/quality.yml` 有 `web-console-quality` job | 但本地 `build-release.ps1` 不跑同款检查 |
| G2 失败模式 | 本地构建通过但 CI 失败 → 浪费 CI 时间 |

**修法：** B 的修法应尽量对齐 CI job 的内容（`pnpm install --frozen-lockfile` → `typecheck` → `test` → `build`）。

---

## 优先级与依赖

```
优先级      依赖
───────    ────────
P0  A orca.yaml setup    ← 无依赖，分钟级
P0  B build-release.ps1  ← 无依赖，半小时级
P1  C Go embed           ← 依赖 B(web-console/dist 存在)
P1  D 路由选择            ← 依赖 C(embed 存在)
P2  E 凭证                ← 人工操作
P3  F Docker CLI          ← 环境配置
P1  G CI 对齐             ← 与 B 同步修
```

---

## 建议执行顺序

1. **P0 A**：`orca.yaml` 加 `pnpm install --dir web-console --frozen-lockfile` ← 分钟级，当晚可做
2. **P0 B**：`build-release.ps1` 补 web-console 构建步（B1–B5）← 半小时级，可测
3. **P1 C+D**：`frontend_assets_embedded.go` + `router/web-router.go` 加 `web-console/dist` embed + `FRONTEND_MODE=vue` ← 与 B 并行
4. **P1 G**：CI 对齐验证
5. **P2 E**：你设 `TH_E2E_*` → 跑 E2E 验证
6. **P3 F**：安装 Docker CLI 或依赖 CI image

---

## 相关文档

| 路径 | 角色 |
|------|------|
| `docs/operations/web-console-cutover-plan.md` | 生产 cutover 八门 G1–G8 |
| `docs/operations/web-console-cutover-rollback.md` | 回滚操作手册 |
| `docs/ARCHITECTURE_TARGET.md` | 架构目标契约 |
| `docs/legacy-frontend-gate.md` | React 功能冻结 |
| `deploy/separated/README.md` | 分离部署说明 |
| `web-console/README.md` | Vue 开发环境说明 |
| `web-console/E2E.md` | E2E 测试说明 |
| `scripts/e2e-web-console-login.ps1` | E2E 登录脚本 |
| `scripts/build-release.ps1` | 构建发布脚本 |
| `orca.yaml` | Orca worktree setup |
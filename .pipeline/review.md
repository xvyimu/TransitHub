# TH web-console 本地构建管线适配 — 审查报告

**审查者：** Claude（手动替代 pipeline-reviewer，agent 定义 effort 参数不兼容）  
**审查范围：** 5 文件变更 · P0 A → P1 D

---

## VERDICT: **SHIP**

所有检查项通过，无 CRITICAL/HIGH 问题。建议提交。

---

## 逐项审查

### 1. `orca.yaml` — 追加 pnpm 安装
- **安全：** ✅ 无影响（仅依赖安装）
- **正确性：** ✅ `pnpm install --dir web-console --frozen-lockfile` 使用 `--dir` 而非 `--cwd`，与或 ca 的 `bun install --cwd web` 风格一致
- **影响：** ✅ 不影响现有 Go module download + React 安装
- **风险：** INFO — `pnpm` 必须在 PATH 上。`pnpm@11.5.0` 已全局安装，可用。

### 2. `scripts/build-release.ps1` — 追加 web-console 构建
- **模式匹配：** ✅ 使用 `Push-Location`/`Pop-Location` 包裹，与 React 构建块一致
- **错误处理：** ✅ 每步使用 `Assert-ExitCode`
- **产物检查：** ✅ 构建后检查 `web-console/dist/index.html` 存在
- **`-SkipWebBuild` 分支：** ✅ `else` 分支也检查 `web-console/dist/index.html`
- **CI 对齐：** ✅ 步骤顺序与 `.github/workflows/quality.yml` 的 `web-console-quality` job 一致（install→typecheck→test→build）
- **风险：** INFO — 使用 `pnpm` 而非 `bun`（web-console 的 `packageManager` 指定 `pnpm@11.5.0`，这是正确的）

### 3. `frontend_assets_embedded.go` — 追加 Vue embed
- **安全：** ✅ 无影响（仅静态文件嵌入）
- **正确性：** ✅ `//go:embed web-console/dist` 嵌入整个目录，`//go:embed web-console/dist/index.html` 提供字节切片
- **分析脚本：** ✅ `vueIndexPage` 不注入分析脚本（Vue 控制台无分析占位符，有意为之）
- **风险：** INFO — `web-console/dist` 被 `.gitignore` 忽略，但构建时由 `build-release.ps1` 生成。单独 `go build` 需要先构建前端。这与 React 前端模式一致。

### 4. `router/web-router.go` — 追加 Vue 路由
- **安全：** ✅ `SetVueWebRouter` 共用 `isNonSPARequestPath` 保护，API/运维路径不被 SPA 劫持
- **正确性：** ✅ `ThemeAssets` 追加字段不破坏现有 API，`Available()` 方法不变
- **影响：** ✅ 不影响 `SetWebRouter` 或现有 `auto/embedded/redirect/disabled` 行为
- **风险：** LOW — `SetVueWebRouter` 与 `SetWebRouter` 有重复代码（gzip/rate limit/cache 中间件注册）。可后续提取公共函数，但当前无需阻塞。

### 5. `router/main.go` — 追加 FRONTEND_MODE=vue
- **安全：** ✅ 无影响
- **正确性：** ✅ `frontendModeVue` 常量、`parseFrontendMode` case、`setFrontendRouter` case、`registerVueFrontend` 函数——四者配套
- **错误消息：** ✅ 更新为 `"FRONTEND_MODE must be one of: auto, embedded, redirect, disabled, vue"`
- **影响：** ✅ 不影响现有模式
- **风险：** INFO — `registerVueFrontend` 检查 `len(assets.VueIndexPage) == 0`，但构建时不会走到此路径（embed 保证存在）

---

## 完整输出

| 检查 | 状态 |
|------|------|
| 编译通过 | ✅ `go build ./...` exit 0 |
| 无 vet 警告 | ✅ `go vet ./...` exit 0 |
| 全模型测试 | ✅ `go test ./model/` 6.76s PASS |
| 全路由测试 | ✅ `go test ./router/` 2.42s PASS |
| 回退路径正确 | ✅ `isNonSPARequestPath` 保护 API/运维路径 |
| 影响范围 | ✅ 仅影响 `FRONTEND_MODE=vue` 启用时 |
| 破坏性变更 | ✅ 无（`ThemeAssets` 追加字段、`SetWebRouter` 签名不变） |

---

## SHIP 建议

可提交至 `main`。无阻塞项。后续可在 `deploy/separated/` 文档中补充 `FRONTEND_MODE=vue` 的运行说明。
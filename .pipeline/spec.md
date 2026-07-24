# TH: Vue3 web-console 本地构建管线适配

## 目标

让 TransitHub 本地构建管线（`orca.yaml` + `build-release.ps1` + Go embed + 路由）适配 `web-console/`（Vue3），使开发者能通过 `build-release.ps1` 产出含 Vue console 的二进制，且 `FRONTEND_MODE=vue` 时路由到 Vue 前端。

## 变更文件清单

| # | 文件 | 优先级 | 依赖 |
|---|------|--------|------|
| 1 | `orca.yaml` | P0 A | 无 |
| 2 | `scripts/build-release.ps1` | P0 B | 无 |
| 3 | `frontend_assets_embedded.go` | P1 C | 依赖 2（需要 `web-console/dist` 存在） |
| 4 | `router/web-router.go` | P1 D | 依赖 3 |
| 5 | `router/main.go` | P1 D | 依赖 3 |

## 修改内容

### 1. `orca.yaml` — 追加 pnpm 安装

**当前 `scripts.setup`：**
```yaml
scripts:
  setup: |
    go mod download
    bun install --cwd web
```

**改为：**
```yaml
scripts:
  setup: |
    go mod download
    bun install --cwd web
    pnpm install --dir web-console --frozen-lockfile
```

### 2. `scripts/build-release.ps1` — 追加 web-console 构建

在 `$SkipWebBuild` 为 false 时，React 构建完成后（`web/classic` 构建之后），追加 web-console 构建步骤：

```powershell
# --- web-console (Vue3) build ---
Push-Location (Join-Path $repoRoot "web-console")
try {
  & pnpm install --frozen-lockfile
  Assert-ExitCode "web-console pnpm install"
  & pnpm typecheck
  Assert-ExitCode "web-console typecheck"
  & pnpm test
  Assert-ExitCode "web-console test"
  & pnpm build
  Assert-ExitCode "web-console build"
} finally {
  Pop-Location
}
# Verify dist exists
$vueDistIndex = Join-Path $repoRoot "web-console\dist\index.html"
if (-not (Test-Path -LiteralPath $vueDistIndex)) {
  throw "Missing web-console frontend asset: web-console/dist/index.html"
}
```

注意：与 React 构建路径不同的地方：
- web-console 根在工作区根（`web-console/`），不是 `web/default` 或 `web/classic`
- 用 `pnpm` 不是 `bun`
- 使用 `pnpm typecheck` 而不是 `vue-tsc -b`（因为 `package.json scripts.typecheck` 已定义）

### 3. `frontend_assets_embedded.go` — 追加 Vue embed

在 `frontend_assets_embedded.go` 中，在现有 `//go:embed` 指令之后追加：

```go
//go:embed web-console/dist
var vueBuildFS embed.FS

//go:embed web-console/dist/index.html
var vueIndexPage []byte
```

在 `prepareFrontendAssets()` 函数中，追加 Vue 资源到返回的 `router.ThemeAssets`：

```go
return router.ThemeAssets{
    DefaultBuildFS:   buildFS,
    DefaultIndexPage: indexPage,
    ClassicBuildFS:   classicBuildFS,
    ClassicIndexPage: classicIndexPage,
    VueBuildFS:       vueBuildFS,    // 追加
    VueIndexPage:     vueIndexPage,  // 追加
}
```

### 4. `router/web-router.go` — 追加 Vue 路由

**ThemeAssets struct 追加字段：**
```go
type ThemeAssets struct {
    DefaultBuildFS   embed.FS
    DefaultIndexPage []byte
    ClassicBuildFS   embed.FS
    ClassicIndexPage []byte
    VueBuildFS       embed.FS        // 追加
    VueIndexPage     []byte          // 追加
}
```

**`Available()` 方法**：追加 `len(assets.VueIndexPage) > 0` 检查（可选，为兼容不传 Vue 的调用方）。

**`SetWebRouter()` 函数**：在函数开头，`FRONTEND_MODE=vue` 时：

```go
func SetWebRouter(router *gin.Engine, assets ThemeAssets, frontendMode string) {
    if frontendMode == "vue" {
        vueFS := common.EmbedFolder(assets.VueBuildFS, "web-console/dist")
        router.Use(gzip.Gzip(gzip.DefaultCompression))
        router.Use(middleware.GlobalWebRateLimit())
        router.Use(middleware.Cache())
        router.Use(static.Serve("/", vueFS))
        router.NoRoute(func(c *gin.Context) {
            c.Set(middleware.RouteTagKey, "web")
            if isNonSPARequestPath(c.Request.RequestURI) {
                controller.RelayNotFound(c)
                return
            }
            c.Header("Cache-Control", "no-cache")
            c.Data(http.StatusOK, "text/html; charset=utf-8", assets.VueIndexPage)
        })
        return
    }
    // 原有 default/classic 逻辑不变
    ...
}
```

注意：需要修改 `SetWebRouter` 签名，增加 `frontendMode string` 参数。所有调用方也需要更新。

### 5. `router/main.go` — 追加 FRONTEND_MODE=vue 枚举

**`parseFrontendMode()` 函数**（约 60 行）：在错误消息中追加 `"vue"` 枚举值。

**`SetAPIRouter()` 函数**（约 96 行）：`FRONTEND_MODE=vue` 时，调用 `SetWebRouter` 传入 `"vue"`。

## 依赖关系

```
orca.yaml      ─ 无依赖，最先改
build-release.ps1  ─ 无依赖，可并行改
                    │
                    ▼
frontend_assets_embedded.go  ─ 依赖 build-release.ps1 之后 web-console/dist 可编译
                    │
                    ▼
router/web-router.go  ─ 依赖 embed 添加
router/main.go        ─ 依赖 web-router.go 的签名变更
```

## 验证方法

### 1. orca.yaml
```bash
cd D:\TransitHub\src
pnpm install --dir web-console --frozen-lockfile
# 应退出 0，无错误
```

### 2. build-release.ps1
```powershell
cd D:\TransitHub\src
.\scripts\build-release.ps1 -SkipTests
# 应产出含 web-console/dist 的二进制，exit 0
```

### 3. Go embed + 路由
```bash
cd D:\TransitHub\src
go build -trimpath -o new-api-test.exe .
# 应编译通过
FRONTEND_MODE=vue ./new-api-test.exe
# 应启动成功，访问 / 应返回 Vue console 页面
```

### 4. model 全绿
```bash
go test ./model/ -count=1 -timeout 120s
# PASS
```

## 风险/注意事项

1. `build-release.ps1` 的 `SkipWebBuild` 开关：如果传 `-SkipWebBuild`，不会构建任何前端（包括 web-console），此时 Go embed 编译会失败（缺少 `web-console/dist`）。这是预期行为——`SkipWebBuild` 用于已有二进制场景。
2. `FRONTEND_MODE=vue` 是新增的模式，不影响现有 `auto/embedded/redirect/disabled` 行为。
3. `SetWebRouter` 签名变更会影响所有调用方，需检查 `router/main.go` 全部调用点。
4. `web-console/dist` 不提交到 git（`.gitignore` 已忽略），构建时生成。
5. `web-console` 的 `pnpm install` 需要网络，`orca.yaml` 的 setup 会在 worktree 创建时自动执行。
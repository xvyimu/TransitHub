# TH web-console 本地构建管线适配 — 变更摘要

## 变更文件

| # | 文件 | 变更类型 | 说明 |
|---|------|----------|------|
| 1 | `orca.yaml` | 追加 | 在 `scripts.setup` 中追加 `pnpm install --dir web-console --frozen-lockfile` |
| 2 | `scripts/build-release.ps1` | 追加 | 在 `$SkipWebBuild` 块内，React 构建后追加 web-console 构建（pnpm install→typecheck→test→build）+ 产物检查；`-SkipWebBuild` else 分支也追加 `web-console/dist/index.html` 检查 |
| 3 | `frontend_assets_embedded.go` | 追加 | 追加 `vueBuildFS embed.FS` + `vueIndexPage []byte` embed 指令；`prepareFrontendAssets()` 返回值追加 `VueBuildFS`/`VueIndexPage` |
| 4 | `router/web-router.go` | 追加 | `ThemeAssets` struct 追加 `VueBuildFS`/`VueIndexPage` 字段；追加 `SetVueWebRouter()` 函数（Vue SPA 路由注册） |
| 5 | `router/main.go` | 追加 | 追加 `frontendModeVue` 常量；`parseFrontendMode` 加 `vue` case + 更新错误消息；`setFrontendRouter` 加 `frontendModeVue` case；追加 `registerVueFrontend()` 函数 |

## 验证结果

- `go build ./...` ✅ 无错误
- `go vet ./...` ✅ 无错误
- `go test ./model/` ✅ PASS (6.757s)

## 使用方式

```bash
# 构建含 Vue console 的二进制
cd D:\TransitHub\src
.\scripts\build-release.ps1 -SkipTests

# 运行 Vue console
FRONTEND_MODE=vue ./new-api-*.exe
```

## 未涉及

- P2 E（E2E 凭证，人工操作）
- P3 F（Docker CLI，环境配置）
- 生产 D7 flip
- React 前端删除
- `.github/workflows/quality.yml` CI 修改
@AGENTS.md

## 快速入口
- 栈：**Go** + **Gin** + **GORM** · 3 库兼容 · 开发 UI = **Vue3+Naive**（web-console）· 生产默认 **React**（web/default）至 D7
- 测试：`go test ./...` · `pwsh scripts/check-architecture-guards.ps1` · `pnpm test`（web-console）
- 红线：不 D7 flip · 不 main 直推 · 不出 exe 未经授权 · 不动 `:3000` 官方服务
- 架构：JSON 只经 `common.Marshal/Unmarshal`；`encoding/json` 直调禁新增（守卫脚本）
- 先读：`docs/PROJECT.md` · `docs/adr/` · `.planning/` 当前计划
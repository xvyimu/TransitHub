# TransitHub

**GitHub：** [xvyimu/TransitHub](https://github.com/xvyimu/TransitHub)  
**Go module：** `github.com/xvyimu/TransitHub`  
**产品显示名（本机）：** NewAPI · 安装/进程名可能仍为 `new-api` / `new-api-fixed`  
**本地路径：** `D:\TransitHub\src`（junction `D:\newapi`）  
**许可：** [AGPL-3.0](./LICENSE) · 附加说明见 [NOTICE](./NOTICE)

> 本仓是 **xvyimu 独立维护** 的 TransitHub。  
> GitHub **已脱离** upstream fork 网络；本地 **仅** `origin` → 本仓。  
> 代码谱系：QuantumNous / Calcium-Ion `new-api` · 祖先 `songquanpeng/one-api`（**AGPL 义务与 NOTICE 署名必须保留**）。  
> 产品能力仍是 New API 系 LLM 网关；仓库与模块身份是 **TransitHub**。  
> 身份卡：[GITHUB_IDENTITY.md](./GITHUB_IDENTITY.md)

## 它是什么

**下一代 LLM 网关 + AI 资产管理系统**（聚合多上游、令牌/配额、渠道运营、管理后台）。

**形态与栈 SSOT：** [`docs/PROJECT.md`](./docs/PROJECT.md) · Agent：[`AGENTS.md`](./AGENTS.md)。

本机生产形态（运维剖面）：

- **LOCAL-ONLY** · `127.0.0.1:3000` · SQLite WAL  
- 自适应选路 **shadow**、退款 outbox、Ops 健康条、价源 probe（默认仅 diff）  
- 运行剖面：`../agent_docs/CURRENT_STATE.md`（docs 可能在 git 外）

## 开发体系（本仓）

| 项 | 约定 |
|----|------|
| remote | 仅 `origin` = `xvyimu/TransitHub` |
| 模块路径 | `github.com/xvyimu/TransitHub` |
| 发版 | `scripts/build-release.ps1` → `../scripts/deploy-production-release.ps1` |
| OTEL | 默认 `OTEL_TRACES_ENABLED=false` · `OTEL_LOGS_ENABLED=false` |
| 上游 cherry-pick | **非默认流程**；需要时人工评估，不自动跟踪 QuantumNous/Calcium-Ion |

```powershell
cd D:\TransitHub\src
git remote -v
git push origin HEAD
```

## 快速健康检查（本机）

```powershell
curl.exe -sS -D - http://127.0.0.1:3000/api/status -o NUL
# 期望 X-New-Api-Version 与 agent_docs/CURRENT_STATE 一致
```

## OTEL（本仓增量）

默认 traces/logs **双关**。Collector/Loki 配置：`deploy/otel/` · 启动：`scripts/start-otel-stack.ps1`。

## 许可与归属

- **AGPL-3.0** — [LICENSE](./LICENSE)  
- **NOTICE** — 上游 QuantumNous 署名、§7 UI 归属、原项目链接必须保留  
- 再分发/网络服务请遵守 AGPL（含源码提供义务）

上游参考（谱系，非本仓 remote）：  
https://github.com/QuantumNous/new-api · 文档 https://docs.newapi.pro

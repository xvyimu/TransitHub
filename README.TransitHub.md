# TransitHub

**GitHub：** [xvyimu/TransitHub](https://github.com/xvyimu/TransitHub)  
**产品显示名（本机）：** NewAPI · 安装/进程名可能仍为 `new-api` / `new-api-fixed`  
**本地路径：** `D:\newapi`（git 根 `D:\newapi\src`）  
**许可：** [AGPL-3.0](./LICENSE)（继承 new-api / one-api 谱系）  

> 本仓由 `xvyimu/new-api` **改名**为 **TransitHub**，并 **脱离 upstream fork 网络**。  
> 上游参考：`QuantumNous/new-api` · `Calcium-Ion/new-api` · 谱系祖先 `songquanpeng/one-api`。  
> **产品能力与协议兼容**仍是 New API 系 LLM 网关；GitHub 仓库身份是 TransitHub。

## 它是什么

**下一代 LLM 网关 + AI 资产管理系统**（聚合多上游、令牌/配额、渠道运营、管理后台）。

本机生产形态（本 fork 运维剖面）：

- **LOCAL-ONLY** · `127.0.0.1:3000` · SQLite WAL  
- 自适应选路 **shadow**、退款 outbox、Ops 健康条、价源 probe（默认仅 diff）  
- 运行剖面：`../agent_docs/CURRENT_STATE.md`（docs 可能在 git 外）

## 上游文档

完整多语言说明、部署与功能列表仍以历史 README 结构为准（本文件为 **仓库身份 + 本机运维入口**）。  
上游项目页：https://github.com/QuantumNous/new-api · 文档：https://docs.newapi.pro

## 本仓目录（摘要）

```text
main.go / router / controller / service / model / relay /
web/default · web/classic
pkg/observability          # metrics + OTEL traces（opt-in）
scripts/                   # build-release / deploy（本机）
```

本地 remote 名可能仍为 `fork`，指向 `xvyimu/TransitHub`：

```powershell
cd D:\newapi\src
git remote -v
git push fork HEAD
```

## 快速健康检查（本机）

```powershell
curl.exe -sS -D - http://127.0.0.1:3000/api/status -o NUL
# 期望 X-New-Api-Version 与 agent_docs/CURRENT_STATE 一致
```

## OTEL（本仓增量）

默认 **`OTEL_TRACES_ENABLED=false`**。设计见仓外 `D:\newapi\docs\adr-otel-2026-07-21.md`（若存在）。

## 许可

AGPL-3.0 — 见 [LICENSE](./LICENSE)。衍生自 new-api；再分发请遵守 AGPL。

# TransitHub · 产品分层方案（PRODUCT-LAYERS）

> **组合总纲：** `D:\orca\.planning\portfolio-product-docs-program-2026-07-23\PORTFOLIO-PRODUCT-PROGRAM.md`  
> **形态与栈 SSOT：** [`PROJECT.md`](./PROJECT.md)（L1 不重复）  
> **tip 波次：** `a72c558c` · 更新文档时以 `git rev-parse` 为准

---

## L0 · 产品身份

| 项 | 内容 |
|----|------|
| **一句话** | 自托管的 LLM API **网关 + 运营控制台**：聚合上游、令牌配额、渠道中继与后台运营。 |
| **核心问题** | 如何在自有基础设施上稳定、可运营地暴露多模型 API，而不是再造一个聊天前端。 |
| **主用户** | **运维/平台开发者**（部署与加固）· **站点运营**（渠道/用户/令牌） |
| **次用户** | 通过 API 调用的业务系统 |
| **明确不做** | 纯前端 SaaS 主形态 · 小程序/移动 App 主线 · 无 ADR 的双前端同屏乱写 · 公网裸奔无鉴权 |
| **价值** | 可二次分发（**AGPL** 义务）· Docker/本机部署 · 管理台可 strangler 演进 |

---

## L1 · 形态与栈

见 [`PROJECT.md`](./PROJECT.md) §1–§2：Go/Gin/GORM · 三库 · 管理台过渡策略。

---

## L2 · 运行与边界

| 项 | 内容 |
|----|------|
| Git 根 | `D:\TransitHub\src` |
| 运行根 | `D:\TransitHub`（data/env/exe **仓外**） |
| 布局 | 根 [`LAYOUT.md`](../../LAYOUT.md)（相对 git 根上一级） |
| 入口 | `D:\projects\TransitHub` |
| 门闩 | **D7 生产 flip** 人 gate（无口令不执行）；TH_E2E 凭据齐才能 full verify |
| 密钥 | `.env` / data **不进 git** |

---

## L3 · 架构与扩展

| 层 | 职责 |
|----|------|
| Router → Controller → Service → Model | 请求链路 |
| `web/default` | 主管理台 UI（视觉 A0/A1 已对齐 Atelier 结构） |
| `web-console` / classic | 过渡/遗留 · 见 PROJECT 前端闸门 |
| 扩展点 | 渠道适配 · 中继策略（flag 默认安全）· 观测 OTEL 可选 |
| **禁止扩展** | 第二后端语言平行网关 · 无 ADR 双写 React+Vue 同屏 · 计费与控制台混成无边界单进程乱耦 |

依赖方向：业务 **禁止** 直接 `encoding/json`（走 `common/json.go`）。

---

## L4 · 验收与质量

| 场景 | 命令/证据 |
|------|-----------|
| 后端改动 | `go test` 触及包 / `./...` 按任务 |
| 管理台改动 | `bun`/`pnpm` 以 `web/default` 脚本：typecheck · build |
| 安全改动 | 相关 middleware 单测 · 不无测合入 |
| 发布 | 不把 `data/`、密钥、live exe 打进源码仓 |
| 覆盖率 | **触达模块测绿**；不虚报全局 % |

兼容：SQLite / MySQL / PostgreSQL 声明见 PROJECT；新 schema 走 migrations。

---

## L5 · 协作与合规

| 项 | 内容 |
|----|------|
| 许可 | **AGPL-3.0** · 见根 `LICENSE` · `NOTICE` 署名 |
| 安全 | [`../SECURITY.md`](../SECURITY.md)（相对 docs）或根 SECURITY |
| 贡献 | [`../CONTRIBUTING.md`](../CONTRIBUTING.md) · 欢迎 Issue/PR · 私有维护者可直合 |
| 谱系 | new-api / one-api 衍生 · 独立 `origin` |

---

## L6 · 路线图与维护节奏

| 周期 | 内容 |
|------|------|
| 近 | 生产 D7 人 gate 议程（独立会话）· b5/b6 残值评估 |
| 中 | 管理台 strangler 按 cutover 计划 · 依赖 minor |
| 远 | 观测/稳定性打磨 · 文档与 NOTICE 同步 |
| 文档 | 改形态/栈 → 先 PROJECT；本文件 L0/L4 随发版检 |
| 安全 | 跟进 CVE/依赖 · SECURITY 渠道响应 |
| 性能 | 渠道与中继路径 profiling；禁无基准的「全面重写」 |

---

## 文档地图

| 文档 | 用途 |
|------|------|
| PROJECT.md | 形态栈 SSOT |
| PRODUCT-LAYERS.md | 本方案 |
| ARCHITECTURE_TARGET.md | 目标架构 |
| LAYOUT.md（运行根） | 路径与清理配额 |
| ops / agent_docs | 运行态 |

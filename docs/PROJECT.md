# TransitHub · 形态与技术栈（SSOT）

> **产品（本机显示名）：** NewAPI · **GitHub / module：** [xvyimu/TransitHub](https://github.com/xvyimu/TransitHub) · `github.com/xvyimu/TransitHub`  
> **Git 根 / 开发目录：** `D:\TransitHub\src`（入口 `D:\projects\TransitHub`）  
> **本机运维剖面：** LOCAL-ONLY · `127.0.0.1:3000` · SQLite WAL（见运维 CURRENT_STATE）  
> 全局门闩：`~/CLAUDE.md` §8 · `~/.claude/specs/principle.md`「形态与技术栈」。  
> **本文件 = 本产品形态与唯一技术栈权威。** 谱系来自 new-api / one-api（**AGPL**）；独立仓仅 `origin`。小修不重选型。  
> **前端过渡：** 与 [`ARCHITECTURE_TARGET.md`](./ARCHITECTURE_TARGET.md) · [`legacy-frontend-gate.md`](./legacy-frontend-gate.md) · [`operations/web-console-cutover-plan.md`](./operations/web-console-cutover-plan.md) **一致**；本文写清「开发主路径 vs 生产默认 vs 禁平行」。

---

## 1. 产品形态（唯一）

| 项 | 结论 |
|----|------|
| **形态** | **自托管 LLM 网关 + 管理控制台**（后端服务 + Web Admin） |
| **交付** | 本机/服务器进程（Go 二进制）+ 嵌入或分开发货前端；Docker / `deploy/separated/` 可选 |
| **能力面** | 多上游聚合、令牌/配额、渠道、中继、管理后台、本机 shadow/退款 outbox 等增量 |
| **不是** | 纯前端 SaaS、小程序、移动 APP、桌面 Electron **作为主产品形态**（`electron/` 可为壳，不另起第二产品线） |

**做 / 不做（形态级）**

| 做 | 不做 |
|----|------|
| 统一 API 网关 + 运营后台 | 另起第二套后端语言平行网关 |
| SQLite / MySQL / PostgreSQL **三库兼容** | 只为某一库写死 SQL 方言 |
| 管理台 **strangler**：新能力进 Vue `web-console/` | **无 ADR + 无 D7 人 gate** 双写 React+Vue 同一屏、或把 classic 升为唯一主路径 |

---

## 2. 唯一技术栈

### 2.1 后端与数据（稳定）

| 层 | 技术 | 约束 |
|----|------|------|
| 后端 | **Go**（`go.mod` 为准）· **Gin** · **GORM v2** | 分层：Router → Controller → Service → Model |
| JSON | **仅** `common/json.go` 封装 | 业务代码禁止直接 `encoding/json` marshal/unmarshal |
| DB | **SQLite + MySQL ≥5.7.8 + PostgreSQL ≥9.6** 同支持 | 新 schema 走 `migrations/`；见 TARGET §4 |
| 缓存 | Redis（go-redis）+ 内存缓存 | |
| 鉴权 | JWT · WebAuthn · 多种 OAuth | |
| 观测（本仓增量） | OTEL 可选；默认 traces/logs **关** | |
| 许可 | **AGPL-3.0** + NOTICE 署名义务 | 不可抹谱系 |

### 2.2 管理控制台（过渡态 · 必读）

> **一句话：** **开发主路径 = Vue `web-console/`**；**生产默认交付仍 = React `web/default/`**，直到 cutover gate 齐 + **D7 人 gate**。不是「两套都主路径」。

| 角色 | 路径 | 栈 | 规则 |
|------|------|-----|------|
| **开发主路径**（新管理 UI） | `web-console/` | **Vue 3** · **TypeScript** · **Vite** · **Naive UI** · **pnpm** | 新控制台能力默认落这里；CI：`web-console-quality` |
| **生产默认 / 回滚面**（切流前） | `web/default/` | **React 19** · **TypeScript** · **Rsbuild** · Base UI · Tailwind · **Bun** | **LEGACY-HOTFIX only**：安全、严重回归、embed/回滚构建、现网 typo；**禁止**新功能与 Vue 双写同屏 |
| **L2 冻结** | `web/classic/` | React 18 系 · Vite · Semi（历史） | 无新屏、无功能追平；仅安全/严重回归 |
| **交付缝** | `FRONTEND_MODE` · `deploy/separated/` · ADR-0001 | Go embed / 外部前端镜像 | **未授权不得**改生产 `FRONTEND_MODE` 或默认 Vue 切流（**D7 = 人 gate**） |

| 阶段 | 默认用户看到的 Admin | Agent 写新 UI 写哪里 |
|------|----------------------|----------------------|
| **现在（Phase1 · 未 D7）** | React `web/default`（回滚资产） | **`web-console/`** |
| **D7 之后（人确认 flip）** | Vue `web-console` 为生产默认 | 仍 `web-console/`；React 仅回滚/收口 |

**禁止（防漂移）：**

- 同一管理能力在 React 与 Vue **长期双实现**（见 `legacy-frontend-gate.md` Resolution B）  
- 无 ADR 引入 Nest / FastAPI 第二后端，或用 Next 重写整个 Admin  
- 把 classic 升为「第三主路径」  
- 把 D7 写成「文档里默认真 flip」——**TARGET 与 cutover-plan 仍不授权生产 cutover**

Agent 约定：根 [`AGENTS.md`](../AGENTS.md) · [`CLAUDE.md`](../CLAUDE.md) · 产品说明 [`README.TransitHub.md`](../README.TransitHub.md) · 架构 [`ARCHITECTURE_TARGET.md`](./ARCHITECTURE_TARGET.md) · [`ARCHITECTURE_ASIS.md`](./ARCHITECTURE_ASIS.md)。

---

## 3. 选型理由（取舍）

- **网关形态：** 必须服务端长连接/中继/计费，Web-only 或小程序不够。  
- **Go + Gin：** 谱系与性能；保持与 AGPL 上游可对照的结构，降低独立维护成本。  
- **Vue `web-console` 为新 UI 主路径：** strangler 增量替换管理台；Naive/Vite/pnpm 与 Phase1 只读台已落地；CI 已钉质量门。  
- **React `web/default` 保留至 D7：** 生产默认与 ≤5 min 回滚面；避免未过 gate 的流量切换。  
- **三库兼容：** 部署面宽；禁止「只在本机 SQLite 能跑」的迁移。  
- **唯一后端栈：** 不平行引入 Nest/FastAPI；不无 ADR 用 Next 重写整个 Admin。

---

## 4. 防漂移

1. 改业务前读本文 + `AGENTS.md`（JSON 封装、DB 兼容、分层）+ **TARGET 前端边界**。  
2. **新管理 UI → `web-console/`**；`web/default` 仅 LEGACY-HOTFIX；`web/classic` L2 冻结。  
3. **生产切 Vue / 改 `FRONTEND_MODE`：** 仅当 cutover G1–G8 证据齐 **且** 用户明确 D7 授权（见 cutover-plan / rollback）。  
4. 换栈/换形态 → ADR（`docs/`）+ **更新本文** → 确认后实现。  
5. 不擅自恢复 `upstream` remote 或自动化跟踪 QuantumNous（非默认流程）。

---

## 5. 相关文档（前端过渡）

| 文档 | 角色 |
|------|------|
| [`ARCHITECTURE_TARGET.md`](./ARCHITECTURE_TARGET.md) | Phase1 目标契约 · cutover gates · **不授权**生产 flip |
| [`ARCHITECTURE_ASIS.md`](./ARCHITECTURE_ASIS.md) | 现状测绘（含三前端事实） |
| [`legacy-frontend-gate.md`](./legacy-frontend-gate.md) | 新 UI 只进 web-console · 禁双写 |
| [`operations/web-console-cutover-plan.md`](./operations/web-console-cutover-plan.md) | G1–G8 · D7 人 gate |
| [`operations/web-console-cutover-rollback.md`](./operations/web-console-cutover-rollback.md) | ≤5 min 回滚 |
| [`adr/0001-frontend-backend-delivery-seam.md`](./adr/0001-frontend-backend-delivery-seam.md) | 交付缝 |

**修订：** 2026-07-23 · 对齐 Vue strangler 过渡态与 TARGET/legacy-frontend-gate（修正「仅 React 主路径」过时表述）。

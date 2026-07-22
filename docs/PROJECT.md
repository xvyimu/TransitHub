# TransitHub · 形态与技术栈（SSOT）

> **产品（本机显示名）：** NewAPI · **GitHub / module：** [xvyimu/TransitHub](https://github.com/xvyimu/TransitHub) · `github.com/xvyimu/TransitHub`  
> **Git 根 / 开发目录：** `D:\TransitHub\src`（入口 `D:\projects\TransitHub`）  
> **本机运维剖面：** LOCAL-ONLY · `127.0.0.1:3000` · SQLite WAL（见运维 CURRENT_STATE）  
> 全局门闩：`~/CLAUDE.md` §8 · `~/.claude/specs/principle.md`「形态与技术栈」。  
> **本文件 = 本产品形态与唯一技术栈权威。** 谱系来自 new-api / one-api（**AGPL**）；独立仓仅 `origin`。小修不重选型。

---

## 1. 产品形态（唯一）

| 项 | 结论 |
|----|------|
| **形态** | **自托管 LLM 网关 + 管理控制台**（后端服务 + Web Admin） |
| **交付** | 本机/服务器进程（Go 二进制）+ 嵌入或同仓前端构建物；Docker 可选 |
| **能力面** | 多上游聚合、令牌/配额、渠道、中继、管理后台、本机 shadow/退款 outbox 等增量 |
| **不是** | 纯前端 SaaS、小程序、移动 APP、桌面 Electron 壳（本仓主形态） |

**做 / 不做（形态级）**

| 做 | 不做 |
|----|------|
| 统一 API 网关 + 运营后台 | 另起第二套后端语言平行网关 |
| SQLite / MySQL / PostgreSQL **三库兼容** | 只为某一库写死 SQL 方言 |
| 默认前端主题 `web/default`（React 19） | 无 ADR 下把 classic 升为唯一主路径 |

---

## 2. 唯一技术栈

| 层 | 技术 | 约束 |
|----|------|------|
| 后端 | **Go**（`go.mod` 为准）· **Gin** · **GORM v2** | 分层：Router → Controller → Service → Model |
| JSON | **仅** `common/json.go` 封装 | 业务代码禁止直接 `encoding/json` marshal/unmarshal |
| 前端主路径 | **React 19** · **TypeScript** · **Rsbuild** · Base UI · Tailwind | 包管理：**Bun**（`web/default`） |
| 前端 legacy | `web/classic`（React 18 / Vite / Semi） | 维护级；新功能默认 default 主题 |
| DB | **SQLite + MySQL ≥5.7.8 + PostgreSQL ≥9.6** 同支持 | |
| 缓存 | Redis（go-redis）+ 内存缓存 | |
| 鉴权 | JWT · WebAuthn · 多种 OAuth | |
| 观测（本仓增量） | OTEL 可选；默认 traces/logs **关** | |
| 许可 | **AGPL-3.0** + NOTICE 署名义务 | 不可抹谱系 |

Agent 约定：根 [`AGENTS.md`](../AGENTS.md) · [`CLAUDE.md`](../CLAUDE.md) · 产品说明 [`README.TransitHub.md`](../README.TransitHub.md)。

---

## 3. 选型理由（取舍）

- **网关形态：** 必须服务端长连接/中继/计费，Web-only 或小程序不够。
- **Go + Gin：** 谱系与性能；保持与 AGPL 上游可对照的结构，降低独立维护成本。
- **React 管理台：** 运营 UI 成熟路径；Bun/Rsbuild 为仓内已选工具链。
- **三库兼容：** 部署面宽；禁止「只在本机 SQLite 能跑」的迁移。
- **唯一栈：** 不平行引入 Nest/FastAPI 第二后端；不无 ADR 用 Next 重写整个 Admin。

---

## 4. 防漂移

1. 改业务前读本文 + `AGENTS.md` 规则（JSON 封装、DB 兼容、分层）。  
2. 新功能默认 `web/default`；动 classic 须说明。  
3. 换栈/换形态 → ADR（`docs/`）+ 更新本文 → 确认后实现。  
4. 不擅自恢复 `upstream` remote 或自动化跟踪 QuantumNous（非默认流程）。

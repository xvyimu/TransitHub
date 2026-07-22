# TransitHub · Architecture As-Is

| 字段 | 值 |
|------|-----|
| 测绘日 | 2026-07-22 |
| 真路径 | `D:\TransitHub\src` |
| 模块 | `github.com/xvyimu/TransitHub` |
| 分支快照 | `main` @ `57d0891c`（Phase1 基线；实现 worktree 可另有 feature branch） |
| 上游血缘 | QuantumNous **new-api** fork / 换皮 |
| 目标契约 | [`ARCHITECTURE_TARGET.md`](ARCHITECTURE_TARGET.md)（仓内 Phase1 边界与验证 gate） |
| 产品标签（建议） | **P0 旗舰**（Go 加深 + Vue 绞杀 gate） |
| 性质 | **只读测绘**；本文不授权业务改码 |
| 可见性 | 仓内 `docs/`；工作树按相同相对路径呈现 |

---

## 1. 一句话现状

**单体 Go AI API 网关**：Gin 路由 + GORM 多方言 SQL + Redis 可选缓存 + 40+ 上游 relay 适配 + 双 React 主题嵌入/分开发货 + Electron 壳，并有一个独立的 **Vue 3 `web-console` Phase1 控制台**。
**无 Python、无 C/固件**。Vue 控制台尚未成为默认交付物；Go 网关职责与 SQL/Shell/Git 目标对齐，AI 领域逻辑仍绑定在 Go 热路径。

---

## 2. 语言与体量（近似，排除 node_modules/dist）

| 语言/形态 | 文件数 | 行数（约） | SSOT 关系 |
|-----------|--------|------------|-----------|
| Go | 757 | ~129k | **目标内** — 网关/计费/渠道/鉴权 |
| TSX | 672 | ~132k | **偏离** — React 面板主体 |
| TS | 340 | ~43k | **偏离**（前端）/ 可复用类型 |
| JSX | 340 | ~93k | **偏离** — classic 主题 |
| JS | 67 | ~8k | Electron / 脚本 / 杂项 |
| MD | 49 | ~8k | 文档 |
| CSS | 5 | ~3k | Tailwind 时代残余 |
| SQL 文件 | 0 主仓；`bin/migration_v*.sql` 遗留 2 份 | — | **规范缺口** — 现靠 GORM AutoMigrate |
| Vue / Python / C | `web-console/` / **0** / **0** | — | Vue 控制台为 Phase1 只读绞杀；Python/C 未开 |

**前端栈事实**

| 主题 | 路径 | 栈 |
|------|------|-----|
| default（主） | `web/default/` | React 19 · TypeScript · Rsbuild · Base UI · Tailwind · TanStack Router/Query/Table · i18next · Bun workspace |
| classic | `web/classic/` | React · Vite 系 · Semi Design（历史） |
| Electron | `electron/` | Electron 39 · 壳加载本地/远程控制台 |
| `web-console`（Phase1） | `web-console/` | Vue 3 · TypeScript · Naive UI · Vite · pnpm；未切流 |

---

## 3. 目录 / 模块图

```text
TransitHub (monorepo)
├── main.go                 # run_mode / plane / worker / scheduler / HTTP
├── router/                 # Gin：API · relay · video · dashboard · web
├── middleware/             # auth · rate · CORS · distribute · security headers …
├── controller/             # ~90 handlers（管理面 + 计费 webhook + health）
├── service/                # 计费会话 · 渠道选择/熔断/自适应 · refund · task …
├── model/                  # GORM 模型 + AutoMigrate + channel cache/merge
├── relay/                  # 协议 handler + channel/* 上游适配器（38+）
├── dto/ · types/ · constant/
├── common/                 # JSON 封装 · Redis · 配额数学 · env · 限流
├── setting/                # 比例/模型/运营/性能配置
├── oauth/ · i18n/ · logger/
├── pkg/                    # billingexpr · cachex · ionet · observability · perf_metrics
├── web/{default,classic}/  # React 双主题
├── web-console/             # Vue3 + Naive UI Phase1 控制台
├── electron/               # 桌面壳
├── deploy/{separated,otel,prometheus}
├── docs/{openapi,adr,operations}
├── scripts/ · bin/ · .github/
└── frontend_assets_{embedded,external}.go   # 交付缝（ADR-0001）
```

### 3.1 逻辑分层（已实现）

```text
Client / SPA / SDK
        │  HTTPS JSON · SSE · WS
        ▼
┌───────────────────────────────────────────┐
│  Go process (planes: all | management | relay)
│  router → middleware → controller → service → model
│                         ↘ relay/channel → upstream AI
│  workers/schedulers: batch updater, system tasks, refund outbox …
└───────────────────────────────────────────┘
        │
        ├── SQL (SQLite | MySQL | PostgreSQL)  main + optional LOG_SQL_DSN
        ├── Redis (optional)
        ├── ClickHouse (log DSN only)
        └── 对象/外呼：OAuth · 支付 webhook · 上游厂商
```

**与目标分层契约对照**

| 目标层 | As-Is | 判定 |
|--------|-------|------|
| Vue3+NaiveUI 面板 | `web-console/` Phase1 只读页；React 仍为默认 | **P0 gate 进行中** |
| Go 网关 | 已是核心，边界大但职责正确 | **保留强化** |
| Python AI-Core | 无；自适应选路/计费表达式在 Go | **缺口 P1** |
| SQL | `migrations/` 已建立；baseline 目前只验证 SQLite | **部分** |
| 设备 C/STM32 | 无 | **SIDE 未开** |

---

## 4. 运行面与交付缝（绞杀切点已存在）

### 4.1 运行模式

- `RUN_MODE` / `APP_PLANE`：`all` | `management` | `relay` 等（`router.Plane*`）
- HTTP 探活：`/healthz` `/livez` `/readyz`
- 可选 OTEL（`pkg/observability`，fail-open）

### 4.2 前端交付（ADR-0001）— **首选绞杀切点**

| 机制 | 作用 |
|------|------|
| `FRONTEND_MODE=auto\|embedded\|redirect\|disabled` | 嵌入 SPA / 跳转独立前端 / 纯 API |
| `-tags frontend_external` | 二进制不含 embed 资源 |
| `deploy/separated/` | Nginx SPA + 反代同域 `/api` `/v1` … |
| `FRONTEND_BASE_URL` | redirect 目标源 |

**含义**：新 `web-console`（Vue）可与旧 React **同仓并行、同域切换**，无需先拆 Git 仓。

---

## 5. 对外 API 面（边界清单）

OpenAPI 资产：`docs/openapi/api.json`（~166KB）、`docs/openapi/relay.json`（~172KB）— 存在但需与路由表做一次对齐审计（T1）。

### 5.1 管理 / 控制台 API（`/api/*`，节选）

| 前缀 | 职责 | 鉴权倾向 |
|------|------|----------|
| `/api/setup` `/api/status` | 安装与状态 | 公开/半公开 |
| `/api/user/*` | 注册登录、用户自助、Passkey | 混合 |
| `/api/oauth/*` | GitHub/Discord/OIDC/微信/Telegram 等 | 限流 |
| `/api/channel/*` | 渠道 CRUD、测试、upstream、**duplicates/merge** | Admin + authz |
| `/api/token` `/api/usage` | 令牌与用量 | 用户/Admin |
| `/api/log` `/api/data` | 日志与统计 | Admin/用户分层 |
| `/api/subscription/*` | 订阅与管理 | 混合 |
| `/api/option` `/api/models` `/api/vendors` | 配置与模型元数据 | Admin |
| `/api/authz/*` | 权限目录 | Admin |
| `/api/system-task` `/api/system-info` | 系统任务与信息 | Admin |
| `/api/redemption` `/api/group` `/api/prefill_group` | 兑换码/分组 | Admin |
| `/api/ratio_sync` `/api/performance` `/api/perf-metrics` | 倍率同步与性能 | Admin/用户 |
| Webhooks | `/api/stripe|creem|waffo/webhook` | 签名校验 |
| Dashboard 兼容 | `/dashboard/billing/*` `/v1/dashboard/billing/*` | Token |

**渠道合并（本分支能力）**：`GET /api/channel/duplicates`，`POST …/merge/preview`，`POST …/merge`。

### 5.2 Relay / 兼容协议（高并发 IO — **Go 必留**）

| 前缀 | 说明 |
|------|------|
| `/v1/models` `/v1/chat/completions` `/v1/completions` `/v1/responses` … | OpenAI 兼容 |
| `/v1/messages` | Claude 形态 |
| `/v1/images/*` `/v1/embeddings` 等 | 多模态 |
| `/v1/realtime` | WebSocket |
| `/v1beta/models` | Gemini |
| `/pg/chat/completions` | Playground（UserAuth + Distribute） |
| `/v1/video*` `/kling/*` `/jimeng` 等 | 异步任务视频 |
| Midjourney / Suno 等 | 历史兼容路径（router/video 与 mj 分组） |

### 5.3 信任边界（As-Is 已基本符合 SSOT）

- 面板经同源或 CORS 调 Go；**无**面板直连「Python 内网」问题（因无 Python）。
- 密钥/渠道 key 在服务端与 DB；前端仅管理操作。
- `/metrics` 设计为边缘不代理、内网 scrape（ADR）。

---

## 6. 数据存储

### 6.1 主库（`SQL_DSN`）

| 方言 | 驱动 | 备注 |
|------|------|------|
| SQLite | glebarez/sqlite | 默认；连接池收紧 + PRAGMA |
| MySQL | gorm mysql | 中文 charset 检查 |
| PostgreSQL | gorm postgres | 保留字列名分支 |

`migrations/` 已是新 schema 变更的文件真源，含 `main/000001_baseline` 与 `cmd/dbmigrate`。现阶段 empty-DB 自动验证仅覆盖 SQLite；既有安装和 MySQL/PostgreSQL 仍受 AutoMigrate 兼容约束，详见 [`migrations/README.md`](../migrations/README.md)。

### 6.2 日志库（`LOG_SQL_DSN`）

- 可与主库分离；支持 **ClickHouse**（仅日志侧）。
- `Log` 模型 AutoMigrate 到 LOG_DB。

### 6.3 缓存

- **Redis**（`REDIS_CONN_STRING`）：可选；关闭则内存路径。
- 渠道/令牌/用户缓存：`model/*_cache.go`、`pkg/cachex`。

### 6.4 AutoMigrate 核心实体（主库）

`Channel, Token, User, PasskeyCredential, Option, Redemption, Ability, Log, Midjourney, TopUp, QuotaData, Task, Model, Vendor, PrefillGroup, Setup, TwoFA(+Backup), Checkin, SubscriptionOrder, UserSubscription, SubscriptionPreConsumeRecord, CustomOAuthProvider, UserOAuthBinding, PerfMetric, SystemInstance, SystemTask(+Lock), CasbinRule, AuthzRole, RefundIntent, SubscriptionPlan`（SQLite 特殊建表路径）。

### 6.5 SQL 与 SSOT

- **SQL 作为通用工具**：已用，但 **迁移规范未产品化** → Phase1 T2。
- Postgres 优先的目标与「三方言必须同时支持」的 fork 约束冲突：迁移期需 **兼容层**，不能单切 PG 砍 SQLite/MySQL（除非产品决策降级部署形态）。

---

## 7. 关键子系统（网关内）

| 子系统 | 位置 | 说明 |
|--------|------|------|
| 渠道选择 / 熔断 / 自适应影子 | `service/channel_*.go` | 热路径；适合 **留 Go**，策略评测可旁路 Python |
| 计费 / 预扣 / 差额 / 表达式 | `service/billing*.go` `pkg/billingexpr` `common/quota_math.go` | **正确性高风险**；迁移最后动 |
| Refund outbox | `service/refund_outbox.go` `model/refund_intent.go` | 异步一致性 |
| 系统任务 | `model/system_task*.go` `service/system_task.go` | scheduler/worker |
| 鉴权 | JWT · Session · WebAuthn · OAuth · Casbin authz | 必须 **单点在 Go** |
| Relay 适配器 | `relay/channel/*`（38+ 厂商） | 长期 Go |
| 可观测 | OTEL 可选 · Prometheus 告警样例 · RUM | 横切 |

---

## 8. 与目标栈偏离清单

| ID | 项 | 现状 | 主参考 | 标签 | 策略（摘要） |
|----|----|------|--------|------|--------------|
| D1 | 管理面板 | React19 + BaseUI + Tailwind；classic Semi；Vue `web-console` Phase1 只读 | Vue3 + NaiveUI + TS | **P0 gate** | 见 [`ARCHITECTURE_TARGET.md`](ARCHITECTURE_TARGET.md)；未完成 gate 前 React 仍默认 |
| D2 | classic 主题 | 第二套 React | 同层一种主实现 | **L2** | 冻结；不跟 default 双写新功能 |
| D3 | AI 核心 | 无 Python；逻辑在 Go service | Python AI-Core | **P1** | 旁路评测/批跑；**热路径留 Go**（R2 证据：IO/计费） |
| D4 | SQL 迁移 | SQL migrations 已建，empty baseline 仅 SQLite | 可审 SQL migrations | **P0/T2** | 三方言基线与 CI 策略仍待完成 |
| D5 | OpenAPI 契约 | 有 json，未必与路由同步 | Go 网关契约真源 | **P0/T1** | 路由表 + 校验 |
| D6 | Electron | 独立壳 | 非面板主栈 | **TOOL/L2** | 维持 |
| D7 | C/嵌入式 | 无 | 副线 | **SIDE** | 仓外 |
| D8 | 模块/身份 | `xvyimu/TransitHub` | 独立产品 | **OK** | 保持 |
| D9 | 支付多提供商 | Stripe/Creem/Waffo/Epay… | 网关职责 | **OK** | 留 Go |
| D10 | 交付缝 | embed + separated | 分发灵活 | **OK 切点** | 面板策略共用 |
| D11 | Go 网关 | Gin+GORM+relay 深 | Go | **OK / 加深** | **主参考命中；禁止换语言** |

### 8.1 面板 React 偏离 · R2 决策表（保留 / 绞杀 Vue / 证据）

主参考默认 **TS+Vue3+NaiveUI**。TransitHub 面板已是 **深 React**（default features 二十余域 + classic + embed/Electron）。按总文档 **R2/R4/R6** 三选一，**推荐项标 ★**（须用户/协调员 gate 后再生效）。

| 选项 | 含义 | 对比维度与证据 | 代价 | 回滚 |
|------|------|----------------|------|------|
| **A. 中期保留 React（default）** | 面板主实现仍 React；治理 classic；**不**为对齐强制 Vue | **契合度**：`web/default` 已 React19+TanStack+Rsbuild+i18n，与 `/api` cookie 会话、Playground SSE 已耦合。**工期**：Vue 全量重写 ≈ 面板全功能二次交付，相对「Go 边界/计费/渠道」P0 收益低。**心智**：团队若已在 React 树上修生产问题，并行 Vue 违反 R4 双写风险。 | 与产品线「默认 Vue」不一致；招聘/跨仓组件难复用 | 无（维持现状） |
| **B. 绞杀 Vue（分阶段）★ 默认推荐若要强对齐主参考** | 新建 `web-console/`（Vue3+NaiveUI）；`FRONTEND_MODE`+`deploy/separated` 同域切流；旧 React **LEGACY** 只修 bug | **架构清晰**：交付缝 **已存在**（ADR-0001、`frontend_external`、nginx 同域）→ 绞杀成本低于「无缝 monorepo」。**同层一种**：切流完成后只留 Vue（R4）。**可测**：先只读页（status/渠道列表/日志）行为对比。 | 高：features 域多；双轨期 CI/i18n/权限矩阵；NaiveUI≠BaseUI 交互重做 | 边缘指回 embedded React 二进制；配置级回滚（ADR 已写） |
| **C. 混合长期双面板** | React 与 Vue 长期并存 | **无产品证据** 支撑双主实现 → **R3 禁止** | 双倍缺陷面 | N/A |

**测绘师建议（非最终 gate）**

1. **Go：保留并加深**（D11）— 主参考命中，证据=体量与热路径，无需 R2 替代语言。  
2. **面板：若用户确认「产品线统一 Vue」→ 选 B（绞杀）**；若确认「TransitHub 面板例外、优先网关正确性」→ 选 **A**，标签改为 **L2 面板偏离 + 书面 R2 接受**，P0 只做 Go/OpenAPI/SQL。  
3. **禁止 C**。  
4. **无论 A/B**：`web/classic` → **L2 冻结**；禁止再开第三套 UI。

**可测量验收（若选 B）**

- 同域登录 cookie 一轮；`/api/status` 与一页渠道只读数据一致；  
- 切流开关可 5 分钟内回到 React embed；  
- 新功能 CODEOWNERS：禁止进 `web/default`（仅 security hotfix）。

---

## 9. 可绞杀切点（Strangler 优先序）

1. **交付缝（已有）**  
   `FRONTEND_MODE=disabled|redirect` + `deploy/separated` → 新 Vue 控制台同域反代。  
2. **只读页面优先**  
   健康/状态、渠道列表只读、日志只读、system-info → 行为易对齐。  
3. **写路径后置**  
   渠道 merge、密钥、充值、option 写入 → 依赖 authz + CSRF/cookie。  
4. **API 版本**  
   现有 `/api/*` 保持；新面板可先消费旧 API，再抽 `/api/v2`（可选）。  
5. **Python 旁路**  
   新进程 + Go 出站（鉴权透传）；**永不**让面板直连 Python。  
6. **不要先动**  
   `relay/*` 热路径、配额数学、三方言 AutoMigrate 大爆炸。

---

## 10. 迁移风险

| 风险 | 等级 | 说明 | 缓解 |
|------|------|------|------|
| 计费/配额回归 | **高** | int32 配额、预扣差额、表达式 | 冻结计费改动；属性/边界测全绿再碰 |
| Session/Cookie 跨前端 | **高** | 分域会破登录 | **强制同域 Nginx**（ADR 已论证） |
| OAuth 回调 URL | **中** | 换控制台源要重配 | 运维清单；先同 host |
| SSE/WS 缓冲 | **中** | 反代默认缓冲 | separated 模板已有方向；验收 realtime |
| 三方言 SQL | **中高** | 迁 migrations 易漏方言 | 矩阵 CI：sqlite/mysql/pg |
| React→Vue 功能面 | **高** | features 二十余域 | 分期绞杀；CODEOWNERS 禁新功能进旧面板 |
| OpenAPI 漂移 | **中** | 文档与路由不一致 | T1 对表 + CI diff |
| 自适应选路语义 | **中** | 影子对比日志已有 | 迁出前固定契约测试 |
| Electron 分发 | **低** | 壳可继续指新 URL | 非 P0 |
| 范围蔓延 | **高** | fork 体量大 | 严格 Phase；禁止微重构占坑 |

---

## 11. 建议标签（本仓）

| 标签 | 应用到 |
|------|--------|
| **P0** | Go 边界加固、OpenAPI、SQL 迁移规范；面板若 gate=B 则 Vue 脚手架+只读绞杀 |
| **P1** | Python AI-Core 旁路（评测/批跑）；高级运营页（仅 B） |
| **L2** | web/classic；若 gate=A 则 default React 整面板标「已接受偏离」 |
| **LEGACY** | gate=B 时 `web/default` 只修 bug/安全 |
| **TOOL** | `electron/`、部分 `scripts/` |
| **SIDE** | 嵌入式（仓外） |
| **OK** | Go relay 网关、鉴权、渠道、计费核心位置（语言选择正确） |

---

## 12. 与总规划 Phase 的映射

| Phase1 步 | As-Is 结论 |
|-----------|------------|
| T1 模块边界 + OpenAPI | 模块清晰；补「路由真源 vs openapi」 |
| T2 SQL schema/迁移规范 | `migrations/` 与 SQLite CI 已落地；MySQL/PostgreSQL baseline 是未关闭 gate |
| T3 `web-console` Vue3+NaiveUI | Vue 工程已存在；默认交付仍受 [`ARCHITECTURE_TARGET.md`](ARCHITECTURE_TARGET.md) gate 约束 |
| T4 只读页绞杀 | health/status、渠道、模型、日志有只读切片；写路径后置 |
| T5 旧 React deprecated | 仅 B；A 则写 R2 接受说明替代 deprecated |

---

## 13. 明确不做（测绘结论）

- 不在本阶段把 relay 热路径改写为 Python。  
- 不引入第二 Go Web 框架；不引入 Java/C#/Rust 主栈。  
- 不把嵌入式代码塞进本 monorepo。  
- 不为「整洁」重写计费（无契约与全量回归不碰）。  
- 不以「支付回调模板」或无关 ISS 微重构占用 P0 带宽。

---

## 14. 证据索引（路径）

| 主题 | 路径 |
|------|------|
| 入口 | `main.go` |
| 路由聚合 | `router/main.go` `api-router.go` `relay-router.go` `web-router.go` `channel-router.go` |
| 交付缝 ADR | `docs/adr/0001-frontend-backend-delivery-seam.md` |
| 分发 | `deploy/separated/*` `frontend_assets_*.go` |
| DB | `model/main.go` |
| OpenAPI | `docs/openapi/api.json` `relay.json` |
| 前端 | `web/package.json` `web/default/package.json` `web-console/package.json` |
| Target SSOT | `docs/ARCHITECTURE_TARGET.md` |

---

## 15. 测绘元数据

- 只读：未改业务代码、未 commit、未 push。  
- 语言行数为 PowerShell 粗计，供占比决策，非严格 cloc。  
- 工作区测绘时主仓分支含 channel-merge 等演进；模块边界结论对同源 worktree 同样适用。
- **下一步**：按 `ARCHITECTURE_TARGET.md` 关闭 Vue CI/NOTICE、三方言迁移和非生产 smoke gate；在明确授权前禁止切流。

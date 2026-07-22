# Phase1 Bid-A · Go 网关架构加深

| 字段 | 值 |
|------|-----|
| 角色 | Go 网关架构师 · 槽 `th-2` |
| 日期 | 2026-07-22 |
| 性质 | **只出方案**；禁止业务代码 commit/push |
| 决议 | `D:\orca\docs\architecture-decision-2026-07-22-approved.md`（P0=TransitHub；面板=B 绞杀 Vue） |
| Master | `D:\orca\docs\architecture-stack-refactor-master-2026-07-22.md`（R1–R6） |
| As-Is | `D:\TransitHub\src\docs\ARCHITECTURE_ASIS.md`（worktree 无镜像，以主仓为准） |
| 标签 | Bid-A · T1 主责；与 Bid-B（Vue）、Bid-C（SQL）接口对齐 |

---

## 1. 目标与非目标

### 1.1 目标（Phase1 · 网关侧）

1. **模块边界书面化**：固定 `router → middleware → controller → service → model` 与 `relay/channel` 旁路职责；禁止新功能把计费/选路写进 controller 或跨层直调。
2. **OpenAPI 草稿可验收**：以**路由注册为真源**，对 `docs/openapi/api.json`（管理面 ~131 paths）与 `relay.json`（~35 paths）做对齐审计 + 缺口清单；为新面板定义 **Console API 契约子集**（登录会话 + 健康/状态 + 一页只读）。
3. **Plane 拆分可运维**：文档化并验收 `RUN_MODE` × `APP_PLANE` × `FRONTEND_MODE` 矩阵；明确 management 纯 API 形态给 Vue 绞杀用。
4. **与新面板的 API 契约**：同域 cookie 会话、CSRF/同源假设、CORS 边界、`/api/status` / 探活 / 渠道只读等首批路径的请求/响应约定。
5. **可观测与限流边界**：HTTP 指标/RUM/日志相关键（`request_id`/`trace_id`）与限流层级（全局 API / critical / model / body）职责不交叉；management 与 relay 指标可按 `plane` 标签分面。
6. **Phase2 缝预留**：AI 热路径（relay IO、预扣/结算、渠道选择熔断）**保留在 Go**（R2）；仅预留「旁路 AI-Core」出站接口形状，**不实现 Python**。

### 1.2 非目标

| 不做 | 原因 |
|------|------|
| 改 controller/service 业务逻辑 / 计费公式 / 渠道 merge 语义 | 本任务只写方案；计费属高风险冻结面 |
| 把 relay 热路径迁 Python | 决议 + As-Is D3/D11：热路径留 Go |
| 换 Web 框架 / 拆 Go monorepo 为多 Git 仓 | ADR-0001 已否决双仓；Gin 加深 |
| 实现 Vue 面板 / SQL migrations 目录落地 | Bid-B / Bid-C；本方案只定义对接面 |
| 全量 OpenAPI 100% 自动生成上 CI（实现期可做） | Phase1 先人工对表 + 契约子集；生成管线 Phase1 末或 Phase3 |
| 新功能进 `web/default` | 决议 B：旧 React 只读维护 |

---

## 2. 改动文件/目录范围

### 2.1 本竞标阶段（文档 only · 本 PR/任务）

| 路径 | 动作 |
|------|------|
| `docs/phase1-bid-go-gateway.md` | **本文件** |
| （可选协调员合并后）`docs/ARCHITECTURE_TARGET.md` | 一页纸指向 Master；**不**在本 worker 强行写，除非协调员指定 |

### 2.2 Phase1 实现期（模块二 · 方案通过后）建议触达范围

**允许加深（边界/契约/运维，非业务语义）**

| 目录/文件 | 用途 |
|-----------|------|
| `docs/openapi/` | 对齐后的草稿；可增 `console-v1.yaml` 或 `api-console-subset.json` |
| `docs/architecture/` 或 `docs/gateway/`（新建） | 模块边界图、plane 矩阵、契约表、限流/可观测边界 |
| `router/*.go` | 仅：路由表导出辅助、注释契约、**不改** handler 业务 |
| `middleware/` | 仅：文档化 + 必要时 tag/metric 一致性；避免改鉴权语义 |
| `pkg/observability/` | 指标命名/plane 标签文档与小补丁（若需） |
| `deploy/separated/*`、`docs/operations/*` | 与 Vue 同域反代、plane 运维交叉引用 |
| `main.go` / `runtime_mode.go` | 仅配置解析说明与只读验收脚本，不改默认行为语义 |
| `scripts/` | 路由 vs OpenAPI diff 脚本（只读检查） |

**明确禁止触达（Phase1）**

- `service/billing*.go`、`common/quota_math.go`、`pkg/billingexpr/**` 语义
- `relay/channel/**` 适配器行为、上游请求形态
- `model/` 字段/AutoMigrate 大爆炸（归 Bid-C）
- `web/default`、`web/classic` 新功能
- `controller/*` 写路径业务（登录/充值/merge 等）除**契约文档**外不改实现

### 2.3 建议新增文档制品（实现期）

```text
docs/gateway/
  MODULE_BOUNDARIES.md      # 分层与依赖方向
  PLANE_MATRIX.md           # RUN_MODE × APP_PLANE × FRONTEND_MODE
  CONSOLE_API_CONTRACT.md   # 新面板首批路径契约
  RATE_LIMIT_BOUNDARIES.md  # 限流层级
  OBSERVABILITY.md          # metrics/RUM/日志字段
  OPENAPI_AUDIT.md          # 路由真源 vs openapi  diff 结果
docs/openapi/
  console-subset.yaml       # 或 json：Vue 首批消费面
```

---

## 3. 依赖与上下游

### 3.1 上游（依赖）

| 依赖 | 说明 |
|------|------|
| 用户决议 | 面板 B 绞杀；P0=TH |
| As-Is 测绘 | 模块图、偏离 D1–D11、绞杀切点 ADR-0001 |
| 现有交付缝 | `FRONTEND_MODE`、`frontend_external`、`deploy/separated` |
| 现有 plane | `router.PlaneAll|Relay|Management` + `RUN_MODE` |
| Bid-C | schema 稳定后，契约里的字段类型以 DB 真源为准（只读页） |

### 3.2 下游（被依赖）

| 下游 | 本方案提供 |
|------|------------|
| Bid-B Vue `web-console` | Console API 子集、同域 cookie、status/health、错误码约定 |
| 运维 / deploy | plane 与 FRONTEND_MODE 推荐组合（management + disabled） |
| Phase2 AI-Core | Go 出站「旁路客户端」接口缝（HTTP 内部，面板永不直连） |
| Phase3 横切 | Auth/JWT 单点在 Go 的规范落点确认 |

### 3.3 信任边界（不变）

```text
[ Vue web-console ] ──HTTPS 同域──► [ Go management plane /api ]
                                         │
                    ┌────────────────────┼────────────────────┐
                    ▼                    ▼                    ▼
              SQL/Redis            [ Go relay plane ]    (Phase2) AI-Core
                                    /v1 /v1beta …         仅 Go 出站
```

- 面板 **不** 直连内部 AI 端口（即使 Phase2 也无例外）。
- 密钥与渠道 key 只在服务端。
- `/metrics` 不进公网 console origin（separated 边缘 404）。

### 3.4 逻辑分层（加深后的依赖方向）

```text
router/          注册路径 + 中间件链；无业务
middleware/      鉴权、限流、CORS、分发、审计、route_tag
controller/      HTTP 绑定、校验入参、调 service；薄
service/         领域编排：计费会话、渠道、退款、任务…
model/           GORM 与缓存；无 HTTP 类型泄漏为佳
relay/           协议 handler + channel 适配；仅热路径
dto/ types/      请求响应与中性类型
common/ setting/ 横切工具与配置
pkg/*            billingexpr、observability、cachex…
```

**依赖规则（书面门禁）**

- `controller` → `service` → `model`；禁止 `controller` 大段 GORM。
- `relay` 可读 `service`/`model` 计费与渠道接口；禁止 `model` import `controller`。
- 新代码禁止引入第二 HTTP 框架。

---

## 4. 分步执行计划（S0… · 每步验收）

> 下列为 **Phase1 网关加深** 实现序列（模块二）。S0 为方案门；本 worker 只完成 S0 文档。

### S0 · 方案竞标与合成（当前）

| 项 | 内容 |
|----|------|
| 做 | 本 bid + 与 Bid-B/C 交叉点表 |
| 验收 | 协调员合成 Phase1 执行规格；用户/协调员 **gate** 通过 |
| 产出 | `docs/phase1-bid-go-gateway.md` |

### S1 · 模块边界图与公共 API 清单（Master T1）

| 项 | 内容 |
|----|------|
| 做 | 固化分层图；导出 **路由注册表**（按 plane：`management` = api+dashboard；`relay` = relay+video；`all` = 并集 + web） |
| 证据源 | `router/main.go` `SetRouterForPlane`；`api-router.go` `channel-router.go` `relay-router.go` `video-router.go` `dashboard.go` `web-router.go` |
| 验收 | `docs/gateway/MODULE_BOUNDARIES.md` + 机器可读路由清单（脚本扫 `router/*.go` 或测试快照） |
| 不做 | 重命名包、搬文件大重构 |

### S2 · OpenAPI 审计与草稿范围

| 项 | 内容 |
|----|------|
| 做 | 对比 Gin 注册路径 vs `docs/openapi/api.json` / `relay.json`；记录 **缺失 / 多余 / 方法不一致** |
| 草稿范围 | **P0 Console 子集**（见 §4.1）必须完整描述；全量管理 API 标「已知漂移，分期修」 |
| 验收 | `OPENAPI_AUDIT.md` 有 diff 表；`console-subset` 可被 Bid-B mock/对接 |
| 工具 | 可选 `scripts/openapi_route_diff.py`（只读） |

#### 4.1 OpenAPI / Console 契约 · Phase1 强制子集

| 路径 | 用途 | 鉴权 | 面板优先级 |
|------|------|------|------------|
| `GET /healthz` `GET /livez` | 探活；响应含 `plane` | 无 | P0 |
| `GET /readyz` | 就绪 | 无 | P0 |
| `GET /api/status` | 控制台启动配置（version、OAuth 开关、theme…） | 公开/半公开 | **P0 首屏** |
| `GET /api/setup` | 是否已安装 | 公开 | P0 |
| `POST /api/user/login`（及 2FA/passkey 相关若首版需要） | 登录 | Critical RL + Turnstile 策略按现状 | P0 |
| `GET /api/user/logout` | 登出 | 会话 | P0 |
| `GET /api/user/self` | 当前用户 | UserAuth | P0 |
| `GET /api/channel/`（列表只读） | 绞杀只读页 | Admin + authz | P0 T4 |
| `GET /api/system-info` 或现有 system 只读等价 | 健康信息页 | Admin | P0 |
| `POST /api/rum` | Web Vitals（无 PII） | 匿名 body limit | P1 可后 |
| Relay `/v1/*` | **不在** Vue 首版管理契约内 | TokenAuth | 文档仅交叉引用 relay.json |

**契约约定（实现期写入 CONSOLE_API_CONTRACT）**

- 会话：同域 cookie（`SESSION_COOKIE_*`）；**禁止**默认跨站 SPA + 宽 CORS 作为主路径。
- 错误体：沿用现有 `success`/`message` 模式；子集路径列出现状字段，不发明 v2 直到需要。
- **Phase1 不强制** 上 `/api/v2`；新面板先消费现有 `/api/*`。

### S3 · Plane 与前端交付矩阵验收

| 组合 | 用途 | 验收命令/行为（示意） |
|------|------|------------------------|
| `RUN_MODE=all` `APP_PLANE=all` `FRONTEND_MODE=auto` | 兼容单体 | 旧行为：embed SPA |
| `RUN_MODE=serve` `APP_PLANE=management` `FRONTEND_MODE=disabled` | **Vue 反代后端** | 仅 management 路由；无 NoRoute SPA |
| `RUN_MODE=serve` `APP_PLANE=relay` | 热路径实例 | 无 `/api` 管理写面；有 `/v1` + health |
| `RUN_MODE=worker` / `scheduler` / `migrate` | 非 HTTP 或短命 | 见 `docs/operations/runtime-separation.md` |

| 验收 | `PLANE_MATRIX.md` 与 `runtime-separation.md` 一致；手工或集成测：错误 plane 启动失败信息可读 |

### S4 · 限流边界文档化（+ 仅必要时补齐 route_tag）

| 层级 | 位置 | 作用面 |
|------|------|--------|
| Global API | `middleware.GlobalAPIRateLimit` on `/api` | 管理面全局 |
| Critical | 登录/注册/OAuth/重置密码等 | 防爆破 |
| Email verification | 专用中间件 | 邮箱滥发 |
| Model request | `model-rate-limit` | Relay/模型维度 |
| Body limit | 匿名/全局 body | 防大包 |
| In-memory / Redis | `common` + `common/limiter` | 部署形态切换 |

| 验收 | `RATE_LIMIT_BOUNDARIES.md`：哪一层保护 management vs relay；**不**在 Phase1 重写 Redis 列表限流算法 |

### S5 · 可观测边界

| 信号 | 位置 | 边界 |
|------|------|------|
| Prometheus HTTP | `pkg/observability`：`plane`,`route_class`,`method`,`route`,`status` | 用 `FullPath` 模板，禁原始高基数 path |
| `/metrics` | `METRICS_ENABLED` + 可选 `METRICS_TOKEN` | 内网 scrape；separated 边缘不代理 |
| RUM | `POST /api/rum` | 仅 name/value/rating；无 URL/用户/token |
| 日志 | `request_id` / `trace_id` | 与 SLO 文档 drill-down 一致 |
| 自适应指标 | Redis 快照（既有） | 仍属 Go 热路径周边；Phase2 评测旁路不替换 |

| 验收 | `OBSERVABILITY.md` + 对照 `docs/operations/slo.md`；management P95 与 relay 可用性分面描述清晰 |

### S6 · 与 Bid-B 联调缝（文档 + 配置，不改业务）

| 项 | 内容 |
|----|------|
| 做 | 确认 `deploy/separated` 反代前缀列表含 `/api` `/v1` … `/healthz`；Vue base 同域 |
| 验收 | Bid-B 登录一轮 + `/api/status` 字段字典一致；切 `FRONTEND_MODE` 可回滚 React embed（配置级） |

### S7 · Phase2 接口缝（仅设计，不实现）

| 缝 | 形状（草案） | 约束 |
|----|--------------|------|
| `service/aicore` 客户端（未来） | 内部 HTTP：`POST /internal/v1/eval` 等 | mTLS 或网络策略；**转发**已鉴权的 tenant/user id，不传面板 cookie |
| 配置 | `AI_CORE_BASE_URL` 空=关闭 | fail-open 或显式错误策略在 Phase2 定 |
| 热路径 | relay 完成/计费 **不** 等待 AI-Core | 异步/旁路 |

| 验收 | 文档一节 + 空接口注释位置建议（如 `service/` 下 package 名）；**无 Python 代码** |

### S8 · 旧面板与治理（配合 Bid-B）

| 项 | 内容 |
|----|------|
| 做 | CODEOWNERS/文档：禁止新功能进 `web/default`；Go 侧无变更也可 |
| 验收 | 与 Master T5 一致 |

---

## 5. 风险与回滚

| 风险 | 等级 | 缓解 | 回滚 |
|------|------|------|------|
| OpenAPI 全量对齐工期爆炸 | 中 | 子集优先；全量只出 diff 清单 | 保留现有 json，不删 |
| 误改鉴权/限流导致锁死登录 | 高 | Phase1 **默认不改** middleware 语义 | git revert 中间件 diff |
| 文档与路由再次漂移 | 中 | S1 路由清单脚本进 CI（实现期） | 关闭 CI gate |
| Plane 配置错误致生产只开了 relay | 中 | 启动日志打印 plane；就绪探针 | `APP_PLANE=all` |
| 与 Vue 跨域破 cookie | 高 | **强制同域** Nginx（ADR-0001） | 回 embedded |
| 范围蔓延进计费/relay | 高 | 本 bid 禁止列表 + 审查 | 拒绝 PR |
| Bid-C 迁移与契约字段不一致 | 中 | 只读页字段先冻结；写路径后置 | 契约标 draft |

**配置级回滚（网关交付）**

1. `RUN_MODE=all` `APP_PLANE=all`，`FRONTEND_MODE` 空/auto → 单体 embed。  
2. 去掉 `frontend_external` 构建标签，恢复集成镜像。  
3. DB：本方案不引入破坏性迁移。

---

## 6. 测试建议

| 类型 | 内容 |
|------|------|
| 已有单测 | `router/main_test.go` plane/frontend mode；`middleware/*_test` CORS/auth；保持绿 |
| 路由快照 | 测试或脚本：`SetRouterForPlane` 三 plane 注册路径集合稳定 |
| OpenAPI diff | CI 可选：fail on **console-subset** 路径缺失（非全量） |
| 契约手工 | `GET /api/status` 字段白名单；login cookie `Set-Cookie` 同站 |
| 分 plane 冒烟 | management 实例 `GET /v1/models` 应 404；relay 实例 `GET /api/status` 应 404 |
| 限流 | critical 路径 429 行为不回归（表测或集成） |
| 指标 | `METRICS_ENABLED=true` 时 `/metrics` 含 `newapi_http_requests_total` 且带 `plane` |
| 禁止 | 无业务语义的随机 fuzz；不碰计费属性测除非独立任务 |

---

## 7. 相对主参考偏离（若有）+ 证据

| 点 | 主参考 | 本方案立场 | 证据 / R2 |
|----|--------|------------|-----------|
| 网关语言 | Go | **加深 Go，不换语言** | As-Is ~129k Go、relay 38+ 适配、计费与配额在 Go；换语言无收益（R1/R4） |
| AI 层 | Python AI-Core | **热路径留 Go**；Python=Phase2 旁路 | 延迟/IO/预扣原子性；Master 亦允许「热路径 Go 推理代理」；决议非立即 MindSync 抽离 |
| 面板 | Vue3+NaiveUI | **不实现面板**；契约迎合绞杀 B | 决议 #3；ADR-0001 交付缝 |
| API 版本 | 清晰 OpenAPI | **先审计子集**，不全量重写 | 现有 api.json/relay.json 体量大且可能漂移（As-Is D5） |
| SQL | 规范 migrations | **不负责落地**；依赖 Bid-C | 决议 Bid-C |
| 同层一种实现 | R4 | 管理 API 仍一套 Gin；不并行第二 BFF | — |

**无**「用 Rust/C# 重写网关」或「管理流量走 Node BFF」类偏离提案。

---

## 8. 工作量粗估

| 步骤 | 人日（约） | 角色 |
|------|------------|------|
| S0 方案（本任务） | 0.5 | Go 架构师 |
| S1 边界图 + 路由清单 | 1–1.5 | Go |
| S2 OpenAPI 审计 + console-subset | 1.5–2.5 | Go + 轻量前端对接 |
| S3 Plane 矩阵文档与冒烟 | 0.5–1 | Go / DevOps |
| S4–S5 限流/可观测文档 | 0.5–1 | Go |
| S6 与 Vue 契约联调支持 | 1（并行 Bid-B） | Go + Vue |
| S7 Phase2 缝设计 | 0.5 | Go |
| S8 治理文档 | 0.25 | 任意 |
| **Phase1 网关合计** | **约 5.5–8 人日** | 不含 Vue/SQL 正文 |

并行：S2 与 Bid-B 脚手架可重叠；S6 在 T3 登录页后。

---

## 9. 与其它 Bid 的接口表（协调员合成用）

| 交叉点 | Bid-A（本） | Bid-B Vue | Bid-C SQL |
|--------|-------------|-----------|-----------|
| 登录 | 契约：cookie/status | 实现 UI | 用户表稳定 |
| 只读渠道页 | 路径与 authz | 页面 | channel 列类型 |
| 部署 | `FRONTEND_MODE=disabled` + plane | Nginx SPA | migrate job 顺序 |
| 禁止 | 业务 handler 大改 | 新功能进 React | 无审 migration |

---

## 10. 成功标准（Bid-A 视角 · Phase1 末）

- [ ] 模块边界与 plane 矩阵文档合入主仓 `docs/`  
- [ ] Console API 子集 OpenAPI/契约可供 Vue 联调  
- [ ] OpenAPI 审计清单存在（允许全量未清零）  
- [ ] management+disabled 与 relay 分面冒烟通过  
- [ ] 限流/可观测边界成文，且无未审 middleware 语义变更  
- [ ] AI-Core 仅预留缝；热路径仍在 Go  
- [ ] **无** 业务逻辑 commit 混入方案门前工作  

---

## 11. 证据索引

| 主题 | 路径 |
|------|------|
| Plane 路由 | `router/main.go`（`SetRouterForPlane`） |
| 管理 API | `router/api-router.go`、`channel-router.go` |
| Relay | `router/relay-router.go`、`video-router.go` |
| 交付缝 | `docs/adr/0001-frontend-backend-delivery-seam.md` |
| 运行时分离 | `docs/operations/runtime-separation.md` |
| SLO | `docs/operations/slo.md` |
| 指标 | `pkg/observability/metrics.go` |
| OpenAPI | `docs/openapi/api.json`、`relay.json` |
| As-Is | `D:\TransitHub\src\docs\ARCHITECTURE_ASIS.md` |
| 决议 / Master | `D:\orca\docs\architecture-decision-2026-07-22-approved.md`、`architecture-stack-refactor-master-2026-07-22.md` |

---

*本文件为 Phase1 模块一竞标产物（Bid-A）。协调员择优合成执行规格前，不授权业务改码。*

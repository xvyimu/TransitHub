# Go 网关模块边界（Phase1 · WP-G G1）

| 字段 | 值 |
|------|-----|
| 日期 | 2026-07-22 |
| 规格 | `D:\orca\docs\phase1-execution-spec-transithub-2026-07-22.md` §3 WP-G |
| 详案 | `docs/phase1-bid-go-gateway.md` |
| 真源 | **Gin 路由注册**（`router/*.go`），非 OpenAPI |
| 性质 | 文档加深；**不改** controller/service 业务语义 |

---

## 1. 分层与依赖方向

```text
Client / SPA / SDK
        │  HTTPS · JSON · SSE · WS
        ▼
┌─ router/ ─────────────────────────────────────────────┐
│  注册路径 + 中间件链；按 APP_PLANE 裁剪面                │
│  main.go → api / dashboard / relay / video / web       │
└───────────────┬───────────────────────────────────────┘
                ▼
┌─ middleware/ ─────────────────────────────────────────┐
│  CORS · Auth(User/Admin/Root/Token) · RateLimit        │
│  Distribute · RouteTag · BodyLimit · Audit · Recover   │
└───────────────┬───────────────────────────────────────┘
                ▼
┌─ controller/ ─────────────────────────────────────────┐
│  HTTP 绑定、入参校验、调用 service；保持薄              │
└───────────────┬───────────────────────────────────────┘
        ┌───────┴────────┐
        ▼                ▼
┌─ service/ ──┐   ┌─ relay/ ──────────────────────────┐
│ 计费会话     │   │ 协议 handler + channel/* 上游适配  │
│ 渠道选择/熔断│   │ **热路径 IO 留 Go**（Phase1/2）    │
│ refund/task  │   └────────────────────────────────────┘
└──────┬──────┘
       ▼
┌─ model/ ────┐     ┌─ dto/ types/ constant/ ─┐
│ GORM · cache│     │ 请求/响应与中性类型       │
└─────────────┘     └─────────────────────────┘
       │
       ▼
 SQL · Redis(可选) · ClickHouse(日志可选)
```

**横切包**

| 包 | 职责 |
|----|------|
| `common/` | JSON 封装、env、配额数学、Redis、内存限流 |
| `setting/` | 比例/模型/运营/性能配置 |
| `oauth/` · `i18n/` · `logger/` | 登录提供商、文案、日志 |
| `pkg/billingexpr` | 计费表达式（**冻结语义**） |
| `pkg/observability` | Prometheus HTTP / RUM 直方图 |
| `pkg/cachex` · `perf_metrics` · `ionet` | 缓存与辅助 |

---

## 2. 依赖规则门禁（Phase1）

| 规则 | 说明 |
|------|------|
| `controller → service → model` | 禁止 controller 大段 GORM |
| `relay` 可读 service/model | 禁止 `model` import `controller`/`router` |
| 禁止第二 HTTP 框架 | 继续 Gin |
| 禁止新业务写进 `web/default` | 面板绞杀见 WP-V |
| 禁止改计费/relay 热路径语义 | WP-G 仅文档/只读脚本 |
| JSON | 业务代码走 `common.Marshal/Unmarshal*` |

---

## 3. Plane 路由装配（代码真源）

`router.SetRouterForPlane`（`router/main.go`）：

| `APP_PLANE` | 注册内容 |
|-------------|----------|
| `all` | probes + **management**（API + dashboard）+ **relay**（relay + video）+ frontend（按 `FRONTEND_MODE`） |
| `management` | probes + API + dashboard + frontend |
| `relay` | probes + relay + video；**无** `/api` 管理面、无 SPA |

探活（**所有** HTTP plane）：

| Method | Path | Handler |
|--------|------|---------|
| GET | `/healthz` | 内联 `{"status":"ok","plane":...}` |
| GET | `/livez` | 同上 |
| GET | `/readyz` | `controller.GetReadiness`（DB + 可选 Redis） |

`RUN_MODE` 控制进程是否起 HTTP / worker / scheduler / migrate，见 `runtime_mode.go` 与 `docs/gateway/PLANE_MATRIX.md`。

---

## 4. 包 → 面映射

| 注册入口 | 文件 | Plane | Route tag（指标） |
|----------|------|-------|-------------------|
| `SetApiRouter` | `api-router.go` + `channel-router.go` + `authz-router.go` | management / all | `api` |
| `SetDashboardRouter` | `dashboard.go` | management / all | `old_api` |
| `SetRelayRouter` | `relay-router.go` | relay / all | `relay` |
| `SetVideoRouter` | `video-router.go` | relay / all | `relay` |
| `SetWebRouter` / redirect / disabled | `web-router.go` + `main.go` | management / all | `web` |

---

## 5. 管理面 API 分组（`/api/*`）

> 路径片段由 `scripts/scan_gin_routes.py` 静态扫描；完整列表见 `ROUTE_TABLE.md`。  
> 扫描计数（2026-07-22）：`api-router` 224 片段 + `channel` 41 + `authz` 1 ≈ 管理面主体。

| 前缀 | 鉴权倾向 | 说明 |
|------|----------|------|
| `/api/setup` `/api/status` `/api/notice` … | 公开/半公开 | 安装与控制台启动配置 |
| `/api/user/*` | 混合 | 注册登录、self、Admin 用户 CRUD |
| `/api/oauth/*` | Critical RL | OAuth / 非标登录 |
| `/api/channel/*` | Admin + authz 权限点 | 渠道 CRUD/测试/merge 等 |
| `/api/token` `/api/usage` | User / Token | 令牌与用量 |
| `/api/log` `/api/data` | Admin/User 分层 | 日志与统计 |
| `/api/subscription/*` | User/Admin | 订阅 |
| `/api/option` `/api/models` `/api/vendors` | Root/Admin | 配置与模型元数据 |
| `/api/authz/catalog` | Admin | 权限目录 |
| `/api/system-task` `/api/system-info` | Root | 系统任务与实例 |
| `/api/redemption` `/api/group` `/api/prefill_group` | Admin | 兑换码/分组 |
| `/api/ratio_sync` `/api/performance` `/api/perf-metrics` | Root/User | 倍率与性能 |
| Webhooks | 签名校验 | Stripe/Creem/Waffo… |
| Dashboard 兼容 | TokenAuth | `/dashboard/billing/*`、`/v1/dashboard/billing/*` |

---

## 6. Relay 面（热路径 · 留 Go）

| 前缀 | 说明 |
|------|------|
| `/v1/*` | OpenAI 兼容 chat/completions/responses/images/audio/… |
| `/v1/messages` | Claude |
| `/v1/realtime` | WebSocket |
| `/v1beta/*` | Gemini |
| `/pg/chat/completions` | Playground（UserAuth + Distribute） |
| `/mj` `/:mode/mj` | Midjourney |
| `/suno/*` | Suno 任务 |
| `/v1/video*` `/kling/*` `/jimeng` | 视频任务（`video-router`） |

中间件典型链：`TokenAuth` → `ModelRequestRateLimit` → `Distribute` → `controller.Relay` / `RelayTask`。

**Phase1 禁止**改适配器行为与计费结算语义。

---

## 7. 与新面板（WP-V）的边界

- 面板只调 **management** `/api/*` + 探活；**不**消费 relay 作为控制台契约（Playground 后续另议）。
- 会话：同域 cookie；推荐 `FRONTEND_MODE=disabled` + Nginx 同域反代（ADR-0001）。
- Console 子集契约：`CONSOLE_API_CONTRACT.md`。

---

## 8. Phase2 缝（仅指向）

AI-Core 出站形状见 `PHASE2_AI_CORE_SEAM.md`。热路径与预扣**不**同步阻塞 AI-Core。

---

## 9. 复现路由清单

```bash
# 仓库根
python scripts/scan_gin_routes.py
# → docs/gateway/_route_scan_raw.json
```

人工整理表：`docs/gateway/ROUTE_TABLE.md`。  
OpenAPI 对齐：`docs/gateway/OPENAPI_AUDIT.md`。

---

## 10. 明确不在本模块边界内

- `web/default` / `web/classic` 功能演进（LEGACY）
- `migrations/` 工具链（WP-S）
- Python AI-Core 实现
- Electron 壳行为

# 路由注册表（按 plane · Phase1 G1）

**真源**：`router/*.go` 静态阅读 + `scripts/scan_gin_routes.py`（片段级，非运行时 `Routes()`）。  
**生成日**：2026-07-22 · HEAD 实现分支 `xvyimu/th-2`。

扫描片段计数：

| 文件 | 片段数 |
|------|--------|
| `api-router.go` | 224 |
| `channel-router.go` | 41 |
| `authz-router.go` | 1 |
| `dashboard.go` | 4 |
| `relay-router.go` | 52 |
| `video-router.go` | 11 |
| `main.go`（probes） | 3 |
| `web-router.go` | 0（静态/NoRoute，无 REST 片段） |
| **合计** | **336** |

原始 JSON：`docs/gateway/_route_scan_raw.json`（可再生成，勿手改当 SSOT）。

---

## A. 全 plane 探活

| Method | Path | Plane |
|--------|------|-------|
| GET | `/healthz` | all, management, relay |
| GET | `/livez` | all, management, relay |
| GET | `/readyz` | all, management, relay |

---

## B. Management plane（`APP_PLANE=management|all`）

### B.1 Console 子集（P0 · 契约钉死见 CONSOLE_API_CONTRACT）

| Method | Full path | Auth / 中间件 |
|--------|-----------|----------------|
| GET | `/api/status` | 公开；`GlobalAPIRateLimit` |
| POST | `/api/user/login` | Critical + Turnstile + body limit |
| GET | `/api/user/logout` | 无强制登录 |
| GET | `/api/user/self` | `UserAuth` |
| GET | `/healthz` `/livez` `/readyz` | 无 |

### B.2 其它管理前缀（摘要 · 完整片段见扫描 JSON）

挂载：`SetApiRouter` → Group `/api`。

| 组 | 代表路径 | Auth |
|----|----------|------|
| setup/status | `GET/POST /api/setup` `GET /api/status` `POST /api/rum` | 混合 |
| user | `/api/user/register` `login` `self` admin CRUD | 混合 |
| oauth | `/api/oauth/:provider` wechat/telegram | Critical |
| subscription | `/api/subscription/*` admin | User/Admin |
| option | `/api/option/*` | Root |
| channel | `/api/channel/*`（见 permission 表） | Admin+authz |
| authz | `GET /api/authz/catalog` | Admin |
| token | `/api/token/*` | User |
| usage | `GET /api/usage/token/` | TokenAuthReadOnly |
| redemption | `/api/redemption/*` | Admin |
| log | `/api/log/*` | Admin/User |
| system-task | `/api/system-task/*` | Root |
| system-info | `/api/system-info/*` | Root |
| data | `/api/data/*` | Admin/User |
| group / prefill | `/api/group/` `/api/prefill_group/` | Admin |
| mj/task（管理查询） | `/api/mj/*` `/api/task/*` | User/Admin |
| vendors / models | `/api/vendors/*` `/api/models/*` | Admin |
| performance / ratio_sync | `/api/performance/*` `/api/ratio_sync/*` | Root |
| webhooks | `/api/stripe|creem|waffo/webhook` | 签名 |

### B.3 Channel 权限路由（完整 · `channel-router.go`）

Base：`/api/channel` + `AdminAuth`；另：`POST /:id/key` 需 Root + SecureVerification。

| Method | Path（相对 `/api/channel`） | Permission |
|--------|----------------------------|------------|
| GET | `/` | ChannelRead |
| GET | `/search` | ChannelRead |
| GET | `/models` | ChannelRead |
| GET | `/models_enabled` | ChannelRead |
| GET | `/ops` | ChannelRead |
| GET | `/:id` | ChannelRead |
| GET | `/test` `/test/:id` | ChannelOperate |
| GET | `/update_balance` `/update_balance/:id` | ChannelOperate |
| POST | `/` | ChannelSensitiveWrite |
| PUT | `/` | ChannelWrite |
| POST | `/status/batch` `/:id/status` | ChannelOperate |
| DELETE | `/disabled` `/:id` | ChannelSensitiveWrite |
| POST | `/tag/disabled` `/tag/enabled` | ChannelOperate |
| PUT | `/tag` | ChannelWrite |
| POST | `/batch` `/fix` `/copy/:id` … | 见源码表 |
| … | ollama / codex / upstream_updates / multi_key | 见源码 |

### B.4 Dashboard 兼容（TokenAuth）

| Method | Path |
|--------|------|
| GET | `/dashboard/billing/subscription` |
| GET | `/v1/dashboard/billing/subscription` |
| GET | `/dashboard/billing/usage` |
| GET | `/v1/dashboard/billing/usage` |

> 注意：`/v1/dashboard/*` 在 **management** 面；与 relay 的 `/v1/chat/*` 不同组。

---

## C. Relay plane（`APP_PLANE=relay|all`）

### C.1 模型列表

| Method | Path | Auth |
|--------|------|------|
| GET | `/v1/models` | TokenAuth（按头分流 OpenAI/Claude/Gemini） |
| GET | `/v1/models/:model` | TokenAuth |
| GET | `/v1beta/models` | TokenAuth |
| GET | `/v1beta/openai/models` | TokenAuth |

### C.2 Playground

| Method | Path | Auth |
|--------|------|------|
| POST | `/pg/chat/completions` | UserAuth + Distribute |

### C.3 OpenAI / Claude / 多模态（`/v1` + TokenAuth + ModelRL + Distribute）

| Method | Path |
|--------|------|
| GET | `/v1/realtime`（WS） |
| POST | `/v1/messages` |
| POST | `/v1/completions` `/v1/chat/completions` |
| POST | `/v1/responses` `/v1/responses/compact` |
| POST | `/v1/edits` `/v1/images/generations` `/v1/images/edits` |
| POST | `/v1/embeddings` `/v1/audio/*` `/v1/rerank` |
| POST | `/v1/engines/:model/embeddings` `/v1/models/*path` |
| POST | `/v1/moderations` |
| * | 若干 `RelayNotImplemented`：files / fine-tunes / variations… |

### C.4 MJ / Suno / Gemini generate

| Method | Path |
|--------|------|
| * | `/mj/*`、`/:mode/mj/*`（`registerMjRouterGroup`） |
| POST/GET | `/suno/submit/:action` `/suno/fetch` `/suno/fetch/:id` |
| POST | `/v1beta/models/*path` |

### C.5 Video（`SetVideoRouter`）

| Method | Path |
|--------|------|
| GET | `/v1/videos/:task_id/content` | TokenOrUserAuth |
| POST/GET | `/v1/video/generations` `.../:task_id` |
| POST | `/v1/videos` `/v1/videos/:video_id/remix` |
| GET | `/v1/videos/:task_id` |
| POST/GET | `/kling/v1/videos/text2video` `image2video` (+ task_id) |
| POST | `/jimeng/` |

---

## D. Frontend 交付（非 REST 表）

| `FRONTEND_MODE` | 行为 |
|-----------------|------|
| `auto` / `embedded` | `SetWebRouter` 嵌入双主题 |
| `redirect` | NoRoute 301 → `FRONTEND_BASE_URL` |
| `disabled` | 无 SPA；纯 API（**Vue 反代推荐**） |

`APP_PLANE=relay` 时不注册 frontend。

---

## E. 维护

1. 改路由后重跑 `python scripts/scan_gin_routes.py`。  
2. 更新本表「摘要」或 Console 子集行。  
3. 同步 `OPENAPI_AUDIT.md` / `console-subset`（G2/G3）。

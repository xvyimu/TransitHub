# RUN_MODE × APP_PLANE × FRONTEND_MODE 矩阵（Phase1 WP-G G4）

| 字段 | 值 |
|------|-----|
| 日期 | 2026-07-22 |
| 代码 | `runtime_mode.go` · `router/main.go` · `docs/operations/runtime-separation.md` |
| ADR | `docs/adr/0001-frontend-backend-delivery-seam.md` |

---

## 1. RUN_MODE（进程职责）

| Mode | HTTP | Worker 领任务 | Scheduler 建定时任务 | 迁移后退出 |
|------|------|---------------|----------------------|------------|
| `all`（默认） | ✅ | ✅ | ✅ | 否 |
| `serve` | ✅ | 否 | 否 | 否 |
| `worker` | 否 | ✅ | 否 | 否 |
| `scheduler` | 否 | 否 | ✅ | 否 |
| `migrate` | 否 | 否 | 否 | 是 |

约束：`worker` / `scheduler` / `migrate` 要求 `NODE_TYPE` 非 slave（见 `parseRuntimeConfig`）。

---

## 2. APP_PLANE（HTTP 路由裁剪）

仅当 `RUN_MODE` 提供 HTTP（`all`/`serve`）时有意义。

| Plane | 路由 |
|-------|------|
| `all` | Relay + management API + dashboard + frontend（按 FRONTEND_MODE） |
| `relay` | Relay + video + probes；**无** `/api` 管理、无 SPA |
| `management` | `/api` + dashboard + frontend + probes；**无** relay 热路径 |

Probes 始终：`/healthz` `/livez` `/readyz`（响应含 `plane`）。

---

## 3. FRONTEND_MODE（与 plane 正交）

| Mode | 行为 |
|------|------|
| `auto`（默认） | 兼容：非 master 且设 `FRONTEND_BASE_URL` → redirect；否则 embed |
| `embedded` | 强制嵌入双主题；缺资源则启动失败 |
| `redirect` | 非 API 路径 301 到 `FRONTEND_BASE_URL`（绝对 origin） |
| `disabled` | 纯 API；未知路径 Gin 404 — **Vue separated 推荐** |

构建：`frontend_external` tag → 无 embed 资源，需 `disabled` 或 `redirect`。

`APP_PLANE=relay` 时**不**走 frontend 注册（在 `SetRouterForPlane` 中 early return）。

---

## 4. 推荐组合

| 场景 | RUN_MODE | APP_PLANE | FRONTEND_MODE | 说明 |
|------|----------|-----------|---------------|------|
| 本地单进程开发 | `all` | `all` | `auto`/`embedded` | 旧体验 |
| **Vue 绞杀后端** | `serve` | `management` | **`disabled`** | Nginx 同域挂 `web-console` |
| Relay 水平扩展 | `serve` | `relay` | （忽略） | 仅热路径 |
| 集成镜像回滚 | `all` | `all` | `auto` | 配置级回 React embed |
| 迁移 Job | `migrate` | — | — | 先于切流量（WP-S） |
| 后台任务 | `worker` / `scheduler` | — | — | master 节点 |

Separated 拓扑（摘自 runtime-separation）：

```text
browser → Nginx:8080
            ├─ SPA (web-console or React dist)
            └─ /api /v1 /v1beta /mj /pg /suno /kling /jimeng
               /healthz /livez /readyz  →  Go:3000 (FRONTEND_MODE=disabled)
```

`/metrics`：**不要**在公网 console origin 反代。

---

## 5. 冒烟验收

| # | 操作 | 期望 |
|---|------|------|
| 1 | `APP_PLANE=management` 访问 `GET /v1/models` | 404 / 未注册 |
| 2 | `APP_PLANE=relay` 访问 `GET /api/status` | 404 / 未注册 |
| 3 | 任一 HTTP plane `GET /healthz` | 200 且 `plane` 正确 |
| 4 | management + disabled + 反代后浏览器登录 | cookie 写在同 host |
| 5 | 回滚 unset FRONTEND_MODE 或 auto + 集成二进制 | 旧 SPA 可开 |

---

## 6. 回滚

1. 配置：`RUN_MODE=all` `APP_PLANE=all`，清空或 `FRONTEND_MODE=auto`。  
2. 产物：部署非 `frontend_external` 集成镜像。  
3. 无需 DB down-migration（本矩阵不改 schema）。

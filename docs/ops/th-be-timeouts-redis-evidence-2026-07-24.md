# M-TH-be-timeouts-redis · timeouts + Redis hot-path evidence · **D7 NOT EXECUTED**

> 模块：`M-TH-be-timeouts-redis` · 只读测绘 `common/` `model/` `middleware/` + 超时 env 锚点 · 可小 docs 修。  
> **不做**：`go.mod` major bump · 生产 DSN · 改业务代码 · push 默认支 · 假绿 · D7 flip。  
> 先读：`AGENTS.md` · scout 文件 `docs/ops/th-backend-stable-scout-evidence-2026-07-24.md` **本 worktree 不存在**（记 gap，不挡本模块）。  
> 相关：[`w1-gin-redis-spike.md`](./w1-gin-redis-spike.md)（Gin/redis major 仍 defer）。

| Field | Value |
|-------|--------|
| Worktree | `C:\Users\yuanjia\orca\workspaces\src\th-be-timeouts-redis` |
| Branch | `xvyimu/th-be-timeouts-redis` |
| Base tip (pre-commit) | `f7a8b9bd` (docs(ops): TH E2E operator card) |
| Date | **2026-07-24** |
| Scope | env 超时/连接池表 · Redis fail-closed/open 分类 · 相关 `go test` · 风险建议 |
| D7 | **NOT EXECUTED** |

---

## 1. 超时 / 连接池表（env · 默认 · 代码锚点）

单位：除特别标注外均为 **秒** 或 **计数**。`GetEnvOrDefault` 实现见 `common/env.go`。

### 1.1 入站 HTTP server（`main.newHTTPServer`）

| Env | Default | Unit | Anchor | Notes |
|-----|---------|------|--------|-------|
| `HTTP_READ_HEADER_TIMEOUT_SECONDS` | `10` | s | `main.go:333` | 慢客户端 header 防护 |
| `HTTP_IDLE_TIMEOUT_SECONDS` | `120` | s | `main.go:334` | keep-alive idle |
| `HTTP_MAX_HEADER_BYTES` | `1048576` (1<<20) | bytes | `main.go:335` | 非超时，同表便于运维 |
| `SHUTDOWN_TIMEOUT_SECONDS` | `120` | s | `main.go:310` | SSE 友好 graceful shutdown |
| `READINESS_TIMEOUT_SECONDS` | `3`（≤0 钳到 3） | s | `controller/health.go:29-32` | readiness 总预算（DB+可选 Redis） |

**刻意未设：** 全局 `WriteTimeout` / `ReadTimeout`（`.env.example` 注释：流式响应不设全局 WriteTimeout）。

测试：`main_server_test.go`（defaults + overrides）。本 worktree 根包 `go test .` 因 `frontend_assets_embedded.go` 缺 `web/classic/dist` **setup failed**（见 §3）——与超时逻辑无关。

### 1.2 出站 Relay HTTP transport / client

| Env | Default | Unit | Anchor | Notes |
|-----|---------|------|--------|-------|
| `RELAY_TIMEOUT` | `0` | s | `common/init.go:117` · `common/constants.go:186` | **0 = 不限制**。**不得**挂到主 relay `http.Client.Timeout`（会切断长流）；见 `service/http_client.go:63-70`。例外：AWS invoke ctx `relay/channel/aws/relay-aws.go:43-47`；SSRF protected client `service/protected_fetch_client.go:94-96` |
| `RELAY_IDLE_CONN_TIMEOUT` | `90` | s | `common/init.go:118` · `common/http_client.go:30` | Transport idle |
| `RELAY_DIAL_TIMEOUT` | `10` | s | `common/init.go:119` · `common/http_client.go:19` | `net.Dialer.Timeout` |
| `RELAY_TLS_HANDSHAKE_TIMEOUT` | `10` | s | `common/init.go:120` · `common/http_client.go:31` | |
| `RELAY_RESPONSE_HEADER_TIMEOUT` | `120` | s | `common/init.go:121` · `common/http_client.go:32` | **上游无响应头**硬界；流式 body 仍靠 ctx / streaming |
| `RELAY_EXPECT_CONTINUE_TIMEOUT` | `1` | s | `common/init.go:122` · `common/http_client.go:33` | |
| `RELAY_MAX_IDLE_CONNS` | `500` | n | `common/init.go:123` · `common/http_client.go:28` | 池规模 |
| `RELAY_MAX_IDLE_CONNS_PER_HOST` | `100` | n | `common/init.go:124` · `common/http_client.go:29` | 每 host 池 |

装配入口：`service.InitHttpClient` → `common.NewOutboundHTTPTransport`（`service/http_client.go:56-60` · `common/http_client.go:16-38`）。

测试：`common/http_client_test.go`（生命周期字段 / header 超时 / cancel）。

### 1.3 流式 / 扫描 / 任务

| Env | Default | Unit | Anchor | Notes |
|-----|---------|------|--------|-------|
| `STREAMING_TIMEOUT` | `300` | s | `common/init.go:150` · `constant/env.go` · `relay/helper/stream_scanner.go:88-93` | **无数据** ticker 切断流；空补全可调大 |
| `STREAM_SCANNER_MAX_BUFFER_MB` | `128` | MB | `common/init.go:153` | 扫描缓冲上界（非时间） |
| `TASK_TIMEOUT_MINUTES` | `1440` | min | `common/init.go:173` | 异步任务超时退款；`0` 禁用 |
| `TASK_QUERY_LIMIT` | `1000` | n | `common/init.go:171` | 轮询批量 |

### 1.4 SQL 连接池

| Env | Default | Unit | Anchor | Notes |
|-----|---------|------|--------|-------|
| `SQL_MAX_IDLE_CONNS` | MySQL/PG **100** · SQLite **2** | n | `model/main.go:65-79` · `configureConnectionPool` | SQLite 大池易 OOM(14) |
| `SQL_MAX_OPEN_CONNS` | MySQL/PG **1000** · SQLite **4** | n | 同上 | |
| `SQL_MAX_LIFETIME` | `60` | s | 同上 | `sql.DB.SetConnMaxLifetime` |
| `SQLITE_PATH` | `one-api.db?_busy_timeout=30000` | path+DSN | `common/database.go:44` · `common/init.go:67-76` | env 覆盖若缺 `_busy_timeout=` 会补 **30000ms** |

测试：`model/main_pool_test.go`（方言默认 + 覆盖）。

### 1.5 Redis 连接 / 同步

| Env | Default | Unit | Anchor | Notes |
|-----|---------|------|--------|-------|
| `REDIS_CONN_STRING` | *(empty)* | URL | `common/redis.go:25-40` | 空 → `RedisEnabled=false`，**不** Fatal |
| `REDIS_POOL_SIZE` | `10` | n | `common/redis.go:39` | `opt.PoolSize` |
| `SYNC_FREQUENCY` | `60` | s | `common/init.go:115` · `common/redis.go:19-21,30-32` | 缓存 TTL 秒数（`RedisKeyCacheSeconds`） |
| *(hardcoded)* | `5` | s | `common/redis.go:42-45` | 启动 `Ping` 超时；**无 env** |

### 1.6 旁路 HTTP（模型同步等 · 非主 relay 热路径）

| Env | Default | Unit | Anchor | Notes |
|-----|---------|------|--------|-------|
| `SYNC_HTTP_TIMEOUT_SECONDS` | `10` 或 `15`（按调用点） | s | `controller/model_sync.go:93,303,501` | 同步上游模型列表 |
| `SYNC_HTTP_RETRY` | `3` | n | `controller/model_sync.go:135` | |
| `SYNC_HTTP_MAX_MB` | `10` | MB | `controller/model_sync.go:140` | 响应体上界 |

### 1.7 分层语义（运维一眼）

```text
Client ──► HTTP_READ_HEADER / IDLE ──► Gin
                │
                ▼
         Relay out: DIAL / TLS / RESPONSE_HEADER / IDLE_CONN
                │  (Client.Timeout 默认 0；RELAY_TIMEOUT 不挂主 client)
                ▼
         Stream body: STREAMING_TIMEOUT ticker + request ctx cancel
                │
         Shutdown: SHUTDOWN_TIMEOUT_SECONDS (SSE 收尾)
```

---

## 2. Redis 热点：fail-closed vs fail-open

### 2.1 分类约定

| Class | 含义 |
|-------|------|
| **Fail-closed (process)** | 错误 → 进程退出 / Fatal，拒绝带病启动 |
| **Fail-closed (request)** | Redis 运行时错误 → 该请求 5xx/拒绝，不降级放行 |
| **Fail-open → DB / memory** | Redis miss/error → 回落 DB 或内存实现，业务继续 |
| **Best-effort** | 写缓存失败只打日志，DB 已成功则主路径不回滚 |
| **Disabled-soft** | `REDIS_CONN_STRING` 空：`RedisEnabled=false`，走无 Redis 路径 |

### 2.2 启动与健康

| Hotspot | Class | Anchor | Behavior |
|---------|-------|--------|----------|
| 无 `REDIS_CONN_STRING` | Disabled-soft | `common/redis.go:25-28` | `RedisEnabled=false`，继续启动 |
| ParseURL 失败 | Fail-closed (process) | `common/redis.go:36-37` | `FatalLog` |
| 启动 `Ping` 失败 | Fail-closed (process) | `common/redis.go:45-47` | `FatalLog`（已配置 Redis 则必须可达） |
| Readiness Redis | Fail-closed (probe) | `controller/health.go:36-49` | 仅当 `RedisEnabled`：Ping 失败 → 503 `component=redis` |
| `main` 强制 memory cache | Side-effect | `main.go:77-79` | Redis 开则 `MemoryCacheEnabled=true`（兼容旧行为） |

### 2.3 限流 / 防护（运行时）

| Hotspot | Class | Anchor | Behavior |
|---------|-------|--------|----------|
| 全局限流 factory 选型 | Fail-open → memory *at init* | `middleware/rate-limit.go:76-87` | `RedisEnabled` 决定 redis vs memory 实现；**选定 redis 后** |
| IP / user Redis 限流读失败 | Fail-closed (request) | `middleware/rate-limit.go:26-30,159-163` | `LLen` err → **500** abort |
| 模型请求限流 Redis 检查失败 | Fail-closed (request) | `middleware/model-rate-limit.go:87-90,110-113` | `rate_limit_check_failed` 500 |
| 模型限流选型 | Fail-open → memory *at call* | `middleware/model-rate-limit.go:194-198` | 未开 Redis 用 memory |
| 邮箱验证限流 | Fail-open → memory | `middleware/email-verification-rate-limit.go:74-79` | |
| 通知限流 | Fail-open → memory；Redis 读硬错 fail-closed 返回 error | `service/notify-limit.go:50-84` | `RedisGet` 非 nil 错 → error；禁用 Redis → memory |
| Lua 令牌桶 | Fail-closed (request path) | `common/limiter/limiter.go:56-67` | `EvalSha` err → error 上抛；ScriptLoad 失败仅日志（后续 Eval 会炸） |

### 2.4 Auth / token / user 缓存

| Hotspot | Class | Anchor | Behavior |
|---------|-------|--------|----------|
| Token by key | Fail-open → DB | `model/token.go:273-280` | cache miss/err 不返回，落 DB |
| User cache `GetUserCache` | Fail-open → DB | `model/user_cache.go:118-129` | Redis 失败 → `GetUserById` |
| `GetUserQuota` / Group / Setting | Fail-open → DB | `model/user.go:1046-1052` 等 | 同上 |
| Session 用户刷新 | Fail-open → DB（经 GetUserCache） | `middleware/auth.go:55-68` | `RDB==nil && DB==nil` 才 fail-closed（单测） |
| Token/user 缓存 **写** | Best-effort | `model/token.go:386-392,420-426` 等 | `gopool` + `SysLog` |
| 退款后 token 缓存 incr | Best-effort | `model/refund_intent.go:271-277` | async log only |
| Invalidate token/user cache | Best-effort / no-op if disabled | `model/token.go:503-506` · `user_cache.go:56-58` | 禁用 Redis 直接 nil |
| HybridCache（affinity/subscription） | Fail-open → memory when redis off；redis on 时 Get 错上抛 | `pkg/cachex/hybrid_cache.go:59-108` · affinity `service/channel_affinity.go:98-99` | 2s op timeout |
| Adaptive metrics / channel metrics | Best-effort skip | `service/channel_metrics.go` · `main.go:80-94` | 无 Redis 不写 |

### 2.5 原始 Redis helper 风险点

| Helper | Timeout | Class | Note |
|--------|---------|-------|------|
| `RedisSet/Get/Del/H*` | **`context.Background()` 无 deadline** | hang risk | `common/redis.go:64+` — 网络半开可挂死调用方 |
| HybridCache ops | 2s / scan 30s / del 10s | bounded | `pkg/cachex/hybrid_cache.go:14-17` |
| Channel affinity clear | 2s | bounded | `service/channel_affinity.go:213` |

### 2.6 速查矩阵

| 场景 | Redis 未配置 | Redis 已配但运行时挂 |
|------|--------------|----------------------|
| 进程启动 | 继续（Disabled-soft） | 启动 Ping 已 Fatal；运行中依赖路径见下 |
| 全局限流 | memory | 请求 **500**（fail-closed） |
| 模型限流 | memory | 请求 **500** |
| Token/User 读缓存 | 直 DB | 回落 DB（fail-open） |
| Quota 扣减 | 直 DB | DB 同步扣；缓存 async best-effort |
| Readiness | 不查 Redis | **503** redis component |
| 缓存失效 | no-op | 失败打日志；可能短暂脏读 |

---

## 3. 测试（本 worktree 实测）

命令与 exit：

```text
go test ./common/ -run 'Timeout|HTTP|Relay|Redis|Outbound' -count=1
→ ok  github.com/xvyimu/TransitHub/common  ~2.7–3.5s
→ EXIT=0
  PASS: TestNewOutboundHTTPTransportUsesLifecycleTimeouts
  PASS: TestOutboundTransportTimesOutWaitingForResponseHeaders
  PASS: TestOutboundTransportHonorsRequestCancellation
  (+ TestInitSessionCookieSettingsRequiresHTTPSURL 匹配 HTTP 字样)

go test ./model/ -run 'Pool|SQL|Redis|Connection' -count=1
→ ok  github.com/xvyimu/TransitHub/model  ~2.2–2.5s
→ EXIT=0
  PASS: TestConnectionPoolConfigPreservesExistingDefaults
  PASS: TestConnectionPoolConfigUsesSQLiteSafeDefaults
  PASS: TestConnectionPoolConfigHonorsExplicitOverrides
  PASS: TestHardDeleteUserPurgesAuthenticationDataWhenRedisFails  (Redis 失败仍清鉴权数据)

go test . -run 'HTTP|Server|Timeout' -count=1
→ FAIL [setup failed]
→ EXIT=1
  frontend_assets_embedded.go:20:12: pattern web/classic/dist: no matching files found
  （嵌入前端资源缺失；非超时逻辑回归。main_server_test.go 源码存在但本包无法编译。）
```

**诚实结论：** 超时 transport + SQL 池单测 **绿（exit 0）**。根包 server 超时单测 **本 wt 未跑通**（frontend embed 构建前置），不声称全绿。

---

## 4. 风险 + 建议（不强制改代码）

| # | Risk | Severity | Suggestion |
|---|------|----------|------------|
| R1 | `RELAY_TIMEOUT=0` + 仅靠 `RESPONSE_HEADER_TIMEOUT`：header 已回但 body 僵死依赖 `STREAMING_TIMEOUT` | Med | 运维文档固定：流式盯 `STREAMING_TIMEOUT`；非流/AWS 才考虑 `RELAY_TIMEOUT>0` |
| R2 | `common.Redis*` 无 per-op timeout | **High** | 后续小改：统一 `context.WithTimeout`（如 2s，与 cachex 对齐）；**本模块不改代码** |
| R3 | 限流路径 Redis 瞬时故障 → 全局限流 **500**（fail-closed） | Med | 可接受偏安全；若要可用性，需显式产品决策「限流 fail-open」并防击穿 |
| R4 | 启动要求 Redis 可达（已配时 Fatal） | Low–Med | 多实例滚动注意顺序；readiness 已覆盖运行中探测 |
| R5 | Token 缓存异步与 DB 扣费：缓存 stale 窗口 | Med | 已有 DB floor；多实例务必开 Redis；禁用 Redis 时 token 状态写回路径不同（`token.go:207+`） |
| R6 | SQLite 默认小池 vs MySQL 默认大池同一 env 名 | Med | 文档已写；SQLite 误设 `SQL_MAX_OPEN_CONNS=1000` 仍危险 |
| R7 | `go-redis/v8` EOL · major defer | Low (hygiene) | 见 `w1-gin-redis-spike.md`；独立 wt，禁止与 D7 同 PR |
| R8 | 根包测试依赖 `web/classic/dist` | Low (CI) | CI/本地先 build classic 或 tags；勿当超时回归红 |

**运维默认建议（非强制）：**

1. 生产多实例：**必须** `REDIS_CONN_STRING` + 合理 `REDIS_POOL_SIZE`（默认 10 偏低时可升）。  
2. 流式 LLM：保持 `RELAY_TIMEOUT=0`；调 `RELAY_RESPONSE_HEADER_TIMEOUT` / `STREAMING_TIMEOUT`。  
3. 入站：保留 `HTTP_READ_HEADER_TIMEOUT_SECONDS`；勿贸然加全局 WriteTimeout。  
4. SQLite 开发机：勿复制 MySQL 的 `SQL_MAX_*` 大数。  
5. Readiness：编排层依赖 `/ready` 时认清 Redis 挂 = 整实例 not-ready。

---

## 5. 边界与非目标核对

| Item | Status |
|------|--------|
| 超时/连接池表 | **Done** §1 |
| Redis fail-closed/open 表 | **Done** §2 |
| 相关 go test + exit | **Done** §3 |
| 风险建议（无强制代码改） | **Done** §4 |
| `go.mod` major bump | **Not done**（禁） |
| 生产 DSN | **Not used** |
| D7 flip / push 默认支 | **NOT EXECUTED** |
| Scout 先读文件 | **Missing in this wt**（记录 gap） |

---

## 6. 结束状态

```text
MODULE: M-TH-be-timeouts-redis
STATUS: DONE + in-review
ARTIFACT: docs/ops/th-be-timeouts-redis-evidence-2026-07-24.md
D7: NOT EXECUTED
RECEIPT: th-coord
```

# Phase2 · Go → AI-Core 出站缝（文档 only · WP-G G6）

| 字段 | 值 |
|------|-----|
| 日期 | 2026-07-22 |
| 状态 | **设计预留**；Phase1 **不实现** Python / 不新增出站客户端代码 |
| 依据 | 执行规格 §0：AI 热路径留 Go；Better-wins 无测量不搬 |

---

## 1. 目标与非目标

**目标（Phase2+）**

- 允许 Go 网关在**旁路**场景调用内部 AI-Core（评测、批跑、智能路由说明、非实时策略实验）。
- 面板与外部客户端**永不**直连 AI-Core。

**非目标（本文件 / Phase1）**

- 实现 `services/ai` / Python 进程。
- 将 `relay` 完成路径改为同步依赖 AI-Core。
- 改计费预扣/结算时序。

---

## 2. 信任边界

```text
[ web-console ] ──► [ Go management/relay ]
                         │
                         │  仅内网 / mTLS / 网络策略
                         ▼
                   [ AI-Core : Python ]
                         │
                         ▼
                      可选队列 / 对象存储
```

| 规则 | 说明 |
|------|------|
| 入口 | 只有 Go 服务账号或内网调用 AI-Core |
| 身份 | Go 已完成 User/Token 鉴权后，**透传** `user_id` / `token_id` / `request_id` 元数据；**不**转发浏览器 cookie |
| 密钥 | 上游厂商 key 仍只在 Go/DB；AI-Core 默认不持有渠道密钥（除非未来独立密钥仓） |
| 失败 | 旁路调用失败 **不得** 导致主请求计费不一致；默认 fail-open 跳过评测或记日志 |

---

## 3. 配置缝（建议名 · 未接线）

| 环境变量 | 含义 | 默认 |
|----------|------|------|
| `AI_CORE_BASE_URL` | 如 `http://ai-core.internal:8081` | 空 = **功能关闭** |
| `AI_CORE_TIMEOUT_MS` | 出站超时 | 建议 ≤ 3000（旁路） |
| `AI_CORE_AUTH_TOKEN` | 服务间 token | 空则仅靠网络策略 |

包位置建议（未来）：`service/aicore` 或 `pkg/aicore` 客户端；**禁止**从 `controller` 直打 HTTP。

---

## 4. 接口形状（草案 · 非实现）

### 4.1 健康

`GET {AI_CORE_BASE_URL}/healthz` → `200 {"status":"ok"}`

### 4.2 评测 / 批跑（示例）

`POST {AI_CORE_BASE_URL}/internal/v1/eval`

```json
{
  "request_id": "…",
  "user_id": 123,
  "task_type": "prompt_eval",
  "payload": { }
}
```

响应：`202` 接受异步 或 `200` 小结果；错误 `4xx/5xx` 由 Go 记录，不写负向配额。

### 4.3 与热路径关系

| 路径 | 是否调用 AI-Core |
|------|------------------|
| `/v1/chat/completions` 等 relay | **默认否**；仅可选异步 shadow（类似现有 adaptive shadow 日志） |
| management 触发的「评测任务」 | 是（Phase2 产品能力） |
| 计费 settle | **否** |

---

## 5. 可观测

- 出站指标建议：`newapi_aicore_requests_total{code}` / latency（Phase2 再加）。
- 日志必带 `request_id`；禁止把 prompt 全文默认打到 info。

---

## 6. 验收（Phase2 时）

- [ ] `AI_CORE_BASE_URL` 空时零出站  
- [ ] 面板抓包无 AI-Core host  
- [ ] 杀掉 AI-Core 时 relay 主路径仍 200（契约测试）  
- [ ] 鉴权透传无 cookie  

Phase1 仅本文件存在即过 G6。

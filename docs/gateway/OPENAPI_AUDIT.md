# OpenAPI 审计（Gin 真源 vs 文档 · Phase1 WP-G G2）

| 字段 | 值 |
|------|-----|
| 日期 | 2026-07-22 |
| 真源 | `router/*.go` 注册（见 `ROUTE_TABLE.md` / `_route_scan_raw.json`） |
| 文档 | `docs/openapi/api.json`（OpenAPI 3.0.1 · 后台管理 · **131 paths / 157 ops**） |
|  | `docs/openapi/relay.json`（AI 模型接口 · **35 paths / 38 ops**） |
| 方法 | 静态路径集合对比 + Console 子集人工核对；**非**运行时 `gin.Routes()` dump |
| 结论摘要 | 管理/Relay 文档**滞后于代码**；Console 核心登录路径在 api.json 中**存在但无响应 schema**；探活与大量新路由**未收录** |

---

## 1. 规模对照

| 面 | Gin（扫描片段 / 整理全路径） | OpenAPI ops | 关系 |
|----|------------------------------|-------------|------|
| Management `/api` + dashboard | api-router 224 + channel 41 + authz 1 + dashboard 4 片段 | api.json **157** ops | 文档明显偏少；且无 dashboard 四条 |
| Relay + video | ~49 条整理全路径（relay 52 + video 11 片段） | relay.json **38** ops | 缺 playground/suno/部分 models 变体 |
| Probes | 3 | **0** | 完全未收录 |

---

## 2. Console 子集（P0）核对

| Method Path | Gin | api.json / relay.json | 备注 |
|-------------|-----|----------------------|------|
| `GET /healthz` | ✅ | ❌ | 仅运行时；应进 console-subset |
| `GET /livez` | ✅ | ❌ | 同上 |
| `GET /readyz` | ✅ | ❌ | 同上 |
| `GET /api/status` | ✅ | ✅ path 存在 | 无详细 response schema |
| `POST /api/user/login` | ✅ | ✅ | 有 body username/password；无 Set-Cookie/2FA data 描述 |
| `GET /api/user/logout` | ✅ | ✅ | 无 body schema |
| `GET /api/user/self` | ✅ | ✅ | 无 data 字段表 |
| `GET /frontend-healthz` | 前端容器（separated） | ❌ | 非 Go 进程；可选 |

**判定**：Vue MVP **可以**按实机路径对接；**不能**只靠现有 OpenAPI 生成客户端类型。以 `CONSOLE_API_CONTRACT.md` 为准。

---

## 3. 管理面 · Gin 有 / OpenAPI 无（抽样 · 高相关）

下列路径在代码中注册，**api.json 无对应 path**（2026-07-22 检查）：

| Path / 区域 | 证据 |
|-------------|------|
| `POST /api/rum` | `api-router.go` RUM |
| `GET /api/perf-metrics` `.../summary` | UserAuth 性能摘要 |
| `GET /api/authz/catalog` | `authz-router.go` |
| `POST /api/waffo` / `waffo-pancake` webhooks 与多条支付 | 较新支付面 |
| `/api/system-task/*` `/api/system-info/*` | Root 系统任务 |
| `/api/subscription/*` 大部 | 订阅计费（若 json 无整组则全缺） |
| `/api/channel/ops` 及大量 channel 批处理/codex/ollama/upstream_updates | channel 表 41 vs openapi channel ~22 paths |
| channel `duplicates` / `merge`（若分支有） | ASIS 提及；以当前分支 `channel-router` 为准（本快照 permission 表无 merge 行则标 N/A） |
| dashboard billing 四条 | `dashboard.go`；api.json 未列 |

**OpenAPI 有、需与 Gin 再核对方法/弃用**

- 历史 oauth 细路径（`/api/oauth/github` 等）vs 现网统一 `GET /api/oauth/:provider`
- 部分 `/api/verify/*` 命名与 `POST /api/verify` 是否一致

---

## 4. Relay 面 · loose diff（`{var}` 归一）

**Gin 有、relay.json 缺（loose）**

| 项 |
|----|
| `GET /v1/models/{var}`（单模型 retrieve） |
| `GET /v1beta/openai/models` |
| `POST /pg/chat/completions` |
| `POST /v1/edits` |
| `POST /v1/images/variations`（NotImplemented 仍挂路由） |
| `POST /v1/models/{var}`（Gemini 兼容 path） |
| `POST /v1beta/models/{var}` |
| `POST /v1/videos/{var}/remix` |
| Suno：`POST /suno/submit/{var}` `POST /suno/fetch` `GET /suno/fetch/{var}` |
| 若干 files/fine-tunes NotImplemented 在 Gin 挂了、文档或有或无 |

**OpenAPI 有、Gin 形态不同**

| OpenAPI | Gin |
|---------|-----|
| `POST /v1beta/models/{model}:generateContent` | 实际多为 `POST /v1beta/models/*path` 通配 |

MJ 子路径：`registerMjRouterGroup` 细节未逐条进 relay.json 时一律标 **drift**。

---

## 5. 质量问题（双方）

| 问题 | 说明 |
|------|------|
| 响应 schema 空洞 | api.json 大量仅 `200 成功`，无法生成类型 |
| security 占位 | `Combination343` 等非标准、难映射到 UserAuth/TokenAuth |
| 双文件边界 | dashboard 的 `/v1/dashboard/*` 既非纯 relay 也不在 api 前缀清晰处 |
| 参数风格 | OpenAPI `{id}` vs Gin `:id` |
| 无 CI 门禁 | 改路由不强制更新 json |

---

## 6. Phase1 处理策略（Better-wins）

| 优先级 | 动作 | 不做 |
|--------|------|------|
| P0 | 钉死 **console-subset**（契约文档 + 可选 yaml） | 全量重生成 157+ 条 |
| P0 | 本审计表入库 | 改业务 handler |
| P1 | `scripts/openapi_route_diff.py` 只读 diff（G5） | 强制 CI fail 全量 |
| P2 | 按域补 OpenAPI（channel / subscription / system-task） | 为对齐而改路由语义 |

**停损**：全量人工对齐人日 ≫ 收益 → 保持「真源=Gin + 子集契约」，全量标 L2 技术债。

---

## 7. 复现

```bash
python scripts/scan_gin_routes.py
# 可选 G5：
python scripts/openapi_route_diff.py
```

附件（可选生成物）：`docs/gateway/_openapi_api_ops.txt`、`_openapi_relay_ops.txt`。

---

## 8. 签字清单

- [x] 规模对照  
- [x] Console 子集逐条  
- [x] 管理面缺失抽样  
- [x] Relay loose diff  
- [x] Phase1 策略与非目标  

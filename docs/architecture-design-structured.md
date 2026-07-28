# TransitHub · 结构化架构设计（v2）

| 项 | 值 |
|----|-----|
| 产品 | 自托管 LLM 网关 + 控制台（显示名 NewAPI） |
| GitHub | [xvyimu/TransitHub](https://github.com/xvyimu/TransitHub) |
| 路径 | `D:\projects\TransitHub\src` |
| 文档版本 | **v2 · 2026-07-29** |
| **栈权威** | **[`PROJECT.md`](./PROJECT.md)** · [`ARCHITECTURE_TARGET.md`](./ARCHITECTURE_TARGET.md) |
| 许可 | **AGPL-3.0** + NOTICE |

> 方法：[arc42](https://docs.arc42.org/home/) · [C4](https://c4model.com/) · 组合优化说明  

---

## 0. 五问

| # | 答 |
|---|----|
| 是什么？ | 多上游 LLM API 网关 + 运营后台 |
| 为谁？ | 自托管运维 / 平台开发 |
| 不做？ | 第二后端语言 · 无 D7 双写 UI · 文档默认真 flip 生产前端 |
| 验收？ | `go test` 相关包 · web-console 质量门 · 架构守卫 |
| 协作？ | 公有 · AGPL · Issue/PR |

---

## 1. 背景与目标

统一 40+ 上游、令牌/配额/中继/计费。管理台 **Vue strangler**：开发主路径 `web-console/`，生产默认 React `web/default/` 直至 **D7 人 gate**。

| 质量属性 | 表述 | 验证 |
|----------|------|------|
| 正确性 | 中继/配额路径可测 | go test |
| 兼容 | SQLite/MySQL/PG | migrations CI |
| 安全 | JSON 统一封装；鉴权 | common 守卫 · 审 |
| 可回滚 | 前端 ≤5 min 回滚面 | cutover-rollback 文档 |
| 可运营 | 渠道/日志/令牌 | console 契约 |

---

## 2. 总体架构（C4）

### Context

```text
 [API 客户端/SDK] ──token──► TransitHub
 [运营人员] ──浏览器──► Admin UI
                              │
                              ▼
                        上游 LLM 厂商
```

### Container

```text
 Go binary (Gin)
  router → middleware → controller → service → model
                         ↘ relay/channel → upstream

 Data: SQLite | MySQL | PG  +  Redis可选

 UI:
  web-console/ (Vue 开发主路径)
  web/default/ (React 生产默认/回滚) 直到 D7
  web/classic/ (L2 冻结)
```

---

## 3. 选型理由

| 选 | 因 | 不选 |
|----|----|------|
| Go 网关 | 长连接/性能/谱系 | Nest/FastAPI 平行 |
| 三库 | 部署面 | 单 SQLite 假兼容 |
| common JSON | 可守卫 | 业务直 marshal |
| Vue console | strangler | 无 ADR 全站 Next |
| React 至 D7 | 回滚 | 文档假 flip |

---

## 4. 核心模块与接口

| 模块 | 接口要点 |
|------|----------|
| controller | 管理 HTTP；JSON via common；类型可用 RawMessage |
| service/model | 业务与持久化 |
| relay/channel | 上游协议 |
| migrations | 三方言；空库基线 ≠ 生产 force |
| web-console | 契约驱动视图 |
| legacy-guard | 阻 legacy 新功能 |

---

## 5. 资产复用

new-api 谱系对照 + AGPL NOTICE；React 热修 only；classic 冻结；DB 备份后迁移。

---

## 6. 信任边界与风险

| 边界 | 风险 | 缓解 |
|------|------|------|
| 公网 API | 盗用 token | 鉴权·配额·审计日志 |
| 上游 | 数据出域 | 渠道配置·最小化日志 |
| Admin | 会话劫持 | 鉴权·HTTPS 部署纪律 |
| 前端切流 | 未门禁 flip | D7 人闸·G1–G8 |
| 迁移 | 毁库 | empty-DB 范围·备份 |

---

## 7. 14 天计划

| 日 | 主题 | DoD |
|----|------|-----|
| 1–2 | 文档 | 与 TARGET 一致 |
| 3–5 | JSON/controller | go test 绿 |
| 6–7 | console | 质量门 |
| 8–9 | relay 小步 | 测/笔记 |
| 10–11 | migrate runbook | CI job |
| 12 | deps 分诊 | 不盲升 |
| 13–14 | D7 差距表 | **不 flip** |

---

## 8. 验收命令（L4）

| 命令 | 用途 |
|------|------|
| `go test ./common ./controller ./dto ./model`（按面加减） | 后端 |
| `go test ./...` 子集 | 扩大回归 |
| web-console 仓内 test/typecheck | 新 UI |
| `scripts/check-architecture-guards.ps1` | JSON 守卫 |
| legacy-guard workflow | PR 门 |

---

## 9. 相关文档

`PROJECT.md` · `ARCHITECTURE_TARGET.md` · `ARCHITECTURE_ASIS.md` · `legacy-frontend-gate.md` · `ops/th-*-2026-07-28.md` · cutover-plan

---

*v2 · 2026-07-29*

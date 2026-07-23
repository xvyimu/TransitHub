# TH E2E · 操作员一页纸（2026-07-25）

> **给操作员**：插非生产凭据 → 跑 W4 pack → 读 exit 表。  
> **W4 exit 0 ≠ D7。** D7 仅在 G1–G8 齐 **且** 人类口令 `D7 flip 现在`（或等价）后执行。  
> **禁止**把真实 `TH_E2E_*` 密码写入 git / 本文件 / PR / issue。  
> 本脚本 **永不**改生产 `FRONTEND_MODE`。

相关：[`th-e2e-gate-card.md`](./th-e2e-gate-card.md) · [`w4-d7-nonprod-verify.md`](./w4-d7-nonprod-verify.md) · [`w2-cutover-e2e-credentials.md`](./w2-cutover-e2e-credentials.md) · 全量 failure catalog [`th-day-e2e-harness-2026-07-24.md`](./th-day-e2e-harness-2026-07-24.md)

---

## 1. 铸造非生产凭据（不写真密码）

| 步骤 | 做什么 | 完成标准 |
|------|--------|----------|
| 1 | 确认目标是 **local / staging**，不是生产 DSN / 生产 origin | `TH_API_BASE` 已口头约定 |
| 2 | 空库：完成 `POST /api/setup`（或产品 setup 向导），设 **强** root 密码 | setup 完成 |
| 3 | 共享 staging：建 **专用 e2e admin**（UI 或 admin API），角色含 AdminAuth + ChannelRead | 账号可登录 |
| 4 | 密码只进 **会话 env / CI secret / 本地密钥库** | git 无明文 |
| 5 | 会话导出（见下节）后跑 W4 pack | `login=0 channels=0` |

细则与常见坑（默认 root 失效、setup 占用首 admin）：[`w2-cutover-e2e-credentials.md`](./w2-cutover-e2e-credentials.md)。

**不要做：**

- 生产密码收集 / 生产 DSN 做 migrate dry-run
- 把 secret 贴进 markdown、commit message、agent 报告
- 用 `scripts/e2e-web-console-login.ps1` **判 G2 绿**（该脚本缺 env 时可能默认 `root`/`123456`，假绿风险）

---

## 2. 会话级环境变量（占位符）

```powershell
# 仓库根（本 wt 或 D:\TransitHub\src）
# 仅非生产 — 永不 commit 下列真实值

$env:TH_API_BASE = 'http://127.0.0.1:3000'   # 或 staging origin；默认即此
$env:TH_E2E_USER = '<non-prod-admin>'
$env:TH_E2E_PASS = '<non-prod-secret>'       # 勿记入 shell 历史以外的共享介质

# 快速复检（跳过重 build）
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild

# 首次 / 源码漂移后：去掉 Skip* 全量
# pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1
```

| 变量 | 用途 | 缺省 |
|------|------|------|
| `TH_API_BASE` | API 根（healthz / login / channels） | `http://127.0.0.1:3000` |
| `TH_E2E_USER` | G2 login + G3 channels RO | **无**（W4 不静默 root） |
| `TH_E2E_PASS` | 同上；与 USER **成对必填** | **无** |

空白 / 仅空白字符 → 视作未设 → exit **10**。

---

## 3. 入口脚本

| 角色 | 路径 | 判 G2/G3 绿？ |
|------|------|----------------|
| **G2/G3 编排器（唯一判绿）** | `scripts/w4-d7-nonprod-verify.ps1` | **是** — 无凭据 exit **10**，不默认 root |
| Legacy 登录 e2e | `scripts/e2e-web-console-login.ps1` | **否** — 默认可 `root`/`123456` |
| OpenAPI 合同（静态） | `scripts/validate-console-contract.py` | 仅 G3 **contract**（无 live） |

编排器步骤映射：healthz → `/api/status` → contract → login+self → channels RO（无 key）→ 可选 `web-console` build / `go build -tags frontend_external`。见 [`w4-d7-nonprod-verify.md`](./w4-d7-nonprod-verify.md)。

---

## 4. 期望 exit 表（`w4-d7-nonprod-verify.ps1`）

| Code | 含义 | 操作员动作 |
|------|------|------------|
| **0** | 所选步骤全绿（未 `-SkipAuth` 时含 login+channels） | 仍要 G4 镜像 · G6 soak · G7 回滚演练 · **G8 人 gate**；**≠ D7** |
| **1** | 登录失败（错密 / 封禁 / setup 未完） | 铸造/重置非生产 admin |
| **2** | login 后 self/session 失败 | 查 cookie / `New-Api-User` |
| **3** | 后端不可达（healthz） | 起非生产 API；查 `TH_API_BASE` |
| **4** | channels RO 失败或列表泄漏 key | Admin 角色 + ChannelRead；泄漏则提 bug |
| **5** | 合同校验或 build 失败 | 修 OpenAPI / `web-console` / Go build |
| **6** | `/api/status` 失败 | 后端路由 / 就绪 |
| **10** | **凭据未设** 或 `-SkipAuth` | 导出非生产 `TH_E2E_USER`+`TH_E2E_PASS`。**不是 pass** |

### 无凭据 dry（期望 10）

```powershell
# 确保本 shell 无 TH_E2E_*（或仅空白）
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
# 期望: SUMMARY exit=10  … login=10 channels=10
# healthz/status/contract 可绿，整体仍 10 — 禁止当 G2/G3 绿
```

| 条件 | 期望 |
|------|------|
| `TH_E2E_USER` / `TH_E2E_PASS` 任一空或仅空白 | exit **10** · `login=10 channels=10` |
| `-SkipAuth` | exit **10** · `login=skip channels=skip` |
| healthz/contract 绿 + 无 auth | **仍** exit **10** |

### 本会话 dry 记录（2026-07-25 · 无凭据）

| 项 | 值 |
|----|-----|
| 命令 | `pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild` |
| Exit | **10** |
| SUMMARY（脱敏） | `healthz=0 status=0 contract=0 login=10 channels=10 console_build=skip backend_build=skip` |
| BLOCK | `credentials incomplete — missing: TH_E2E_USER + TH_E2E_PASS` |
| D7 | **NOT EXECUTED** |
| push | **not done** |

---

## 5. W4 exit 0 ≠ D7（必读）

| 动作 | 是 D7？ |
|------|---------|
| 推/合 `docs/ops/*` · 本 operator card | **否** |
| W4 pack exit **0**（非生产） | **否** — 仅「可申请人 gate 议程」 |
| 改生产 `FRONTEND_MODE` / 切流量 | **是** — 需 G8 口令；本脚本 **永不**执行 |
| 人类口令 `D7 flip 现在`（或等价）且 G1–G8 齐 | **才是** D7 执行条件 |

**一句话状态模板：**

```
TH-E2E: blocked|green · D7 NOT flipped · no push
```

例：`W4 exit 10 · healthz=0 status=0 contract=0 login=10 channels=10`。

---

## 6. 解锁顺序（有凭据后）

1. 会话 export `TH_E2E_USER` + `TH_E2E_PASS`（非生产）
2. 快速：`…\w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild` → 要 `exit=0` · `login=0` · `channels=0`
3. 全量（无 Skip* build）再绿一次
4. **仍不是 D7**：G4 · G6 · G7 · **G8** 口令

---

## 7. 明确非目标

| 项 | 状态 |
|----|------|
| 生产 D7 flip | **NOT EXECUTED**（本卡不授权） |
| 生产 `FRONTEND_MODE` | 本 pack **不碰** |
| 缺凭据假绿 | **禁止**（exit 10） |
| 真密码进 git | **禁止** |
| 用 legacy login e2e 判 G2 | **禁止** |

# TH E2E gate card（一页纸）

> **G2/G3 live ≠ D7。** 推 docs / 合 PR / 本机 exit 0 **都不是** 生产 cutover。  
> D7 仅在 G1–G8 齐 **且** 人类口令 `D7 flip 现在`（或等价）后执行。  
> **禁止**把真实 `TH_E2E_*` 密码写入 git / 本文件。

## 入口脚本

| 角色 | 路径 | 判绿？ |
|------|------|--------|
| **G2/G3 编排器（唯一判绿）** | `scripts/w4-d7-nonprod-verify.ps1` | **是** — 无凭据 exit **10**，不静默 root |
| Legacy 登录 e2e | `scripts/e2e-web-console-login.ps1` | **否** — 默认可 `root`/`123456`；**勿用本脚本判 G2 绿** |
| OpenAPI 合同（静态） | `scripts/validate-console-contract.py` | G3 **contract** only（无 live） |
| 凭据铸造（操作员） | `docs/ops/w2-cutover-e2e-credentials.md` | — |
| 全量 failure catalog | `docs/ops/th-day-e2e-harness-2026-07-24.md` | — |
| Pack how-to | `docs/ops/w4-d7-nonprod-verify.md` | — |

```powershell
# 无凭据 preflight — 期望 exit 10（不是 pass）
pwsh -NoProfile -File scripts/w4-d7-nonprod-verify.ps1 -SkipConsoleBuild -SkipBackendBuild
```

## Exit 速查（W4 pack · 无假绿）

| Code | 含义 | 操作员动作 |
|------|------|------------|
| **0** | 所选步骤全绿（含 login+channels 若未 SkipAuth） | 仍要 G4 镜像 · G6 soak · G7 回滚演练 · **G8 人 gate**；**≠ D7** |
| **1** | 登录失败（错密 / 封禁 / setup 未完） | 铸造/重置非生产 admin；见 w2 credentials |
| **3** | 后端不可达（healthz） | 起非生产 API；查 `TH_API_BASE` |
| **5** | 合同校验或 build 失败 | 修 OpenAPI / `web-console` / Go build |
| **10** | **凭据未设** 或 `-SkipAuth` | 导出非生产 `TH_E2E_USER`+`TH_E2E_PASS`。**不是 pass** |

其它：`2` self/session · `4` channels RO / key 泄漏 · `6` `/api/status` — 见 harness catalog。

## 无凭据 → 期望 10

| 条件 | 期望 |
|------|------|
| `TH_E2E_USER` / `TH_E2E_PASS` 任一空或仅空白 | exit **10** · `login=10 channels=10` |
| `-SkipAuth` | exit **10** · `login=skip channels=skip` |
| healthz/contract 可绿 + 无 auth | **仍** exit **10** — **禁止**当 G2/G3 绿 |

## G2/G3 vs D7（分离）

| 动作 | 是 D7？ |
|------|---------|
| 推/合 `docs/ops/*` · gate card · harness | **否** |
| W4 pack exit **0**（非生产） | **否** — 仅「可申请人 gate 议程」 |
| 改生产 `FRONTEND_MODE` / 切流量 | **是** — 需 G8 口令；本脚本 **永不**执行 |
| 本会话默认 | **D7 FLIP: NOT EXECUTED** · **不 push**（除非人类另令） |

## 环境名（无值）

| 变量 | 用途 |
|------|------|
| `TH_API_BASE` | API 根；默认 `http://127.0.0.1:3000`（仅非生产） |
| `TH_E2E_USER` / `TH_E2E_PASS` | G2 login + G3 channels RO；**必填成对** |

## 一句话状态模板

```
TH-E2E: blocked|green · D7 NOT flipped · no push
```

Detail 例：`W4 exit 10 · healthz=0 status=0 contract=0 login=10 channels=10`。

# TransitHub · code-review findings 消化（2026-07-24）

源：`D:\orca\.planning\portfolio-stack-policy-2026-07-24\code-review\transithub-findings.md`  
**D7 NOT EXECUTED** · **无新代码 P0**

## P0

无新增代码 P0。运营 shadow 真切流禁关 ≠ 本波改 env。

## P1 → 动作

| id | 动作 | wt |
|----|------|-----|
| TH-CR-001 shadow keep | **不改代码** · 禁关 shadow · 采样另授 | 未开（ops 人） |
| TH-CR-002 三前端纪律 | 已有 W8 legacy scan + INTEGRATE | 继承 |
| TH-CR-003 refund 幂等 | **开 fix** | `th-cr-refund-idempotency-tests` |
| TH-CR-004 TLS insecure | 文档进 host-bind 包 | `th-cr-host-bind-docs` |
| TH-CR-005 HOST 绑定 | **开 fix docs** | `th-cr-host-bind-docs` |
| TH-CR-006 中继面 | AGENTS 纪律 · 无本波刀 | — |
| TH-CR-007 渠道 ops | 人工 · 无本波刀 | — |

## P2

TH-CR-010 G2 缺 `TH_E2E_*` → **honest blocked**（长波已 evidence）  
TH-CR-012 三库 → W9 已记 refund_intents 漂移  

## 并行

live ≤3：W11 a11y-debt-2 + 2× cr-fix。禁与关 shadow / D7 flip 并行。

# TransitHub · Long Wave · W12 INTEGRATE（**等人 · 不自动 flip**）

> **Date:** 2026-07-24  
> **Coord tip:** `xvyimu/th-coord` (see `git log -1`)  
> **D7 FLIP: NOT EXECUTED** · production `FRONTEND_MODE` **untouched**  
> **G0:** D = A+C non-prod evidence + backend stability  
> Companion: [GATE-MATRIX.md](./GATE-MATRIX.md) · [WEEK-BACKLOG.md](./WEEK-BACKLOG.md) · [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md)

## 一句话

本周交付 = **非生产 cutover 证据包（绿或诚实 blocked）+ 后端路径级报告**。  
**不是** 生产 Vue 切流。合入 docs/feature 支可以；**main 合入 / D7 需人**。

## 0. 队列验收（实现项）

| ID | Branch / tip | Status | Key exits / note |
|----|--------------|--------|------------------|
| W1a | `th-console-quality` · `4afcf5b3` | **ACCEPT** | pnpm ×4 **0** · W4 pack **10** |
| W1b | backend-stable-scout · `d1dd3278` | **ACCEPT** | pool/common **0** · migrate sqlite **0** |
| W2 | g2-e2e · `d1957b64` | **ACCEPT · blocked** | pack **10** · no `TH_E2E_*` |
| W3 | g3-channels · `4daf0ba9` | **ACCEPT** | contract **0** · live blocked |
| W4 | g4-image · `4c2560bf` | **ACCEPT · blocked local** | docker absent · **CI SSOT** |
| W5 | g5-regression · `d6e3dfae` | **ACCEPT · green** | `frontend_external` **0** · common/model **0** |
| W6 | g6-soak · `f4669be9` | **ACCEPT · blocked** | full soak not run |
| W7 | g7-rollback · `98ce2dfe` | **ACCEPT · blocked timed** | dry-run paths/help **0** |
| W8 | legacy-gate-scan · `98ddd6bd` | **ACCEPT** | branch empty vs main · historical 可疑 pre-gate |
| W9 | be-migrate-3db · `44ab1b5e` | **ACCEPT** | only **`refund_intents`** baseline drift · three-dialect **0** |
| W10 | be-timeouts-redis · `f640bd5d` | **ACCEPT** | common/model **0** · root embed **1** honest · Redis no-deadline risk |
| W11 | console-a11y-debt | **in-flight** (or harvest next) | — |
| W12 | this pack | **coord** | 等人 integrate / G8 |

## 1. GATE 签字态（诚实）

| Gate | Signable? | Status |
|------|-----------|--------|
| G1 | **YES green** | Module2 present |
| G2 | **blocked** | need `TH_E2E_*` · exit 10 recorded |
| G3 live | **blocked** | needs G2 · contract **green** |
| G4 local | **blocked** | CI job `image-reproducibility` SSOT |
| G5 | **YES green** | W5 re-verified |
| G6 | **blocked** | no 24h soak |
| G7 | **blocked** | no timed ≤5m drill |
| G8 | **blocked** | need human `D7 flip 现在` |

**Flip readiness: NO.**

## 2. 建议合入序（feature → main · **人批**）

1. **P0 evidence 批**（几乎纯 `docs/ops/*-evidence-*.md` + `th-long-wave/*`）：  
   cherry-pick / PR from `xvyimu/th-coord`（已含多数 harvest）或按模块支：  
   g2 · g3 · g4 · g5 · g6 · g7 · legacy · migrate-3db · timeouts-redis · console-quality · backend-scout  
2. **不要** 与生产 `FRONTEND_MODE` / 删 `web/default` 同 PR。  
3. **可选 follow-up 实现**（另开 wt · 非本包自动做）：  
   - `000002` add `refund_intents` (W9)  
   - Redis helper `context.WithTimeout` (W10 R2 High · 需单测)  
   - console unit/a11y 债 (W1a / W11)  
4. **禁止** 本包内 go.mod Gin/redis major · D7 · asar/CSP 生产改。

## 3. 环境仍阻塞（操作员）

| Need | Unblocks |
|------|----------|
| Non-prod `TH_E2E_USER` + `TH_E2E_PASS` | G2/G3 live green |
| Docker CLI **or** CI digest paste | G4 local or dossier paste |
| Staging owner ≥24h | G6 |
| Timed rollback on non-prod | G7 |
| Human phrase **`D7 flip 现在`** | G8 only after G1–G7 |

## 4. G8

仅 [G8-HUMAN-CHECKLIST.md](./G8-HUMAN-CHECKLIST.md)。总控 **永不** 自 flip。

## 5. Cross-product note (not TH work)

**Codexveil** `cv-coord`: W1–W12 **ACCEPT** · `cv-long-wave/INTEGRATE.md` = **READY_FOR_HUMAN_GATE** · live 0 · **勿停** 其总控；TH 不代合 CV main。

## 6. Coord stance

- 继续 7m：收 W11 → 更新本文件 §0 → feature push  
- **不** push `main`  
- **不** D7 / 生产 CSP / asar  
- findings 路径出现再开 fix wt  

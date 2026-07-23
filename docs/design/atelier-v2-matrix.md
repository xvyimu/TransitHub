# V2 · TransitHub web × Atelier 差异矩阵

**日期：** 2026-07-23  
**仓 tip 基线：** `3732bbc9`（实现前 main）  
**强度：** **A0/A1 only**（无 A2 霓虹 / 无 default-on glass / 无 dual-primary 换品牌）  
**SSOT：** `D:\orca\.planning\portfolio-visual-fluent-glass-2026-07-23\atelier-token-ssot.md`  
**Pattern：** Chronicle V1a · ChronoPortal V1b  
**栈：** React · TanStack Router · Tailwind v4 · shadcn/base-ui · rsbuild · `web/default`（**不换栈**）  
**范围：** **web/default only** · skip `web/classic`

---

## Design Read

> 高密度 API 管理台 · 默认蓝 primary + 多 preset 轴 · 保留 TransitHub 交互品牌 · 对齐 Atelier **半径 4/8 · chrome blur ≤12 · 间距纪律 · 克制顶轨**，非 MindSync glass-shell 克隆。

| Dial     | 值                                    |
| -------- | ------------------------------------- |
| VARIANCE | 4–5                                   |
| MOTION   | 3（沿用现有 + reduced-motion）        |
| DENSITY  | 5（表格/侧栏实心；禁大面积霜）        |

---

## 1. 现状 vs 目标

| 维           | TransitHub 现状 (`web/default`)              | Atelier SSOT          | V2 裁决                                                                 |
| ------------ | -------------------------------------------- | --------------------- | ----------------------------------------------------------------------- |
| 品牌主色     | `--primary` 蓝 oklch + multi-preset          | CTA 橙 `#f97316`      | **保留蓝 primary / presets**；橙仅 `--cta` 附加                         |
| 画布 / 表面  | 实心 card/table/sidebar                      | 主内容实心            | **KEEP** solid surfaces；不加 default glass                             |
| 圆角         | `--radius: 1rem` · sm/md calc 比例           | 4 / 8                 | **默认** `--radius: 0.5rem`；**固定** `--radius-sm: 4px` / `--md: 8px` |
| 用户圆角轴   | `data-theme-radius` sm…xl                    | 可回退                | **KEEP** 轴；explicit `xl` = soft 选择，非回归                          |
| Named presets| 部分 mood 仍设 `1rem`                        | —                     | **Leave** preset 心情半径；只改 `:root` 默认 + chrome 硬编码 call sites |
| Header 霜    | public `backdrop-blur-2xl`（≈40px）          | chrome ≤12            | **token** `--atelier-panel-blur: 12px` + `.th-chrome-blur`              |
| Auth header  | 透明实心、无轨                               | 2–3px brand 轨        | `.th-chrome-rail` · primary→border 渐变（非彩虹）                       |
| Card 原语    | `rounded-xl` 硬编码                          | card 8                | → `rounded-md`（8px family）                                            |
| Public footer| `rounded-2xl` + `backdrop-blur-sm`           | chrome 8 / blur OK    | radius → `rounded-md`；blur-sm 保留                                     |
| CTA 按钮     | Button default = primary                     | CTA 一点              | **令牌 only**；不改 default Button 主色                                 |
| Classic web  | Semi 独立栈                                  | —                     | **Skip**                                                                |
| Go / D7 / relay | 业务后端                                  | —                     | **本波零改**                                                            |

---

## 2. 本波文件范围

| 做 | 不做 |
| -- | ---- |
| `web/default/src/styles/theme.css` 半径 + Atelier 附加令牌 + rail/blur 类 | 换框架 / 换 primary 为橙 |
| `public-header.tsx` blur≤12 · rounded-md | 重写 nav IA / features 全树 `rounded-*` |
| `header.tsx` `.th-chrome-rail` | glassShell 产品 flag / `?glassShell=1` |
| `card.tsx` 8px family | `web/classic/**` |
| `footer.tsx` public shell radius | middleware / relay / D7 / deploy live |
| `docs/design/atelier-v2-matrix.md`（本文件） | A2 snippets / custom cursor / View Transitions |
| — | theme cookie 键 / provider 行为变更 |

---

## 3. 令牌决策摘要

```css
--radius: 0.5rem;           /* default path base (was 1rem) */
--radius-sm: 4px;           /* controls — fixed */
--radius-md: 8px;           /* cards / panels — fixed */
--radius-lg: var(--radius); /* tracks axis / presets */
/* xl+ still scale from --radius so data-theme-radius=xl stays soft */

--cta / --cta-ink / --cta-soft   /* additive; Button default stays primary */
--atelier-panel-blur: 12px
--atelier-rail: 3px
--space-1..5: 4 / 8 / 16 / 24 / 32   /* legal ladder for NEW CSS */

.th-chrome-blur   /* backdrop-filter: blur(var(--atelier-panel-blur)) */
.th-chrome-rail   /* 3px primary→border top rail; pointer-events: none */
```

Dark：`--cta` 共享；`--cta-ink: #ffedd5`；**不**把 `--primary` 改成橙。

**Default path** = Atelier 4/8 family.  
**Explicit user `data-theme-radius='xl'`** = soft product choice, not a regression.

---

## 4. 验收

- [ ] `pnpm typecheck`（`web/default`）exit 0
- [ ] `pnpm build`（`web/default`）exit 0
- [ ] 视觉：默认圆角更利落、public chrome blur ≤12；蓝 primary + presets 仍在
- [ ] Auth header 有克制顶轨；主内容 / 表格 / 侧栏仍为实心
- [ ] 无 middleware / D7 / relay / deploy / classic diff
- [ ] 无 A2 / default-on glass flag / radical-snippets
- [ ] 用户 theme radius/preset cookies 仍生效

---

## 5. 回滚

```text
git checkout HEAD -- \
  web/default/src/styles/theme.css \
  web/default/src/components/layout/components/public-header.tsx \
  web/default/src/components/layout/components/header.tsx \
  web/default/src/components/ui/card.tsx \
  web/default/src/components/layout/components/footer.tsx
rm docs/design/atelier-v2-matrix.md
```

或单 commit revert。

---

## 6. 后续（非本波）

- CTA 实际接到稀有按钮（本波只定义令牌）
- features/** 深层 `rounded-2xl` 审计（仅高流量壳层）
- Named preset mood radii 可选收紧（需产品确认「max soft」预设例外）
- A2 intensity / radical snippets sandbox only
- D7 / security headers — 独立议程，**不得**借本视觉 commit 声称

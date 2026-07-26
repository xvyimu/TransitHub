# D-LEG · web/default legacy-guard 清单与断言建议

> **不删 `web/default`。** 该树是 D7 前生产默认 + ≤5min 回滚面（见 `docs/legacy-frontend-gate.md` · `docs/PROJECT.md` §2.2）。
> 本文只产出**守卫断言 + 混入清单**,供 CI / 人评审拦截「新功能进 legacy」。

## 1. 现状扫描（本 wt · base main）

锚点：`869d200d`（independent TransitHub stack，栈定型基线）。

| 事实 | 结论 |
|------|------|
| `git log a72c558c..HEAD -- web/default/` | **空** — Atelier 之后无新 web/default 改动 |
| `feat(web/default)` 全历史 | 4 笔：`a42b3976` v1.0 前端、`8b2b03d2` Base UI 大改、`a64f26d1` Anthropic 主题、`a72c558c` V2 Atelier A0/A1 |
| 最近端混入 | **`a72c558c`（2026-07-23）V2 Atelier A0/A1 structural chrome** — 在 stack 定型（`869d200d` 2026-07-22）之后 |

### 混入判定：`a72c558c`

改动面：`web/default/src/{components/layout,components/ui/card,styles/theme.css}` + `docs/design/atelier-v2-matrix.md`。

- 提交自述「No Go/middleware/classic/deploy changes」「additive tokens / structural chrome」——**视觉结构性 chrome**,非新设置页/表单/OAuth/billing。
- 未落 `docs/legacy-frontend-gate.md` §Allowed（安全 / 生产回归 / 回滚构建 / 已坏 i18n typo）任一类。
- 提交信息**无** `LEGACY-HOTFIX` 标记,**无**事件链接。
- 结论：**属 gate 定义的 Forbidden「Refactors "while we are here" / 新视觉功能」灰区**。已落库、`main` 祖先,本波**不回滚**（回滚 chrome 风险 > 收益,且会动回滚面）；改为**钉 guard 防后续再混入**,并把该笔记为「pre-guard legacy debt · accepted」。

## 2. Guard 断言建议（防后续再混入）

现有 `.github/workflows/quality.yml` 有 `web-quality`（typecheck/test/build/bundle budget）但**无 legacy-guard 语义门**。建议新增 PR 级断言（不改生产、不删树）：

### G-LEG-1 · 提交/PR 标记门（阻断级）
对 diff 命中 `web/default/**` 的 PR：
- 若改动**不**在 allowlist 路径（见 G-LEG-3），PR body 必须含 `LEGACY-HOTFIX:` + 事件/回归链接,否则 fail。
- 断言伪码：
  ```sh
  changed=$(git diff --name-only "$BASE"...HEAD -- 'web/default/**')
  if [ -n "$changed" ] && ! grep -qi 'LEGACY-HOTFIX' <<<"$PR_BODY"; then
    echo "::error::web/default touched without LEGACY-HOTFIX tag"; exit 1
  fi
  ```

### G-LEG-2 · commit message 前缀门（软警告 → 可升阻断）
- 拦 `feat(web/default)` / `feat(web):`（指向 default 树）新提交:
  ```sh
  git log --format='%h %s' "$BASE"...HEAD -- 'web/default/**' \
    | grep -E '^\w+ feat\(web(/default)?\):' && {
      echo "::warning::new feat() on legacy web/default — needs cutover exception"; }
  ```
- 例外：commit body 含 `LEGACY-HOTFIX` + 链接。

### G-LEG-3 · 允许路径白名单（回滚必需）
仅以下类别视为合法 legacy 改动,其余进阻断复核：
- 安全修复（XSS / auth bypass / 该树依赖 CVE）
- 生产回归修复（附 incident 链接）
- 保持 embed/回滚可构建的 build/tooling 修复
- 已在生产坏掉的 i18n typo

### G-LEG-4 · 反双写断言
同一 PR **同时**新增 `web/default/src/features/**` 与 `web-console/src/**` 同名能力 → fail（对齐 gate Resolution B「禁 React+Vue 长期双写」）。
```sh
d=$(git diff --name-only "$BASE"...HEAD -- 'web/default/src/features/**')
v=$(git diff --name-only "$BASE"...HEAD -- 'web-console/src/**')
[ -n "$d" ] && [ -n "$v" ] && echo "::error::possible React+Vue dual-write in one PR"
```

## 3. 落地状态

- [x] G-LEG-1/4 落成 CI 阻断门 — 独立 `.github/workflows/legacy-guard.yml`（PR 触发,阻断级），脚本 `scripts/check-legacy-guard.ps1`。放独立 workflow（非 `quality.yml`）以便语义门独立演进,并避免与并行任务争抢 `quality.yml`。
- [x] G-LEG-2 已实现为软警告（`::warning::`），观察噪声阶段,未阻断；后续视噪声升阻断。
- [x] G-LEG-3 allowlist 落进脚本注释 + `$allowlistPatterns`（当前仅 `web/default/bun.lock*` 纯 build/tooling），其余类别（安全 / 生产回归 / build 保绿 / i18n typo）经 `LEGACY-HOTFIX:` 标记 + 链接放行。
- [ ] CODEOWNERS 草案（gate 文档已列）启用后,`/web/default/ @xvyimu` 强制人评审。
- 已存债：`a72c558c` 标 `accepted pre-guard legacy debt`,不追溯回滚（guard 只查 merge base 之后的改动）。

### 本地自测（可复现）

```pwsh
# pass — 无 web/default 改动的分支
pwsh -NoProfile -File scripts/check-legacy-guard.ps1 -BaseRef origin/main -PrBody ''

# fail (G-LEG-1) — 改了 web/default 但 PR body 无标记：造一个改动再跑
#   'test' | Set-Content web/default/src/scratch-guard-test.tsx
#   git add -A && git commit -m 'test: legacy guard fixture'
#   pwsh -NoProfile -File scripts/check-legacy-guard.ps1 -BaseRef origin/main -PrBody ''
#   -> ::error::G-LEG-1: web/default touched (missing LEGACY-HOTFIX marker) ... exit 1
# pass 同场景 -PrBody 'LEGACY-HOTFIX: prod regression #123'
```

## 4. 红线复核
- ✅ 未删 `web/default`。
- ✅ 未改生产 `FRONTEND_MODE` / 未 flip。
- ✅ 未引入 React+Vue 双写。
- ✅ CI 门为纯 diff 断言,可逆低风险；仅拦「新功能混入 legacy」,不改生产代码 / 不删树。

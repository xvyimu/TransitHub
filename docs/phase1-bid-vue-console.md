# Phase1 Bid-B · Vue3+NaiveUI `web-console` 绞杀路径

| 字段 | 值 |
|------|-----|
| 角色 | Vue 面板架构师 · 槽 th-3 |
| 竞标 ID | Bid-B / phase1-bid-vue-console |
| 日期 | 2026-07-22 |
| 决议 | **B 绞杀**：新建 `web-console`（Vue3+TS+NaiveUI）；**禁止长期双写**；旧 React 只读维护至切流完成 |
| 输入 SSOT | `D:\orca\docs\architecture-decision-2026-07-22-approved.md` · Master `architecture-stack-refactor-master-2026-07-22.md` · ASIS `D:\TransitHub\src\docs\ARCHITECTURE_ASIS.md` |
| 性质 | **只出方案**；本任务不改 controller/service 业务逻辑、不 commit 业务、不 push |
| 主交付缝证据 | ADR-0001 · `FRONTEND_MODE` · `deploy/separated/` · `frontend_assets_{embedded,external}.go` |

---

## 1. 目标与非目标

### 1.1 目标（Phase1 面板切片 · T3–T5）

1. **新建同仓 `web-console/`**：Vue3 + TypeScript + **Vite** + **Naive UI** + Vue Router + Pinia + Vue I18n；与 Go 后端同源 `/api/*` 契约对接（先消费现有 API，不强制 `/api/v2`）。
2. **首屏可运行切片（MVP）**  
   - 登录（`POST /api/user/login` + cookie 会话）  
   - 登出 / 当前用户（`GET /api/user/logout` · `GET /api/user/self`）  
   - 健康/状态首屏：聚合展示 `GET /healthz|livez|readyz` + `GET /api/status`（及可选 `/frontend-healthz`）  
3. **信息架构（IA）定稿**：导航骨架 + 路由域与旧 `web/default/src/features/*` 映射表；Phase1 只实现「认证 + 健康」与空壳路由占位，写路径后置。
4. **切流路径可配置**：利用已有 `FRONTEND_MODE` + `deploy/separated` 同域 Nginx 反代；扩展 frontend 镜像可构建 `web-console`（不拆 Git 仓）。
5. **旧 React 废弃策略落地为文档/门禁草案**：`web/default` → LEGACY；`web/classic` → L2 冻结；CODEOWNERS / CI 注释禁止新功能进旧面板（实现阶段再改文件）。
6. **可 5 分钟配置级回滚**到 embedded React（ADR-0001 已论证）。

### 1.2 非目标（本 Phase / 本 Bid 明确不做）

| 不做 | 理由 |
|------|------|
| 全量重写 20+ features（渠道 merge、钱包、Playground SSE、设置矩阵…） | 绞杀分期；T4 只读优先 |
| 改 `controller/` `service/` 业务逻辑、计费、relay | 决议：模块一只出方案；热路径禁碰 |
| 长期 React+Vue 双主实现 | 决议禁止 C；双写仅迁移窗口 |
| 拆双仓 / 默认改 CORS 跨域控制台 | ADR-0001：同域反代优先 |
| Electron 壳重做 | TOOL/L2；可后指新 URL |
| 引入第二 UI 库并行（Element/Ant 等） | R4 同层一种 |
| 为 Vue 重写而发明 `/api/v2` | 先对齐现有 OpenAPI/`/api/*` |
| OAuth/Passkey/2FA 全覆盖进 MVP | MVP 用户名密码登录；多因子后续波次 |
| classic 主题迁移 | L2 冻结，不进绞杀主线 |

---

## 2. 改动文件/目录范围

### 2.1 新增（模块二实现时；本 Bid 仅规划）

```text
web-console/                         # 新建 · 绞杀目标面板（主实现）
  package.json                       # 独立 Vite 工程（可不进 web/ bun workspace，避免拖 classic）
  vite.config.ts
  tsconfig*.json
  index.html
  .env.development                   # 可选 dev proxy → :3000
  src/
    main.ts
    App.vue
    styles/                          # 全局 + Naive 主题覆写（克制）
    router/index.ts                  # 路由 + 鉴权守卫
    stores/auth.ts                   # Pinia：会话 / self
    api/
      http.ts                        # fetch/axios：withCredentials、同源 baseURL=''
      auth.ts                        # login / logout / self
      status.ts                      # /api/status + probes
    views/
      LoginView.vue
      HealthView.vue                 # 健康检查首屏
      NotFoundView.vue
    layouts/
      AuthLayout.vue
      ConsoleLayout.vue              # 顶栏/侧栏骨架（后续页挂载点）
    components/                      # 通用：N* 封装、错误边界、空态
    i18n/                            # vue-i18n；首批 en/zh
    types/api.ts                     # 与现有响应壳对齐（success/data/message）
  README.md                          # 本地 dev / 与 Go 联调

deploy/separated/
  Dockerfile.frontend.vue            # 或参数化现有 Dockerfile.frontend 的 BUILD_CONTEXT=web-console
  # nginx.conf.template 复用；SPA root 改为 web-console/dist

docs/
  phase1-bid-vue-console.md          # 本文（本任务唯一落盘）
  # 后续（执行规格合成后）：ARCHITECTURE_TARGET.md 面板章节、CODEOWNERS 草案
```

### 2.2 触碰但应「最小、配置向」（实现阶段）

| 路径 | 改动性质 |
|------|----------|
| `deploy/separated/Dockerfile.frontend` | 可选：多阶段/ARG 选择 `web/default` vs `web-console`；**默认仍可 build React** 直到 cutover gate |
| `deploy/separated/README.md` · `docs/operations/runtime-separation.md` | 文档：Vue 镜像与切流步骤 |
| `Makefile` | `build-console` / `docker-frontend-vue` 目标 |
| `.github/workflows/quality.yml` | 增加 `web-console` typecheck/build 与可选 image（勿删 React 构建直至 T5） |
| `CODEOWNERS`（若无则新增） | `web/default/**` `web/classic/**` → 仅 security/hotfix 审批 |
| 根 `README*.md` | 一句指向新控制台；**不**擦除受保护 branding |

### 2.3 明确不改（本切片）

- `controller/**` `service/**` `relay/**` `model/**`（除非 Bid-A 另开 OpenAPI 对齐）
- `web/default/**` `web/classic/**` 业务代码（仅 LEGACY 标注与门禁）
- `frontend_assets_*.go` 逻辑语义（继续 embed 旧主题；Vue 走 external + separated）
- 计费 / 配额 / AutoMigrate

### 2.4 信息架构（目标导航 · 与旧 features 映射）

旧 `web/default/src/features/*` → 新域（Phase1 仅 **Auth / Health** 实装，其余 **占位路由 + 403/Coming soon**）：

| 新 IA 域 | 路由前缀 | 旧 feature | Phase1 |
|----------|----------|------------|--------|
| 认证 | `/login` | `auth` | **实装**（密码登录） |
| 健康/运维概览 | `/health` 或 `/`（登录后默认） | 部分 `dashboard`/`system-info` | **实装**（status + probes） |
| 渠道 | `/channels` | `channels` | 占位 → T4 只读列表 |
| 模型/供应商 | `/models` | `models` | 后置 |
| 令牌/密钥 | `/keys` | `keys` | 后置 |
| 用量日志 | `/logs` | `usage-logs` | T4 候选只读 |
| 用户/权限 | `/users` | `users` | 后置 |
| 订阅/钱包 | `/billing` | `wallet` `subscriptions` | 后置（高风险） |
| 系统设置 | `/settings` | `system-settings` | 后置 |
| 系统信息 | `/system` | `system-info` | T4 候选 |
| Playground/Chat | `/playground` | `playground` `chat` | **最后**（SSE） |
| Setup | `/setup` | `setup` | 后置；首次安装仍可用旧面板或纯 API |
| 个人中心 | `/profile` | `profile` | 后置 |
| 定价/排行/关于 | 次要 | `pricing` `rankings` `about` `legal` | 后置 |

**IA 原则**：运营控制台优先「只读可观测 + 渠道/日志」；面向终端用户的营销首页/复杂 Playground 不阻塞绞杀主线。

---

## 3. 依赖与上下游

### 3.1 上游依赖（阻塞 / 协作）

| 依赖 | 来源 | 本 Bid 要求 |
|------|------|-------------|
| 决议 B | 已批准 | 不重开 A/C 选型，除非附录强证据（无） |
| 交付缝 | ADR-0001 已存在 | **不重造**；Vue 只挂到 seam |
| Go `/api/*` 稳定 | Bid-A OpenAPI/路由表 | MVP 用现有路径；契约漂移由 Bid-A 审计 |
| Session cookie | 现有 `SESSION_COOKIE_*` + `UserAuth` | **强制同域**；`withCredentials: true` |
| 分发 | `deploy/separated` Nginx 已代理 `/api` `/healthz`… | SPA `try_files` 不变 |

### 3.2 下游消费

| 消费者 | 期望 |
|--------|------|
| 协调员 Phase1 执行规格 | 采纳脚手架 + 切流步骤 + 废弃策略 |
| 模块二 Vue 实现师 | 按 S0–S6 落地 `web-console` |
| 运维 / CI | `FRONTEND_MODE=disabled` + Vue frontend image |
| 旧 React 维护者 | 仅 hotfix；新功能拒收 |

### 3.3 API 对接清单（MVP · 复用现有，零后端改动）

| 用途 | 方法 / 路径 | 鉴权 | 备注 |
|------|-------------|------|------|
| 登录 | `POST /api/user/login?turnstile=` | 匿名 + 限流 + Turnstile（若开） | Body: `username` `password`；Set-Cookie |
| 2FA（非 MVP 必做） | `POST /api/user/login/2fa` | 会话中间态 | 若 status 要求 2FA 再接 |
| 登出 | `GET /api/user/logout` | 会话 | |
| 当前用户 | `GET /api/user/self` | UserAuth | 守卫与顶栏 |
| 系统状态 | `GET /api/status` | 公开/半公开 | 首屏卡片主数据源 |
| 进程探活 | `GET /healthz` `GET /livez` `GET /readyz` | 公开 | Nginx 已反代；readyz 含依赖 |
| 前端容器探活 | `GET /frontend-healthz` | Nginx 本地 | 编排健康检查，不探后端 |

**HTTP 客户端约定（对齐旧 `web/default/src/lib/api.ts`）**

- `baseURL = ''`（同源）
- `credentials: 'include'` / axios `withCredentials: true`
- 业务响应壳：兼容现有 `{ success, message, data }`（以 OpenAPI/实抓为准；实现时以 `docs/openapi/api.json` + 实机为准）
- 401：清 Pinia 会话 → `/login`；不弹跨域 CORS 方案

### 3.4 脚手架选型（锁定）

| 项 | 选择 | 理由 |
|----|------|------|
| 构建 | **Vite 6+** | 生态默认、Naive 官方示例友好；与旧 Rsbuild 解耦 |
| 语言 | TypeScript strict | 主参考 TS |
| UI | **Naive UI** | 产品线 SSOT；管理台组件密度高 |
| 路由 | vue-router 4 | 标准 |
| 状态 | Pinia | 会话/轻量 UI |
| 请求 | **axios**（与旧面板一致）或 ofetch | 团队已熟悉拦截器模式；二选一写进实现规格，推荐 axios 降低契约心智 |
| i18n | vue-i18n | 对齐多语言产品；MVP en+zh |
| 包管理 | **pnpm 或 bun**（二选一，推荐 **pnpm** 独立锁文件） | 不绑死 `web/` workspace 的 React catalog |
| 质量 | `vue-tsc` + ESLint + prettier（克制） | CI 门禁 |
| 测试 | Vitest 单测 API 封装 + Playwright 烟雾（后） | 见 §6 |

**不选**：Nuxt（SSR 对 cookie 管理台无收益且复杂化 separated 部署）；Element Plus（偏离 SSOT）；微前端 qiankun（过度）。

### 3.5 与 MindSync 面板

MindSync 已是 Vue3+NaiveUI（决议：不动）。本仓 `web-console` **不**做跨仓 monorepo 组件强共享；可共享的仅是「约定」：Naive 版本大版本对齐、i18n 键风格、错误 toast 语义。避免过早抽 `packages/ui`。

---

## 4. 分步执行计划（S0… 每步验收）

> 本任务只写方案。下列步骤供模块二实现；**S0 可在方案 gate 后立即做且仍无业务后端改动**。

### S0 · 仓库落位与治理（0.5–1d）

- 创建 `web-console/` 空脚手架（Vite vue-ts）。
- 文档：`web-console/README.md`（dev 端口、proxy、与 `FRONTEND_MODE` 关系）。
- 草案：`CODEOWNERS` / CONTRIBUTING 一节「新 UI 只进 web-console」。

**验收**：`pnpm i && pnpm dev` 出空白页；CI 可选 `working-directory: web-console` typecheck 任务草稿。

### S1 · 工程基线（1–1.5d）

- Naive UI 按需/全量引入 + 暗色可选（跟随系统即可）。
- `api/http.ts`：credentials、错误归一、401 处理。
- 环境：`.env.development` 中 `server.proxy` → `http://127.0.0.1:3000`（仅 dev；生产靠 Nginx）。
- vue-i18n en/zh 最小字典。

**验收**：浏览器 Network 可见代理后的 `GET /api/status` 200（后端已起）。

### S2 · 登录闭环（1.5–2d）

- `LoginView`：用户名/密码；可选 Turnstile 位（若 `/api/status` 指示开启则展示，否则隐藏）。
- 成功：拉 `/api/user/self` → 进控制台 layout。
- 失败：展示 `message`；不记密码。
- 路由守卫：无会话访问 `/health` → `/login?redirect=`（redirect 白名单防开放重定向；对齐旧 `safe-redirect` 思路）。

**验收**：同域（dev proxy 或 separated）登录后 Cookie 存在；刷新仍保持会话；登出后 self 401。

### S3 · 健康检查首屏（1d）

- `HealthView`：卡片展示  
  - 前端：`/frontend-healthz`（若存在）  
  - 后端：`/healthz` `/livez` `/readyz` 状态码 + JSON 摘要  
  - 业务：`/api/status` 关键字段（版本、是否开启注册等——以实响应为准，禁止臆造字段）
- 手动刷新 + 进入页自动拉一次。

**验收**：后端 stop 时 readyz/health 红态可见；恢复后变绿；与 curl 结果一致。

### S4 · 布局与 IA 骨架（1d）

- `ConsoleLayout`：侧栏按 §2.4 域列出；未实装路由 → 「迁移中」页。
- 顶栏：用户名、登出。
- 默认登录后落地 `/health`。

**验收**：信息架构可评审；无死链（占位页 200）。

### S5 · 切流与镜像（1.5–2d）

**推荐生产切流（同域）**

```text
browser → Nginx(web-console dist :8080)
            ├── / /assets /frontend-healthz
            └── /api /v1 /v1beta /mj /pg /suno /kling /jimeng
                /healthz /livez /readyz  →  Go :3000
                                            FRONTEND_MODE=disabled
                                            -tags frontend_external
```

步骤：

1. 构建纯后端：`go build -tags frontend_external`，`FRONTEND_MODE=disabled`。
2. 构建 Vue 前端镜像（新 Dockerfile 或 ARG）。
3. `TRUSTED_PROXY_CIDRS` 含代理网段。
4. Smoke：扩展 `deploy/separated/smoke.ps1|sh` 增加登录+status 断言（或独立 `smoke-console`）。
5. **灰度**：先内网/预发；默认 Docker 集成镜像仍可 embed React 直至组织 cutover 日。

**备选**：`FRONTEND_MODE=redirect` + `FRONTEND_BASE_URL` 跨源——**不推荐**默认（Cookie/CORS/OAuth 面扩大，ADR 已否）。

**验收**：

- `GET /frontend-healthz` → ok  
- 登录 cookie 首方可用  
- `GET /api/status` 与旧面板同数据  
- 5 分钟回滚：切回 integrated 镜像 / `FRONTEND_MODE=auto|embedded`

### S6 · 旧 React 废弃策略（0.5–1d · 文档+门禁）

| 资产 | 策略 |
|------|------|
| `web/default` | **LEGACY**：只修安全/严重回归；**禁止新功能** |
| `web/classic` | **L2 冻结**：不跟 default 双写 |
| Embed 路径 | cutover 前默认保留；cutover 后主路径 separated+Vue；embed 可保留一版应急 |
| CODEOWNERS | `web/default/**` `web/classic/**` 高审；`web-console/**` 为面板默认 |
| CI | 可选：对 `web/default` 的非文档 PR label `legacy-ui` 人工确认 |
| 发布说明 | 用户可见：控制台技术栈迁移；URL 尽量不变（同域） |

**验收**：书面策略合入执行规格；实现阶段 PR 模板勾选「UI 变更路径」。

### S7 · T4 预告（不在本 Bid 工期，仅接口）

下一波只读页建议顺序：`/api/channel` 列表只读 → 日志只读 → system-info → 再写路径。每页：旧 React 对照表 + 截图/JSON diff。

---

## 5. 风险与回滚

| ID | 风险 | 等级 | 缓解 | 回滚 |
|----|------|------|------|------|
| R1 | Session/Cookie 分域失效 | **高** | 强制同域 Nginx；禁默认跨域 SPA | 回 embedded 集成镜像 |
| R2 | Turnstile/OAuth/2FA 行为不一致 | 中 | MVP 仅密码；status 驱动能力开关；OAuth 回调 host 不变 | 旧面板登录 |
| R3 | 响应壳/字段与 OpenAPI 漂移 | 中 | 与 Bid-A 对表；实现时实抓 fixture | 适配层改客户端 |
| R4 | 双轨期心智与缺陷面 | **高** | 短窗口；CODEOWNERS 禁新功能进 React；R4 禁长期双写 | 加速 cutover 或（仅事故）停 Vue 入口 |
| R5 | SSE/WS 缓冲（后续 Playground） | 中 | 复用现成 nginx `proxy_buffering off`；Playground 最后迁 | 旧面板 playground |
| R6 | i18n 与权限矩阵漏迁 | 中 | 分期；admin 路由守卫对齐 role 字段 | 功能留旧面板直至对齐 |
| R7 | CI 时间/镜像膨胀 | 低 | ARG 切换前端上下文；不必双镜像长期发布 | 只发一端 |
| R8 | 范围蔓延「先把渠道 CRUD 做完」 | **高** | 严格执行 S2–S3 MVP；写路径 gate | 砍 scope |
| R9 | 保护标识被误改 | 低 | 不碰 branding/license 清洗 | git revert |

**回滚手册（运维 5 分钟）**

1. 边缘指回 **integrated** 镜像/二进制（embed React）。  
2. 清空或 `FRONTEND_MODE=auto`（勿 `disabled`）。  
3. 无需 DB down-migration（面板交付无 schema 依赖）。  
4. 验证：`/api/status` + 旧登录页可用。

---

## 6. 测试建议

### 6.1 自动化

| 层 | 内容 |
|----|------|
| 单元 | `auth`/`status` API 封装：mock axios；401→清会话；成功→self |
| 组件 | Login 表单校验（空用户名/密码）；Health 卡片三态（ok/degraded/down） |
| 契约 | 可选：对 `docs/openapi/api.json` 抽 login/status/self 做 response 类型生成或手工 fixture |
| 烟雾 | 扩展 separated smoke：`/frontend-healthz`、`/healthz`、login→self→logout |
| 回归 | 切流后抽样：cookie `Secure`/`SameSite` 在 HTTPS 预发 |

### 6.2 手动验收清单（MVP 出门）

- [ ] Dev：Vite proxy 登录成功  
- [ ] Separated compose：同域登录 + 刷新保持  
- [ ] 错误密码提示与旧面板语义一致（message）  
- [ ] 后端宕机：Health 页可见失败，不白屏  
- [ ] 登出后前进/后退不落到半登录态  
- [ ] `/metrics` 在公共边缘仍 404（安全不回归）  
- [ ] 回滚到 embedded React 成功  

### 6.3 非目标测试

- 不做假压测/随机 fuzz 撑覆盖率  
- 不在本阶段对计费/渠道 merge 做 E2E  

---

## 7. 相对主参考偏离（若有）+ 证据

| 项 | 主参考 | 本方案 | 判定 |
|----|--------|--------|------|
| 面板栈 | TS + Vue3 + NaiveUI | **一致** | 无偏离 |
| 绞杀 | 新面板并行至切完 | `web-console` + 短双轨 | 一致 |
| 禁长期双写 | R4 | CODEOWNERS + LEGACY | 一致 |
| 同域交付 | 分层契约 HTTPS/JSON | 复用 ADR-0001 | 一致 |
| 构建器 | （未规定） | **Vite** 而非 Nuxt/Rsbuild | **可接受实现选择**；非栈偏离 |
| 包管理 | （未规定） | 独立 pnpm（推荐） | 可接受；若全仓强制 bun 可改为 bun，**不**构成 R2 |
| API | OpenAPI | 先消费现有 `/api` | 与 Master T3 一致；v2 非必须 |
| 附录 A 暂留 React | 主参考允许 R2 例外 | **不采纳** | 见附录：代价不足以推翻已批决议 B |

**结论**：本 Bid **无**需 R2 推翻主参考的偏离。唯一需协调员确认的实现细节是包管理器（pnpm vs bun）与 axios vs ofetch——建议默认 **pnpm + axios**。

---

## 8. 工作量粗估

| 阶段 | 人日（1 名熟悉 Vue 的实现师） | 说明 |
|------|-------------------------------|------|
| S0 治理 + 空架 | 0.5–1 | |
| S1 基线 | 1–1.5 | |
| S2 登录 | 1.5–2 | 含 Turnstile 位与守卫 |
| S3 健康首屏 | 1 | |
| S4 IA 骨架 | 1 | |
| S5 切流/镜像/smoke | 1.5–2 | 含文档 |
| S6 废弃策略落文档/门禁 | 0.5–1 | |
| **Phase1 面板合计** | **约 7–10 人日** | 不含 T4 只读页迁移 |
| T4 单只读页（预告） | 1–3 / 页 | 视表格与权限 |
| 全 features 绞杀完成 | **数月级** | 20+ 域；需独立路线图，非本 Bid |

**关键路径**：S2 会话正确性 + S5 同域切流。  
**并行**：S5 镜像可与 S3/S4 并行（不同人）。  
**与 Bid-A/C**：不阻塞脚手架；登录契约以现网为准，OpenAPI 审计可并行修正文档。

---

## 附录 A · 「暂留 React」代价对照（不推翻决议 B）

> 仅满足任务「对比暂留 React」；**用户已批准 B**，本附录 **不** 启动 R2 改判，除非出现新的压倒性证据（当前无）。

| 维度 | 暂留 React（A） | 绞杀 Vue（B · 本方案） |
|------|-----------------|------------------------|
| 与产品线 SSOT | 永久偏离；跨仓（MindSync Vue）组件/招聘不一致 | 对齐主参考 |
| 工期（Phase1） | 0 面板重写；P0 全给 Go/SQL | 7–10 人日得可切流 MVP |
| 契合度 | 极高：132k+ TSX 已接 cookie/SSE | 低初期；交付缝已存在降低部署成本 |
| 双写风险 | 无（单实现） | 迁移窗有；须强门禁 |
| 回滚 | N/A | 配置级回滚已有 |
| 机会成本 | 继续在 React 加功能 → 未来若再 Vue 更贵 | 新功能进 Vue，旧债冻结 |
| 决议状态 | **未采纳** | **已批准** |

**测绘时 A 的合理场景**（历史）：团队只保网关正确性、面板例外书面 R2。  
**当前**：组合 SSOT + 用户批准 B + 交付缝就绪 → **执行 B**；A 仅作事故回滚形态（embed React），不是发展方向。

---

## 附录 B · 切流 Runbook 摘要（实现后粘贴运维手册）

```bash
# 1) 后端纯 API
go build -trimpath -tags frontend_external -o th-backend .
FRONTEND_MODE=disabled TRUSTED_PROXY_CIDRS=... ./th-backend

# 2) 前端（示例）
# docker build -f deploy/separated/Dockerfile.frontend.vue -t th-console:local .
# 或 compose profile console-vue

# 3) Smoke
FRONTEND_BASE=http://127.0.0.1:8080 ./deploy/separated/smoke.sh
# + 手动：登录 /health 三探活

# 4) 回滚
# 部署上一版 integrated 镜像；FRONTEND_MODE 置空或 auto
```

---

## 附录 C · 证据索引

| 主题 | 路径 |
|------|------|
| 决议 B | `D:\orca\docs\architecture-decision-2026-07-22-approved.md` |
| Master T3–T5 | `D:\orca\docs\architecture-stack-refactor-master-2026-07-22.md` §4 Phase1 |
| ASIS 面板/交付缝 | `D:\TransitHub\src\docs\ARCHITECTURE_ASIS.md` §4 §8.1 §9 |
| ADR 交付缝 | `docs/adr/0001-frontend-backend-delivery-seam.md` |
| FRONTEND_MODE | `router/main.go` `parseFrontendMode` / `setFrontendRouter` |
| Separated | `deploy/separated/*` |
| 登录 API 消费 | `web/default/src/features/auth/api.ts` · `router/api-router.go` userRoute |
| 探活 | `router/main.go` `/healthz` `/livez` `/readyz` |
| 旧 features 域 | `web/default/src/features/*`（约 20+） |

---

## 附录 D · 协调员合成时注意

1. **Bid-B 与 Bid-A**：面板 MVP **不**等待完整 OpenAPI 重写；但 login/status/self 字段以 A 的路由真源为准做一次对表。  
2. **Bid-B 与 Bid-C**：无直接阻塞；面板不依赖 migrations 目录。  
3. **默认包管理/HTTP 库**：请在执行规格写死（建议 pnpm + axios），避免实现师分叉。  
4. **Cutover 日**：建议 MVP 合入后 ≥1 个迭代再默认切公共流量；此前 Vue 为预发/内部。

---

*Bid-B 完。只读方案文档；未改业务代码、未 push。*

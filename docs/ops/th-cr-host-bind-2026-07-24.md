# HOST 绑定与网络暴露面 · 运维指南（TH-CR-004 / TH-CR-005）

> **本机运维剖面：LOCAL-ONLY。** 生产/本机默认绑定 `127.0.0.1`，不做公网暴露推荐。
> 本文仅文档；不改监听默认行为代码。栈锁见 [`docs/PROJECT.md`](../PROJECT.md)。
> **D7 NOT EXECUTED** · 不改生产配置 · 不 push 默认分支。

---

## 1. HOST 绑定行为（以 `main.go` 为准）

监听地址由 `HOST` + `PORT` 决定（`main.go` §280–289）：

```go
// HOST optional: empty/"0.0.0.0"/"::" → all interfaces; "127.0.0.1" → loopback only.
host := strings.TrimSpace(os.Getenv("HOST"))
addr := ":" + port
if host != "" && host != "0.0.0.0" && host != "::" {
    addr = net.JoinHostPort(host, port)
}
```

| `HOST` 取值 | 实际监听 | 暴露面 | 风险 |
|-------------|----------|--------|------|
| **未设 / 空** | `:PORT`（如 `:3000`） | **所有接口**（含 LAN/公网 IP） | ⚠️ **默认即全接口**。任何可路由到本机的网络都能访问 |
| `0.0.0.0` | `:PORT` | **所有 IPv4 接口** | ⚠️ 同上，等价全暴露 |
| `::` | `:PORT` | **所有 IPv6 接口** | ⚠️ 同上 |
| **`127.0.0.1`** | `127.0.0.1:PORT` | **仅本机回环** | ✅ **本机推荐剖面** |
| `::1` | `[::1]:PORT` | 仅本机 IPv6 回环 | ✅ 本机（IPv6） |
| 具体网卡 IP（如 `10.0.0.5`） | 该 IP:PORT | 该网段可见 | 视网络边界而定，需配套防火墙 |

**关键点：** 代码默认（不设 `HOST`）= **绑定全接口**，不是回环。LOCAL-ONLY 剖面必须**显式**设 `HOST=127.0.0.1`，否则本机在同一 LAN / 公网 IP 上会被直接访问。

### 本机推荐剖面

```pwsh
# LOCAL-ONLY：仅回环，配额/令牌/管理台不出本机
$env:HOST = '127.0.0.1'
$env:PORT = '3000'
```

若必须绑定网卡 IP 或 `0.0.0.0`（例如反代在另一台机器），**不属本机剖面**，须自行承担网络边界责任：前置反向代理 + 防火墙白名单 + 认证，且不在本文档范围内推荐。

---

## 2. TRUSTED_PROXY_CIDRS 误配风险

`TRUSTED_PROXY_CIDRS` 决定哪些来源的 `X-Forwarded-For` / `X-Real-IP` 被信任（`trusted_proxy.go`）。**误配会导致客户端 IP 被伪造**：把不受控网段（或 `0.0.0.0/0`）列入信任，任意客户端即可通过伪造转发头冒充任意 IP，绕过基于 IP 的限流/审计/封禁。默认留空 = 忽略转发头（安全）；仅当端口 **不**直接暴露、且前置的是自己可控的反代时，才设 `TRUSTED_PROXY_CIDRS=127.0.0.1/32,::1/128`（本机反代）或 compose 网桥网段。

---

## 3. TLS / SMTP insecure（TH-CR-004）

两个 opt-in 开关，**默认 `false`，生产禁开**（`common/init.go` §97、§108）：

| 变量 | 默认 | 作用 | 生产 |
|------|------|------|------|
| `TLS_INSECURE_SKIP_VERIFY` | `false` | 跳过上游 AI provider TLS 证书校验 | ❌ 禁。开则上游全线 MITM 面 |
| `SMTP_INSECURE_SKIP_VERIFY`（含旧名 `SMTP_TLS_INSECURE_SKIP_VERIFY`） | `false` | 跳过 SMTP 邮件服务器 TLS 校验 | ❌ 禁。开则邮件通道 MITM |

仅在明确的本机调试且理解风险时临时开启，用完立即清除。LOCAL-ONLY 剖面风险相对低，但**公网/共享网络部署下开启 = 致命**。

---

## 4. 部署前检查清单

| # | 检查项 | 期望（LOCAL-ONLY） | 命令 / 依据 |
|---|--------|--------------------|-------------|
| 1 | `HOST` 已显式设置 | `HOST=127.0.0.1` | `echo $env:HOST` |
| 2 | 未依赖默认（默认=全接口） | 不留空、不 `0.0.0.0` / `::` | 见 §1 表 |
| 3 | 实际监听确认为回环 | `TCP 127.0.0.1:3000 LISTENING` | `Get-NetTCPConnection -LocalPort 3000 -State Listen` |
| 4 | 无外部接口监听 | 无 `0.0.0.0:3000` / LAN IP | 同上，检查 `LocalAddress` |
| 5 | `TRUSTED_PROXY_CIDRS` 正确 | 留空，或仅本机反代网段 | `echo $env:TRUSTED_PROXY_CIDRS` |
| 6 | `TLS_INSECURE_SKIP_VERIFY` | `false` / 未设 | `echo $env:TLS_INSECURE_SKIP_VERIFY` |
| 7 | `SMTP_INSECURE_SKIP_VERIFY` | `false` / 未设 | `echo $env:SMTP_INSECURE_SKIP_VERIFY` |
| 8 | pprof（若开）已锁回环 | `ENABLE_PPROF` 未设或知悉其绑 `127.0.0.1:8005` | `main.go` §189–195 |
| 9 | 防火墙（若必须外绑） | 白名单 + 反代 + 认证 | 不在本机剖面推荐 |

监听核对示例：

```pwsh
Get-NetTCPConnection -LocalPort 3000 -State Listen |
  Select-Object LocalAddress, LocalPort, State
# 期望 LocalAddress = 127.0.0.1（或 ::1）。若见 0.0.0.0 / :: → HOST 未锁，属暴露风险。
```

---

## 5. `.env.example` 核对

当前 `.env.example` **没有** `HOST` 绑定注释项（仅有 Pyroscope 的 `# HOSTNAME=...`，语义无关）。
本模块边界为 docs-only，`.env.example` 现无 HOST 一行注释，故**未新增**（避免改动配置模板）。
若后续要落地，建议加一行（示例，非本次改动）：

```
# 监听地址；留空/0.0.0.0/:: = 所有接口（暴露面大），本机剖面请设 127.0.0.1
# HOST=127.0.0.1
```

---

## 6. 边界声明

- 仅 `docs/ops/`；未改 `main.go` 监听默认行为，未改生产配置。
- 未把 `0.0.0.0` 写成推荐；本机剖面推荐值恒为 `127.0.0.1`。
- **D7 NOT EXECUTED**；未 push 默认分支。

修订：2026-07-24 · 对照 `transithub-findings.md` TH-CR-004 / TH-CR-005。

# M-TH-cr-host-bind-docs · evidence · 2026-07-24

> **D7 FLIP: NOT EXECUTED** · no production FRONTEND_MODE · no default-branch push

| Field | Value |
|-------|--------|
| Module | **M-TH-cr-host-bind-docs** |
| Branch | `xvyimu/th-cr-host-bind-docs` |
| Findings | **TH-CR-004** TLS insecure · **TH-CR-005** HOST bind |
| Guide (SSOT body) | [th-cr-host-bind-2026-07-24.md](./th-cr-host-bind-2026-07-24.md) |
| Status | **DONE** · **in-review** |

## Delivered

1. HOST empty/0.0.0.0/:: = all interfaces risk table + LOCAL-ONLY `HOST=127.0.0.1`
2. TRUSTED_PROXY_CIDRS misconfig risk
3. TLS_INSECURE_SKIP_VERIFY / SMTP insecure default false · prod ban
4. Deploy pre-check checklist
5. .env.example HOST gap noted (docs-only; template not edited this knife)

## Verification

| Check | Result |
|-------|--------|
| Guide path present | `docs/ops/th-cr-host-bind-2026-07-24.md` |
| Business code change | **none** |
| D7 / FRONTEND_MODE | **untouched** |

## Risk (one line)

**Risk:** default empty HOST still binds all interfaces — operators who assume loopback-by-default will expose admin/token surfaces on LAN/public IP without realizing.

STATUS: DONE + in-review · D7 NOT EXECUTED

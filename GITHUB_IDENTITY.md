# GitHub identity · TransitHub

| | |
|--|--|
| **Repository** | https://github.com/xvyimu/TransitHub |
| **Go module** | `github.com/xvyimu/TransitHub` |
| **Former names** | `xvyimu/new-api` · local product UI may still say NewAPI |
| **Fork network** | Detached (not a GitHub fork of upstream) |
| **Upstream lineage** | QuantumNous / Calcium-Ion `new-api` · ancestor `songquanpeng/one-api` |
| **License** | **AGPL-3.0** (inherited; see [LICENSE](./LICENSE) + [NOTICE](./NOTICE)) |
| **Local path** | 工作区 `D:\TransitHub` · git `D:\TransitHub\src`（入口 `D:\projects\TransitHub`） |
| **Git remote** | **only** `origin` → `xvyimu/TransitHub` |

## Independent development system

This tree is developed as **TransitHub**:

- Default remote is **only** `origin` (`xvyimu/TransitHub`). Upstream remotes (`Calcium-Ion`, `QuantumNous`, local `_qn_tmp`) are **not** configured.
- Feature work branches from the live/feature line under TransitHub; merge target is this repo’s default branch(es), not upstream.
- Go imports and `-ldflags -X` version symbols use `github.com/xvyimu/TransitHub/...`.
- Ops / ADR / handoff live under this product’s docs (`../docs`, `../agent_docs` when present).

## What is intentionally preserved (AGPL)

- Full **AGPL-3.0** text in [LICENSE](./LICENSE).
- [NOTICE](./NOTICE): QuantumNous copyright, §7 attribution text, and original project link.
- UI / footer / About attribution required by NOTICE (must not strip).
- Product capability remains New API–lineage LLM gateway; **repository identity** is TransitHub.

## What is intentionally *not* done

- No relicense to MIT/proprietary.
- No deletion of upstream copyright or NOTICE obligations.
- No claim of “original work with no lineage.”
- History rewrite of past commits is out of scope of this identity pass.

## Push

```powershell
cd D:\TransitHub\src
git remote -v   # expect only origin → xvyimu/TransitHub
git push origin HEAD
```

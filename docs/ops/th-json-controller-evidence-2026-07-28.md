# M-TH-json-ctrl — Evidence Document

**Date:** 2026-07-28  
**Task:** M-TH-json-ctrl · WAVE-DEBT-LONG  
**Branch:** `th-json-controller-2026-07-28`  
**Base tip:** `governance-cleanup-2026-07-28` @ `cb61e983`  
**Scope:** `controller/` + `common/json.go` only. No D7, no relay rewrite, no push to `main`.

## Summary

Prior waves already moved controller `json.Marshal` / `json.Unmarshal` / `json.NewDecoder` call sites to `common.*`. This batch:

1. Inventories remaining `encoding/json` imports under `controller/`.
2. Adds `common.Valid` so business code need not call `encoding/json.Valid` directly.
3. Migrates the only remaining controller **call** site (`channel_status_tag.go` ×2) to `common.Valid`.
4. Documents remaining controller imports as **type-only** exceptions (`json.RawMessage`, `json.Number`).

## Inventory (before)

```text
controller/channel-test.go       encoding/json   type only (json.RawMessage)
controller/channel_crud.go       encoding/json   type only ([]json.RawMessage)
controller/channel_status_tag.go encoding/json   CALLS: json.Valid ×2
controller/model_sync.go         encoding/json   type only (json.RawMessage field)
controller/ratio_sync.go         encoding/json   type only (json.Number, json.RawMessage)
controller/token_test.go         encoding/json   type only (json.RawMessage in test DTO)
```

`rg 'json\.(Marshal|Unmarshal|NewEncoder|NewDecoder)\s*\(' controller` → **0 hits** (already clean from prior merge `940da124`).

## Files Changed

| File | Change |
|------|--------|
| `common/json.go` | Add `Valid(data []byte) bool` wrapper over `json.Valid`. |
| `common/json_test.go` | Table-ish assertions for `Valid` true/false cases. |
| `controller/channel_status_tag.go` | `json.Valid` → `common.Valid` (2 sites); drop `encoding/json` import. |

## Remaining controller `encoding/json` (exceptions)

AGENTS.md / PROJECT.md allow **type** references (`json.RawMessage`, `json.Number`) while forbidding direct marshal/unmarshal. Remaining files keep the import for types only:

| File | Why kept |
|------|----------|
| `controller/channel-test.go` | `json.RawMessage` literals for Responses-format test input |
| `controller/channel_crud.go` | `[]json.RawMessage` when parsing multi-key JSON arrays |
| `controller/model_sync.go` | `upstreamModel.Endpoints json.RawMessage` |
| `controller/ratio_sync.go` | `case json.Number:` + probe `Data json.RawMessage` |
| `controller/token_test.go` | test response DTO `Data json.RawMessage` |

**No remaining `json.Valid` / `json.Marshal` / `json.Unmarshal` / `json.NewEncoder` / `json.NewDecoder` in `controller/`.**

## Verification

### Tests (run 2026-07-28, this worktree)

```text
go test ./common/ ./controller/ ./dto/ ./model/
# exit 0
ok  github.com/xvyimu/TransitHub/common      5.285s
ok  github.com/xvyimu/TransitHub/controller  2.448s
ok  github.com/xvyimu/TransitHub/dto         1.731s
ok  github.com/xvyimu/TransitHub/model       8.191s
```

### Architecture guard

```text
pwsh scripts/check-architecture-guards.ps1
```

Guard only fails on **newly added** direct marshal/unmarshal lines; this change removes a call and does not add any.

## Out of scope / next debt

- `relay/channel/**`, `dto/**`, `service/**`, `model/**` still import `encoding/json` for types and some call sites — separate waves.
- No production binary, no `:3000` service, no `main` push, no D7 flip.

## Explicit non-actions

- **Did not** fast-forward `main` (see companion decision card).
- **Did not** delete migrations or touch live DSN.

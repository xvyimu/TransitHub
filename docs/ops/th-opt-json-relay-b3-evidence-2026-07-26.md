# M-TH3-json-relay-b3 — Evidence Document

**Date:** 2026-07-26
**Branch:** `xvyimu/th-opt-json-relay-b3`
**Base:** `xvyimu/th-audit-p0-fixes-2026-07-25` (tip `6db2805b`)
**Commit:** `c92581cd`

## Summary

WAVE-OPTIMIZE B3: Replaced direct `encoding/json` calls with `common.*` JSON helpers in 3 `relay/channel/` adapter files.

## Files Changed

| File | Changes |
|------|---------|
| `relay/channel/cohere/relay-cohere.go` | 6 calls: 3 `json.Unmarshal` → `common.Unmarshal`, 3 `json.Marshal` → `common.Marshal`. Removed `encoding/json` import. |
| `relay/channel/palm/relay-palm.go` | 3 calls: 2 `json.Unmarshal` → `common.Unmarshal`, 1 `json.Marshal` → `common.Marshal`. Removed `encoding/json` import. |
| `relay/channel/siliconflow/relay-siliconflow.go` | 2 calls: 1 `json.Unmarshal` → `common.Unmarshal`, 1 `json.Marshal` → `common.Marshal`. Added `common` import, removed `encoding/json` import. |

**Total:** 11 `encoding/json` direct calls eliminated, 3 packages cleaned.

## Verification

### Build
```
go build ./relay/channel/cohere  ✓
go build ./relay/channel/palm     ✓
go build ./relay/channel/siliconflow ✓
```

### Tests
```
go test ./relay/channel/...  ✓
```
All 10 test packages passed (47 packages total, 37 with no test files).

### Architecture Guards
```
pwsh scripts/check-architecture-guards.ps1  ✓
```
Architecture guard passed (base d72f137...; only added Go lines checked).

## Scope Notes

- No changes to `controller/`, `common/`, or `relay/helper/` (covered in prior waves).
- No executables built, no live services touched.
- Branch pushed to `origin/xvyimu/th-opt-json-relay-b3` (not `main`).
# TH-CR-refund-idempotency-tests · Evidence

**模块 ID：** M-TH-cr-refund-idempotency-tests
**日期：** 2026-07-24
**对照 findings：** `code-review/transithub-findings.md` · **TH-CR-003**
**边界：** `model/` · `service/` 退款 outbox/intent 幂等与 CAS 单测；docs/ops evidence
**栈锁遵守：** Go · 三库 · JSON 仅 `common/json.go` · 未改 go.mod · 未 D7 · 未关 shadow · 未改 FRONTEND_MODE

---

## 1. 定位现有幂等/CAS 路径

| 路径 | 幂等/CAS 机制 |
|------|---------------|
| `service/refund_outbox.go` `processRefundIntent` | 钱包→订阅→令牌分步，各步 `*_done` 标记 + `IsRefundStepAlreadyDone` 幂等跳过 |
| `model/refund_intent.go` `CreateRefundIntentIfAbsent` | 按 `idempotency_key` 唯一键去重，冲突返回已有行（`created=false`） |
| `model/refund_intent.go` `refundIntentStepCAS` | `status NOT IN (succeeded,dead) AND <done>=false` 条件更新；`RowsAffected=0 → ErrRefundStepAlreadyDone` |
| `model/refund_intent.go` `ApplyRefundWalletStep / ApplyRefundTokenStep / ApplyRefundSubscriptionExtraStep` | 额度变更与 done 标记同事务 CAS，防崩溃/并发双倍回补 |
| `model/subscription.go` `RefundSubscriptionPreConsume` | 按 `request_id` 幂等：`status==refunded` 直接返回 no-op |
| `model/refund_intent.go` `ClaimRefundIntents` | 逐行条件更新认领，`RowsAffected=0` 表示被并发抢走 |

### 已有测试（改动前）
- `model/refund_intent_test.go`：`ApplyRefundWalletStep` / `ApplyRefundTokenStep` / `ApplyRefundSubscriptionExtraStep` 原子性+幂等；Wallet 终态回滚。
- `service/task_billing_test.go`：`TestCASGuardedRefund_Win/Lose`、`TestRefundTaskQuota_*`。

### 识别缺口
1. `CreateRefundIntentIfAbsent` 无直接单测锁「同 idempotency key 不重复插入/不覆盖原额度」。
2. `RefundSubscriptionPreConsume` 无直接单测锁「同 request_id 重复退款不双倍回补」及空 requestId 拒绝。

---

## 2. 补充最小表驱动单测

文件：`model/refund_intent_test.go`（+83 行）

- `TestCreateRefundIntentIfAbsent_DeduplicatesByIdempotencyKey`
  - 首次入队 `created=true` 且 `status=pending`；
  - 同键第二次不同 payload → `created=false`、返回原行、原额度保留（第二 payload 被忽略）；
  - `count(idempotency_key)==1` 锁唯一。
- `TestRefundSubscriptionPreConsume_IdempotentNoDoubleRefund`
  - 首次退款 `AmountUsed 200→150`、record `status=refunded`；
  - 同 requestId 再退为 no-op，`AmountUsed` 仍 150（不双倍回补）。
- `TestRefundSubscriptionPreConsume_EmptyRequestIdRejected`
  - 空/空白 requestId 返回 error。

均为 testify/require 断言、确定性输入/精确期望，遵守后端测试质量纪律。

---

## 3. 验证命令 + exit code

```
$ cd /d/TransitHub/src && go test ./model ./service -count=1
ok  	github.com/xvyimu/TransitHub/model	2.096s
ok  	github.com/xvyimu/TransitHub/service	2.076s
EXIT=0
```

聚焦验证（-run Refund|Outbox|Idempot -v）：

```
$ go test ./model ./service -run 'Refund|Outbox|Idempot' -count=1 -v
--- PASS: TestApplyRefundWalletStep_AtomicAndIdempotent
--- PASS: TestApplyRefundTokenStep_AtomicAndIdempotent
--- PASS: TestApplyRefundWalletStep_RollsBackQuotaIfDoneUpdateImpossible
--- PASS: TestApplyRefundSubscriptionExtraStep_AtomicAndIdempotent
--- PASS: TestCreateRefundIntentIfAbsent_DeduplicatesByIdempotencyKey        (含于 model 包)
--- PASS: TestRefundSubscriptionPreConsume_IdempotentNoDoubleRefund          (含于 model 包)
--- PASS: TestRefundSubscriptionPreConsume_EmptyRequestIdRejected            (含于 model 包)
ok  	github.com/xvyimu/TransitHub/model
--- PASS: TestRefundTaskQuota_Wallet / _Subscription / _ZeroQuota / _NoToken
--- PASS: TestCASGuardedRefund_Win / _Lose
ok  	github.com/xvyimu/TransitHub/service
EXIT=0
```

全量 `./model ./service` 未超时，已跑全量，无需 -run 裁剪；两种运行均如实记录。

---

## 4. 风险

单测仅在 SQLite（内存）跑，`RefundSubscriptionPreConsume` 的 `lockForUpdate` 在 SQLite 下不发 `FOR UPDATE`（正常降级），MySQL/PG 的真实行锁并发路径未在本单测覆盖；幂等的正确性由 `status==refunded` 短路保证，与锁无关。

---

**DONE · in-review · D7 NOT EXECUTED**

# Spec: B→O 目录同步 API（P1）

## POST /api/v1/integrations/catalog/sync

### INT-CAT-01 合法 upsert

- **Given** `INTEGRATION_ENABLED=true`，有效 service token
- **When** POST 1 条合法 device
- **Then** `code=0`，`accepted=1`，`catalog_version` 已更新

### INT-CAT-02 幂等 batch_id

- **Given** 相同 `batch_id` 重复提交
- **When** 第二次 POST
- **Then** 返回首次结果，不重复写入

### INT-CAT-03 缺 device_id

- **When** devices[0].device_id 为空
- **Then** `rejected_rows` 含 `invalid_field`

### INT-CAT-04 未授权

- **When** 无 token 或 token 错误
- **Then** HTTP 401，`code=1001`

### INT-CAT-05 与 CSV 等效（catalog 字段）

- **Given** 同一 fixtures/devices_valid.csv 内容
- **When** API sync 与 CSV import 各执行一次
- **Then** `catalog_version`、group、model、hw 一致

---

## 版本倒退

### INT-CAT-V01 version_rollback_ignored

- **Given** `reported_version=v2.4.0`（来自 prior report-status Success）
- **When** sync `current_version=v2.1.0`，默认 `version_policy=preserve_device_reported`
- **Then** `catalog_version=v2.1.0`，`reported_version` 仍为 v2.4.0
- **And** `warning_rows[0].code=version_rollback_ignored`

### INT-CAT-V02 trust_catalog 强制覆盖

- **Given** `reported_version=v2.4.0`
- **When** sync v2.1.0 且 `version_policy=trust_catalog` + 行级 `force_catalog_version=true`
- **Then** `reported_version=v2.1.0`，写审计

### INT-CAT-V03 version_ahead_of_device

- **Given** `reported_version=v2.3.0`
- **When** sync `current_version=v2.4.0`
- **Then** `catalog_version=v2.4.0`，`reported_version` 仍为 v2.3.0
- **And** 可选 warning `version_ahead_of_device`

### INT-CAT-V04 check-update 不受 catalog 虚高影响

- **Given** catalog v2.4，reported v2.3，Running 任务目标 v2.4
- **When** C check-update 带 `current_version=v2.3`
- **Then** 仍返回 has_update=true（不因 B 已认为 v2.4 而 already_latest）

---

## 身份不一致

### INT-CAT-I01 group 变更允许

- **Given** 注册表 group=org-1001
- **When** sync group=org-1002
- **Then** accepted，warning `identity_changed` 或静默更新

### INT-CAT-I02 model 变更拒绝

- **Given** product_model=V9
- **When** sync product_model=V10 且无 `force_identity_update`
- **Then** rejected，`identity_change_rejected`

### INT-CAT-I03 model 变更 force

- **When** sync V10 且 `force_identity_update=true`
- **Then** accepted，warning `identity_change_applied`

### INT-CAT-I04 快照不受 group 变更影响

- **Given** Running 任务快照含 device A（创建时 group=org-1001）
- **When** sync 将 A 改为 org-1002
- **Then** 快照仍含 A；check-update 仍按快照命中（identity 校验策略以实现为准）

---

## 乱序与多来源

### INT-CAT-O01 stale 行跳过

- **Given** 设备 catalog_synced_at=T1
- **When** sync 行带 `updated_at=T0`（T0<T1）
- **Then** 跳过，`warning stale_device_record`

### INT-CAT-O02 API 与 CSV 新者胜

- **Given** API sync at T1，CSV import at T2
- **When** 两者 catalog_version 不同
- **Then** T2 的 catalog 字段生效

### INT-CAT-O03 stale_batch_before

- **When** batch 内 `exported_at` 早于 `sync_options.stale_batch_before`
- **Then** 整批拒绝 `batch_out_of_order`

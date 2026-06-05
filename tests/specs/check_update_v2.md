# Spec: check-update v2（P0）

## CU-01 not_in_catalog

- **Given** device_id 不在 t_device
- **When** check-update
- **Then** HTTP 403，`code=2003`，`reason=not_in_catalog`

## CU-02 identity_mismatch

- **Given** 注册表 group=org-1001，请求 group=org-9999
- **When** check-update
- **Then** HTTP 409，`code=2006`

## CU-03 already_latest

- **Given** reported_version >= package.target_version
- **When** check-update
- **Then** `code=2001`，`reason=already_latest`，含 `retry_after_sec`

## CU-04 task snapshot

- **Given** Running 任务快照不含 device A
- **When** A 的 group/model/hw 与任务规则相同但不在快照
- **Then** `code=2001`，`reason=no_running_task`（或 not_in_snapshot）

## CU-05 Success 回写

- **Given** report-status Success，target_version=v2.4.0
- **When** 查 t_device
- **Then** reported_version=v2.4.0

## CU-06 retry_after_sec

- **When** 任意 2001 响应
- **Then** data.retry_after_sec > 0

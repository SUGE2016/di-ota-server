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

## CU-07 check-update 只读不写库

- **Given** 库内 reported_version=v2.4.0
- **When** 请求 current_version=v2.3.0（低于库内版本）
- **Then** 决策仍用 v2.4.0，`reason=already_latest`
- **And** 不产生 `UPDATE t_device`（不 TouchDeviceLastSeen / TouchDeviceReportedVersion）

## CU-08 download_url 公网 host

- **Given** S3 内网 endpoint=minio:9000，PublicBaseURL=localhost:9000
- **When** has_update=true
- **Then** data.download_url 不含 `minio:9000`，使用 PublicBaseURL

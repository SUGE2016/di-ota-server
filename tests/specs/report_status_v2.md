# Spec: report-status v2

## RS-01 Success 幂等仍回写 reported_version

- **Given** 幂等键已存在（重复 Success 上报）
- **And** 库内 reported_version=v2.3.0，请求 target_version=v2.4.0
- **When** POST report-status
- **Then** HTTP 200，返回缓存响应
- **And** 仍执行 `TouchDeviceReportedVersion`，reported_version=v2.4.0

## RS-02 Success 幂等版本已达标

- **Given** 幂等键已存在
- **And** 库内 reported_version 已 >= target_version
- **When** POST report-status Success
- **Then** HTTP 200，不重复 UPDATE t_device

## RS-03 非法状态迁移

- **Given** 当前状态 Success
- **When** 上报 Downloading
- **Then** HTTP 409，`code=2005`

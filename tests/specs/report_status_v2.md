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

## RS-03 Success 后中间态忽略

- **Given** 当前状态 Success
- **When** 上报 Downloading（宽松或严格）
- **Then** HTTP 200，`ignored=true`，库内状态仍为 Success

## RS-04 严格模式非法回退

- **Given** product_model 策略为 strict，当前状态 Upgrading
- **When** 上报 Downloading
- **Then** HTTP 409，`code=2005`

## RS-05 宽松模式 Failed→Success

- **Given** 默认宽松（无策略行），当前状态 Failed
- **And** task 对应包版本 = v2.4.0
- **When** 上报 Success + target_version=v2.4.0
- **Then** HTTP 200，状态变为 Success，回写 reported_version

## RS-06 Success 校验

- **Given** 上报 Success
- **When** 缺少 target_version，或与 task 包版本不一致，或 task 不存在
- **Then** HTTP 400，`code=1002`

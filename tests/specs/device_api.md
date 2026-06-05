# Spec: 设备端 API

## 背景

设备直连 `ota-server`：`POST /device/v1/check-update`、`POST /device/v1/report-status`。可选 Bearer 鉴权（`DEVICE_API_AUTH_ENABLED`）。

## 场景

### API-01 鉴权关闭

- **Given** `DEVICE_API_AUTH_ENABLED=false`
- **When** 无 Authorization 调用设备接口
- **Then** 请求继续处理（非 401）

### API-02 鉴权开启缺 Token

- **Given** `DEVICE_API_AUTH_ENABLED=true`
- **When** 无 Bearer
- **Then** HTTP 401，`code=1001`

### API-03 鉴权开启有效 Token

- **Given** token=`shared-token`
- **When** `Authorization: Bearer shared-token`
- **Then** 通过鉴权

### API-04 check-update 缺字段

- **Given** 请求体缺少 device_id
- **When** POST check-update
- **Then** HTTP 400

### API-05 check-update 无任务

- **Given** 无匹配 Running 任务
- **When** POST check-update
- **Then** HTTP 200，`code=2001`，`has_update=false`

### API-06 report-status 缺 task_id

- **Given** 合法鉴权
- **When** 缺少 task_id
- **Then** HTTP 400

### API-07 report-status 非法状态

- **Given** status=`unknown`
- **When** POST report-status
- **Then** HTTP 400，`invalid status`

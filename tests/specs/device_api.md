# Spec: 设备端 API

## 背景

设备直连 `ota-server`：`POST /device/v1/check-update`、`POST /device/v1/report-status`。

启用 `DEVICE_API_AUTH_ENABLED=true` 时，使用 **per-device HMAC**（`Authorization: Device device_id=...,timestamp=...,signature=...`），详见 `doc/device-integration-v9.md` §3。管理端 Bearer JWT **不**用于设备 API。

## 场景

### API-01 鉴权关闭

- **Given** `DEVICE_API_AUTH_ENABLED=false`
- **When** 无 Authorization 调用设备接口
- **Then** 请求继续处理（非 401）

### API-02 鉴权开启缺签名

- **Given** `DEVICE_API_AUTH_ENABLED=true`
- **When** 无 `Authorization: Device ...`
- **Then** HTTP 401，`code=1001`

### API-03 鉴权开启有效 HMAC

- **Given** 设备已 provision `device_secret`
- **When** 按 §3.3 构造合法 HMAC 请求头
- **Then** 通过鉴权

### API-04 secret 未 provision

- **Given** 设备在注册表但 `device_secret` 为空
- **When** 合法 HMAC 或鉴权关闭后调用
- **Then** HTTP 403，`code=2007`（鉴权阶段）

### API-05 check-update 缺字段

- **Given** 请求体缺少 device_id
- **When** POST check-update
- **Then** HTTP 400

### API-06 check-update 无任务

- **Given** 无匹配 Running 任务
- **When** POST check-update
- **Then** HTTP 200，`code=2001`，`has_update=false`

### API-07 report-status 缺 task_id

- **Given** 合法鉴权
- **When** 缺少 task_id
- **Then** HTTP 400

### API-08 report-status 非法状态

- **Given** status=`unknown`
- **When** POST report-status
- **Then** HTTP 400，`invalid status`

### API-09 check-update download_url

- **Given** 有可用升级且 S3 PublicBaseURL 已配置
- **When** POST check-update 返回 has_update=true
- **Then** download_url 使用公网 host（见 CU-08）

### API-10 管理端设备列表筛选

- **When** GET `/api/v1/devices?group=&product_model=&tag=&abnormal=true`
- **Then** 返回筛选结果、total 及 `secret_provisioned` 状态

### API-11 管理端 provision secret

- **When** `PUT /api/v1/devices/:device_id/device-secret` + Bearer JWT
- **Then** 写入 secret，响应 `provisioned=true`；列表/详情 `secret_provisioned=true`

### API-12 管理端设备升级历史

- **When** GET `/api/v1/devices/:device_id/upgrade-records`
- **Then** 返回该设备 report-status 写入的记录列表

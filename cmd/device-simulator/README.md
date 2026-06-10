# OTA Device Simulator

模拟设备 OTA 升级流程：**check-update** → **report-status**（可选下载探测）。

鉴权方式为 **per-device HMAC**（`device_secret`），与 `doc/device-integration-v9.md` §3 一致；**不使用** Bearer / OIDC token。

## Web 界面（推荐）

管理台同域访问，经 Vite/nginx 代理到 `ota-api`：

| 环境 | URL |
|------|-----|
| 本地 dev | `http://localhost:5173/#/simulator` |
| demogo | `https://demogo.work/ota/#/simulator` |

登录页底部也有「设备端模拟器」入口（无需登录管理台即可使用）。

**填写项：**

- `device_id` — 设备 SN（须在平台注册表）
- `device_secret` — 产线 secret（须先 provision，见下）
- `current_version` — 当前固件版本

**功能：**

- **检查更新**：`POST /device/v1/check-update`（仅 `device_id` + `current_version`）
- **完整升级流程**：check-update → 可选 download → pending → … → success
- 签名 path 固定为 `/device/v1/...`（不含 `/ota` 等控制台前缀）
- 右侧实时 JSON 日志

## CLI（可选）

```bash
go run ./cmd/device-simulator/main.go \
  -url http://localhost:8080 \
  -device-id AMS000001 \
  -device-secret test-secret-local-001 \
  -version v2.3.0 \
  -dry-run
```

`-device-secret` 为空且服务端 `DEVICE_API_AUTH_ENABLED=false` 时跳过签名（仅本地开发）。

## 前置条件

1. OTA 服务已启动（本地：`docker compose up -d`）
2. 设备已导入注册表（CSV 或 catalog/sync）
3. **已为该 SN provision `device_secret`**（管理端 `PUT /api/v1/devices/:id/device-secret`）
4. 若要测完整升级：存在 **Running** 任务且设备在任务快照中

### 本地快速 provision 示例

```bash
JWT=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123456"}' | jq -r '.data.access_token')

curl -s -X POST http://localhost:8080/api/v1/devices/import-csv \
  -H "Authorization: Bearer $JWT" -F "file=@tests/fixtures/devices_valid.csv"

curl -s -X PUT http://localhost:8080/api/v1/devices/AMS000001/device-secret \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{"device_secret":"test-secret-local-001"}'
```

## 相关文档

- 设备对接与 HMAC 算法：`doc/device-integration-v9.md`
- 部署与环境变量：`doc/deploy-manual.md`

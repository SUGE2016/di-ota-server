# OTA Server Monorepo

## 目录
- `backend/`: Go API 与 Worker
- `frontend/`: React 管理台
- `docker-compose.yml`: 本地一键启动
- `docker-compose.staging.yml`: 预发覆盖配置
- `.env.example`: 环境变量模板
- `doc/device-integration-v9.md`: **设备对接手册（HMAC 鉴权，权威）**
- `cmd/device-simulator/`: 设备端 CLI 模拟器

## 快速启动
1. 复制环境变量文件：
   - `cp .env.example .env`
   - 按注释填写密钥占位符（`<...>`）
2. 启动服务：
   - `docker compose up -d`
   - WSL 宿主机 5432 被占用时：`docker compose -f docker-compose.yml -f docker-compose.local.yml up -d`（Postgres 映射到 5433）
3. 访问地址：
   - API 健康检查：`http://localhost:8080/healthz`
   - 管理台：`http://localhost:5173`
   - 设备模拟器：`http://localhost:5173/#/simulator`
   - RabbitMQ 管理台：`http://localhost:15672`（guest/guest）
   - Keycloak 管理台：`http://localhost:18080`（admin/admin）
   - MinIO Console：`http://localhost:9001`（minioadmin/minioadmin）

## 本地联调默认配置
- ⚠️ 以下默认账号/密码仅用于开发联调，严禁用于生产环境。
- OIDC Provider：Keycloak（realm: `ota-dev`）
- OIDC Client：`ota-console`
- OIDC Client Secret：`ota-dev-client-secret`
- OIDC 测试用户：`dev-admin` / `Admin@123456`
- 本地账号登录（可选）：`LOCAL_AUTH_ENABLED=true` 时使用 `admin` / `Admin@123456`（见 `.env` bcrypt 配置）
- S3 兼容存储：MinIO
- 默认 Bucket：`ota-packages`（由 `minio-init` 自动创建）

## OIDC 与 S3 关键环境变量
- OIDC:
   - `OIDC_ENABLED=true`
   - `OIDC_ISSUER_URL=http://keycloak:8080/realms/ota-dev`
   - `OIDC_AUTHORIZE_URL=`（可选覆盖）
   - `OIDC_TOKEN_URL=`（可选覆盖）
   - `OIDC_USERINFO_URL=`（可选覆盖）
   - `OIDC_CLIENT_ID=ota-console`
   - `OIDC_CLIENT_SECRET=ota-dev-client-secret`
   - `OIDC_REDIRECT_URL=http://localhost:8080/api/v1/auth/sso/callback`
   - `OIDC_STATE_SIGNING_KEY=<state-signing-key>`（建议独立配置，未配置时回落 `JWT_SECRET`）
   - `OIDC_STATE_TTL_SEC=300`
- API 运行:
   - `API_AUTO_MIGRATE_ON_START=false`（默认关闭，建议通过迁移工具执行）
   - `DEVICE_API_AUTH_ENABLED=false`（本地默认关闭；**生产 / demogo 须为 `true`**）
   - `DEVICE_AUTH_TIMESTAMP_TOLERANCE_SEC=300`（HMAC 时间戳容忍窗，秒）
- Worker:
   - `WORKER_TASK_STATS_RETENTION_HOURS=168`（默认保留最近 7 天快照）
- Local Auth:
   - `LOCAL_AUTH_ENABLED=false`
   - `LOCAL_ADMIN_USERNAME=admin`
   - `LOCAL_ADMIN_PASSWORD_HASH=<bcrypt-hash>`
   - 密码需预先生成 bcrypt 哈希后填入该变量。
- S3:
   - `S3_ENDPOINT=http://minio:9000`
   - `S3_BUCKET=ota-packages`
   - `S3_ACCESS_KEY_ID=minioadmin`
   - `S3_SECRET_ACCESS_KEY=minioadmin`
   - `S3_PUBLIC_BASE_URL=http://localhost:9000/ota-packages`

## 设备鉴权（per-device HMAC）

设备 API（`/device/v1/*`）**不使用**管理台 Bearer / OIDC token，也**不再使用**共享 `DEVICE_API_TOKEN`。

| 角色 | 鉴权方式 |
|------|----------|
| 管理台 / 运维 | `Authorization: Bearer <JWT>`（登录或 SSO） |
| V9 设备 / 模拟器 | `Authorization: Device device_id=...,timestamp=...,signature=...` |

管理台角色（可组合）：

| 角色 | 能力 |
|------|------|
| `admin` | 全部能力 |
| `secret_admin` | 仅 device_secret provision（独立页面 + API） |
| `release` | 包/任务/设备目录，**无** secret 写入 |
| `readonly` / `audit` | 只读 |

流程概要：

1. **CSV 导入设备**（注册 SN，不含 secret）
2. **管理台 → 设备 Secret 管理**（`secret_admin` / `admin`）写入 secret，或 `PUT .../device-secret` / `POST .../device-secrets/import-csv`
3. 设备固件 / Web 模拟器 / CLI 模拟器使用**相同** secret 做 HMAC 签名

详细算法与错误码见 [`doc/device-integration-v9.md`](doc/device-integration-v9.md) §3。

## 当前已实现
- API 服务（Gin）
- Worker 统计任务（周期快照 + 超阈值自动回滚）
- 本地登录接口：`POST /api/v1/auth/login`（默认禁用，启用后校验数据库用户 + bcrypt）
- OIDC 开关接口：
   - `GET /api/v1/auth/sso/login`
   - `GET /api/v1/auth/sso/callback`
   - 回调已支持 state 校验 + code->token->userinfo 交换（未配置时可回落 Mock）
   - 默认本地 Provider：Keycloak（realm: `ota-dev`，client: `ota-console`）
- 管理端接口：
   - `GET /api/v1/ping`
   - `POST /api/v1/packages/upload-url`
   - `POST /api/v1/packages/complete`
   - `POST /api/v1/packages`
   - `GET /api/v1/devices`（支持筛选与分页；返回 `secret_provisioned` 状态）
   - `GET /api/v1/devices/:device_id`
   - `PUT /api/v1/devices/:device_id/device-secret`（provision HMAC secret）
   - `GET /api/v1/devices/:device_id/upgrade-records`
   - `GET /api/v1/devices/csv-template`
   - `POST /api/v1/devices/import-csv`
   - `POST /api/v1/release-tasks`
   - `POST /api/v1/release-tasks/:task_id/actions`（pause/resume/rollback）
   - `GET /api/v1/release-tasks/:task_id/audits`
- 设备端接口：
   - `POST /device/v1/check-update`（请求体：`device_id` + `current_version`；启用 HMAC 后须带 `Authorization: Device ...`）
   - 返回 S3/MinIO 预签名下载 URL（或 HMAC 代理 URL）
   - `POST /device/v1/report-status`（支持幂等键；要求 `task_id`，支持 `source_version`、`target_version`、`error_code`）
- 本地对象存储：MinIO（自动创建 bucket: `ota-packages`）
- `sqlc` 查询定义与生成代码（`backend/queries/ota.sql`、`backend/internal/store`）
- React 管理台（包/任务/设备/用户；设备详情页可 provision secret；告警仍为 Mock，详见 [`doc/管理台缺口清单.md`](doc/管理台缺口清单.md)）
- PostgreSQL 初始化脚本：`backend/migrations/001_init.sql`

## 安全密钥说明
- `JWT_SECRET`：管理端登录后签发 JWT 的签名密钥。
- `DEVICE_SIGNING_SECRET`：历史/内部签名相关配置（设备 API 现用 per-device `device_secret` + HMAC，见对接手册）。
- `OIDC_CLIENT_SECRET`：SSO 对接 OIDC 提供方的客户端密钥。
- `device_secret`：每台设备唯一，产线烧录 + 平台 provision；**不在列表/详情 API 中回显明文**。

## 架构改进（v2，2026-06-05）

在保留 **C 直连 O** 与 **O 升级决策权威** 前提下，允许 **B↔O 受控集成**（目录 sync + 待升级 hint），详见 [`doc/架构改进方案_v2_2026-06-05.md`](doc/架构改进方案_v2_2026-06-05.md)。

### 已实现（v2）

- P0：设备注册表校验、任务快照、`check-update` 拒绝码、Success 回写 `reported_version`
- P1：`POST /api/v1/integrations/catalog/sync`（含版本倒退/身份冲突策略）
- P2：`GET /api/v1/integrations/pending-upgrades`
- 设备模拟器：[`cmd/device-simulator/README.md`](cmd/device-simulator/README.md)（Web：`/#/simulator`；CLI 支持 `-device-secret`）

### 本地联调（v2）

```bash
# 1. 启动栈（需 API_AUTO_MIGRATE_ON_START=true 或手动执行迁移）
docker compose up -d

# 2. 管理端登录
JWT=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123456"}' | jq -r '.data.access_token')

# 3. 导入设备（CSV 不含 secret）
curl -H "Authorization: Bearer $JWT" -F file=@tests/fixtures/devices_valid.csv \
  http://localhost:8080/api/v1/devices/import-csv

# 4. Provision device_secret（管理台设备详情页亦可操作）
curl -X PUT -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{"device_secret":"test-secret-local-001"}' \
  http://localhost:8080/api/v1/devices/AMS000001/device-secret

# 5. 创建 Running 任务后，用模拟器走升级流程
go run ./cmd/device-simulator/main.go \
  -url http://localhost:8080 \
  -device-id AMS000001 \
  -device-secret test-secret-local-001 \
  -dry-run
```

Integration 环境变量见 `.env.example`：`INTEGRATION_ENABLED`、`INTEGRATION_SERVICE_TOKEN`。

## 第三方接入说明
1. `ota-server` 可作为独立第三方 OTA 服务对外提供 `/device/v1/check-update` 与 `/device/v1/report-status`。
2. 设备直接调用 `ota-server`，主系统不代理 OTA 请求、不共享 OTA 数据库。
3. OTA 任务选设备依赖 CSV 设备清单导入，模板字段为 `device_id,product_code,product_model,hardware_version,current_version,device_group,tags`（**不含** `device_secret`）。
4. 每台设备须单独 provision `device_secret`（管理台或 `PUT .../device-secret`）。
5. 设备升级检查成功后应保存 `task_id`，并在升级结果上报时回传 `task_id`。
6. 设备结果上报建议包含：`device_id`、`task_id`、`status`、`source_version`、`target_version`；失败场景补充 `error_code`。

## 下一步
1. 增加 OIDC 交换失败场景的细粒度错误码与重试策略。
2. 补充包上传、任务流转与设备上报的集成测试。

## 包上传最小联调流程
1. 调用 `POST /api/v1/packages/upload-url` 获取 `upload_url` 与 `package_id`。
2. 使用返回的 `upload_url` 直接 `PUT` 二进制到 MinIO/S3。
3. 调用 `POST /api/v1/packages/complete`，提交 `package_id`、版本信息、哈希、签名和文件大小。
4. 服务端会校验对象存在与大小（以及可用时的哈希元数据），通过后入库为可发布包。

## 常见问题
1. `sqlc` 命令找不到：
   - 使用 `$(go env GOPATH)/bin/sqlc` 执行，或把 `$(go env GOPATH)/bin` 加入 `PATH`。
2. `sqlc generate` 报 `sqlc.yaml does not exist`：
   - 需要在 `backend/` 目录执行生成命令。
3. Keycloak 或 MinIO 首次拉取较慢：
   - `docker compose up -d` 可能需要等待镜像拉取完成后再检查状态。
4. 设备 API 返回 `403 device_secret_not_provisioned`：
   - 在管理台设备详情页写入 secret，或调用 `PUT /api/v1/devices/:id/device-secret`。
5. 设备 API 返回 `401 unauthorized`：
   - 检查 HMAC 签名：path 为 `/device/v1/...`（不含 `/ota` 前缀）、timestamp 在容忍窗内、secret 与平台一致。详见 `doc/device-integration-v9.md` §3。

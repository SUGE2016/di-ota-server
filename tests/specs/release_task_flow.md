# Spec: 发布任务端到端（集成，待 Docker 环境）

## 背景

计划 2.0 P0 要求：**上传 → 发布 → 设备上报 → 暂停/回滚** 全链路可测。

## 场景（integration/e2e，标记 build tag `e2e`）

### E2E-01 包上传闭环

1. POST `/api/v1/packages/upload-url`
2. PUT 二进制到 MinIO
3. POST `/api/v1/packages/complete`
4. 包状态变为 Published

### E2E-02 CSV 导入 + 创建任务

1. 主系统导出 CSV（或 fixtures/devices_valid.csv）
2. POST `/api/v1/devices/import-csv`
3. POST `/api/v1/release-tasks`
4. 任务状态 Draft → Running

### E2E-03 设备升级闭环

1. POST `/device/v1/check-update` → 返回 task_id、download_url
2. POST `/device/v1/report-status` 顺序上报 Pending…Success
3. GET 任务详情可见设备状态

### E2E-04 任务暂停

1. Running 任务
2. POST `/api/v1/release-tasks/{id}/actions` action=pause
3. check-update 不再命中

### E2E-05 自动熔断

1. Worker 统计失败率超阈值
2. 任务自动 Paused

> 运行 `go test -tags=e2e ./tests/integration/` 前需 `docker compose up`；栈不可达时自动 skip。

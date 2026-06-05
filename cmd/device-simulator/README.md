# OTA Device Simulator

模拟设备 OTA 升级流程：**check-update** → **report-status**（可选下载探测）。

## Web 界面（推荐）

管理台同域访问，走 Vite 代理到 `ota-api`：

```text
http://localhost:5173/#/simulator
```

登录页底部也有「设备端模拟器」入口。

功能：

- 配置 device_id / group / model / hw / version / Bearer token
- **检查更新**：单次 `POST /device/v1/check-update`
- **完整升级流程**：check-update → 可选 download → pending → … → success
- 右侧实时 JSON 日志

## CLI（可选）

```bash
go run ./cmd/device-simulator/main.go \
  -url http://localhost:8080 \
  -device-id AMS000001 \
  -dry-run
```

## 前置条件

1. OTA 服务已启动
2. 设备已导入注册表（CSV 或 catalog/sync）
3. 存在 Running 任务且设备在任务快照中

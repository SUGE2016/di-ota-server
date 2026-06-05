# OTA Server 测试（TDD）

本目录存放 **TDD 行为规格** 与 **测试夹具**。可执行测试代码在 `backend/` 下（Go 不允许外部模块 import `internal` 包）。

## 目录结构

```
tests/
├── README.md           # 本文件
├── specs/              # Given-When-Then 行为规格（先写 spec，再写测试）
└── fixtures/           # CSV 等共享夹具

backend/
├── internal/server/*_test.go     # 单元测试
├── tests/integration/*_test.go   # HTTP + sqlmock 集成测试
└── cmd/worker/main_test.go       # Worker 测试
```

## TDD 工作流

1. **Red**：在 `specs/` 写场景 → 在 `backend/.../*_test.go` 写失败测试
2. **Green**：实现最小代码使测试通过
3. **Refactor**：重构并保持测试绿

## 运行测试

```bash
# 全部单元 + 集成
cd backend && go test ./...

# 仅集成测试
cd backend && go test ./tests/integration/...

# E2E 占位（需 docker compose）
cd backend && go test -tags=e2e ./tests/integration/ -run TestE2E -v
```

## 覆盖矩阵

| 模块 | Spec | 单元测试 | 集成测试 |
|------|------|----------|----------|
| 设备 CSV | specs/device_csv.md | internal/server/device_csv*_test.go | — |
| 设备 API | specs/device_api.md | internal/server/router_state_test.go | tests/integration/device_auth_test.go |
| 升级状态机 | specs/upgrade_state_machine.md | upgrade_status_test.go, canary_test.go | device_report_status_test.go |
| 发布全链路 | specs/release_task_flow.md | — | e2e_flow_test.go（占位） |
| B→O 目录同步 | specs/integration_catalog.md | — | 待 P1 实现 |
| 待升级 hint | specs/integration_hint.md | — | 待 P2 实现 |
| check-update v2 | specs/check_update_v2.md | — | 待 P0 实现 |
| 健康检查 | — | — | health_test.go |
| Worker 统计 | — | cmd/worker/main_test.go | — |

## 夹具

- `fixtures/devices_valid.csv` — 合法样例
- `fixtures/devices_invalid_header.csv` — 表头错误
- `fixtures/devices_duplicate.csv` — 重复 device_id
- `fixtures/devices_missing_required.csv` — 缺必填字段

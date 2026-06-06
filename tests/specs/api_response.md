# Spec: 管理端 API JSON 响应

管理端列表/详情 API 不得将 `sql.NullString` / `sql.NullTime` 结构体直接序列化为 `{String,Valid}`。

## AR-01 发布任务列表 version

- **Given** 任务关联包 version=v2.4.0
- **When** GET /api/v1/release-tasks
- **Then** `data[].version` 为 plain string `"v2.4.0"`

## AR-02 用户列表 last_login_at 为空

- **Given** 用户从未登录（last_login_at IS NULL）
- **When** GET /api/v1/users
- **Then** `data.users[].last_login_at` 为 JSON `null`，非对象

## AR-03 审计日志 schedule_time 类字段

- **Given** 任务 schedule_time 为空
- **When** GET 任务详情或审计
- **Then** 时间字段为 `null` 或 ISO 字符串，非 `{Time,Valid}`

# Spec: 发布任务设备快照

Running 任务需将匹配 scope 的设备写入 `t_task_target`，供 check-update 按快照匹配。

## TS-01 创建 Running 任务

- **Given** scope 内存在 active 设备
- **When** POST 创建任务且初始 state=Running
- **Then** 调用 `buildTaskSnapshot`，写入 t_task_target

## TS-02 resume 补快照

- **Given** 任务 state=Paused
- **When** POST action=resume
- **Then** state→Running，再次 `buildTaskSnapshot`

## TS-03 零设备退回 Draft

- **Given** scope 内无 active 设备
- **When** 创建 Running 任务
- **Then** 任务 state 退回 Draft，HTTP 400

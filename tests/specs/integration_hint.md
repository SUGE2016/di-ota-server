# Spec: 待升级清单与 ota_hint（P2）

## GET /api/v1/integrations/pending-upgrades

### INT-HINT-01 与 check-update 一致

- **Given** Running 任务 + 快照含 device A，A 命中 canary 且版本可升
- **When** GET pending-upgrades 与 POST check-update
- **Then** 同一 `task_id`、`target_version`

### INT-HINT-02 已成功设备不出现

- **Given** device A 对该 task 已 Success
- **When** GET pending-upgrades
- **Then** items 不含 A

### INT-HINT-03 游标分页

- **Given** items > limit
- **When** 用返回 cursor 请求下一页
- **Then** 无重复 device_id

### INT-HINT-04 hint 过期

- **Given** `hint_expires_at` 已过期
- **When** B 仍缓存旧 hint
- **Then** C check-update 以 O 实时决策为准（可升或不可升）

## B 心跳 ota_hint（契约）

### INT-HINT-05 字段完整

- **When** need_upgrade=true
- **Then** 必含 task_id、target_version、hint_expires_at

### INT-HINT-06 hint 不替代 check-update

- **Given** B 返回 need_upgrade=true
- **When** C 仅升级未调 O
- **Then** 不允许（SDK 必须 check-update + download + report）

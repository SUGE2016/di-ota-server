# Spec: 设备管理（管理台 + API）

## DM-01 设备列表组合筛选

- **Given** 设备注册表有多条记录
- **When** GET `/api/v1/devices?group=&product_model=&tag=&search=&abnormal=`
- **Then** 返回匹配子集与 `total` 计数一致

## DM-02 异常设备快速过滤

- **Given** 设备 `eligibility_state=blocked` 或 `inconsistency_flags` 非空
- **When** GET `/api/v1/devices?abnormal=true`
- **Then** 仅返回异常设备

## DM-03 设备升级历史

- **Given** 设备存在 `t_upgrade_record`
- **When** GET `/api/v1/devices/:device_id/upgrade-records`
- **Then** 按 `created_at DESC` 返回记录，含 status / task_id / 版本字段

## DM-04 详情页时间线

- **Given** 管理台打开设备详情
- **When** 页面加载
- **Then** 展示注册表概览 + 升级历史 Timeline

## DM-05 设备不存在

- **When** GET 未知 device_id 的 upgrade-records
- **Then** HTTP 404，`code=2004`

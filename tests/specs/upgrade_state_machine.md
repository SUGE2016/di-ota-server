# Spec: 设备升级状态机

## 背景

`report-status` 使用 `normalizeUpgradeStatus` 归一化状态，`canTransitUpgradeStatus` 校验转移。

## 正向路径

```
Pending → Downloading → Downloaded → Verifying → Upgrading → Success
```

## 场景

### SM-01 别名归一化

| 输入 | 期望 |
|------|------|
| `downloading` | Downloading |
| `download_success` | Downloaded |
| `upgrade_success` | Success |
| `rolling_back` | Rollbacking |

### SM-02 非法状态

- **When** status=`not-a-state`
- **Then** normalize 失败

### SM-03 首次上报

- **Given** 无历史记录（prev 空）
- **When** 任意合法 next
- **Then** 允许

### SM-04 重复上报

- **Given** prev=Downloading
- **When** next=Downloading
- **Then** 允许（幂等）

### SM-05 顺序前进

- **Given** prev=Downloading
- **When** next=Downloaded
- **Then** 允许

### SM-06 顺序倒退

- **Given** prev=Upgrading
- **When** next=Downloading
- **Then** 拒绝

### SM-07 终态不可变

- **Given** prev=Success
- **When** next=Failed
- **Then** 拒绝

### SM-08 失败后可回滚

- **Given** prev=Failed
- **When** next=Rollbacking
- **Then** 允许

### SM-09 任意阶段可 Failed

- **Given** prev=Verifying
- **When** next=Failed
- **Then** 允许

## 灰度（canary）

### CN-01 0% 不命中

- **When** canaryPercent=0
- **Then** false

### CN-02 100% 全命中

- **When** canaryPercent=100
- **Then** true

### CN-03 确定性

- **Given** 固定 taskID + deviceID + percent
- **When** 多次计算
- **Then** 结果一致

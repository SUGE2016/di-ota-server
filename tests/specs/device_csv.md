# Spec: 设备 CSV 导入

## 背景

主系统通过 CSV 离线同步设备影子表，字段：`device_id,product_code,product_model,hardware_version,current_version,device_group,tags`。

## 场景

### CSV-01 合法导入

- **Given** 表头正确、必填字段齐全、tags 为 JSON object
- **When** 调用 `parseDeviceCSV`
- **Then** 返回 1 行数据，无 rowErrors

### CSV-02 表头错误

- **Given** 表头缺少字段
- **When** 解析 CSV
- **Then** row[1] 报错 `header must be ...`

### CSV-03 重复 device_id

- **Given** 同一 CSV 内两行相同 device_id
- **When** 解析
- **Then** 保留首行，第二行报 `duplicate device_id`

### CSV-04 tags 非 object

- **Given** tags 为 `[]`
- **When** 解析
- **Then** 报错 `tags must be a JSON object`

### CSV-05 缺必填字段

- **Given** product_code 为空
- **When** 解析
- **Then** 报错 `device_id/product_code/... are required`

### CSV-06 默认分组

- **Given** device_group 为空
- **When** 解析成功
- **Then** device_group 设为 `default`

### CSV-07 空文件

- **Given** 无数据行
- **When** 解析
- **Then** 报错 `empty csv`

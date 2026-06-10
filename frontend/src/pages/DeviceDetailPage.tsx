import { useEffect, useState } from 'react';
import { Alert, Breadcrumb, Button, Card, Descriptions, Form, Input, Space, Tag, Timeline, Typography, message } from 'antd';
import { ApiOutlined, KeyOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { DeviceCatalogItem, UpgradeRecord, deviceAPI } from '../api';

const { Paragraph, Title } = Typography;

const statusColor: Record<string, string> = {
  Success: 'green',
  Failed: 'red',
  Downloading: 'blue',
  Applying: 'blue',
  Pending: 'default',
};

const FLAG_LABELS: Record<string, string> = {
  version_rollback_ignored: '目录同步版本低于 OTA 上报，已忽略回退',
  version_ahead_of_device: '目录版本高于设备 OTA 上报',
  identity_changed: '产品型号或硬件版本与历史记录不一致',
};

export function DeviceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [device, setDevice] = useState<DeviceCatalogItem | null>(null);
  const [records, setRecords] = useState<UpgradeRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [secretForm] = Form.useForm<{ device_secret: string }>();
  const [savingSecret, setSavingSecret] = useState(false);

  const loadDevice = async (deviceId: string) => {
    const [dev, history] = await Promise.all([
      deviceAPI.get(deviceId),
      deviceAPI.upgradeRecords(deviceId),
    ]);
    setDevice(dev);
    setRecords(history.records);
  };

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    loadDevice(id)
      .catch((e: Error) => message.error(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleSaveSecret = async (values: { device_secret: string }) => {
    if (!id) return;
    const secret = values.device_secret.trim();
    setSavingSecret(true);
    try {
      await deviceAPI.setSecret(id, secret);
      setDevice((prev) => (prev ? { ...prev, secret_provisioned: true } : prev));
      message.success('device_secret 已写入平台。请复制下方 secret 到设备固件或模拟器（页面不会再次显示）。');
    } catch (e) {
      message.error(e instanceof Error ? e.message : String(e));
    } finally {
      setSavingSecret(false);
    }
  };

  const handleGenerateSecret = () => {
    const bytes = new Uint8Array(24);
    crypto.getRandomValues(bytes);
    const secret = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
    secretForm.setFieldsValue({ device_secret: secret });
    message.info('已生成随机 secret，请点击「保存到平台」后复制到模拟器');
  };

  if (loading) {
    return null;
  }

  if (!device) {
    return (
      <div className="ota-page">
        <Card className="ota-card">
          <Space direction="vertical">
            <Title level={4}>设备不存在</Title>
            <Button onClick={() => navigate('/devices')}>返回设备管理</Button>
          </Space>
        </Card>
      </div>
    );
  }

  const tagEntries = device.tags && typeof device.tags === 'object'
    ? Object.entries(device.tags as Record<string, unknown>)
    : [];
  const flags = device.inconsistency_flags ?? [];

  return (
    <div className="ota-page">
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => navigate('/devices')}>设备管理</a> },
          { title: device.device_id },
        ]}
      />

      {(device.eligibility_state === 'blocked' || flags.length > 0) && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="设备存在异常标记"
          description={
            <Space direction="vertical" size={4}>
              {device.eligibility_state === 'blocked' && <span>升级资格：已阻断（blocked）</span>}
              {flags.length > 0 && (
                <span>
                  目录冲突：
                  {flags.map((flag) => (
                    <Tag key={flag} color="orange" style={{ marginInlineStart: 8 }}>{flag}</Tag>
                  ))}
                </span>
              )}
            </Space>
          }
        />
      )}

      {(flags.length > 0 || device.catalog_version || device.reported_version) && (
        <Card className="ota-card" title="目录与 OTA 版本对比" style={{ marginBottom: 16 }}>
          <Descriptions bordered column={{ xs: 1, sm: 3 }} size="small">
            <Descriptions.Item label="目录版本 (catalog)">{device.catalog_version || '-'}</Descriptions.Item>
            <Descriptions.Item label="OTA 上报 (reported)">{device.reported_version || '-'}</Descriptions.Item>
            <Descriptions.Item label="决策用当前版本">{device.current_version || '-'}</Descriptions.Item>
          </Descriptions>
          {flags.length > 0 && (
            <div style={{ marginTop: 12 }}>
              <Space direction="vertical" size={4}>
                {flags.map((flag) => (
                  <span key={flag}>
                    <Tag color="orange">{flag}</Tag>
                    <Typography.Text type="secondary">{FLAG_LABELS[flag] ?? '目录同步冲突标记'}</Typography.Text>
                  </span>
                ))}
              </Space>
            </div>
          )}
        </Card>
      )}

      <Card className="ota-card" title="设备鉴权 Secret (HMAC)" style={{ marginBottom: 16 }}>
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message="CSV 导入不含 device_secret"
          description="设备 CSV 仅注册 SN 与型号信息。每台设备须在此单独 provision secret，并与产线烧录 / 模拟器填写值一致。平台不会回显已有 secret。"
        />
        <Descriptions bordered column={{ xs: 1, sm: 2 }} size="small" style={{ marginBottom: 16 }}>
          <Descriptions.Item label="Provision 状态">
            <Tag color={device.secret_provisioned ? 'green' : 'orange'}>
              {device.secret_provisioned ? '已配置' : '未配置'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="设备 API 鉴权">
            Authorization: Device device_id=...,timestamp=...,signature=...
          </Descriptions.Item>
        </Descriptions>
        <Form form={secretForm} layout="vertical" onFinish={handleSaveSecret}>
          <Form.Item
            name="device_secret"
            label="写入 / 更新 device_secret"
            rules={[
              { required: true, message: '请输入 device_secret' },
              { max: 128, message: '最长 128 字符' },
            ]}
            extra="保存成功后请立即复制；刷新页面或离开后将无法从平台回读明文。"
          >
            <Input.Password placeholder="输入新 secret（保存后覆盖旧值）" visibilityToggle />
          </Form.Item>
          <Space wrap>
            <Button icon={<KeyOutlined />} onClick={handleGenerateSecret}>
              生成随机 Secret
            </Button>
            <Button type="primary" htmlType="submit" loading={savingSecret}>
              保存到平台
            </Button>
            <Button icon={<ApiOutlined />} onClick={() => navigate('/simulator')}>
              打开设备模拟器
            </Button>
          </Space>
        </Form>
      </Card>

      <Card className="ota-card" title="设备概览">
        <Descriptions bordered column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label="设备 ID">{device.device_id}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={device.eligibility_state === 'active' ? 'green' : 'red'}>{device.eligibility_state}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="产品代码">{device.product_code}</Descriptions.Item>
          <Descriptions.Item label="产品型号">{device.product_model}</Descriptions.Item>
          <Descriptions.Item label="硬件版本">{device.hardware_version}</Descriptions.Item>
          <Descriptions.Item label="设备分组">{device.device_group}</Descriptions.Item>
          <Descriptions.Item label="当前版本">{device.current_version || '-'}</Descriptions.Item>
          <Descriptions.Item label="OTA 上报版本">{device.reported_version || '-'}</Descriptions.Item>
          <Descriptions.Item label="目录版本">{device.catalog_version || '-'}</Descriptions.Item>
          <Descriptions.Item label="目录来源">{device.catalog_source || '-'}</Descriptions.Item>
          <Descriptions.Item label="最后心跳">
            {device.last_heartbeat ? new Date(device.last_heartbeat).toLocaleString() : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="注册时间">
            {device.registered_at ? new Date(device.registered_at).toLocaleString() : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="标签" span={2}>
            <Space size={[8, 8]} wrap>
              {tagEntries.length === 0 ? '-' : tagEntries.map(([k, v]) => (
                <span key={k} className="ota-list-chip">{k}: {String(v)}</span>
              ))}
            </Space>
          </Descriptions.Item>
        </Descriptions>
        <div style={{ marginTop: 16 }}>
          <Button onClick={() => navigate('/devices')}>返回列表</Button>
        </div>
      </Card>

      <Card className="ota-card" title="升级历史" style={{ marginTop: 16 }}>
        {records.length === 0 ? (
          <Paragraph type="secondary">暂无升级记录（设备上报 report-status 后将出现在此）。</Paragraph>
        ) : (
          <Timeline
            items={records.map((record) => ({
              color: statusColor[record.status] ?? 'gray',
              children: (
                <Space direction="vertical" size={2}>
                  <Space wrap>
                    <Tag color={statusColor[record.status] ?? 'default'}>{record.status}</Tag>
                    <span>{new Date(record.created_at).toLocaleString()}</span>
                  </Space>
                  <span>
                    任务 <Button type="link" size="small" onClick={() => navigate(`/tasks/${record.task_id}`)}>{record.task_id}</Button>
                  </span>
                  {(record.source_version || record.target_version) && (
                    <span>
                      版本：{record.source_version || '-'} → {record.target_version || '-'}
                    </span>
                  )}
                  {record.error_code && <Tag color="red">{record.error_code}</Tag>}
                </Space>
              ),
            }))}
          />
        )}
      </Card>
    </div>
  );
}

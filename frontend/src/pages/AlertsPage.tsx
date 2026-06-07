import { useEffect, useState } from 'react';
import { Button, Card, Select, Space, Table, Tag, Tooltip, Typography, message } from 'antd';
import { CheckOutlined, ReloadOutlined, StopOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { AlertItem, alertAPI } from '../api';

const { Paragraph, Title } = Typography;

const severityColor: Record<string, string> = {
  critical: 'red',
  warning: 'orange',
  info: 'blue',
};

const statusColor: Record<string, string> = {
  open: 'red',
  acknowledged: 'gold',
  closed: 'green',
};

export function AlertsPage() {
  const navigate = useNavigate();
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');
  const [selected, setSelected] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const load = async () => {
    setLoading(true);
    try {
      const data = await alertAPI.list({
        limit: pageSize,
        offset: (page - 1) * pageSize,
        status: statusFilter,
        severity: severityFilter,
      });
      setAlerts(data.alerts);
      setTotal(data.total);
      setSelected([]);
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载告警失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [page, statusFilter, severityFilter]);

  const runAction = async (action: 'acknowledge' | 'close') => {
    if (selected.length === 0) {
      message.warning('请先选择告警');
      return;
    }
    try {
      const res = await alertAPI.batchAction(action, selected);
      message.success(`已更新 ${res.updated} 条告警`);
      await load();
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '操作失败');
    }
  };

  const columns = [
    { title: '告警类型', dataIndex: 'alert_type', key: 'alert_type', width: 140, ellipsis: true },
    {
      title: '级别',
      dataIndex: 'severity',
      key: 'severity',
      width: 90,
      render: (value: string) => <Tag color={severityColor[value]}>{value}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value: string) => <Tag color={statusColor[value]}>{value}</Tag>,
    },
    {
      title: '资源',
      key: 'resource',
      width: 200,
      ellipsis: true,
      render: (_: unknown, r: AlertItem) => {
        const label = `${r.resource_type}/${r.resource_id}`;
        return (
          <Tooltip title={label}>
            <Typography.Link
              ellipsis
              style={{ display: 'block', maxWidth: '100%' }}
              onClick={() => {
                if (r.resource_type === 'task') navigate(`/tasks/${r.resource_id}`);
                else if (r.resource_type === 'device') navigate(`/devices/${r.resource_id}`);
              }}
            >
              {label}
            </Typography.Link>
          </Tooltip>
        );
      },
    },
    { title: '说明', dataIndex: 'message', key: 'message', ellipsis: true },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170, render: (v: string) => new Date(v).toLocaleString() },
  ];

  return (
    <div className="ota-page">
      <div>
        <Title level={3} className="ota-page-title">告警中心</Title>
        <Paragraph className="ota-page-subtitle">任务熔断与设备升级异常事件，支持确认与关闭。</Paragraph>
      </div>

      <Card className="ota-card">
        <div className="ota-toolbar">
          <div className="ota-toolbar-left">
            <Select
              className="ota-toolbar-control-select"
              allowClear
              placeholder="全部状态"
              value={statusFilter || undefined}
              options={[
                { label: 'open', value: 'open' },
                { label: 'acknowledged', value: 'acknowledged' },
                { label: 'closed', value: 'closed' },
              ]}
              onChange={(v) => { setStatusFilter(v ?? ''); setPage(1); }}
            />
            <Select
              className="ota-toolbar-control-select"
              allowClear
              placeholder="全部级别"
              value={severityFilter || undefined}
              options={[
                { label: 'critical', value: 'critical' },
                { label: 'warning', value: 'warning' },
                { label: 'info', value: 'info' },
              ]}
              onChange={(v) => { setSeverityFilter(v ?? ''); setPage(1); }}
            />
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void load()}>刷新</Button>
            <Button icon={<CheckOutlined />} onClick={() => void runAction('acknowledge')}>批量确认</Button>
            <Button icon={<StopOutlined />} onClick={() => void runAction('close')}>批量关闭</Button>
          </Space>
        </div>

        <Table
          rowKey="alert_id"
          tableLayout="fixed"
          rowSelection={{ selectedRowKeys: selected, onChange: (keys) => setSelected(keys as string[]) }}
          columns={columns}
          dataSource={alerts}
          loading={loading}
          pagination={{ current: page, pageSize, total, showSizeChanger: false, onChange: setPage }}
          scroll={alerts.length > 0 ? { x: 1100 } : undefined}
          locale={{ emptyText: '暂无告警事件' }}
        />
      </Card>
    </div>
  );
}

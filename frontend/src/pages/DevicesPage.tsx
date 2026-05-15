import { useEffect, useState } from 'react';
import { Alert, Button, Card, Input, Select, Space, Table, Tag, Typography, Upload, message } from 'antd';
import { DownloadOutlined, ReloadOutlined, UploadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { DeviceCatalogItem, deviceAPI } from '../api';

const { Paragraph, Title, Text } = Typography;

export function DevicesPage() {
  const navigate = useNavigate();
  const [devices, setDevices] = useState<DeviceCatalogItem[]>([]);
  const [keyword, setKeyword] = useState('');
  const [groupFilter, setGroupFilter] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [importing, setImporting] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const data = await deviceAPI.list(200, 0);
      setDevices(data.devices);
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const handleDownloadTemplate = async () => {
    try {
      const res = await deviceAPI.downloadTemplate();
      const url = URL.createObjectURL(new Blob([res.data], { type: 'text/csv;charset=utf-8' }));
      const link = document.createElement('a');
      link.href = url;
      link.download = 'ota-device-template.csv';
      link.click();
      URL.revokeObjectURL(url);
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const handleImport = async (file: File) => {
    setImporting(true);
    try {
      const result = await deviceAPI.importCSV(file);
      message.success(`导入成功：${result.imported_count} 台设备`);
      load();
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setImporting(false);
    }
  };

  const columns = [
    { title: '设备 ID', dataIndex: 'device_id', key: 'device_id' },
    { title: '产品代码', dataIndex: 'product_code', key: 'product_code' },
    { title: '型号', dataIndex: 'product_model', key: 'product_model' },
    { title: '硬件版本', dataIndex: 'hardware_version', key: 'hardware_version' },
    { title: '设备分组', dataIndex: 'device_group', key: 'device_group', render: (value: string) => <Tag color="blue">{value}</Tag> },
    { title: '当前版本', dataIndex: 'current_version', key: 'current_version' },
    { title: '导入/更新时间', dataIndex: 'last_heartbeat', key: 'last_heartbeat', render: (value: string) => value ? new Date(value).toLocaleString() : '-' },
    {
      title: '操作',
      key: 'actions',
      fixed: 'right' as const,
      render: (_: unknown, record: DeviceCatalogItem) => (
        <Button type="link" onClick={() => navigate(`/devices/${record.device_id}`)}>
          查看详情
        </Button>
      ),
    },
  ];

  const groupOptions = Array.from(new Set(devices.map((d) => d.device_group).filter(Boolean)))
    .map((group) => ({ label: group, value: group }));

  const filteredDevices = devices.filter((device) => {
    if (groupFilter && device.device_group !== groupFilter) return false;
    if (!keyword.trim()) return true;
    const key = keyword.toLowerCase();
    return [device.device_id, device.product_code, device.product_model, device.hardware_version, device.current_version, device.device_group]
      .join(' ')
      .toLowerCase()
      .includes(key);
  });

  return (
    <div className="ota-page">
      <div>
        <Title level={3} className="ota-page-title">设备管理</Title>
        <Paragraph className="ota-page-subtitle">通过 CSV 导入第三方设备清单，作为 OTA 任务选设备的影子目录。</Paragraph>
      </div>

      <Alert
        type="info"
        showIcon
        message="设备目录来自 CSV"
        description="此处只保存 OTA 选型与分组需要的最小字段，不同步主系统隐私或业务数据。"
      />

      <Card className="ota-card">
        <div className="ota-toolbar">
          <div className="ota-toolbar-left">
            <Input.Search
              className="ota-toolbar-control-search"
              placeholder="搜索设备 ID / 型号 / 产品代码"
              allowClear
              onSearch={setKeyword}
              onChange={(e) => setKeyword(e.target.value)}
            />
            <Select
              className="ota-toolbar-control-select"
              allowClear
              placeholder="设备分组"
              value={groupFilter}
              options={groupOptions}
              onChange={(value) => setGroupFilter(value)}
            />
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={load}>刷新</Button>
            <Button icon={<DownloadOutlined />} onClick={handleDownloadTemplate}>模板</Button>
            <Upload accept=".csv,text/csv" showUploadList={false} beforeUpload={(file) => { handleImport(file); return false; }}>
              <Button type="primary" icon={<UploadOutlined />} loading={importing}>导入 CSV</Button>
            </Upload>
          </Space>
        </div>

        <Table
          rowKey="device_id"
          columns={columns}
          dataSource={filteredDevices}
          loading={loading}
          pagination={{ pageSize: 10, showSizeChanger: false }}
          scroll={filteredDevices.length > 0 ? { x: 1200 } : undefined}
        />
      </Card>

      <div className="ota-section-grid">
        <Card className="ota-card ota-section-span-8" title="设备视角需要继续补的区域">
          <div className="ota-stack">
            <Text>设备详情页接入真实升级历史</Text>
            <Text>升级历史时间线</Text>
            <Text>异常设备快速过滤</Text>
            <Text>按产品型号 / 标签 / 分组的组合查询</Text>
          </div>
        </Card>
        <Card className="ota-card ota-section-span-4" title="当前页面重点">
          <div className="ota-stack">
            <Text>CSV 导入设备影子目录</Text>
            <Text>按分组筛选发布范围</Text>
            <Text>避免依赖主系统在线接口</Text>
          </div>
        </Card>
      </div>
    </div>
  );
}
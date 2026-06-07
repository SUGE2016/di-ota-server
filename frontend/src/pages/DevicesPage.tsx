import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Input, Select, Space, Switch, Table, Tag, Typography, Upload, message } from 'antd';
import { DownloadOutlined, ReloadOutlined, UploadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { DeviceCatalogItem, deviceAPI } from '../api';
import { tableActionColumn, listTableProps, listTableScroll } from '../utils/tableActionColumn';

const { Paragraph, Title, Text } = Typography;

export function DevicesPage() {
  const navigate = useNavigate();
  const [devices, setDevices] = useState<DeviceCatalogItem[]>([]);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState('');
  const [appliedKeyword, setAppliedKeyword] = useState('');
  const [groupFilter, setGroupFilter] = useState<string | undefined>(undefined);
  const [modelFilter, setModelFilter] = useState<string | undefined>(undefined);
  const [tagFilter, setTagFilter] = useState('');
  const [appliedTagFilter, setAppliedTagFilter] = useState('');
  const [abnormalOnly, setAbnormalOnly] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [loading, setLoading] = useState(true);
  const [importing, setImporting] = useState(false);
  const [filterOptions, setFilterOptions] = useState<{ groups: string[]; models: string[] }>({ groups: [], models: [] });

  const loadFilterOptions = useCallback(async () => {
    try {
      const data = await deviceAPI.list({ limit: 200, offset: 0 });
      setFilterOptions({
        groups: Array.from(new Set(data.devices.map((d) => d.device_group).filter(Boolean))),
        models: Array.from(new Set(data.devices.map((d) => d.product_model).filter(Boolean))),
      });
    } catch {
      /* ignore */
    }
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await deviceAPI.list({
        limit: pageSize,
        offset: (page - 1) * pageSize,
        search: appliedKeyword.trim(),
        group: groupFilter,
        product_model: modelFilter,
        tag: appliedTagFilter.trim(),
        abnormal: abnormalOnly,
      });
      setDevices(data.devices);
      setTotal(data.total);
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载设备失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, appliedKeyword, groupFilter, modelFilter, appliedTagFilter, abnormalOnly]);

  useEffect(() => { void loadFilterOptions(); }, [loadFilterOptions]);
  useEffect(() => { void load(); }, [load]);

  const handleDownloadTemplate = async () => {
    try {
      const res = await deviceAPI.downloadTemplate();
      const url = URL.createObjectURL(new Blob([res.data], { type: 'text/csv;charset=utf-8' }));
      const link = document.createElement('a');
      link.href = url;
      link.download = 'ota-device-template.csv';
      link.click();
      URL.revokeObjectURL(url);
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '下载模板失败');
    }
  };

  const handleImport = async (file: File) => {
    setImporting(true);
    try {
      const result = await deviceAPI.importCSV(file);
      message.success(`导入成功：${result.imported_count} 台设备`);
      setPage(1);
      await loadFilterOptions();
      await load();
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '导入失败');
    } finally {
      setImporting(false);
    }
  };

  const applyFilters = () => {
    setAppliedKeyword(keyword);
    setAppliedTagFilter(tagFilter);
    setPage(1);
  };

  const columns = [
    { title: '设备 ID', dataIndex: 'device_id', key: 'device_id' },
    { title: '产品代码', dataIndex: 'product_code', key: 'product_code' },
    { title: '型号', dataIndex: 'product_model', key: 'product_model' },
    { title: '硬件版本', dataIndex: 'hardware_version', key: 'hardware_version' },
    { title: '设备分组', dataIndex: 'device_group', key: 'device_group', render: (value: string) => <Tag color="blue">{value}</Tag> },
    { title: '当前版本', dataIndex: 'current_version', key: 'current_version' },
    { title: 'OTA 上报', dataIndex: 'reported_version', key: 'reported_version', render: (v: string) => v || '-' },
    {
      title: '状态',
      key: 'status',
      render: (_: unknown, record: DeviceCatalogItem) => {
        const flags = record.inconsistency_flags ?? [];
        if (record.eligibility_state === 'blocked') {
          return <Tag color="red">已阻断</Tag>;
        }
        if (flags.length > 0) {
          return <Tag color="orange">目录异常</Tag>;
        }
        return <Tag color="green">正常</Tag>;
      },
    },
    { title: '导入/更新时间', dataIndex: 'last_heartbeat', key: 'last_heartbeat', render: (value: string) => value ? new Date(value).toLocaleString() : '-' },
    tableActionColumn<DeviceCatalogItem>(
      (_, record) => (
        <Space size={4} className="ota-table-actions">
          <Button type="link" size="small" onClick={() => navigate(`/devices/${record.device_id}`)}>
            查看详情
          </Button>
        </Space>
      ),
      { actionLabels: ['查看详情'] },
    ),
  ];

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
        description="只保存 OTA 选型与分组需要的最小字段。开启「异常设备」可筛选已阻断或目录冲突的设备；点击详情可查看升级历史。"
      />

      <Card className="ota-card">
        <div className="ota-toolbar">
          <div className="ota-toolbar-left">
            <Input.Search
              className="ota-toolbar-control-search"
              placeholder="搜索设备 ID / 型号 / 产品代码"
              allowClear
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onSearch={applyFilters}
            />
            <Select
              className="ota-toolbar-control-select"
              allowClear
              placeholder="设备分组"
              value={groupFilter}
              options={filterOptions.groups.map((group) => ({ label: group, value: group }))}
              onChange={(value) => { setGroupFilter(value); setPage(1); }}
            />
            <Select
              className="ota-toolbar-control-select"
              allowClear
              placeholder="产品型号"
              value={modelFilter}
              options={filterOptions.models.map((model) => ({ label: model, value: model }))}
              onChange={(value) => { setModelFilter(value); setPage(1); }}
            />
            <Input
              className="ota-toolbar-control-select"
              placeholder="标签关键词"
              allowClear
              value={tagFilter}
              onChange={(e) => setTagFilter(e.target.value)}
              onPressEnter={applyFilters}
            />
            <Space size={4}>
              <Text type="secondary">异常设备</Text>
              <Switch checked={abnormalOnly} onChange={(checked) => { setAbnormalOnly(checked); setPage(1); }} />
            </Space>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void load()}>刷新</Button>
            <Button icon={<DownloadOutlined />} onClick={() => void handleDownloadTemplate()}>模板</Button>
            <Upload accept=".csv,text/csv" showUploadList={false} beforeUpload={(file) => { void handleImport(file); return false; }}>
              <Button type="primary" icon={<UploadOutlined />} loading={importing}>导入 CSV</Button>
            </Upload>
          </Space>
        </div>

        <Table
          {...listTableProps}
          rowKey="device_id"
          columns={columns}
          dataSource={devices}
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: false,
            onChange: (next) => setPage(next),
          }}
          scroll={listTableScroll(1280, devices.length)}
        />
      </Card>
    </div>
  );
}

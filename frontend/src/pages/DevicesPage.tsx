import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Select, Space, Switch, Table, Tag, Upload, message } from 'antd';
import { ApiOutlined, DownloadOutlined, ReloadOutlined, UploadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { DeviceCatalogItem, deviceAPI } from '../api';
import { tableActionColumn, listTableProps, listTableScroll, serverTablePagination } from '../utils/tableActionColumn';
import { tableIdLinkColumn } from '../utils/tableIdLinkColumn';
import { tableEllipsisColumn, tableCompactColumn } from '../utils/tableEllipsisColumn';

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
    tableIdLinkColumn<DeviceCatalogItem>('设备 ID', 'device_id', (id) => navigate(`/devices/${id}`)),
    tableEllipsisColumn<DeviceCatalogItem>('产品代码', 'product_code'),
    tableEllipsisColumn<DeviceCatalogItem>('型号', 'product_model'),
    tableEllipsisColumn<DeviceCatalogItem>('硬件版本', 'hardware_version'),
    tableEllipsisColumn<DeviceCatalogItem>('设备分组', 'device_group'),
    tableEllipsisColumn<DeviceCatalogItem>('当前版本', 'current_version'),
    tableEllipsisColumn<DeviceCatalogItem>('OTA 上报', 'reported_version', {
      render: (v) => (v ? String(v) : '-'),
    }),
    tableCompactColumn<DeviceCatalogItem>('状态', 'status', (_: unknown, record: DeviceCatalogItem) => {
      const flags = record.inconsistency_flags ?? [];
      if (record.eligibility_state === 'blocked') {
        return <Tag color="red">已阻断</Tag>;
      }
      if (flags.length > 0) {
        return <Tag color="orange">目录异常</Tag>;
      }
      return <Tag color="green">正常</Tag>;
    }),
    tableEllipsisColumn<DeviceCatalogItem>('导入/更新时间', 'last_heartbeat', {
      size: 'date',
      render: (v) => (v ? new Date(String(v)).toLocaleString() : '-'),
    }),
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
            <label className="ota-toolbar-filter-switch">
              <span className="ota-toolbar-filter-switch-label">仅异常设备</span>
              <Switch
                size="small"
                checked={abnormalOnly}
                onChange={(checked) => { setAbnormalOnly(checked); setPage(1); }}
              />
            </label>
          </div>
          <Space>
            <Button icon={<ApiOutlined />} onClick={() => navigate('/simulator')}>设备端模拟器</Button>
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
          pagination={serverTablePagination(page, pageSize, total, setPage)}
          scroll={listTableScroll(columns, devices.length)}
          locale={{ emptyText: '暂无设备，请先导入 CSV 或使用模板' }}
        />
      </Card>
    </div>
  );
}

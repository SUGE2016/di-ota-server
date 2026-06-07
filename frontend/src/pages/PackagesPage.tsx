import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, Select, Space, message, Modal, Table, Tag } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Package, packageAPI } from '../api';
import { tableActionColumn, listTableProps, listTableScroll, clientTablePagination } from '../utils/tableActionColumn';
import { tableIdLinkColumn } from '../utils/tableIdLinkColumn';
import { tableEllipsisColumn, tableCompactColumn } from '../utils/tableEllipsisColumn';

const statusColor: Record<string, string> = {
  Published: 'green',
  Draft: 'orange',
  Deprecated: 'default',
  Disabled: 'red',
  Archived: 'default',
};

export function PackagesPage() {
  const [packages, setPackages] = useState<Package[]>([]);
  const [keyword, setKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [form] = Form.useForm();
  const [uploading, setUploading] = useState(false);
  const navigate = useNavigate();

  const load = async () => {
    try {
      const data = await packageAPI.list(100, 0);
      setPackages(data.packages);
    } catch (e: any) {
      message.error(e.message);
    }
    setLoading(false);
  };

  useEffect(() => { load(); }, []);

  const handleUpload = async (values: any) => {
    setUploading(true);
    try {
      const file = values.file?.file;
      if (!file) { message.error('请选择文件'); return; }

      const { package_id, upload_url } = await packageAPI.uploadUrl({
        file_name: file.name,
        content_type: file.type || 'application/octet-stream',
      });

      await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': 'application/octet-stream' },
      });

      await packageAPI.complete({
        package_id,
        product_code: values.product_code,
        version: values.version,
        file_hash: values.file_hash || '',
        signature: values.signature || '',
        file_size: file.size,
      });

      message.success('上传成功');
      setUploadOpen(false);
      form.resetFields();
      load();
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setUploading(false);
    }
  };

  const handleDeprecate = (pkg: Package) => {
    Modal.confirm({
      title: `下架固件包 ${pkg.package_id}?`,
      onOk: async () => {
        try {
          await packageAPI.updateStatus(pkg.package_id, 'Deprecated');
          message.success('已下架');
          load();
        } catch (e: any) {
          message.error(e.message);
        }
      },
    });
  };

  const columns = [
    tableIdLinkColumn<Package>('包 ID', 'package_id', (id) => navigate(`/packages/${id}`)),
    tableEllipsisColumn<Package>('产品代码', 'product_code'),
    tableEllipsisColumn<Package>('版本', 'version'),
    tableCompactColumn<Package>('状态', 'status', (v: string) => <Tag color={statusColor[v]}>{v}</Tag>, 'status'),
    tableEllipsisColumn<Package>('创建时间', 'created_at', {
      size: 'date',
      render: (v) => new Date(String(v)).toLocaleString(),
    }),
    tableActionColumn<Package>(
      (_, r) => (
        r.status === 'Published' ? (
          <Space size={4} className="ota-table-actions">
            <Button type="link" size="small" onClick={() => handleDeprecate(r)}>下架</Button>
          </Space>
        ) : null
      ),
      { actionLabels: ['下架'] },
    ),
  ];

  const filteredPackages = packages.filter((p) => {
    if (statusFilter && p.status !== statusFilter) return false;
    if (!keyword.trim()) return true;
    const key = keyword.toLowerCase();
    return [p.package_id, p.product_code, p.version].join(' ').toLowerCase().includes(key);
  });

  return (
    <div className="ota-page">
      <Card
        className="ota-card"
        extra={<Button type="primary" icon={<UploadOutlined />} onClick={() => setUploadOpen(true)}>上传固件包</Button>}
      >
        <div className="ota-toolbar">
          <div className="ota-toolbar-left">
            <Input.Search
              allowClear
              placeholder="搜索包ID/产品代码/版本"
              className="ota-toolbar-control-search"
              onSearch={setKeyword}
              onChange={(e) => setKeyword(e.target.value)}
            />
            <Select
              allowClear
              placeholder="状态筛选"
              className="ota-toolbar-control-select"
              value={statusFilter}
              onChange={(v) => setStatusFilter(v)}
              options={[
                { value: 'Published', label: 'Published' },
                { value: 'Draft', label: 'Draft' },
                { value: 'Deprecated', label: 'Deprecated' },
                { value: 'Disabled', label: 'Disabled' },
                { value: 'Archived', label: 'Archived' },
              ]}
            />
          </div>
          <span className="ota-muted">共 {filteredPackages.length} 条</span>
        </div>

        <Table
          {...listTableProps}
          columns={columns}
          dataSource={filteredPackages}
          loading={loading}
          rowKey="package_id"
          pagination={clientTablePagination(12)}
          scroll={listTableScroll(columns, filteredPackages.length)}
        />

        <Modal width="min(560px, calc(100vw - 24px))" title="上传固件包" open={uploadOpen} onCancel={() => { setUploadOpen(false); form.resetFields(); }} footer={null}>
          <Form form={form} layout="vertical" onFinish={handleUpload}>
            <Form.Item name="product_code" label="产品代码" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="version" label="版本号" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="file_hash" label="文件哈希">
              <Input />
            </Form.Item>
            <Form.Item name="signature" label="签名">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="file" label="固件文件" rules={[{ required: true }]}>
              <Input type="file" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={uploading} block>上传</Button>
            </Form.Item>
          </Form>
        </Modal>
      </Card>
    </div>
  );
}

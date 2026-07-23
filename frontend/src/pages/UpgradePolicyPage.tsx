import { useEffect, useState } from 'react';
import { Alert, Card, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { productModelPolicyAPI, type ProductModelPolicy } from '../api';
import { listTableProps, listTableScroll } from '../utils/tableActionColumn';
import { tableEllipsisColumn } from '../utils/tableEllipsisColumn';

const { Text } = Typography;

export function UpgradePolicyPage() {
  const [loading, setLoading] = useState(true);
  const [savingModel, setSavingModel] = useState<string | null>(null);
  const [defaultMode, setDefaultMode] = useState('relaxed');
  const [policies, setPolicies] = useState<ProductModelPolicy[]>([]);

  const load = async () => {
    setLoading(true);
    try {
      const data = await productModelPolicyAPI.list();
      setDefaultMode(data.default_mode || 'relaxed');
      setPolicies(data.policies || []);
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleModeChange = async (productModel: string, mode: string) => {
    setSavingModel(productModel);
    try {
      await productModelPolicyAPI.update(productModel, mode);
      message.success(`${productModel} 已设为 ${mode === 'strict' ? '严格' : '宽松'}`);
      await load();
    } catch (e) {
      message.error(e instanceof Error ? e.message : '保存失败');
    } finally {
      setSavingModel(null);
    }
  };

  const columns: ColumnsType<ProductModelPolicy> = [
    tableEllipsisColumn<ProductModelPolicy>('产品型号', 'product_model'),
    {
      title: '模式',
      dataIndex: 'report_status_mode',
      width: 180,
      render: (mode: string, record) => (
        <Select
          value={mode || 'relaxed'}
          style={{ width: 140 }}
          loading={savingModel === record.product_model}
          options={[
            { value: 'relaxed', label: '宽松' },
            { value: 'strict', label: '严格' },
          ]}
          onChange={(v) => handleModeChange(record.product_model, v)}
        />
      ),
    },
    {
      title: '配置来源',
      dataIndex: 'explicit',
      width: 120,
      render: (explicit: boolean) =>
        explicit ? <Tag>已保存</Tag> : <Text type="secondary">默认</Text>,
    },
    {
      title: '更新人',
      dataIndex: 'updated_by',
      width: 120,
      render: (v: string) => v || '—',
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 200,
      render: (v: string | null) => v || '—',
    },
  ];

  return (
    <div className="ota-page">
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="升级策略（按 product_model）"
        description={
          <span>
            默认模式：<Tag color="blue">{defaultMode === 'strict' ? '严格' : '宽松'}</Tag>
            宽松允许乱序/回退重置与 Failed→Success，减少网络抖动导致的反复升级；严格保持原状态机（非法迁移 409）。
            Success 后中间态一律忽略。仅 admin 可修改。
          </span>
        }
      />

      <Card className="ota-card" title="型号策略">
        <Table<ProductModelPolicy>
          {...listTableProps}
          loading={loading}
          rowKey="product_model"
          dataSource={policies}
          scroll={listTableScroll(columns, policies.length)}
          pagination={false}
          columns={columns}
        />
        {!loading && policies.length === 0 && (
          <Space style={{ marginTop: 12 }}>
            <Text type="secondary">暂无设备型号。导入设备后将自动出现在此列表。</Text>
          </Space>
        )}
      </Card>
    </div>
  );
}

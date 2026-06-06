import { Card, Typography } from 'antd';
import { BellOutlined } from '@ant-design/icons';

const { Paragraph, Title } = Typography;

export function AlertsPage() {
  return (
    <div className="ota-page">
      <div>
        <Title level={3} className="ota-page-title">告警中心</Title>
        <Paragraph className="ota-page-subtitle">
          集中查看升级失败、任务异常等告警，并跳转至相关任务或设备。
        </Paragraph>
      </div>

      <Card className="ota-card">
        <div className="ota-empty-panel">
          <BellOutlined className="ota-empty-panel-icon" />
          <Title level={4} style={{ marginBottom: 8 }}>告警功能尚未开放</Title>
          <Paragraph type="secondary" style={{ marginBottom: 0, maxWidth: 480 }}>
            当前请通过「发布任务」详情查看失败率，或在「设备管理」中筛选异常设备并查看升级历史。
          </Paragraph>
        </div>
      </Card>
    </div>
  );
}

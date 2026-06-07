import { useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Form,
  Input,
  Row,
  Space,
  Switch,
  Typography,
  message,
} from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import {
  UPGRADE_STATUS_STEPS,
  checkUpdate,
  describeNoUpdate,
  probeDownload,
  reportStatus,
  type DeviceConfig,
} from '../lib/deviceSimulator';
import { APP_VERSION_LABEL, COPYRIGHT_HOLDER } from '../constants/version';

const { Title, Text, Paragraph } = Typography;

type LogEntry = {
  id: number;
  time: string;
  title: string;
  body: string;
  ok: boolean;
};

const defaultConfig: DeviceConfig = {
  deviceId: 'AMS000001',
  group: 'org-1001',
  productModel: 'V9',
  hardwareVersion: '1.0',
  currentVersion: 'v2.3.0',
  apiToken: '',
};

function appendLog(prev: LogEntry[], title: string, body: unknown, ok: boolean, idRef: { n: number }) {
  idRef.n += 1;
  return [
    {
      id: idRef.n,
      time: new Date().toLocaleTimeString(),
      title,
      body: typeof body === 'string' ? body : JSON.stringify(body, null, 2),
      ok,
    },
    ...prev,
  ];
}

export function DeviceSimulatorPage() {
  const [form] = Form.useForm<DeviceConfig & { tryDownload: boolean }>();
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [running, setRunning] = useState(false);
  const idRef = useMemo(() => ({ n: 0 }), []);

  const pushLog = (title: string, body: unknown, ok: boolean) => {
    setLogs((prev) => appendLog(prev, title, body, ok, idRef));
  };

  const getConfig = (): DeviceConfig => {
    const v = form.getFieldsValue();
    return {
      deviceId: v.deviceId,
      group: v.group,
      productModel: v.productModel,
      hardwareVersion: v.hardwareVersion,
      currentVersion: v.currentVersion,
      apiToken: v.apiToken || '',
    };
  };

  const handleCheckUpdate = async () => {
    setRunning(true);
    try {
      const cfg = getConfig();
      const res = await checkUpdate(cfg);
      pushLog('check-update', res, true);
      if (res.data?.current_version) {
        form.setFieldValue('currentVersion', res.data.current_version);
      }
      if (res.data?.has_update) {
        message.success(`发现升级 task_id=${res.data.task_id}`);
      } else {
        message.warning(describeNoUpdate(res.data));
      }
      return res.data;
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      pushLog('check-update 失败', msg, false);
      message.error(msg);
      return null;
    } finally {
      setRunning(false);
    }
  };

  const handleFullFlow = async () => {
    setRunning(true);
    try {
      const cfg = getConfig();
      const tryDownload = form.getFieldValue('tryDownload') === true;
      const check = await checkUpdate(cfg);
      pushLog('check-update', check, true);

      if (check.data?.current_version) {
        form.setFieldValue('currentVersion', check.data.current_version);
      }

      if (!check.data?.has_update) {
        const hint = describeNoUpdate(check.data);
        pushLog('流程结束', { reason: check.data?.reason, current_version: check.data?.current_version, hint }, true);
        message.warning(hint);
        return;
      }

      const taskId = check.data.task_id || '';
      const targetVersion = check.data.target_version || '';
      const sourceVersion = check.data.current_version || cfg.currentVersion;
      if (!taskId) {
        throw new Error('响应缺少 task_id');
      }

      if (tryDownload && check.data.download_url) {
        try {
          const dl = await probeDownload(check.data.download_url);
          pushLog('download', dl, dl.status >= 200 && dl.status < 400);
        } catch (e) {
          pushLog('download 失败', e instanceof Error ? e.message : String(e), false);
        }
      }

      for (const status of UPGRADE_STATUS_STEPS) {
        await new Promise((r) => setTimeout(r, 300));
        const rep = await reportStatus(cfg, { taskId, status, targetVersion, sourceVersion });
        pushLog(`report-status (${status})`, rep, true);
      }

      form.setFieldValue('currentVersion', targetVersion);
      message.success('升级流程模拟完成');
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      pushLog('流程失败', msg, false);
      message.error(msg);
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="ota-simulator-wrap">
      <div className="ota-simulator-header">
        <div>
          <Title level={3} style={{ margin: 0 }}>OTA 设备端模拟器</Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            模拟设备调用 check-update / report-status，用于联调升级链路。
          </Paragraph>
        </div>
        <Link to="/devices">返回设备管理</Link>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card title="设备参数" className="ota-card">
            <Alert
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
              message="重复联调说明"
              description="升级成功后服务端会记住 reported_version，同一版本任务无法再次完整走通。需要再次演练时，请发布更高版本固件包并创建新的 Running 任务。"
            />
            <Form
              form={form}
              layout="vertical"
              initialValues={{ ...defaultConfig, tryDownload: false }}
            >
              <Form.Item name="deviceId" label="device_id" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="group" label="group" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="productModel" label="product_model" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="hardwareVersion" label="hardware_version" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="currentVersion" label="current_version">
                <Input />
              </Form.Item>
              <Form.Item name="apiToken" label="Bearer Token（可选）">
                <Input.Password placeholder="DEVICE_API_TOKEN" />
              </Form.Item>
              <Form.Item name="tryDownload" label="完整流程时尝试下载" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Space wrap>
                <Button onClick={handleCheckUpdate} loading={running}>
                  检查更新
                </Button>
                <Button type="primary" onClick={handleFullFlow} loading={running}>
                  完整升级流程
                </Button>
                <Button onClick={() => setLogs([])} disabled={running}>
                  清空日志
                </Button>
              </Space>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={14}>
          <Card title="请求日志" className="ota-card ota-simulator-log-card">
            <div className="ota-simulator-log">
              {logs.length === 0 ? (
                <Text type="secondary">执行操作后在此显示 JSON 响应</Text>
              ) : (
                logs.map((item) => (
                  <div key={item.id} className={`ota-simulator-log-item ${item.ok ? 'ok' : 'err'}`}>
                    <div className="ota-simulator-log-head">
                      <Text strong>{item.title}</Text>
                      <Text type="secondary">{item.time}</Text>
                    </div>
                    <pre>{item.body}</pre>
                  </div>
                ))
              )}
            </div>
          </Card>
        </Col>
      </Row>

      <div className="ota-simulator-footer">
        <span>© {COPYRIGHT_HOLDER}</span>
        <span>{APP_VERSION_LABEL}</span>
      </div>
    </div>
  );
}

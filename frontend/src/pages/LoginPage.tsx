import { useState } from 'react';
import { Button, Divider, Form, Input, Card, message, Typography } from 'antd';
import { UserOutlined, LockOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../api';
import useAuthStore from '../stores/authStore';
import { APP_VERSION_LABEL, COPYRIGHT_HOLDER } from '../constants/version';
import { OtaLogo } from '../components/OtaLogo';

const { Paragraph } = Typography;

export function LoginPage() {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const res = await authAPI.login(values.username, values.password);
      useAuthStore.getState().setAuth(res.access_token);
      message.success('登录成功');
      navigate('/dashboard');
    } catch {
      message.error('用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="ota-login-wrap">
      <div className="ota-login-aside">
        <div className="ota-login-brand">
          <OtaLogo className="ota-login-brand-logo" size={40} />
          <div className="ota-login-brand-lines">
            <span className="ota-login-brand-text">OTA 管理台</span>
            <span className="ota-login-brand-tagline">固件发布与灰度升级管理</span>
          </div>
        </div>
        <Paragraph className="ota-login-aside-copy">
          管理固件包、发布灰度任务、跟踪升级状态，构建稳定可控的设备升级流程。
        </Paragraph>
      </div>

      <div className="ota-login-card-zone">
        <Card className="ota-card ota-login-card" style={{ width: 420 }}>
          <Form onFinish={onFinish} size="large" layout="vertical" className="ota-login-form">
            <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input prefix={<UserOutlined />} placeholder="请输入账号" />
            </Form.Item>
            <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
            </Form.Item>
            <Form.Item style={{ marginBottom: 6 }}>
              <Button type="primary" htmlType="submit" loading={loading} block>
                登录
              </Button>
            </Form.Item>
          </Form>
          <Divider plain>或</Divider>
          <Button icon={<SafetyCertificateOutlined />} size="large" block disabled>
            SSO 登录（待接入）
          </Button>
        </Card>
        <div className="ota-login-foot">
          <Link to="/simulator" className="ota-login-foot-link">设备端模拟器</Link>
          <span className="ota-login-foot-meta">
            © {COPYRIGHT_HOLDER} · {APP_VERSION_LABEL}
          </span>
        </div>
      </div>
    </div>
  );
}

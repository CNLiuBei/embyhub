import React, { useState, useEffect } from 'react';
import { Card, Steps, Button, Input, Form, message, Result, Spin } from 'antd';
import { 
  CheckCircleOutlined, 
  DatabaseOutlined, 
  MailOutlined, 
  UserOutlined,
  SafetyOutlined,
  PlayCircleOutlined
} from '@ant-design/icons';
import { post, get } from '@/utils/request';
import './Setup.css';

const { Step } = Steps;

// 安装完成页面组件 - 等待后端重启后自动跳转
const FinishedPage: React.FC = () => {
  const [status, setStatus] = useState<'restarting' | 'ready' | 'error'>('restarting');
  const [countdown, setCountdown] = useState(5);

  useEffect(() => {
    let attempts = 0;
    const maxAttempts = 30; // 最多等待30秒

    const checkBackend = async () => {
      try {
        // 尝试访问登录API检查后端是否就绪
        const response = await fetch('/api/auth/login', { 
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: '', password: '' })
        });
        // 只要能响应（即使是400错误）说明后端已就绪
        if (response.status !== 404) {
          setStatus('ready');
          return;
        }
      } catch {
        // 连接失败，继续等待
      }
      
      attempts++;
      if (attempts >= maxAttempts) {
        setStatus('error');
      } else {
        setTimeout(checkBackend, 1000);
      }
    };

    // 延迟3秒后开始检查（给后端重启时间）
    setTimeout(checkBackend, 3000);
  }, []);

  useEffect(() => {
    if (status === 'ready' && countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }
    if (status === 'ready' && countdown === 0) {
      window.location.href = '/login';
    }
  }, [status, countdown]);

  return (
    <div className="setup-container">
      <Card className="setup-card">
        <Result
          status={status === 'error' ? 'warning' : 'success'}
          title={status === 'restarting' ? '🔄 正在重启服务...' : status === 'ready' ? '🎉 初始化完成！' : '⚠️ 重启超时'}
          subTitle={
            <div style={{ textAlign: 'center' }}>
              {status === 'restarting' && (
                <>
                  <Spin style={{ marginBottom: 16 }} />
                  <p>系统正在重启，请稍候...</p>
                </>
              )}
              {status === 'ready' && (
                <p>{countdown} 秒后自动跳转到登录页...</p>
              )}
              {status === 'error' && (
                <p>后端服务重启超时，请手动重启后端服务</p>
              )}
            </div>
          }
          extra={[
            <Button 
              type="primary" 
              key="login" 
              onClick={() => window.location.href = '/login'}
              disabled={status === 'restarting'}
            >
              {status === 'restarting' ? '等待中...' : '前往登录'}
            </Button>
          ]}
        />
      </Card>
    </div>
  );
};

// 配置类型
interface SetupConfig {
  server: { port: number; mode: string };
  database: { host: string; port: number; user: string; password: string; dbname: string; sslmode: string; maxIdleConns: number; maxOpenConns: number; connMaxLifetime: number };
  redis: { host: string; port: number; password: string; db: number };
  jwt: { secret: string; expireHours: number };
  emby: { serverUrl: string; apiKey: string };
  email: { host: string; port: number; user: string; password: string; from: string };
  log: { level: string; filename: string };
  cors: { allowOrigins: string[]; allowMethods: string[]; allowHeaders: string[]; exposeHeaders: string[]; allowCredentials: boolean; maxAge: number };
}

const Setup: React.FC = () => {
  const [current, setCurrent] = useState(0);
  const [loading, setLoading] = useState(true);
  const [stepLoading, setStepLoading] = useState(false);
  const [stepPassed, setStepPassed] = useState<boolean[]>([false, false, false, false, false]);
  const [config, setConfig] = useState<SetupConfig | null>(null);
  const [adminInfo, setAdminInfo] = useState({ user: 'admin', pass: '', email: '' });
  const [finished, setFinished] = useState(false);

  // 检查初始化状态
  useEffect(() => {
    checkStatus();
  }, []);

  const checkStatus = async () => {
    try {
      const res: any = await get('/setup/status');
      if (res.code === 200 && res.data?.initialized) {
        message.info('系统已初始化，正在跳转到登录页...');
        setTimeout(() => {
          window.location.href = '/login';
        }, 1000);
        return;
      }
      // 获取默认配置
      const configRes: any = await get('/setup/config');
      if (configRes.code === 200) {
        setConfig(configRes.data);
      } else {
        message.error('获取配置失败');
      }
    } catch (e) {
      message.error('获取配置失败，请确保后端服务已启动');
    } finally {
      setLoading(false);
    }
  };

  // 验证授权码
  const verifyLicense = async (values: { license: string }) => {
    setStepLoading(true);
    try {
      const res: any = await post('/setup/verify-license', values);
      if (res.code === 200) {
        message.success(res.message);
        markStepPassed(0);
      } else {
        message.error(res.message);
      }
    } catch (e) {
      message.error('验证失败');
    } finally {
      setStepLoading(false);
    }
  };

  // 测试数据库
  const testDatabase = async () => {
    if (!config) return;
    setStepLoading(true);
    try {
      const res: any = await post('/setup/test-database', config.database);
      if (res.code === 200) {
        message.success(res.message);
        markStepPassed(1);
      } else {
        message.error(res.message);
      }
    } catch (e) {
      message.error('测试失败');
    } finally {
      setStepLoading(false);
    }
  };

  // 测试 Emby
  const testEmby = async () => {
    if (!config) return;
    setStepLoading(true);
    try {
      const res: any = await post('/setup/test-emby', config.emby);
      if (res.code === 200) {
        message.success(res.message);
        markStepPassed(2);
      } else {
        message.error(res.message);
      }
    } catch (e) {
      message.error('测试失败');
    } finally {
      setStepLoading(false);
    }
  };

  // 测试邮件
  const testEmail = async () => {
    if (!config) return;
    setStepLoading(true);
    try {
      const res: any = await post('/setup/test-email', config.email);
      if (res.code === 200) {
        message.success(res.message);
        markStepPassed(3);
      } else {
        message.error(res.message);
      }
    } catch (e) {
      message.error('测试失败');
    } finally {
      setStepLoading(false);
    }
  };

  // 完成安装
  const finishSetup = async () => {
    if (!config || !adminInfo.pass || !adminInfo.email) {
      message.error('请填写完整管理员信息');
      return;
    }
    setStepLoading(true);
    try {
      const res: any = await post('/setup/finish', {
        config,
        admin_user: adminInfo.user,
        admin_pass: adminInfo.pass,
        admin_email: adminInfo.email,
      });
      if (res.code === 200) {
        message.success(res.message);
        setFinished(true);
      } else {
        message.error(res.message);
      }
    } catch (e) {
      message.error('安装失败');
    } finally {
      setStepLoading(false);
    }
  };

  const markStepPassed = (step: number) => {
    const newPassed = [...stepPassed];
    newPassed[step] = true;
    setStepPassed(newPassed);
  };

  const updateConfig = (section: keyof SetupConfig, field: string, value: any) => {
    if (!config) return;
    setConfig({
      ...config,
      [section]: { ...config[section], [field]: value }
    });
  };

  const steps = [
    { title: '授权验证', icon: <SafetyOutlined /> },
    { title: '数据库', icon: <DatabaseOutlined /> },
    { title: 'Emby', icon: <PlayCircleOutlined /> },
    { title: '邮件', icon: <MailOutlined /> },
    { title: '完成', icon: <UserOutlined /> },
  ];

  if (loading) {
    return (
      <div className="setup-container">
        <Spin size="large">
          <div style={{ padding: 50, textAlign: 'center' }}>加载中...</div>
        </Spin>
      </div>
    );
  }

  if (finished) {
    return <FinishedPage />;
  }

  // 步骤内容渲染
  const renderStepContent = () => {
    if (!config) return null;

    switch (current) {
      case 0: // 授权验证
        return (
          <div className="step-content">
            <div className="step-icon">🔐</div>
            <h2>授权验证</h2>
            <p className="step-desc">请输入授权码以继续安装</p>
            <Form onFinish={verifyLicense} layout="vertical">
              <Form.Item name="license" rules={[{ required: true, message: '请输入授权码' }]}>
                <Input size="large" placeholder="请输入授权码" style={{ textAlign: 'center' }} />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={stepLoading} block size="large">
                  验证授权
                </Button>
              </Form.Item>
            </Form>
            <div className="step-tip">
              💡 试用授权码：<code>EMBY-FREE-TRIAL</code>
            </div>
          </div>
        );

      case 1: // 数据库配置
        return (
          <div className="step-content">
            <div className="step-icon">🗄️</div>
            <h2>数据库配置</h2>
            <p className="step-desc">配置 PostgreSQL 数据库连接</p>
            <Form layout="vertical">
              <div className="form-row">
                <Form.Item label="主机地址" className="form-item-half">
                  <Input value={config.database.host} onChange={e => updateConfig('database', 'host', e.target.value)} />
                </Form.Item>
                <Form.Item label="端口" className="form-item-half">
                  <Input type="number" value={config.database.port} onChange={e => updateConfig('database', 'port', parseInt(e.target.value))} />
                </Form.Item>
              </div>
              <div className="form-row">
                <Form.Item label="用户名" className="form-item-half">
                  <Input value={config.database.user} onChange={e => updateConfig('database', 'user', e.target.value)} />
                </Form.Item>
                <Form.Item label="密码" className="form-item-half">
                  <Input.Password value={config.database.password} onChange={e => updateConfig('database', 'password', e.target.value)} />
                </Form.Item>
              </div>
              <Form.Item label="数据库名">
                <Input value={config.database.dbname} onChange={e => updateConfig('database', 'dbname', e.target.value)} />
              </Form.Item>
              <Button type="primary" onClick={testDatabase} loading={stepLoading} icon={stepPassed[1] ? <CheckCircleOutlined /> : undefined}>
                {stepPassed[1] ? '连接成功' : '测试连接'}
              </Button>
            </Form>
          </div>
        );

      case 2: // Emby 配置
        return (
          <div className="step-content">
            <div className="step-icon">🎬</div>
            <h2>Emby 服务器配置</h2>
            <p className="step-desc">配置 Emby 媒体服务器连接</p>
            <Form layout="vertical">
              <Form.Item label="Emby 服务器地址">
                <Input value={config.emby.serverUrl} onChange={e => updateConfig('emby', 'serverUrl', e.target.value)} placeholder="http://localhost:8096" />
              </Form.Item>
              <Form.Item label="API Key">
                <Input value={config.emby.apiKey} onChange={e => updateConfig('emby', 'apiKey', e.target.value)} placeholder="在 Emby 后台获取" />
              </Form.Item>
              <div className="step-tip">
                💡 API Key 可在 Emby 管理后台 → 设置 → 高级 → API 密钥 中创建
              </div>
              <Button type="primary" onClick={testEmby} loading={stepLoading} icon={stepPassed[2] ? <CheckCircleOutlined /> : undefined}>
                {stepPassed[2] ? '连接成功' : '测试连接'}
              </Button>
            </Form>
          </div>
        );

      case 3: // 邮件配置
        return (
          <div className="step-content">
            <div className="step-icon">📧</div>
            <h2>邮件服务配置</h2>
            <p className="step-desc">配置 SMTP 邮件服务</p>
            <Form layout="vertical">
              <div className="form-row">
                <Form.Item label="SMTP 服务器" className="form-item-half">
                  <Input value={config.email.host} onChange={e => updateConfig('email', 'host', e.target.value)} />
                </Form.Item>
                <Form.Item label="端口" className="form-item-half">
                  <Input type="number" value={config.email.port} onChange={e => updateConfig('email', 'port', parseInt(e.target.value))} />
                </Form.Item>
              </div>
              <div className="form-row">
                <Form.Item label="邮箱账号" className="form-item-half">
                  <Input value={config.email.user} onChange={e => updateConfig('email', 'user', e.target.value)} />
                </Form.Item>
                <Form.Item label="密码/授权码" className="form-item-half">
                  <Input.Password value={config.email.password} onChange={e => updateConfig('email', 'password', e.target.value)} />
                </Form.Item>
              </div>
              <div className="step-tip">
                💡 常用配置：QQ邮箱 smtp.qq.com:587 | 阿里企业邮箱 smtp.qiye.aliyun.com:465
              </div>
              <Button type="primary" onClick={testEmail} loading={stepLoading} icon={stepPassed[3] ? <CheckCircleOutlined /> : undefined}>
                {stepPassed[3] ? '连接成功' : '测试连接'}
              </Button>
            </Form>
          </div>
        );

      case 4: // 完成设置
        return (
          <div className="step-content">
            <div className="step-icon">👤</div>
            <h2>管理员账户</h2>
            <p className="step-desc">设置系统管理员信息</p>
            <Form layout="vertical">
              <Form.Item label="管理员用户名">
                <Input value={adminInfo.user} onChange={e => setAdminInfo({...adminInfo, user: e.target.value})} />
              </Form.Item>
              <Form.Item label="管理员密码">
                <Input.Password value={adminInfo.pass} onChange={e => setAdminInfo({...adminInfo, pass: e.target.value})} placeholder="请设置密码" />
              </Form.Item>
              <Form.Item label="管理员邮箱">
                <Input value={adminInfo.email} onChange={e => setAdminInfo({...adminInfo, email: e.target.value})} placeholder="admin@example.com" />
              </Form.Item>
              <Button type="primary" onClick={finishSetup} loading={stepLoading} size="large" block>
                完成初始化
              </Button>
            </Form>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="setup-container">
      <Card className="setup-card">
        <div className="setup-header">
          <h1>Emby Hub 初始化向导</h1>
        </div>
        
        <Steps current={current} className="setup-steps">
          {steps.map((item, index) => (
            <Step key={index} title={item.title} icon={item.icon} />
          ))}
        </Steps>

        <div className="setup-content">
          {renderStepContent()}
        </div>

        <div className="setup-footer">
          {current > 0 && (
            <Button onClick={() => setCurrent(current - 1)}>上一步</Button>
          )}
          {current < steps.length - 1 && (
            <Button 
              type="primary" 
              onClick={() => setCurrent(current + 1)} 
              disabled={!stepPassed[current]}
            >
              下一步
            </Button>
          )}
        </div>
      </Card>
    </div>
  );
};

export default Setup;

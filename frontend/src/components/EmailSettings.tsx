import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Switch, InputNumber, message, Space, Divider, Tag, Row, Col } from 'antd';
import { MailOutlined, SendOutlined, LockOutlined, GlobalOutlined, ThunderboltOutlined, CloudOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { getConfigs, updateConfig } from '@/api/config';

interface EmailConfig {
  email_provider: string;
  email_from: string;
  email_from_name: string;
  smtp_host: string;
  smtp_port: number;
  smtp_user: string;
  smtp_password: string;
  smtp_ssl: boolean;
  resend_api_key: string;
  aliyun_access_key_id: string;
  aliyun_access_key_secret: string;
  aliyun_region: string;
}

// 服务商名称映射
const providerNames: Record<string, string> = {
  aliyun_smtp: '阿里企业邮箱',
  aliyun: '阿里云推送',
  resend: 'Resend',
  smtp: 'SMTP',
};

const EmailSettings: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testEmail, setTestEmail] = useState('');
  const [provider, setProvider] = useState('smtp');
  const [serviceStatus, setServiceStatus] = useState<{ configured: boolean; provider: string; from: string }>({
    configured: false,
    provider: '',
    from: '',
  });

  // 加载配置
  const loadConfigs = async () => {
    setLoading(true);
    try {
      const response = await getConfigs();
      if (response.code === 200 && response.data) {
        const data = response.data as any;
        const configList = Array.isArray(data) ? data : (data.list || []);
        
        const configMap: Record<string, string> = {};
        configList.forEach((item: { config_key: string; config_value: string }) => {
          configMap[item.config_key] = item.config_value;
        });
        
        // 根据配置读取服务商
        const currentProvider = configMap.email_provider || 'smtp';
        setProvider(currentProvider);
        
        const emailFrom = configMap.email_from || configMap.smtp_from || '';
        
        form.setFieldsValue({
          email_provider: currentProvider,
          email_from: emailFrom,
          email_from_name: configMap.email_from_name || configMap.smtp_from_name || 'Emby Hub',
          smtp_host: configMap.smtp_host || '',
          smtp_port: parseInt(configMap.smtp_port) || 465,
          smtp_user: configMap.smtp_user || '',
          smtp_password: configMap.smtp_password || '',
          smtp_ssl: configMap.smtp_ssl === 'true',
          resend_api_key: configMap.resend_api_key || '',
          aliyun_access_key_id: configMap.aliyun_access_key_id || '',
          aliyun_access_key_secret: configMap.aliyun_access_key_secret || '',
          aliyun_region: configMap.aliyun_region || 'cn-hangzhou',
        });

        // 检查服务是否已配置
        let configured = false;
        if (currentProvider === 'aliyun_smtp' || currentProvider === 'smtp') {
          configured = !!configMap.smtp_host && !!configMap.smtp_user;
        } else if (currentProvider === 'resend') {
          configured = !!configMap.resend_api_key;
        } else if (currentProvider === 'aliyun') {
          configured = !!configMap.aliyun_access_key_id;
        }
        setServiceStatus({ configured, provider: currentProvider, from: emailFrom });
      }
    } catch (error) {
      message.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, []);

  // 保存配置
  const handleSave = async (values: EmailConfig) => {
    setSaving(true);
    try {
      // 通用配置
      const configs: { key: string; value: string }[] = [
        { key: 'email_provider', value: provider },
        { key: 'email_from', value: values.email_from || '' },
        { key: 'email_from_name', value: values.email_from_name || '' },
      ];

      // 根据服务商添加对应配置
      if (provider === 'aliyun_smtp') {
        // 阿里企业邮箱使用SMTP，预设服务器配置
        configs.push(
          { key: 'smtp_host', value: 'smtp.qiye.aliyun.com' },
          { key: 'smtp_port', value: '465' },
          { key: 'smtp_user', value: values.smtp_user || '' },
          { key: 'smtp_password', value: values.smtp_password || '' },
          { key: 'smtp_ssl', value: 'true' },
        );
      } else if (provider === 'smtp') {
        configs.push(
          { key: 'smtp_host', value: values.smtp_host || '' },
          { key: 'smtp_port', value: String(values.smtp_port || 465) },
          { key: 'smtp_user', value: values.smtp_user || '' },
          { key: 'smtp_password', value: values.smtp_password || '' },
          { key: 'smtp_ssl', value: values.smtp_ssl ? 'true' : 'false' },
        );
      } else if (provider === 'resend') {
        configs.push({ key: 'resend_api_key', value: values.resend_api_key || '' });
      } else if (provider === 'aliyun') {
        configs.push(
          { key: 'aliyun_access_key_id', value: values.aliyun_access_key_id || '' },
          { key: 'aliyun_access_key_secret', value: values.aliyun_access_key_secret || '' },
          { key: 'aliyun_region', value: values.aliyun_region || 'cn-hangzhou' },
        );
      }

      for (const config of configs) {
        await updateConfig(config.key, config.value);
      }

      message.success('邮件配置保存成功');
    } catch (error) {
      message.error('保存配置失败');
    } finally {
      setSaving(false);
    }
  };

  // 测试邮件
  const handleTest = async () => {
    if (!testEmail) {
      message.error('请输入测试邮箱地址');
      return;
    }

    setTesting(true);
    try {
      const response = await fetch(`/api/email/test?to=${encodeURIComponent(testEmail)}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      const data = await response.json();
      if (data.code === 200) {
        message.success('测试邮件发送成功，请检查收件箱');
      } else {
        message.error(data.message || '发送失败');
      }
    } catch (error) {
      message.error('测试失败');
    } finally {
      setTesting(false);
    }
  };

  return (
    <Card 
      title={<><MailOutlined /> 邮件服务配置</>} 
      extra={
        serviceStatus.configured ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            {providerNames[serviceStatus.provider] || serviceStatus.provider} · {serviceStatus.from}
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">未配置</Tag>
        )
      }
      loading={loading}
      style={{ marginTop: 16 }}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSave}
        style={{ maxWidth: 600 }}
      >
        {/* 邮件服务商选择 */}
        <Form.Item name="email_provider" label="选择邮件服务">
          <Row gutter={[12, 12]}>
            {[
              { key: 'aliyun_smtp', icon: <CloudOutlined />, name: '阿里企业邮箱', desc: '推荐' },
              { key: 'aliyun', icon: <CloudOutlined />, name: '阿里云推送', desc: 'DirectMail' },
              { key: 'resend', icon: <ThunderboltOutlined />, name: 'Resend', desc: '海外推荐' },
              { key: 'smtp', icon: <GlobalOutlined />, name: '其他SMTP', desc: '自定义' },
            ].map((item) => (
              <Col span={12} key={item.key}>
                <div
                  onClick={() => {
                    setProvider(item.key);
                    form.setFieldValue('email_provider', item.key);
                    if (item.key === 'aliyun_smtp') {
                      form.setFieldsValue({ smtp_host: 'smtp.qiye.aliyun.com', smtp_port: 465, smtp_ssl: true });
                    }
                  }}
                  style={{
                    padding: '16px',
                    borderRadius: 8,
                    border: provider === item.key ? '2px solid #1890ff' : '1px solid #d9d9d9',
                    background: provider === item.key ? '#e6f7ff' : '#fff',
                    cursor: 'pointer',
                    transition: 'all 0.2s',
                  }}
                >
                  <div style={{ fontSize: 20, marginBottom: 4 }}>{item.icon}</div>
                  <div style={{ fontWeight: 500 }}>{item.name}</div>
                  <div style={{ fontSize: 12, color: '#999' }}>{item.desc}</div>
                </div>
              </Col>
            ))}
          </Row>
        </Form.Item>

        {provider === 'aliyun_smtp' && (
          <div style={{ background: '#f0f5ff', borderRadius: 8, padding: 16, marginBottom: 16 }}>
            <div style={{ color: '#1890ff', fontWeight: 500, marginBottom: 12 }}>
              <CloudOutlined /> 阿里云企业邮箱配置
            </div>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="smtp_user" label="邮箱账号" rules={[{ required: true, message: '请输入邮箱账号' }]} style={{ marginBottom: 0 }}>
                  <Input prefix={<MailOutlined />} placeholder="admin@yourdomain.com" autoComplete="off" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="smtp_password" label="邮箱密码" rules={[{ required: true, message: '请输入邮箱密码' }]} style={{ marginBottom: 0 }}>
                  <Input.Password prefix={<LockOutlined />} placeholder="邮箱密码" autoComplete="new-password" />
                </Form.Item>
              </Col>
            </Row>
          </div>
        )}

        {provider === 'aliyun' && (
          <div style={{ background: '#fff7e6', borderRadius: 8, padding: 16, marginBottom: 16 }}>
            <div style={{ color: '#fa8c16', fontWeight: 500, marginBottom: 12 }}>
              <CloudOutlined /> 阿里云DirectMail · <a href="https://www.aliyun.com/product/directmail" target="_blank" rel="noreferrer" style={{ fontSize: 12 }}>开通服务 →</a>
            </div>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="aliyun_access_key_id" label="AccessKey ID" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Input prefix={<LockOutlined />} placeholder="LTAI5t..." autoComplete="off" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="aliyun_access_key_secret" label="AccessKey Secret" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Input.Password prefix={<LockOutlined />} placeholder="Secret" autoComplete="off" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="aliyun_region" label="地域" style={{ marginBottom: 0 }}>
              <Input placeholder="cn-hangzhou" style={{ width: 200 }} />
            </Form.Item>
          </div>
        )}

        {provider === 'resend' && (
          <div style={{ background: '#f6ffed', borderRadius: 8, padding: 16, marginBottom: 16 }}>
            <div style={{ color: '#52c41a', fontWeight: 500, marginBottom: 12 }}>
              <ThunderboltOutlined /> Resend · 每月免费3000封 · <a href="https://resend.com" target="_blank" rel="noreferrer" style={{ fontSize: 12 }}>获取API Key →</a>
            </div>
            <Form.Item name="resend_api_key" label="API Key" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
              <Input.Password prefix={<LockOutlined />} placeholder="re_xxxxxxxx" autoComplete="off" />
            </Form.Item>
          </div>
        )}

        {provider === 'smtp' && (
          <div style={{ background: '#f5f5f5', borderRadius: 8, padding: 16, marginBottom: 16 }}>
            <div style={{ color: '#666', fontWeight: 500, marginBottom: 12 }}>
              <GlobalOutlined /> 自定义SMTP服务器
            </div>
            <Row gutter={16}>
              <Col span={16}>
                <Form.Item name="smtp_host" label="服务器地址" style={{ marginBottom: 8 }}>
                  <Input prefix={<GlobalOutlined />} placeholder="smtp.qq.com" />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="smtp_port" label="端口" style={{ marginBottom: 8 }}>
                  <InputNumber style={{ width: '100%' }} min={1} max={65535} placeholder="465" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="smtp_user" label="用户名" style={{ marginBottom: 8 }}>
                  <Input prefix={<MailOutlined />} placeholder="邮箱地址" autoComplete="off" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="smtp_password" label="密码/授权码" style={{ marginBottom: 8 }}>
                  <Input.Password prefix={<LockOutlined />} placeholder="授权码" autoComplete="new-password" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="smtp_ssl" valuePropName="checked" style={{ marginBottom: 0 }}>
              <Switch checkedChildren="SSL加密" unCheckedChildren="不加密" />
            </Form.Item>
          </div>
        )}

        <Divider />

        {/* 通用配置 */}
        <Form.Item
          name="email_from"
          label="发件人邮箱"
          rules={[{ required: true, message: '请输入发件人邮箱' }, { type: 'email', message: '邮箱格式不正确' }]}
          extra={provider === 'aliyun' ? '需在阿里云控制台配置发信地址' : provider === 'resend' ? '需在Resend中验证域名' : ''}
        >
          <Input prefix={<MailOutlined />} placeholder="noreply@yourdomain.com" />
        </Form.Item>

        <Form.Item name="email_from_name" label="发件人名称">
          <Input placeholder="Emby Hub" />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={saving}>保存配置</Button>
        </Form.Item>
      </Form>

      {/* 测试邮件 */}
      <div style={{ 
        marginTop: 24, 
        padding: '20px', 
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', 
        borderRadius: 12,
        color: '#fff'
      }}>
        <div style={{ marginBottom: 16, fontWeight: 600, fontSize: 16 }}>
          <SendOutlined style={{ marginRight: 8 }} />
          发送测试邮件
        </div>
        <Space.Compact style={{ width: '100%', maxWidth: 400 }}>
          <Input 
            placeholder="输入收件邮箱地址" 
            value={testEmail} 
            onChange={(e) => setTestEmail(e.target.value)} 
            style={{ height: 40 }}
          />
          <Button 
            type="default" 
            onClick={handleTest} 
            loading={testing}
            style={{ height: 40, fontWeight: 500 }}
          >
            发送测试
          </Button>
        </Space.Compact>
        <div style={{ marginTop: 10, fontSize: 12, opacity: 0.8 }}>
          保存配置后发送测试邮件验证配置是否正确
        </div>
      </div>
    </Card>
  );
};

export default EmailSettings;

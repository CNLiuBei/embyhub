import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, message, Space, Descriptions } from 'antd';
import { SaveOutlined, ReloadOutlined } from '@ant-design/icons';
import { getConfigs, updateConfig } from '@/api/config';
import type { SystemConfig } from '@/types';
import EmailSettings from '@/components/EmailSettings';

const SystemConfigPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [form] = Form.useForm();

  // 加载配置
  const loadConfigs = async () => {
    setLoading(true);
    try {
      const response = await getConfigs();
      if (response.code === 200 && response.data) {
        // 如果是数组，直接使用；如果是对象，取list属性
        const configList = Array.isArray(response.data) 
          ? response.data 
          : ((response.data as any).list || []);
        setConfigs(configList);
        
        // 设置表单初始值
        const initialValues: any = {};
        configList.forEach((config: SystemConfig) => {
          initialValues[config.config_key] = config.config_value;
        });
        form.setFieldsValue(initialValues);
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

  // 提交所有配置
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      // 逐个更新配置
      for (const key in values) {
        await updateConfig(key, values[key]);
      }
      
      message.success('配置保存成功');
      loadConfigs();
    } catch (error) {
      message.error('配置保存失败');
    }
  };

  return (
    <div>
      {/* 页面头部 */}
      <div className="page-header">
        <h1>系统设置</h1>
        <p>配置Emby服务器连接和系统参数</p>
      </div>

      <Card
        title="Emby服务器配置"
        loading={loading}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadConfigs}>
              刷新
            </Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSubmit}>
              保存配置
            </Button>
          </Space>
        }
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="emby_server_url"
            label="Emby服务器地址"
            rules={[{ required: true, message: '请输入Emby服务器地址' }]}
          >
            <Input placeholder="http://192.168.1.100:8096" />
          </Form.Item>

          <Form.Item
            name="emby_api_key"
            label="Emby API密钥"
            rules={[{ required: true, message: '请输入Emby API密钥' }]}
          >
            <Input.Password placeholder="请输入API密钥" autoComplete="off" />
          </Form.Item>

          <Form.Item
            name="emby_sync_interval"
            label="同步周期（秒）"
            rules={[{ required: true, message: '请输入同步周期' }]}
          >
            <Input type="number" placeholder="3600" />
          </Form.Item>
        </Form>
      </Card>

      <Card title="配置列表" style={{ marginTop: 24 }} loading={loading}>
        <Descriptions bordered column={1}>
          {configs
            .filter(c => !c.config_key.startsWith('smtp_')) // 过滤掉SMTP配置（在邮件设置中显示）
            .map((config) => (
            <Descriptions.Item
              key={config.config_key}
              label={config.description || config.config_key}
            >
              {config.config_value || '-'}
            </Descriptions.Item>
          ))}
        </Descriptions>
      </Card>

      {/* 邮件服务配置 */}
      <EmailSettings />
    </div>
  );
};

export default SystemConfigPage;

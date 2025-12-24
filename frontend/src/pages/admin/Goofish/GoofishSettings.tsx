import React, { useState, useEffect } from 'react';
import { Card, Form, Input, InputNumber, Switch, Button, Alert, Typography, Space, Divider, App } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { goofishApi, GoofishConfig, SaveConfigRequest } from '../../../services/goofishApi';

const { Text, Paragraph } = Typography;

const GoofishSettings: React.FC = () => {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [config, setConfig] = useState<GoofishConfig | null>(null);

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const res = await goofishApi.getConfig();
      // 后端返回格式: {code: 200, data: {...}, message: "success"}
      // axios响应: res.data = {code: 200, data: {...}, message: "success"}
      const configData = res.data.data;
      setConfig(configData);
      form.setFieldsValue({
        app_id: configData.app_id || undefined,
        app_secret: configData.app_secret || '',
        mch_id: configData.mch_id || '',
        mch_secret: configData.mch_secret || '',
        enabled: configData.enabled || false,
      });
    } catch (err) {
      console.error('获取配置失败:', err);
      message.error('获取配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConfig();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSave = async (values: SaveConfigRequest) => {
    setSaving(true);
    try {
      // 如果密钥没有修改（与原值相同），则不提交
      const submitData: SaveConfigRequest = {
        app_id: values.app_id,
        mch_id: values.mch_id,
        enabled: values.enabled,
      };
      // 只有当密钥被修改时才提交
      if (values.app_secret && values.app_secret !== config?.app_secret) {
        submitData.app_secret = values.app_secret;
      }
      if (values.mch_secret && values.mch_secret !== config?.mch_secret) {
        submitData.mch_secret = values.mch_secret;
      }
      await goofishApi.saveConfig(submitData);
      message.success('保存成功');
      fetchConfig();
    } catch (error: any) {
      message.error(error.response?.data?.error || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <Card title="闲管家对接配置" loading={loading}>
        <Alert
          message="配置说明"
          description={
            <div>
              <p>1. 在闲管家开放平台创建应用，获取 app_id 和 app_secret</p>
              <p>2. 设置货源商户ID (mch_id) 和密钥 (mch_secret)，这两个值由您自定义</p>
              <p>3. 将下方的接口网关地址填写到闲管家平台</p>
              <p>4. 启用后，闲管家平台即可调用本系统的卡密接口</p>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        {config?.gateway_url && (
          <Card size="small" style={{ marginBottom: 24, background: '#f5f5f5' }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Text strong>接口网关地址：</Text>
              <Paragraph copyable style={{ margin: 0 }}>
                {config.gateway_url}
              </Paragraph>
              <Text type="secondary">请将此地址填写到闲管家平台的货源接口配置中</Text>
            </Space>
          </Card>
        )}

        <Form
          form={form}
          name="goofish_settings_form"
          layout="vertical"
          onFinish={handleSave}
          initialValues={{ enabled: false }}
        >
          <Divider orientation="left">闲管家平台配置</Divider>
          
          <Form.Item
            name="app_id"
            label="应用ID (app_id)"
            rules={[{ required: true, message: '请输入应用ID' }]}
            tooltip="闲管家开放平台提供的应用ID"
          >
            <InputNumber 
              style={{ width: '100%' }} 
              placeholder="请输入闲管家应用ID"
              controls={false}
            />
          </Form.Item>

          <Form.Item
            name="app_secret"
            label="应用密钥 (app_secret)"
            tooltip="闲管家开放平台提供的应用密钥，点击小眼睛查看完整密钥"
          >
            <Input.Password placeholder="请输入应用密钥" />
          </Form.Item>

          <Divider orientation="left">货源商户配置</Divider>

          <Form.Item
            name="mch_id"
            label="商户ID (mch_id)"
            rules={[{ required: true, message: '请输入商户ID' }]}
            tooltip="您自定义的货源商户ID，用于签名验证"
          >
            <Input placeholder="请输入商户ID，如: 1001" />
          </Form.Item>

          <Form.Item
            name="mch_secret"
            label="商户密钥 (mch_secret)"
            tooltip="您自定义的商户密钥，用于签名验证，点击小眼睛查看完整密钥"
          >
            <Input.Password placeholder="请输入商户密钥" />
          </Form.Item>

          <Divider />

          <Form.Item
            name="enabled"
            label="启用状态"
            valuePropName="checked"
          >
            <Switch checkedChildren="已启用" unCheckedChildren="已禁用" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={saving} icon={<SaveOutlined />}>
              保存配置
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default GoofishSettings;

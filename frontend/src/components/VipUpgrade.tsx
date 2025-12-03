import React, { useState } from 'react';
import { Modal, Form, Input, Button, message, Result } from 'antd';
import { CrownOutlined, KeyOutlined } from '@ant-design/icons';
import { useVipCard } from '@/api/cardKey';

interface VipUpgradeProps {
  visible: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

const VipUpgrade: React.FC<VipUpgradeProps> = ({ visible, onClose, onSuccess }) => {
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [resultData, setResultData] = useState<{ vip_expire_at: string } | null>(null);
  const [form] = Form.useForm();

  const handleSubmit = async (values: { card_code: string }) => {
    setLoading(true);
    try {
      const response: any = await useVipCard(values.card_code.toUpperCase());
      if (response.code === 200) {
        setResultData(response.data);
        setSuccess(true);
        form.resetFields();
      } else {
        message.error(response.message || 'VIP升级失败');
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || 'VIP升级失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (success) {
      onSuccess?.();
    }
    setSuccess(false);
    setResultData(null);
    onClose();
  };

  return (
    <Modal
      title={
        success ? null : (
          <span>
            <CrownOutlined style={{ color: '#faad14', marginRight: 8 }} />
            VIP会员升级
          </span>
        )
      }
      open={visible}
      onCancel={handleClose}
      footer={null}
      width={400}
    >
      {success ? (
        <Result
          status="success"
          icon={<CrownOutlined style={{ color: '#faad14', fontSize: 72 }} />}
          title="VIP升级成功！"
          subTitle={
            <div>
              <p style={{ marginBottom: 8 }}>恭喜您成为VIP会员</p>
              <p style={{ color: '#666' }}>
                VIP到期时间：{resultData?.vip_expire_at ? new Date(resultData.vip_expire_at).toLocaleDateString('zh-CN') : '-'}
              </p>
            </div>
          }
          extra={[
            <Button type="primary" key="close" onClick={handleClose}>
              确定
            </Button>
          ]}
        />
      ) : (
        <>
          <div style={{ marginBottom: 16, padding: 12, background: '#fffbe6', borderRadius: 4 }}>
            <p style={{ margin: 0, fontSize: 13 }}>
              💡 使用VIP升级码可以延长您的VIP会员时长
            </p>
            <p style={{ margin: '8px 0 0', fontSize: 12, color: '#666' }}>
              • 如果您已是VIP，时长会在原有基础上叠加<br/>
              • 如果您不是VIP，将从现在开始计算
            </p>
          </div>

          <Form form={form} onFinish={handleSubmit} layout="vertical">
            <Form.Item
              name="card_code"
              label="VIP升级码"
              rules={[
                { required: true, message: '请输入VIP升级码' },
                { pattern: /^TL\|[A-Z0-9]{24}$/, message: '卡密格式不正确' },
              ]}
            >
              <Input
                prefix={<KeyOutlined />}
                placeholder="TL|XXXXXXXXXXXXXXXXXXXXXXXX"
                style={{ textTransform: 'uppercase' }}
                size="large"
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={loading} block size="large">
                <CrownOutlined /> 立即升级
              </Button>
            </Form.Item>
          </Form>

          <div style={{ marginTop: 16, textAlign: 'center', color: '#999', fontSize: 12 }}>
            没有VIP升级码？请联系管理员获取
          </div>
        </>
      )}
    </Modal>
  );
};

export default VipUpgrade;

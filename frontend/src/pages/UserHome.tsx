import React, { useState } from 'react';
import { Row, Col, Tag, Button, Modal, Form, Input, message } from 'antd';
import { UserOutlined, CrownOutlined, KeyOutlined, LockOutlined, MailOutlined, CalendarOutlined, CheckCircleOutlined, CloseCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { useSelector } from 'react-redux';
import type { RootState } from '@/store';
import VipUpgrade from '@/components/VipUpgrade';
import request from '@/utils/request';

const UserHome: React.FC = () => {
  const userInfo = useSelector((state: RootState) => state.auth.userInfo);
  const [vipModalVisible, setVipModalVisible] = useState(false);
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  // 修改密码
  const handleChangePassword = async (values: { new_password: string }) => {
    setLoading(true);
    try {
      const response: any = await request.put('/auth/password', {
        password: values.new_password,
      });
      if (response.code === 200) {
        message.success('密码修改成功，Emby密码已同步更新');
        form.resetFields();
        setPasswordModalVisible(false);
      } else {
        message.error(response.message || '修改失败');
      }
    } catch (error) {
      message.error('修改密码失败');
    } finally {
      setLoading(false);
    }
  };

  // 检查VIP状态
  const isVip = userInfo?.vip_level === 1;
  const vipExpireAt = userInfo?.vip_expire_at ? new Date(userInfo.vip_expire_at) : null;
  const isVipExpired = vipExpireAt ? vipExpireAt < new Date() : true;
  const remainingDays = vipExpireAt && !isVipExpired 
    ? Math.ceil((vipExpireAt.getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24))
    : 0;

  return (
    <div style={{ padding: '0 4px' }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ fontSize: 28, fontWeight: 700, color: '#1d1d1f', margin: 0, letterSpacing: '-0.5px' }}>
          个人中心
        </h1>
        <p style={{ color: '#86868b', marginTop: 4, fontSize: 14, margin: '4px 0 0' }}>
          管理您的账号信息和VIP会员
        </p>
      </div>

      <Row gutter={[20, 20]}>
        {/* 用户信息卡片 */}
        <Col xs={24} lg={12}>
          <div style={{
            background: 'rgba(255, 255, 255, 0.5)',
            backdropFilter: 'blur(20px) saturate(180%)',
            WebkitBackdropFilter: 'blur(20px) saturate(180%)',
            borderRadius: 16,
            padding: 24,
            boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
            border: '1px solid rgba(255, 255, 255, 0.4)',
            height: '100%',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <div style={{
                  width: 48,
                  height: 48,
                  borderRadius: 12,
                  background: 'linear-gradient(135deg, #007aff 0%, #5856d6 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}>
                  <UserOutlined style={{ fontSize: 24, color: '#fff' }} />
                </div>
                <div>
                  <div style={{ fontSize: 18, fontWeight: 600, color: '#1d1d1f' }}>账号信息</div>
                  <div style={{ fontSize: 13, color: '#86868b' }}>管理您的基本信息</div>
                </div>
              </div>
              <Button 
                icon={<LockOutlined />} 
                onClick={() => setPasswordModalVisible(true)}
                style={{ borderRadius: 8 }}
              >
                修改密码
              </Button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px', background: '#f9f9f9', borderRadius: 10 }}>
                <UserOutlined style={{ fontSize: 18, color: '#007aff' }} />
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 12, color: '#86868b' }}>用户名</div>
                  <div style={{ fontSize: 15, fontWeight: 500, color: '#1d1d1f' }}>{userInfo?.username}</div>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px', background: '#f9f9f9', borderRadius: 10 }}>
                <MailOutlined style={{ fontSize: 18, color: '#34c759' }} />
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 12, color: '#86868b' }}>邮箱</div>
                  <div style={{ fontSize: 15, fontWeight: 500, color: '#1d1d1f' }}>{userInfo?.email || '未设置'}</div>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px', background: '#f9f9f9', borderRadius: 10 }}>
                {userInfo?.emby_user_id ? (
                  <CheckCircleOutlined style={{ fontSize: 18, color: '#34c759' }} />
                ) : (
                  <CloseCircleOutlined style={{ fontSize: 18, color: '#8e8e93' }} />
                )}
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 12, color: '#86868b' }}>Emby账号</div>
                  <div style={{ fontSize: 15, fontWeight: 500, color: '#1d1d1f' }}>
                    {userInfo?.emby_user_id ? '已关联' : '未关联'}
                  </div>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px', background: '#f9f9f9', borderRadius: 10 }}>
                <CalendarOutlined style={{ fontSize: 18, color: '#ff9500' }} />
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 12, color: '#86868b' }}>注册时间</div>
                  <div style={{ fontSize: 15, fontWeight: 500, color: '#1d1d1f' }}>
                    {userInfo?.created_at ? new Date(userInfo.created_at).toLocaleDateString('zh-CN') : '-'}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Col>

        {/* VIP状态卡片 */}
        <Col xs={24} lg={12}>
          <div style={{
            background: isVip && !isVipExpired 
              ? 'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)'
              : '#fff',
            borderRadius: 16,
            padding: 24,
            boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
            height: '100%',
            position: 'relative',
            overflow: 'hidden',
          }}>
            {isVip && !isVipExpired && (
              <div style={{
                position: 'absolute',
                top: -30,
                right: -30,
                width: 120,
                height: 120,
                background: 'rgba(255,255,255,0.2)',
                borderRadius: '50%',
              }} />
            )}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <div style={{
                  width: 48,
                  height: 48,
                  borderRadius: 12,
                  background: isVip && !isVipExpired 
                    ? 'linear-gradient(135deg, #f5af19 0%, #f12711 100%)'
                    : '#e5e5e5',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}>
                  <CrownOutlined style={{ fontSize: 24, color: '#fff' }} />
                </div>
                <div>
                  <div style={{ fontSize: 18, fontWeight: 600, color: isVip && !isVipExpired ? '#8b4513' : '#1d1d1f' }}>
                    VIP会员
                  </div>
                  <div style={{ fontSize: 13, color: isVip && !isVipExpired ? '#a0522d' : '#86868b' }}>
                    享受专属特权
                  </div>
                </div>
              </div>
              <Button 
                type="primary"
                icon={<KeyOutlined />} 
                onClick={() => setVipModalVisible(true)}
                style={{ 
                  borderRadius: 8,
                  background: isVip && !isVipExpired ? '#f5af19' : undefined,
                  borderColor: isVip && !isVipExpired ? '#f5af19' : undefined,
                }}
              >
                {isVip && !isVipExpired ? '续费VIP' : '升级VIP'}
              </Button>
            </div>

            {isVip && !isVipExpired ? (
              <div style={{ textAlign: 'center', padding: '20px 0' }}>
                <div style={{ 
                  display: 'inline-flex', 
                  alignItems: 'baseline', 
                  gap: 4,
                  marginBottom: 12,
                }}>
                  <span style={{ fontSize: 56, fontWeight: 700, color: '#8b4513' }}>{remainingDays}</span>
                  <span style={{ fontSize: 18, color: '#a0522d' }}>天</span>
                </div>
                <div style={{ color: '#a0522d', fontSize: 14 }}>
                  到期时间：{vipExpireAt?.toLocaleDateString('zh-CN')}
                </div>
                <Tag color="gold" style={{ marginTop: 16, fontSize: 14, padding: '4px 16px' }}>
                  尊贵VIP会员
                </Tag>
              </div>
            ) : (
              <div style={{ textAlign: 'center', padding: '30px 0' }}>
                <CrownOutlined style={{ fontSize: 56, color: '#d9d9d9' }} />
                <div style={{ marginTop: 16, fontSize: 16, color: '#86868b' }}>
                  {isVip ? 'VIP已过期' : '您还不是VIP会员'}
                </div>
                <div style={{ marginTop: 8, color: '#86868b', fontSize: 14 }}>
                  升级VIP享受更多观影特权
                </div>
              </div>
            )}
          </div>
        </Col>
      </Row>

      {/* 使用说明 */}
      <div style={{
        marginTop: 20,
        background: 'rgba(255, 255, 255, 0.5)',
        backdropFilter: 'blur(20px) saturate(180%)',
        WebkitBackdropFilter: 'blur(20px) saturate(180%)',
        borderRadius: 16,
        padding: 24,
        boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
        border: '1px solid rgba(255, 255, 255, 0.4)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
          <InfoCircleOutlined style={{ fontSize: 20, color: '#007aff' }} />
          <span style={{ fontSize: 16, fontWeight: 600, color: '#1d1d1f' }}>使用说明</span>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {[
            '您的Emby账号与本系统账号已关联，可直接登录Emby服务器观看内容',
            '如需升级VIP，请点击上方"升级VIP"按钮，输入VIP升级码即可',
            'VIP到期后可使用新的VIP码续费，时长会叠加',
            '如有问题请联系管理员',
          ].map((text, index) => (
            <div key={index} style={{ 
              display: 'flex', 
              alignItems: 'flex-start', 
              gap: 10,
              padding: '10px 14px',
              background: '#f9f9f9',
              borderRadius: 8,
            }}>
              <span style={{ 
                minWidth: 20, 
                height: 20, 
                borderRadius: '50%', 
                background: '#007aff', 
                color: '#fff', 
                fontSize: 12,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}>
                {index + 1}
              </span>
              <span style={{ color: '#1d1d1f', fontSize: 14, lineHeight: 1.6 }}>{text}</span>
            </div>
          ))}
        </div>
      </div>

      {/* VIP升级弹窗 */}
      <VipUpgrade 
        visible={vipModalVisible} 
        onClose={() => setVipModalVisible(false)}
        onSuccess={() => window.location.reload()}
      />

      {/* 修改密码弹窗 */}
      <Modal
        title={<><LockOutlined /> 修改密码</>}
        open={passwordModalVisible}
        onCancel={() => {
          setPasswordModalVisible(false);
          form.resetFields();
        }}
        footer={null}
      >
        <Form form={form} onFinish={handleChangePassword} layout="vertical">
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6位' },
            ]}
          >
            <Input.Password placeholder="请输入新密码（至少6位）" autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label="确认密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" autoComplete="new-password" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" loading={loading} block>
              确认修改
            </Button>
          </Form.Item>
        </Form>
        <div style={{ marginTop: 12, color: '#999', fontSize: 12 }}>
          注意：修改密码后Emby账号密码也会同步更新
        </div>
      </Modal>
    </div>
  );
};

export default UserHome;

/**
 * 卡密兑换页面
 */

import React, { useState } from 'react';
import { Card, Input, Button, Result, Space, List, Typography, App } from 'antd';
import { GiftOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { cardUserApi } from '../../services/cardApi';
import type { RedeemOrder } from '../../types/card.types';
import { formatDateTime } from '../../utils/cardHelpers';

const { Title, Text, Paragraph } = Typography;

const RedeemCard: React.FC = () => {
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [code, setCode] = useState('');
  const [redeemResult, setRedeemResult] = useState<RedeemOrder | null>(null);

  // 获取兑换历史
  const { data: historyData, refetch: refetchHistory } = useQuery({
    queryKey: ['redeemHistory'],
    queryFn: () => cardUserApi.getRedeemHistory({ page: 1, page_size: 5 })
  });

  // 兑换mutation
  const redeemMutation = useMutation({
    mutationFn: (cardCode: string) => cardUserApi.redeem({ code: cardCode.trim() }),
    onSuccess: (data) => {
      setRedeemResult(data);
      setCode('');
      message.success('🎉 兑换成功！会员权益已生效');
      
      // 刷新兑换历史
      refetchHistory();
      
      // 刷新所有可能的用户相关查询
      queryClient.invalidateQueries({ queryKey: ['userInfo'] });
      queryClient.invalidateQueries({ queryKey: ['redeemHistory'] });
      
      // 延迟后重新加载页面以确保状态更新
      setTimeout(() => {
        window.location.reload();
      }, 2000);
    },
    onError: (error: any) => {
      message.error(error.response?.data?.message || '兑换失败，请检查卡密是否正确');
    }
  });

  const handleRedeem = () => {
    const trimmedCode = code.trim();
    
    if (!trimmedCode) {
      message.warning('请输入卡密');
      return;
    }
    
    // 基本格式验证：至少包含前缀和24位字符
    if (trimmedCode.length < 28) {
      message.error('卡密格式不正确，请检查后重试');
      return;
    }

    redeemMutation.mutate(trimmedCode);
  };

  // 显示成功结果
  if (redeemResult) {
    return (
      <div style={{ maxWidth: 600, margin: '40px auto' }}>
        <Result
          status="success"
          icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
          title="兑换成功！"
          subTitle={`恭喜您成为会员！会员时长：${redeemResult.duration}天`}
          extra={[
            <div key="info" style={{ textAlign: 'left', margin: '20px 0' }}>
              <Card>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text strong>会员到期时间：</Text>
                    <Text>{formatDateTime(redeemResult.expire_time)}</Text>
                  </div>
                  <div>
                    <Text strong>订单号：</Text>
                    <Text copyable>{redeemResult.order_no}</Text>
                  </div>
                </Space>
              </Card>
              
              <Card style={{ marginTop: 16 }} title="🎁 您可以享受以下特权">
                <ul style={{ paddingLeft: 20 }}>
                  <li>✓ 无限观看所有影片</li>
                  <li>✓ 高清4K画质</li>
                  <li>✓ 多设备同时在线</li>
                  <li>✓ 离线下载功能</li>
                </ul>
              </Card>
            </div>,
            <Space key="actions">
              <Button type="primary" size="large" href="/media">
                去看看新内容
              </Button>
              <Button size="large" onClick={() => setRedeemResult(null)}>
                返回
              </Button>
            </Space>
          ]}
        />
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 600, margin: '40px auto' }}>
      <Card
        title={
          <div style={{ textAlign: 'center' }}>
            <GiftOutlined style={{ fontSize: 40, color: '#1890ff', marginBottom: 16 }} />
            <Title level={3} style={{ margin: 0 }}>兑换卡密</Title>
          </div>
        }
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Text strong style={{ marginBottom: 8, display: 'block' }}>
              请输入您的卡密
            </Text>
            <Input
              size="large"
              placeholder="True-XXXXXXXXXXXXXXXXXXXXXXXX"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              maxLength={50}
              onPressEnter={handleRedeem}
              style={{ fontSize: 14, letterSpacing: 1, fontFamily: 'monospace' }}
            />
            <div style={{ marginTop: 8, color: '#8c8c8c', fontSize: 12 }}>
              💡 卡密格式：前缀-24位字符<br />
              💡 示例：True-ABCDEFGHJKLMNPQRSTUWXYZ2
            </div>
          </div>

          <Button
            type="primary"
            size="large"
            block
            onClick={handleRedeem}
            loading={redeemMutation.isPending}
            icon={<GiftOutlined />}
          >
            立即兑换
          </Button>

          <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
            ⏱️ 今日剩余兑换次数：根据限流规则自动管理
          </div>
        </Space>
      </Card>

      {historyData && historyData.list.length > 0 && (
        <Card title="📜 我的兑换记录" style={{ marginTop: 24 }}>
          <List
            dataSource={historyData.list}
            renderItem={(item: RedeemOrder) => (
              <List.Item>
                <List.Item.Meta
                  title={`${item.duration}天会员`}
                  description={formatDateTime(item.redeem_time)}
                />
                <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 20 }} />
              </List.Item>
            )}
          />
          {historyData.total > 5 && (
            <div style={{ textAlign: 'center', marginTop: 16 }}>
              <Button type="link" href="/user/redeem-history">
                查看更多
              </Button>
            </div>
          )}
        </Card>
      )}

      <Card title="📋 兑换说明" style={{ marginTop: 24 }}>
        <Paragraph>
          <ul style={{ paddingLeft: 20 }}>
            <li>每个卡密仅可使用一次</li>
            <li>兑换成功后立即生效</li>
            <li>会员时长可累加</li>
            <li>如有问题请联系客服</li>
          </ul>
        </Paragraph>
      </Card>
    </div>
  );
};

export default RedeemCard;

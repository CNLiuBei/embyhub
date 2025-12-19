/**
 * 会员购买页面（已废弃）
 * @deprecated 此功能已迁移到新的VIP购买页面
 */

import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import { CrownOutlined } from '@ant-design/icons';

const PurchaseMember: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="flex items-center justify-center" style={{ minHeight: '60vh' }}>
      <Result
        icon={<CrownOutlined style={{ fontSize: 72, color: '#faad14' }} />}
        title="功能已迁移"
        subTitle="会员购买功能已迁移到新的VIP购买页面，提供更好的体验和更安全的支付方式"
        extra={[
          <Button
            type="primary"
            key="vip"
            size="large"
            icon={<CrownOutlined />}
            onClick={() => navigate('/user/vip-purchase')}
          >
            前往VIP购买页面
          </Button>,
          <Button
            key="back"
            size="large"
            onClick={() => navigate('/user/member')}
          >
            返回会员中心
          </Button>,
        ]}
      />
    </div>
  );
};

export default PurchaseMember;

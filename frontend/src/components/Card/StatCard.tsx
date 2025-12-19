/**
 * 统计卡片组件
 */

import React from 'react';
import { Card } from 'antd';
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  StopOutlined
} from '@ant-design/icons';

interface StatCardProps {
  title: string;
  value: number | string;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  icon?: 'file' | 'check' | 'close' | 'stop' | React.ReactNode;
  color?: string;
  loading?: boolean;
  onClick?: () => void;
}

const ICON_MAP: Record<string, React.ReactNode> = {
  file: <FileTextOutlined />,
  check: <CheckCircleOutlined />,
  close: <CloseCircleOutlined />,
  stop: <StopOutlined />
};

export const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  trend,
  icon,
  color = '#1890ff',
  loading = false,
  onClick
}) => {
  const iconElement = typeof icon === 'string' ? ICON_MAP[icon as keyof typeof ICON_MAP] : icon;

  return (
    <Card
      hoverable={!!onClick}
      onClick={onClick}
      loading={loading}
      style={{
        borderRadius: '8px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
        transition: 'all 0.3s ease'
      }}
      styles={{
        body: { padding: '20px 24px' }
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div style={{ flex: 1 }}>
          <div style={{
            fontSize: '14px',
            color: '#8c8c8c',
            marginBottom: '8px'
          }}>
            {title}
          </div>
          
          <div style={{
            fontSize: '30px',
            fontWeight: 600,
            color: '#262626',
            marginBottom: '4px'
          }}>
            {value}
          </div>
          
          {trend && (
            <div style={{
              fontSize: '14px',
              color: trend.isPositive ? '#52c41a' : '#ff4d4f',
              display: 'flex',
              alignItems: 'center',
              gap: '4px'
            }}>
              {trend.isPositive ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
              <span>{Math.abs(trend.value)}%</span>
            </div>
          )}
        </div>
        
        {iconElement && (
          <div style={{
            fontSize: '40px',
            color,
            opacity: 0.8
          }}>
            {iconElement}
          </div>
        )}
      </div>
    </Card>
  );
};

export default StatCard;

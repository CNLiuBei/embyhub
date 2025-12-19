/**
 * 状态标签组件
 */

import React from 'react';
import { Tag } from 'antd';
import { CardStatus } from '../../types/card.types';
import { getStatusConfig } from '../../utils/cardHelpers';

interface StatusBadgeProps {
  status: CardStatus;
  showIcon?: boolean;
  size?: 'small' | 'default';
}

const STATUS_ICONS = {
  [CardStatus.UNUSED]: '⚪',
  [CardStatus.USED]: '✓',
  [CardStatus.EXPIRED]: '✕',
  [CardStatus.DISABLED]: '⊘'
};

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  showIcon = false,
  size = 'default'
}) => {
  const config = getStatusConfig(status);
  
  return (
    <Tag
      color={config.color}
      style={{
        backgroundColor: config.backgroundColor,
        borderColor: config.color,
        fontSize: size === 'small' ? '12px' : '14px',
        padding: size === 'small' ? '2px 8px' : '4px 12px',
        margin: 0
      }}
    >
      {showIcon && STATUS_ICONS[status]} {config.label}
    </Tag>
  );
};

export default StatusBadge;

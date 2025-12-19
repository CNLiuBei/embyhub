/**
 * 卡密相关工具函数
 */

import { CardStatus, CardType, type CardStatusConfig, type CardTypeConfigMap } from '../types/card.types';

/**
 * 卡密状态配置
 */
export const CARD_STATUS_CONFIG: CardStatusConfig = {
  [CardStatus.UNUSED]: {
    label: '未使用',
    color: '#52c41a',
    backgroundColor: '#f6ffed'
  },
  [CardStatus.USED]: {
    label: '已使用',
    color: '#8c8c8c',
    backgroundColor: '#fafafa'
  },
  [CardStatus.EXPIRED]: {
    label: '已过期',
    color: '#ff4d4f',
    backgroundColor: '#fff2f0'
  },
  [CardStatus.DISABLED]: {
    label: '已禁用',
    color: '#bfbfbf',
    backgroundColor: '#f5f5f5'
  }
};

/**
 * 卡密类型配置
 */
export const CARD_TYPE_CONFIG: CardTypeConfigMap = {
  [CardType.MONTHLY]: {
    label: '月卡',
    duration: 30,
    color: '#1890ff'
  },
  [CardType.YEARLY]: {
    label: '年卡',
    duration: 365,
    color: '#722ed1'
  }
};

/**
 * 获取状态标签配置
 */
export const getStatusConfig = (status: CardStatus) => {
  return CARD_STATUS_CONFIG[status] || CARD_STATUS_CONFIG[CardStatus.UNUSED];
};

/**
 * 获取卡密类型配置
 */
export const getCardTypeConfig = (type: CardType) => {
  return CARD_TYPE_CONFIG[type] || CARD_TYPE_CONFIG[CardType.MONTHLY];
};

/**
 * 格式化卡密码
 * 自动添加横杠分隔符
 */
export const formatCardCode = (code: string): string => {
  // 移除所有非字母数字字符
  const cleaned = code.replace(/[^A-Z0-9]/gi, '').toUpperCase();
  
  // 每4个字符添加一个横杠
  const parts = [];
  for (let i = 0; i < cleaned.length; i += 4) {
    parts.push(cleaned.substring(i, i + 4));
  }
  
  return parts.join('-');
};

/**
 * 验证卡密格式
 */
export const validateCardCode = (code: string): boolean => {
  // 移除横杠
  const cleaned = code.replace(/-/g, '');
  
  // 应该是16位字母数字组合
  return /^[A-Z0-9]{16}$/i.test(cleaned);
};

/**
 * 清理卡密码（移除横杠和空格）
 */
export const cleanCardCode = (code: string): string => {
  return code.replace(/[-\s]/g, '').toUpperCase();
};

/**
 * 格式化日期时间
 */
export const formatDateTime = (date: string | undefined): string => {
  if (!date) return '-';
  
  try {
    const d = new Date(date);
    return d.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch {
    return '-';
  }
};

/**
 * 格式化日期
 */
export const formatDate = (date: string | undefined): string => {
  if (!date) return '-';
  
  try {
    const d = new Date(date);
    return d.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    });
  } catch {
    return '-';
  }
};

/**
 * 计算使用率
 */
export const calculateUsageRate = (usedCount: number, totalCount: number): number => {
  if (totalCount === 0) return 0;
  return Math.round((usedCount / totalCount) * 100);
};

/**
 * 格式化百分比
 */
export const formatPercentage = (value: number): string => {
  return `${value}%`;
};

/**
 * 获取剩余天数
 */
export const getRemainingDays = (expireDate: string | undefined): number => {
  if (!expireDate) return 0;
  
  try {
    const now = new Date();
    const expire = new Date(expireDate);
    const diff = expire.getTime() - now.getTime();
    return Math.ceil(diff / (1000 * 60 * 60 * 24));
  } catch {
    return 0;
  }
};

/**
 * 判断是否已过期
 */
export const isExpired = (expireDate: string | undefined): boolean => {
  return getRemainingDays(expireDate) < 0;
};

/**
 * 获取批次号的日期部分
 */
export const getBatchDate = (batchNo: string): string => {
  // 批次号格式: B20251207123456
  if (batchNo.length < 9) return '';
  
  const year = batchNo.substring(1, 5);
  const month = batchNo.substring(5, 7);
  const day = batchNo.substring(7, 9);
  
  return `${year}-${month}-${day}`;
};

/**
 * 生成导出文件名
 */
export const generateExportFilename = (prefix: string, extension: string): string => {
  const now = new Date();
  const timestamp = now.toISOString().replace(/[:.]/g, '-').substring(0, 19);
  return `${prefix}_${timestamp}.${extension}`;
};

/**
 * 复制到剪贴板
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    // 降级方案
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    const success = document.execCommand('copy');
    document.body.removeChild(textarea);
    return success;
  }
};

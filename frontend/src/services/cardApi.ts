/**
 * 卡密API服务
 */

import api from './api';
import type {
  Card,
  CardBatch,
  CardStats,
  CreateBatchRequest,
  CreateBatchResponse,
  CardListRequest,
  BatchListRequest,
  RedeemRequest,
  RedeemOrder,
  RedeemHistoryRequest,
  PaginatedResponse,
  ExportFilter
} from '../types/card.types';

/**
 * 管理员API
 */
export const cardAdminApi = {
  /**
   * 批量生成卡密
   */
  createBatch: async (data: CreateBatchRequest) => {
    const response = await api.post('/admin/card/batch', data);
    return response.data.data as CreateBatchResponse;
  },

  /**
   * 获取批次列表
   */
  getBatchList: async (params: BatchListRequest) => {
    const response = await api.get('/admin/card/batch/list', { params });
    return response.data.data as PaginatedResponse<CardBatch>;
  },

  /**
   * 获取卡密列表
   */
  getCardList: async (params: CardListRequest) => {
    const response = await api.get('/admin/card/list', { params });
    return response.data.data as PaginatedResponse<Card>;
  },

  /**
   * 禁用卡密
   */
  disableCard: (id: number) => {
    return api.post(`/admin/card/${id}/disable`);
  },

  /**
   * 启用卡密
   */
  enableCard: (id: number) => {
    return api.post(`/admin/card/${id}/enable`);
  },

  /**
   * 删除卡密
   */
  deleteCard: (id: number) => {
    return api.delete(`/admin/card/${id}`);
  },

  /**
   * 获取卡密统计
   */
  getStats: async () => {
    const response = await api.get('/admin/card/stats');
    return response.data.data as CardStats;
  },

  /**
   * 导出卡密为CSV
   */
  exportToCsv: (filter: ExportFilter) => {
    return api.get('/admin/card/export/csv', {
      params: filter,
      responseType: 'blob'
    });
  },

  /**
   * 导出卡密为Excel
   */
  exportToExcel: (filter: ExportFilter) => {
    return api.get('/admin/card/export/excel', {
      params: filter,
      responseType: 'blob'
    });
  },

  /**
   * 导出卡密码（纯文本）
   */
  exportCodes: (batchNo: string) => {
    return api.get('/admin/card/export/codes', {
      params: { batch_no: batchNo },
      responseType: 'blob'
    });
  },

  /**
   * 生成使用报告
   */
  generateReport: (batchNo: string) => {
    return api.get('/admin/card/export/report', {
      params: { batch_no: batchNo },
      responseType: 'blob'
    });
  }
};

/**
 * 用户API
 */
export const cardUserApi = {
  /**
   * 兑换卡密
   */
  redeem: async (data: RedeemRequest) => {
    const response = await api.post('/card/redeem', data);
    return response.data.data as RedeemOrder;
  },

  /**
   * 获取兑换历史
   */
  getRedeemHistory: async (params: RedeemHistoryRequest) => {
    const response = await api.get('/card/history', { params });
    return response.data.data as PaginatedResponse<RedeemOrder>;
  }
};

/**
 * 下载文件辅助函数
 */
export const downloadFile = (blob: Blob, filename: string) => {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
};

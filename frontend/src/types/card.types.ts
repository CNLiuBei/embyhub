/**
 * 卡密相关类型定义
 */

// 卡密状态枚举
export enum CardStatus {
  UNUSED = 0,    // 未使用
  USED = 1,      // 已使用
  EXPIRED = 2,   // 已过期
  DISABLED = 3   // 已禁用
}

// 卡密类型枚举
export enum CardType {
  MONTHLY = 1,   // 月卡
  YEARLY = 2     // 年卡
}

// 卡密信息
export interface Card {
  id: number;
  code: string;
  batch_no: string;
  card_type: CardType;
  duration: number;
  status: CardStatus;
  used_by?: string;
  used_by_name?: string;  // 使用者用户名
  used_at?: string;
  expire_at?: string;
  created_by: string;
  remark?: string;
  created_at: string;
  updated_at: string;
}

// 批次信息
export interface CardBatch {
  id: number;
  batch_no: string;
  card_type: CardType;
  duration: number;
  quantity: number;
  used_count: number;
  expire_at?: string;
  created_by: string;
  created_by_name: string;
  remark?: string;
  created_at: string;
}

// 兑换订单
export interface RedeemOrder {
  id: number;
  order_no: string;
  user_id: string;
  card_id: number;
  card_code: string;
  level: number;
  duration: number;
  status: number;
  redeem_time: string;
  expire_time: string;
  created_at: string;
}

// 卡密统计
export interface CardStats {
  total_cards: number;
  unused_cards: number;
  used_cards: number;
  expired_cards: number;
  disabled_cards: number;
  total_batches: number;
}

// 生成卡密请求
export interface CreateBatchRequest {
  card_type: CardType;
  quantity: number;
  duration?: number;
  expire_at?: string;
  remark?: string;
}

// 生成卡密响应
export interface CreateBatchResponse {
  batch: CardBatch;
  cards: Card[];
  codes: string[];
}

// 卡密列表请求
export interface CardListRequest {
  page: number;
  page_size: number;
  batch_no?: string;
  card_type?: CardType;
  status?: CardStatus;
  code?: string;
}

// 批次列表请求
export interface BatchListRequest {
  page: number;
  page_size: number;
}

// 兑换请求
export interface RedeemRequest {
  code: string;
}

// 兑换历史请求
export interface RedeemHistoryRequest {
  page: number;
  page_size: number;
}

// 导出筛选条件
export interface ExportFilter {
  batch_no?: string;
  card_type?: CardType;
  status?: CardStatus;
  limit?: number;
}

// 分页响应
export interface PaginatedResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

// 状态标签配置
export interface StatusBadgeConfig {
  label: string;
  color: string;
  backgroundColor: string;
}

// 卡密类型标签配置
export interface CardTypeConfig {
  label: string;
  duration: number;
  color: string;
}

// 工具函数类型
export type CardStatusConfig = Record<CardStatus, StatusBadgeConfig>;
export type CardTypeConfigMap = Record<CardType, CardTypeConfig>;

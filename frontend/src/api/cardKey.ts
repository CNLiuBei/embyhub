import request from '@/utils/request';

// 卡密类型
export interface CardKey {
  id: number;
  card_code: string;
  card_type: number;
  duration: number;
  status: number;
  used_by?: number;
  used_at?: string;
  expire_at?: string;
  remark?: string;
  created_by: number;
  created_at: string;
  used_by_user?: { username: string };
  created_by_user?: { username: string };
}

// 创建卡密请求
export interface CardKeyCreateRequest {
  count: number;
  card_type: number;
  duration: number;
  remark?: string;
}

// 获取卡密列表
export function getCardKeys(params?: {
  page?: number;
  page_size?: number;
  status?: number;
  card_type?: number;
  keyword?: string;
}) {
  return request.get('/card-keys', { params });
}

// 生成卡密
export function createCardKeys(data: CardKeyCreateRequest) {
  return request.post('/card-keys', data);
}

// 获取卡密详情
export function getCardKey(id: number) {
  return request.get(`/card-keys/${id}`);
}

// 禁用卡密
export function disableCardKey(id: number) {
  return request.put(`/card-keys/${id}/disable`);
}

// 启用卡密
export function enableCardKey(id: number) {
  return request.put(`/card-keys/${id}/enable`);
}

// 删除卡密
export function deleteCardKey(id: number) {
  return request.delete(`/card-keys/${id}`);
}

// 获取卡密统计
export function getCardKeyStatistics() {
  return request.get('/card-keys/statistics');
}

// 验证卡密（公开接口）
export function validateCardKey(cardCode: string) {
  return request.post('/card-keys/validate', { card_code: cardCode });
}

// 使用VIP升级码
export function useVipCard(cardCode: string) {
  return request.post('/card-keys/use-vip', { card_code: cardCode });
}

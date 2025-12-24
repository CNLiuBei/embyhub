import api from './api';

// 闲管家配置
export interface GoofishConfig {
  id: number;
  app_id: number;
  app_secret: string;
  mch_id: string;
  mch_secret: string;
  enabled: boolean;
  gateway_url?: string;
}

export interface SaveConfigRequest {
  app_id: number;
  app_secret?: string;
  mch_id: string;
  mch_secret?: string;
  enabled: boolean;
}

// 商品映射
export interface GoofishGoods {
  id: number;
  goods_no: string;
  goods_name: string;
  card_type: number;
  price: number;
  status: number;
  auto_generate: boolean;
  card_prefix: string;
  duration: number;
  max_auto_generate: number;
  stock: number;
  created_at: string;
  updated_at: string;
}

export interface CreateGoodsRequest {
  goods_no: string;
  goods_name: string;
  card_type: number;
  price: number;
  status?: number;
  auto_generate?: boolean;
  card_prefix?: string;
  duration?: number;
  max_auto_generate?: number;
}

export interface UpdateGoodsRequest {
  goods_name?: string;
  card_type?: number;
  price?: number;
  status?: number;
  auto_generate?: boolean;
  card_prefix?: string;
  duration?: number;
  max_auto_generate?: number;
}

// 订单
export interface GoofishOrder {
  id: number;
  order_no: string;
  out_order_no: string;
  biz_order_no: string;
  goods_no: string;
  goods_name: string;
  buy_quantity: number;
  order_amount: number;
  order_status: number;
  card_codes: string;
  remark: string;
  client_ip: string;
  created_at: string;
  updated_at: string;
}

export interface GoofishOrderCard {
  id: number;
  order_no: string;
  card_id: number;
  card_code: string;
  goods_no: string;
  created_at: string;
}

// API日志
export interface GoofishLog {
  id: number;
  endpoint: string;
  method: string;
  request_body: string;
  response_body: string;
  response_code: number;
  duration: number;
  client_ip: string;
  created_at: string;
}

// API响应包装类型
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 闲管家API服务
export const goofishApi = {
  // 配置管理
  getConfig: () => api.get<ApiResponse<GoofishConfig>>('/admin/goofish/config'),
  saveConfig: (data: SaveConfigRequest) => api.post<ApiResponse<null>>('/admin/goofish/config', data),

  // 商品管理
  getGoodsList: (params?: { page?: number; page_size?: number; status?: number; keyword?: string }) =>
    api.get<ApiResponse<{ total: number; list: GoofishGoods[] }>>('/admin/goofish/goods', { params }),
  createGoods: (data: CreateGoodsRequest) => api.post<ApiResponse<null>>('/admin/goofish/goods', data),
  updateGoods: (id: number, data: UpdateGoodsRequest) => api.put<ApiResponse<null>>(`/admin/goofish/goods/${id}`, data),
  deleteGoods: (id: number) => api.delete<ApiResponse<null>>(`/admin/goofish/goods/${id}`),

  // 订单查询
  getOrderList: (params?: {
    page?: number;
    page_size?: number;
    order_no?: string;
    goods_no?: string;
    biz_order_no?: string;
    status?: number;
    start_time?: string;
    end_time?: string;
  }) => api.get<ApiResponse<{ total: number; list: GoofishOrder[] }>>('/admin/goofish/orders', { params }),
  getOrderDetail: (orderNo: string) =>
    api.get<ApiResponse<{ order: GoofishOrder; cards: GoofishOrderCard[] }>>(`/admin/goofish/orders/${orderNo}`),

  // 日志查询
  getLogs: (params?: {
    page?: number;
    page_size?: number;
    endpoint?: string;
    start_time?: string;
    end_time?: string;
  }) => api.get<ApiResponse<{ total: number; list: GoofishLog[] }>>('/admin/goofish/logs', { params }),
  cleanOldLogs: () => api.post<ApiResponse<{ count: number }>>('/admin/goofish/logs/clean'),

  // 自动生成商品映射
  autoGenerateGoods: () => api.post<ApiResponse<{ created: number; skipped: number; details: string[] }>>('/admin/goofish/goods/auto-generate'),
};

export default goofishApi;

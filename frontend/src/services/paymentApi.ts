import api from './api'

// VIP套餐类型
export interface VipPlan {
  id: number
  name: string
  description: string
  price: number // 分
  duration_days: number
  is_active: boolean
}

// 支付结果
export interface PaymentResult {
  order_no: string
  qr_code: string
  amount: number
}

// 订单状态
export interface OrderStatus {
  order_no: string
  trade_no: string
  trade_status: string
  amount: number
  plan_name: string
  status: string
  created_at: string
  paid_at: string | null
}

// 订单列表响应
export interface OrderListResponse {
  list: OrderStatus[]
  total: number
}

// 会员变动记录
export interface MemberChangeLog {
  id: number
  source: string
  source_name: string
  order_no: string
  change_days: number
  amount: number
  before_expire_at: string | null
  after_expire_at: string | null
  remark: string
  created_at: string
}

// 会员变动记录响应
export interface MemberChangeLogResponse {
  list: MemberChangeLog[]
  total: number
}

// 支付宝配置
export interface AlipayConfig {
  id: number
  app_id: string
  app_public_key: string
  app_private_key: string
  alipay_public_key: string
  notify_url: string
  enabled: boolean
  is_production: boolean
}

// 支付日志
export interface AlipayLog {
  id: number
  order_no: string
  action: string
  request_body: string
  response_body: string
  status: string
  error_msg: string
  client_ip: string
  duration: number
  created_at: string
}

// 支付API
export const paymentApi = {
  // 获取VIP套餐列表
  getVipPlans: () =>
    api.get<{ code: number; data: VipPlan[] }>('/payment/plans'),

  // 创建支付宝支付订单
  createAlipayPayment: (planId: number) =>
    api.post<{ code: number; data: PaymentResult }>('/payment/alipay/create', { plan_id: planId }),

  // 查询订单状态
  getOrderStatus: (orderNo: string) =>
    api.get<{ code: number; data: OrderStatus }>(`/payment/order/${orderNo}`),

  // 获取订单列表
  getOrderList: (params: { page: number; page_size: number; status?: string }) =>
    api.get<{ code: number; data: OrderListResponse }>('/payment/orders', { params }),

  // 获取会员变动记录
  getMemberChangeLogs: (params: { page: number; page_size: number }) =>
    api.get<{ code: number; data: MemberChangeLogResponse }>('/payment/member-logs', { params }),
}

// 支付宝管理API（管理员）
export const alipayAdminApi = {
  // 获取配置
  getConfig: () =>
    api.get<{ code: number; data: AlipayConfig }>('/admin/alipay/config'),

  // 保存配置
  saveConfig: (data: {
    app_id: string
    app_public_key?: string
    app_private_key?: string
    alipay_public_key: string
    notify_url: string
    enabled: boolean
    is_production: boolean
  }) =>
    api.put('/admin/alipay/config', data),

  // 测试连接
  testConnection: () =>
    api.post('/admin/alipay/test'),

  // 获取日志
  getLogs: (params: { page: number; page_size: number; order_no?: string }) =>
    api.get<{ code: number; data: { list: AlipayLog[]; total: number } }>('/admin/alipay/logs', { params }),

  // VIP套餐管理
  getVipPlans: () =>
    api.get<{ code: number; data: VipPlan[] }>('/admin/alipay/plans'),

  createVipPlan: (data: { name: string; description: string; price: number; duration_days: number }) =>
    api.post<{ code: number; data: VipPlan }>('/admin/alipay/plans', data),

  updateVipPlan: (id: number, data: { name: string; description: string; price: number; duration_days: number; is_active: boolean }) =>
    api.put<{ code: number; data: VipPlan }>(`/admin/alipay/plans/${id}`, data),

  deleteVipPlan: (id: number) =>
    api.delete(`/admin/alipay/plans/${id}`),

  toggleVipPlanStatus: (id: number) =>
    api.post<{ code: number; data: VipPlan }>(`/admin/alipay/plans/${id}/toggle`),
}

export default paymentApi


// Cloudflare 隧道状态
export interface TunnelStatus {
  installed: boolean
  version: string
  logged_in: boolean
  configured: boolean
  running: boolean
  tunnel_name: string
  full_domain: string
  local_port: number
  error_msg: string
  notify_url: string
}

// Cloudflare 隧道配置
export interface TunnelConfig {
  id: number
  tunnel_id: string
  tunnel_name: string
  domain: string
  subdomain: string
  full_domain: string
  local_port: number
  status: string
  config_path: string
  cred_path: string
  error_msg: string
  created_at: string
  updated_at: string
}

// Cloudflare 隧道管理 API
export const tunnelApi = {
  // 获取隧道状态
  getStatus: () =>
    api.get<{ code: number; data: TunnelStatus }>('/admin/tunnel/status'),

  // 获取隧道配置
  getConfig: () =>
    api.get<{ code: number; data: TunnelConfig }>('/admin/tunnel/config'),

  // 下载 cloudflared（超时5分钟）
  downloadCloudflared: () =>
    api.post<{ code: number; message: string; data: { path: string } }>('/admin/tunnel/download', {}, { timeout: 300000 }),

  // 创建隧道（超时5分钟，因为需要等待浏览器授权）
  createTunnel: (data: { tunnel_name: string; domain: string; subdomain: string; local_port?: number }) =>
    api.post<{ code: number; data: TunnelConfig }>('/admin/tunnel/create', data, { timeout: 300000 }),

  // 启动隧道
  startTunnel: () =>
    api.post('/admin/tunnel/start'),

  // 停止隧道
  stopTunnel: () =>
    api.post('/admin/tunnel/stop'),

  // 重启隧道
  restartTunnel: () =>
    api.post('/admin/tunnel/restart'),

  // 删除隧道
  deleteTunnel: () =>
    api.delete('/admin/tunnel'),
}

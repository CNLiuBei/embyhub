// 通用响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

// 分页请求参数
export interface PaginationParams {
  page: number
  page_size: number
}

// 分页响应数据
export interface PaginationResponse<T> {
  total: number
  list: T[]
}

// 用户类型
export interface User {
  user_id: number
  username: string
  email: string
  emby_user_id: string
  role_id: number
  status: number
  vip_level: number        // VIP等级：0=普通 1=VIP
  vip_expire_at?: string   // VIP到期时间
  created_at: string
  updated_at: string
  role?: Role
}

// 用户创建请求
export interface UserCreateRequest {
  username: string
  password: string
  email?: string
  emby_user_id?: string
  role_id: number
}

// 用户更新请求
export interface UserUpdateRequest {
  email?: string
  emby_user_id?: string
  role_id?: number
  status?: number
}

// 角色类型
export interface Role {
  role_id: number
  role_name: string
  description: string
  created_at: string
  permissions?: Permission[]
}

// 角色创建请求
export interface RoleCreateRequest {
  role_name: string
  description?: string
}

// 权限类型
export interface Permission {
  permission_id: number
  permission_name: string
  permission_key: string
  description: string
}

// 访问记录类型
export interface AccessRecord {
  record_id: number
  user_id: number
  access_time: string
  resource: string
  ip_address: string
  device_info: string
  user?: User
}

// 访问记录查询请求
export interface AccessRecordQuery extends PaginationParams {
  user_id?: number
  start_time?: string
  end_time?: string
  resource?: string
}

// 系统配置类型
export interface SystemConfig {
  config_key: string
  config_value: string
  description: string
  updated_at: string
}

// 统计数据类型
export interface Statistics {
  total_users: number
  active_users: number
  today_access: number
  top_users: TopUserItem[]
  access_trend: AccessTrendItem[]
  user_growth?: GrowthTrendItem[]
  vip_stats?: VipStatistics
  cardkey_stats?: CardKeyStatistics
}

// 增长趋势项
export interface GrowthTrendItem {
  date: string
  new_users: number
  total_users: number
}

// VIP统计
export interface VipStatistics {
  total_vip: number
  expired_vip: number
  expiring_3_day: number
  expiring_7_day: number
}

// 卡密统计
export interface CardKeyStatistics {
  total_cards: number
  unused_cards: number
  used_cards: number
  disabled_cards: number
}

// Top用户项
export interface TopUserItem {
  user_id: number
  username: string
  access_count: number
}

// 访问趋势项
export interface AccessTrendItem {
  date: string
  count: number
}

// Emby用户类型
export interface EmbyUser {
  id: string
  name: string
  has_password: boolean
  has_configured_password: boolean
  last_login_date: string
  last_activity_date: string
}

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  token: string
  user_info: User
}

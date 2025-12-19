import axios from 'axios'
import { store } from '../store'
import { logout } from '../store/authSlice'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    const { code, message } = response.data
    // 后端返回 code: 0 或 code: 200 都表示成功
    if (code === 0 || code === 200) {
      return response  // 返回完整response，保持data结构
    }
    return Promise.reject(new Error(message || '请求失败'))
  },
  (error) => {
    if (error.response?.status === 401) {
      store.dispatch(logout())
      window.location.href = '/login'
    }
    return Promise.reject(error.response?.data?.message || error.message)
  }
)

// 用户相关API
export const userApi = {
  login: (data: { account: string; password: string }) =>
    api.post('/user/login', data),
  
  sendRegisterCode: (email: string) =>
    api.post('/user/send-register-code', { email }),
  
  register: (data: { username: string; email: string; code: string; password: string; nickname?: string; invite_code?: string }) =>
    api.post('/user/register', data),
  
  logout: () =>
    api.post('/user/logout'),
  
  getInfo: () =>
    api.get('/user/info'),
  
  updateInfo: (data: { nickname?: string; avatar?: string }) =>
    api.put('/user/update', data),
  
  changePassword: (data: { old_password: string; new_password: string }) =>
    api.put('/user/password', data),
  
  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/user/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  
  refreshToken: (refreshToken: string) =>
    api.post('/user/refresh-token', { refresh_token: refreshToken }),
  
  forgotPassword: (email: string) =>
    api.post('/user/forgot-password', { email }),
  
  resetPassword: (data: { email: string; code: string; password: string }) =>
    api.post('/user/reset-password', data),
  
  // 邮箱修改
  sendChangeEmailCode: (newEmail: string) =>
    api.post('/user/send-change-email-code', { new_email: newEmail }),
  
  changeEmail: (data: { new_email: string; code: string }) =>
    api.put('/user/email', data),
  
  // 设备管理
  getDevices: () =>
    api.get('/user/devices'),
  
  removeDevice: (deviceId: string) =>
    api.delete(`/user/devices/${deviceId}`),
  
  removeAllDevices: () =>
    api.delete('/user/devices'),
}

// 卡密相关API
export const cardApi = {
  redeem: (code: string) =>
    api.post('/card/redeem', { code }),
  
  getHistory: (params: { page: number; page_size: number }) =>
    api.get('/card/history', { params }),
}

// 公开续费API（无需登录）
export const publicApi = {
  renewByCard: (data: { account: string; code: string }) =>
    axios.post('/api/v1/card/renew', data),
  
  // 获取网站设置（公开接口）
  getSiteSettings: () =>
    axios.get('/api/v1/settings/site'),

  // 获取充值链接（公开接口）
  getRechargeLinks: () =>
    axios.get('/api/v1/settings/recharge-links'),

  // 获取积分卡购买链接（公开接口）
  getPointsRechargeLinks: () =>
    axios.get('/api/v1/settings/points-recharge-links'),
}

// 媒体库相关API
export const mediaApi = {
  // 获取媒体库列表
  getList: () =>
    api.get('/media/list'),
  
  // 获取媒体库中的媒体列表
  getMediaDBItems: (guid: string, params: { page?: number; page_size?: number }) =>
    api.get(`/media/db/${guid}/items`, { params }),
  
  // 获取媒体详情
  getMediaDetail: (guid: string) =>
    api.get(`/media/${guid}`),
  
  // 获取媒体的季列表（电视剧）
  getMediaSeasons: (guid: string) =>
    api.get(`/media/${guid}/seasons`),
  
  // 获取季的剧集列表
  getSeasonEpisodes: (seasonGuid: string) =>
    api.get(`/media/season/${seasonGuid}/episodes`),
  
  // 获取剧集的所有集数（不分季）
  getAllEpisodes: (seriesGuid: string) =>
    api.get(`/media/${seriesGuid}/episodes`),
  
  // 搜索媒体
  searchMedia: (params: { keyword: string; page?: number; page_size?: number }) =>
    api.get('/media/search', { params }),
  
  // 获取媒体库统计
  getMediaDBSum: (guid: string) =>
    api.get(`/media/db/${guid}/sum`),
}

// 会员相关API
export const memberApi = {
  getInfo: () =>
    api.get('/member/info'),
  
  renew: (data: { level: number; duration: number }) =>
    api.post('/member/renew', data),
  
  getOrders: (params: { page: number; page_size: number }) =>
    api.get('/member/orders', { params }),
}

// 积分相关API
export const pointsApi = {
  getMyPoints: () =>
    api.get('/points/my'),
  
  signIn: () =>
    api.post('/points/sign-in'),
  
  getSignInStatus: () =>
    api.get('/points/sign-in/status'),
  
  getRecords: (params: { page: number; page_size: number; type?: number }) =>
    api.get('/points/records', { params }),
  
  getExchangeRules: () =>
    api.get('/points/exchange-rules'),
  
  exchange: (ruleId: number) =>
    api.post('/points/exchange', { rule_id: ruleId }),
  
  // 积分卡密兑换
  redeemCard: (code: string) =>
    api.post('/points/redeem', { code }),
  
  // 积分排行榜（分页）
  getRanking: (params: { page: number; page_size: number }) =>
    api.get('/points/ranking', { params }),
  
  // 我的排名
  getMyRank: () =>
    api.get('/points/my-rank'),
}

// 积分卡管理API（管理员）
export const pointsCardApi = {
  // 批量生成积分卡
  createBatch: (data: { points: number; quantity: number; remark?: string }) =>
    api.post('/admin/points-card/batch', data),
  
  // 获取批次列表
  getBatchList: (params: { page: number; page_size: number }) =>
    api.get('/admin/points-card/batch/list', { params }),
  
  // 删除批次
  deleteBatch: (batchNo: string) =>
    api.delete(`/admin/points-card/batch/${batchNo}`),
  
  // 获取卡密列表
  getCardList: (params: { page: number; page_size: number; batch_no?: string; status?: number; keyword?: string }) =>
    api.get('/admin/points-card/list', { params }),
  
  // 禁用卡密
  disableCard: (id: number) =>
    api.post(`/admin/points-card/${id}/disable`),
  
  // 启用卡密
  enableCard: (id: number) =>
    api.post(`/admin/points-card/${id}/enable`),
  
  // 删除卡密
  deleteCard: (id: number) =>
    api.delete(`/admin/points-card/${id}`),
  
  // 获取统计
  getStats: () =>
    api.get('/admin/points-card/stats'),
  
  // 导出卡密
  exportCards: (batchNo: string) =>
    api.get('/admin/points-card/export', { params: { batch_no: batchNo } }),
}

// 管理员相关API
export const adminApi = {
  getUserList: (params: {
    page: number
    page_size: number
    keyword?: string       // 账号/邮箱/昵称搜索
    member_level?: number
    status?: number
  }) => api.get('/admin/user/list', { params }),
  
  getUserDetail: (id: string) =>
    api.get(`/admin/user/${id}`),
  
  updateUserStatus: (id: string, status: number) =>
    api.put(`/admin/user/${id}/status`, { status }),
  
  batchUpdateStatus: (userIds: string[], status: number) =>
    api.put('/admin/user/batch-status', { user_ids: userIds, status }),
  
  batchSetMember: (userIds: string[], days: number) =>
    api.put('/admin/user/batch-member', { user_ids: userIds, days }),
  
  resetPassword: (id: string, password: string) =>
    api.put(`/admin/user/${id}/reset-password`, { password }),
  
  updateUserRole: (id: string, role: number) =>
    api.put(`/admin/user/${id}/role`, { role }),
  
  setMember: (id: string, data: { level: number; days: number }) =>
    api.post(`/admin/user/${id}/set-member`, data),
  
  deleteUser: (id: string) =>
    api.delete(`/admin/user/${id}`),
  
  getLoginLogs: (id: string, params: { page: number; page_size: number }) =>
    api.get(`/admin/user/${id}/login-logs`, { params }),
  
  getOperationLogs: (params: { page: number; page_size: number }) =>
    api.get('/admin/operation-logs', { params }),
  
  getUserStats: () =>
    api.get('/admin/stat/user'),
  
  getDailyStats: (days?: number) =>
    api.get('/admin/stat/daily', { params: { days } }),
  
  getVisitRanking: (limit?: number) =>
    api.get('/admin/stat/ranking', { params: { limit } }),
  
  // 积分管理
  getPointsStats: () =>
    api.get('/admin/points/stats'),
  
  adjustPoints: (data: { user_id: string; points: number; remark?: string }) =>
    api.post('/admin/points/adjust', data),
  
  giftPointsToAll: (data: { points: number; remark?: string; send_email?: boolean; email_title?: string; email_body?: string }) =>
    api.post('/admin/points/gift-all', data),
  
  getPointsExchangeRules: () =>
    api.get('/admin/points/exchange-rules'),
  
  createPointsExchangeRule: (data: { name: string; points: number; member_days: number; description?: string; sort_order?: number }) =>
    api.post('/admin/points/exchange-rules', data),
  
  updatePointsExchangeRule: (id: number, data: { name?: string; points?: number; member_days?: number; description?: string; enabled?: boolean; sort_order?: number }) =>
    api.put(`/admin/points/exchange-rules/${id}`, data),
  
  deletePointsExchangeRule: (id: number) =>
    api.delete(`/admin/points/exchange-rules/${id}`),

  // 积分自动赠送规则
  getPointsGiftRules: () =>
    api.get('/admin/points/gift-rules'),

  createPointsGiftRule: (data: {
    name: string
    rule_type: number
    points: number
    target_type?: string
    member_level?: number
    execute_time?: string
    execute_day?: number
    execute_month?: number
    send_notification?: boolean
    notification_title?: string
    notification_body?: string
    enabled?: boolean
  }) =>
    api.post('/admin/points/gift-rules', data),

  updatePointsGiftRule: (id: number, data: {
    name?: string
    rule_type?: number
    points?: number
    target_type?: string
    member_level?: number
    execute_time?: string
    execute_day?: number
    execute_month?: number
    send_notification?: boolean
    notification_title?: string
    notification_body?: string
    enabled?: boolean
  }) =>
    api.put(`/admin/points/gift-rules/${id}`, data),

  deletePointsGiftRule: (id: number) =>
    api.delete(`/admin/points/gift-rules/${id}`),

  togglePointsGiftRule: (id: number) =>
    api.post(`/admin/points/gift-rules/${id}/toggle`),

  executePointsGiftRule: (id: number) =>
    api.post(`/admin/points/gift-rules/${id}/execute`),

  getPointsGiftLogs: (params: { rule_id?: number; page: number; page_size: number }) =>
    api.get('/admin/points/gift-logs', { params }),
  
  // 卡密管理
  createCardBatch: (data: { card_type: number; quantity: number; duration?: number; remark?: string }) =>
    api.post('/admin/card/batch', data),
  
  getCardBatchList: (params: { page: number; page_size: number }) =>
    api.get('/admin/card/batch/list', { params }),
  
  getCardList: (params: { page: number; page_size: number; batch_no?: string; card_type?: number; status?: number }) =>
    api.get('/admin/card/list', { params }),
  
  disableCard: (id: number) =>
    api.post(`/admin/card/${id}/disable`),
  
  enableCard: (id: number) =>
    api.post(`/admin/card/${id}/enable`),
  
  exportCards: (batchNo: string) =>
    api.get('/admin/card/export', { params: { batch_no: batchNo } }),
  
  getCardStats: () =>
    api.get('/admin/card/stats'),
  
  // 系统设置
  getEmailSettings: () =>
    api.get('/admin/settings/email'),
  
  saveEmailSettings: (data: object) =>
    api.put('/admin/settings/email', data),
  
  testEmailSettings: (to: string) =>
    api.post('/admin/settings/email/test', { to }),
  
  getDomainSettings: () =>
    api.get('/admin/settings/domain'),
  
  saveDomainSettings: (data: { enabled: boolean; domains: string[] }) =>
    api.put('/admin/settings/domain', data),
  
  // 注册设置
  getRegisterSettings: () =>
    api.get('/admin/settings/register'),
  
  saveRegisterSettings: (data: { enabled: boolean; gift_member_days: number; auto_disable_on_exp: boolean }) =>
    api.put('/admin/settings/register', data),
  
  // Emby/媒体服务设置
  getEmbySettings: () =>
    api.get('/admin/settings/emby'),
  
  saveEmbySettings: (data: { enabled: boolean; mode: string; base_url: string; api_key?: string; admin_user?: string; admin_pass?: string; template_user?: string }) =>
    api.put('/admin/settings/emby', data),
  
  testEmbyConnection: () =>
    api.post('/admin/settings/emby/test'),
  
  // 网站设置
  getSiteSettings: () =>
    api.get('/admin/settings/site'),
  
  saveSiteSettings: (data: { title: string; description?: string; keywords?: string; logo?: string; favicon?: string; footer?: string; icp?: string }) =>
    api.put('/admin/settings/site', data),
  
  uploadLogo: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/admin/settings/site/logo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // 充值链接设置
  getRechargeLinksSettings: () =>
    api.get('/admin/settings/recharge-links'),
  
  saveRechargeLinksSettings: (data: { links: Array<{ card_type: number; name: string; url: string; enabled: boolean }> }) =>
    api.put('/admin/settings/recharge-links', data),

  // 积分卡购买链接设置
  getPointsRechargeLinksSettings: () =>
    api.get('/admin/settings/points-recharge-links'),
  
  savePointsRechargeLinksSettings: (data: { links: Array<{ points: number; name: string; url: string; enabled: boolean }> }) =>
    api.put('/admin/settings/points-recharge-links', data),
  
  // 图床设置
  getImageHostSettings: () =>
    api.get('/admin/settings/image-host'),
  
  saveImageHostSettings: (data: { enabled: boolean; base_url: string }) =>
    api.put('/admin/settings/image-host', data),
  
  // Emby用户管理
  getEmbyUsers: () =>
    api.get('/admin/emby/users'),
  
  getEmbyUserByUsername: (username: string) =>
    api.get(`/admin/emby/user/${encodeURIComponent(username)}`),
  
  // Emby设备和会话管理
  getEmbySessions: () =>
    api.get('/admin/emby/sessions'),
  
  getEmbySessionsByUsername: (username: string) =>
    api.get(`/admin/emby/sessions/${encodeURIComponent(username)}`),
  
  killEmbySession: (sessionId: string) =>
    api.post(`/admin/emby/sessions/${encodeURIComponent(sessionId)}/kill`),
  
  getEmbyDevices: () =>
    api.get('/admin/emby/devices'),
  
  getEmbyDevicesByUsername: (username: string) =>
    api.get(`/admin/emby/user/${encodeURIComponent(username)}/devices`),
  
  deleteEmbyDevice: (deviceId: string) =>
    api.delete(`/admin/emby/devices/${encodeURIComponent(deviceId)}`),
  
  getEmbyUserStreamLimit: (username: string) =>
    api.get(`/admin/emby/user/${encodeURIComponent(username)}/stream-limit`),
  
  setEmbyUserStreamLimit: (username: string, limit: number) =>
    api.put(`/admin/emby/user/${encodeURIComponent(username)}/stream-limit`, { limit }),

  // 客户端白名单管理
  getClientWhitelistSettings: () =>
    api.get('/admin/settings/client-whitelist'),

  saveClientWhitelistSettings: (data: { enabled: boolean; clients: Array<{ name: string; display_name: string; enabled: boolean }> }) =>
    api.put('/admin/settings/client-whitelist', data),

  addClientToWhitelist: (name: string, displayName: string) =>
    api.post('/admin/settings/client-whitelist/add', { name, display_name: displayName }),

  removeClientFromWhitelist: (name: string) =>
    api.delete(`/admin/settings/client-whitelist/${encodeURIComponent(name)}`),

  updateClientStatus: (name: string, enabled: boolean) =>
    api.put(`/admin/settings/client-whitelist/${encodeURIComponent(name)}/status`, { enabled }),

  enforceClientWhitelist: () =>
    api.post('/admin/emby/enforce-client-whitelist'),

  // 会话限制设置
  getSessionLimitSettings: () =>
    api.get('/admin/settings/session-limit'),

  saveSessionLimitSettings: (data: { 
    enabled: boolean; 
    max_sessions: number; 
    auto_kill_oldest: boolean;
    play_limit_enabled: boolean;
    max_playing_sessions: number;
    auto_stop_oldest_play: boolean;
  }) =>
    api.put('/admin/settings/session-limit', data),

  enforceSessionLimit: () =>
    api.post('/admin/emby/enforce-session-limit'),

  enforcePlayLimit: () =>
    api.post('/admin/emby/enforce-play-limit'),

  getSessionLimitStatus: () =>
    api.get('/admin/emby/session-limit-status'),

  // 播放限制设置
  getPlayLimitSettings: () =>
    api.get('/admin/settings/play-limit'),

  savePlayLimitSettings: (data: { enabled: boolean; max_playing: number }) =>
    api.put('/admin/settings/play-limit', data),

  // 用户清理设置
  getUserCleanupSettings: () =>
    api.get('/admin/settings/user-cleanup'),

  saveUserCleanupSettings: (data: { enabled: boolean; inactive_days: number; expired_days: number; delete_emby_account: boolean }) =>
    api.put('/admin/settings/user-cleanup', data),

  // 用户设备策略管理（使用Emby的EnableAllDevices/EnabledDevices）
  getEmbyUserDevicePolicy: (username: string) =>
    api.get(`/admin/emby/user/${encodeURIComponent(username)}/device-policy`),

  setEmbyUserDevicePolicy: (username: string, enableAllDevices: boolean, enabledDevices: string[]) =>
    api.put(`/admin/emby/user/${encodeURIComponent(username)}/device-policy`, {
      enable_all_devices: enableAllDevices,
      enabled_devices: enabledDevices,
    }),

  // 按客户端名称设置用户设备策略
  setEmbyUserClientPolicy: (username: string, enableAllDevices: boolean, enabledClients: string[]) =>
    api.put(`/admin/emby/user/${encodeURIComponent(username)}/client-policy`, {
      enable_all_devices: enableAllDevices,
      enabled_clients: enabledClients,
    }),

  addDeviceToEmbyUserWhitelist: (username: string, deviceId: string) =>
    api.post(`/admin/emby/user/${encodeURIComponent(username)}/device-whitelist`, { device_id: deviceId }),

  removeDeviceFromEmbyUserWhitelist: (username: string, deviceId: string) =>
    api.delete(`/admin/emby/user/${encodeURIComponent(username)}/device-whitelist/${encodeURIComponent(deviceId)}`),

  applyGlobalDeviceWhitelistToUser: (username: string) =>
    api.post(`/admin/emby/user/${encodeURIComponent(username)}/apply-global-whitelist`),

  applyGlobalDeviceWhitelistToAllUsers: () =>
    api.post('/admin/emby/apply-global-whitelist-all'),
  
  // 用户同步
  syncUserToEmby: (id: string, password?: string) =>
    api.post(`/admin/sync/user/${id}/to-emby`, { password }),
  
  syncUserStatus: (id: string) =>
    api.post(`/admin/sync/user/${id}/status`),
  
  syncUserPassword: (id: string, password: string) =>
    api.post(`/admin/sync/user/${id}/password`, { password }),
  
  importEmbyUser: (username: string) =>
    api.post('/admin/sync/import-emby-user', { username }),
  
  syncAllUsers: () =>
    api.post('/admin/sync/all'),
  
  importAllEmbyUsers: () =>
    api.post('/admin/sync/import-all'),
  
  // 公告管理
  getAnnouncements: (params: { page: number; page_size: number; status?: number }) =>
    api.get('/admin/announcement/list', { params }),
  
  createAnnouncement: (data: { title: string; content: string; type: number; is_top: boolean }) =>
    api.post('/admin/announcement', data),
  
  updateAnnouncement: (id: number, data: { title: string; content: string; type: number; is_top: boolean }) =>
    api.put(`/admin/announcement/${id}`, data),
  
  deleteAnnouncement: (id: number) =>
    api.delete(`/admin/announcement/${id}`),
  
  publishAnnouncement: (id: number) =>
    api.post(`/admin/announcement/${id}/publish`),
  
  offlineAnnouncement: (id: number) =>
    api.post(`/admin/announcement/${id}/offline`),
  
  // IP黑名单
  getIPBlacklist: (params: { page: number; page_size: number }) =>
    api.get('/admin/ip-blacklist', { params }),
  
  addIPBlacklist: (data: { ip: string; reason?: string; duration?: number }) =>
    api.post('/admin/ip-blacklist', data),
  
  removeIPBlacklist: (ip: string) =>
    api.delete(`/admin/ip-blacklist/${ip}`),
  
  // 系统监控
  getSystemHealth: () =>
    api.get('/admin/system/health'),
  
  getSystemStats: () =>
    api.get('/admin/system/stats'),
  
  // 批量导入
  importUsers: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/admin/import/users', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  
  getImportTemplate: () =>
    api.get('/admin/import/template', { responseType: 'blob' }),
  
  // 数据备份
  createBackup: () =>
    api.post('/admin/backup'),
  
  getBackupList: () =>
    api.get('/admin/backup/list'),
  
  downloadBackup: (filename: string) =>
    api.get(`/admin/backup/download/${filename}`, { responseType: 'blob' }),
  
  deleteBackup: (filename: string) =>
    api.delete(`/admin/backup/${filename}`),
  
  restoreBackup: (filename: string) =>
    api.post(`/admin/backup/restore/${filename}`),
  
  // 邀请管理（管理后台）
  getInviteStats: () =>
    api.get('/admin/invite/stats'),
  
  getInviteRecords: (params: { page: number; page_size: number }) =>
    api.get('/admin/invite/records', { params }),
  
  setInviteRewardDays: (days: number) =>
    api.post('/admin/invite/reward', { days }),
}

// 用户邀请API
// 公告API（用户端）
export const announcementApi = {
  // 获取已发布公告列表
  getPublished: () =>
    api.get('/announcements'),
  
  // 获取公告详情
  getById: (id: number) =>
    api.get(`/announcement/${id}`),
}

export const inviteApi = {
  // 获取我的邀请信息
  getMyInviteInfo: () =>
    api.get('/user/invite/info'),
  
  // 获取我的邀请记录
  getMyInviteRecords: () =>
    api.get('/user/invite/records'),
  
  // 获取邀请排行榜
  getInviteRanking: (limit?: number) =>
    api.get('/user/invite/ranking', { params: { limit } }),
}

// 论坛API
export const forumApi = {
  // 节点
  getNodes: () =>
    api.get('/forum/nodes'),
  
  // 话题
  createTopic: (data: { node_id: number; title: string; content: string; content_type?: string; images?: string[] }) =>
    api.post('/forum/topic', data),
  
  getTopicList: (params: { node_id?: number; page?: number; page_size?: number; order_by?: string }) =>
    api.get('/forum/topics', { params }),
  
  getTopicDetail: (id: number) =>
    api.get(`/forum/topic/${id}`),
  
  updateTopic: (id: number, data: { title: string; content: string; images?: string[] }) =>
    api.put(`/forum/topic/${id}`, data),
  
  deleteTopic: (id: number) =>
    api.delete(`/forum/topic/${id}`),
  
  likeTopic: (id: number) =>
    api.post(`/forum/topic/${id}/like`),
  
  favoriteTopic: (id: number) =>
    api.post(`/forum/topic/${id}/favorite`),
  
  // 评论
  createComment: (data: { topic_id: number; content: string; images?: string[]; parent_id?: number; reply_to_user?: string }) =>
    api.post('/forum/comment', data),
  
  getCommentList: (params: { topic_id: number; page?: number; page_size?: number }) =>
    api.get('/forum/comments', { params }),
  
  getCommentReplies: (id: number, params?: { page?: number; page_size?: number }) =>
    api.get(`/forum/comment/${id}/replies`, { params }),
  
  deleteComment: (id: number) =>
    api.delete(`/forum/comment/${id}`),
  
  likeComment: (id: number) =>
    api.post(`/forum/comment/${id}/like`),
  
  // 我的
  getMyTopics: (params?: { page?: number; page_size?: number }) =>
    api.get('/forum/my/topics', { params }),
  
  getMyFavorites: (params?: { page?: number; page_size?: number }) =>
    api.get('/forum/my/favorites', { params }),
}

// 私信API
export const pmApi = {
  // 发送私信
  sendMessage: (data: { to_user_id: string; content: string; images?: string[] }) =>
    api.post('/pm/send', data),
  
  // 获取会话列表
  getConversations: (params?: { page?: number; page_size?: number }) =>
    api.get('/pm/conversations', { params }),
  
  // 获取与某用户的消息
  getMessages: (userId: string, params?: { page?: number; page_size?: number }) =>
    api.get(`/pm/messages/${userId}`, { params }),
  
  // 标记已读
  markAsRead: (userId: string) =>
    api.post(`/pm/read/${userId}`),
  
  // 删除消息
  deleteMessage: (id: number) =>
    api.delete(`/pm/message/${id}`),
  
  // 撤回消息
  recallMessage: (id: number) =>
    api.post(`/pm/message/${id}/recall`),
  
  // 获取未读数
  getUnreadCount: () =>
    api.get('/pm/unread-count'),
  
  // 搜索用户（用于发起私信）
  searchUsers: (keyword: string) =>
    api.get('/pm/search-users', { params: { keyword } }),
  
  // 检查是否可以发送私信
  canSendMessage: (userId: string) =>
    api.get(`/pm/can-send/${userId}`),
  
  // 静音/取消静音会话
  muteConversation: (userId: string) =>
    api.post(`/pm/mute/${userId}`),
}

// 黑名单API
export const blacklistApi = {
  // 拉黑用户
  blockUser: (userId: string, reason?: string) =>
    api.post(`/blacklist/${userId}`, { reason }),
  
  // 取消拉黑
  unblockUser: (userId: string) =>
    api.delete(`/blacklist/${userId}`),
  
  // 获取黑名单列表
  getBlacklist: (params?: { page?: number; page_size?: number }) =>
    api.get('/blacklist/list', { params }),
  
  // 检查是否被拉黑
  isBlocked: (userId: string) =>
    api.get(`/blacklist/check/${userId}`),
}

// 用户关注API
export const followApi = {
  // 关注/取消关注
  toggleFollow: (userId: string) =>
    api.post(`/follow/${userId}`),
  
  // 获取关注列表
  getFollowings: (userId?: string) =>
    api.get(userId ? `/follow/followings/${userId}` : '/follow/followings'),
  
  // 获取粉丝列表
  getFollowers: (userId?: string) =>
    api.get(userId ? `/follow/followers/${userId}` : '/follow/followers'),
  
  // 获取关注统计
  getFollowStats: (userId?: string) =>
    api.get(userId ? `/follow/stats/${userId}` : '/follow/stats'),
}

// 图床API（Telegram图床）
export const imageHostApi = {
  // 缓存的图床地址
  _baseUrl: '',
  
  // 获取图床地址（从后端设置或缓存）
  getBaseUrl: async (): Promise<string> => {
    if (imageHostApi._baseUrl) {
      return imageHostApi._baseUrl
    }
    try {
      const res = await axios.get('/api/v1/settings/image-host')
      if (res.data?.code === 0 && res.data?.data?.base_url) {
        imageHostApi._baseUrl = res.data.data.base_url
        return imageHostApi._baseUrl
      }
    } catch {
      // 获取失败使用默认值
    }
    imageHostApi._baseUrl = 'https://img.liubei.org'
    return imageHostApi._baseUrl
  },
  
  // 上传图片到图床
  upload: async (file: File, tags?: string, folder?: string): Promise<{ success: boolean; data?: { id: string; url: string; filename: string }; error?: string }> => {
    const baseUrl = await imageHostApi.getBaseUrl()
    const formData = new FormData()
    formData.append('file', file)
    if (tags) formData.append('tags', tags)
    if (folder) formData.append('folder', folder || 'forum')
    
    try {
      const response = await fetch(`${baseUrl}/api/upload`, {
        method: 'POST',
        body: formData,
      })
      return await response.json()
    } catch (error) {
      return { success: false, error: (error as Error).message }
    }
  },
  
  // 批量上传图片
  uploadMultiple: async (files: File[], folder?: string): Promise<string[]> => {
    const urls: string[] = []
    for (const file of files) {
      const result = await imageHostApi.upload(file, '', folder || 'forum')
      if (result.success && result.data?.url) {
        urls.push(result.data.url)
      }
    }
    return urls
  },
  
  // 清除缓存（设置更新后调用）
  clearCache: () => {
    imageHostApi._baseUrl = ''
  },
}

// 论坛管理API（管理员）
export const forumAdminApi = {
  // 节点管理
  getNodes: () =>
    api.get('/admin/forum/nodes'),
  
  createNode: (data: { name: string; description?: string; icon?: string; sort_order?: number }) =>
    api.post('/admin/forum/node', data),
  
  updateNode: (id: number, data: { name?: string; description?: string; icon?: string; sort_order?: number; status?: number }) =>
    api.put(`/admin/forum/node/${id}`, data),
  
  deleteNode: (id: number) =>
    api.delete(`/admin/forum/node/${id}`),
  
  // 话题管理
  getTopicList: (params: { node_id?: number; status?: number; keyword?: string; page?: number; page_size?: number }) =>
    api.get('/admin/forum/topics', { params }),
  
  deleteTopic: (id: number) =>
    api.delete(`/admin/forum/topic/${id}`),
  
  setTopicTop: (id: number, isTop: boolean) =>
    api.post(`/admin/forum/topic/${id}/top`, { is_top: isTop }),
  
  setTopicRecommend: (id: number, isRecommend: boolean) =>
    api.post(`/admin/forum/topic/${id}/recommend`, { is_recommend: isRecommend }),
  
  // 评论管理
  deleteComment: (id: number) =>
    api.delete(`/admin/forum/comment/${id}`),
}

export default api

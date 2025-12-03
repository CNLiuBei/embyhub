// Emby同步API
import { get, post } from '@/utils/request'
import type { EmbyUser } from '@/types'

// 测试Emby连接
export const testEmbyConnection = () => {
  return post('/emby/test')
}

// 别名导出
export const testConnection = testEmbyConnection

// 同步Emby用户
export const syncEmbyUsers = () => {
  return post<{ sync_count: number }>('/emby/sync')
}

// 别名导出
export const syncUsers = syncEmbyUsers

// 获取Emby用户列表
export const getEmbyUsers = () => {
  return get<EmbyUser[]>('/emby/users')
}

// ========== 媒体库相关API ==========

// 媒体库类型
export interface MediaLibrary {
  Id?: string           // Views API 返回 Id
  Name: string
  CollectionType: string
  ItemId?: string       // VirtualFolders API 返回 ItemId
  Locations?: string[]
  PrimaryImageTag?: string
}

// 媒体项目类型
export interface MediaItem {
  Id: string
  Name: string
  Type: string
  Overview?: string
  ProductionYear?: number
  CommunityRating?: number
  OfficialRating?: string
  RunTimeTicks?: number
  PremiereDate?: string
  DateCreated?: string
  Genres?: string[]
  ImageTags?: {
    Primary?: string
    Thumb?: string
  }
  BackdropImageTags?: string[]
  ChildCount?: number
  RecursiveItemCount?: number
  SeriesName?: string
  SeasonName?: string
  IndexNumber?: number
  ParentIndexNumber?: number
}

// 获取Emby服务器URL
export const getServerUrl = () => {
  return get<{ server_url: string }>('/media/server-url')
}

// 获取媒体库列表
export const getLibraries = () => {
  return get<MediaLibrary[]>('/media/libraries')
}

// 获取媒体项目列表
export const getItems = (params: {
  parent_id?: string
  type?: string
  page?: number
  page_size?: number
  sort_by?: string
  sort_order?: string
  search?: string  // 搜索关键词
}) => {
  return get<{ list: MediaItem[], total: number }>('/media/items', params)
}

// 获取单个媒体详情
export const getItem = (id: string) => {
  return get<MediaItem>(`/media/items/${id}`)
}

// 获取最新媒体
export const getLatestItems = (params?: { parent_id?: string, limit?: number }) => {
  return get<MediaItem[]>('/media/latest', params)
}

// 获取媒体图片URL（需要拼接Emby服务器地址）
export const getImageUrl = (serverUrl: string, itemId: string, imageType: string = 'Primary', tag?: string, maxWidth?: number) => {
  let url = `${serverUrl}/Items/${itemId}/Images/${imageType}`
  const params = []
  if (tag) params.push(`tag=${tag}`)
  if (maxWidth) params.push(`maxWidth=${maxWidth}`)
  if (params.length > 0) url += '?' + params.join('&')
  return url
}

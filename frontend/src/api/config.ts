// 系统配置API
import { get, put } from '@/utils/request'
import type { SystemConfig } from '@/types'

// 获取配置列表
export const getConfigs = () => {
  return get<SystemConfig[]>('/configs')
}

// 更新配置
export const updateConfig = (key: string, value: string) => {
  return put(`/configs/${key}`, { config_value: value })
}

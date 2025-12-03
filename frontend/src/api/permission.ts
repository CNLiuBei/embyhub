// 权限管理API
import { get } from '@/utils/request'
import type { Permission } from '@/types'

// 获取权限列表
export const getPermissionList = () => {
  return get<Permission[]>('/permissions')
}

// 别名导出，保持向后兼容
export const getPermissions = getPermissionList

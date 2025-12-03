// 角色管理API
import { get, post, put, del } from '@/utils/request'
import type { Role, RoleCreateRequest } from '@/types'

// 获取角色列表
export const getRoleList = () => {
  return get<Role[]>('/roles')
}

// 别名导出，保持向后兼容
export const getRoles = getRoleList

// 获取角色详情
export const getRoleDetail = (id: number) => {
  return get<Role>(`/roles/${id}`)
}

// 创建角色
export const createRole = (data: RoleCreateRequest) => {
  return post<Role>('/roles', data)
}

// 更新角色
export const updateRole = (id: number, data: Partial<RoleCreateRequest>) => {
  return put<Role>(`/roles/${id}`, data)
}

// 删除角色
export const deleteRole = (id: number) => {
  return del(`/roles/${id}`)
}

// 为角色分配权限
export const assignPermissions = (id: number, permissionIds: number[]) => {
  return post(`/roles/${id}/permissions`, { permission_ids: permissionIds })
}

// 用户管理API
import { get, post, put, del } from '@/utils/request'
import type { User, UserCreateRequest, UserUpdateRequest, PaginationResponse } from '@/types'

// 获取用户列表
export const getUserList = (params: any) => {
  return get<PaginationResponse<User>>('/users', params)
}

// 别名导出，保持向后兼容
export const getUsers = getUserList

// 获取用户详情
export const getUserDetail = (id: number) => {
  return get<User>(`/users/${id}`)
}

// 创建用户
export const createUser = (data: UserCreateRequest) => {
  return post<User>('/users', data)
}

// 更新用户
export const updateUser = (id: number, data: UserUpdateRequest) => {
  return put<User>(`/users/${id}`, data)
}

// 删除用户
export const deleteUser = (id: number) => {
  return del(`/users/${id}`)
}

// 重置密码
export const resetPassword = (id: number, password: string) => {
  return put(`/users/${id}/password`, { password })
}

// 批量更新用户状态
export const batchUpdateStatus = (userIds: number[], status: number) => {
  return put('/users/batch/status', { user_ids: userIds, status })
}

// 设置用户VIP
export const setUserVip = (userId: number, days: number) => {
  return put(`/users/${userId}/vip`, { days })
}

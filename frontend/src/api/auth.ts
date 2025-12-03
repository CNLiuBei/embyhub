// 认证相关API
import { post, get } from '@/utils/request'
import type { LoginRequest, LoginResponse, User } from '@/types'

// 登录
export const login = (data: LoginRequest) => {
  return post<LoginResponse>('/auth/login', data)
}

// 发送邮箱验证码
export const sendEmailCode = (data: { email: string; type?: string }) => {
  return post<{ message: string }>('/email/send-code', data)
}

// 注册（邮箱验证方式）
export const register = (data: { email: string; code: string; username: string; password: string }) => {
  return post<{
    user_id: number;
    username: string;
    email: string;
    emby_user_id?: string;
    message: string;
  }>('/auth/register', data)
}

// 登出
export const logout = () => {
  return post('/auth/logout')
}

// 获取当前用户信息
export const getCurrentUser = () => {
  return get<User>('/auth/current')
}

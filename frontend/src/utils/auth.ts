import type { User } from '@/types'

const TOKEN_KEY = 'token'
const USER_INFO_KEY = 'userInfo'

// 获取Token
export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY)
}

// 设置Token
export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token)
}

// 移除Token
export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY)
}

// 获取用户信息
export const getUserInfo = (): User | null => {
  const userInfoStr = localStorage.getItem(USER_INFO_KEY)
  if (userInfoStr) {
    try {
      return JSON.parse(userInfoStr)
    } catch (error) {
      console.error('解析用户信息失败:', error)
      return null
    }
  }
  return null
}

// 设置用户信息
export const setUserInfo = (userInfo: User): void => {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo))
}

// 移除用户信息
export const removeUserInfo = (): void => {
  localStorage.removeItem(USER_INFO_KEY)
}

// 清除所有认证信息
export const clearAuth = (): void => {
  removeToken()
  removeUserInfo()
}

// 检查是否已登录
export const isAuthenticated = (): boolean => {
  return !!getToken()
}

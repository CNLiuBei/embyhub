import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import { message } from 'antd'
import type { ApiResponse } from '@/types'

// 创建axios实例
const request = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Token刷新状态
let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

// 订阅Token刷新
const subscribeTokenRefresh = (callback: (token: string) => void) => {
  refreshSubscribers.push(callback)
}

// 通知所有订阅者Token已刷新
const onTokenRefreshed = (token: string) => {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

// 刷新Token
const refreshToken = async (): Promise<string | null> => {
  try {
    const response = await axios.post('/api/auth/refresh', null, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
    })
    if (response.data.code === 200 && response.data.data?.token) {
      const newToken = response.data.data.token
      localStorage.setItem('token', newToken)
      return newToken
    }
    return null
  } catch {
    return null
  }
}

// 检查Token是否需要刷新（剩余时间少于30分钟）
const shouldRefreshToken = (): boolean => {
  const token = localStorage.getItem('token')
  if (!token) return false
  
  try {
    // 解析JWT获取过期时间（简单解析，不验证签名）
    const payload = JSON.parse(atob(token.split('.')[1]))
    const expiresAt = payload.exp * 1000 // 转换为毫秒
    const now = Date.now()
    const thirtyMinutes = 30 * 60 * 1000
    
    // 如果剩余时间少于30分钟，需要刷新
    return expiresAt - now < thirtyMinutes && expiresAt > now
  } catch {
    return false
  }
}

// 请求拦截器
request.interceptors.request.use(
  async (config) => {
    // 从localStorage获取token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
      
      // 检查是否需要刷新Token（排除刷新请求本身）
      if (shouldRefreshToken() && !config.url?.includes('/auth/refresh')) {
        if (!isRefreshing) {
          isRefreshing = true
          const newToken = await refreshToken()
          isRefreshing = false
          
          if (newToken) {
            onTokenRefreshed(newToken)
            config.headers.Authorization = `Bearer ${newToken}`
          }
        } else {
          // 等待Token刷新完成
          return new Promise((resolve) => {
            subscribeTokenRefresh((newToken) => {
              config.headers.Authorization = `Bearer ${newToken}`
              resolve(config)
            })
          })
        }
      }
    }
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>): any => {
    const res = response.data

    // 如果code不是200，说明业务逻辑出错
    if (res.code !== 200) {
      message.error(res.message || '请求失败')
      
      // 401 未授权，跳转到登录页
      if (res.code === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
        window.location.href = '/login'
      }
      
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    
    return res
  },
  (error: AxiosError) => {
    console.error('响应错误:', error)
    
    if (error.response) {
      const status = error.response.status
      
      switch (status) {
        case 401:
          message.error('未授权，请重新登录')
          localStorage.removeItem('token')
          localStorage.removeItem('userInfo')
          window.location.href = '/login'
          break
        case 403:
          message.error('权限不足')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error('服务器错误')
          break
        default:
          message.error(error.message || '网络错误')
      }
    } else if (error.request) {
      message.error('网络连接失败，请检查网络')
    } else {
      message.error('请求失败')
    }
    
    return Promise.reject(error)
  }
)

// 封装GET请求
export const get = <T = any>(url: string, params?: any, config?: AxiosRequestConfig) => {
  return request.get<any, ApiResponse<T>>(url, { params, ...config })
}

// 封装POST请求
export const post = <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => {
  return request.post<any, ApiResponse<T>>(url, data, config)
}

// 封装PUT请求
export const put = <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => {
  return request.put<any, ApiResponse<T>>(url, data, config)
}

// 封装DELETE请求
export const del = <T = any>(url: string, config?: AxiosRequestConfig) => {
  return request.delete<any, ApiResponse<T>>(url, config)
}

export default request

import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import type { User } from '@/types'
import { setToken, setUserInfo, clearAuth } from '@/utils/auth'

interface AuthState {
  token: string | null
  userInfo: User | null
  isAuthenticated: boolean
}

const initialState: AuthState = {
  token: localStorage.getItem('token'),
  userInfo: localStorage.getItem('userInfo') 
    ? JSON.parse(localStorage.getItem('userInfo')!) 
    : null,
  isAuthenticated: !!localStorage.getItem('token'),
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    // 设置认证信息
    setAuthInfo: (state, action: PayloadAction<{ token: string; userInfo: User }>) => {
      const { token, userInfo } = action.payload
      state.token = token
      state.userInfo = userInfo
      state.isAuthenticated = true
      
      // 保存到localStorage
      setToken(token)
      setUserInfo(userInfo)
    },
    
    // 清除认证信息
    clearAuthInfo: (state) => {
      state.token = null
      state.userInfo = null
      state.isAuthenticated = false
      
      // 清除localStorage
      clearAuth()
    },
    
    // 更新用户信息
    updateUserInfo: (state, action: PayloadAction<User>) => {
      state.userInfo = action.payload
      setUserInfo(action.payload)
    },
  },
})

export const { setAuthInfo, clearAuthInfo, updateUserInfo } = authSlice.actions
export default authSlice.reducer

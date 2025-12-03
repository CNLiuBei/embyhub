import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import type { User, PaginationResponse } from '@/types'

interface UserState {
  userList: PaginationResponse<User> | null
  loading: boolean
  currentUser: User | null
}

const initialState: UserState = {
  userList: null,
  loading: false,
  currentUser: null,
}

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    // 设置用户列表
    setUserList: (state, action: PayloadAction<PaginationResponse<User>>) => {
      state.userList = action.payload
    },
    
    // 设置加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    },
    
    // 设置当前用户
    setCurrentUser: (state, action: PayloadAction<User | null>) => {
      state.currentUser = action.payload
    },
    
    // 清空用户数据
    clearUserData: (state) => {
      state.userList = null
      state.currentUser = null
    },
  },
})

export const { setUserList, setLoading, setCurrentUser, clearUserData } = userSlice.actions
export default userSlice.reducer

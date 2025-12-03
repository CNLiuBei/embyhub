import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { Spin } from 'antd'
import type { RootState } from './store'

// 页面懒加载（减少首屏加载时间）
const Login = lazy(() => import('./pages/Login'))
const Register = lazy(() => import('./pages/Register'))
const ForgotPassword = lazy(() => import('./pages/ForgotPassword'))
const Layout = lazy(() => import('./components/Layout'))
const Dashboard = lazy(() => import('./pages/Dashboard'))
const UserHome = lazy(() => import('./pages/UserHome'))
const UserList = lazy(() => import('./pages/User/UserList'))
const RoleList = lazy(() => import('./pages/Role/RoleList'))
const PermissionList = lazy(() => import('./pages/Permission/PermissionList'))
const AccessRecordList = lazy(() => import('./pages/AccessRecord/AccessRecordList'))
const SystemConfig = lazy(() => import('./pages/System/SystemConfig'))
const EmbySync = lazy(() => import('./pages/Emby/EmbySync'))
const CardKeyList = lazy(() => import('./pages/CardKey/CardKeyList'))
const MediaLibrary = lazy(() => import('./pages/Media/MediaLibrary'))
const Setup = lazy(() => import('./pages/Setup'))

// 加载中组件
const PageLoading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
    <Spin size="large">
      <div style={{ padding: 50, textAlign: 'center' }}>加载中...</div>
    </Spin>
  </div>
)

// 智能首页组件：根据角色显示不同页面
const SmartHome = () => {
  const userInfo = useSelector((state: RootState) => state.auth.userInfo)
  // 超级管理员和管理员看仪表盘，普通用户看个人中心
  if (userInfo?.role_id === 1 || userInfo?.role_id === 2) {
    return <Dashboard />
  }
  return <UserHome />
}

function App() {
  const isAuthenticated = useSelector((state: RootState) => state.auth.isAuthenticated)

  return (
    <BrowserRouter
      future={{
        v7_startTransition: true,
        v7_relativeSplatPath: true,
      }}
    >
      <Suspense fallback={<PageLoading />}>
        <Routes>
          <Route path="/setup" element={<Setup />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />
          <Route
            path="/"
            element={
              isAuthenticated ? <Layout /> : <Navigate to="/login" replace />
            }
          >
            <Route index element={<SmartHome />} />
            <Route path="users" element={<UserList />} />
            <Route path="roles" element={<RoleList />} />
            <Route path="permissions" element={<PermissionList />} />
            <Route path="access-records" element={<AccessRecordList />} />
            <Route path="system" element={<SystemConfig />} />
            <Route path="emby" element={<EmbySync />} />
            <Route path="card-keys" element={<CardKeyList />} />
            <Route path="media" element={<MediaLibrary />} />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

export default App

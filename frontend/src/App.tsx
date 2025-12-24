import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { App as AntdApp, Spin } from 'antd'
import { RootState } from './store'
import { SiteSettingsProvider } from './hooks/useSiteSettings.tsx'

// 布局组件（不懒加载）
import AdminLayout from './layouts/AdminLayout'
import UserLayout from './layouts/UserLayout'

// 页面组件（懒加载）
const Login = lazy(() => import('./pages/Login'))
const Register = lazy(() => import('./pages/Register'))
const ForgotPassword = lazy(() => import('./pages/ForgotPassword'))
const RenewCard = lazy(() => import('./pages/RenewCard'))
const About = lazy(() => import('./pages/About'))

// 用户中心页面
const UserCenter = lazy(() => import('./pages/user/UserCenter'))
const MemberCenter = lazy(() => import('./pages/user/MemberCenter'))
const VipPurchase = lazy(() => import('./pages/user/VipPurchase'))
const PurchaseMember = lazy(() => import('./pages/user/PurchaseMember'))
const MediaLibrary = lazy(() => import('./pages/user/MediaLibrary'))
const MediaDetail = lazy(() => import('./pages/user/MediaDetail'))
const WatchHistory = lazy(() => import('./pages/user/WatchHistory'))
const Favorites = lazy(() => import('./pages/user/Favorites'))
const Notifications = lazy(() => import('./pages/user/Notifications'))
const InviteFriends = lazy(() => import('./pages/user/InviteFriends'))
const Points = lazy(() => import('./pages/user/Points'))
const Forum = lazy(() => import('./pages/user/Forum'))
const ForumTopic = lazy(() => import('./pages/user/ForumTopic'))
const Messages = lazy(() => import('./pages/user/Messages'))

// 管理后台页面
const AdminDashboard = lazy(() => import('./pages/admin/Dashboard'))
const UserManagement = lazy(() => import('./pages/admin/UserManagement'))
const UserDetail = lazy(() => import('./pages/admin/UserDetail'))
const CardManagement = lazy(() => import('./pages/admin/CardManagement'))
const Settings = lazy(() => import('./pages/admin/Settings'))
const OperationLogs = lazy(() => import('./pages/admin/OperationLogs'))
const Announcements = lazy(() => import('./pages/admin/Announcements'))
const IPBlacklist = lazy(() => import('./pages/admin/IPBlacklist'))
const SystemMonitor = lazy(() => import('./pages/admin/SystemMonitor'))
const BackupManagement = lazy(() => import('./pages/admin/BackupManagement'))
const InviteManagement = lazy(() => import('./pages/admin/InviteManagement'))
const EmbyDeviceManagement = lazy(() => import('./pages/admin/EmbyDeviceManagement'))
const PointsManagement = lazy(() => import('./pages/admin/PointsManagement'))
const RechargeLinks = lazy(() => import('./pages/admin/RechargeLinks'))
const ForumManagement = lazy(() => import('./pages/admin/ForumManagement'))
const ExternalCardAPI = lazy(() => import('./pages/admin/ExternalCardAPI'))

// 闲管家对接页面
const GoofishSettings = lazy(() => import('./pages/admin/Goofish/GoofishSettings'))
const GoofishGoods = lazy(() => import('./pages/admin/Goofish/GoofishGoods'))
const GoofishOrders = lazy(() => import('./pages/admin/Goofish/GoofishOrders'))
const GoofishLogs = lazy(() => import('./pages/admin/Goofish/GoofishLogs'))

// 支付宝配置页面
const AlipaySettings = lazy(() => import('./pages/admin/AlipaySettings'))

// 加载指示器
const PageLoading = () => (
  <div className="flex items-center justify-center h-screen">
    <Spin size="large" />
  </div>
)

// 路由守卫
const PrivateRoute = ({ children, requiredRole = 0 }: { children: React.ReactNode; requiredRole?: number }) => {
  const { isAuthenticated, user } = useSelector((state: RootState) => state.auth)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  if (requiredRole > 0 && (user?.role || 0) < requiredRole) {
    return <Navigate to="/user" replace />
  }
  
  return <>{children}</>
}

function App() {
  return (
    <SiteSettingsProvider>
      <AntdApp>
        <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
          <Suspense fallback={<PageLoading />}>
          <Routes>
          {/* 公开路由 */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot" element={<ForgotPassword />} />
          <Route path="/renew" element={<RenewCard />} />
          <Route path="/about" element={<About />} />
          
          {/* 用户中心路由 */}
          <Route path="/user" element={
            <PrivateRoute>
              <UserLayout />
            </PrivateRoute>
          }>
            <Route index element={<UserCenter />} />
            <Route path="member" element={<MemberCenter />} />
            <Route path="vip-purchase" element={<VipPurchase />} />
            <Route path="purchase-member" element={<PurchaseMember />} />
            <Route path="media" element={<MediaLibrary />} />
            <Route path="media/:guid" element={<MediaDetail />} />
            <Route path="history" element={<WatchHistory />} />
            <Route path="favorites" element={<Favorites />} />
            <Route path="notifications" element={<Notifications />} />
            <Route path="invite" element={<InviteFriends />} />
            <Route path="points" element={<Points />} />
            <Route path="forum" element={<Forum />} />
            <Route path="forum/topic/:id" element={<ForumTopic />} />
            <Route path="messages" element={<Messages />} />
          </Route>
          
          {/* 管理后台路由 */}
          <Route path="/admin" element={
            <PrivateRoute requiredRole={2}>
              <AdminLayout />
            </PrivateRoute>
          }>
            <Route index element={<AdminDashboard />} />
            <Route path="users" element={<UserManagement />} />
            <Route path="users/:id" element={<UserDetail />} />
            <Route path="emby-devices" element={<EmbyDeviceManagement />} />
            <Route path="cards" element={<CardManagement />} />
            <Route path="announcements" element={<Announcements />} />
            <Route path="ip-blacklist" element={<IPBlacklist />} />
            <Route path="system" element={<SystemMonitor />} />
            <Route path="backup" element={<BackupManagement />} />
            <Route path="invite" element={<InviteManagement />} />
            <Route path="points" element={<PointsManagement />} />
            <Route path="recharge-links" element={<RechargeLinks />} />
            <Route path="forum" element={<ForumManagement />} />
            <Route path="external-api" element={<ExternalCardAPI />} />
            <Route path="goofish/settings" element={<GoofishSettings />} />
            <Route path="goofish/goods" element={<GoofishGoods />} />
            <Route path="goofish/orders" element={<GoofishOrders />} />
            <Route path="goofish/logs" element={<GoofishLogs />} />
            <Route path="alipay" element={<AlipaySettings />} />
            <Route path="logs" element={<OperationLogs />} />
            <Route path="settings" element={<Settings />} />
          </Route>
          
            {/* 默认重定向 */}
            <Route path="/" element={<Navigate to="/login" replace />} />
          </Routes>
          </Suspense>
        </BrowserRouter>
      </AntdApp>
    </SiteSettingsProvider>
  )
}

export default App

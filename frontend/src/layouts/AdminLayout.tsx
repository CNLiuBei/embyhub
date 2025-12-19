import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import { useQueryClient } from '@tanstack/react-query'
import { Layout, Menu, Dropdown } from 'antd'
import {
  DashboardOutlined,
  UserOutlined,
  CreditCardOutlined,
  LogoutOutlined,
  HomeOutlined,
  SkinOutlined,
  SettingOutlined,
  FileTextOutlined,
  NotificationOutlined,
  StopOutlined,
  MonitorOutlined,
  CloudOutlined,
  GiftOutlined,
  DesktopOutlined,
  TrophyOutlined,
  ShoppingCartOutlined,
  CommentOutlined,
} from '@ant-design/icons'
import UserAvatar from '../components/UserAvatar'
import { logout } from '../store/authSlice'
import { RootState } from '../store'
import { userApi } from '../services/api'
import { useTheme } from '../theme/ThemeContext'
import { useSiteSettings } from '../hooks/useSiteSettings'
import { ThemeName } from '../theme'
const { Content } = Layout

const AdminLayout = () => {
  const [collapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const dispatch = useDispatch()
  const queryClient = useQueryClient()
  const { user } = useSelector((state: RootState) => state.auth)
  const { theme, themeName, setTheme } = useTheme()
  const { settings: siteSettings } = useSiteSettings()
  
  // 从网站标题中提取品牌名称
  const brandName = siteSettings.title?.split(' - ')[0] || theme.brand.name
  const shortName = brandName.charAt(0)

  // 根据主题判断是否使用深色文字
  const isDarkBg = themeName === 'dark'
  const textColor = isDarkBg ? 'text-white' : 'text-gray-800'
  const textColorMuted = isDarkBg ? 'text-white/60' : 'text-gray-500'
  const borderColor = isDarkBg ? 'border-white/10' : 'border-gray-200/50'
  const hoverBg = isDarkBg ? 'hover:bg-white/10' : 'hover:bg-gray-100'
  const glassBg = isDarkBg ? 'bg-black/20' : 'bg-white/80'
  const glassBorder = isDarkBg ? 'border-white/20' : 'border-gray-200/50'

  const themeOptions: { key: ThemeName; label: string }[] = [
    { key: 'default', label: '🌤️ 默认主题' },
    { key: 'dark', label: '🌙 暗色主题' },
    { key: 'pink', label: '🌸 粉色主题' },
  ]

  const menuItems = [
    { key: '/admin', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/admin/users', icon: <UserOutlined />, label: '用户管理' },
    { key: '/admin/emby-devices', icon: <DesktopOutlined />, label: '客户端授权' },
    { key: '/admin/cards', icon: <CreditCardOutlined />, label: '会员卡管理' },
    { key: '/admin/recharge-links', icon: <ShoppingCartOutlined />, label: '充值链接' },
    { key: '/admin/points', icon: <TrophyOutlined />, label: '积分管理' },
    { key: '/admin/forum', icon: <CommentOutlined />, label: '论坛管理' },
    { key: '/admin/invite', icon: <GiftOutlined />, label: '邀请码' },
    { key: '/admin/announcements', icon: <NotificationOutlined />, label: '公告管理' },
    { key: '/admin/ip-blacklist', icon: <StopOutlined />, label: 'IP黑名单' },
    { key: '/admin/system', icon: <MonitorOutlined />, label: '系统监控' },
    { key: '/admin/backup', icon: <CloudOutlined />, label: '数据备份' },
    { key: '/admin/logs', icon: <FileTextOutlined />, label: '操作日志' },
    { key: '/admin/settings', icon: <SettingOutlined />, label: '系统设置' },
  ]

  const handleLogout = async () => {
    try {
      await userApi.logout()
    } catch {
      // ignore
    }
    // 清除所有React Query缓存
    queryClient.clear()
    dispatch(logout())
    navigate('/login')
  }

  const userMenuItems = [
    { key: 'user', label: '个人中心', icon: <HomeOutlined />, onClick: () => navigate('/user') },
    { 
      key: 'theme', 
      label: '主题设置', 
      icon: <SkinOutlined />,
      children: themeOptions.map(t => ({
        key: t.key,
        label: (
          <span className={themeName === t.key ? 'font-bold text-blue-500' : ''}>
            {t.label} {themeName === t.key && '✓'}
          </span>
        ),
        onClick: () => setTheme(t.key),
      })),
    },
    { type: 'divider' as const },
    { key: 'logout', label: '退出登录', icon: <LogoutOutlined />, onClick: handleLogout },
  ]

  return (
    <Layout style={{ minHeight: '100vh', background: 'var(--bg-base, #f5f7fa)' }}>
      {/* 卡片式侧边栏 */}
      <div 
        className={`fixed left-4 top-4 bottom-4 w-[220px] rounded-2xl shadow-lg overflow-hidden flex flex-col max-md:hidden ${glassBg} backdrop-blur-xl border ${glassBorder}`}
      >
        {/* Logo */}
        <div className={`h-16 flex items-center px-5 border-b ${borderColor}`}>
          <div className="flex items-center gap-3 min-w-0">
            {siteSettings.logo ? (
              <img src={siteSettings.logo} alt="Logo" className="w-9 h-9 rounded-full object-contain shadow-lg flex-shrink-0" />
            ) : (
              <div className="w-9 h-9 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center shadow-lg flex-shrink-0">
                <span className="text-white font-bold text-sm">{shortName}</span>
              </div>
            )}
            {!collapsed && <span className={`${textColor} font-bold text-sm truncate whitespace-nowrap`}>{brandName}</span>}
          </div>
        </div>
        
        {/* 菜单 */}
        <div className="flex-1 overflow-auto py-3 px-2">
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            className="border-none bg-transparent"
            style={{ 
              background: 'transparent',
            }}
          />
        </div>
        
        {/* 底部用户信息 */}
        <div className={`px-3 py-4 border-t ${borderColor}`}>
          <Dropdown menu={{ items: userMenuItems }} placement="topRight" trigger={['click']}>
            <div className={`flex items-center gap-3 p-2 rounded-xl cursor-pointer ${hoverBg} transition-all duration-200`}>
              <UserAvatar src={user?.avatar} name={user?.nickname || user?.username} size={40} />
              {!collapsed && (
                <div className="flex-1 min-w-0">
                  <div className={`${textColor} font-semibold text-sm truncate`}>{user?.nickname || 'admin'}</div>
                  <div className={`${textColorMuted} text-xs`}>管理员</div>
                </div>
              )}
            </div>
          </Dropdown>
        </div>
      </div>

      {/* 内容区域 - 固定卡片，内容滚动 */}
      <div 
        className={`fixed left-[252px] top-4 bottom-4 right-4 rounded-2xl shadow-lg overflow-hidden flex flex-col max-md:hidden ${glassBg} backdrop-blur-xl border ${glassBorder}`}
      >
        <Content className="flex-1 overflow-auto p-4">
          <Outlet />
        </Content>
      </div>
      
      {/* 移动端顶部导航 */}
      <div className={`hidden max-md:flex fixed top-0 left-0 right-0 h-14 ${glassBg} backdrop-blur-xl border-b ${borderColor} z-50 items-center px-4 justify-between`}>
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {siteSettings.logo ? (
            <img src={siteSettings.logo} alt="Logo" className="w-8 h-8 rounded-full object-contain flex-shrink-0" />
          ) : (
            <div className="w-8 h-8 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center flex-shrink-0">
              <span className="text-white font-bold text-xs">{shortName}</span>
            </div>
          )}
          <span className={`${textColor} font-bold text-sm truncate whitespace-nowrap`}>{brandName}</span>
        </div>
        <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
          <div className="flex items-center gap-2 cursor-pointer">
            <UserAvatar src={user?.avatar} name={user?.nickname || user?.username} size={32} />
          </div>
        </Dropdown>
      </div>

      {/* 移动端内容区域 */}
      <div className="hidden max-md:block pt-14 pb-16 min-h-screen">
        <Content className="p-3">
          <Outlet />
        </Content>
      </div>
        
      {/* 移动端底部导航 */}
      <div className={`hidden max-md:flex fixed bottom-0 left-0 right-0 h-14 ${glassBg} backdrop-blur-xl border-t ${borderColor} z-50 safe-area-bottom`}>
        {menuItems.slice(0, 4).map((item) => (
          <div
            key={item.key}
            onClick={() => navigate(item.key)}
            className={`flex-1 flex flex-col items-center justify-center cursor-pointer transition-colors ${
              location.pathname === item.key 
                ? 'text-blue-500' 
                : `${textColorMuted} hover:${textColor}`
            }`}
          >
            <span className="text-lg">{item.icon}</span>
            <span className="text-[10px] mt-0.5 font-medium">{String(item.label).slice(0, 4)}</span>
          </div>
        ))}
        <Dropdown 
          menu={{ 
            items: [
              ...menuItems.slice(4).map(item => ({
                key: item.key,
                label: item.label,
                icon: item.icon,
                onClick: () => navigate(item.key),
              })),
              { type: 'divider' as const },
              ...userMenuItems.filter(item => item.key !== 'theme'),
            ]
          }} 
          placement="top" 
          trigger={['click']}
        >
          <div className={`flex-1 flex flex-col items-center justify-center cursor-pointer ${textColorMuted} hover:${textColor} transition-colors`}>
            <SettingOutlined className="text-lg" />
            <span className="text-[10px] mt-0.5 font-medium">更多</span>
          </div>
        </Dropdown>
      </div>
    </Layout>
  )
}

export default AdminLayout

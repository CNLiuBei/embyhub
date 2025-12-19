import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useDispatch } from 'react-redux'
import { useQueryClient } from '@tanstack/react-query'
import { Form, Input, Button, Checkbox, App } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { setCredentials } from '../store/authSlice'
import { userApi } from '../services/api'
import { useTheme } from '../theme/ThemeContext'
import { useSiteSettings } from '../hooks/useSiteSettings.tsx'

const Login = () => {
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const dispatch = useDispatch()
  const queryClient = useQueryClient()
  const { theme } = useTheme()
  const { settings: siteSettings } = useSiteSettings()
  
  // 从网站标题中提取品牌名称（去掉副标题部分）
  const brandName = siteSettings.title?.split(' - ')[0] || theme.brand.name
  const shortName = brandName.charAt(0)

  const handleSubmit = async (values: { account: string; password: string }) => {
    try {
      setLoading(true)
      const response = await userApi.login(values)
      
      // 安全地提取数据
      const data = response?.data?.data || response?.data || response
      
      if (!data || !data.user) {
        throw new Error('登录响应数据格式错误')
      }
      
      // 清除之前用户的所有缓存
      queryClient.clear()
      
      dispatch(setCredentials({
        user: data.user,
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
      }))
      message.success('登录成功')
      
      // 安全地检查用户角色
      const userRole = data.user?.role ?? 0
      // role >= 2 是管理员，进入管理后台
      if (userRole >= 2) {
        navigate('/admin')
      } else {
        navigate('/user')
      }
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-bg flex items-center justify-center min-h-screen">
      {/* 左上角品牌 */}
      <div className="absolute top-6 left-6 flex items-center gap-3">
        {siteSettings.logo ? (
          <img src={siteSettings.logo} alt="Logo" className="w-10 h-10 rounded-full object-contain shadow-lg" />
        ) : (
          <div className="w-10 h-10 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center shadow-lg">
            <span className="text-white font-bold">{shortName}</span>
          </div>
        )}
        <span className="text-white text-lg font-bold drop-shadow-lg whitespace-nowrap">{brandName}</span>
      </div>
      
      {/* 登录卡片 */}
      <div className="glass-card p-10 w-[420px] shadow-2xl">
        {/* Logo */}
        <div className="flex justify-center mb-8">
          {siteSettings.logo ? (
            <img src={siteSettings.logo} alt="Logo" className="w-20 h-20 rounded-full object-contain shadow-xl" />
          ) : (
            <div className="w-20 h-20 rounded-full bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 flex items-center justify-center shadow-xl">
              <span className="text-white text-3xl font-bold">{shortName}</span>
            </div>
          )}
        </div>
        
        <h2 className="text-center text-2xl font-bold text-gray-800 mb-2">欢迎回来</h2>
        <p className="text-center text-gray-500 mb-8">登录您的 {brandName} 账户</p>
        
        <Form onFinish={handleSubmit} size="large">
          <Form.Item name="account" rules={[{ required: true, message: '请输入账号或邮箱' }]}>
            <Input 
              prefix={<UserOutlined className="text-gray-400" />} 
              placeholder="用户名或邮箱" 
              className="h-12 rounded-xl"
              autoComplete="username"
            />
          </Form.Item>
          
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password 
              prefix={<LockOutlined className="text-gray-400" />} 
              placeholder="密码" 
              className="h-12 rounded-xl"
              autoComplete="current-password"
            />
          </Form.Item>
          
          <div className="flex justify-between items-center mb-6">
            <Checkbox defaultChecked>记住我</Checkbox>
            <Link to="/forgot" className="text-blue-500 hover:text-blue-600">忘记密码？</Link>
          </div>
          
          <Form.Item className="mb-4">
            <Button 
              type="primary" 
              htmlType="submit" 
              loading={loading} 
              block
              className="h-12 rounded-xl text-base font-semibold shadow-lg hover:shadow-xl transition-shadow"
            >
              登 录
            </Button>
          </Form.Item>
        </Form>
        
        <div className="text-center text-gray-500 space-y-2">
          <div>还没有账户？ <Link to="/register" className="text-blue-500 font-medium hover:text-blue-600">立即注册</Link></div>
          <div>会员到期？ <Link to="/renew" className="text-orange-500 font-medium hover:text-orange-600">使用卡密续费</Link></div>
          <div className="pt-2"><Link to="/about" className="text-gray-400 hover:text-gray-600">关于我们</Link></div>
        </div>
      </div>
      
      {/* 底部版权 */}
      <div className="absolute bottom-6 text-white/60 text-sm">
        {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
      </div>
    </div>
  )
}

export default Login

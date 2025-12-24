import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Form, Input, Button, Steps, App } from 'antd'
import { MailOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons'
import { userApi } from '../services/api'
import { useTheme } from '../theme/ThemeContext'
import { useSiteSettings } from '../hooks/useSiteSettings.tsx'

const ForgotPassword = () => {
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [currentStep, setCurrentStep] = useState(0)
  const [email, setEmail] = useState('')
  const [countdown, setCountdown] = useState(0)
  const navigate = useNavigate()
  const { theme } = useTheme()
  const { settings: siteSettings } = useSiteSettings()
  const [form] = Form.useForm()
  
  // 从网站标题中提取品牌名称
  const brandName = siteSettings.title?.split(' - ')[0] || theme.brand.name
  const shortName = brandName.charAt(0)

  // 发送验证码
  const handleSendCode = async (values: { email: string }) => {
    try {
      setLoading(true)
      await userApi.forgotPassword(values.email)
      setEmail(values.email)
      message.success('验证码已发送到您的邮箱')
      setCurrentStep(1)
      // 开始倒计时
      setCountdown(60)
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  // 重新发送验证码
  const handleResend = async () => {
    if (countdown > 0) return
    try {
      setLoading(true)
      await userApi.forgotPassword(email)
      message.success('验证码已重新发送')
      setCountdown(60)
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  // 重置密码
  const handleResetPassword = async (values: { code: string; password: string }) => {
    try {
      setLoading(true)
      await userApi.resetPassword({
        email,
        code: values.code,
        password: values.password,
      })
      message.success('密码重置成功，请使用新密码登录')
      navigate('/login')
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : String(err))
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
      
      {/* 重置密码卡片 */}
      <div className="glass-card p-10 w-[450px] shadow-2xl">
        {/* Logo */}
        <div className="flex justify-center mb-6">
          {siteSettings.logo ? (
            <img src={siteSettings.logo} alt="Logo" className="w-16 h-16 rounded-full object-contain shadow-xl" />
          ) : (
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-orange-400 via-red-500 to-pink-500 flex items-center justify-center shadow-xl">
              <LockOutlined className="text-white text-2xl" />
            </div>
          )}
        </div>
        
        <h2 className="text-center text-2xl font-bold text-gray-800 mb-2">找回密码</h2>
        <p className="text-center text-gray-500 mb-6">通过邮箱验证码重置您的密码</p>
        
        {/* 步骤指示器 */}
        <Steps
          current={currentStep}
          size="small"
          className="mb-8"
          items={[
            { title: '验证邮箱' },
            { title: '重置密码' },
          ]}
        />
        
        {currentStep === 0 ? (
          // 第一步：输入邮箱
          <Form onFinish={handleSendCode} size="large">
            <Form.Item 
              name="email" 
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '邮箱格式不正确' },
              ]}
            >
              <Input 
                prefix={<MailOutlined className="text-gray-400" />} 
                placeholder="请输入注册时使用的邮箱" 
                className="h-12 rounded-xl"
              />
            </Form.Item>
            
            <Form.Item className="mb-4">
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={loading} 
                block
                className="h-12 rounded-xl text-base font-semibold shadow-lg hover:shadow-xl transition-shadow"
              >
                发送验证码
              </Button>
            </Form.Item>
          </Form>
        ) : (
          // 第二步：输入验证码和新密码
          <Form form={form} name="resetPasswordStepForm" onFinish={handleResetPassword} size="large">
            <div className="mb-4 p-3 bg-blue-50 rounded-xl text-sm text-blue-600">
              验证码已发送至 <span className="font-medium">{email}</span>
            </div>
            
            <Form.Item 
              name="code" 
              rules={[
                { required: true, message: '请输入验证码' },
                { len: 6, message: '验证码为6位数字' },
              ]}
            >
              <Input 
                prefix={<SafetyOutlined className="text-gray-400" />} 
                placeholder="请输入6位验证码" 
                className="h-12 rounded-xl"
                maxLength={6}
                suffix={
                  <Button 
                    type="link" 
                    size="small" 
                    disabled={countdown > 0}
                    onClick={handleResend}
                    loading={loading}
                  >
                    {countdown > 0 ? `${countdown}s后重发` : '重新发送'}
                  </Button>
                }
              />
            </Form.Item>
            
            <Form.Item 
              name="password" 
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码至少8位' },
                { 
                  pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/,
                  message: '密码必须包含大小写字母和数字'
                },
              ]}
            >
              <Input.Password 
                prefix={<LockOutlined className="text-gray-400" />} 
                placeholder="新密码 (至少8位，含大小写和数字)" 
                className="h-12 rounded-xl"
              />
            </Form.Item>
            
            <Form.Item
              name="confirmPassword"
              dependencies={['password']}
              rules={[
                { required: true, message: '请确认新密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve()
                    }
                    return Promise.reject(new Error('两次密码不一致'))
                  },
                }),
              ]}
            >
              <Input.Password 
                prefix={<LockOutlined className="text-gray-400" />} 
                placeholder="确认新密码" 
                className="h-12 rounded-xl"
              />
            </Form.Item>
            
            <Form.Item className="mb-4">
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={loading} 
                block
                className="h-12 rounded-xl text-base font-semibold shadow-lg hover:shadow-xl transition-shadow"
              >
                重置密码
              </Button>
            </Form.Item>
            
            <div className="text-center">
              <Button 
                type="link" 
                onClick={() => setCurrentStep(0)}
                className="text-gray-500"
              >
                ← 返回上一步
              </Button>
            </div>
          </Form>
        )}
        
        <div className="text-center text-gray-500 mt-4">
          想起密码了？ <Link to="/login" className="text-blue-500 font-medium hover:text-blue-600">立即登录</Link>
        </div>
      </div>
      
      {/* 底部版权 */}
      <div className="absolute bottom-6 text-white/60 text-sm">
        {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
      </div>
    </div>
  )
}

export default ForgotPassword

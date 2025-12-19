import { useState, useEffect } from 'react'
import { useNavigate, Link, useSearchParams } from 'react-router-dom'
import { Form, Input, Button, Select, App } from 'antd'
import { LockOutlined, UserOutlined, SmileOutlined, SafetyOutlined, MailOutlined, GiftOutlined } from '@ant-design/icons'
import { userApi } from '../services/api'
import { useTheme } from '../theme/ThemeContext'
import { useSiteSettings } from '../hooks/useSiteSettings.tsx'

// 允许的邮箱后缀
const emailDomains = [
  { value: 'qq.com', label: 'qq.com' },
  { value: '163.com', label: '163.com' },
  { value: '126.com', label: '126.com' },
  { value: 'gmail.com', label: 'gmail.com' },
  { value: 'outlook.com', label: 'outlook.com' },
  { value: 'hotmail.com', label: 'hotmail.com' },
  { value: 'icloud.com', label: 'icloud.com' },
  { value: 'foxmail.com', label: 'foxmail.com' },
  { value: 'sina.com', label: 'sina.com' },
  { value: 'aliyun.com', label: 'aliyun.com' },
  { value: '139.com', label: '139.com' },
  { value: '189.cn', label: '189.cn' },
]

const Register = () => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const { theme } = useTheme()
  const { settings: siteSettings } = useSiteSettings()
  
  // 从网站标题中提取品牌名称
  const brandName = siteSettings.title?.split(' - ')[0] || theme.brand.name
  const shortName = brandName.charAt(0)
  
  // 从 URL 参数获取邀请码
  const inviteCode = searchParams.get('invite') || ''

  // 初始化倒计时（从 localStorage 读取）
  useEffect(() => {
    const checkCountdown = () => {
      const endTime = localStorage.getItem('registerCodeEndTime')
      if (endTime) {
        const remaining = Math.floor((parseInt(endTime) - Date.now()) / 1000)
        if (remaining > 0) {
          setCountdown(remaining)
        } else {
          localStorage.removeItem('registerCodeEndTime')
          setCountdown(0)
        }
      }
    }
    
    checkCountdown()
    const timer = setInterval(checkCountdown, 1000)
    
    return () => clearInterval(timer)
  }, [])

  // 发送验证码
  const handleSendCode = async () => {
    try {
      const prefix = form.getFieldValue('emailPrefix')
      const domain = form.getFieldValue('emailDomain')
      if (!prefix) {
        message.warning('请先输入邮箱')
        return
      }
      if (!domain) {
        message.warning('请选择邮箱后缀')
        return
      }
      const email = `${prefix}@${domain}`
      setSendingCode(true)
      await userApi.sendRegisterCode(email)
      message.success('验证码已发送')
      // 开始倒计时并存储到 localStorage
      const endTime = Date.now() + 60000 // 60秒后
      localStorage.setItem('registerCodeEndTime', endTime.toString())
      setCountdown(60)
    } catch (err: unknown) {
      message.error(String(err) || '发送失败')
    } finally {
      setSendingCode(false)
    }
  }

  const handleSubmit = async (values: { username: string; emailPrefix: string; emailDomain: string; code: string; password: string; nickname?: string; invite_code?: string }) => {
    try {
      setLoading(true)
      const email = `${values.emailPrefix}@${values.emailDomain}`
      await userApi.register({
        username: values.username,
        email,
        code: values.code,
        password: values.password,
        nickname: values.nickname,
        invite_code: values.invite_code,
      })
      message.success('注册成功，请登录')
      navigate('/login')
    } catch (err: unknown) {
      message.error(String(err) || '注册失败')
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
      
      {/* 注册卡片 */}
      <div className="glass-card p-10 w-[420px] shadow-2xl">
        {/* Logo */}
        <div className="flex justify-center mb-6">
          {siteSettings.logo ? (
            <img src={siteSettings.logo} alt="Logo" className="w-16 h-16 rounded-full object-contain shadow-xl" />
          ) : (
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-green-400 via-blue-500 to-purple-500 flex items-center justify-center shadow-xl">
              <span className="text-white text-2xl font-bold">{shortName}</span>
            </div>
          )}
        </div>
        
        <h2 className="text-center text-2xl font-bold text-gray-800 mb-2">创建账户</h2>
        <p className="text-center text-gray-500 mb-6">加入 {brandName}，开启精彩之旅</p>
        
        <Form form={form} onFinish={handleSubmit} size="large" initialValues={{ emailPrefix: '', emailDomain: 'qq.com', invite_code: inviteCode }}>
          <Form.Item name="username" rules={[
            { required: true, message: '请输入账号' },
            { min: 4, message: '账号至少4个字符' },
            { max: 20, message: '账号最多20个字符' },
            { pattern: /^[a-zA-Z0-9_]+$/, message: '只能包含字母、数字和下划线' },
          ]}>
            <Input 
              prefix={<UserOutlined className="text-gray-400" />} 
              placeholder="账号 (4-20个字符)" 
              className="h-11 rounded-xl"
            />
          </Form.Item>
          
          <Form.Item label={null} className="mb-4">
            <div className="input-glow">
              <div className="flex items-center h-11 bg-white rounded-[10px] border-none">
                <MailOutlined className="text-gray-400 ml-3 mr-2 flex-shrink-0" />
              <Form.Item name="emailPrefix" noStyle rules={[
                { required: true, message: '请输入邮箱' },
                { pattern: /^[a-zA-Z0-9._-]+$/, message: '格式不正确' },
              ]}>
                <input 
                  placeholder="邮箱账号" 
                  className="h-10 flex-1 bg-transparent outline-none border-none min-w-0 text-sm"
                  style={{ boxShadow: 'none' }}
                  autoComplete="email"
                />
              </Form.Item>
              <Form.Item name="emailDomain" noStyle rules={[{ required: true }]}>
                <Select 
                  className="h-10 flex-shrink-0"
                  style={{ width: 130 }}
                  popupMatchSelectWidth={false}
                  optionLabelProp="label"
                  variant="borderless"
                >
                  {emailDomains.map(d => (
                    <Select.Option key={d.value} value={d.value} label={`@${d.label}`}>
                      @{d.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
              </div>
            </div>
          </Form.Item>

          <Form.Item name="code" rules={[
            { required: true, message: '请输入验证码' },
            { len: 6, message: '验证码为6位数字' },
          ]}>
            <Input 
              prefix={<SafetyOutlined className="text-gray-400" />} 
              placeholder="6位数字验证码" 
              className="h-11 rounded-xl"
              maxLength={6}
              autoComplete="one-time-code"
              suffix={
                <Button 
                  type={countdown > 0 ? 'text' : 'link'}
                  size="small"
                  disabled={countdown > 0 || sendingCode}
                  loading={sendingCode}
                  onClick={handleSendCode}
                  className={countdown > 0 ? 'text-gray-400' : 'text-blue-500 font-medium'}
                >
                  {countdown > 0 ? `${countdown}s 后重发` : '获取验证码'}
                </Button>
              }
            />
          </Form.Item>
          
          <Form.Item name="password" rules={[
            { required: true, message: '请输入密码' },
            { min: 8, message: '密码至少8位' },
            { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/, message: '需包含大小写字母和数字' },
          ]}>
            <Input.Password 
              prefix={<LockOutlined className="text-gray-400" />} 
              placeholder="密码 (大小写字母+数字，至少8位)" 
              className="h-11 rounded-xl"
              autoComplete="new-password"
            />
          </Form.Item>
          
          <Form.Item
            name="confirmPassword"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
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
              placeholder="确认密码" 
              className="h-11 rounded-xl"
              autoComplete="new-password"
            />
          </Form.Item>

          <Form.Item name="nickname">
            <Input 
              prefix={<SmileOutlined className="text-gray-400" />} 
              placeholder="昵称 (选填，默认使用账号)" 
              className="h-11 rounded-xl"
            />
          </Form.Item>

          <Form.Item name="invite_code">
            <Input 
              prefix={<GiftOutlined className="text-gray-400" />} 
              placeholder="邀请码 (选填)" 
              className="h-11 rounded-xl"
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
              注 册
            </Button>
          </Form.Item>
        </Form>
        
        <div className="text-center text-gray-500">
          已有账户？ <Link to="/login" className="text-blue-500 font-medium hover:text-blue-600">立即登录</Link>
        </div>
      </div>
      
      {/* 底部版权 */}
      <div className="absolute bottom-6 text-white/60 text-sm">
        {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
      </div>
    </div>
  )
}

export default Register

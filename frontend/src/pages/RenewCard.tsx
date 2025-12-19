import { useState } from 'react'
import { Form, Input, Button, App, Result } from 'antd'
import { UserOutlined, KeyOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { publicApi } from '../services/api'
import { useTheme } from '../theme/ThemeContext'
import { useSiteSettings } from '../hooks/useSiteSettings.tsx'

interface RenewResult {
  order_no: string
  expire_time: string
  duration: number
}

const RenewCard = () => {
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [result, setResult] = useState<RenewResult | null>(null)
  const [form] = Form.useForm()
  const { message } = App.useApp()
  const { theme } = useTheme()
  const { settings: siteSettings } = useSiteSettings()
  
  // 从网站标题中提取品牌名称
  const brandName = siteSettings.title?.split(' - ')[0] || theme.brand.name
  const shortName = brandName.charAt(0)

  const handleSubmit = async (values: { account: string; code: string }) => {
    try {
      setLoading(true)
      const response = await publicApi.renewByCard(values)
      if (response.data.code === 200) {
        setResult(response.data.data)
        setSuccess(true)
        message.success('续费成功！')
      } else {
        message.error(response.data.message || '续费失败')
      }
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : 
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message || '续费失败，请检查账号和卡密'
      message.error(errorMsg)
    } finally {
      setLoading(false)
    }
  }

  if (success && result) {
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

        <div className="glass-card p-10 w-[420px] shadow-2xl">
          <Result
            status="success"
            icon={<CheckCircleOutlined className="text-green-500" />}
            title={<span className="text-gray-800">续费成功！</span>}
            subTitle={
              <div className="space-y-2 text-left mt-4 text-gray-600">
                <p><strong>订单号：</strong>{result.order_no}</p>
                <p><strong>会员时长：</strong>{result.duration} 天</p>
                <p><strong>到期时间：</strong>{result.expire_time}</p>
              </div>
            }
            extra={[
              <Button type="primary" key="login" size="large" className="h-12 rounded-xl">
                <Link to="/login">立即登录</Link>
              </Button>,
              <Button key="back" size="large" className="h-12 rounded-xl" onClick={() => { setSuccess(false); setResult(null); form.resetFields() }}>
                继续续费
              </Button>,
            ]}
          />
        </div>

        {/* 底部版权 */}
        <div className="absolute bottom-6 text-white/60 text-sm">
          {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
        </div>
      </div>
    )
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

      {/* 续费卡片 */}
      <div className="glass-card p-10 w-[420px] shadow-2xl">
        {/* Logo */}
        <div className="flex justify-center mb-8">
          {siteSettings.logo ? (
            <img src={siteSettings.logo} alt="Logo" className="w-20 h-20 rounded-full object-contain shadow-xl" />
          ) : (
            <div className="w-20 h-20 rounded-full bg-gradient-to-br from-orange-500 via-red-500 to-pink-500 flex items-center justify-center shadow-xl">
              <KeyOutlined className="text-white text-3xl" />
            </div>
          )}
        </div>
        
        <h2 className="text-center text-2xl font-bold text-gray-800 mb-2">会员续费</h2>
        <p className="text-center text-gray-500 mb-8">使用卡密为您的账户续费会员</p>

        <Form
          form={form}
          onFinish={handleSubmit}
          size="large"
        >
          <Form.Item
            name="account"
            rules={[{ required: true, message: '请输入您的账号或邮箱' }]}
          >
            <Input 
              prefix={<UserOutlined className="text-gray-400" />} 
              placeholder="账号或邮箱"
              className="h-12 rounded-xl"
            />
          </Form.Item>

          <Form.Item
            name="code"
            rules={[{ required: true, message: '请输入卡密' }]}
          >
            <Input 
              prefix={<KeyOutlined className="text-gray-400" />} 
              placeholder="请输入卡密码"
              className="h-12 rounded-xl"
            />
          </Form.Item>

          <Form.Item className="mb-4">
            <Button 
              type="primary" 
              htmlType="submit" 
              block 
              loading={loading}
              className="h-12 rounded-xl text-base font-semibold shadow-lg hover:shadow-xl transition-shadow"
            >
              立即续费
            </Button>
          </Form.Item>
        </Form>

        <div className="text-center text-gray-500 space-y-2">
          <div className="text-sm">续费成功后账户将自动恢复，可正常登录使用</div>
          <div>
            <Link to="/login" className="text-blue-500 font-medium hover:text-blue-600">返回登录</Link>
            <span className="mx-2 text-gray-300">|</span>
            <Link to="/register" className="text-blue-500 font-medium hover:text-blue-600">注册账户</Link>
          </div>
        </div>
      </div>

      {/* 底部版权 */}
      <div className="absolute bottom-6 text-white/60 text-sm">
        {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
      </div>
    </div>
  )
}

export default RenewCard

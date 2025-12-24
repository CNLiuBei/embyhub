import { useState, useEffect } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Form, Input, Button, Upload, Modal, Radio, Row, Col, Progress, Tag, App } from 'antd'
import { UserOutlined, EditOutlined, CameraOutlined, MailOutlined, LockOutlined, SafetyOutlined, CrownOutlined, TrophyOutlined, GiftOutlined } from '@ant-design/icons'
import DeviceManagement from './DeviceManagement'
import UserAvatar from '../../components/UserAvatar'
import SignInFlame, { getFlameConfig } from '../../components/SignInFlame'
import { RootState } from '../../store'
import { updateUser } from '../../store/authSlice'
import { userApi, pointsApi } from '../../services/api'
import dayjs from 'dayjs'

interface SignInStatus {
  signed_today: boolean
  continue_days: number
}

const UserCenter = () => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { user } = useSelector((state: RootState) => state.auth)
  const dispatch = useDispatch()
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [passwordModalOpen, setPasswordModalOpen] = useState(false)
  const [emailModalOpen, setEmailModalOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [editForm] = Form.useForm()
  const [passwordForm] = Form.useForm()
  const [emailForm] = Form.useForm()

  // 获取积分
  const { data: pointsData } = useQuery({
    queryKey: ['myPoints'],
    queryFn: async () => {
      const res = await pointsApi.getMyPoints()
      return res.data.data as { points: number }
    },
  })

  // 获取签到状态
  const { data: signInStatus } = useQuery({
    queryKey: ['signInStatus'],
    queryFn: async () => {
      const res = await pointsApi.getSignInStatus()
      return res.data.data as SignInStatus
    },
  })

  // 签到
  const signInMutation = useMutation({
    mutationFn: () => pointsApi.signIn(),
    onSuccess: (res) => {
      const result = res.data.data
      message.success(`签到成功！获得 ${result.points} 积分`)
      queryClient.invalidateQueries({ queryKey: ['myPoints'] })
      queryClient.invalidateQueries({ queryKey: ['signInStatus'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '签到失败')
    },
  })

  // 倒计时效果
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 角色配置
  const roleConfig = [
    { name: '普通用户', color: '#999', bgColor: 'from-gray-400 to-gray-500', tagColor: 'default' },
    { name: '会员用户', color: '#13c2c2', bgColor: 'from-cyan-400 to-cyan-600', tagColor: 'cyan' },
    { name: '管理员', color: '#1890ff', bgColor: 'from-blue-400 to-blue-600', tagColor: 'blue' },
    { name: '超级管理员', color: '#722ed1', bgColor: 'from-purple-400 to-purple-600', tagColor: 'purple' },
  ]

  const handleEditSubmit = async (values: { nickname: string; gender: number }) => {
    try {
      setLoading(true)
      await userApi.updateInfo(values)
      dispatch(updateUser(values))
      message.success('更新成功')
      setEditModalOpen(false)
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '更新失败')
    } finally {
      setLoading(false)
    }
  }

  const handlePasswordSubmit = async (values: { old_password: string; new_password: string; confirm_password: string }) => {
    if (values.new_password !== values.confirm_password) {
      message.error('两次密码输入不一致')
      return
    }
    try {
      setLoading(true)
      await userApi.changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      })
      message.success('密码修改成功')
      setPasswordModalOpen(false)
      passwordForm.resetFields()
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '修改失败')
    } finally {
      setLoading(false)
    }
  }

  const openEditModal = () => {
    editForm.setFieldsValue({
      nickname: user?.nickname,
      gender: user?.gender || 0,
    })
    setEditModalOpen(true)
  }

  // 发送修改邮箱验证码
  const handleSendEmailCode = async () => {
    try {
      const newEmail = emailForm.getFieldValue('new_email')
      if (!newEmail) {
        message.error('请输入新邮箱')
        return
      }
      setSendingCode(true)
      await userApi.sendChangeEmailCode(newEmail)
      message.success('验证码已发送到新邮箱')
      setCountdown(60)
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '发送失败')
    } finally {
      setSendingCode(false)
    }
  }

  // 修改邮箱
  const handleEmailSubmit = async (values: { new_email: string; code: string }) => {
    try {
      setLoading(true)
      await userApi.changeEmail(values)
      dispatch(updateUser({ email: values.new_email }))
      message.success('邮箱修改成功')
      setEmailModalOpen(false)
      emailForm.resetFields()
      setCountdown(0)
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '修改失败')
    } finally {
      setLoading(false)
    }
  }

  const openEmailModal = () => {
    emailForm.resetFields()
    setCountdown(0)
    setEmailModalOpen(true)
  }

  // 头像上传处理
  const handleAvatarUpload = async (file: File) => {
    try {
      setLoading(true)
      const response = await userApi.uploadAvatar(file)
      const avatarUrl = response.data.data.avatar
      dispatch(updateUser({ avatar: avatarUrl }))
      message.success('头像上传成功')
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '头像上传失败')
    } finally {
      setLoading(false)
    }
    return false // 阻止默认上传行为
  }

  const userRole = user?.role || 0
  const currentRoleConfig = roleConfig[userRole] || roleConfig[0]
  const isAdmin = userRole >= 2 // 管理员或超级管理员

  // 计算会员剩余天数
  const getMemberDaysLeft = () => {
    if (isAdmin) return -1 // 管理员返回-1表示长期
    if (!user?.member_expire) return 0
    const expireDate = new Date(user.member_expire)
    const now = new Date()
    const diff = expireDate.getTime() - now.getTime()
    return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
  }

  const daysLeft = getMemberDaysLeft()
  
  // 格式化到期时间
  const formatExpireTime = () => {
    if (isAdmin) return '长期'
    if (!user?.member_expire) return '未开通'
    return dayjs(user.member_expire).format('YYYY-MM-DD')
  }

  return (
    <div className="space-y-6">
      <Row gutter={24}>
        {/* 左侧：用户信息卡片 */}
        <Col span={8}>
          <div className="glass-card p-6 text-center">
            {/* 头像 */}
            <div className="relative inline-block mb-4">
              <UserAvatar 
                src={user?.avatar} 
                name={user?.nickname || user?.username}
                size={120}
                className="shadow-lg"
              />
              <Upload 
                showUploadList={false} 
                accept="image/*"
                beforeUpload={handleAvatarUpload}
                disabled={loading}
              >
                <div className="absolute bottom-0 right-0 w-8 h-8 rounded-full bg-white shadow-md flex items-center justify-center cursor-pointer hover:bg-gray-50">
                  <CameraOutlined className="text-gray-600" />
                </div>
              </Upload>
            </div>
            
            {/* 昵称 */}
            <h2 className="text-xl font-bold text-gray-800 mb-1">{user?.nickname}</h2>
            
            {/* 角色标签 */}
            <div className="mb-4">
              <Tag color={currentRoleConfig.tagColor} className="text-sm px-3 py-1">
                {userRole >= 1 && <CrownOutlined className="mr-1" />}
                {currentRoleConfig.name}
              </Tag>
            </div>
            
            {/* 会员/管理员信息 */}
            {userRole >= 1 && (
              <div className="mt-4 px-2">
                {isAdmin ? (
                  <div className="text-center p-3 rounded-xl bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-100">
                    <div className="text-sm text-gray-500 mb-1">权限</div>
                    <div className="text-lg font-bold" style={{ color: currentRoleConfig.color }}>长期有效</div>
                  </div>
                ) : (
                  <div 
                    className="p-4 rounded-xl border-2"
                    style={{ 
                      background: daysLeft <= 7 
                        ? 'linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%)' 
                        : daysLeft <= 30 
                          ? 'linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%)'
                          : 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%)',
                      borderColor: daysLeft <= 7 ? '#fca5a5' : daysLeft <= 30 ? '#fcd34d' : '#6ee7b7'
                    }}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-gray-600 text-sm font-medium">会员剩余</span>
                      <span 
                        className="text-2xl font-bold"
                        style={{ color: daysLeft <= 7 ? '#dc2626' : daysLeft <= 30 ? '#d97706' : '#059669' }}
                      >
                        {daysLeft} <span className="text-sm font-normal">天</span>
                      </span>
                    </div>
                    <Progress 
                      percent={Math.min(100, (daysLeft / 365) * 100)} 
                      showInfo={false}
                      strokeColor={daysLeft <= 7 ? '#ef4444' : daysLeft <= 30 ? '#f59e0b' : '#10b981'}
                      trailColor="#e5e7eb"
                      size="small"
                    />
                    <div className="flex items-center justify-between mt-2">
                      <span className="text-xs text-gray-500">到期：{formatExpireTime()}</span>
                      {daysLeft <= 7 && (
                        <span className="text-xs text-red-500 font-medium animate-pulse">⚠️ 即将到期</span>
                      )}
                      {daysLeft > 7 && daysLeft <= 30 && (
                        <span className="text-xs text-amber-600 font-medium">⏰ 请及时续费</span>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
            
            {/* 积分签到卡片 */}
            <div 
              className="mt-4 relative overflow-hidden rounded-xl cursor-pointer group"
              style={{ background: 'linear-gradient(135deg, #f97316 0%, #eab308 100%)' }}
              onClick={() => navigate('/user/points')}
            >
              <div className="absolute inset-0 bg-white/10 opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="p-4 relative">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center">
                      <TrophyOutlined className="text-white text-lg" />
                    </div>
                    <div className="text-left">
                      <div className="text-white/80 text-xs">我的积分</div>
                      <div className="text-white font-bold text-xl">{pointsData?.points || 0}</div>
                    </div>
                  </div>
                  <Button
                    type="default"
                    size="small"
                    icon={signInStatus?.signed_today ? <GiftOutlined /> : null}
                    onClick={(e) => {
                      e.stopPropagation()
                      if (!signInStatus?.signed_today) signInMutation.mutate()
                    }}
                    loading={signInMutation.isPending}
                    disabled={signInStatus?.signed_today}
                    className="border-white/50 text-white hover:bg-white/20"
                    style={{ 
                      background: signInStatus?.signed_today ? 'rgba(255,255,255,0.2)' : 'rgba(255,255,255,0.3)',
                      borderColor: 'rgba(255,255,255,0.5)',
                      color: 'white'
                    }}
                  >
                    {signInStatus?.signed_today ? '已签' : '签到'}
                  </Button>
                </div>
                {/* 火苗显示 */}
                <div className="mt-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <SignInFlame days={signInStatus?.continue_days || 0} size="small" />
                    <div className="text-left">
                      <div className="text-white font-medium text-sm">
                        {getFlameConfig(signInStatus?.continue_days || 0).name}
                      </div>
                      <div className="text-white/70 text-xs">
                        连续 {signInStatus?.continue_days || 0} 天
                      </div>
                    </div>
                  </div>
                  {(signInStatus?.continue_days || 0) >= 7 && (
                    <div className="text-white/80 text-xs bg-white/20 px-2 py-1 rounded-full">
                      🎉 坚持就是胜利
                    </div>
                  )}
                </div>
              </div>
            </div>
            
            <Button 
              type="primary" 
              icon={<EditOutlined />} 
              onClick={openEditModal}
              className="mt-4"
              block
            >
              编辑资料
            </Button>
          </div>
          
          {/* 设备管理 - 放在头像卡片下方 */}
          <div className="mt-6">
            <DeviceManagement />
          </div>
        </Col>
        
        {/* 右侧：详细信息 */}
        <Col span={16}>
          {/* 基本信息 */}
          <div className="glass-card p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2">
              <UserOutlined className="text-blue-500" />
              基本信息
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-gray-50 rounded-lg">
                <div className="text-gray-500 text-sm mb-1">用户名</div>
                <div className="text-gray-800 font-medium">{user?.username}</div>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <div className="text-gray-500 text-sm mb-1">邮箱</div>
                <div className="text-gray-800 font-medium flex items-center gap-2">
                  <MailOutlined className="text-blue-500" />
                  {user?.email || '未绑定'}
                </div>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <div className="text-gray-500 text-sm mb-1">注册时间</div>
                <div className="text-gray-800 font-medium">
                  {user?.created_at ? dayjs(user.created_at).format('YYYY-MM-DD') : '-'}
                </div>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <div className="text-gray-500 text-sm mb-1">账号状态</div>
                <Tag color="green">正常</Tag>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <div className="text-gray-500 text-sm mb-1">用户角色</div>
                <Tag color={currentRoleConfig.tagColor}>{currentRoleConfig.name}</Tag>
              </div>
              <div 
                className="p-4 rounded-lg border-2"
                style={{ 
                  background: isAdmin 
                    ? '#f0f9ff'
                    : daysLeft <= 7 
                      ? '#fef2f2' 
                      : daysLeft <= 30 
                        ? '#fffbeb'
                        : '#f0fdf4',
                  borderColor: isAdmin
                    ? '#93c5fd'
                    : daysLeft <= 7 ? '#fca5a5' : daysLeft <= 30 ? '#fcd34d' : '#86efac'
                }}
              >
                <div className="text-gray-500 text-sm mb-1">会员到期</div>
                <div className="flex items-center gap-2">
                  <span 
                    className="font-bold text-lg"
                    style={{ 
                      color: isAdmin 
                        ? '#2563eb'
                        : daysLeft <= 7 ? '#dc2626' : daysLeft <= 30 ? '#d97706' : '#059669' 
                    }}
                  >
                    {formatExpireTime()}
                  </span>
                  {!isAdmin && daysLeft > 0 && (
                    <Tag 
                      color={daysLeft <= 7 ? 'red' : daysLeft <= 30 ? 'orange' : 'green'}
                      className="ml-1"
                    >
                      {daysLeft}天
                    </Tag>
                  )}
                </div>
              </div>
            </div>
          </div>
          
          {/* 账号安全 */}
          <div className="glass-card p-6">
            <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2">
              <SafetyOutlined className="text-green-500" />
              账号安全
            </h3>
            <div className="space-y-4">
              <div className="flex justify-between items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
                    <LockOutlined className="text-blue-500" />
                  </div>
                  <div>
                    <div className="font-medium text-gray-800">登录密码</div>
                    <div className="text-gray-500 text-sm">定期修改密码可以保护账号安全</div>
                  </div>
                </div>
                <Button onClick={() => setPasswordModalOpen(true)}>修改密码</Button>
              </div>
              
              <div className="flex justify-between items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center">
                    <MailOutlined className="text-green-500" />
                  </div>
                  <div>
                    <div className="font-medium text-gray-800">绑定邮箱</div>
                    <div className="text-gray-500 text-sm">已绑定：{user?.email || '未绑定'}</div>
                  </div>
                </div>
                <Button onClick={openEmailModal}>修改邮箱</Button>
              </div>
            </div>
          </div>
        </Col>
      </Row>

      {/* 编辑资料弹窗 */}
      <Modal
        title="编辑资料"
        open={editModalOpen}
        onCancel={() => setEditModalOpen(false)}
        footer={null}
      >
        <Form form={editForm} name="editProfileForm" onFinish={handleEditSubmit} layout="vertical" className="mt-4">
          <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}>
            <Input placeholder="请输入昵称" prefix={<UserOutlined className="text-gray-400" />} />
          </Form.Item>
          <Form.Item name="gender" label="性别">
            <Radio.Group>
              <Radio value={0}>保密</Radio>
              <Radio value={1}>男</Radio>
              <Radio value={2}>女</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              保存修改
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* 修改密码弹窗 */}
      <Modal
        title="修改密码"
        open={passwordModalOpen}
        onCancel={() => setPasswordModalOpen(false)}
        footer={null}
      >
        <Form form={passwordForm} name="changePasswordForm" onFinish={handlePasswordSubmit} layout="vertical" className="mt-4">
          <Form.Item name="old_password" label="原密码" rules={[{ required: true, message: '请输入原密码' }]}>
            <Input.Password placeholder="请输入原密码" prefix={<LockOutlined className="text-gray-400" />} />
          </Form.Item>
          <Form.Item name="new_password" label="新密码" rules={[
            { required: true, message: '请输入新密码' },
            { min: 8, message: '密码至少8位' },
          ]}>
            <Input.Password placeholder="请输入新密码" prefix={<LockOutlined className="text-gray-400" />} />
          </Form.Item>
          <Form.Item name="confirm_password" label="确认密码" rules={[{ required: true, message: '请确认新密码' }]}>
            <Input.Password placeholder="请再次输入新密码" prefix={<LockOutlined className="text-gray-400" />} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              确认修改
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* 修改邮箱弹窗 */}
      <Modal
        title="修改邮箱"
        open={emailModalOpen}
        onCancel={() => setEmailModalOpen(false)}
        footer={null}
      >
        <Form form={emailForm} name="changeEmailForm" onFinish={handleEmailSubmit} layout="vertical" className="mt-4">
          <Form.Item label="当前邮箱">
            <Input value={user?.email || '未绑定'} disabled prefix={<MailOutlined className="text-gray-400" />} />
          </Form.Item>
          <Form.Item 
            name="new_email" 
            label="新邮箱" 
            rules={[
              { required: true, message: '请输入新邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input placeholder="请输入新邮箱" prefix={<MailOutlined className="text-gray-400" />} />
          </Form.Item>
          <Form.Item 
            name="code" 
            label="验证码" 
            rules={[
              { required: true, message: '请输入验证码' },
              { len: 6, message: '验证码为6位' }
            ]}
          >
            <div className="flex gap-2">
              <Input 
                placeholder="请输入6位验证码" 
                maxLength={6}
                className="flex-1"
              />
              <Button 
                onClick={handleSendEmailCode} 
                loading={sendingCode}
                disabled={countdown > 0}
              >
                {countdown > 0 ? `${countdown}s` : '获取验证码'}
              </Button>
            </div>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              确认修改
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default UserCenter

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useSelector } from 'react-redux'
import { Button, Tag, Table, Modal, Input, App, Space, Progress, Dropdown } from 'antd'
import { CrownOutlined, CheckCircleOutlined, GiftOutlined, ClockCircleOutlined, CalendarOutlined, ShoppingCartOutlined, DownOutlined } from '@ant-design/icons'
import { RootState } from '../../store'
import { cardApi, publicApi } from '../../services/api'
import dayjs from 'dayjs'
import type { MenuProps } from 'antd'

interface RedeemHistory {
  id: number
  code: string
  card_type: number
  duration: number
  redeemed_at: string
}

interface RechargeLink {
  card_type: number
  name: string
  url: string
  enabled: boolean
}

const MemberCenter = () => {
  const { message } = App.useApp()
  const { user } = useSelector((state: RootState) => state.auth)
  const queryClient = useQueryClient()
  
  // 弹窗状态
  const [redeemModalOpen, setRedeemModalOpen] = useState(false)
  
  // 表单状态
  const [cardCode, setCardCode] = useState('')
  
  const userRole = user?.role || 0
  const isAdmin = userRole >= 2



  const { data: historyData } = useQuery({
    queryKey: ['redeemHistory'],
    queryFn: async () => {
      const response = await cardApi.getHistory({ page: 1, page_size: 10 })
      return response.data.data
    },
  })

  // 获取充值链接
  const { data: rechargeLinksData } = useQuery({
    queryKey: ['publicRechargeLinks'],
    queryFn: async () => {
      const response = await publicApi.getRechargeLinks()
      return response.data.data as { links: RechargeLink[] }
    },
  })

  const rechargeLinks = rechargeLinksData?.links || []
  const hasRechargeLinks = rechargeLinks.length > 0

  // 兑换会员卡密
  const redeemMutation = useMutation({
    mutationFn: (code: string) => cardApi.redeem(code),
    onSuccess: () => {
      message.success('兑换成功！会员已升级')
      setRedeemModalOpen(false)
      setCardCode('')
      queryClient.invalidateQueries({ queryKey: ['redeemHistory'] })
      window.location.reload()
    },
    onError: (err: Error) => {
      message.error(err.message || '兑换失败')
    },
  })




  // 角色配置
  const roleConfig = [
    { role: 0, name: '普通用户', color: '#999', tagColor: 'default', benefits: ['基础观影功能', '标清画质', '有广告'] },
    { role: 1, name: '会员用户', color: '#13c2c2', tagColor: 'cyan', benefits: ['无限观影', '免广告', '高清画质', '专属客服'] },
    { role: 2, name: '管理员', color: '#1890ff', tagColor: 'blue', benefits: ['所有会员权益', '用户管理', '卡密管理', '长期有效'] },
    { role: 3, name: '超级管理员', color: '#722ed1', tagColor: 'purple', benefits: ['所有管理员权益', '系统设置', '角色管理', '长期有效'] },
  ]
  
  // 计算会员剩余天数
  const getDaysLeft = () => {
    if (isAdmin) return -1 // 管理员返回-1表示长期
    if (!user?.member_expire) return 0
    const expireDate = new Date(user.member_expire)
    const now = new Date()
    const diff = expireDate.getTime() - now.getTime()
    return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
  }
  
  const daysLeft = getDaysLeft()

  const historyColumns = [
    { 
      title: '卡密码', 
      dataIndex: 'card_code', 
      key: 'card_code',
      width: 200,
      render: (code: string) => <code className="text-xs bg-gray-100 px-2 py-1 rounded">{code}</code>
    },
    {
      title: '天数',
      dataIndex: 'duration',
      key: 'duration',
      width: 80,
      render: (days: number) => <Tag color="blue">{days}天</Tag>,
    },
    {
      title: '兑换时间',
      dataIndex: 'redeem_time',
      key: 'redeem_time',
      width: 150,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '到期时间',
      dataIndex: 'expire_time',
      key: 'expire_time',
      width: 150,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
  ]

  const currentRole = roleConfig[userRole] || roleConfig[0]

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* 当前会员状态 */}
      <div className="glass-card p-6">
        <div className="flex items-center justify-between mb-6">
          <Space size="large">
            <CrownOutlined style={{ fontSize: 56, color: currentRole.color }} />
            <div>
              <div className="flex items-center gap-2 mb-1">
                <span className="text-2xl font-bold">{currentRole.name}</span>
                <Tag color={currentRole.tagColor}>{currentRole.name}</Tag>
              </div>
              {userRole >= 1 && (
                <Space className="text-gray-500" size="middle">
                  <span>
                    <CalendarOutlined className="mr-1" />
                    到期: {isAdmin ? '长期' : (user?.member_expire ? dayjs(user.member_expire).format('YYYY-MM-DD') : '未开通')}
                  </span>
                  {!isAdmin && daysLeft > 0 && (
                    <span className="text-green-600 font-medium">
                      <ClockCircleOutlined className="mr-1" />
                      剩余 {daysLeft} 天
                    </span>
                  )}
                </Space>
              )}
            </div>
          </Space>
          {userRole < 2 && (
            <Space>
              {hasRechargeLinks ? (
                rechargeLinks.length === 1 ? (
                  <Button 
                    type="primary" 
                    size="large" 
                    icon={<ShoppingCartOutlined />} 
                    onClick={() => window.open(rechargeLinks[0].url, '_blank')}
                  >
                    购买{rechargeLinks[0].name}
                  </Button>
                ) : (
                  <Dropdown
                    menu={{
                      items: rechargeLinks.map(link => ({
                        key: link.card_type,
                        label: link.name,
                        onClick: () => window.open(link.url, '_blank'),
                      })) as MenuProps['items'],
                    }}
                    placement="bottomRight"
                  >
                    <Button type="primary" size="large" icon={<ShoppingCartOutlined />}>
                      购买会员 <DownOutlined />
                    </Button>
                  </Dropdown>
                )
              ) : (
                <Button 
                  type="primary" 
                  size="large" 
                  icon={<ShoppingCartOutlined />} 
                  onClick={() => message.info('请联系管理员获取购买链接')}
                >
                  购买会员
                </Button>
              )}
              <Button size="large" icon={<GiftOutlined />} onClick={() => setRedeemModalOpen(true)}>
                兑换卡密
              </Button>
            </Space>
          )}
        </div>
        
        {/* 会员进度 */}
        {userRole === 1 && !isAdmin && daysLeft > 0 && (
          <div className="mt-4">
            <div className="flex justify-between text-sm text-gray-500 mb-2">
              <span>会员有效期</span>
              <span>{daysLeft} / 365 天</span>
            </div>
            <Progress
              percent={Math.min(100, (daysLeft / 365) * 100)}
              strokeColor={currentRole.color}
              trailColor="#e5e7eb"
              showInfo={false}
            />
          </div>
        )}
        
        {/* 权益展示 */}
        <div className="mt-6 grid grid-cols-4 gap-4">
          {currentRole.benefits.map((benefit, index) => (
            <div key={index} className="text-center p-3 bg-gray-50 rounded-lg">
              <CheckCircleOutlined className="text-green-500 text-xl mb-1" />
              <div className="text-sm text-gray-700">{benefit}</div>
            </div>
          ))}
        </div>
      </div>

      {/* 兑换记录 */}
      <div className="glass-card p-6 flex-1">
        <div className="font-semibold mb-4">兑换记录</div>
        <Table
          columns={historyColumns}
          dataSource={(historyData as { list: RedeemHistory[] })?.list || []}
          rowKey="id"
          pagination={false}
          size="small"
          locale={{ emptyText: '暂无兑换记录' }}
        />
      </div>

      {/* 兑换卡密弹窗 */}
      <Modal
        title="兑换卡密"
        open={redeemModalOpen}
        onCancel={() => { setRedeemModalOpen(false); setCardCode('') }}
        onOk={() => redeemMutation.mutate(cardCode)}
        confirmLoading={redeemMutation.isPending}
        okText="兑换"
      >
        <div className="py-4">
          <div className="mb-2 text-gray-600">请输入卡密码：</div>
          <Input
            size="large"
            placeholder="输入卡密码"
            value={cardCode}
            onChange={(e) => setCardCode(e.target.value.trim())}
            onPressEnter={() => cardCode && redeemMutation.mutate(cardCode)}
          />
          <div className="mt-4 text-gray-500 text-sm">
            提示：卡密码区分大小写，请确保输入正确。
          </div>
        </div>
      </Modal>

    </div>
  )
}

export default MemberCenter

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, Button, Tag, Table, App, Progress, Modal, Empty, Input, Dropdown, Spin, Row, Col } from 'antd'
import {
  GiftOutlined,
  CheckCircleOutlined,
  HistoryOutlined,
  SwapOutlined,
  CalendarOutlined,
  FireOutlined,
  TrophyOutlined,
  CreditCardOutlined,
  ShoppingCartOutlined,
  DownOutlined,
  CrownOutlined,
} from '@ant-design/icons'
import { pointsApi, publicApi } from '../../services/api'
import UserAvatar from '../../components/UserAvatar'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import dayjs from 'dayjs'
import type { MenuProps } from 'antd'

interface RankingItem {
  rank: number
  user_id: string
  username: string
  nickname: string
  avatar: string
  points: number
}

interface SignInStatus {
  signed_today: boolean
  continue_days: number
  month_count: number
  sign_dates: string[]
}

interface PointsRecord {
  id: number
  type: number
  type_name: string
  points: number
  points_before: number
  points_after: number
  remark: string
  created_at: string
}

interface ExchangeRule {
  id: number
  name: string
  points: number
  member_days: number
  description: string
  enabled: boolean
}

interface PointsRechargeLink {
  points: number
  name: string
  url: string
  enabled: boolean
}

const calcSignInPoints = (currentContinueDays: number) => {
  const base = 5
  const nextContinueDays = currentContinueDays + 1
  const bonus = Math.min(nextContinueDays - 1, 5)
  return base + bonus
}

const Points = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const { user } = useSelector((state: RootState) => state.auth)
  const [activeTab, setActiveTab] = useState<'records' | 'exchange'>('records')
  const [page, setPage] = useState(1)
  const [exchangeModalOpen, setExchangeModalOpen] = useState(false)
  const [selectedRule, setSelectedRule] = useState<ExchangeRule | null>(null)
  const [calendarMonth, setCalendarMonth] = useState(dayjs())
  const [redeemCode, setRedeemCode] = useState('')
  const [rechargeModalOpen, setRechargeModalOpen] = useState(false)
  const [rankingModalOpen, setRankingModalOpen] = useState(false)
  const [rankingPage, setRankingPage] = useState(1)

  // 获取积分
  const { data: pointsData } = useQuery({
    queryKey: ['myPoints'],
    queryFn: async () => {
      const res = await pointsApi.getMyPoints()
      return res.data.data as { points: number }
    },
  })

  // 获取签到状态
  const { data: signInStatus, refetch: refetchSignIn } = useQuery({
    queryKey: ['signInStatus'],
    queryFn: async () => {
      const res = await pointsApi.getSignInStatus()
      return res.data.data as SignInStatus
    },
  })

  // 获取积分记录
  const { data: recordsData, isLoading: recordsLoading } = useQuery({
    queryKey: ['pointsRecords', page],
    queryFn: async () => {
      const res = await pointsApi.getRecords({ page, page_size: 10 })
      return res.data.data as { list: PointsRecord[]; total: number }
    },
  })

  // 获取兑换规则
  const { data: exchangeRules } = useQuery({
    queryKey: ['exchangeRules'],
    queryFn: async () => {
      const res = await pointsApi.getExchangeRules()
      return res.data.data as ExchangeRule[]
    },
  })

  // 获取积分卡购买链接
  const { data: pointsRechargeData } = useQuery({
    queryKey: ['pointsRechargeLinks'],
    queryFn: async () => {
      const res = await publicApi.getPointsRechargeLinks()
      return res.data.data as { links: PointsRechargeLink[] }
    },
  })

  const pointsRechargeLinks = pointsRechargeData?.links || []
  const hasPointsRechargeLinks = pointsRechargeLinks.length > 0

  // 获取积分排行榜（首页前10）
  const { data: rankingData, isLoading: rankingLoading } = useQuery({
    queryKey: ['pointsRanking'],
    queryFn: async () => {
      const res = await pointsApi.getRanking({ page: 1, page_size: 10 })
      return res.data.data as { list: RankingItem[]; total: number }
    },
  })

  // 获取积分排行榜（弹窗分页）
  const { data: rankingModalData, isLoading: rankingModalLoading } = useQuery({
    queryKey: ['pointsRankingModal', rankingPage, rankingModalOpen],
    queryFn: async () => {
      const res = await pointsApi.getRanking({ page: rankingPage, page_size: 20 })
      return res.data.data as { list: RankingItem[]; total: number }
    },
    enabled: rankingModalOpen,
  })

  // 获取我的排名
  const { data: myRankData } = useQuery({
    queryKey: ['myPointsRank'],
    queryFn: async () => {
      const res = await pointsApi.getMyRank()
      return res.data.data as { rank: number }
    },
  })

  // 签到
  const signInMutation = useMutation({
    mutationFn: () => pointsApi.signIn(),
    onSuccess: (res) => {
      const result = res.data.data
      message.success(`签到成功！获得 ${result.points} 积分，连续签到 ${result.continue_days} 天`)
      queryClient.invalidateQueries({ queryKey: ['myPoints'] })
      queryClient.invalidateQueries({ queryKey: ['pointsRanking'] })
      queryClient.invalidateQueries({ queryKey: ['myPointsRank'] })
      refetchSignIn()
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || err.message || '签到失败')
    },
  })

  // 兑换
  const exchangeMutation = useMutation({
    mutationFn: (ruleId: number) => pointsApi.exchange(ruleId),
    onSuccess: () => {
      message.success('兑换成功！')
      setExchangeModalOpen(false)
      setSelectedRule(null)
      queryClient.invalidateQueries({ queryKey: ['myPoints'] })
      queryClient.invalidateQueries({ queryKey: ['pointsRanking'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || err.message || '兑换失败')
    },
  })

  // 卡密充值
  const redeemCardMutation = useMutation({
    mutationFn: (code: string) => pointsApi.redeemCard(code),
    onSuccess: (res) => {
      const result = res.data.data
      message.success(`充值成功！获得 ${result.points} 积分`)
      setRedeemCode('')
      queryClient.invalidateQueries({ queryKey: ['myPoints'] })
      queryClient.invalidateQueries({ queryKey: ['pointsRecords'] })
      queryClient.invalidateQueries({ queryKey: ['pointsRanking'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || err.message || '充值失败')
    },
  })

  const handleExchange = (rule: ExchangeRule) => {
    setSelectedRule(rule)
    setExchangeModalOpen(true)
  }

  const confirmExchange = () => {
    if (selectedRule) {
      exchangeMutation.mutate(selectedRule.id)
    }
  }

  // 生成日历数据
  const generateCalendarDays = () => {
    const startOfMonth = calendarMonth.startOf('month')
    const endOfMonth = calendarMonth.endOf('month')
    const startDay = startOfMonth.day()
    const daysInMonth = endOfMonth.date()
    const days: { date: dayjs.Dayjs; isCurrentMonth: boolean }[] = []
    for (let i = startDay; i > 0; i--) days.push({ date: startOfMonth.subtract(i, 'day'), isCurrentMonth: false })
    for (let i = 1; i <= daysInMonth; i++) days.push({ date: startOfMonth.date(i), isCurrentMonth: true })
    const remaining = 42 - days.length
    for (let i = 1; i <= remaining; i++) days.push({ date: endOfMonth.add(i, 'day'), isCurrentMonth: false })
    return days
  }

  const calendarDays = generateCalendarDays()
  const today = dayjs().format('YYYY-MM-DD')
  const weekDays = ['日', '一', '二', '三', '四', '五', '六']

  const columns = [
    {
      title: '类型',
      dataIndex: 'type_name',
      key: 'type_name',
      render: (name: string, record: PointsRecord) => <Tag color={record.points > 0 ? 'green' : 'red'}>{name}</Tag>,
    },
    {
      title: '积分',
      dataIndex: 'points',
      key: 'points',
      render: (points: number) => (
        <span className={points > 0 ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>
          {points > 0 ? '+' : ''}{points}
        </span>
      ),
    },
    { title: '余额', dataIndex: 'points_after', key: 'points_after', width: 80 },
    { title: '说明', dataIndex: 'remark', key: 'remark', ellipsis: true },
    { title: '时间', dataIndex: 'created_at', key: 'created_at', width: 100, render: (time: string) => dayjs(time).format('MM-DD HH:mm') },
  ]

  return (
    <Row gutter={16} className="h-full">
      {/* 左侧：积分卡片 + 排行榜 */}
      <Col xs={24} lg={8} className="flex flex-col gap-4">
        {/* 积分卡片 */}
        <div className="bg-gradient-to-br from-orange-500 via-orange-400 to-yellow-400 rounded-2xl p-5 text-white shadow-lg">
          <div className="flex items-center justify-between mb-4">
            <div>
              <div className="text-white/70 text-sm">我的积分</div>
              <div className="text-4xl font-bold">{pointsData?.points || 0}</div>
            </div>
            <TrophyOutlined className="text-5xl opacity-30" />
          </div>
          <div className="flex gap-6 text-sm">
            <div className="flex items-center gap-1">
              <FireOutlined />
              <span>连续 {signInStatus?.continue_days || 0} 天</span>
            </div>
            <div className="flex items-center gap-1">
              <CalendarOutlined />
              <span>本月 {signInStatus?.month_count || 0} 天</span>
            </div>
          </div>
          {myRankData && (
            <div className="mt-3 pt-3 border-t border-white/20 text-sm">
              <CrownOutlined className="mr-1" /> 排名第 <span className="font-bold">{myRankData.rank}</span> 名
            </div>
          )}
          {/* 充值按钮 */}
          <div className="mt-4 pt-3 border-t border-white/20">
            <Button type="default" ghost block icon={<CreditCardOutlined />} onClick={() => setRechargeModalOpen(true)}>
              积分充值
            </Button>
          </div>
        </div>

        {/* 排行榜 - 填满剩余空间 */}
        <Card 
          size="small" 
          title={<><CrownOutlined className="text-yellow-500 mr-2" />积分排行榜 TOP10</>} 
          extra={<Button type="link" size="small" onClick={() => setRankingModalOpen(true)}>查看全部</Button>}
          className="shadow-sm flex-1" 
          styles={{ body: { padding: '12px' } }}
        >
          {rankingLoading ? (
            <div className="flex justify-center py-8"><Spin /></div>
          ) : rankingData?.list && rankingData.list.length > 0 ? (
            <div className="space-y-1">
              {rankingData.list.map((item) => (
                <div
                  key={item.user_id}
                  className={`flex items-center gap-3 py-2 px-3 rounded-lg ${item.user_id === user?.id ? 'bg-orange-50 border border-orange-200' : 'hover:bg-gray-50'}`}
                >
                  <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                    item.rank === 1 ? 'bg-yellow-400 text-white' : item.rank === 2 ? 'bg-gray-300 text-gray-700' : item.rank === 3 ? 'bg-orange-300 text-white' : 'bg-gray-100 text-gray-500'
                  }`}>
                    {item.rank}
                  </div>
                  <UserAvatar size={28} src={item.avatar} name={item.nickname} />
                  <span className="flex-1 text-sm truncate text-gray-700">{item.nickname}</span>
                  <span className="text-orange-500 font-bold text-sm">{item.points}</span>
                </div>
              ))}
            </div>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据" />
          )}
        </Card>
      </Col>

      {/* 右侧：签到日历 + Tab内容 */}
      <Col xs={24} lg={16} className="flex flex-col gap-4">
        {/* 签到日历 */}
        <Card
          size="small"
          className="shadow-sm"
          title={
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <CalendarOutlined className="text-blue-500" />
                <Button size="small" type="text" onClick={() => setCalendarMonth(calendarMonth.subtract(1, 'month'))}>‹</Button>
                <span className="font-medium">{calendarMonth.format('YYYY年MM月')}</span>
                <Button size="small" type="text" onClick={() => setCalendarMonth(calendarMonth.add(1, 'month'))}>›</Button>
              </div>
              <Button
                type="primary"
                icon={<GiftOutlined />}
                onClick={() => signInMutation.mutate()}
                loading={signInMutation.isPending}
                disabled={signInStatus?.signed_today}
              >
                {signInStatus?.signed_today ? '已签到' : `签到 +${calcSignInPoints(signInStatus?.continue_days || 0)}`}
              </Button>
            </div>
          }
        >
          <div className="grid grid-cols-7 gap-1 mb-1">
            {weekDays.map((day) => (
              <div key={day} className="text-center text-gray-400 text-xs py-1">{day}</div>
            ))}
          </div>
          <div className="grid grid-cols-7 gap-1">
            {calendarDays.map(({ date, isCurrentMonth }, index) => {
              const dateStr = date.format('YYYY-MM-DD')
              const isSigned = signInStatus?.sign_dates?.includes(dateStr)
              const isToday = dateStr === today
              const isPast = date.isBefore(dayjs(), 'day')
              const isFuture = date.isAfter(dayjs(), 'day')
              const calcFuturePoints = () => {
                const currentContinue = signInStatus?.continue_days || 0
                const todaySigned = signInStatus?.signed_today
                const daysFromToday = date.diff(dayjs().startOf('day'), 'day')
                const predictedContinue = todaySigned ? currentContinue + daysFromToday : currentContinue + 1 + daysFromToday
                return 5 + Math.min(Math.max(predictedContinue - 1, 0), 5)
              }
              return (
                <div
                  key={index}
                  className={`py-2 rounded-lg text-center transition-all ${!isCurrentMonth ? 'opacity-20' : ''} ${isToday ? 'ring-2 ring-orange-400 bg-orange-50' : ''} ${isSigned && isCurrentMonth ? 'bg-green-50' : ''} ${isPast && isCurrentMonth && !isSigned ? 'bg-gray-50' : ''} ${isFuture && isCurrentMonth ? 'bg-blue-50/50' : ''}`}
                >
                  <div className={`text-sm font-medium ${isToday ? 'text-orange-600' : isSigned ? 'text-green-600' : 'text-gray-600'}`}>{date.date()}</div>
                  {isCurrentMonth && (
                    <div className="text-xs mt-0.5">
                      {isSigned ? <CheckCircleOutlined className="text-green-500" /> : isToday && !signInStatus?.signed_today ? (
                        <span className="text-orange-500 font-bold">+{calcSignInPoints(signInStatus?.continue_days || 0)}</span>
                      ) : isPast ? <span className="text-gray-300">-</span> : isFuture ? <span className="text-blue-400">+{calcFuturePoints()}</span> : null}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
          <div className="mt-2 text-xs text-gray-400 flex gap-4">
            <span>📌 第1天5分，每天+1，最高10分</span>
            <span className="text-blue-400">💡 蓝色为预测</span>
          </div>
        </Card>

        {/* Tab内容 - 填满剩余空间 */}
        <Card
          size="small"
          className="shadow-sm flex-1"
          title={
            <div className="flex gap-2">
              <Button type={activeTab === 'records' ? 'primary' : 'default'} size="small" icon={<HistoryOutlined />} onClick={() => setActiveTab('records')}>积分记录</Button>
              <Button type={activeTab === 'exchange' ? 'primary' : 'default'} size="small" icon={<SwapOutlined />} onClick={() => setActiveTab('exchange')}>积分兑换</Button>
            </div>
          }
        >
          {activeTab === 'records' && (
            <Table columns={columns} dataSource={recordsData?.list || []} rowKey="id" loading={recordsLoading} size="small"
              pagination={{ current: page, pageSize: 5, total: recordsData?.total || 0, onChange: setPage, showTotal: (t) => `共 ${t} 条`, size: 'small' }}
            />
          )}

          {activeTab === 'exchange' && (
            exchangeRules && exchangeRules.length > 0 ? (
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                {exchangeRules.map((rule) => (
                  <div key={rule.id} className="border rounded-lg p-3 hover:shadow-md transition-shadow bg-gradient-to-br from-white to-orange-50/30">
                    <div className="text-center">
                      <div className="text-xl font-bold text-orange-500">{rule.points}</div>
                      <div className="text-xs text-gray-400">积分</div>
                      <div className="font-medium text-gray-800 text-sm mt-1">{rule.name}</div>
                      <Progress percent={Math.min(100, ((pointsData?.points || 0) / rule.points) * 100)} size="small" showInfo={false} strokeColor="#f97316" className="my-2" />
                      <Button type="primary" size="small" block disabled={(pointsData?.points || 0) < rule.points} onClick={() => handleExchange(rule)}>
                        {(pointsData?.points || 0) >= rule.points ? '兑换' : `差${rule.points - (pointsData?.points || 0)}分`}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            ) : <Empty description="暂无兑换规则" />
          )}
        </Card>
      </Col>

      {/* 兑换确认弹窗 */}
      <Modal title="确认兑换" open={exchangeModalOpen} onCancel={() => setExchangeModalOpen(false)} onOk={confirmExchange} confirmLoading={exchangeMutation.isPending} okText="确认兑换">
        {selectedRule && (
          <div className="py-4 text-center">
            <div className="text-lg mb-2">使用 <span className="text-orange-500 font-bold">{selectedRule.points}</span> 积分兑换</div>
            <div className="text-xl font-bold text-blue-500 mb-2">{selectedRule.name}</div>
            <div className="text-gray-500">兑换后获得 {selectedRule.member_days} 天会员</div>
          </div>
        )}
      </Modal>

      {/* 积分充值弹窗 */}
      <Modal title="积分充值" open={rechargeModalOpen} onCancel={() => setRechargeModalOpen(false)} footer={null} width={400}>
        <div className="py-4">
          <div className="text-center mb-6">
            <CreditCardOutlined className="text-5xl text-orange-400" />
            <div className="text-gray-500 text-sm mt-2">输入积分卡密码进行充值</div>
          </div>
          <Input size="large" placeholder="请输入卡密" value={redeemCode} onChange={(e) => setRedeemCode(e.target.value.trim())} onPressEnter={() => redeemCode && redeemCardMutation.mutate(redeemCode)} className="font-mono" />
          <Button type="primary" size="large" block className="mt-4" loading={redeemCardMutation.isPending} disabled={!redeemCode} onClick={() => redeemCardMutation.mutate(redeemCode)}>立即充值</Button>
          {hasPointsRechargeLinks && (
            <div className="mt-4 pt-4 border-t text-center">
              <div className="text-gray-400 text-xs mb-2">还没有积分卡？</div>
              {pointsRechargeLinks.length === 1 ? (
                <Button type="link" icon={<ShoppingCartOutlined />} onClick={() => window.open(pointsRechargeLinks[0].url, '_blank')}>购买积分卡</Button>
              ) : (
                <Dropdown menu={{ items: pointsRechargeLinks.map((link, i) => ({ key: i, label: link.name, onClick: () => window.open(link.url, '_blank') })) as MenuProps['items'] }}>
                  <Button type="link" icon={<ShoppingCartOutlined />}>购买积分卡 <DownOutlined /></Button>
                </Dropdown>
              )}
            </div>
          )}
        </div>
      </Modal>

      {/* 积分排行榜弹窗 */}
      <Modal 
        title={<><CrownOutlined className="text-yellow-500 mr-2" />积分排行榜</>} 
        open={rankingModalOpen} 
        onCancel={() => { setRankingModalOpen(false); setRankingPage(1) }} 
        footer={null} 
        width={500}
      >
        {rankingModalLoading ? (
          <div className="flex justify-center py-8"><Spin /></div>
        ) : rankingModalData?.list && rankingModalData.list.length > 0 ? (
          <div>
            <div className="space-y-1 max-h-96 overflow-auto">
              {rankingModalData.list.map((item) => (
                <div
                  key={item.user_id}
                  className={`flex items-center gap-3 py-2 px-3 rounded-lg ${item.user_id === user?.id ? 'bg-orange-50 border border-orange-200' : 'hover:bg-gray-50'}`}
                >
                  <div className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                    item.rank === 1 ? 'bg-yellow-400 text-white' : item.rank === 2 ? 'bg-gray-300 text-gray-700' : item.rank === 3 ? 'bg-orange-300 text-white' : 'bg-gray-100 text-gray-500'
                  }`}>
                    {item.rank}
                  </div>
                  <UserAvatar size={32} src={item.avatar} name={item.nickname} />
                  <span className="flex-1 text-sm truncate text-gray-700">{item.nickname}</span>
                  <span className="text-orange-500 font-bold">{item.points}</span>
                </div>
              ))}
            </div>
            <div className="mt-4 flex justify-center gap-2">
              <Button size="small" disabled={rankingPage <= 1} onClick={() => setRankingPage(p => p - 1)}>上一页</Button>
              <span className="text-gray-500 text-sm leading-6">第 {rankingPage} 页 / 共 {Math.ceil((rankingModalData?.total || 0) / 20)} 页</span>
              <Button size="small" disabled={rankingPage >= Math.ceil((rankingModalData?.total || 0) / 20)} onClick={() => setRankingPage(p => p + 1)}>下一页</Button>
            </div>
          </div>
        ) : (
          <Empty description="暂无数据" />
        )}
      </Modal>
    </Row>
  )
}

export default Points

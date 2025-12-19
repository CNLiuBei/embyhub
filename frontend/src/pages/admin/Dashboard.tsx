import { useQuery } from '@tanstack/react-query'
import { Row, Col } from 'antd'
import {
  UserOutlined,
  CheckCircleOutlined,
  RiseOutlined,
  CrownOutlined,
  TeamOutlined,
  CalendarOutlined,
  ThunderboltOutlined,
  GiftOutlined,
} from '@ant-design/icons'
import { XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, AreaChart, Area } from 'recharts'
import { adminApi, pointsCardApi } from '../../services/api'

interface UserStats {
  total_users: number
  today_register: number
  week_register: number
  month_register: number
  total_members: number
  active_users: number
}

interface DailyStat {
  date: string
  register_count: number
}

interface CardStats {
  total_cards: number
  unused_cards: number
  used_cards: number
  disabled_cards: number
}

interface PointsCardStats {
  total_cards: number
  unused_cards: number
  used_cards: number
  disabled_cards: number
  total_points: number
  used_points: number
}

interface InviteStats {
  total_invites: number
  today_invites: number
  week_invites: number
  month_invites: number
  reward_days: number
}

const COLORS = ['#52c41a', '#1890ff', '#ff4d4f']
const POINTS_COLORS = ['#722ed1', '#1890ff', '#ff4d4f']

const Dashboard = () => {
  const { data: stats } = useQuery<UserStats>({
    queryKey: ['userStats'],
    queryFn: async () => {
      const response = await adminApi.getUserStats()
      return response.data.data as UserStats
    },
  })

  const { data: dailyStats } = useQuery<DailyStat[]>({
    queryKey: ['dailyStats'],
    queryFn: async () => {
      const response = await adminApi.getDailyStats(14)
      return response.data.data as DailyStat[]
    },
  })

  const { data: cardStats } = useQuery<CardStats>({
    queryKey: ['cardStats'],
    queryFn: async () => {
      const response = await adminApi.getCardStats()
      return response.data.data as CardStats
    },
  })

  const { data: pointsCardStats } = useQuery<PointsCardStats>({
    queryKey: ['pointsCardStats'],
    queryFn: async () => {
      const response = await pointsCardApi.getStats()
      return response.data.data as PointsCardStats
    },
  })

  const { data: inviteStats } = useQuery<InviteStats>({
    queryKey: ['inviteStats'],
    queryFn: async () => {
      const response = await adminApi.getInviteStats()
      return response.data.data as InviteStats
    },
  })

  const { data: visitRanking } = useQuery({
    queryKey: ['visitRanking'],
    queryFn: async () => {
      const response = await adminApi.getVisitRanking(10)
      return response.data.data as { rank: number; username: string; visits: number }[]
    },
  })

  const cardPieData = [
    { name: '未使用', value: cardStats?.unused_cards || 0 },
    { name: '已使用', value: cardStats?.used_cards || 0 },
    { name: '已禁用', value: cardStats?.disabled_cards || 0 },
  ]

  const pointsCardPieData = [
    { name: '未使用', value: pointsCardStats?.unused_cards || 0 },
    { name: '已使用', value: pointsCardStats?.used_cards || 0 },
    { name: '已禁用', value: pointsCardStats?.disabled_cards || 0 },
  ]

  // 计算会员转化率
  const memberRate = stats?.total_users ? ((stats.total_members / stats.total_users) * 100).toFixed(1) : '0'
  // 计算活跃率
  const activeRate = stats?.total_users ? ((stats.active_users / stats.total_users) * 100).toFixed(1) : '0'

  return (
    <div className="space-y-4">
      {/* 顶部统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={12} md={6}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center shadow-lg">
                <UserOutlined className="text-white text-xl" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-gray-500 text-xs">总用户数</div>
                <div className="text-2xl font-bold text-gray-800">{stats?.total_users || 0}</div>
                <div className="text-xs text-green-500">+{stats?.today_register || 0} 今日</div>
              </div>
            </div>
          </div>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-green-400 to-green-600 flex items-center justify-center shadow-lg">
                <CheckCircleOutlined className="text-white text-xl" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-gray-500 text-xs">活跃用户</div>
                <div className="text-2xl font-bold text-gray-800">{stats?.active_users || 0}</div>
                <div className="text-xs text-blue-500">{activeRate}% 活跃率</div>
              </div>
            </div>
          </div>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-yellow-400 to-orange-500 flex items-center justify-center shadow-lg">
                <CrownOutlined className="text-white text-xl" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-gray-500 text-xs">VIP会员</div>
                <div className="text-2xl font-bold text-gray-800">{stats?.total_members || 0}</div>
                <div className="text-xs text-orange-500">{memberRate}% 转化率</div>
              </div>
            </div>
          </div>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-400 to-purple-600 flex items-center justify-center shadow-lg">
                <RiseOutlined className="text-white text-xl" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-gray-500 text-xs">本周注册</div>
                <div className="text-2xl font-bold text-gray-800">{stats?.week_register || 0}</div>
                <div className="text-xs text-purple-500">本月 {stats?.month_register || 0}</div>
              </div>
            </div>
          </div>
        </Col>
      </Row>

      {/* 中间区域 */}
      <Row gutter={[16, 16]}>
        {/* 注册趋势图 */}
        <Col xs={24} lg={16}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <CalendarOutlined className="text-blue-500" />
                <span className="font-semibold">近14日注册趋势</span>
              </div>
              <div className="text-xs text-gray-400">
                总计 {dailyStats?.reduce((sum, d) => sum + d.register_count, 0) || 0} 人
              </div>
            </div>
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={dailyStats || []}>
                <defs>
                  <linearGradient id="colorRegister" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#1890ff" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#1890ff" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" tickFormatter={(v) => v?.slice(5)} tick={{ fontSize: 11 }} />
                <YAxis tick={{ fontSize: 11 }} />
                <Tooltip
                  contentStyle={{ borderRadius: 8, border: 'none', boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }}
                  formatter={(value: number) => [`${value} 人`, '注册数']}
                  labelFormatter={(label) => `日期: ${label}`}
                />
                <Area type="monotone" dataKey="register_count" stroke="#1890ff" strokeWidth={2} fill="url(#colorRegister)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </Col>

        {/* 邀请统计 */}
        <Col xs={24} lg={8}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-2 mb-4">
              <GiftOutlined className="text-pink-500" />
              <span className="font-semibold">邀请统计</span>
            </div>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-gray-500 text-sm">总邀请数</span>
                <span className="text-xl font-bold text-gray-800">{inviteStats?.total_invites || 0}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-500 text-sm">今日邀请</span>
                <span className="text-lg font-semibold text-green-500">{inviteStats?.today_invites || 0}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-500 text-sm">本周邀请</span>
                <span className="text-lg font-semibold text-blue-500">{inviteStats?.week_invites || 0}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-500 text-sm">本月邀请</span>
                <span className="text-lg font-semibold text-purple-500">{inviteStats?.month_invites || 0}</span>
              </div>
              <div className="pt-2 border-t border-gray-100">
                <div className="flex justify-between items-center">
                  <span className="text-gray-500 text-sm">邀请奖励</span>
                  <span className="text-orange-500 font-semibold">{inviteStats?.reward_days || 0} 天/人</span>
                </div>
              </div>
            </div>
          </div>
        </Col>
      </Row>

      {/* 底部区域 */}
      <Row gutter={[16, 16]}>
        {/* 会员卡统计 */}
        <Col xs={24} md={8}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-2 mb-4">
              <CrownOutlined className="text-yellow-500" />
              <span className="font-semibold">会员卡统计</span>
              <span className="ml-auto text-xs text-gray-400">共 {cardStats?.total_cards || 0} 张</span>
            </div>
            <Row>
              <Col span={12}>
                <ResponsiveContainer width="100%" height={120}>
                  <PieChart>
                    <Pie data={cardPieData} cx="50%" cy="50%" innerRadius={30} outerRadius={50} dataKey="value" paddingAngle={2}>
                      {cardPieData.map((_, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                  </PieChart>
                </ResponsiveContainer>
              </Col>
              <Col span={12} className="flex flex-col justify-center">
                <div className="flex items-center gap-2 mb-2">
                  <span className="w-3 h-3 rounded-full bg-green-500" />
                  <span className="text-xs text-gray-500">未使用</span>
                  <span className="ml-auto text-green-500 font-bold text-sm">{cardStats?.unused_cards || 0}</span>
                </div>
                <div className="flex items-center gap-2 mb-2">
                  <span className="w-3 h-3 rounded-full bg-blue-500" />
                  <span className="text-xs text-gray-500">已使用</span>
                  <span className="ml-auto text-blue-500 font-bold text-sm">{cardStats?.used_cards || 0}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-red-500" />
                  <span className="text-xs text-gray-500">已禁用</span>
                  <span className="ml-auto text-red-500 font-bold text-sm">{cardStats?.disabled_cards || 0}</span>
                </div>
              </Col>
            </Row>
          </div>
        </Col>

        {/* 积分卡统计 */}
        <Col xs={24} md={8}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-2 mb-4">
              <ThunderboltOutlined className="text-purple-500" />
              <span className="font-semibold">积分卡统计</span>
              <span className="ml-auto text-xs text-gray-400">共 {pointsCardStats?.total_cards || 0} 张</span>
            </div>
            <Row>
              <Col span={12}>
                <ResponsiveContainer width="100%" height={120}>
                  <PieChart>
                    <Pie data={pointsCardPieData} cx="50%" cy="50%" innerRadius={30} outerRadius={50} dataKey="value" paddingAngle={2}>
                      {pointsCardPieData.map((_, index) => (
                        <Cell key={`cell-${index}`} fill={POINTS_COLORS[index % POINTS_COLORS.length]} />
                      ))}
                    </Pie>
                  </PieChart>
                </ResponsiveContainer>
              </Col>
              <Col span={12} className="flex flex-col justify-center">
                <div className="flex items-center gap-2 mb-2">
                  <span className="w-3 h-3 rounded-full bg-purple-500" />
                  <span className="text-xs text-gray-500">未使用</span>
                  <span className="ml-auto text-purple-500 font-bold text-sm">{pointsCardStats?.unused_cards || 0}</span>
                </div>
                <div className="flex items-center gap-2 mb-2">
                  <span className="w-3 h-3 rounded-full bg-blue-500" />
                  <span className="text-xs text-gray-500">已使用</span>
                  <span className="ml-auto text-blue-500 font-bold text-sm">{pointsCardStats?.used_cards || 0}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-red-500" />
                  <span className="text-xs text-gray-500">已禁用</span>
                  <span className="ml-auto text-red-500 font-bold text-sm">{pointsCardStats?.disabled_cards || 0}</span>
                </div>
              </Col>
            </Row>
            <div className="mt-3 pt-3 border-t border-gray-100 flex justify-between text-xs">
              <span className="text-gray-400">总积分</span>
              <span className="text-purple-500 font-semibold">{pointsCardStats?.total_points?.toLocaleString() || 0}</span>
            </div>
          </div>
        </Col>

        {/* 今日访问排行 */}
        <Col xs={24} md={8}>
          <div className="glass-card p-4 h-full">
            <div className="flex items-center gap-2 mb-4">
              <TeamOutlined className="text-blue-500" />
              <span className="font-semibold">今日访问排行</span>
            </div>
            <div className="space-y-2 max-h-[180px] overflow-auto">
              {visitRanking?.length ? (
                visitRanking.slice(0, 5).map((item, index) => (
                  <div key={item.rank} className="flex items-center gap-2 py-1">
                    <div
                      className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold ${
                        index === 0
                          ? 'bg-yellow-400 text-white'
                          : index === 1
                            ? 'bg-gray-300 text-gray-600'
                            : index === 2
                              ? 'bg-orange-300 text-white'
                              : 'bg-gray-100 text-gray-500'
                      }`}
                    >
                      {item.rank}
                    </div>
                    <span className="flex-1 text-sm text-gray-700 truncate">{item.username}</span>
                    <span className="text-blue-500 font-semibold text-sm">{item.visits}</span>
                  </div>
                ))
              ) : (
                <div className="text-center text-gray-400 py-8 text-sm">暂无数据</div>
              )}
            </div>
          </div>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard

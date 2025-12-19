import { useQuery } from '@tanstack/react-query'
import { Card, Button, Table, App, Statistic, Row, Col, Empty } from 'antd'
import { CopyOutlined, GiftOutlined, UserAddOutlined, TrophyOutlined } from '@ant-design/icons'
import { inviteApi } from '../../services/api'
import UserAvatar from '../../components/UserAvatar'
import dayjs from 'dayjs'

interface InviteInfo {
  invite_code: string
  invite_count: number
  reward_days: number
}

interface InviteRecord {
  id: number
  invitee: string
  reward_days: number
  created_at: string
}

interface RankingItem {
  rank: number
  nickname: string
  avatar: string
  invite_count: number
}

const InviteFriends = () => {
  const { message } = App.useApp()

  const { data: info } = useQuery({
    queryKey: ['myInviteInfo'],
    queryFn: async () => {
      const response = await inviteApi.getMyInviteInfo()
      return response.data.data as InviteInfo
    },
  })

  const { data: records } = useQuery({
    queryKey: ['myInviteRecords'],
    queryFn: async () => {
      const response = await inviteApi.getMyInviteRecords()
      return response.data.data as { list: InviteRecord[]; total: number }
    },
  })

  const { data: ranking } = useQuery({
    queryKey: ['inviteRanking'],
    queryFn: async () => {
      const response = await inviteApi.getInviteRanking(10)
      return response.data.data as RankingItem[]
    },
  })

  const inviteLink = info?.invite_code 
    ? `${window.location.origin}/register?invite=${info.invite_code}`
    : ''

  const copyLink = () => {
    navigator.clipboard.writeText(inviteLink)
    message.success('邀请链接已复制到剪贴板')
  }

  const copyCode = () => {
    if (info?.invite_code) {
      navigator.clipboard.writeText(info.invite_code)
      message.success('邀请码已复制')
    }
  }

  const columns = [
    { title: '被邀请人', dataIndex: 'invitee', key: 'invitee' },
    { title: '奖励天数', dataIndex: 'reward_days', key: 'reward_days', width: 100 },
    {
      title: '邀请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      {/* 邀请信息卡片 */}
      <Card className="glass-card">
        <div className="text-center mb-6">
          <GiftOutlined className="text-5xl text-purple-500 mb-3" />
          <h2 className="text-xl font-bold m-0">邀请好友，获得奖励</h2>
          <p className="text-gray-500 mt-2">
            每成功邀请一位好友注册，您将获得 <span className="text-purple-500 font-bold">{info?.reward_days || 0}天</span> 会员奖励
          </p>
        </div>

        {/* 邀请码和链接 */}
        <div className="bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg p-6 mb-6">
          <div className="text-center mb-4">
            <div className="text-gray-500 text-sm mb-1">我的邀请码</div>
            <div className="text-3xl font-bold text-purple-600 tracking-widest">
              {info?.invite_code || '---'}
            </div>
            <Button type="link" icon={<CopyOutlined />} onClick={copyCode}>
              复制邀请码
            </Button>
          </div>
          
          <div className="flex items-center gap-2">
            <input
              type="text"
              readOnly
              value={inviteLink}
              className="flex-1 px-3 py-2 bg-white rounded border text-sm text-gray-600"
            />
            <Button type="primary" icon={<CopyOutlined />} onClick={copyLink}>
              复制链接
            </Button>
          </div>
        </div>

        {/* 统计 */}
        <Row gutter={16}>
          <Col span={12}>
            <Card size="small" className="text-center">
              <Statistic 
                title="已邀请人数" 
                value={info?.invite_count || 0} 
                prefix={<UserAddOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={12}>
            <Card size="small" className="text-center">
              <Statistic 
                title="累计获得天数" 
                value={(info?.invite_count || 0) * (info?.reward_days || 0)} 
                suffix="天"
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>
      </Card>

      <Row gutter={16}>
        {/* 邀请记录 */}
        <Col span={14}>
          <Card title="邀请记录" className="glass-card h-full">
            {records?.list?.length ? (
              <Table
                columns={columns}
                dataSource={records.list}
                rowKey="id"
                pagination={false}
                size="small"
              />
            ) : (
              <Empty description="暂无邀请记录" />
            )}
          </Card>
        </Col>

        {/* 邀请排行榜 */}
        <Col span={10}>
          <Card 
            title={<><TrophyOutlined className="text-yellow-500 mr-2" />邀请排行榜</>} 
            className="glass-card h-full"
          >
            {ranking?.length ? (
              <div className="space-y-3">
                {ranking.map((item, index) => (
                  <div key={index} className="flex items-center gap-3">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-sm font-bold ${
                      index === 0 ? 'bg-yellow-400 text-white' :
                      index === 1 ? 'bg-gray-300 text-white' :
                      index === 2 ? 'bg-orange-400 text-white' :
                      'bg-gray-100 text-gray-500'
                    }`}>
                      {item.rank}
                    </div>
                    <UserAvatar src={item.avatar} name={item.nickname} size="small" />
                    <span className="flex-1 truncate">{item.nickname}</span>
                    <span className="text-purple-500 font-medium">{item.invite_count}人</span>
                  </div>
                ))}
              </div>
            ) : (
              <Empty description="暂无排行数据" />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default InviteFriends

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Button, InputNumber, App, Statistic, Row, Col, Card } from 'antd'
import { SettingOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'
import dayjs from 'dayjs'

interface InviteRecord {
  id: number
  inviter: string
  invitee: string
  reward_days: number
  created_at: string
}

interface InviteStats {
  total_users: number
  invited_users: number
  total_invites: number
  reward_days: number
}

const InviteManagement = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [rewardDays, setRewardDays] = useState<number>(7)

  const { data, isLoading } = useQuery({
    queryKey: ['inviteRecords', page, pageSize],
    queryFn: async () => {
      const response = await adminApi.getInviteRecords({ page, page_size: pageSize })
      return response.data.data as { list: InviteRecord[]; total: number }
    },
  })

  const { data: stats } = useQuery({
    queryKey: ['inviteStats'],
    queryFn: async () => {
      const response = await adminApi.getInviteStats()
      const data = response.data.data as InviteStats
      setRewardDays(data.reward_days)
      return data
    },
  })

  const setRewardMutation = useMutation({
    mutationFn: (days: number) => adminApi.setInviteRewardDays(days),
    onSuccess: () => {
      message.success('设置成功')
      queryClient.invalidateQueries({ queryKey: ['inviteStats'] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '设置失败'),
  })

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '邀请人', dataIndex: 'inviter', key: 'inviter' },
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
    <div className="flex flex-col h-full gap-4">
      {/* 统计卡片 */}
      <Row gutter={16}>
        <Col span={6}>
          <Card size="small" className="glass-card">
            <Statistic title="总用户数" value={stats?.total_users || 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" className="glass-card">
            <Statistic title="被邀请注册" value={stats?.invited_users || 0} valueStyle={{ color: '#52c41a' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" className="glass-card">
            <Statistic title="邀请记录数" value={stats?.total_invites || 0} valueStyle={{ color: '#1890ff' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" className="glass-card">
            <Statistic title="当前奖励天数" value={stats?.reward_days || 0} suffix="天" valueStyle={{ color: '#722ed1' }} />
          </Card>
        </Col>
      </Row>

      {/* 奖励设置 */}
      <div className="glass-card p-4">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-semibold m-0">邀请奖励设置</h2>
            <p className="text-gray-500 text-sm m-0 mt-1">用户邀请好友注册成功后，邀请人获得的会员天数奖励</p>
          </div>
          <div className="flex items-center gap-2">
            <InputNumber 
              min={0} 
              max={365} 
              value={rewardDays} 
              onChange={(v) => setRewardDays(v || 0)}
            />
            <span className="text-gray-500">天</span>
            <Button 
              type="primary" 
              icon={<SettingOutlined />}
              loading={setRewardMutation.isPending}
              onClick={() => setRewardMutation.mutate(rewardDays)}
            >
              保存设置
            </Button>
          </div>
        </div>
      </div>

      {/* 邀请记录 */}
      <div className="glass-card p-4 flex-1">
        <h3 className="font-semibold mb-4">邀请记录</h3>
        <Table
          columns={columns}
          dataSource={data?.list || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: page,
            pageSize,
            total: data?.total || 0,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) },
          }}
        />
      </div>
    </div>
  )
}

export default InviteManagement

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, Avatar, Button, Spin, App, Empty, Popconfirm } from 'antd'
import { UserOutlined, StopOutlined, DeleteOutlined } from '@ant-design/icons'
import { blacklistApi } from '../../services/api'
import dayjs from 'dayjs'

interface BlacklistItem {
  id: number
  user_id: string
  blocked_id: string
  reason: string
  created_at: string
  blocked_user?: {
    id: string
    username: string
    nickname: string
    avatar: string
  }
}

const Blacklist = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page] = useState(1)

  // 获取黑名单列表
  const { data, isLoading } = useQuery({
    queryKey: ['blacklist', page],
    queryFn: async () => {
      const res = await blacklistApi.getBlacklist({ page, page_size: 50 })
      return res.data.data as { list: BlacklistItem[]; total: number }
    },
  })

  // 取消拉黑
  const unblockMutation = useMutation({
    mutationFn: (userId: string) => blacklistApi.unblockUser(userId),
    onSuccess: () => {
      message.success('已取消拉黑')
      queryClient.invalidateQueries({ queryKey: ['blacklist'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '操作失败')
    },
  })

  return (
    <div className="max-w-3xl mx-auto">
      <Card
        title={
          <div className="flex items-center gap-2">
            <StopOutlined className="text-red-500" />
            <span>黑名单管理</span>
          </div>
        }
        className="glass-card"
      >
        <Spin spinning={isLoading}>
          {data?.list?.length ? (
            <div className="space-y-3">
              {data.list.map((item) => (
                <div
                  key={item.id}
                  className="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <Avatar
                      src={item.blocked_user?.avatar}
                      size={48}
                      icon={<UserOutlined />}
                    >
                      {item.blocked_user?.nickname?.[0]}
                    </Avatar>
                    <div>
                      <div className="font-medium">
                        {item.blocked_user?.nickname || item.blocked_user?.username || '未知用户'}
                      </div>
                      <div className="text-gray-400 text-sm">
                        {item.reason ? `原因: ${item.reason}` : '无备注'}
                      </div>
                      <div className="text-gray-400 text-xs">
                        拉黑时间: {dayjs(item.created_at).format('YYYY-MM-DD HH:mm')}
                      </div>
                    </div>
                  </div>
                  <Popconfirm
                    title="确定要取消拉黑吗？"
                    onConfirm={() => unblockMutation.mutate(item.blocked_id)}
                    okText="确定"
                    cancelText="取消"
                  >
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      loading={unblockMutation.isPending}
                    >
                      取消拉黑
                    </Button>
                  </Popconfirm>
                </div>
              ))}
            </div>
          ) : (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="黑名单为空"
            />
          )}
        </Spin>
      </Card>
    </div>
  )
}

export default Blacklist

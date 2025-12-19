import { useState } from 'react'
import { List, Button, Badge, Empty, message, Tag, Space } from 'antd'
import { DeleteOutlined, CheckOutlined, BellOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../services/api'

interface NotificationItem {
  id: number
  title: string
  content: string
  type: number
  is_read: boolean
  created_at: string
}

const typeMap: Record<number, { color: string; text: string }> = {
  1: { color: 'blue', text: '系统' },
  2: { color: 'green', text: '会员' },
  3: { color: 'orange', text: '活动' },
}

const Notifications = () => {
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['notifications', page],
    queryFn: async () => {
      const response = await api.get('/notification/list', { params: { page, page_size: 20 } })
      return response.data.data
    },
  })

  const { data: unreadCount } = useQuery({
    queryKey: ['unreadCount'],
    queryFn: async () => {
      const response = await api.get('/notification/unread-count')
      return response.data.data
    },
  })

  const readMutation = useMutation({
    mutationFn: (id: number) => api.post('/notification/read', { id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['unreadCount'] })
    },
  })

  const readAllMutation = useMutation({
    mutationFn: () => api.post('/notification/read-all'),
    onSuccess: () => {
      message.success('已全部标记为已读')
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['unreadCount'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.delete('/notification/delete', { data: { id } }),
    onSuccess: () => {
      message.success('已删除')
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
  })

  const list = (data as { list: NotificationItem[]; total: number })?.list || []
  const total = (data as { list: NotificationItem[]; total: number })?.total || 0
  const unread = (unreadCount as { count: number })?.count || 0

  return (
    <div className="glass-card p-6 h-full flex flex-col">
      <div className="flex justify-between items-center mb-4">
        <Space>
          <BellOutlined />
          <span className="font-semibold">消息通知</span>
          {unread > 0 && <Badge count={unread} />}
        </Space>
        {unread > 0 && (
          <Button icon={<CheckOutlined />} onClick={() => readAllMutation.mutate()}>
            全部已读
          </Button>
        )}
      </div>
      {list.length === 0 ? (
        <Empty description="暂无消息" />
      ) : (
        <List
          loading={isLoading}
          dataSource={list}
          pagination={{
            current: page,
            total,
            pageSize: 20,
            onChange: setPage,
          }}
          renderItem={(item: NotificationItem) => (
            <List.Item
              className={!item.is_read ? 'bg-blue-50' : ''}
              actions={[
                !item.is_read && (
                  <Button
                    key="read"
                    type="link"
                    size="small"
                    onClick={() => readMutation.mutate(item.id)}
                  >
                    标为已读
                  </Button>
                ),
                <Button
                  key="delete"
                  type="link"
                  danger
                  size="small"
                  icon={<DeleteOutlined />}
                  onClick={() => deleteMutation.mutate(item.id)}
                />,
              ].filter(Boolean)}
            >
              <List.Item.Meta
                title={
                  <Space>
                    {!item.is_read && <Badge status="processing" />}
                    <Tag color={typeMap[item.type]?.color}>{typeMap[item.type]?.text}</Tag>
                    {item.title}
                  </Space>
                }
                description={
                  <div>
                    <div className="mb-1">{item.content}</div>
                    <div className="text-gray-400 text-xs">
                      {item.created_at?.slice(0, 19).replace('T', ' ')}
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}
    </div>
  )
}

export default Notifications

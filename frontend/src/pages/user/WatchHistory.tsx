import { useState } from 'react'
import { List, Button, Empty, Modal, message, Avatar } from 'antd'
import { DeleteOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../services/api'

interface WatchItem {
  id: number
  video_id: string
  video_name: string
  duration: number
  progress: number
  updated_at: string
}

const WatchHistory = () => {
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['watchHistory', page],
    queryFn: () => api.get('/watch/history', { params: { page, page_size: 20 } }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.delete('/watch/history', { data: { id } }),
    onSuccess: () => {
      message.success('已删除')
      queryClient.invalidateQueries({ queryKey: ['watchHistory'] })
    },
  })

  const clearMutation = useMutation({
    mutationFn: () => api.delete('/watch/history/clear'),
    onSuccess: () => {
      message.success('已清空')
      queryClient.invalidateQueries({ queryKey: ['watchHistory'] })
    },
  })

  const handleClear = () => {
    Modal.confirm({
      title: '确认清空',
      content: '确定要清空所有观影记录吗？此操作不可恢复。',
      onOk: () => clearMutation.mutate(),
    })
  }

  const formatProgress = (progress: number, duration: number) => {
    const percent = duration > 0 ? Math.round((progress / duration) * 100) : 0
    const formatTime = (s: number) => {
      const h = Math.floor(s / 3600)
      const m = Math.floor((s % 3600) / 60)
      const sec = s % 60
      return h > 0 ? `${h}:${m.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}` : `${m}:${sec.toString().padStart(2, '0')}`
    }
    return `${formatTime(progress)} / ${formatTime(duration)} (${percent}%)`
  }

  const list = (data as unknown as { list: WatchItem[]; total: number })?.list || []
  const total = (data as unknown as { list: WatchItem[]; total: number })?.total || 0

  return (
    <div className="glass-card p-6 h-full flex flex-col">
      <div className="flex justify-between items-center mb-4">
        <span className="font-semibold">观影记录</span>
        {list.length > 0 && (
          <Button danger onClick={handleClear} loading={clearMutation.isPending}>
            清空记录
          </Button>
        )}
      </div>
      {list.length === 0 ? (
        <Empty description="暂无观影记录" />
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
          renderItem={(item: WatchItem) => (
            <List.Item
              actions={[
                <Button
                  key="play"
                  type="link"
                  icon={<PlayCircleOutlined />}
                  onClick={() => message.info('播放功能待实现')}
                >
                  继续观看
                </Button>,
                <Button
                  key="delete"
                  type="link"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => deleteMutation.mutate(item.id)}
                >
                  删除
                </Button>,
              ]}
            >
              <List.Item.Meta
                avatar={<Avatar shape="square" size={64} icon={<PlayCircleOutlined />} />}
                title={item.video_name}
                description={
                  <div>
                    <div>观看进度: {formatProgress(item.progress, item.duration)}</div>
                    <div className="text-gray-400 text-sm">
                      {item.updated_at?.slice(0, 19).replace('T', ' ')}
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

export default WatchHistory

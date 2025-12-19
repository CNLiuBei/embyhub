import { useState } from 'react'
import { List, Button, Empty, message, Avatar } from 'antd'
import { DeleteOutlined, HeartFilled, PlayCircleOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../services/api'

interface FavoriteItem {
  id: number
  video_id: string
  video_name: string
  created_at: string
}

const Favorites = () => {
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['favorites', page],
    queryFn: () => api.get('/favorite/list', { params: { page, page_size: 20 } }),
  })

  const removeMutation = useMutation({
    mutationFn: (videoId: string) => api.delete('/favorite/remove', { data: { video_id: videoId } }),
    onSuccess: () => {
      message.success('已取消收藏')
      queryClient.invalidateQueries({ queryKey: ['favorites'] })
    },
  })

  const list = (data as unknown as { list: FavoriteItem[]; total: number })?.list || []
  const total = (data as unknown as { list: FavoriteItem[]; total: number })?.total || 0

  return (
    <div className="glass-card p-6 h-full flex flex-col">
      <div className="font-semibold mb-4">我的收藏</div>
      {list.length === 0 ? (
        <Empty description="暂无收藏" />
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
          renderItem={(item: FavoriteItem) => (
            <List.Item
              actions={[
                <Button
                  key="play"
                  type="link"
                  icon={<PlayCircleOutlined />}
                  onClick={() => message.info('播放功能待实现')}
                >
                  播放
                </Button>,
                <Button
                  key="remove"
                  type="link"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => removeMutation.mutate(item.video_id)}
                >
                  取消收藏
                </Button>,
              ]}
            >
              <List.Item.Meta
                avatar={<Avatar shape="square" size={64} icon={<HeartFilled style={{ color: '#ff4d4f' }} />} />}
                title={item.video_name}
                description={`收藏时间: ${item.created_at?.slice(0, 19).replace('T', ' ')}`}
              />
            </List.Item>
          )}
        />
      )}
    </div>
  )
}

export default Favorites

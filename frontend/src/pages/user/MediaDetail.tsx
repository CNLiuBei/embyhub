import { useQuery } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Empty, Spin, Tag, Pagination, Rate, Button } from 'antd'
import { ArrowLeftOutlined, PlayCircleOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { mediaApi } from '../../services/api'
import { useState } from 'react'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'

interface MediaItem {
  guid: string
  title: string
  original_title?: string
  type: string  // Movie/TV/Episode
  poster?: string
  vote_average?: string
  release_date?: string
  year?: number
  overview?: string
  number_of_episodes?: number
  local_number_of_episodes?: number
  duration?: number
}

const MediaDetail = () => {
  const { guid } = useParams<{ guid: string }>()
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const pageSize = 20
  const { user } = useSelector((state: RootState) => state.auth)

  const { data, isLoading } = useQuery({
    queryKey: ['mediaItems', guid, page],
    queryFn: () => mediaApi.getMediaDBItems(guid!, { page, page_size: pageSize }),
    enabled: !!guid,
  })

  const mediaItems: MediaItem[] = data?.data?.data?.items || []
  const total = data?.data?.data?.total || 0
  const imageURL = data?.data?.data?.image_url || ''
  
  // 构建带用户ID的图片URL
  const buildImageURL = (poster: string) => {
    const fullUrl = `${imageURL}${poster}`
    if (!user?.id) return fullUrl
    // 检查URL是否已有查询参数
    const separator = fullUrl.includes('?') ? '&' : '?'
    return `${fullUrl}${separator}uid=${user.id}`
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="glass-card p-4">
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/user/media')}
          type="text"
        >
          返回媒体库列表
        </Button>
      </div>

      {mediaItems.length === 0 ? (
        <div className="glass-card p-6">
          <Empty description="该媒体库暂无内容" />
        </div>
      ) : (
        <>
          <div className="glass-card p-4">
            <h2 className="text-lg font-semibold">
              共 {total} 个影片
            </h2>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
            {mediaItems.map((item) => (
              <Card
                key={item.guid}
                hoverable
                className="glass-card overflow-hidden"
                styles={{ body: { padding: '12px' } }}
                cover={
                  item.poster ? (
                    <div className="relative aspect-[2/3] overflow-hidden bg-gradient-to-br from-gray-200 to-gray-300 group">
                      <img
                        src={buildImageURL(item.poster)}
                        alt={item.title}
                        className="w-full h-full object-cover transition-all duration-300 group-hover:scale-105"
                        loading="lazy"
                        onError={(e) => {
                          const target = e.target as HTMLImageElement
                          // 使用一个简单的占位图URL，或者隐藏图片显示父容器的背景
                          target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjMwMCIgZmlsbD0iI2E1YjRmYyIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwiIGZvbnQtc2l6ZT0iMTgiIGZpbGw9IiNmZmZmZmYiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj7lm77niYfliqDovb3lpLHotKU8L3RleHQ+PC9zdmc+'
                          target.style.opacity = '0.5'
                        }}
                      />
                      {/* 评分显示 - 始终可见 */}
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent pointer-events-none">
                        <div className="absolute bottom-0 left-0 right-0 p-2">
                          {item.vote_average && parseFloat(item.vote_average) > 0 && (
                            <div className="flex items-center gap-1 text-yellow-400 text-xs bg-black/40 rounded px-2 py-1 backdrop-blur-sm inline-flex">
                              <Rate disabled defaultValue={parseFloat(item.vote_average) / 2} count={5} style={{ fontSize: 12 }} />
                              <span className="font-semibold">{parseFloat(item.vote_average).toFixed(1)}</span>
                            </div>
                          )}
                        </div>
                      </div>
                      {/* 播放按钮 - 悬停显示 */}
                      <div className="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        <div className="transform scale-100 group-hover:scale-110 transition-transform duration-300">
                          <PlayCircleOutlined className="text-white text-6xl drop-shadow-2xl" />
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="aspect-[2/3] bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center">
                      <PlayCircleOutlined className="text-white text-6xl opacity-30" />
                    </div>
                  )
                }
              >
                <div className="space-y-1">
                  <div className="font-medium text-sm line-clamp-2" title={item.title}>
                    {item.title}
                  </div>
                  <div className="flex items-center gap-2 text-xs text-gray-500">
                    {item.year && <span>{item.year}</span>}
                    {item.duration && (
                      <span className="flex items-center gap-1">
                        <ClockCircleOutlined />
                        {item.duration}min
                      </span>
                    )}
                  </div>
                  <div className="flex flex-wrap gap-1">
                    <Tag color={item.type === 'TV' ? 'blue' : item.type === 'Movie' ? 'purple' : 'green'} style={{ fontSize: '12px' }}>
                      {item.type === 'TV' ? '电视剧' : item.type === 'Movie' ? '电影' : item.type === 'Episode' ? '剧集' : item.type}
                    </Tag>
                    {item.number_of_episodes && item.number_of_episodes > 0 && (
                      <Tag color="cyan" style={{ fontSize: '12px' }}>
                        {item.local_number_of_episodes || 0}/{item.number_of_episodes}集
                      </Tag>
                    )}
                  </div>
                </div>
              </Card>
            ))}
          </div>

          {total > pageSize && (
            <div className="glass-card p-4 flex justify-center">
              <Pagination
                current={page}
                pageSize={pageSize}
                total={total}
                onChange={setPage}
                showSizeChanger={false}
                showTotal={(total) => `共 ${total} 条`}
              />
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default MediaDetail

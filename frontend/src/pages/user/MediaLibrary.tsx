import { useQuery } from '@tanstack/react-query'
import { Card, Empty, Spin, Pagination, FloatButton, Modal, Input, Badge, Tooltip } from 'antd'
import { VideoCameraOutlined, UpOutlined, CloseOutlined, DownOutlined, StarFilled, SearchOutlined, PlayCircleOutlined, FolderOutlined } from '@ant-design/icons'
import { useState, useEffect, useMemo } from 'react'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import { mediaApi } from '../../services/api'

const { Search } = Input

interface MediaDB {
  guid: string
  title: string
  posters: string[]
  category: string
  image_url: string
}

interface MediaDBSum {
  total: number
  movie: number
  tv: number
  video: number
  favorite: number
}

interface MediaItem {
  guid: string
  title: string
  original_title?: string
  type: string
  poster?: string
  vote_average?: string
  release_date?: string
  year?: number
  overview?: string
  number_of_episodes?: number
  local_number_of_episodes?: number
  duration?: number
  is_favorite?: number
  watched?: number
}

const MediaLibrary = () => {
  const { user } = useSelector((state: RootState) => state.auth)
  const [selectedMediaDB, setSelectedMediaDB] = useState<string | null>(null)
  const [selectedMedia, setSelectedMedia] = useState<MediaItem | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [page, setPage] = useState(1)
  const [showBackTop, setShowBackTop] = useState(false)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [isSearching, setIsSearching] = useState(false)
  const pageSize = 20

  // 获取媒体库列表
  const { data, isLoading, error } = useQuery({
    queryKey: ['mediaList'],
    queryFn: async () => {
      try {
        const response = await mediaApi.getList()
        return response.data.data
      } catch (err) {
        console.error('获取媒体库列表失败:', err)
        throw err
      }
    },
    retry: 1,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
    refetchOnMount: false,
  })

  // 获取选中媒体库的影片列表
  const { data: mediaItemsData, isLoading: isLoadingItems } = useQuery({
    queryKey: ['mediaItems', selectedMediaDB, page],
    queryFn: async () => {
      const response = await mediaApi.getMediaDBItems(selectedMediaDB!, { page, page_size: pageSize })
      return response.data.data
    },
    enabled: !!selectedMediaDB && !isSearching,
    staleTime: 3 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
    placeholderData: (previousData) => previousData,
  })

  // 获取选中媒体库的统计信息
  const { data: mediaDBSumData } = useQuery({
    queryKey: ['mediaDBSum', selectedMediaDB],
    queryFn: async () => {
      const response = await mediaApi.getMediaDBSum(selectedMediaDB!)
      return response.data.data as MediaDBSum
    },
    enabled: !!selectedMediaDB,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
  })

  // 搜索媒体
  const { data: searchData, isLoading: isLoadingSearch } = useQuery({
    queryKey: ['searchMedia', searchKeyword, page],
    queryFn: async () => {
      const response = await mediaApi.searchMedia({ keyword: searchKeyword, page, page_size: pageSize })
      return response.data.data
    },
    enabled: isSearching && searchKeyword.length > 0,
    staleTime: 1 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
  })

  const mediaList: MediaDB[] = data || []
  
  // 根据是否搜索模式选择数据源
  const currentData = isSearching ? searchData : mediaItemsData
  const mediaItems: MediaItem[] = currentData?.items || []
  const total = currentData?.total || 0
  const imageURL = currentData?.image_url || '/api/v1/image/'
  const isLoadingContent = isSearching ? isLoadingSearch : isLoadingItems

  // 处理搜索
  const handleSearch = (value: string) => {
    if (value.trim()) {
      setSearchKeyword(value.trim())
      setIsSearching(true)
      setSelectedMediaDB(null)
      setPage(1)
    }
  }

  // 清除搜索
  const clearSearch = () => {
    setSearchKeyword('')
    setIsSearching(false)
    setPage(1)
  }

  // 计算总统计
  const totalStats = useMemo(() => {
    if (!mediaList.length) return null
    return {
      libraries: mediaList.length,
    }
  }, [mediaList])

  const handleCardClick = (guid: string) => {
    // 如果点击已选中的媒体库，则收起
    if (selectedMediaDB === guid) {
      setSelectedMediaDB(null)
      setPage(1)
      window.scrollTo({ top: 0, behavior: 'smooth' })
      return
    }
    
    // 选择新的媒体库
    setSelectedMediaDB(guid)
    setPage(1)
    
    // 滚动到影片列表区域
    setTimeout(() => {
      document.getElementById('media-items-section')?.scrollIntoView({ 
        behavior: 'smooth',
        block: 'start'
      })
    }, 100)
  }

  const handleCloseMediaItems = () => {
    setSelectedMediaDB(null)
    setPage(1)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleMediaClick = (media: MediaItem) => {
    setSelectedMedia(media)
    setIsModalOpen(true)
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setSelectedMedia(null)
  }



  
  // 监听滚动，显示/隐藏返回顶部按钮
  useEffect(() => {
    const handleScroll = () => {
      setShowBackTop(window.scrollY > 300)
    }
    
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])
  
  // 构建带用户ID的图片URL
  const buildImageURL = (poster: string, baseURL?: string) => {
    const base = baseURL || imageURL
    const fullUrl = `${base}${poster}`
    if (!user?.id) return fullUrl
    // 检查URL是否已有查询参数
    const separator = fullUrl.includes('?') ? '&' : '?'
    return `${fullUrl}${separator}uid=${user.id}`
  }

  const buildMediaLibraryImageURL = (item: MediaDB, poster: string) => {
    const fullUrl = `${item.image_url}${poster}`
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

  if (error) {
    console.error('Media API Error:', error)
    return (
      <div className="glass-card p-6">
        <Empty 
          description={
            <div className="space-y-2">
              <p>请重新登录以获取媒体库访问权限</p>
              <p className="text-xs text-gray-400">
                {String(error)}
              </p>
            </div>
          } 
        />
      </div>
    )
  }

  if (!mediaList || mediaList.length === 0) {
    return (
      <div className="glass-card p-6">
        <Empty description="暂无可访问的媒体库" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 顶部搜索和统计 */}
      <div className="glass-card p-4">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          {/* 搜索框 */}
          <div className="flex-1 max-w-md">
            <Search
              placeholder="搜索电影、电视剧..."
              allowClear
              enterButton={<SearchOutlined />}
              size="large"
              onSearch={handleSearch}
              onChange={(e) => {
                if (!e.target.value) clearSearch()
              }}
              className="w-full"
            />
          </div>
          
          {/* 统计信息 */}
          {totalStats && (
            <div className="flex items-center gap-4 text-sm">
              <Tooltip title="媒体库数量">
                <div className="flex items-center gap-1.5 text-gray-600">
                  <FolderOutlined className="text-blue-500" />
                  <span>{totalStats.libraries} 个媒体库</span>
                </div>
              </Tooltip>
              {mediaDBSumData && (
                <>
                  <Tooltip title="电影数量">
                    <div className="flex items-center gap-1.5 text-gray-600">
                      <PlayCircleOutlined className="text-purple-500" />
                      <span>{mediaDBSumData.movie} 部电影</span>
                    </div>
                  </Tooltip>
                  <Tooltip title="电视剧数量">
                    <div className="flex items-center gap-1.5 text-gray-600">
                      <VideoCameraOutlined className="text-green-500" />
                      <span>{mediaDBSumData.tv} 部剧集</span>
                    </div>
                  </Tooltip>
                </>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 搜索结果提示 */}
      {isSearching && (
        <div className="flex items-center justify-between bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-xl px-5 py-3 shadow-sm">
          <div className="flex items-center gap-2">
            <SearchOutlined className="text-blue-500 text-lg" />
            <span className="text-gray-700">
              搜索 "<span className="font-semibold text-blue-600">{searchKeyword}</span>" 的结果，共 <span className="font-semibold text-blue-600">{total}</span> 个
            </span>
          </div>
          <button
            onClick={clearSearch}
            className="flex items-center gap-1.5 px-4 py-1.5 bg-white hover:bg-blue-500 text-blue-600 hover:text-white border border-blue-300 hover:border-blue-500 rounded-lg text-sm font-medium transition-all duration-200 shadow-sm hover:shadow"
          >
            <FolderOutlined />
            返回媒体库
          </button>
        </div>
      )}

      {/* 媒体库列表 - 非搜索模式显示 */}
      {!isSearching && (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-3">
        {mediaList.map((item) => (
          <Card
            key={item.guid}
            hoverable
            className={`glass-card overflow-hidden cursor-pointer transition-all relative ${
              selectedMediaDB === item.guid ? 'ring-2 ring-blue-500 shadow-lg' : ''
            }`}
            styles={{ body: { padding: 0 } }}
            onClick={() => handleCardClick(item.guid)}
            cover={
              item.posters && item.posters.length > 0 ? (
                <div className="relative aspect-video overflow-hidden bg-gradient-to-br from-gray-200 to-gray-300 group">
                  <img
                    src={buildMediaLibraryImageURL(item, item.posters[0])}
                    alt={item.title}
                    className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
                    loading="lazy"
                    onError={(e) => {
                      const target = e.target as HTMLImageElement
                      target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjIyNSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDAwIiBoZWlnaHQ9IjIyNSIgZmlsbD0iI2E1YjRmYyIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwiIGZvbnQtc2l6ZT0iMTgiIGZpbGw9IiNmZmZmZmYiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj7lm77niYfliqDovb3lpLHotKU8L3RleHQ+PC9zdmc+'
                      target.style.opacity = '0.5'
                    }}
                  />
                  
                  {/* 渐变遮罩 */}
                  <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent pointer-events-none" />
                  
                  {/* 底部信息栏 */}
                  <div className="absolute bottom-0 left-0 right-0 p-2">
                    {/* 媒体库名称 - 左下角 */}
                    <div className="text-white text-sm font-semibold drop-shadow-lg">
                      {item.title}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="aspect-video bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center">
                  <VideoCameraOutlined className="text-white text-2xl opacity-50" />
                </div>
              )
            }
          >
            {/* 选中指示箭头 */}
            {selectedMediaDB === item.guid && (
              <div className="absolute -bottom-3 left-1/2 transform -translate-x-1/2 z-20">
                <DownOutlined className="text-blue-500 text-2xl drop-shadow-lg animate-bounce" />
              </div>
            )}
          </Card>
        ))}
        </div>
      )}

      {/* 影片列表 - 选中媒体库或搜索模式 */}
      {(selectedMediaDB || isSearching) && (
        <div id="media-items-section" className="space-y-4">
          {/* 媒体库统计信息 */}
          {selectedMediaDB && mediaDBSumData && !isSearching && (
            <div className="flex items-center gap-4 text-sm bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg px-4 py-3 border border-blue-100">
              <Badge count={mediaDBSumData.total} showZero color="#6366f1" overflowCount={9999}>
                <span className="text-gray-600 pr-2">总计</span>
              </Badge>
              {mediaDBSumData.movie > 0 && (
                <Badge count={mediaDBSumData.movie} showZero color="#8b5cf6" overflowCount={9999}>
                  <span className="text-gray-600 pr-2">电影</span>
                </Badge>
              )}
              {mediaDBSumData.tv > 0 && (
                <Badge count={mediaDBSumData.tv} showZero color="#10b981" overflowCount={9999}>
                  <span className="text-gray-600 pr-2">剧集</span>
                </Badge>
              )}
            </div>
          )}

          {isLoadingContent ? (
            <div className="flex justify-center items-center h-64">
              <Spin size="large" tip={isSearching ? "搜索中..." : "加载中..."} spinning={true}>
                <div className="h-40" />
              </Spin>
            </div>
          ) : mediaItems.length === 0 ? (
            <div className="glass-card p-6">
              <Empty description={isSearching ? "未找到相关内容" : "该媒体库暂无内容"} />
            </div>
          ) : (
            <>
              <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8 gap-3">
                {mediaItems.map((item) => (
                  <Card
                    key={item.guid}
                    hoverable
                    className="glass-card overflow-hidden cursor-pointer"
                    styles={{ body: { padding: 0 } }}
                    onClick={() => handleMediaClick(item)}
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
                              target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjMwMCIgZmlsbD0iI2E1YjRmYyIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwiIGZvbnQtc2l6ZT0iMTgiIGZpbGw9IiNmZmZmZmYiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj7lm77niYfliqDovb3lpLHotKU8L3RleHQ+PC9zdmc+'
                              target.style.opacity = '0.5'
                            }}
                          />
                          
                          {/* 顶部标签区 */}
                          <div className="absolute top-1 left-1 right-1 flex justify-between items-start">
                            {/* 类型标签 - 左上角 */}
                            <div className="flex flex-col gap-1">
                              {item.type && (
                                <span className={`text-[10px] px-1.5 py-0.5 rounded font-medium ${
                                  item.type === 'Movie' 
                                    ? 'bg-purple-500/90 text-white' 
                                    : item.type === 'Series' 
                                    ? 'bg-blue-500/90 text-white'
                                    : 'bg-gray-500/90 text-white'
                                }`}>
                                  {item.type === 'Movie' ? '电影' : item.type === 'Series' ? '剧集' : item.type}
                                </span>
                              )}
                              {item.year && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-black/50 text-white/90">
                                  {item.year}
                                </span>
                              )}
                            </div>
                            
                            {/* 评分和状态 - 右上角 */}
                            <div className="flex flex-col items-end gap-1">
                              {item.vote_average && parseFloat(item.vote_average) > 0 && (
                                <div className="flex items-center gap-1 text-yellow-400 bg-black/70 rounded px-1.5 py-0.5 backdrop-blur-sm">
                                  <StarFilled style={{ fontSize: '10px' }} />
                                  <span className="font-semibold text-xs">{parseFloat(item.vote_average).toFixed(1)}</span>
                                </div>
                              )}
                              {item.is_favorite === 1 && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-red-500/90 text-white">❤️</span>
                              )}
                              {item.watched === 1 && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-green-500/90 text-white">✓</span>
                              )}
                            </div>
                          </div>
                          
                          {/* 渐变遮罩 */}
                          <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent pointer-events-none" />
                          
                          {/* 底部信息栏 */}
                          <div className="absolute bottom-0 left-0 right-0 flex items-end justify-between p-1.5 gap-1">
                            {/* 标题 - 左下角 */}
                            <div className="flex-1 text-white text-xs font-medium line-clamp-2 drop-shadow-lg">
                              {item.title}
                            </div>
                            
                            {/* 集数 - 右下角 */}
                            {item.number_of_episodes && item.number_of_episodes > 0 && (
                              <div className="flex-shrink-0 bg-black/70 text-white rounded px-1.5 py-0.5 backdrop-blur-sm text-xs font-medium whitespace-nowrap">
                                {item.local_number_of_episodes || 0}/{item.number_of_episodes}
                              </div>
                            )}
                          </div>
                        </div>
                      ) : (
                        <div className="aspect-[2/3] bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center">
                          <VideoCameraOutlined className="text-white text-3xl opacity-30" />
                        </div>
                      )
                    }
                  >
                  </Card>
                ))}
              </div>

              {/* 分页 */}
              {total > pageSize && (
                <div className="flex justify-center mt-6">
                  <Pagination
                    current={page}
                    pageSize={pageSize}
                    total={total}
                    onChange={(newPage) => {
                      setPage(newPage)
                      document.getElementById('media-items-section')?.scrollIntoView({ 
                        behavior: 'smooth',
                        block: 'start'
                      })
                    }}
                    showSizeChanger={false}
                    showTotal={(total) => `共 ${total} 个影片`}
                  />
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* 浮动按钮 */}
      {showBackTop && (
        <FloatButton.Group shape="circle">
          <FloatButton 
            icon={<UpOutlined />} 
            tooltip="返回顶部"
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          />
          {selectedMediaDB && (
            <FloatButton 
              icon={<CloseOutlined />} 
              tooltip="收起影片列表"
              onClick={handleCloseMediaItems}
              type="primary"
            />
          )}
        </FloatButton.Group>
      )}

      {/* 影片详情弹窗 */}
      <Modal
        title={null}
        open={isModalOpen}
        onCancel={handleCloseModal}
        footer={null}
        width={1000}
        centered
        styles={{ body: { padding: 0, maxHeight: '85vh', overflowY: 'auto' } }}
      >
        {selectedMedia ? (
          <div>
            {/* 背景和海报区域 */}
            <div className="relative">
              {selectedMedia.poster && (
                <div className="relative h-80 overflow-hidden">
                  <img
                    src={buildImageURL(selectedMedia.poster)}
                    alt={selectedMedia.title}
                    className="w-full h-full object-cover blur-md scale-110"
                  />
                  <div className="absolute inset-0 bg-gradient-to-t from-black via-black/80 to-black/40" />
                </div>
              )}
              <div className="absolute inset-0 p-6 flex items-start pt-6">
                <div className="flex gap-6 w-full items-start">
                  {selectedMedia.poster && (
                    <img
                      src={buildImageURL(selectedMedia.poster)}
                      alt={selectedMedia.title}
                      className="w-36 h-48 object-cover rounded-xl shadow-2xl flex-shrink-0"
                    />
                  )}
                  <div className="flex-1 space-y-2.5 text-white">
                    <div>
                      <h2 className="text-2xl font-bold drop-shadow-lg mb-1.5 text-white leading-tight">
                        {selectedMedia.title}
                      </h2>
                      {selectedMedia.release_date && (
                        <p className="text-[11px] text-gray-200 bg-black/20 backdrop-blur-sm px-2.5 py-0.5 rounded-full inline-block border border-white/20">
                          📅 {new Date(selectedMedia.release_date).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })}
                        </p>
                      )}
                    </div>
                    <div className="flex flex-wrap items-center gap-1.5 text-xs">
                      {selectedMedia.vote_average && parseFloat(selectedMedia.vote_average) > 0 ? (
                        <span className="bg-gradient-to-r from-yellow-500 to-orange-500 text-white px-3 py-1 rounded-full font-bold shadow-md text-xs">
                          ⭐ {parseFloat(selectedMedia.vote_average).toFixed(1)}
                        </span>
                      ) : null}
                      {selectedMedia.type ? (
                        <span className="bg-gradient-to-r from-blue-500 to-cyan-500 text-white px-3 py-1 rounded-full font-medium shadow-md text-xs">
                          {selectedMedia.type === 'TV' ? '📺 电视剧' : selectedMedia.type === 'Movie' ? '🎬 电影' : selectedMedia.type}
                        </span>
                      ) : null}
                      {selectedMedia.number_of_episodes && selectedMedia.number_of_episodes > 0 ? (
                        <span className="bg-gradient-to-r from-purple-500 to-pink-500 text-white px-3 py-1 rounded-full font-medium shadow-md text-xs">
                          共{selectedMedia.number_of_episodes}集
                        </span>
                      ) : null}
                      {selectedMedia.local_number_of_episodes && selectedMedia.local_number_of_episodes > 0 ? (
                        <span className="bg-gradient-to-r from-green-500 to-emerald-500 text-white px-3 py-1 rounded-full font-medium shadow-md text-xs">
                          ✓ 已更新{selectedMedia.local_number_of_episodes}集
                        </span>
                      ) : null}
                      {selectedMedia.duration && selectedMedia.duration > 60 ? (
                        <span className="bg-white/30 backdrop-blur-sm text-white px-3 py-1 rounded-full font-medium shadow-md text-xs">
                          ⏱ {Math.floor(selectedMedia.duration / 60)}分钟
                        </span>
                      ) : null}
                      {selectedMedia.is_favorite === 1 ? (
                        <span className="bg-gradient-to-r from-red-500 to-pink-500 text-white px-3 py-1 rounded-full font-medium shadow-md text-xs">❤️ 已收藏</span>
                      ) : null}
                      {selectedMedia.watched === 1 ? (
                        <span className="bg-gradient-to-r from-teal-500 to-green-500 text-white px-3 py-1 rounded-full font-medium shadow-md text-xs">✓ 已观看</span>
                      ) : null}
                    </div>
                    
                    {/* 剧情简介 */}
                    {selectedMedia.overview && (
                      <div className="bg-black/20 backdrop-blur-sm rounded-lg p-3 border border-white/10 max-h-32 overflow-hidden">
                        <h3 className="text-xs font-bold mb-1.5 text-white flex items-center gap-1">
                          <span className="text-sm">📖</span>
                          剧情简介
                        </h3>
                        <p className="text-gray-100 text-[11px] leading-relaxed line-clamp-5">
                          {selectedMedia.overview}
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>


          </div>
        ) : (
          <div className="p-6">
            <Empty description="无法加载影片详情" />
          </div>
        )}
      </Modal>
    </div>
  )
}

export default MediaLibrary

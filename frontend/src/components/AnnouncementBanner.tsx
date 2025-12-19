import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { CloseOutlined, DownOutlined, FireOutlined, BellOutlined, RocketOutlined } from '@ant-design/icons'
import { announcementApi } from '../services/api'
import dayjs from 'dayjs'

interface Announcement {
  id: number
  title: string
  content: string
  type: number
  is_top: boolean
  created_at: string
}

const typeConfig: Record<number, { label: string; color: string; icon: React.ReactNode; bg: string }> = {
  1: { label: '通知', color: '#1890ff', icon: <BellOutlined />, bg: 'from-blue-500 to-cyan-500' },
  2: { label: '活动', color: '#fa8c16', icon: <FireOutlined />, bg: 'from-orange-500 to-red-500' },
  3: { label: '更新', color: '#52c41a', icon: <RocketOutlined />, bg: 'from-green-500 to-teal-500' },
}

const AnnouncementBanner = () => {
  const [visible, setVisible] = useState(false)
  const [selectedAnn, setSelectedAnn] = useState<Announcement | null>(null)
  const [userInteracted, setUserInteracted] = useState(false)

  const { data: announcements } = useQuery({
    queryKey: ['userAnnouncements'],
    queryFn: async () => {
      const response = await announcementApi.getPublished()
      return response.data.data as Announcement[]
    },
    staleTime: 5 * 60 * 1000,
  })

  // 有公告时自动显示
  useEffect(() => {
    if (announcements?.length) {
      setVisible(true)
      setUserInteracted(false)
    }
  }, [announcements?.length])

  // 自动隐藏（仅在用户未交互时）
  useEffect(() => {
    if (visible && !userInteracted) {
      const timer = setTimeout(() => setVisible(false), 3000)
      return () => clearTimeout(timer)
    }
  }, [visible, userInteracted])

  if (!announcements?.length || !visible) return null

  return (
    <div className="fixed top-4 right-4 z-50">
      {/* 公告面板 */}
      <div className="w-80 bg-white/95 backdrop-blur-xl rounded-xl overflow-hidden shadow-[0_8px_32px_rgba(0,0,0,0.12)] border border-gray-100">
        {/* 头部 */}
        <div className="bg-gradient-to-r from-violet-500 via-purple-500 to-fuchsia-500 text-white px-4 py-3">
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-sm">系统公告</div>
              <div className="text-white/70 text-xs">
                共 {announcements.length} 条公告
              </div>
            </div>
            <button 
              onClick={() => setVisible(false)} 
              className="w-6 h-6 flex items-center justify-center hover:bg-white/20 rounded-full transition-colors"
            >
              <CloseOutlined className="text-xs" />
            </button>
          </div>
        </div>
        
        {/* 公告列表 */}
        <div className="max-h-64 overflow-y-auto">
          {announcements.map((ann, index) => {
            const config = typeConfig[ann.type] || typeConfig[1]
            const isSelected = selectedAnn?.id === ann.id
            
            return (
              <div 
                key={ann.id}
                className={`
                  cursor-pointer transition-all duration-200
                  ${isSelected ? 'bg-purple-50' : 'hover:bg-gray-50'}
                  ${index !== announcements.length - 1 ? 'border-b border-gray-100' : ''}
                `}
                onClick={() => {
                  setUserInteracted(true)
                  setSelectedAnn(isSelected ? null : ann)
                }}
              >
                <div className="p-3 pl-4">
                  <div className="flex items-start gap-2">
                    <div className={`w-7 h-7 rounded-lg flex items-center justify-center text-white text-xs flex-shrink-0 bg-gradient-to-br ${config.bg}`}>
                      {config.icon}
                    </div>
                    
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5 mb-0.5">
                        <span className="font-medium text-gray-800 text-sm truncate">{ann.title}</span>
                        {ann.is_top && (
                          <span className="text-[9px] bg-red-500 text-white px-1 py-0.5 rounded leading-none">TOP</span>
                        )}
                      </div>
                      <div className="flex items-center gap-1.5 text-[11px] text-gray-400">
                        <span style={{ color: config.color }}>{config.label}</span>
                        <span>·</span>
                        <span>{dayjs(ann.created_at).format('MM-DD HH:mm')}</span>
                      </div>
                      
                      {isSelected && (
                        <div className="mt-2 pt-2 border-t border-dashed border-gray-200 text-xs text-gray-600 leading-relaxed whitespace-pre-wrap">
                          {ann.content}
                        </div>
                      )}
                    </div>
                    
                    <DownOutlined className={`text-gray-300 text-[10px] transition-transform duration-200 mt-1 ${isSelected ? 'rotate-180' : ''}`} />
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

export default AnnouncementBanner

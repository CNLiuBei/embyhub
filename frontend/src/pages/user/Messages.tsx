import { useState, useEffect, useRef, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLocation } from 'react-router-dom'
import {
  Card,
  Avatar,
  Badge,
  Input,
  Button,
  Spin,
  App,
  Modal,
  Select,
  Tooltip,
  Dropdown,
  Tabs,
  Popconfirm,
} from 'antd'
import type { MenuProps } from 'antd'
import {
  SendOutlined,
  ArrowLeftOutlined,
  PlusOutlined,
  MessageOutlined,
  UserOutlined,
  SoundOutlined,
  StopOutlined,
  UndoOutlined,
  TeamOutlined,
  UserDeleteOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { pmApi, followApi, blacklistApi } from '../../services/api'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import UserProfileCard from '../../components/UserProfileCard'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

interface Conversation {
  id: number
  user1_id: string
  user2_id: string
  last_message_at: string
  other_user?: {
    id: string
    nickname: string
    avatar: string
  }
  last_message?: {
    content: string
  }
  unread_count: number
  is_muted: boolean
}

interface PrivateMessage {
  id: number
  from_user_id: string
  to_user_id: string
  content: string
  status: number
  created_at: string
  from_user?: {
    id: string
    nickname: string
    avatar: string
  }
}

const MESSAGE_STATUS_RECALLED = 3

interface SearchUser {
  id: string
  username: string
  nickname: string
  avatar: string
}

interface FollowUser {
  id: string
  username: string
  nickname: string
  avatar: string
  bio?: string
}

interface BlacklistItem {
  id: number
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

const Messages = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const location = useLocation()
  const currentUser = useSelector((state: RootState) => state.auth.user)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const [activeTab, setActiveTab] = useState('messages')
  const [followTab, setFollowTab] = useState<'followings' | 'followers'>('followings')
  const [selectedUser, setSelectedUser] = useState<{
    id: string
    nickname: string
    avatar: string
  } | null>(null)
  const [messageContent, setMessageContent] = useState('')
  const [messagePage, setMessagePage] = useState(1)
  const [newConversationOpen, setNewConversationOpen] = useState(false)
  const [searchUsers, setSearchUsers] = useState<SearchUser[]>([])
  const [searching, setSearching] = useState(false)

  // 从其他页面跳转过来时，自动选择目标用户
  useEffect(() => {
    const state = location.state as { targetUser?: { id: string; nickname: string; avatar: string } } | null
    if (state?.targetUser) {
      setSelectedUser({
        id: state.targetUser.id,
        nickname: state.targetUser.nickname,
        avatar: state.targetUser.avatar,
      })
      setActiveTab('messages')
      window.history.replaceState({}, document.title)
    }
  }, [location.state])

  // ============= 私信相关 =============
  const { data: conversationsData, isLoading: conversationsLoading } = useQuery({
    queryKey: ['conversations'],
    queryFn: async () => {
      const res = await pmApi.getConversations({ page: 1, page_size: 50 })
      return res.data.data as { list: Conversation[]; total: number }
    },
  })

  const { data: messagesData, isLoading: messagesLoading, isFetching: messagesFetching } = useQuery({
    queryKey: ['messages', selectedUser?.id, messagePage],
    queryFn: async () => {
      if (!selectedUser) return null
      const res = await pmApi.getMessages(selectedUser.id, { page: messagePage, page_size: 50 })
      return res.data.data as { list: PrivateMessage[]; total: number }
    },
    enabled: !!selectedUser,
  })

  // ============= 关注相关 =============
  const { data: followingsData, isLoading: followingsLoading } = useQuery({
    queryKey: ['followings'],
    queryFn: async () => {
      const res = await followApi.getFollowings()
      return res.data.data as { list: FollowUser[]; total: number }
    },
    enabled: activeTab === 'follow',
  })

  const { data: followersData, isLoading: followersLoading } = useQuery({
    queryKey: ['followers'],
    queryFn: async () => {
      const res = await followApi.getFollowers()
      return res.data.data as { list: FollowUser[]; total: number }
    },
    enabled: activeTab === 'follow',
  })

  const { data: followStats } = useQuery({
    queryKey: ['followStats'],
    queryFn: async () => {
      const res = await followApi.getFollowStats()
      return res.data.data as { followings: number; followers: number }
    },
    enabled: activeTab === 'follow',
  })

  // ============= 黑名单相关 =============
  const { data: blacklistData, isLoading: blacklistLoading } = useQuery({
    queryKey: ['blacklist'],
    queryFn: async () => {
      const res = await blacklistApi.getBlacklist({ page: 1, page_size: 50 })
      return res.data.data as { list: BlacklistItem[]; total: number }
    },
    enabled: activeTab === 'blacklist',
  })

  // ============= Mutations =============
  const markAsReadMutation = useMutation({
    mutationFn: (userId: string) => pmApi.markAsRead(userId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['conversations'] }),
  })

  const sendMutation = useMutation({
    mutationFn: (data: { to_user_id: string; content: string }) => pmApi.sendMessage(data),
    onSuccess: () => {
      setMessageContent('')
      queryClient.invalidateQueries({ queryKey: ['messages', selectedUser?.id] })
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '发送失败')
    },
  })

  const recallMutation = useMutation({
    mutationFn: (messageId: number) => pmApi.recallMessage(messageId),
    onSuccess: () => {
      message.success('消息已撤回')
      queryClient.invalidateQueries({ queryKey: ['messages', selectedUser?.id] })
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '撤回失败')
    },
  })

  const muteMutation = useMutation({
    mutationFn: (userId: string) => pmApi.muteConversation(userId),
    onSuccess: (res) => {
      message.success(res.data.data?.muted ? '已静音' : '已取消静音')
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    },
  })

  const unfollowMutation = useMutation({
    mutationFn: (userId: string) => followApi.toggleFollow(userId),
    onSuccess: () => {
      message.success('已取消关注')
      queryClient.invalidateQueries({ queryKey: ['followings'] })
      queryClient.invalidateQueries({ queryKey: ['followStats'] })
    },
  })

  const unblockMutation = useMutation({
    mutationFn: (userId: string) => blacklistApi.unblockUser(userId),
    onSuccess: () => {
      message.success('已取消拉黑')
      queryClient.invalidateQueries({ queryKey: ['blacklist'] })
    },
  })

  // ============= Handlers =============
  const handleSelectConversation = useCallback(
    (user: { id: string; nickname: string; avatar: string }, unreadCount: number) => {
      setSelectedUser(user)
      setMessagePage(1)
      if (unreadCount > 0) markAsReadMutation.mutate(user.id)
    },
    [markAsReadMutation]
  )

  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>()
  const handleSearchUsers = useCallback((keyword: string) => {
    if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current)
    if (!keyword.trim()) { setSearchUsers([]); return }
    setSearching(true)
    searchTimeoutRef.current = setTimeout(async () => {
      try {
        const res = await pmApi.searchUsers(keyword)
        setSearchUsers(res.data.data || [])
      } catch { setSearchUsers([]) }
      finally { setSearching(false) }
    }, 300)
  }, [])

  const handleSelectUser = (user: SearchUser) => {
    setSelectedUser({ id: user.id, nickname: user.nickname || user.username, avatar: user.avatar })
    setNewConversationOpen(false)
    setSearchUsers([])
    setMessagePage(1)
  }

  const handleSendMessage = (user: FollowUser) => {
    setSelectedUser({ id: user.id, nickname: user.nickname || user.username, avatar: user.avatar })
    setActiveTab('messages')
  }

  useEffect(() => {
    if (!messagesFetching) messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messagesData, messagesFetching])

  const handleSend = () => {
    if (!messageContent.trim() || !selectedUser) return
    sendMutation.mutate({ to_user_id: selectedUser.id, content: messageContent })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() }
  }

  const sortedMessages = messagesData?.list?.slice().reverse() || []
  const totalUnread = conversationsData?.list?.reduce((sum, conv) => sum + (conv.unread_count || 0), 0) || 0

  // ============= 渲染私信列表 =============
  const renderConversationList = () => (
    <div className="flex-1 overflow-auto">
      <Spin spinning={conversationsLoading}>
        {conversationsData?.list?.length ? (
          <div className="divide-y divide-gray-50">
            {conversationsData.list.map((conv) => {
              const isSelected = selectedUser?.id === conv.other_user?.id
              const hasUnread = conv.unread_count > 0
              const convMenuItems: MenuProps['items'] = [{
                key: 'mute',
                icon: conv.is_muted ? <SoundOutlined /> : <StopOutlined />,
                label: conv.is_muted ? '取消静音' : '静音会话',
                onClick: (e) => { e.domEvent.stopPropagation(); if (conv.other_user) muteMutation.mutate(conv.other_user.id) },
              }]
              return (
                <Dropdown key={conv.id} menu={{ items: convMenuItems }} trigger={['contextMenu']}>
                  <div
                    className={`relative flex items-center gap-3 px-4 py-3 cursor-pointer transition-all duration-200 ${isSelected ? 'bg-gradient-to-r from-blue-50 to-blue-50/50' : 'hover:bg-gray-50/80'}`}
                    onClick={() => conv.other_user && handleSelectConversation(conv.other_user, conv.unread_count)}
                  >
                    {isSelected && <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-8 bg-blue-500 rounded-r" />}
                    <div className="relative flex-shrink-0" onClick={(e) => e.stopPropagation()}>
                      {conv.other_user ? (
                        <UserProfileCard user={conv.other_user} placement="right">
                          <Avatar src={conv.other_user.avatar} size={48} icon={<UserOutlined />} className={`cursor-pointer hover:ring-2 hover:ring-blue-300 transition-all ${isSelected ? 'ring-2 ring-blue-200' : ''}`}>
                            {conv.other_user.nickname?.[0]}
                          </Avatar>
                        </UserProfileCard>
                      ) : (
                        <Avatar size={48} icon={<UserOutlined />} />
                      )}
                      {hasUnread && <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] flex items-center justify-center bg-red-500 text-white text-xs rounded-full px-1">{conv.unread_count > 99 ? '99+' : conv.unread_count}</span>}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex justify-between items-center mb-1">
                        <span className={`truncate flex items-center gap-1 ${hasUnread ? 'font-semibold text-gray-900' : 'font-medium text-gray-700'}`}>
                          {conv.other_user?.nickname || '未知用户'}
                          {conv.is_muted && <StopOutlined className="text-gray-400 text-xs" />}
                        </span>
                        <span className="text-gray-400 text-xs flex-shrink-0 ml-2">{dayjs(conv.last_message_at).fromNow()}</span>
                      </div>
                      <div className={`text-sm truncate ${hasUnread ? 'text-gray-800' : 'text-gray-500'}`}>{conv.last_message?.content || '暂无消息'}</div>
                    </div>
                  </div>
                </Dropdown>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-16 text-gray-400">
            <MessageOutlined className="text-4xl text-gray-300 mb-4" />
            <p className="mb-3">暂无私信</p>
            <Button type="primary" ghost size="small" icon={<PlusOutlined />} onClick={() => setNewConversationOpen(true)}>发起新对话</Button>
          </div>
        )}
      </Spin>
    </div>
  )

  // ============= 渲染关注列表 =============
  const renderFollowList = () => {
    const list = followTab === 'followings' ? followingsData?.list : followersData?.list
    const loading = followTab === 'followings' ? followingsLoading : followersLoading
    return (
      <div className="flex-1 overflow-auto flex flex-col">
        {/* 标签切换 */}
        <div className="flex bg-gray-50/50 mx-3 mt-3 rounded-lg p-1">
          <button
            className={`flex-1 py-2 text-sm rounded-md transition-all duration-200 ${
              followTab === 'followings'
                ? 'bg-white text-blue-600 shadow-sm font-medium'
                : 'text-gray-500 hover:text-gray-700'
            }`}
            onClick={() => setFollowTab('followings')}
          >
            <TeamOutlined className="mr-1" />
            关注 <span className="text-xs opacity-70">({followStats?.followings || 0})</span>
          </button>
          <button
            className={`flex-1 py-2 text-sm rounded-md transition-all duration-200 ${
              followTab === 'followers'
                ? 'bg-white text-blue-600 shadow-sm font-medium'
                : 'text-gray-500 hover:text-gray-700'
            }`}
            onClick={() => setFollowTab('followers')}
          >
            <UserOutlined className="mr-1" />
            粉丝 <span className="text-xs opacity-70">({followStats?.followers || 0})</span>
          </button>
        </div>

        {/* 列表 */}
        <div className="flex-1 overflow-auto mt-2">
          <Spin spinning={loading}>
            {list?.length ? (
              <div className="px-2 pb-2 space-y-1">
                {list.map((user) => (
                  <div
                    key={user.id}
                    className="group flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-gradient-to-r hover:from-blue-50/80 hover:to-purple-50/50 transition-all duration-200"
                  >
                    <div className="relative">
                      <UserProfileCard user={user} placement="right">
                        <Avatar
                          src={user.avatar}
                          size={48}
                          icon={<UserOutlined />}
                          className="ring-2 ring-white shadow-sm cursor-pointer hover:ring-blue-300 transition-all"
                        >
                          {user.nickname?.[0]}
                        </Avatar>
                      </UserProfileCard>
                      <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-400 rounded-full border-2 border-white" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-gray-800 truncate">{user.nickname || user.username}</div>
                      <div className="text-gray-400 text-xs truncate mt-0.5">
                        {user.bio || (followTab === 'followings' ? '你关注了TA' : 'TA关注了你')}
                      </div>
                    </div>
                    <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Tooltip title="发私信">
                        <Button
                          type="primary"
                          ghost
                          size="small"
                          icon={<MessageOutlined />}
                          onClick={(e) => { e.stopPropagation(); handleSendMessage(user) }}
                          className="border-blue-200"
                        />
                      </Tooltip>
                      {followTab === 'followings' && (
                        <Popconfirm
                          title="确定取消关注？"
                          description="取消后可以重新关注"
                          onConfirm={() => unfollowMutation.mutate(user.id)}
                          okText="确定"
                          cancelText="取消"
                        >
                          <Button type="text" size="small" danger icon={<UserDeleteOutlined />} />
                        </Popconfirm>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-16 text-gray-400">
                <div className="w-20 h-20 rounded-full bg-gradient-to-br from-gray-100 to-gray-50 flex items-center justify-center mb-4">
                  <TeamOutlined className="text-3xl text-gray-300" />
                </div>
                <p className="text-gray-500 mb-1">{followTab === 'followings' ? '还没有关注任何人' : '还没有粉丝'}</p>
                <p className="text-xs text-gray-400">{followTab === 'followings' ? '去论坛发现有趣的人吧' : '分享精彩内容吸引更多关注'}</p>
              </div>
            )}
          </Spin>
        </div>
      </div>
    )
  }

  // ============= 渲染黑名单 =============
  const renderBlacklist = () => (
    <div className="flex-1 overflow-auto">
      <Spin spinning={blacklistLoading}>
        {blacklistData?.list?.length ? (
          <div className="px-2 py-2 space-y-1">
            {blacklistData.list.map((item) => (
              <div
                key={item.id}
                className="group flex items-center gap-3 px-3 py-3 rounded-xl bg-red-50/30 hover:bg-red-50/60 transition-all duration-200"
              >
                <div className="relative">
                  {item.blocked_user ? (
                    <UserProfileCard user={item.blocked_user} placement="right">
                      <Avatar
                        src={item.blocked_user.avatar}
                        size={48}
                        icon={<UserOutlined />}
                        className="ring-2 ring-red-100 grayscale opacity-70 cursor-pointer hover:opacity-100 hover:grayscale-0 transition-all"
                      >
                        {item.blocked_user.nickname?.[0]}
                      </Avatar>
                    </UserProfileCard>
                  ) : (
                    <Avatar size={48} icon={<UserOutlined />} className="ring-2 ring-red-100 grayscale opacity-70" />
                  )}
                  <div className="absolute -bottom-0.5 -right-0.5 w-5 h-5 bg-red-100 rounded-full flex items-center justify-center">
                    <StopOutlined className="text-red-400 text-xs" />
                  </div>
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-gray-600 truncate">
                    {item.blocked_user?.nickname || item.blocked_user?.username || '未知用户'}
                  </div>
                  <div className="text-gray-400 text-xs mt-0.5">
                    {dayjs(item.created_at).format('YYYY年MM月DD日')} 加入黑名单
                  </div>
                </div>
                <Popconfirm
                  title="确定移出黑名单？"
                  description="移出后对方可以再次给你发消息"
                  onConfirm={() => unblockMutation.mutate(item.blocked_id)}
                  okText="确定"
                  cancelText="取消"
                  okButtonProps={{ danger: true }}
                >
                  <Button
                    size="small"
                    className="opacity-60 group-hover:opacity-100 transition-opacity"
                    icon={<DeleteOutlined />}
                  >
                    移除
                  </Button>
                </Popconfirm>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-16 text-gray-400">
            <div className="w-20 h-20 rounded-full bg-gradient-to-br from-green-50 to-emerald-50 flex items-center justify-center mb-4">
              <StopOutlined className="text-3xl text-green-300" />
            </div>
            <p className="text-gray-500 mb-1">黑名单为空</p>
            <p className="text-xs text-gray-400">这里很干净，没有被拉黑的用户</p>
          </div>
        )}
      </Spin>
    </div>
  )

  return (
    <div className="h-[calc(100vh-180px)] flex gap-4">
      {/* 左侧面板 */}
      <div className="w-80 flex-shrink-0 flex flex-col glass-card rounded-xl overflow-hidden">
        <div className="px-4 py-2 border-b border-gray-100 bg-white/80 backdrop-blur">
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            size="small"
            items={[
              { key: 'messages', label: <span className="flex items-center gap-1"><MessageOutlined />私信{totalUnread > 0 && <Badge count={totalUnread} size="small" />}</span> },
              { key: 'follow', label: <span className="flex items-center gap-1"><TeamOutlined />关注</span> },
              { key: 'blacklist', label: <span className="flex items-center gap-1"><StopOutlined />黑名单</span> },
            ]}
          />
        </div>
        {activeTab === 'messages' && (
          <div className="px-4 py-2 border-b border-gray-100">
            <Button type="primary" size="small" icon={<PlusOutlined />} onClick={() => setNewConversationOpen(true)} block>发起新私信</Button>
          </div>
        )}
        {activeTab === 'messages' && renderConversationList()}
        {activeTab === 'follow' && renderFollowList()}
        {activeTab === 'blacklist' && renderBlacklist()}
      </div>

      {/* 聊天区域 */}
      <Card
        className="glass-card flex-1 flex flex-col"
        title={selectedUser ? (
          <div className="flex items-center gap-3">
            <Button type="text" icon={<ArrowLeftOutlined />} className="md:hidden -ml-2" onClick={() => setSelectedUser(null)} />
            <Avatar src={selectedUser.avatar} size={36}>{selectedUser.nickname?.[0]}</Avatar>
            <div className="font-medium">{selectedUser.nickname}</div>
          </div>
        ) : <div className="text-gray-400">选择一个会话开始聊天</div>}
        styles={{ body: { flex: 1, display: 'flex', flexDirection: 'column', padding: 0, overflow: 'hidden', minHeight: 0 } }}
      >
        {selectedUser ? (
          <>
            <div className="flex-1 overflow-auto p-4 space-y-3 bg-gray-50/50">
              <Spin spinning={messagesLoading}>
                {sortedMessages.length ? sortedMessages.map((msg) => {
                  const isMe = msg.from_user_id === currentUser?.id
                  const isRecalled = msg.status === MESSAGE_STATUS_RECALLED
                  const canRecall = isMe && !isRecalled && dayjs().diff(dayjs(msg.created_at), 'minute') < 5
                  return (
                    <div key={msg.id} className={`group flex gap-2 ${isMe ? 'flex-row-reverse' : ''}`}>
                      <Avatar src={msg.from_user?.avatar} size={32} className="flex-shrink-0">{msg.from_user?.nickname?.[0]}</Avatar>
                      <div className={`max-w-[70%] ${isMe ? 'text-right' : ''}`}>
                        <div className={`flex items-center gap-1 ${isMe ? 'flex-row-reverse' : ''}`}>
                          <div className={`inline-block px-3 py-2 rounded-2xl shadow-sm ${isRecalled ? 'bg-gray-100 text-gray-400 italic' : isMe ? 'bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-br-md' : 'bg-white text-gray-800 rounded-bl-md'}`}>
                            <p className="whitespace-pre-wrap break-words">{msg.content}</p>
                          </div>
                          {canRecall && (
                            <Tooltip title="撤回">
                              <Button type="text" size="small" icon={<UndoOutlined />} className="opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-red-500" onClick={() => recallMutation.mutate(msg.id)} loading={recallMutation.isPending} />
                            </Tooltip>
                          )}
                        </div>
                        <div className="text-gray-400 text-xs mt-1 px-1">{dayjs(msg.created_at).format('HH:mm')}</div>
                      </div>
                    </div>
                  )
                }) : <div className="flex flex-col items-center justify-center h-full text-gray-400"><MessageOutlined className="text-4xl mb-2" /><p>开始你们的对话吧</p></div>}
                <div ref={messagesEndRef} />
              </Spin>
            </div>
            <div className="p-3 border-t border-gray-100 bg-white">
              <div className="flex gap-2 items-end">
                <Input.TextArea value={messageContent} onChange={(e) => setMessageContent(e.target.value)} onKeyDown={handleKeyDown} placeholder="输入消息，Enter发送..." autoSize={{ minRows: 1, maxRows: 4 }} maxLength={1000} className="flex-1" />
                <Button type="primary" icon={<SendOutlined />} onClick={handleSend} loading={sendMutation.isPending} disabled={!messageContent.trim()} className="h-8">发送</Button>
              </div>
              <div className="text-gray-400 text-xs mt-1 text-right">{messageContent.length}/1000</div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-gray-400 bg-gray-50/30">
            <MessageOutlined className="text-6xl mb-4 text-gray-300" />
            <p className="text-lg mb-2">选择一个会话</p>
            <p className="text-sm">或发起新的私信对话</p>
          </div>
        )}
      </Card>

      {/* 新建会话弹窗 */}
      <Modal title={<span><PlusOutlined /> 发起新私信</span>} open={newConversationOpen} onCancel={() => { setNewConversationOpen(false); setSearchUsers([]) }} footer={null} width={420}>
        <div className="py-4">
          <Select showSearch placeholder="搜索用户名或昵称..." filterOption={false} onSearch={handleSearchUsers} loading={searching}
            notFoundContent={searching ? <Spin size="small" /> : <div className="py-4 text-center text-gray-400">输入关键词搜索用户</div>}
            style={{ width: '100%' }} size="large"
            onSelect={(_, option) => { const user = searchUsers.find((u) => u.id === option.value); if (user) handleSelectUser(user) }}
            options={searchUsers.map((user) => ({
              value: user.id,
              label: <div className="flex items-center gap-3 py-2"><Avatar src={user.avatar} size={40}>{(user.nickname || user.username)?.[0]}</Avatar><div><div className="font-medium">{user.nickname || user.username}</div><div className="text-gray-400 text-xs">@{user.username}</div></div></div>,
            }))}
          />
        </div>
      </Modal>
    </div>
  )
}

export default Messages

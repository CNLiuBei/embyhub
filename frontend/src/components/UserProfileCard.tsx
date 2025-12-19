import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Popover, Spin, App, Tooltip } from 'antd'
import {
  UserOutlined,
  MessageOutlined,
  UserAddOutlined,
  UserDeleteOutlined,
  ManOutlined,
  WomanOutlined,
  StopOutlined,
} from '@ant-design/icons'
import { followApi, pmApi, blacklistApi } from '../services/api'
import { useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { RootState } from '../store'

interface UserInfo {
  id: string
  username?: string
  nickname?: string
  avatar?: string
  bio?: string
  gender?: number
}

interface UserProfileCardProps {
  user: UserInfo
  children: React.ReactNode
  placement?: 'top' | 'bottom' | 'left' | 'right' | 'topLeft' | 'topRight' | 'bottomLeft' | 'bottomRight'
}

const UserProfileCard = ({ user, children, placement = 'top' }: UserProfileCardProps) => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const currentUser = useSelector((state: RootState) => state.auth.user)
  const [open, setOpen] = useState(false)

  const isMe = currentUser?.id === user.id

  // 获取关注统计
  const { data: followStats, isLoading } = useQuery({
    queryKey: ['userProfileStats', user.id],
    queryFn: async () => {
      const res = await followApi.getFollowStats(user.id)
      return res.data.data as { followings: number; followers: number; is_following: boolean }
    },
    enabled: open && !isMe,
  })

  // 检查是否可以发送私信
  const { data: canSendData } = useQuery({
    queryKey: ['canSendMessage', user.id],
    queryFn: async () => {
      const res = await pmApi.canSendMessage(user.id)
      return res.data.data as { can_send: boolean; reason: string }
    },
    enabled: open && !isMe,
  })

  // 检查是否被拉黑
  const { data: blockedData } = useQuery({
    queryKey: ['isBlocked', user.id],
    queryFn: async () => {
      const res = await blacklistApi.isBlocked(user.id)
      return res.data.data as { blocked: boolean; reason: string }
    },
    enabled: open && !isMe,
  })

  // 关注/取消关注
  const followMutation = useMutation({
    mutationFn: () => followApi.toggleFollow(user.id),
    onSuccess: (res) => {
      const followed = res.data.data?.followed
      message.success(followed ? '关注成功' : '已取消关注')
      queryClient.invalidateQueries({ queryKey: ['userProfileStats', user.id] })
    },
    onError: () => {
      message.error('操作失败')
    },
  })

  // 拉黑用户
  const blockMutation = useMutation({
    mutationFn: () => blacklistApi.blockUser(user.id),
    onSuccess: () => {
      message.success('已拉黑该用户')
      queryClient.invalidateQueries({ queryKey: ['isBlocked', user.id] })
      queryClient.invalidateQueries({ queryKey: ['canSendMessage', user.id] })
      setOpen(false)
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '操作失败')
    },
  })

  // 取消拉黑
  const unblockMutation = useMutation({
    mutationFn: () => blacklistApi.unblockUser(user.id),
    onSuccess: () => {
      message.success('已取消拉黑')
      queryClient.invalidateQueries({ queryKey: ['isBlocked', user.id] })
      queryClient.invalidateQueries({ queryKey: ['canSendMessage', user.id] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '操作失败')
    },
  })

  // 发起私信
  const handleSendMessage = () => {
    setOpen(false)
    navigate('/user/messages', {
      state: {
        targetUser: {
          id: user.id,
          nickname: user.nickname || user.username,
          avatar: user.avatar,
        },
      },
    })
  }

  // 性别图标
  const GenderIcon = () => {
    if (user.gender === 1) return <ManOutlined className="text-blue-500" />
    if (user.gender === 2) return <WomanOutlined className="text-pink-500" />
    return null
  }

  const cardContent = (
    <div className="w-64">
      {/* 用户信息 */}
      <div className="flex items-start gap-3 mb-3">
        <Avatar src={user.avatar} size={56} icon={<UserOutlined />}>
          {(user.nickname || user.username)?.[0]}
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1">
            <span className="font-semibold text-base truncate">
              {user.nickname || user.username || '未知用户'}
            </span>
            <GenderIcon />
          </div>
          {user.username && user.nickname && (
            <div className="text-gray-400 text-xs">@{user.username}</div>
          )}
          {user.bio && (
            <div className="text-gray-500 text-sm mt-1 line-clamp-2">{user.bio}</div>
          )}
        </div>
      </div>

      {/* 关注统计 */}
      {!isMe && (
        <div className="flex items-center gap-4 text-sm text-gray-500 mb-3 pb-3 border-b border-gray-100">
          {isLoading ? (
            <Spin size="small" />
          ) : (
            <>
              <span>
                关注 <strong className="text-gray-700">{followStats?.followings || 0}</strong>
              </span>
              <span>
                粉丝 <strong className="text-gray-700">{followStats?.followers || 0}</strong>
              </span>
            </>
          )}
        </div>
      )}

      {/* 操作按钮 */}
      {!isMe && (
        <div className="space-y-2">
          <div className="flex gap-2">
            <Button
              type={followStats?.is_following ? 'default' : 'primary'}
              icon={followStats?.is_following ? <UserDeleteOutlined /> : <UserAddOutlined />}
              onClick={() => followMutation.mutate()}
              loading={followMutation.isPending}
              className="flex-1"
              size="small"
            >
              {followStats?.is_following ? '取消关注' : '关注'}
            </Button>
            <Tooltip title={canSendData?.can_send === false ? canSendData.reason : ''}>
              <Button
                icon={<MessageOutlined />}
                onClick={handleSendMessage}
                className="flex-1"
                size="small"
                disabled={canSendData?.can_send === false}
              >
                私信
              </Button>
            </Tooltip>
          </div>
          {/* 拉黑按钮 */}
          <Button
            type="text"
            danger={!blockedData?.blocked}
            icon={<StopOutlined />}
            onClick={() => blockedData?.blocked ? unblockMutation.mutate() : blockMutation.mutate()}
            loading={blockMutation.isPending || unblockMutation.isPending}
            className="w-full"
            size="small"
          >
            {blockedData?.blocked ? '取消拉黑' : '拉黑'}
          </Button>
        </div>
      )}

      {isMe && (
        <div className="text-center text-gray-400 text-sm py-2">这是你自己</div>
      )}
    </div>
  )

  return (
    <Popover
      content={cardContent}
      trigger="click"
      placement={placement}
      open={open}
      onOpenChange={setOpen}
      overlayClassName="user-profile-popover"
    >
      <span className="cursor-pointer inline-block">{children}</span>
    </Popover>
  )
}

export default UserProfileCard

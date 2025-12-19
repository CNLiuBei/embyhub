import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Tabs,
  List,
  Avatar,
  Button,
  Empty,
  Spin,
  App,
  Input,
  Badge,
} from 'antd'
import {
  UserOutlined,
  UserAddOutlined,
  UserDeleteOutlined,
  TeamOutlined,
  SearchOutlined,
  MessageOutlined,
} from '@ant-design/icons'
import { followApi } from '../../services/api'
import { useNavigate } from 'react-router-dom'

interface FollowUser {
  id: string
  username: string
  nickname: string
  avatar: string
  bio?: string
}

interface FollowStats {
  followings: number
  followers: number
  is_following?: boolean
}

const Follow = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState('followings')
  const [searchKeyword, setSearchKeyword] = useState('')

  // 获取关注统计
  const { data: statsData } = useQuery({
    queryKey: ['followStats'],
    queryFn: async () => {
      const res = await followApi.getFollowStats()
      return res.data.data as FollowStats
    },
  })

  // 获取关注列表
  const { data: followingsData, isLoading: followingsLoading } = useQuery({
    queryKey: ['followings'],
    queryFn: async () => {
      const res = await followApi.getFollowings()
      return res.data.data as { list: FollowUser[]; total: number }
    },
    enabled: activeTab === 'followings',
  })

  // 获取粉丝列表
  const { data: followersData, isLoading: followersLoading } = useQuery({
    queryKey: ['followers'],
    queryFn: async () => {
      const res = await followApi.getFollowers()
      return res.data.data as { list: FollowUser[]; total: number }
    },
    enabled: activeTab === 'followers',
  })

  // 关注/取消关注
  const toggleFollowMutation = useMutation({
    mutationFn: (userId: string) => followApi.toggleFollow(userId),
    onSuccess: (res) => {
      const followed = res.data.data?.followed
      message.success(followed ? '关注成功' : '已取消关注')
      queryClient.invalidateQueries({ queryKey: ['followings'] })
      queryClient.invalidateQueries({ queryKey: ['followers'] })
      queryClient.invalidateQueries({ queryKey: ['followStats'] })
    },
    onError: () => {
      message.error('操作失败')
    },
  })

  // 发起私信
  const handleSendMessage = (user: FollowUser) => {
    navigate('/user/messages', { state: { targetUser: user } })
  }

  // 过滤用户列表
  const filterUsers = (users: FollowUser[] | undefined) => {
    if (!users) return []
    if (!searchKeyword.trim()) return users
    const keyword = searchKeyword.toLowerCase()
    return users.filter(
      (u) =>
        u.nickname?.toLowerCase().includes(keyword) ||
        u.username?.toLowerCase().includes(keyword)
    )
  }

  const renderUserList = (
    users: FollowUser[] | undefined,
    loading: boolean,
    isFollowingList: boolean
  ) => {
    const filteredUsers = filterUsers(users)

    if (loading) {
      return (
        <div className="flex justify-center py-12">
          <Spin />
        </div>
      )
    }

    if (!filteredUsers.length) {
      return (
        <Empty
          description={
            searchKeyword
              ? '未找到匹配的用户'
              : isFollowingList
                ? '还没有关注任何人'
                : '还没有粉丝'
          }
          className="py-12"
        />
      )
    }

    return (
      <List
        dataSource={filteredUsers}
        renderItem={(user) => (
          <List.Item
            key={user.id}
            className="hover:bg-gray-50 transition-colors rounded-lg px-4"
            actions={[
              <Button
                key="message"
                type="text"
                icon={<MessageOutlined />}
                onClick={() => handleSendMessage(user)}
              >
                私信
              </Button>,
              isFollowingList ? (
                <Button
                  key="unfollow"
                  danger
                  icon={<UserDeleteOutlined />}
                  onClick={() => toggleFollowMutation.mutate(user.id)}
                  loading={toggleFollowMutation.isPending}
                >
                  取消关注
                </Button>
              ) : (
                <Button
                  key="follow"
                  type="primary"
                  ghost
                  icon={<UserAddOutlined />}
                  onClick={() => toggleFollowMutation.mutate(user.id)}
                  loading={toggleFollowMutation.isPending}
                >
                  回关
                </Button>
              ),
            ]}
          >
            <List.Item.Meta
              avatar={
                <Avatar src={user.avatar} size={48} icon={<UserOutlined />}>
                  {user.nickname?.[0]}
                </Avatar>
              }
              title={
                <span className="font-medium">{user.nickname || user.username}</span>
              }
              description={
                <div className="text-gray-500">
                  <div className="text-xs">@{user.username}</div>
                  {user.bio && (
                    <div className="text-sm mt-1 line-clamp-1">{user.bio}</div>
                  )}
                </div>
              }
            />
          </List.Item>
        )}
      />
    )
  }

  const tabItems = [
    {
      key: 'followings',
      label: (
        <span className="flex items-center gap-2">
          <UserAddOutlined />
          关注
          <Badge
            count={statsData?.followings || 0}
            showZero
            style={{ backgroundColor: '#1890ff' }}
          />
        </span>
      ),
      children: renderUserList(followingsData?.list, followingsLoading, true),
    },
    {
      key: 'followers',
      label: (
        <span className="flex items-center gap-2">
          <TeamOutlined />
          粉丝
          <Badge
            count={statsData?.followers || 0}
            showZero
            style={{ backgroundColor: '#52c41a' }}
          />
        </span>
      ),
      children: renderUserList(followersData?.list, followersLoading, false),
    },
  ]

  return (
    <div className="max-w-4xl mx-auto">
      <Card
        className="glass-card"
        title={
          <div className="flex items-center gap-2">
            <TeamOutlined className="text-blue-500" />
            <span>我的关注</span>
          </div>
        }
        extra={
          <div className="flex items-center gap-4 text-sm">
            <span>
              关注 <strong className="text-blue-500">{statsData?.followings || 0}</strong>
            </span>
            <span>
              粉丝 <strong className="text-green-500">{statsData?.followers || 0}</strong>
            </span>
          </div>
        }
      >
        {/* 搜索框 */}
        <div className="mb-4">
          <Input
            placeholder="搜索用户..."
            prefix={<SearchOutlined className="text-gray-400" />}
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            allowClear
            className="max-w-xs"
          />
        </div>

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          className="follow-tabs"
        />
      </Card>
    </div>
  )
}

export default Follow

import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Tag,
  Space,
  Button,
  Divider,
  List,
  App,
  Spin,
  Empty,
  Popconfirm,
  Modal,
  Form,
  Input,
} from 'antd'
import {
  ArrowLeftOutlined,
  LikeOutlined,
  LikeFilled,
  StarOutlined,
  StarFilled,
  MessageOutlined,
  EyeOutlined,
  DeleteOutlined,
  SendOutlined,
  EditOutlined,
} from '@ant-design/icons'
import { forumApi, followApi } from '../../services/api'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import UserProfileCard from '../../components/UserProfileCard'
import UserAvatar from '../../components/UserAvatar'
import MarkdownEditor from '../../components/MarkdownEditor'
import CommentEditor, { CommentEditorRef } from '../../components/CommentEditor'
import EmojiGifPicker from '../../components/EmojiGifPicker'
import MarkdownContent from '../../components/MarkdownContent'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

interface ForumComment {
  id: number
  topic_id: number
  user_id: string
  content: string
  parent_id: number
  reply_to_user: string | null
  like_count: number
  reply_count: number
  location: string
  created_at: string
  user?: {
    id: string
    nickname: string
    avatar: string
  }
  reply_to_name?: string
  replies?: ForumComment[]
  is_liked: boolean
}

const ForumTopic = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const currentUser = useSelector((state: RootState) => state.auth.user)

  const [commentContent, setCommentContent] = useState('')
  const commentEditorRef = useRef<CommentEditorRef>(null)
  const [replyTo, setReplyTo] = useState<{ id: number; user: string; name: string } | null>(null)
  const [commentPage, setCommentPage] = useState(1)
  const [isFollowing, setIsFollowing] = useState(false)
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [editForm] = Form.useForm()
  const [editContent, setEditContent] = useState('')

  // 获取话题详情
  const { data: topic, isLoading } = useQuery({
    queryKey: ['forumTopic', id],
    queryFn: async () => {
      const res = await forumApi.getTopicDetail(Number(id))
      return res.data.data
    },
    enabled: !!id,
  })

  // 获取评论列表
  const { data: commentsData } = useQuery({
    queryKey: ['forumComments', id, commentPage],
    queryFn: async () => {
      const res = await forumApi.getCommentList({
        topic_id: Number(id),
        page: commentPage,
        page_size: 20,
      })
      return res.data.data as { list: ForumComment[]; total: number }
    },
    enabled: !!id,
  })

  // 点赞话题
  const likeMutation = useMutation({
    mutationFn: () => forumApi.likeTopic(Number(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['forumTopic', id] })
    },
  })

  // 收藏话题
  const favoriteMutation = useMutation({
    mutationFn: () => forumApi.favoriteTopic(Number(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['forumTopic', id] })
    },
  })

  // 点赞评论
  const likeCommentMutation = useMutation({
    mutationFn: (commentId: number) => forumApi.likeComment(commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['forumComments', id] })
    },
  })

  // 发表评论
  const commentMutation = useMutation({
    mutationFn: (data: { topic_id: number; content: string; parent_id?: number; reply_to_user?: string }) =>
      forumApi.createComment(data),
    onSuccess: () => {
      message.success('评论成功')
      setCommentContent('')
      setReplyTo(null)
      queryClient.invalidateQueries({ queryKey: ['forumComments', id] })
      queryClient.invalidateQueries({ queryKey: ['forumTopic', id] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '评论失败')
    },
  })

  // 删除评论
  const deleteCommentMutation = useMutation({
    mutationFn: (commentId: number) => forumApi.deleteComment(commentId),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['forumComments', id] })
      queryClient.invalidateQueries({ queryKey: ['forumTopic', id] })
    },
  })

  // 编辑话题
  const updateMutation = useMutation({
    mutationFn: (data: { title: string; content: string }) =>
      forumApi.updateTopic(Number(id), data),
    onSuccess: () => {
      message.success('更新成功')
      setEditModalOpen(false)
      editForm.resetFields()
      setEditContent('')
      queryClient.invalidateQueries({ queryKey: ['forumTopic', id] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '更新失败')
    },
  })

  // 删除话题
  const deleteMutation = useMutation({
    mutationFn: () => forumApi.deleteTopic(Number(id)),
    onSuccess: () => {
      message.success('删除成功')
      navigate('/user/forum')
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  // 获取关注状态
  const { data: followStats } = useQuery({
    queryKey: ['followStats', topic?.user?.id],
    queryFn: async () => {
      if (!topic?.user?.id) return null
      const res = await followApi.getFollowStats(topic.user.id)
      return res.data.data as { followings: number; followers: number; is_following: boolean }
    },
    enabled: !!topic?.user?.id && topic?.user?.id !== currentUser?.id,
  })

  // 同步关注状态
  useEffect(() => {
    if (followStats) {
      setIsFollowing(followStats.is_following)
    }
  }, [followStats])

  // 关注用户
  const followMutation = useMutation({
    mutationFn: (userId: string) => followApi.toggleFollow(userId),
    onSuccess: (res) => {
      const followed = res.data.data?.followed
      setIsFollowing(followed)
      message.success(followed ? '关注成功' : '已取消关注')
      queryClient.invalidateQueries({ queryKey: ['followStats', topic?.user?.id] })
    },
    onError: () => {
      message.error('操作失败')
    },
  })

  const handleComment = () => {
    if (!commentContent.trim()) {
      message.warning('请输入评论内容')
      return
    }

    const data: { topic_id: number; content: string; parent_id?: number; reply_to_user?: string } = {
      topic_id: Number(id),
      content: commentContent,
    }

    if (replyTo) {
      data.parent_id = replyTo.id
      data.reply_to_user = replyTo.user
    }

    commentMutation.mutate(data)
  }

  const openEditModal = () => {
    if (topic) {
      editForm.setFieldsValue({
        title: topic.title,
      })
      setEditContent(topic.content)
      setEditModalOpen(true)
    }
  }

  const handleEdit = async () => {
    try {
      const values = await editForm.validateFields()
      if (!editContent.trim()) {
        message.error('请输入内容')
        return
      }
      updateMutation.mutate({ title: values.title, content: editContent })
    } catch {
      // validation error
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    )
  }

  if (!topic) {
    return <Empty description="话题不存在" />
  }

  return (
    <div className="space-y-4">
      {/* 返回按钮 */}
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/user/forum')}
      >
        返回论坛
      </Button>

      {/* 话题内容 */}
      <Card className="glass-card">
        <div className="space-y-4">
          {/* 标题 */}
          <div className="flex items-center gap-2 flex-wrap">
            {topic.is_top && <Tag color="red">置顶</Tag>}
            {topic.is_recommend && <Tag color="orange">推荐</Tag>}
            {topic.node && <Tag color="blue">{topic.node.name}</Tag>}
            <h1 className="text-xl font-bold m-0">{topic.title}</h1>
          </div>

          {/* 作者信息 */}
          <div className="flex items-center justify-between">
            <Space>
              <UserProfileCard user={topic.user} placement="bottomLeft">
                <UserAvatar src={topic.user?.avatar} name={topic.user?.nickname} size={40} className="cursor-pointer hover:opacity-80 transition-opacity" />
              </UserProfileCard>
              <div>
                <div className="font-medium">{topic.user?.nickname}</div>
                <div className="text-gray-500 text-sm">
                  {dayjs(topic.created_at).format('YYYY-MM-DD HH:mm')}
                  {topic.location && (
                    <span className="ml-2 text-gray-400">· {topic.location}</span>
                  )}
                </div>
              </div>
            </Space>
            <Space>
              {topic.user?.id === currentUser?.id && (
                <>
                  <Button
                    size="small"
                    icon={<EditOutlined />}
                    onClick={openEditModal}
                  >
                    编辑
                  </Button>
                  <Popconfirm
                    title="确认删除"
                    description="确定要删除这个话题吗？此操作不可恢复。"
                    onConfirm={() => deleteMutation.mutate()}
                    okText="删除"
                    cancelText="取消"
                    okButtonProps={{ danger: true }}
                  >
                    <Button
                      size="small"
                      danger
                      icon={<DeleteOutlined />}
                      loading={deleteMutation.isPending}
                    >
                      删除
                    </Button>
                  </Popconfirm>
                </>
              )}
              {topic.user?.id !== currentUser?.id && (
                <Button
                  size="small"
                  type={isFollowing ? 'default' : 'primary'}
                  onClick={() => followMutation.mutate(topic.user?.id)}
                  loading={followMutation.isPending}
                >
                  {isFollowing ? '已关注' : '关注'}
                </Button>
              )}
            </Space>
          </div>

          <Divider />

          {/* 内容 */}
          <MarkdownContent content={topic.content} />

          <Divider />

          {/* 操作栏 */}
          <div className="flex items-center justify-between">
            <Space size="large" className="text-gray-500">
              <span>
                <EyeOutlined className="mr-1" />
                {topic.view_count} 浏览
              </span>
              <span>
                <MessageOutlined className="mr-1" />
                {topic.comment_count} 评论
              </span>
            </Space>
            <Space>
              <Button
                type={topic.is_liked ? 'primary' : 'default'}
                icon={topic.is_liked ? <LikeFilled /> : <LikeOutlined />}
                onClick={() => likeMutation.mutate()}
              >
                {topic.like_count}
              </Button>
              <Button
                type={topic.is_faved ? 'primary' : 'default'}
                icon={topic.is_faved ? <StarFilled /> : <StarOutlined />}
                onClick={() => favoriteMutation.mutate()}
              >
                {topic.favorite_count}
              </Button>
            </Space>
          </div>
        </div>
      </Card>

      {/* 评论区 */}
      <Card title={`评论 (${commentsData?.total || 0})`} className="glass-card">
        {/* 评论列表 - 优化样式 */}
        {commentsData?.list?.length ? (
          <List
            dataSource={commentsData.list}
            pagination={{
              current: commentPage,
              pageSize: 20,
              total: commentsData.total,
              onChange: setCommentPage,
              showSizeChanger: false,
              className: 'mt-4',
            }}
            renderItem={(comment) => (
              <List.Item key={comment.id} className="block !py-4 hover:bg-gray-50 rounded-lg transition-colors">
                <div className="flex gap-3">
                  <UserProfileCard user={comment.user || { id: comment.user_id }}>
                    <UserAvatar 
                      src={comment.user?.avatar} 
                      name={comment.user?.nickname}
                      size={42}
                      className="cursor-pointer hover:opacity-80 transition-opacity flex-shrink-0"
                    />
                  </UserProfileCard>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium text-gray-900">{comment.user?.nickname}</span>
                      <span className="text-gray-400 text-xs">
                        {dayjs(comment.created_at).fromNow()}
                      </span>
                      {comment.location && (
                        <span className="text-gray-400 text-xs">· {comment.location}</span>
                      )}
                    </div>
                    <div className="mt-2 text-gray-700 text-[15px] leading-relaxed">
                      <MarkdownContent content={comment.content} className="prose-sm" compact />
                    </div>
                    <div className="mt-2 flex items-center gap-4">
                      <Button
                        type="text"
                        size="small"
                        className={`px-2 ${comment.is_liked ? 'text-red-500' : 'text-gray-500'} hover:text-red-500`}
                        icon={comment.is_liked ? <LikeFilled /> : <LikeOutlined />}
                        onClick={() => likeCommentMutation.mutate(comment.id)}
                      >
                        {comment.like_count || '赞'}
                      </Button>
                      <Button
                        type="text"
                        size="small"
                        className="px-2 text-gray-500 hover:text-blue-500"
                        icon={<MessageOutlined />}
                        onClick={() =>
                          setReplyTo({
                            id: comment.id,
                            user: comment.user_id,
                            name: comment.user?.nickname || '',
                          })
                        }
                      >
                        回复
                      </Button>
                      {comment.user_id === currentUser?.id && (
                        <Popconfirm
                          title="确定删除这条评论？"
                          onConfirm={() => deleteCommentMutation.mutate(comment.id)}
                        >
                          <Button 
                            type="text" 
                            size="small" 
                            className="px-2 text-gray-400 hover:text-red-500"
                            icon={<DeleteOutlined />}
                          >
                            删除
                          </Button>
                        </Popconfirm>
                      )}
                    </div>

                    {/* 回复列表 - 优化样式 */}
                    {comment.replies && comment.replies.length > 0 && (
                      <div className="mt-3 bg-gray-50 rounded-lg p-3 space-y-3">
                        {comment.replies.map((reply) => (
                          <div key={reply.id} className="flex gap-2">
                            <UserProfileCard user={reply.user || { id: reply.user_id }}>
                              <UserAvatar 
                                src={reply.user?.avatar} 
                                name={reply.user?.nickname}
                                size={28}
                                className="cursor-pointer hover:opacity-80 transition-opacity flex-shrink-0"
                              />
                            </UserProfileCard>
                            <div className="flex-1 min-w-0">
                              <div className="text-sm">
                                <span className="font-medium text-gray-900">{reply.user?.nickname}</span>
                                {reply.reply_to_name && (
                                  <span className="text-gray-500">
                                    {' '}回复{' '}
                                    <span className="text-blue-500">@{reply.reply_to_name}</span>
                                  </span>
                                )}
                                <span className="text-gray-400 text-xs ml-2">
                                  {dayjs(reply.created_at).fromNow()}
                                </span>
                                {reply.location && (
                                  <span className="text-gray-400 text-xs ml-1">· {reply.location}</span>
                                )}
                              </div>
                              <div className="text-gray-700 text-sm mt-1">
                                <MarkdownContent content={reply.content} className="prose-xs" />
                              </div>
                            </div>
                          </div>
                        ))}
                        {comment.reply_count > 3 && (
                          <Button type="link" size="small" className="pl-8">
                            查看全部 {comment.reply_count} 条回复 →
                          </Button>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </List.Item>
            )}
          />
        ) : (
          <Empty 
            description="暂无评论，快来抢沙发吧~" 
            className="py-12"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        )}

        {/* 评论输入框 - 放在最下方 */}
        <div className="mt-6 pt-4 border-t border-gray-100">
          {replyTo && (
            <div className="mb-2 px-2 py-1 bg-blue-50 rounded text-sm text-blue-600 flex items-center justify-between">
              <span>回复 @{replyTo.name}</span>
              <Button
                type="link"
                size="small"
                onClick={() => setReplyTo(null)}
                className="text-blue-500"
              >
                取消
              </Button>
            </div>
          )}
          <div className="flex gap-3 items-start">
            <UserAvatar src={currentUser?.avatar} name={currentUser?.nickname} size={40} />
            <div className="flex-1">
              <CommentEditor
                ref={commentEditorRef}
                value={commentContent}
                onChange={setCommentContent}
                placeholder={replyTo ? `回复 @${replyTo.name}...` : '写下你的评论...'}
                minHeight={200}
                maxLength={1000}
                disabled={commentMutation.isPending}
                renderActions={() => (
                  <>
                    <EmojiGifPicker
                      onEmojiSelect={(emoji) => commentEditorRef.current?.insertEmoji(emoji)}
                      onGifSelect={(url) => commentEditorRef.current?.insertGif(url)}
                      disabled={commentMutation.isPending}
                    />
                    <Button
                      type="primary"
                      icon={<SendOutlined />}
                      onClick={handleComment}
                      loading={commentMutation.isPending}
                      disabled={!commentContent.trim()}
                    >
                      发表评论
                    </Button>
                  </>
                )}
              />
            </div>
          </div>
        </div>
      </Card>

      {/* 编辑话题弹窗 */}
      <Modal
        title="编辑话题"
        open={editModalOpen}
        onOk={handleEdit}
        onCancel={() => {
          setEditModalOpen(false)
          editForm.resetFields()
          setEditContent('')
        }}
        confirmLoading={updateMutation.isPending}
        width={800}
      >
        <Form form={editForm} name="editTopicDetailForm" layout="vertical">
          <Form.Item
            name="title"
            label="标题"
            rules={[
              { required: true, message: '请输入标题' },
              { max: 128, message: '标题最多128个字符' },
            ]}
          >
            <Input placeholder="请输入标题" />
          </Form.Item>
          <Form.Item label="内容" required>
            <MarkdownEditor
              value={editContent}
              onChange={setEditContent}
              placeholder="支持 Markdown 格式，可直接粘贴或拖拽图片..."
              height={350}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ForumTopic

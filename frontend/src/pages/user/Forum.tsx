import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
  Card,
  List,
  Tag,
  Space,
  Button,
  Tabs,
  Empty,
  Spin,
  App,
  Input,
  Modal,
  Form,
  Select,
  Dropdown,
} from 'antd'
import type { MenuProps } from 'antd'
import {
  MessageOutlined,
  LikeOutlined,
  LikeFilled,
  EyeOutlined,
  StarOutlined,
  StarFilled,
  PlusOutlined,
  FireOutlined,
  ClockCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  MoreOutlined,
  FileTextOutlined,
  HeartOutlined,
} from '@ant-design/icons'
import { forumApi } from '../../services/api'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import UserProfileCard from '../../components/UserProfileCard'
import UserAvatar from '../../components/UserAvatar'
import MarkdownEditor from '../../components/MarkdownEditor'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

interface ForumNode {
  id: number
  name: string
  description: string
  icon: string
  topic_count: number
}

interface ForumTopic {
  id: number
  node_id: number
  user_id: string
  title: string
  content: string
  images: string
  view_count: number
  comment_count: number
  like_count: number
  favorite_count: number
  is_top: boolean
  is_recommend: boolean
  location: string
  created_at: string
  user?: {
    id: string
    nickname: string
    avatar: string
  }
  node?: ForumNode
  is_liked: boolean
  is_faved: boolean
}

const Forum = () => {
  const { message, modal } = App.useApp()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const currentUser = useSelector((state: RootState) => state.auth.user)
  const [activeNode, setActiveNode] = useState<number>(0)
  const [orderBy, setOrderBy] = useState<string>('latest')
  const [page, setPage] = useState(1)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [editingTopic, setEditingTopic] = useState<ForumTopic | null>(null)
  const [viewMode, setViewMode] = useState<'all' | 'my' | 'favorites'>('all')
  const [form] = Form.useForm()
  const [editForm] = Form.useForm()
  const [createContent, setCreateContent] = useState('')
  const [editContent, setEditContent] = useState('')

  // 获取节点列表
  const { data: nodesData } = useQuery({
    queryKey: ['forumNodes'],
    queryFn: async () => {
      const res = await forumApi.getNodes()
      return res.data.data as ForumNode[]
    },
  })

  // 获取话题列表
  const { data: topicsData, isLoading } = useQuery({
    queryKey: ['forumTopics', activeNode, orderBy, page, viewMode],
    queryFn: async () => {
      if (viewMode === 'my') {
        const res = await forumApi.getMyTopics({ page, page_size: 20 })
        return res.data.data as { list: ForumTopic[]; total: number }
      }
      if (viewMode === 'favorites') {
        const res = await forumApi.getMyFavorites({ page, page_size: 20 })
        return res.data.data as { list: ForumTopic[]; total: number }
      }
      const res = await forumApi.getTopicList({
        node_id: activeNode || undefined,
        order_by: orderBy,
        page,
        page_size: 20,
      })
      return res.data.data as { list: ForumTopic[]; total: number }
    },
  })

  // 点赞
  const likeMutation = useMutation({
    mutationFn: (id: number) => forumApi.likeTopic(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['forumTopics'] })
    },
  })

  // 收藏
  const favoriteMutation = useMutation({
    mutationFn: (id: number) => forumApi.favoriteTopic(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['forumTopics'] })
    },
  })

  // 发布话题
  const createMutation = useMutation({
    mutationFn: (data: { node_id: number; title: string; content: string; content_type?: string }) =>
      forumApi.createTopic(data),
    onSuccess: () => {
      message.success('发布成功')
      setCreateModalOpen(false)
      form.resetFields()
      setCreateContent('')
      queryClient.invalidateQueries({ queryKey: ['forumTopics'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '发布失败')
    },
  })

  // 编辑话题
  const updateMutation = useMutation({
    mutationFn: (data: { id: number; title: string; content: string }) =>
      forumApi.updateTopic(data.id, { title: data.title, content: data.content }),
    onSuccess: () => {
      message.success('更新成功')
      setEditModalOpen(false)
      setEditingTopic(null)
      editForm.resetFields()
      setEditContent('')
      queryClient.invalidateQueries({ queryKey: ['forumTopics'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '更新失败')
    },
  })

  // 删除话题
  const deleteMutation = useMutation({
    mutationFn: (id: number) => forumApi.deleteTopic(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['forumTopics'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      if (!createContent.trim()) {
        message.error('请输入内容')
        return
      }
      createMutation.mutate({ 
        node_id: values.node_id, 
        title: values.title, 
        content: createContent,
        content_type: 'markdown'
      })
    } catch {
      // validation error
    }
  }

  const handleEdit = async () => {
    try {
      const values = await editForm.validateFields()
      if (editingTopic) {
        if (!editContent.trim()) {
          message.error('请输入内容')
          return
        }
        updateMutation.mutate({ 
          id: editingTopic.id, 
          title: values.title, 
          content: editContent 
        })
      }
    } catch {
      // validation error
    }
  }

  const openEditModal = (topic: ForumTopic, e: React.MouseEvent) => {
    e.stopPropagation()
    setEditingTopic(topic)
    editForm.setFieldsValue({
      title: topic.title,
    })
    setEditContent(topic.content)
    setEditModalOpen(true)
  }

  const handleDelete = (topic: ForumTopic, e: React.MouseEvent) => {
    e.stopPropagation()
    modal.confirm({
      title: '确认删除',
      content: `确定要删除话题「${topic.title}」吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutate(topic.id),
    })
  }

  const getTopicActions = (topic: ForumTopic): MenuProps['items'] => {
    const items: MenuProps['items'] = []
    if (topic.user_id === currentUser?.id) {
      items.push(
        { key: 'edit', icon: <EditOutlined />, label: '编辑' },
        { key: 'delete', icon: <DeleteOutlined />, label: '删除', danger: true }
      )
    }
    return items
  }

  const nodeItems = [
    { key: '0', label: '全部' },
    ...(nodesData?.map((node) => ({
      key: String(node.id),
      label: (
        <span>
          {node.icon && <span className="mr-1">{node.icon}</span>}
          {node.name}
          <span className="text-gray-400 text-xs ml-1">({node.topic_count})</span>
        </span>
      ),
    })) || []),
  ]

  return (
    <div className="space-y-4">
      {/* 顶部操作栏 */}
      <Card size="small" className="glass-card">
        <div className="flex justify-between items-center">
          <Tabs
            activeKey={String(activeNode)}
            onChange={(key) => {
              setActiveNode(Number(key))
              setViewMode('all')
              setPage(1)
            }}
            items={nodeItems}
            size="small"
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalOpen(true)}
          >
            发帖
          </Button>
        </div>
      </Card>

      {/* 排序和筛选选项 */}
      <div className="flex justify-between items-center">
        <div className="flex gap-2">
          <Button
            type={viewMode === 'all' && orderBy === 'latest' ? 'primary' : 'default'}
            size="small"
            icon={<ClockCircleOutlined />}
            onClick={() => { setViewMode('all'); setOrderBy('latest'); setPage(1) }}
          >
            最新
          </Button>
          <Button
            type={viewMode === 'all' && orderBy === 'hot' ? 'primary' : 'default'}
            size="small"
            icon={<FireOutlined />}
            onClick={() => { setViewMode('all'); setOrderBy('hot'); setPage(1) }}
          >
            热门
          </Button>
          <Button
            type={viewMode === 'all' && orderBy === 'recommend' ? 'primary' : 'default'}
            size="small"
            icon={<StarOutlined />}
            onClick={() => { setViewMode('all'); setOrderBy('recommend'); setPage(1) }}
          >
            推荐
          </Button>
        </div>
        <div className="flex gap-2">
          <Button
            type={viewMode === 'my' ? 'primary' : 'default'}
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => { setViewMode('my'); setPage(1) }}
          >
            我的帖子
          </Button>
          <Button
            type={viewMode === 'favorites' ? 'primary' : 'default'}
            size="small"
            icon={<HeartOutlined />}
            onClick={() => { setViewMode('favorites'); setPage(1) }}
          >
            我的收藏
          </Button>
        </div>
      </div>

      {/* 话题列表 */}
      <Card className="glass-card">
        <Spin spinning={isLoading}>
          {topicsData?.list?.length ? (
            <List
              itemLayout="vertical"
              dataSource={topicsData.list}
              pagination={{
                current: page,
                pageSize: 20,
                total: topicsData.total,
                onChange: setPage,
                showSizeChanger: false,
              }}
              renderItem={(topic) => (
                <List.Item
                  key={topic.id}
                  className="cursor-pointer hover:bg-gray-50 rounded-lg transition-colors px-3"
                  onClick={() => navigate(`/user/forum/topic/${topic.id}`)}
                  actions={[
                    <Space
                      key="actions"
                      onClick={(e) => e.stopPropagation()}
                      className="text-gray-500"
                    >
                      <span>
                        <EyeOutlined className="mr-1" />
                        {topic.view_count}
                      </span>
                      <span
                        className={`cursor-pointer ${topic.is_liked ? 'text-red-500' : ''}`}
                        onClick={() => likeMutation.mutate(topic.id)}
                      >
                        {topic.is_liked ? <LikeFilled className="mr-1" /> : <LikeOutlined className="mr-1" />}
                        {topic.like_count}
                      </span>
                      <span>
                        <MessageOutlined className="mr-1" />
                        {topic.comment_count}
                      </span>
                      <span
                        className={`cursor-pointer ${topic.is_faved ? 'text-yellow-500' : ''}`}
                        onClick={() => favoriteMutation.mutate(topic.id)}
                      >
                        {topic.is_faved ? <StarFilled className="mr-1" /> : <StarOutlined className="mr-1" />}
                        {topic.favorite_count}
                      </span>
                      {topic.user_id === currentUser?.id && (
                        <Dropdown
                          menu={{
                            items: getTopicActions(topic),
                            onClick: ({ key, domEvent }) => {
                              domEvent.stopPropagation()
                              if (key === 'edit') openEditModal(topic, domEvent as unknown as React.MouseEvent)
                              if (key === 'delete') handleDelete(topic, domEvent as unknown as React.MouseEvent)
                            },
                          }}
                          trigger={['click']}
                        >
                          <Button
                            type="text"
                            size="small"
                            icon={<MoreOutlined />}
                            onClick={(e) => e.stopPropagation()}
                          />
                        </Dropdown>
                      )}
                    </Space>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <div onClick={(e) => e.stopPropagation()}>
                        <UserProfileCard user={topic.user || { id: topic.user_id }}>
                          <UserAvatar src={topic.user?.avatar} name={topic.user?.nickname} size={40} className="cursor-pointer hover:opacity-80 transition-opacity" />
                        </UserProfileCard>
                      </div>
                    }
                    title={
                      <div className="flex items-center gap-2">
                        {topic.is_top && <Tag color="red">置顶</Tag>}
                        {topic.is_recommend && <Tag color="orange">推荐</Tag>}
                        {topic.node && (
                          <Tag color="blue">{topic.node.name}</Tag>
                        )}
                        <span className="font-medium text-base">{topic.title}</span>
                      </div>
                    }
                    description={
                      <div className="text-gray-500 text-sm">
                        <span>{topic.user?.nickname}</span>
                        <span className="mx-2">·</span>
                        <span>{dayjs(topic.created_at).fromNow()}</span>
                        {topic.location && (
                          <>
                            <span className="mx-2">·</span>
                            <span className="text-gray-400">{topic.location}</span>
                          </>
                        )}
                      </div>
                    }
                  />
                  <div className="text-gray-600 line-clamp-2">
                    {topic.content
                      .replace(/!\[.*?\]\(.*?\)/g, '[图片]')  // 替换 Markdown 图片为 [图片]
                      .replace(/\[.*?\]\(.*?\)/g, '')         // 移除 Markdown 链接
                      .replace(/<[^>]+>/g, '')                // 移除 HTML 标签
                      .replace(/[#*`~>|-]/g, '')              // 移除 Markdown 符号
                      .replace(/\n+/g, ' ')                   // 换行转空格
                      .trim()
                      .slice(0, 200)}
                  </div>
                </List.Item>
              )}
            />
          ) : (
            <Empty description="暂无话题" />
          )}
        </Spin>
      </Card>

      {/* 发帖弹窗 */}
      <Modal
        title="发布话题"
        open={createModalOpen}
        onOk={handleCreate}
        onCancel={() => { setCreateModalOpen(false); setCreateContent(''); form.resetFields() }}
        confirmLoading={createMutation.isPending}
        width={800}
      >
        <Form form={form} name="createTopicForm" layout="vertical">
          <Form.Item
            name="node_id"
            label="选择板块"
            rules={[{ required: true, message: '请选择板块' }]}
          >
            <Select placeholder="请选择板块">
              {nodesData?.map((node) => (
                <Select.Option key={node.id} value={node.id}>
                  {node.icon} {node.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
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
              value={createContent}
              onChange={setCreateContent}
              placeholder="支持 Markdown 格式，可直接粘贴或拖拽图片..."
              height={350}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑话题弹窗 */}
      <Modal
        title="编辑话题"
        open={editModalOpen}
        onOk={handleEdit}
        onCancel={() => {
          setEditModalOpen(false)
          setEditingTopic(null)
          editForm.resetFields()
          setEditContent('')
        }}
        confirmLoading={updateMutation.isPending}
        width={800}
      >
        <Form form={editForm} name="editTopicForm" layout="vertical">
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

export default Forum

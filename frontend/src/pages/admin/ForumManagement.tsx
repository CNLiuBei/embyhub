import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  Switch,
  App,
  Tabs,
  Popconfirm,
  Avatar,
  Select,
  Spin,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PushpinOutlined,
  StarOutlined,
  CloudUploadOutlined,
} from '@ant-design/icons'
import { forumAdminApi, adminApi, imageHostApi } from '../../services/api'
import dayjs from 'dayjs'

interface ForumNode {
  id: number
  name: string
  description: string
  icon: string
  sort_order: number
  topic_count: number
  status: number
  created_at: string
}

interface ForumTopic {
  id: number
  node_id: number
  user_id: string
  title: string
  view_count: number
  comment_count: number
  like_count: number
  is_top: boolean
  is_recommend: boolean
  status: number
  created_at: string
  user?: {
    id: string
    nickname: string
    avatar: string
  }
  node?: {
    name: string
  }
}

const ForumManagement = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState('nodes')

  // 节点相关状态
  const [nodeModalOpen, setNodeModalOpen] = useState(false)
  const [editingNode, setEditingNode] = useState<ForumNode | null>(null)
  const [nodeForm] = Form.useForm()

  // 话题相关状态
  const [topicPage, setTopicPage] = useState(1)
  const [topicNodeFilter, setTopicNodeFilter] = useState<number | undefined>()
  const [topicStatusFilter, setTopicStatusFilter] = useState<number>(-1)
  const [topicKeyword, setTopicKeyword] = useState('')

  // 图床设置状态
  const [imageHostForm] = Form.useForm()
  const [imageHostLoading, setImageHostLoading] = useState(false)

  // 获取节点列表
  const { data: nodes, isLoading: nodesLoading } = useQuery({
    queryKey: ['adminForumNodes'],
    queryFn: async () => {
      const res = await forumAdminApi.getNodes()
      return res.data.data as ForumNode[]
    },
  })

  // 获取图床设置
  const { data: imageHostSettings, isLoading: imageHostSettingsLoading } = useQuery({
    queryKey: ['imageHostSettings'],
    queryFn: async () => {
      const res = await adminApi.getImageHostSettings()
      return res.data.data as { enabled: boolean; base_url: string }
    },
  })

  // 图床设置表单初始化
  useEffect(() => {
    if (imageHostSettings) {
      imageHostForm.setFieldsValue(imageHostSettings)
    }
  }, [imageHostSettings, imageHostForm])

  // 获取话题列表
  const { data: topicsData, isLoading: topicsLoading } = useQuery({
    queryKey: ['adminForumTopics', topicPage, topicNodeFilter, topicStatusFilter, topicKeyword],
    queryFn: async () => {
      const res = await forumAdminApi.getTopicList({
        node_id: topicNodeFilter,
        status: topicStatusFilter,
        keyword: topicKeyword,
        page: topicPage,
        page_size: 20,
      })
      return res.data.data as { list: ForumTopic[]; total: number }
    },
  })

  // 节点操作
  const createNodeMutation = useMutation({
    mutationFn: (data: { name: string; description?: string; icon?: string; sort_order?: number }) =>
      forumAdminApi.createNode(data),
    onSuccess: () => {
      message.success('创建成功')
      setNodeModalOpen(false)
      nodeForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['adminForumNodes'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '创建失败')
    },
  })

  const updateNodeMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<ForumNode> }) =>
      forumAdminApi.updateNode(id, data),
    onSuccess: () => {
      message.success('更新成功')
      setNodeModalOpen(false)
      setEditingNode(null)
      nodeForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['adminForumNodes'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '更新失败')
    },
  })

  const deleteNodeMutation = useMutation({
    mutationFn: (id: number) => forumAdminApi.deleteNode(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['adminForumNodes'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  // 话题操作
  const deleteTopicMutation = useMutation({
    mutationFn: (id: number) => forumAdminApi.deleteTopic(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['adminForumTopics'] })
    },
  })

  const setTopicTopMutation = useMutation({
    mutationFn: ({ id, isTop }: { id: number; isTop: boolean }) =>
      forumAdminApi.setTopicTop(id, isTop),
    onSuccess: () => {
      message.success('设置成功')
      queryClient.invalidateQueries({ queryKey: ['adminForumTopics'] })
    },
  })

  const setTopicRecommendMutation = useMutation({
    mutationFn: ({ id, isRecommend }: { id: number; isRecommend: boolean }) =>
      forumAdminApi.setTopicRecommend(id, isRecommend),
    onSuccess: () => {
      message.success('设置成功')
      queryClient.invalidateQueries({ queryKey: ['adminForumTopics'] })
    },
  })

  // 保存图床设置
  const handleSaveImageHost = async () => {
    try {
      const values = await imageHostForm.validateFields()
      setImageHostLoading(true)
      await adminApi.saveImageHostSettings(values)
      message.success('保存成功')
      // 清除前端缓存
      imageHostApi.clearCache()
      queryClient.invalidateQueries({ queryKey: ['imageHostSettings'] })
    } catch (err) {
      message.error((err as Error).message || '保存失败')
    } finally {
      setImageHostLoading(false)
    }
  }

  // 节点表格列
  const nodeColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    {
      title: '图标',
      dataIndex: 'icon',
      key: 'icon',
      width: 60,
      render: (icon: string) => <span className="text-xl">{icon || '📁'}</span>,
    },
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    { title: '排序', dataIndex: 'sort_order', key: 'sort_order', width: 60 },
    { title: '话题数', dataIndex: 'topic_count', key: 'topic_count', width: 80 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number, record: ForumNode) => (
        <Switch
          checked={status === 1}
          onChange={(checked) =>
            updateNodeMutation.mutate({ id: record.id, data: { status: checked ? 1 : 0 } })
          }
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: unknown, record: ForumNode) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingNode(record)
              nodeForm.setFieldsValue(record)
              setNodeModalOpen(true)
            }}
          />
          <Popconfirm
            title="确定删除该节点？"
            onConfirm={() => deleteNodeMutation.mutate(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // 话题表格列
  const topicColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: ForumTopic) => (
        <div>
          {record.is_top && <Tag color="red">置顶</Tag>}
          {record.is_recommend && <Tag color="orange">推荐</Tag>}
          <span>{title}</span>
        </div>
      ),
    },
    {
      title: '作者',
      dataIndex: 'user',
      key: 'user',
      width: 120,
      render: (user: ForumTopic['user']) => (
        <Space>
          <Avatar src={user?.avatar} size="small">
            {user?.nickname?.[0]}
          </Avatar>
          <span>{user?.nickname}</span>
        </Space>
      ),
    },
    {
      title: '板块',
      dataIndex: 'node',
      key: 'node',
      width: 100,
      render: (node: ForumTopic['node']) => <Tag>{node?.name}</Tag>,
    },
    { title: '浏览', dataIndex: 'view_count', key: 'view_count', width: 60 },
    { title: '评论', dataIndex: 'comment_count', key: 'comment_count', width: 60 },
    { title: '点赞', dataIndex: 'like_count', key: 'like_count', width: 60 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        const map: Record<number, { color: string; text: string }> = {
          0: { color: 'green', text: '正常' },
          1: { color: 'red', text: '已删除' },
          2: { color: 'orange', text: '待审核' },
        }
        const s = map[status] || { color: 'default', text: '未知' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    {
      title: '发布时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: ForumTopic) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<PushpinOutlined />}
            onClick={() =>
              setTopicTopMutation.mutate({ id: record.id, isTop: !record.is_top })
            }
          >
            {record.is_top ? '取消置顶' : '置顶'}
          </Button>
          <Button
            type="link"
            size="small"
            icon={<StarOutlined />}
            onClick={() =>
              setTopicRecommendMutation.mutate({ id: record.id, isRecommend: !record.is_recommend })
            }
          >
            {record.is_recommend ? '取消推荐' : '推荐'}
          </Button>
          <Popconfirm
            title="确定删除该话题？"
            onConfirm={() => deleteTopicMutation.mutate(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const handleNodeSubmit = async () => {
    try {
      const values = await nodeForm.validateFields()
      if (editingNode) {
        updateNodeMutation.mutate({ id: editingNode.id, data: values })
      } else {
        createNodeMutation.mutate(values)
      }
    } catch {
      // validation error
    }
  }

  return (
    <div className="space-y-4">
      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'nodes',
              label: '板块管理',
              children: (
                <div className="space-y-4">
                  <div className="flex justify-end">
                    <Button
                      type="primary"
                      icon={<PlusOutlined />}
                      onClick={() => {
                        setEditingNode(null)
                        nodeForm.resetFields()
                        setNodeModalOpen(true)
                      }}
                    >
                      添加板块
                    </Button>
                  </div>
                  <Table
                    columns={nodeColumns}
                    dataSource={nodes || []}
                    rowKey="id"
                    loading={nodesLoading}
                    pagination={false}
                    size="small"
                  />
                </div>
              ),
            },
            {
              key: 'topics',
              label: '话题管理',
              children: (
                <div className="space-y-4">
                  {/* 筛选条件 */}
                  <div className="flex flex-wrap gap-4 items-center">
                    <Select
                      allowClear
                      placeholder="选择板块"
                      style={{ width: 150 }}
                      value={topicNodeFilter}
                      onChange={setTopicNodeFilter}
                    >
                      {nodes?.map((node) => (
                        <Select.Option key={node.id} value={node.id}>
                          {node.icon} {node.name}
                        </Select.Option>
                      ))}
                    </Select>
                    <Select
                      style={{ width: 120 }}
                      value={topicStatusFilter}
                      onChange={setTopicStatusFilter}
                    >
                      <Select.Option value={-1}>全部状态</Select.Option>
                      <Select.Option value={0}>正常</Select.Option>
                      <Select.Option value={1}>已删除</Select.Option>
                      <Select.Option value={2}>待审核</Select.Option>
                    </Select>
                    <Input.Search
                      placeholder="搜索标题"
                      style={{ width: 200 }}
                      value={topicKeyword}
                      onChange={(e) => setTopicKeyword(e.target.value)}
                      onSearch={() => setTopicPage(1)}
                    />
                  </div>
                  <Table
                    columns={topicColumns}
                    dataSource={topicsData?.list || []}
                    rowKey="id"
                    loading={topicsLoading}
                    pagination={{
                      current: topicPage,
                      pageSize: 20,
                      total: topicsData?.total || 0,
                      onChange: setTopicPage,
                      showTotal: (t) => `共 ${t} 条`,
                    }}
                    size="small"
                  />
                </div>
              ),
            },
            {
              key: 'imageHost',
              label: (
                <span>
                  <CloudUploadOutlined className="mr-1" />
                  图床设置
                </span>
              ),
              children: (
                <Spin spinning={imageHostSettingsLoading}>
                  <div className="max-w-xl">
                    <div className="mb-4 p-4 bg-blue-50 rounded-lg text-sm text-blue-700">
                      <p className="font-medium mb-2">📷 Telegram 图床说明</p>
                      <p>用户发帖和评论时上传的图片会自动存储到 Telegram 图床，不占用服务器空间。</p>
                      <p className="mt-1">需要先部署 Telegram 图床服务，详见项目文档。</p>
                    </div>
                    <Form form={imageHostForm} layout="vertical" name="imageHostSettingsForm">
                      <Form.Item
                        name="enabled"
                        label="启用图床"
                        valuePropName="checked"
                      >
                        <Switch checkedChildren="开启" unCheckedChildren="关闭" />
                      </Form.Item>
                      <Form.Item
                        name="base_url"
                        label="图床地址"
                        rules={[{ required: true, message: '请输入图床地址' }]}
                        extra="例如：https://img.liubei.org"
                      >
                        <Input placeholder="请输入图床服务地址" />
                      </Form.Item>
                      <Form.Item>
                        <Button
                          type="primary"
                          onClick={handleSaveImageHost}
                          loading={imageHostLoading}
                        >
                          保存设置
                        </Button>
                      </Form.Item>
                    </Form>
                  </div>
                </Spin>
              ),
            },
          ]}
        />
      </Card>

      {/* 节点编辑弹窗 */}
      <Modal
        title={editingNode ? '编辑板块' : '添加板块'}
        open={nodeModalOpen}
        onOk={handleNodeSubmit}
        onCancel={() => {
          setNodeModalOpen(false)
          setEditingNode(null)
          nodeForm.resetFields()
        }}
        confirmLoading={createNodeMutation.isPending || updateNodeMutation.isPending}
      >
        <Form form={nodeForm} layout="vertical" name="forumNodeForm">
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="请输入板块名称" />
          </Form.Item>
          <Form.Item name="icon" label="图标">
            <Input placeholder="请输入图标（emoji）" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="请输入描述" rows={3} />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber min={0} placeholder="数字越小越靠前" style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ForumManagement

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, Switch, App, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined, StopOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'
import dayjs from 'dayjs'

const { TextArea } = Input

interface Announcement {
  id: number
  title: string
  content: string
  type: number
  is_top: boolean
  status: number
  created_at: string
}

const typeOptions = [
  { label: '通知', value: 1 },
  { label: '活动', value: 2 },
  { label: '更新', value: 3 },
]

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '草稿', color: 'default' },
  1: { text: '已发布', color: 'green' },
  2: { text: '已下线', color: 'red' },
}

const Announcements = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const { data, isLoading } = useQuery({
    queryKey: ['announcements', page, pageSize],
    queryFn: async () => {
      const response = await adminApi.getAnnouncements({ page, page_size: pageSize })
      return response.data.data as { list: Announcement[]; total: number }
    },
  })

  const createMutation = useMutation({
    mutationFn: (data: any) => adminApi.createAnnouncement(data),
    onSuccess: () => {
      message.success('创建成功')
      setModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
    },
    onError: (err: any) => message.error(err.message || '创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => adminApi.updateAnnouncement(id, data),
    onSuccess: () => {
      message.success('更新成功')
      setModalOpen(false)
      setEditingId(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
    },
    onError: (err: any) => message.error(err.message || '更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminApi.deleteAnnouncement(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
    },
  })

  const publishMutation = useMutation({
    mutationFn: (id: number) => adminApi.publishAnnouncement(id),
    onSuccess: () => {
      message.success('发布成功')
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
    },
  })

  const offlineMutation = useMutation({
    mutationFn: (id: number) => adminApi.offlineAnnouncement(id),
    onSuccess: () => {
      message.success('下线成功')
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
    },
  })

  const handleCreate = () => {
    setEditingId(null)
    form.resetFields()
    form.setFieldsValue({ type: 1, is_top: false, status: 0 })
    setModalOpen(true)
  }

  const handleEdit = (record: Announcement) => {
    setEditingId(record.id)
    form.setFieldsValue(record)
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: values })
    } else {
      createMutation.mutate(values)
    }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: number) => {
        const colors = ['', 'blue', 'orange', 'green']
        const names = ['', '通知', '活动', '更新']
        return <Tag color={colors[type]}>{names[type]}</Tag>
      },
    },
    {
      title: '置顶',
      dataIndex: 'is_top',
      key: 'is_top',
      width: 80,
      render: (isTop: boolean) => isTop ? <Tag color="red">置顶</Tag> : '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => {
        const config = statusMap[status] || statusMap[0]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 240,
      fixed: 'right' as const,
      render: (_: unknown, record: Announcement) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          {record.status === 0 && (
            <Button type="link" size="small" icon={<SendOutlined />} onClick={() => publishMutation.mutate(record.id)}>
              发布
            </Button>
          )}
          {record.status === 1 && (
            <Button type="link" size="small" danger icon={<StopOutlined />} onClick={() => offlineMutation.mutate(record.id)}>
              下线
            </Button>
          )}
          <Popconfirm title="确定删除？" onConfirm={() => deleteMutation.mutate(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="glass-card p-4">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold m-0">公告管理</h2>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新建公告
          </Button>
        </div>
      </div>

      <div className="glass-card p-4 flex-1">
        <Table
          columns={columns}
          dataSource={data?.list || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: page,
            pageSize,
            total: data?.total || 0,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) },
          }}
        />
      </div>

      <Modal
        title={editingId ? '编辑公告' : '新建公告'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => { setModalOpen(false); setEditingId(null) }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input placeholder="请输入公告标题" />
          </Form.Item>
          <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入内容' }]}>
            <TextArea rows={6} placeholder="请输入公告内容" />
          </Form.Item>
          <Space size="large">
            <Form.Item name="type" label="类型">
              <Select options={typeOptions} style={{ width: 120 }} />
            </Form.Item>
            <Form.Item name="is_top" label="置顶" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}

export default Announcements

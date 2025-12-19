import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Modal, Form, Input, InputNumber, App, Popconfirm, Tag } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'
import dayjs from 'dayjs'

interface IPRecord {
  id: number
  ip: string
  reason: string
  expire_at: string | null
  created_at: string
}

const IPBlacklist = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()

  const { data, isLoading } = useQuery({
    queryKey: ['ipBlacklist', page, pageSize],
    queryFn: async () => {
      const response = await adminApi.getIPBlacklist({ page, page_size: pageSize })
      return response.data.data as { list: IPRecord[]; total: number }
    },
  })

  const addMutation = useMutation({
    mutationFn: (data: { ip: string; reason: string; duration: number }) => 
      adminApi.addIPBlacklist(data),
    onSuccess: () => {
      message.success('添加成功')
      setModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['ipBlacklist'] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '添加失败'),
  })

  const removeMutation = useMutation({
    mutationFn: (ip: string) => adminApi.removeIPBlacklist(ip),
    onSuccess: () => {
      message.success('移除成功')
      queryClient.invalidateQueries({ queryKey: ['ipBlacklist'] })
    },
  })

  const handleAdd = () => {
    form.resetFields()
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()
    addMutation.mutate(values)
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { 
      title: 'IP地址', 
      dataIndex: 'ip', 
      key: 'ip', 
      width: 150,
      render: (ip: string) => <code className="bg-gray-100 px-2 py-1 rounded">{ip}</code>
    },
    { title: '封禁原因', dataIndex: 'reason', key: 'reason', ellipsis: true },
    {
      title: '过期时间',
      dataIndex: 'expire_at',
      key: 'expire_at',
      width: 170,
      render: (time: string | null) => {
        if (!time) return <Tag color="red">永久封禁</Tag>
        const isExpired = dayjs(time).isBefore(dayjs())
        return (
          <span className={isExpired ? 'text-gray-400' : ''}>
            {dayjs(time).format('YYYY-MM-DD HH:mm')}
            {isExpired && <Tag className="ml-2">已过期</Tag>}
          </span>
        )
      },
    },
    {
      title: '添加时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: IPRecord) => (
        <Popconfirm title="确定移除？" onConfirm={() => removeMutation.mutate(record.ip)}>
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>移除</Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="glass-card p-4">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-semibold m-0">IP黑名单</h2>
            <p className="text-gray-500 text-sm m-0 mt-1">管理被封禁的IP地址，封禁后该IP无法访问系统</p>
          </div>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加IP
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
        title="添加IP到黑名单"
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={addMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item 
            name="ip" 
            label="IP地址" 
            rules={[
              { required: true, message: '请输入IP地址' },
              { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入有效的IP地址' }
            ]}
          >
            <Input placeholder="例如: 192.168.1.1" />
          </Form.Item>
          <Form.Item name="reason" label="封禁原因">
            <Input.TextArea rows={2} placeholder="请输入封禁原因（可选）" />
          </Form.Item>
          <Form.Item name="duration" label="封禁时长（分钟）" extra="留空或填0表示永久封禁">
            <InputNumber min={0} placeholder="0" style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default IPBlacklist

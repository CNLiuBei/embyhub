import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Table, Tag } from 'antd'
import { adminApi } from '../../services/api'
import dayjs from 'dayjs'

interface OperationLog {
  id: number
  admin_id: string
  admin_name: string
  action: string
  target: string
  detail: string
  ip: string
  created_at: string
}

const OperationLogs = () => {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const { data, isLoading } = useQuery({
    queryKey: ['operationLogs', page, pageSize],
    queryFn: async () => {
      const response = await adminApi.getOperationLogs({ page, page_size: pageSize })
      return response.data.data as { list: OperationLog[]; total: number }
    },
  })

  // 操作类型颜色映射
  const getActionColor = (action: string) => {
    if (action.includes('删除')) return 'red'
    if (action.includes('禁用')) return 'orange'
    if (action.includes('启用')) return 'green'
    if (action.includes('重置')) return 'purple'
    if (action.includes('设置') || action.includes('升级')) return 'blue'
    return 'default'
  }

  const columns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '管理员',
      dataIndex: 'admin_name',
      key: 'admin_name',
      width: 100,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (action: string) => <Tag color={getActionColor(action)}>{action}</Tag>,
    },
    {
      title: '目标',
      dataIndex: 'target',
      key: 'target',
      width: 280,
      ellipsis: true,
      render: (target: string) => (
        <code className="text-xs bg-gray-100 px-2 py-1 rounded">{target}</code>
      ),
    },
    {
      title: '详情',
      dataIndex: 'detail',
      key: 'detail',
      ellipsis: true,
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
      width: 130,
    },
  ]

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="glass-card p-4">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold m-0">操作日志</h2>
          <span className="text-gray-500 text-sm">
            记录管理员的所有操作行为，用于安全审计
          </span>
        </div>
      </div>

      <div className="glass-card p-4 flex-1 flex flex-col">
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
            showTotal: (total) => `共 ${total} 条记录`,
            onChange: (p, ps) => {
              setPage(p)
              setPageSize(ps)
            },
          }}
          scroll={{ x: 1000 }}
        />
      </div>
    </div>
  )
}

export default OperationLogs

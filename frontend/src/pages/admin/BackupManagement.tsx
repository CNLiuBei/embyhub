import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Space, App, Popconfirm, Tag } from 'antd'
import { CloudUploadOutlined, DownloadOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'

interface BackupInfo {
  filename: string
  size: number
  size_str: string
  created_at: string
}

const BackupManagement = () => {
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()
  const [creating, setCreating] = useState(false)

  const { data: backups, isLoading, refetch } = useQuery({
    queryKey: ['backups'],
    queryFn: async () => {
      const response = await adminApi.getBackupList()
      return response.data.data as BackupInfo[]
    },
  })

  const createMutation = useMutation({
    mutationFn: () => adminApi.createBackup(),
    onSuccess: () => {
      message.success('备份创建成功')
      queryClient.invalidateQueries({ queryKey: ['backups'] })
      setCreating(false)
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || '备份失败')
      setCreating(false)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (filename: string) => adminApi.deleteBackup(filename),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['backups'] })
    },
  })

  const handleCreate = () => {
    setCreating(true)
    createMutation.mutate()
  }

  const handleDownload = async (filename: string) => {
    try {
      const response = await adminApi.downloadBackup(filename)
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      link.click()
      window.URL.revokeObjectURL(url)
    } catch {
      message.error('下载失败')
    }
  }

  const handleRestore = (filename: string) => {
    modal.confirm({
      title: '危险操作',
      content: '恢复备份将覆盖当前所有数据，此操作不可逆！确定要继续吗？',
      okText: '确定恢复',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await adminApi.restoreBackup(filename)
          message.success('恢复成功，请重启服务')
        } catch (err: any) {
          message.error(err.response?.data?.message || '恢复失败')
        }
      },
    })
  }

  const columns = [
    { 
      title: '文件名', 
      dataIndex: 'filename', 
      key: 'filename',
      render: (filename: string) => <code className="text-sm">{filename}</code>
    },
    { title: '大小', dataIndex: 'size_str', key: 'size_str', width: 120 },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
    {
      title: '操作',
      key: 'action',
      width: 250,
      render: (_: unknown, record: BackupInfo) => (
        <Space>
          <Button 
            type="link" 
            size="small" 
            icon={<DownloadOutlined />}
            onClick={() => handleDownload(record.filename)}
          >
            下载
          </Button>
          <Button 
            type="link" 
            size="small"
            onClick={() => handleRestore(record.filename)}
          >
            恢复
          </Button>
          <Popconfirm title="确定删除？" onConfirm={() => deleteMutation.mutate(record.filename)}>
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
          <div>
            <h2 className="text-lg font-semibold m-0">数据备份</h2>
            <p className="text-gray-500 text-sm m-0 mt-1">备份和恢复数据库数据</p>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()}>刷新</Button>
            <Button 
              type="primary" 
              icon={<CloudUploadOutlined />} 
              onClick={handleCreate}
              loading={creating}
            >
              创建备份
            </Button>
          </Space>
        </div>
      </div>

      <div className="glass-card p-4 flex-1">
        <div className="mb-4">
          <Tag color="orange">提示：恢复备份会覆盖当前数据，请谨慎操作</Tag>
        </div>
        <Table
          columns={columns}
          dataSource={backups || []}
          rowKey="filename"
          loading={isLoading}
          pagination={false}
        />
      </div>
    </div>
  )
}

export default BackupManagement

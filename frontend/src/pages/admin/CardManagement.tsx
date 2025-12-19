import { useState } from 'react'
import { Table, Button, Form, InputNumber, Select, Modal, Input, Space, Row, Col, Tabs, Tag, Statistic, Dropdown, Typography, App } from 'antd'
import { PlusOutlined, DownloadOutlined, StopOutlined, CheckOutlined, CopyOutlined } from '@ant-design/icons'
import type { MenuProps } from 'antd'

const { Paragraph } = Typography
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { cardAdminApi } from '../../services/cardApi'
import type { Card, CardBatch, CardStats } from '../../types/card.types'

type CardItem = Card

const CardManagement = () => {
  const { message } = App.useApp()
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [page, setPage] = useState(1)
  const [batchPage, setBatchPage] = useState(1)
  const [filters, setFilters] = useState<{ batch_no?: string; card_type?: number; status?: number }>({})
  const [form] = Form.useForm()
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([])
  const queryClient = useQueryClient()

  // 获取统计
  const { data: stats } = useQuery<CardStats>({
    queryKey: ['cardStats'],
    queryFn: () => cardAdminApi.getStats(),
  })

  // 获取批次列表
  const { data: batchData } = useQuery({
    queryKey: ['cardBatches', batchPage],
    queryFn: () => cardAdminApi.getBatchList({ page: batchPage, page_size: 10 }),
  })

  // 获取卡密列表
  const { data: cardData, isLoading } = useQuery({
    queryKey: ['cards', page, filters],
    queryFn: () => cardAdminApi.getCardList({ page, page_size: 20, ...filters }),
  })

  // 创建批次
  const createMutation = useMutation({
    mutationFn: (data: { card_type: number; quantity: number; duration?: number; remark?: string }) => 
      cardAdminApi.createBatch(data),
    onSuccess: (data) => {
      message.success(`成功生成 ${data.codes.length} 张卡密`)
      setCreateModalVisible(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardBatches'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
      
      // 下载卡密
      const blob = new Blob([data.codes.join('\n')], { type: 'text/plain' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `cards_${data.batch.batch_no}.txt`
      a.click()
    },
    onError: () => message.error('生成失败'),
  })

  // 禁用卡密
  const disableMutation = useMutation({
    mutationFn: (id: number) => cardAdminApi.disableCard(id),
    onSuccess: () => {
      message.success('已禁用')
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
    },
  })

  // 启用卡密
  const enableMutation = useMutation({
    mutationFn: (id: number) => cardAdminApi.enableCard(id),
    onSuccess: () => {
      message.success('已启用')
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
    },
  })

  // 删除卡密
  const deleteMutation = useMutation({
    mutationFn: (id: number) => cardAdminApi.deleteCard(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
    },
  })

  // 批量禁用
  const batchDisableMutation = useMutation({
    mutationFn: (ids: number[]) => Promise.all(ids.map(id => cardAdminApi.disableCard(id))),
    onSuccess: () => {
      message.success('批量禁用成功')
      setSelectedRowKeys([])
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
    },
  })

  // 批量启用
  const batchEnableMutation = useMutation({
    mutationFn: (ids: number[]) => Promise.all(ids.map(id => cardAdminApi.enableCard(id))),
    onSuccess: () => {
      message.success('批量启用成功')
      setSelectedRowKeys([])
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
    },
  })

  // 批量删除
  const batchDeleteMutation = useMutation({
    mutationFn: (ids: number[]) => Promise.all(ids.map(id => cardAdminApi.deleteCard(id))),
    onSuccess: () => {
      message.success('批量删除成功')
      setSelectedRowKeys([])
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      queryClient.invalidateQueries({ queryKey: ['cardStats'] })
    },
  })

  // 批量删除确认
  const handleBatchDelete = () => {
    Modal.confirm({
      title: '确认批量删除',
      content: `确定要删除选中的 ${selectedRowKeys.length} 张卡密吗？此操作不可恢复！`,
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => batchDeleteMutation.mutate(selectedRowKeys),
    })
  }

  // 获取选中的卡密码
  const getSelectedCodes = () => {
    const cards = (cardData as { list: CardItem[]; total: number })?.list || []
    return cards.filter(c => selectedRowKeys.includes(c.id)).map(c => c.code)
  }

  // 批量复制卡密码
  const handleBatchCopy = async () => {
    const codes = getSelectedCodes()
    if (codes.length === 0) return
    try {
      await navigator.clipboard.writeText(codes.join('\n'))
      message.success(`已复制 ${codes.length} 个卡密码到剪贴板`)
    } catch {
      message.error('复制失败，请手动复制')
    }
  }

  // 批量导出卡密码
  const handleBatchExport = (format: 'txt' | 'csv') => {
    const cards = (cardData as { list: CardItem[]; total: number })?.list || []
    const selectedCards = cards.filter(c => selectedRowKeys.includes(c.id))
    if (selectedCards.length === 0) return
    
    let content = ''
    let filename = ''
    let mimeType = ''
    
    if (format === 'txt') {
      // TXT格式：序号 | 卡密码 | 类型 | 天数
      const rows = selectedCards.map((c, i) => {
        const typeMap: Record<number, string> = { 1: '月卡', 2: '季卡', 3: '半年卡', 4: '年卡' }
        const typeText = typeMap[c.card_type] || '未知'
        return `${i + 1}. ${c.code} | ${typeText} | ${c.duration}天`
      })
      content = rows.join('\n')
      filename = `selected_cards_${Date.now()}.txt`
      mimeType = 'text/plain'
    } else {
      // CSV格式：完整信息
      const header = '序号,卡密码,类型,天数,状态,批次号,使用者,使用时间,创建时间'
      const rows = selectedCards.map((c, i) => {
        const typeMap2: Record<number, string> = { 1: '月卡', 2: '季卡', 3: '半年卡', 4: '年卡' }
        const typeText = typeMap2[c.card_type] || '未知'
        const statusText = c.status === 0 ? '未使用' : c.status === 1 ? '已使用' : c.status === 2 ? '已过期' : '已禁用'
        const usedBy = c.used_by_name || c.used_by || ''
        const usedAt = c.used_at?.slice(0, 19).replace('T', ' ') || ''
        return `${i + 1},${c.code},${typeText},${c.duration},${statusText},${c.batch_no},${usedBy},${usedAt},${c.created_at?.slice(0, 19).replace('T', ' ')}`
      })
      content = [header, ...rows].join('\n')
      filename = `selected_cards_${Date.now()}.csv`
      mimeType = 'text/csv'
    }
    
    const blob = new Blob(['\ufeff' + content], { type: `${mimeType};charset=utf-8` })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
    message.success(`已导出 ${selectedCards.length} 个卡密`)
  }

  const handleDelete = (record: CardItem) => {
    Modal.confirm({
      title: '确认删除',
      content: (
        <div>
          <p>确定要删除此卡密吗？</p>
          <p style={{ color: '#999', fontSize: '12px', fontFamily: 'monospace' }}>卡密码：{record.code}</p>
          <p style={{ color: '#ff4d4f', fontSize: '12px' }}>此操作不可恢复！</p>
        </div>
      ),
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutate(record.id),
    })
  }

  const cardColumns = [
    { 
      title: '卡密码', 
      dataIndex: 'code', 
      key: 'code', 
      width: 320,
      render: (code: string) => (
        <Paragraph 
          copyable={{ 
            text: code,
            tooltips: ['复制', '已复制!'],
            icon: [<CopyOutlined key="copy" />, <CheckOutlined key="copied" style={{ color: '#52c41a' }} />]
          }}
          style={{ margin: 0, fontFamily: 'monospace', fontSize: '13px', wordBreak: 'break-all' }}
        >
          {code}
        </Paragraph>
      )
    },
    { title: '批次号', dataIndex: 'batch_no', key: 'batch_no', width: 150 },
    {
      title: '类型',
      dataIndex: 'card_type',
      key: 'card_type',
      width: 80,
      render: (v: number) => {
        const typeMap: Record<number, { color: string; text: string }> = {
          1: { color: 'blue', text: '月卡' },
          2: { color: 'cyan', text: '季卡' },
          3: { color: 'purple', text: '半年卡' },
          4: { color: 'gold', text: '年卡' },
        }
        return <Tag color={typeMap[v]?.color || 'default'}>{typeMap[v]?.text || '未知'}</Tag>
      },
    },
    { title: '天数', dataIndex: 'duration', key: 'duration', width: 80 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (v: number) => {
        const map: Record<number, { color: string; text: string }> = {
          0: { color: 'green', text: '未使用' },
          1: { color: 'default', text: '已使用' },
          2: { color: 'orange', text: '已过期' },
          3: { color: 'red', text: '已禁用' },
        }
        return <Tag color={map[v]?.color}>{map[v]?.text}</Tag>
      },
    },
    {
      title: '使用情况',
      key: 'usage',
      width: 180,
      render: (_: unknown, record: CardItem) => {
        if (record.status === 1 && (record.used_by_name || record.used_by)) {
          return (
            <div className="text-xs">
              <div className="text-gray-600">使用者: <span className="font-medium">{record.used_by_name || record.used_by}</span></div>
              {record.used_at && <div className="text-gray-400">{record.used_at?.slice(0, 19).replace('T', ' ')}</div>}
            </div>
          )
        }
        return <span className="text-gray-300">-</span>
      },
    },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 160, render: (v: string) => v?.slice(0, 19).replace('T', ' ') },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
      render: (_: unknown, record: CardItem) => (
        <Space>
          {record.status === 0 && (
            <Button size="small" danger icon={<StopOutlined />} onClick={() => disableMutation.mutate(record.id)}>
              禁用
            </Button>
          )}
          {record.status === 3 && (
            <Button size="small" type="primary" icon={<CheckOutlined />} onClick={() => enableMutation.mutate(record.id)}>
              启用
            </Button>
          )}
          <Button 
            size="small" 
            danger 
            type="text" 
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const batchColumns = [
    { title: '批次号', dataIndex: 'batch_no', key: 'batch_no' },
    {
      title: '类型',
      dataIndex: 'card_type',
      key: 'card_type',
      render: (v: number) => {
        const typeMap: Record<number, { color: string; text: string }> = {
          1: { color: 'blue', text: '月卡' },
          2: { color: 'cyan', text: '季卡' },
          3: { color: 'purple', text: '半年卡' },
          4: { color: 'gold', text: '年卡' },
        }
        return <Tag color={typeMap[v]?.color || 'default'}>{typeMap[v]?.text || '未知'}</Tag>
      },
    },
    { title: '天数', dataIndex: 'duration', key: 'duration' },
    { title: '数量', dataIndex: 'quantity', key: 'quantity' },
    { title: '已用', dataIndex: 'used_count', key: 'used_count' },
    { title: '创建者', dataIndex: 'created_by_name', key: 'created_by_name' },
    { title: '备注', dataIndex: 'remark', key: 'remark' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (v: string) => v?.slice(0, 19).replace('T', ' ') },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: CardBatch) => (
        <Dropdown menu={getExportMenu(record.batch_no)} placement="bottomRight">
          <Button size="small" icon={<DownloadOutlined />}>
            导出
          </Button>
        </Dropdown>
      ),
    },
  ]

  const handleExport = async (batchNo: string, format: 'csv' | 'excel' | 'codes' | 'report') => {
    try {
      const token = localStorage.getItem('access_token')
      let url = ''
      let filename = ''
      
      switch (format) {
        case 'csv':
          url = `/api/v1/admin/card/export/csv?batch_no=${batchNo}`
          filename = `cards_${batchNo}.csv`
          break
        case 'excel':
          url = `/api/v1/admin/card/export/excel?batch_no=${batchNo}`
          filename = `cards_${batchNo}.xlsx`
          break
        case 'codes':
          url = `/api/v1/admin/card/export/codes?batch_no=${batchNo}`
          filename = `codes_${batchNo}.txt`
          break
        case 'report':
          url = `/api/v1/admin/card/export/report?batch_no=${batchNo}`
          filename = `report_${batchNo}.xlsx`
          break
      }

      const response = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
      })
      
      if (!response.ok) throw new Error('导出失败')
      
      const blob = await response.blob()
      const downloadUrl = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = downloadUrl
      a.download = filename
      a.click()
      URL.revokeObjectURL(downloadUrl)
      
      message.success('导出成功')
    } catch {
      message.error('导出失败')
    }
  }

  const getExportMenu = (batchNo: string): MenuProps => ({
    items: [
      {
        key: 'csv',
        label: '导出CSV',
        icon: <DownloadOutlined />,
        onClick: () => handleExport(batchNo, 'csv'),
      },
      {
        key: 'excel',
        label: '导出Excel',
        icon: <DownloadOutlined />,
        onClick: () => handleExport(batchNo, 'excel'),
      },
      {
        key: 'codes',
        label: '导出卡密码',
        icon: <DownloadOutlined />,
        onClick: () => handleExport(batchNo, 'codes'),
      },
      {
        type: 'divider',
      },
      {
        key: 'report',
        label: '使用报告',
        icon: <DownloadOutlined />,
        onClick: () => handleExport(batchNo, 'report'),
      },
    ],
  })

  return (
    <div className="flex flex-col h-full gap-4">
      {/* 移动端统计卡片 */}
      <div className="md:hidden grid grid-cols-3 gap-2">
        <div className="glass-card p-3 text-center">
          <div className="text-lg font-bold text-blue-600">{stats?.total_cards || 0}</div>
          <div className="text-xs text-gray-500">总卡密</div>
        </div>
        <div className="glass-card p-3 text-center">
          <div className="text-lg font-bold text-green-600">{stats?.unused_cards || 0}</div>
          <div className="text-xs text-gray-500">未使用</div>
        </div>
        <div className="glass-card p-3 text-center">
          <div className="text-lg font-bold text-gray-600">{stats?.used_cards || 0}</div>
          <div className="text-xs text-gray-500">已使用</div>
        </div>
        <div className="glass-card p-3 text-center">
          <div className="text-lg font-bold text-orange-500">{stats?.expired_cards || 0}</div>
          <div className="text-xs text-gray-500">已过期</div>
        </div>
        <div className="glass-card p-3 text-center">
          <div className="text-lg font-bold text-red-500">{stats?.disabled_cards || 0}</div>
          <div className="text-xs text-gray-500">已禁用</div>
        </div>
        <div className="glass-card p-3 text-center">
          <div className="text-lg font-bold text-purple-600">{stats?.total_batches || 0}</div>
          <div className="text-xs text-gray-500">总批次</div>
        </div>
      </div>
      
      {/* 桌面端统计卡片 */}
      <Row gutter={16} className="hidden md:flex">
        <Col span={4}>
          <div className="glass-card p-4"><Statistic title="总卡密" value={stats?.total_cards || 0} /></div>
        </Col>
        <Col span={4}>
          <div className="glass-card p-4"><Statistic title="未使用" value={stats?.unused_cards || 0} valueStyle={{ color: '#52c41a' }} /></div>
        </Col>
        <Col span={4}>
          <div className="glass-card p-4"><Statistic title="已使用" value={stats?.used_cards || 0} /></div>
        </Col>
        <Col span={4}>
          <div className="glass-card p-4"><Statistic title="已过期" value={stats?.expired_cards || 0} valueStyle={{ color: '#faad14' }} /></div>
        </Col>
        <Col span={4}>
          <div className="glass-card p-4"><Statistic title="已禁用" value={stats?.disabled_cards || 0} valueStyle={{ color: '#ff4d4f' }} /></div>
        </Col>
        <Col span={4}>
          <div className="glass-card p-4"><Statistic title="总批次" value={stats?.total_batches || 0} /></div>
        </Col>
      </Row>

      <div className="glass-card p-4 flex-1">
        <div className="flex justify-between items-center mb-4">
          <span className="font-semibold">卡密管理</span>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
            生成卡密
          </Button>
        </div>
        <Tabs
          items={[
            {
              key: 'cards',
              label: '卡密列表',
              children: (
                <>
                  {/* 移动端筛选 */}
                  <div className="md:hidden mb-4 space-y-2">
                    <div className="flex gap-2">
                      <Select
                        placeholder="类型"
                        allowClear
                        className="flex-1"
                        onChange={(v) => setFilters({ ...filters, card_type: v })}
                        options={[
                          { value: 1, label: '月卡' },
                          { value: 2, label: '季卡' },
                          { value: 3, label: '半年卡' },
                          { value: 4, label: '年卡' },
                        ]}
                      />
                      <Select
                        placeholder="状态"
                        allowClear
                        className="flex-1"
                        onChange={(v) => setFilters({ ...filters, status: v })}
                        options={[
                          { value: 0, label: '未使用' },
                          { value: 1, label: '已使用' },
                          { value: 2, label: '已过期' },
                          { value: 3, label: '已禁用' },
                        ]}
                      />
                    </div>
                  </div>
                  
                  {/* 桌面端筛选和批量操作 - 同一行 */}
                  <div className="mb-4 hidden md:flex items-center justify-between gap-4">
                    <Space>
                      <Input
                        placeholder="批次号"
                        style={{ width: 150 }}
                        onChange={(e) => setFilters({ ...filters, batch_no: e.target.value })}
                      />
                      <Select
                        placeholder="类型"
                        allowClear
                        style={{ width: 120 }}
                        onChange={(v) => setFilters({ ...filters, card_type: v })}
                        options={[
                          { value: 1, label: '月卡' },
                          { value: 2, label: '季卡' },
                          { value: 3, label: '半年卡' },
                          { value: 4, label: '年卡' },
                        ]}
                      />
                      <Select
                        placeholder="状态"
                        allowClear
                        style={{ width: 120 }}
                        onChange={(v) => setFilters({ ...filters, status: v })}
                        options={[
                          { value: 0, label: '未使用' },
                          { value: 1, label: '已使用' },
                          { value: 2, label: '已过期' },
                          { value: 3, label: '已禁用' },
                        ]}
                      />
                    </Space>
                    {/* 批量操作按钮 - 右侧 */}
                    {selectedRowKeys.length > 0 && (
                      <Space>
                        <span className="text-gray-500">已选 <span className="font-bold text-blue-600">{selectedRowKeys.length}</span> 项</span>
                        <Button size="small" icon={<CopyOutlined />} onClick={handleBatchCopy}>复制卡密</Button>
                        <Dropdown menu={{
                          items: [
                            { key: 'txt', label: '导出TXT', onClick: () => handleBatchExport('txt') },
                            { key: 'csv', label: '导出CSV', onClick: () => handleBatchExport('csv') },
                          ]
                        }}>
                          <Button size="small" icon={<DownloadOutlined />}>导出选中</Button>
                        </Dropdown>
                        <Button size="small" danger icon={<StopOutlined />} loading={batchDisableMutation.isPending} onClick={() => batchDisableMutation.mutate(selectedRowKeys)}>批量禁用</Button>
                        <Button size="small" icon={<CheckOutlined />} loading={batchEnableMutation.isPending} onClick={() => batchEnableMutation.mutate(selectedRowKeys)}>批量启用</Button>
                        <Button size="small" danger type="primary" loading={batchDeleteMutation.isPending} onClick={handleBatchDelete}>批量删除</Button>
                        <Button size="small" onClick={() => setSelectedRowKeys([])}>取消</Button>
                      </Space>
                    )}
                  </div>
                  
                  {/* 移动端卡密列表 */}
                  <div className="md:hidden space-y-2">
                    {isLoading ? (
                      <div className="text-center py-8 text-gray-400">加载中...</div>
                    ) : ((cardData as { list: CardItem[]; total: number })?.list || []).length === 0 ? (
                      <div className="text-center py-8 text-gray-400">暂无数据</div>
                    ) : (
                      ((cardData as { list: CardItem[]; total: number })?.list || []).map((card) => (
                        <div 
                          key={card.id} 
                          className={`bg-white rounded-xl p-3 shadow-sm border ${selectedRowKeys.includes(card.id) ? 'border-blue-400 bg-blue-50' : 'border-gray-100'}`}
                          onClick={() => {
                            if (selectedRowKeys.includes(card.id)) {
                              setSelectedRowKeys(selectedRowKeys.filter(k => k !== card.id))
                            } else {
                              setSelectedRowKeys([...selectedRowKeys, card.id])
                            }
                          }}
                        >
                          <div className="flex items-start justify-between gap-2 mb-2">
                            <div className="flex items-start gap-2 flex-1 min-w-0">
                              <input 
                                type="checkbox" 
                                checked={selectedRowKeys.includes(card.id)}
                                onChange={() => {}}
                                className="mt-1 w-4 h-4 accent-blue-500"
                              />
                              <Paragraph 
                                copyable={{ text: card.code }}
                                className="!mb-0 font-mono text-xs break-all"
                                onClick={(e) => e.stopPropagation()}
                              >
                                {card.code}
                              </Paragraph>
                            </div>
                            <Tag color={
                              card.card_type === 1 ? 'blue' : 
                              card.card_type === 2 ? 'cyan' : 
                              card.card_type === 3 ? 'purple' : 'gold'
                            } className="shrink-0">
                              {card.card_type === 1 ? '月卡' : card.card_type === 2 ? '季卡' : card.card_type === 3 ? '半年卡' : '年卡'}
                            </Tag>
                          </div>
                          <div className="flex items-center justify-between text-xs pl-6">
                            <div className="flex items-center gap-2">
                              <Tag color={
                                card.status === 0 ? 'green' : 
                                card.status === 1 ? 'default' : 
                                card.status === 2 ? 'orange' : 'red'
                              }>
                                {card.status === 0 ? '未使用' : card.status === 1 ? '已使用' : card.status === 2 ? '已过期' : '已禁用'}
                              </Tag>
                              <span className="text-gray-400">{card.duration}天</span>
                            </div>
                            <Space size="small" onClick={(e) => e.stopPropagation()}>
                              {card.status === 0 && (
                                <Button size="small" danger type="text" onClick={() => disableMutation.mutate(card.id)}>禁用</Button>
                              )}
                              {card.status === 3 && (
                                <Button size="small" type="text" className="text-green-600" onClick={() => enableMutation.mutate(card.id)}>启用</Button>
                              )}
                            </Space>
                          </div>
                          {/* 使用情况 */}
                          {card.status === 1 && (card.used_by_name || card.used_by) && (
                            <div className="mt-2 pt-2 border-t border-gray-100 pl-6 text-xs text-gray-500">
                              使用者: <span className="text-gray-700">{card.used_by_name || card.used_by}</span>
                              {card.used_at && <span className="ml-2">{card.used_at?.slice(0, 16).replace('T', ' ')}</span>}
                            </div>
                          )}
                        </div>
                      ))
                    )}
                    {/* 移动端分页 */}
                    <div className="flex justify-center py-4">
                      <Space>
                        <Button size="small" disabled={page <= 1} onClick={() => setPage(p => Math.max(1, p - 1))}>上一页</Button>
                        <span className="text-gray-500 text-sm">{page} / {Math.ceil(((cardData as { list: CardItem[]; total: number })?.total || 0) / 20)}</span>
                        <Button size="small" disabled={page >= Math.ceil(((cardData as { list: CardItem[]; total: number })?.total || 0) / 20)} onClick={() => setPage(p => p + 1)}>下一页</Button>
                      </Space>
                    </div>
                  </div>
                  
                  {/* 移动端批量操作栏 */}
                  {selectedRowKeys.length > 0 && (
                    <div className="md:hidden mb-4 p-3 bg-blue-50 rounded-lg">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-gray-600">已选 <span className="font-bold text-blue-600">{selectedRowKeys.length}</span> 项</span>
                        <Button size="small" onClick={() => setSelectedRowKeys([])}>取消</Button>
                      </div>
                      <Space wrap size="small">
                        <Button size="small" icon={<CopyOutlined />} onClick={handleBatchCopy}>复制</Button>
                        <Button size="small" icon={<DownloadOutlined />} onClick={() => handleBatchExport('txt')}>导出</Button>
                        <Button size="small" danger icon={<StopOutlined />} onClick={() => batchDisableMutation.mutate(selectedRowKeys)}>禁用</Button>
                        <Button size="small" icon={<CheckOutlined />} onClick={() => batchEnableMutation.mutate(selectedRowKeys)}>启用</Button>
                        <Button size="small" danger type="primary" onClick={handleBatchDelete}>删除</Button>
                      </Space>
                    </div>
                  )}
                  
                  {/* 桌面端表格 */}
                  <div className="hidden md:block">
                    <Table
                      rowKey="id"
                      columns={cardColumns}
                      dataSource={(cardData as { list: CardItem[]; total: number })?.list || []}
                      loading={isLoading}
                      rowSelection={{
                        selectedRowKeys,
                        onChange: (keys) => setSelectedRowKeys(keys as number[]),
                      }}
                      pagination={{
                        current: page,
                        total: (cardData as { list: CardItem[]; total: number })?.total || 0,
                        pageSize: 20,
                        onChange: setPage,
                      }}
                    />
                  </div>
                </>
              ),
            },
            {
              key: 'batches',
              label: '批次列表',
              children: (
                <>
                  {/* 移动端批次列表 */}
                  <div className="md:hidden space-y-2">
                    {((batchData as { list: CardBatch[]; total: number })?.list || []).map((batch) => (
                      <div key={batch.id} className="bg-white rounded-xl p-3 shadow-sm border border-gray-100">
                        <div className="flex items-center justify-between mb-2">
                          <span className="font-mono text-sm font-medium">{batch.batch_no}</span>
                          <Tag color={
                            batch.card_type === 1 ? 'blue' : 
                            batch.card_type === 2 ? 'cyan' : 
                            batch.card_type === 3 ? 'purple' : 'gold'
                          }>
                            {batch.card_type === 1 ? '月卡' : batch.card_type === 2 ? '季卡' : batch.card_type === 3 ? '半年卡' : '年卡'}
                          </Tag>
                        </div>
                        <div className="flex items-center justify-between text-xs text-gray-500">
                          <div className="flex gap-3">
                            <span>数量: {batch.quantity}</span>
                            <span>已用: {batch.used_count}</span>
                            <span>{batch.duration}天</span>
                          </div>
                          <Dropdown menu={getExportMenu(batch.batch_no)} placement="bottomRight">
                            <Button size="small" type="text" icon={<DownloadOutlined />} />
                          </Dropdown>
                        </div>
                      </div>
                    ))}
                    <div className="flex justify-center py-4">
                      <Space>
                        <Button size="small" disabled={batchPage <= 1} onClick={() => setBatchPage(p => Math.max(1, p - 1))}>上一页</Button>
                        <span className="text-gray-500 text-sm">{batchPage}</span>
                        <Button size="small" onClick={() => setBatchPage(p => p + 1)}>下一页</Button>
                      </Space>
                    </div>
                  </div>
                  
                  {/* 桌面端表格 */}
                  <div className="hidden md:block">
                    <Table
                      rowKey="id"
                      columns={batchColumns}
                      dataSource={(batchData as { list: CardBatch[]; total: number })?.list || []}
                      pagination={{
                        current: batchPage,
                        total: (batchData as { list: CardBatch[]; total: number })?.total || 0,
                        pageSize: 10,
                        onChange: setBatchPage,
                      }}
                    />
                  </div>
                </>
              ),
            },
          ]}
        />
      </div>

      <Modal
        title="生成卡密"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending}
        width={600}
      >
        <Form 
          form={form} 
          layout="vertical" 
          onFinish={(values) => createMutation.mutate(values)}
          initialValues={{ code_prefix: 'True' }}
        >
          <Form.Item name="card_type" label="卡密类型" rules={[{ required: true, message: '请选择卡密类型' }]}>
            <Select
              placeholder="请选择卡密类型"
              options={[
                { value: 1, label: '月卡 (30天)' },
                { value: 2, label: '季卡 (90天)' },
                { value: 3, label: '半年卡 (180天)' },
                { value: 4, label: '年卡 (365天)' },
              ]}
            />
          </Form.Item>
          
          <Form.Item name="quantity" label="生成数量" rules={[{ required: true, message: '请输入生成数量' }]}>
            <InputNumber 
              min={1} 
              max={1000} 
              style={{ width: '100%' }} 
              placeholder="1-1000"
            />
          </Form.Item>
          
          <Form.Item name="duration" label="有效天数" help="不填则使用默认值(月卡30天/季卡90天/半年卡180天/年卡365天)">
            <InputNumber min={1} style={{ width: '100%' }} placeholder="自定义天数" />
          </Form.Item>

          <div style={{ background: '#f5f5f5', padding: '16px', borderRadius: '4px', marginBottom: '16px' }}>
            <div style={{ fontWeight: 'bold', marginBottom: '12px', color: '#1890ff' }}>🔑 卡密格式配置</div>
            
            <Form.Item 
              name="code_prefix" 
              label="卡密前缀" 
              rules={[
                { required: true, message: '请输入卡密前缀' },
                { pattern: /^[A-Za-z0-9]+$/, message: '只能包含字母和数字' },
                { max: 10, message: '前缀不能超过10个字符' }
              ]}
            >
              <Space.Compact style={{ width: '100%' }}>
                <Input 
                  placeholder="True" 
                  style={{ flex: 1 }}
                />
                <Select
                  defaultValue="True"
                  variant="filled"
                  style={{ width: 100 }}
                  onChange={(value) => form.setFieldValue('code_prefix', value)}
                  options={[
                    { value: 'True', label: 'True' },
                    { value: 'VIP', label: 'VIP' },
                    { value: 'SVIP', label: 'SVIP' },
                    { value: 'Premium', label: 'Premium' },
                  ]}
                />
              </Space.Compact>
            </Form.Item>
            
            <div style={{ fontSize: '12px', color: '#8c8c8c', padding: '12px', background: '#fff', borderRadius: '4px' }}>
              <div style={{ marginBottom: '8px' }}>📝 <strong>卡密格式说明：</strong></div>
              <div>• 格式：<code style={{ color: '#1890ff', fontSize: '13px' }}>前缀-24位纯字符</code></div>
              <div style={{ marginTop: '4px' }}>• 示例：<code style={{ color: '#1890ff', fontSize: '13px' }}>True-ABCDEFGHJKLMNPQRSTUWXYZ2</code></div>
              <div style={{ marginTop: '4px', color: '#999' }}>（固定24位大写字母和数字，去除易混淆字符0O1I）</div>
            </div>
          </div>
          
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} placeholder="选填，用于标识此批次" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default CardManagement

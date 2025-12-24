import { useState } from 'react'
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
  Statistic,
  Row,
  Col,
  Popconfirm,
  Tabs,
  Tooltip,
  Select,
  Divider,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TrophyOutlined,
  RiseOutlined,
  FallOutlined,
  UserOutlined,
  GiftOutlined,
  CopyOutlined,
  ExportOutlined,
  CreditCardOutlined,
  BellOutlined,
  MailOutlined,
  TeamOutlined,
  CrownOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { adminApi, pointsCardApi } from '../../services/api'
import dayjs from 'dayjs'

interface PointsStats {
  total_points: number
  today_issued: number
  today_consumed: number
  today_sign_in: number
}

interface ExchangeRule {
  id: number
  name: string
  points: number
  member_days: number
  description: string
  enabled: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

interface PointsCardBatch {
  id: number
  batch_no: string
  points: number
  quantity: number
  used_count: number
  remark: string
  created_by: string
  created_at: string
}

interface PointsCard {
  id: number
  code: string
  batch_no: string
  points: number
  status: number
  used_by: string
  used_at: string
  created_at: string
}

interface PointsCardStats {
  total_batches: number
  total_cards: number
  used_cards: number
  unused_cards: number
  total_points: number
  used_points: number
  unused_points: number
}

interface PointsGiftRule {
  id: number
  name: string
  rule_type: number
  points: number
  target_type: string
  member_level: number | null
  execute_time: string
  execute_day: number
  execute_month: number
  send_notification: boolean
  notification_title: string
  notification_body: string
  enabled: boolean
  last_execute_at: string | null
  next_execute_at: string | null
  created_at: string
}

interface PointsGiftLog {
  id: number
  rule_id: number
  rule_name: string
  points: number
  total_users: number
  success_count: number
  failed_count: number
  execute_at: string
}

const PointsManagement = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState('rules')

  // 兑换规则相关状态
  const [ruleModalOpen, setRuleModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<ExchangeRule | null>(null)
  const [ruleForm] = Form.useForm()

  // 赠送积分相关状态
  const [giftForm] = Form.useForm()

  // 积分卡相关状态
  const [cardModalOpen, setCardModalOpen] = useState(false)
  const [exportModalOpen, setExportModalOpen] = useState(false)
  const [exportedCodes, setExportedCodes] = useState<string[]>([])
  const [exportBatchNo, setExportBatchNo] = useState('')
  const [batchPage, setBatchPage] = useState(1)
  const [cardPage, setCardPage] = useState(1)
  const [selectedBatchNo, setSelectedBatchNo] = useState<string>('')
  const [cardStatusFilter, setCardStatusFilter] = useState<number | undefined>(undefined)
  const [cardKeyword, setCardKeyword] = useState<string>('')
  const [cardForm] = Form.useForm()

  // 自动赠送规则相关状态
  const [giftRuleModalOpen, setGiftRuleModalOpen] = useState(false)
  const [editingGiftRule, setEditingGiftRule] = useState<PointsGiftRule | null>(null)
  const [giftRuleForm] = Form.useForm()
  const [giftLogPage, setGiftLogPage] = useState(1)

  // ========== 积分统计 ==========
  const { data: stats } = useQuery({
    queryKey: ['pointsStats'],
    queryFn: async () => {
      const res = await adminApi.getPointsStats()
      return res.data.data as PointsStats
    },
  })

  // ========== 兑换规则 ==========
  const { data: rules, isLoading: rulesLoading } = useQuery({
    queryKey: ['pointsExchangeRules'],
    queryFn: async () => {
      const res = await adminApi.getPointsExchangeRules()
      return res.data.data as ExchangeRule[]
    },
  })

  const createRuleMutation = useMutation({
    mutationFn: (data: { name: string; points: number; member_days: number; description?: string; sort_order?: number }) =>
      adminApi.createPointsExchangeRule(data),
    onSuccess: () => {
      message.success('创建成功')
      setRuleModalOpen(false)
      ruleForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['pointsExchangeRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '创建失败')
    },
  })

  const updateRuleMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<ExchangeRule> }) =>
      adminApi.updatePointsExchangeRule(id, data),
    onSuccess: () => {
      message.success('更新成功')
      setRuleModalOpen(false)
      setEditingRule(null)
      ruleForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['pointsExchangeRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '更新失败')
    },
  })

  const deleteRuleMutation = useMutation({
    mutationFn: (id: number) => adminApi.deletePointsExchangeRule(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['pointsExchangeRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  // 批量赠送积分
  const giftPointsMutation = useMutation({
    mutationFn: (data: {
      points: number
      remark?: string
      target_type?: string
      member_level?: number
      role?: number
      send_notification?: boolean
      notification_title?: string
      notification_body?: string
      send_email?: boolean
      email_title?: string
      email_body?: string
    }) => adminApi.giftPointsToAll(data),
    onSuccess: (res) => {
      const data = res.data.data
      let msg = `赠送完成！成功 ${data.success_count} 人，失败 ${data.failed_count} 人`
      if (data.notification_sent > 0) {
        msg += `，已发送 ${data.notification_sent} 条站内通知`
      }
      message.success(msg)
      giftForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['pointsStats'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '赠送失败')
    },
  })

  // ========== 积分卡 ==========
  const { data: cardStats } = useQuery({
    queryKey: ['pointsCardStats'],
    queryFn: async () => {
      const res = await pointsCardApi.getStats()
      return res.data.data as PointsCardStats
    },
  })

  const { data: batchData, isLoading: batchLoading } = useQuery({
    queryKey: ['pointsCardBatches', batchPage],
    queryFn: async () => {
      const res = await pointsCardApi.getBatchList({ page: batchPage, page_size: 10 })
      return res.data.data as { list: PointsCardBatch[]; total: number }
    },
  })

  const { data: cardData, isLoading: cardLoading } = useQuery({
    queryKey: ['pointsCards', cardPage, selectedBatchNo, cardStatusFilter, cardKeyword],
    queryFn: async () => {
      const res = await pointsCardApi.getCardList({
        page: cardPage,
        page_size: 15,
        batch_no: selectedBatchNo || undefined,
        status: cardStatusFilter,
        keyword: cardKeyword || undefined,
      })
      return res.data.data as { list: PointsCard[]; total: number }
    },
  })

  const createCardMutation = useMutation({
    mutationFn: (data: { points: number; quantity: number; remark?: string }) =>
      pointsCardApi.createBatch(data),
    onSuccess: (res) => {
      message.success('创建成功')
      setCardModalOpen(false)
      cardForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['pointsCardBatches'] })
      queryClient.invalidateQueries({ queryKey: ['pointsCardStats'] })
      const batch = res.data.data
      handleExport(batch.batch_no)
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '创建失败')
    },
  })

  const disableCardMutation = useMutation({
    mutationFn: (id: number) => pointsCardApi.disableCard(id),
    onSuccess: () => {
      message.success('禁用成功')
      queryClient.invalidateQueries({ queryKey: ['pointsCards'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '禁用失败')
    },
  })

  const enableCardMutation = useMutation({
    mutationFn: (id: number) => pointsCardApi.enableCard(id),
    onSuccess: () => {
      message.success('启用成功')
      queryClient.invalidateQueries({ queryKey: ['pointsCards'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '启用失败')
    },
  })

  const deleteBatchMutation = useMutation({
    mutationFn: (batchNo: string) => pointsCardApi.deleteBatch(batchNo),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['pointsCardBatches'] })
      queryClient.invalidateQueries({ queryKey: ['pointsCardStats'] })
      queryClient.invalidateQueries({ queryKey: ['pointsCards'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  const deleteCardMutation = useMutation({
    mutationFn: (id: number) => pointsCardApi.deleteCard(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['pointsCards'] })
      queryClient.invalidateQueries({ queryKey: ['pointsCardStats'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  // ========== 自动赠送规则 ==========
  const { data: giftRules, isLoading: giftRulesLoading } = useQuery({
    queryKey: ['pointsGiftRules'],
    queryFn: async () => {
      const res = await adminApi.getPointsGiftRules()
      return res.data.data as PointsGiftRule[]
    },
  })

  const { data: giftLogsData } = useQuery({
    queryKey: ['pointsGiftLogs', giftLogPage],
    queryFn: async () => {
      const res = await adminApi.getPointsGiftLogs({ page: giftLogPage, page_size: 10 })
      return res.data.data as { list: PointsGiftLog[]; total: number }
    },
  })

  const createGiftRuleMutation = useMutation({
    mutationFn: (data: Parameters<typeof adminApi.createPointsGiftRule>[0]) =>
      adminApi.createPointsGiftRule(data),
    onSuccess: () => {
      message.success('创建成功')
      setGiftRuleModalOpen(false)
      giftRuleForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['pointsGiftRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '创建失败')
    },
  })

  const updateGiftRuleMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Parameters<typeof adminApi.updatePointsGiftRule>[1] }) =>
      adminApi.updatePointsGiftRule(id, data),
    onSuccess: () => {
      message.success('更新成功')
      setGiftRuleModalOpen(false)
      setEditingGiftRule(null)
      giftRuleForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['pointsGiftRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '更新失败')
    },
  })

  const deleteGiftRuleMutation = useMutation({
    mutationFn: (id: number) => adminApi.deletePointsGiftRule(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['pointsGiftRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '删除失败')
    },
  })

  const toggleGiftRuleMutation = useMutation({
    mutationFn: (id: number) => adminApi.togglePointsGiftRule(id),
    onSuccess: () => {
      message.success('操作成功')
      queryClient.invalidateQueries({ queryKey: ['pointsGiftRules'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '操作失败')
    },
  })

  const executeGiftRuleMutation = useMutation({
    mutationFn: (id: number) => adminApi.executePointsGiftRule(id),
    onSuccess: (res) => {
      const data = res.data.data
      message.success(`执行完成！成功 ${data.success_count} 人，失败 ${data.failed_count} 人`)
      queryClient.invalidateQueries({ queryKey: ['pointsGiftRules'] })
      queryClient.invalidateQueries({ queryKey: ['pointsGiftLogs'] })
      queryClient.invalidateQueries({ queryKey: ['pointsStats'] })
    },
    onError: (err: Error & { response?: { data?: { message?: string } } }) => {
      message.error(err.response?.data?.message || '执行失败')
    },
  })

  // ========== 处理函数 ==========
  const handleAddRule = () => {
    setEditingRule(null)
    ruleForm.resetFields()
    ruleForm.setFieldsValue({ enabled: true, sort_order: 0 })
    setRuleModalOpen(true)
  }

  const handleEditRule = (rule: ExchangeRule) => {
    setEditingRule(rule)
    ruleForm.setFieldsValue(rule)
    setRuleModalOpen(true)
  }

  const handleRuleSubmit = async () => {
    try {
      const values = await ruleForm.validateFields()
      if (editingRule) {
        updateRuleMutation.mutate({ id: editingRule.id, data: values })
      } else {
        createRuleMutation.mutate(values)
      }
    } catch {
      // validation error
    }
  }

  const handleToggleEnabled = (rule: ExchangeRule) => {
    updateRuleMutation.mutate({ id: rule.id, data: { enabled: !rule.enabled } })
  }

  const handleGiftPoints = async () => {
    try {
      const values = await giftForm.validateFields()
      const targetTypeMap: Record<string, string> = {
        all: '所有用户',
        member: '会员用户',
        non_member: '非会员用户',
        role: '指定角色用户',
      }
      const targetDesc = targetTypeMap[values.target_type || 'all'] || '所有用户'
      Modal.confirm({
        title: '确认赠送',
        content: `确定要给【${targetDesc}】赠送 ${values.points} 积分吗？此操作不可撤销！`,
        okText: '确认赠送',
        cancelText: '取消',
        onOk: () => giftPointsMutation.mutate(values),
      })
    } catch {
      // validation error
    }
  }

  const handleCreateCard = async () => {
    try {
      const values = await cardForm.validateFields()
      createCardMutation.mutate(values)
    } catch {
      // validation error
    }
  }

  // 自动赠送规则处理函数
  const handleAddGiftRule = () => {
    setEditingGiftRule(null)
    giftRuleForm.resetFields()
    giftRuleForm.setFieldsValue({ rule_type: 1, target_type: 'all', execute_time: '08:00', send_notification: true, enabled: true })
    setGiftRuleModalOpen(true)
  }

  const handleEditGiftRule = (rule: PointsGiftRule) => {
    setEditingGiftRule(rule)
    giftRuleForm.setFieldsValue(rule)
    setGiftRuleModalOpen(true)
  }

  const handleGiftRuleSubmit = async () => {
    try {
      const values = await giftRuleForm.validateFields()
      if (editingGiftRule) {
        updateGiftRuleMutation.mutate({ id: editingGiftRule.id, data: values })
      } else {
        createGiftRuleMutation.mutate(values)
      }
    } catch {
      // validation error
    }
  }

  const getRuleTypeName = (type: number) => {
    const map: Record<number, string> = { 1: '每日', 2: '每周', 3: '每月', 4: '每年' }
    return map[type] || '未知'
  }

  const getTargetTypeName = (type: string) => {
    const map: Record<string, string> = { all: '所有用户', member: '会员用户', non_member: '非会员用户' }
    return map[type] || type
  }

  const handleExport = async (batchNo: string) => {
    try {
      const res = await pointsCardApi.exportCards(batchNo)
      const data = res.data.data
      setExportedCodes(data.codes)
      setExportBatchNo(batchNo)
      setExportModalOpen(true)
    } catch (err: unknown) {
      const error = err as Error & { response?: { data?: { message?: string } } }
      message.error(error.response?.data?.message || '导出失败')
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success('已复制')
  }

  const copyAllCodes = () => {
    navigator.clipboard.writeText(exportedCodes.join('\n'))
    message.success('已复制所有卡密')
  }

  const downloadCodes = () => {
    const content = exportedCodes.join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `积分卡密_${exportBatchNo}.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  // ========== 表格列定义 ==========
  const ruleColumns = [
    { title: '排序', dataIndex: 'sort_order', key: 'sort_order', width: 60 },
    { title: '规则名称', dataIndex: 'name', key: 'name', render: (name: string) => <span className="font-medium">{name}</span> },
    { title: '所需积分', dataIndex: 'points', key: 'points', render: (points: number) => <span className="text-orange-500 font-bold">{points}</span> },
    { title: '兑换天数', dataIndex: 'member_days', key: 'member_days', render: (days: number) => <Tag color="blue">{days} 天</Tag> },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    {
      title: '状态', dataIndex: 'enabled', key: 'enabled',
      render: (enabled: boolean, record: ExchangeRule) => (
        <Switch checked={enabled} onChange={() => handleToggleEnabled(record)} size="small" />
      ),
    },
    {
      title: '操作', key: 'action', width: 120,
      render: (_: unknown, record: ExchangeRule) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditRule(record)} />
          <Popconfirm title="确定删除？" onConfirm={() => deleteRuleMutation.mutate(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const batchColumns = [
    { title: '批次号', dataIndex: 'batch_no', key: 'batch_no', render: (v: string) => <span className="font-mono text-blue-600">{v}</span> },
    { title: '面值', dataIndex: 'points', key: 'points', render: (v: number) => <span className="text-orange-500 font-bold">{v} 积分</span> },
    { title: '数量', dataIndex: 'quantity', key: 'quantity', render: (v: number) => `${v} 张` },
    {
      title: '使用情况', dataIndex: 'used_count', key: 'used_count',
      render: (used: number, record: PointsCardBatch) => {
        const percent = record.quantity > 0 ? Math.round((used / record.quantity) * 100) : 0
        return (
          <div className="flex items-center gap-2">
            <span>{used}/{record.quantity}</span>
            <Tag color={percent === 100 ? 'default' : percent > 50 ? 'orange' : 'green'}>{percent}%</Tag>
          </div>
        )
      },
    },
    { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true, render: (v: string) => v || '-' },
    { title: '创建者', dataIndex: 'created_by_name', key: 'created_by_name', render: (v: string) => v || '-' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm') },
    {
      title: '操作', key: 'action', width: 200,
      render: (_: unknown, record: PointsCardBatch) => (
        <Space size="small">
          <Button type="link" size="small" icon={<ExportOutlined />} onClick={() => handleExport(record.batch_no)}>导出</Button>
          <Button type="link" size="small" onClick={() => { setSelectedBatchNo(record.batch_no); setActiveTab('cards') }}>查看</Button>
          {record.used_count === 0 && (
            <Popconfirm title="确定删除该批次？" description="将同时删除该批次下所有卡密" onConfirm={() => deleteBatchMutation.mutate(record.batch_no)}>
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  const cardColumns = [
    {
      title: '卡密', dataIndex: 'code', key: 'code', width: 180,
      render: (code: string) => (
        <Space>
          <span className="font-mono">{code}</span>
          <Tooltip title="复制"><CopyOutlined className="text-gray-400 hover:text-blue-500 cursor-pointer" onClick={() => copyToClipboard(code)} /></Tooltip>
        </Space>
      ),
    },
    { title: '批次号', dataIndex: 'batch_no', key: 'batch_no', render: (v: string) => <span className="font-mono text-blue-600 text-xs">{v}</span> },
    { title: '积分', dataIndex: 'points', key: 'points', width: 80, render: (v: number) => <span className="text-orange-500 font-bold">{v}</span> },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 90,
      render: (status: number) => {
        const map: Record<number, { color: string; text: string }> = { 0: { color: 'green', text: '未使用' }, 1: { color: 'default', text: '已使用' }, 3: { color: 'red', text: '已禁用' } }
        const s = map[status] || { color: 'default', text: '未知' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    { title: '使用者', dataIndex: 'used_by', key: 'used_by', render: (v: string) => v || '-' },
    { title: '使用时间', dataIndex: 'used_at', key: 'used_at', render: (t: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm') },
    {
      title: '操作', key: 'action', width: 120,
      render: (_: unknown, record: PointsCard) => {
        if (record.status === 1) return <span className="text-gray-400 text-xs">已使用</span>
        return (
          <Space size="small">
            {record.status === 3 ? (
              <Popconfirm title="确定启用该卡密？" onConfirm={() => enableCardMutation.mutate(record.id)}>
                <Button type="link" size="small">启用</Button>
              </Popconfirm>
            ) : (
              <Popconfirm title="确定禁用该卡密？" onConfirm={() => disableCardMutation.mutate(record.id)}>
                <Button type="link" size="small" danger>禁用</Button>
              </Popconfirm>
            )}
            <Popconfirm title="确定删除该卡密？" onConfirm={() => deleteCardMutation.mutate(record.id)}>
              <Button type="link" size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Space>
        )
      },
    },
  ]

  return (
    <div className="space-y-4">
      {/* 统计卡片 */}
      <Row gutter={16}>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic title="系统总积分" value={stats?.total_points || 0} prefix={<TrophyOutlined className="text-orange-500" />} valueStyle={{ color: '#f97316' }} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic title="今日发放" value={stats?.today_issued || 0} prefix={<RiseOutlined className="text-green-500" />} valueStyle={{ color: '#22c55e' }} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic title="今日消费" value={stats?.today_consumed || 0} prefix={<FallOutlined className="text-red-500" />} valueStyle={{ color: '#ef4444' }} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic title="今日签到" value={stats?.today_sign_in || 0} prefix={<UserOutlined className="text-blue-500" />} valueStyle={{ color: '#3b82f6' }} suffix="人" />
          </Card>
        </Col>
      </Row>

      {/* 主内容 Tabs */}
      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'rules',
              label: <span><GiftOutlined /> 兑换规则</span>,
              children: (
                <div className="space-y-4">
                  <div className="flex justify-between items-center">
                    <span className="text-gray-500 text-sm">用户可用积分兑换会员时长</span>
                    <Button type="primary" size="small" icon={<PlusOutlined />} onClick={handleAddRule}>添加规则</Button>
                  </div>
                  <Table columns={ruleColumns} dataSource={rules || []} rowKey="id" loading={rulesLoading} pagination={false} size="small" />
                  
                  {/* 积分说明 */}
                  <div className="mt-4 pt-4 border-t border-gray-100">
                    <div className="text-gray-600 font-medium mb-3">积分获取说明</div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                      <div className="p-3 bg-green-50 rounded-lg">
                        <div className="font-medium text-green-700 text-sm mb-1">📅 每日签到</div>
                        <div className="text-gray-600 text-xs">第1天5分，每天+1，第6天起10分</div>
                      </div>
                      <div className="p-3 bg-blue-50 rounded-lg">
                        <div className="font-medium text-blue-700 text-sm mb-1">👥 邀请好友</div>
                        <div className="text-gray-600 text-xs">每成功邀请1人：30积分</div>
                      </div>
                      <div className="p-3 bg-purple-50 rounded-lg">
                        <div className="font-medium text-purple-700 text-sm mb-1">🎁 卡密充值</div>
                        <div className="text-gray-600 text-xs">使用积分卡密充值积分</div>
                      </div>
                    </div>
                  </div>
                </div>
              ),
            },
            {
              key: 'batches',
              label: <span><CreditCardOutlined /> 批次管理</span>,
              children: (
                <div className="space-y-4">
                  {/* 统计和操作 */}
                  <div className="flex justify-between items-center">
                    <Space size="large">
                      <Statistic title="总批次" value={cardStats?.total_batches || 0} valueStyle={{ fontSize: 16 }} />
                      <Statistic title="总卡密" value={cardStats?.total_cards || 0} valueStyle={{ fontSize: 16 }} />
                      <Statistic title="已使用" value={cardStats?.used_cards || 0} valueStyle={{ fontSize: 16, color: '#22c55e' }} />
                      <Statistic title="已发放积分" value={cardStats?.used_points || 0} valueStyle={{ fontSize: 16, color: '#f97316' }} />
                    </Space>
                    <Button type="primary" icon={<PlusOutlined />} onClick={() => setCardModalOpen(true)}>生成积分卡</Button>
                  </div>

                  {/* 批次列表 */}
                  <Table
                    columns={batchColumns}
                    dataSource={batchData?.list || []}
                    rowKey="id"
                    loading={batchLoading}
                    size="small"
                    pagination={{ current: batchPage, pageSize: 10, total: batchData?.total || 0, onChange: setBatchPage, size: 'small', showTotal: (t) => `共${t}条` }}
                  />

                  {/* 说明 */}
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-gray-600">
                    <div className="font-medium text-blue-700 mb-1">📋 批次管理说明：</div>
                    <ul className="list-disc list-inside space-y-0.5">
                      <li>每次生成积分卡会创建一个新批次，批次号格式为 PB+时间戳</li>
                      <li>点击"导出"可以导出该批次所有卡密，支持复制和下载</li>
                      <li>如需查看或管理单张卡密，请切换到"卡密管理"标签页</li>
                    </ul>
                  </div>
                </div>
              ),
            },
            {
              key: 'cards',
              label: <span><CreditCardOutlined /> 卡密管理</span>,
              children: (
                <div className="space-y-4">
                  {/* 筛选条件 */}
                  <div className="bg-gradient-to-r from-gray-50 to-blue-50/30 rounded-xl p-5 border border-gray-100">
                    <div className="flex flex-wrap items-center gap-5">
                      <div className="flex items-center gap-2">
                        <span className="text-gray-600 text-sm font-medium">卡密查询</span>
                        <Input.Search
                          placeholder="输入卡密搜索"
                          allowClear
                          className="rounded-lg"
                          style={{ width: 220 }}
                          value={cardKeyword}
                          onChange={(e) => setCardKeyword(e.target.value)}
                          onSearch={() => setCardPage(1)}
                        />
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-gray-600 text-sm font-medium">批次</span>
                        <Select
                          allowClear
                          placeholder="选择批次"
                          className="rounded-lg"
                          style={{ width: 200 }}
                          value={selectedBatchNo || undefined}
                          onChange={(v) => { setSelectedBatchNo(v || ''); setCardPage(1) }}
                        >
                          {(batchData?.list || []).map((b) => (
                            <Select.Option key={b.batch_no} value={b.batch_no}>
                              {b.batch_no} ({b.points}积分)
                            </Select.Option>
                          ))}
                        </Select>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-gray-600 text-sm font-medium">状态</span>
                        <Space size="small">
                          {[
                            { value: undefined, label: '全部' },
                            { value: 0, label: '未使用' },
                            { value: 1, label: '已使用' },
                            { value: 3, label: '已禁用' },
                          ].map((item) => (
                            <Button
                              key={String(item.value)}
                              type={cardStatusFilter === item.value ? 'primary' : 'default'}
                              size="small"
                              className="rounded-lg"
                              onClick={() => { setCardStatusFilter(item.value); setCardPage(1) }}
                            >
                              {item.label}
                            </Button>
                          ))}
                        </Space>
                      </div>
                    </div>
                  </div>

                  {/* 卡密列表 */}
                  <Table
                    columns={cardColumns}
                    dataSource={cardData?.list || []}
                    rowKey="id"
                    loading={cardLoading}
                    size="small"
                    pagination={{
                      current: cardPage,
                      pageSize: 15,
                      total: cardData?.total || 0,
                      onChange: setCardPage,
                      size: 'small',
                      showTotal: (t) => `共${t}条`,
                      showSizeChanger: false,
                    }}
                  />

                  {/* 说明 */}
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-gray-600">
                    <div className="font-medium text-blue-700 mb-1">📋 卡密管理说明：</div>
                    <ul className="list-disc list-inside space-y-0.5">
                      <li>卡密格式：JF-XXXXXXXXXXXXXXXX（JF前缀+16位字符）</li>
                      <li>支持卡密模糊查询，输入部分卡密即可搜索</li>
                      <li>未使用的卡密可以禁用，禁用后用户无法兑换</li>
                      <li>已禁用的卡密可以重新启用</li>
                      <li>已使用的卡密无法进行任何操作</li>
                      <li>用户使用卡密后，积分会立即到账</li>
                    </ul>
                  </div>
                </div>
              ),
            },
            {
              key: 'gift',
              label: <span><GiftOutlined /> 活动赠送</span>,
              children: (
                <div className="space-y-4">
                  {/* 活动赠送说明 */}
                  <div className="bg-gradient-to-r from-orange-50 to-yellow-50 border border-orange-200 rounded-lg p-4">
                    <div className="flex items-center gap-2 text-orange-600 font-medium mb-2">
                      <GiftOutlined className="text-xl" />
                      <span>活动积分赠送</span>
                    </div>
                    <div className="text-gray-600 text-sm">
                      当有活动时，可以给指定用户群体批量赠送积分，支持站内通知和邮件通知。
                    </div>
                  </div>

                  {/* 快捷选择 */}
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="text-gray-600 font-medium mb-3">快捷选择活动类型：</div>
                    <Space wrap size="middle">
                      <Button onClick={() => { giftForm.setFieldsValue({ points: 10, remark: '每日福利', target_type: 'all' }) }}>每日福利 +10</Button>
                      <Button onClick={() => { giftForm.setFieldsValue({ points: 50, remark: '周末福利', target_type: 'all' }) }}>周末福利 +50</Button>
                      <Button onClick={() => { giftForm.setFieldsValue({ points: 100, remark: '节日活动赠送', target_type: 'all' }) }}>节日活动 +100</Button>
                      <Button onClick={() => { giftForm.setFieldsValue({ points: 200, remark: '周年庆活动赠送', target_type: 'all' }) }}>周年庆 +200</Button>
                      <Button onClick={() => { giftForm.setFieldsValue({ points: 50, remark: '会员专属福利', target_type: 'member' }) }}>会员专属 +50</Button>
                      <Button onClick={() => { giftForm.setFieldsValue({ points: 30, remark: '新用户福利', target_type: 'non_member' }) }}>新用户福利 +30</Button>
                    </Space>
                  </div>

                  {/* 赠送表单 */}
                  <div className="bg-white border border-gray-100 rounded-xl p-5 shadow-sm">
                    <Form form={giftForm} layout="vertical" name="pointsGiftForm" initialValues={{ target_type: 'all', send_notification: true }}>
                      {/* 基本信息 */}
                      <div className="text-gray-700 font-medium mb-3 flex items-center gap-2">
                        <GiftOutlined className="text-orange-500" /> 赠送信息
                      </div>
                      <Row gutter={16}>
                        <Col span={8}>
                          <Form.Item name="points" label="赠送积分" rules={[{ required: true, message: '请输入赠送积分' }]}>
                            <InputNumber min={1} max={10000} className="w-full h-10 rounded-lg" placeholder="输入积分数量" />
                          </Form.Item>
                        </Col>
                        <Col span={16}>
                          <Form.Item name="remark" label="赠送原因" rules={[{ required: true, message: '请输入赠送原因' }]}>
                            <Input className="h-10 rounded-lg" placeholder="如：元旦活动赠送、会员专属福利等" />
                          </Form.Item>
                        </Col>
                      </Row>

                      <Divider className="my-4" />

                      {/* 目标用户 */}
                      <div className="text-gray-700 font-medium mb-3 flex items-center gap-2">
                        <TeamOutlined className="text-blue-500" /> 目标用户
                      </div>
                      <Row gutter={16}>
                        <Col span={8}>
                          <Form.Item name="target_type" label="赠送对象">
                            <Select className="h-10 rounded-lg">
                              <Select.Option value="all"><UserOutlined className="mr-1" />所有用户</Select.Option>
                              <Select.Option value="member"><CrownOutlined className="mr-1 text-yellow-500" />会员用户</Select.Option>
                              <Select.Option value="non_member"><UserOutlined className="mr-1 text-gray-400" />非会员用户</Select.Option>
                            </Select>
                          </Form.Item>
                        </Col>
                        <Form.Item noStyle shouldUpdate={(prev, cur) => prev.target_type !== cur.target_type}>
                          {({ getFieldValue }) => getFieldValue('target_type') === 'member' && (
                            <Col span={8}>
                              <Form.Item name="member_level" label="会员等级">
                                <Select allowClear placeholder="全部会员" className="h-10 rounded-lg">
                                  <Select.Option value={1}>月卡会员</Select.Option>
                                  <Select.Option value={2}>年卡会员</Select.Option>
                                </Select>
                              </Form.Item>
                            </Col>
                          )}
                        </Form.Item>
                      </Row>

                      <Divider className="my-4" />

                      {/* 通知设置 */}
                      <div className="text-gray-700 font-medium mb-3 flex items-center gap-2">
                        <BellOutlined className="text-green-500" /> 通知设置
                      </div>
                      <Row gutter={16}>
                        <Col span={8}>
                          <Form.Item name="send_notification" label="站内通知" valuePropName="checked">
                            <Switch checkedChildren="发送" unCheckedChildren="不发送" />
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item name="send_email" label="邮件通知" valuePropName="checked" initialValue={false}>
                            <Switch checkedChildren="发送" unCheckedChildren="不发送" />
                          </Form.Item>
                        </Col>
                      </Row>

                      {/* 站内通知内容 */}
                      <Form.Item noStyle shouldUpdate={(prev, cur) => prev.send_notification !== cur.send_notification}>
                        {({ getFieldValue }) => getFieldValue('send_notification') && (
                          <div className="bg-blue-50/50 rounded-xl p-4 mb-4 border border-blue-100">
                            <div className="text-blue-600 text-sm font-medium mb-3 flex items-center gap-1">
                              <BellOutlined /> 站内通知内容
                            </div>
                            <Row gutter={16}>
                              <Col span={12}>
                                <Form.Item name="notification_title" label="通知标题" extra="留空使用默认标题" className="mb-2">
                                  <Input className="h-10 rounded-lg" placeholder="默认：🎁 积分到账通知" />
                                </Form.Item>
                              </Col>
                              <Col span={12}>
                                <Form.Item name="notification_body" label="通知内容" extra="留空使用默认内容" className="mb-0">
                                  <Input.TextArea rows={2} className="rounded-lg" placeholder="默认：恭喜您获得 X 积分！" />
                                </Form.Item>
                              </Col>
                            </Row>
                          </div>
                        )}
                      </Form.Item>

                      {/* 邮件通知内容 */}
                      <Form.Item noStyle shouldUpdate={(prev, cur) => prev.send_email !== cur.send_email}>
                        {({ getFieldValue }) => getFieldValue('send_email') && (
                          <div className="bg-green-50/50 rounded-xl p-4 mb-4 border border-green-100">
                            <div className="text-green-600 text-sm font-medium mb-3 flex items-center gap-1">
                              <MailOutlined /> 邮件通知内容
                            </div>
                            <Row gutter={16}>
                              <Col span={12}>
                                <Form.Item name="email_title" label="邮件标题" extra="留空使用默认标题" className="mb-2">
                                  <Input className="h-10 rounded-lg" placeholder="默认：积分赠送通知" />
                                </Form.Item>
                              </Col>
                              <Col span={12}>
                                <Form.Item name="email_body" label="邮件内容" extra="可使用 {points} 表示积分数量" className="mb-0">
                                  <Input.TextArea rows={2} className="rounded-lg" placeholder="默认：恭喜您获得 {points} 积分！" />
                                </Form.Item>
                              </Col>
                            </Row>
                          </div>
                        )}
                      </Form.Item>

                      <Form.Item className="mb-0 mt-5">
                        <Button type="primary" size="large" icon={<GiftOutlined />} onClick={handleGiftPoints} loading={giftPointsMutation.isPending} className="h-11 rounded-xl px-8 font-medium shadow-lg hover:shadow-xl transition-shadow">
                          确认赠送
                        </Button>
                      </Form.Item>
                    </Form>
                  </div>

                  {/* 赠送规则说明 */}
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="font-medium text-blue-700 mb-2">📋 赠送规则说明：</div>
                    <ul className="list-disc list-inside text-gray-600 text-sm space-y-1">
                      <li>仅对状态正常的用户赠送积分（禁用、待审核用户不会收到）</li>
                      <li>可选择赠送给所有用户、仅会员用户或仅非会员用户</li>
                      <li>积分将立即到账，用户可在积分中心查看</li>
                      <li>赠送记录将显示在用户的积分明细中，来源显示为"活动奖励"</li>
                      <li>站内通知会即时推送，用户登录后可在通知中心查看</li>
                      <li>邮件通知为异步发送，可能有几分钟延迟</li>
                    </ul>
                  </div>
                </div>
              ),
            },
            {
              key: 'auto-gift',
              label: <span><ClockCircleOutlined /> 自动赠送</span>,
              children: (
                <div className="space-y-4">
                  {/* 说明 */}
                  <div className="bg-gradient-to-r from-purple-50 to-blue-50 border border-purple-200 rounded-lg p-4">
                    <div className="flex items-center gap-2 text-purple-600 font-medium mb-2">
                      <ClockCircleOutlined className="text-xl" />
                      <span>自动赠送规则</span>
                    </div>
                    <div className="text-gray-600 text-sm">
                      设置定时自动赠送积分规则，支持每日、每周、每月定时执行，自动给指定用户群体赠送积分。
                    </div>
                  </div>

                  {/* 操作栏 */}
                  <div className="flex justify-between items-center">
                    <span className="text-gray-500 text-sm">共 {giftRules?.length || 0} 条规则</span>
                    <Button type="primary" icon={<PlusOutlined />} onClick={handleAddGiftRule}>添加规则</Button>
                  </div>

                  {/* 规则列表 */}
                  <Table
                    columns={[
                      { title: '规则名称', dataIndex: 'name', key: 'name', render: (name: string) => <span className="font-medium">{name}</span> },
                      { title: '类型', dataIndex: 'rule_type', key: 'rule_type', width: 80, render: (t: number) => <Tag color="blue">{getRuleTypeName(t)}</Tag> },
                      { title: '积分', dataIndex: 'points', key: 'points', width: 80, render: (p: number) => <span className="text-orange-500 font-bold">{p}</span> },
                      { title: '目标用户', dataIndex: 'target_type', key: 'target_type', width: 100, render: (t: string) => getTargetTypeName(t) },
                      { title: '执行时间', key: 'execute_time', width: 140, render: (_: unknown, r: PointsGiftRule) => {
                        if (r.rule_type === 1) return r.execute_time
                        if (r.rule_type === 2) return `周${r.execute_day} ${r.execute_time}`
                        if (r.rule_type === 3) return `${r.execute_day}号 ${r.execute_time}`
                        if (r.rule_type === 4) return `${r.execute_month}月${r.execute_day}日 ${r.execute_time}`
                        return '-'
                      }},
                      { title: '下次执行', dataIndex: 'next_execute_at', key: 'next_execute_at', width: 150, render: (t: string) => t ? dayjs(t).format('MM-DD HH:mm') : '-' },
                      { title: '上次执行', dataIndex: 'last_execute_at', key: 'last_execute_at', width: 150, render: (t: string) => t ? dayjs(t).format('MM-DD HH:mm') : '从未执行' },
                      { title: '状态', dataIndex: 'enabled', key: 'enabled', width: 80, render: (enabled: boolean, record: PointsGiftRule) => (
                        <Switch checked={enabled} onChange={() => toggleGiftRuleMutation.mutate(record.id)} size="small" />
                      )},
                      { title: '操作', key: 'action', width: 180, render: (_: unknown, record: PointsGiftRule) => (
                        <Space size="small">
                          <Popconfirm title="确定立即执行？" onConfirm={() => executeGiftRuleMutation.mutate(record.id)}>
                            <Button type="link" size="small" className="text-green-600">执行</Button>
                          </Popconfirm>
                          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditGiftRule(record)} />
                          <Popconfirm title="确定删除？" onConfirm={() => deleteGiftRuleMutation.mutate(record.id)}>
                            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
                          </Popconfirm>
                        </Space>
                      )},
                    ]}
                    dataSource={giftRules || []}
                    rowKey="id"
                    loading={giftRulesLoading}
                    pagination={false}
                    size="small"
                  />

                  {/* 执行日志 */}
                  <div className="mt-6">
                    <div className="text-gray-700 font-medium mb-3">执行日志</div>
                    <Table
                      columns={[
                        { title: '规则名称', dataIndex: 'rule_name', key: 'rule_name' },
                        { title: '积分', dataIndex: 'points', key: 'points', width: 80, render: (p: number) => <span className="text-orange-500">{p}</span> },
                        { title: '目标用户', dataIndex: 'total_users', key: 'total_users', width: 80 },
                        { title: '成功', dataIndex: 'success_count', key: 'success_count', width: 80, render: (c: number) => <span className="text-green-600">{c}</span> },
                        { title: '失败', dataIndex: 'failed_count', key: 'failed_count', width: 80, render: (c: number) => c > 0 ? <span className="text-red-500">{c}</span> : c },
                        { title: '执行时间', dataIndex: 'execute_at', key: 'execute_at', render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm:ss') },
                      ]}
                      dataSource={giftLogsData?.list || []}
                      rowKey="id"
                      size="small"
                      pagination={{ current: giftLogPage, pageSize: 10, total: giftLogsData?.total || 0, onChange: setGiftLogPage, size: 'small' }}
                    />
                  </div>

                  {/* 说明 */}
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-gray-600">
                    <div className="font-medium text-blue-700 mb-1">📋 自动赠送说明：</div>
                    <ul className="list-disc list-inside space-y-0.5">
                      <li>每日赠送：每天在指定时间执行</li>
                      <li>每周赠送：每周指定日期的指定时间执行</li>
                      <li>每月赠送：每月指定日期的指定时间执行</li>
                      <li>每年赠送：每年指定月日的指定时间执行（如周年庆）</li>
                      <li>可随时手动执行规则，不影响定时执行</li>
                      <li>禁用规则后将不再自动执行</li>
                    </ul>
                  </div>
                </div>
              ),
            },
          ]}
        />
      </Card>

      {/* 兑换规则弹窗 */}
      <Modal
        title={editingRule ? '编辑规则' : '添加规则'}
        open={ruleModalOpen}
        onCancel={() => { setRuleModalOpen(false); setEditingRule(null); ruleForm.resetFields() }}
        onOk={handleRuleSubmit}
        confirmLoading={createRuleMutation.isPending || updateRuleMutation.isPending}
      >
        <Form form={ruleForm} layout="vertical" name="pointsRuleForm" className="mt-4">
          <Form.Item name="name" label="规则名称" rules={[{ required: true, message: '请输入' }]}>
            <Input placeholder="如：天卡" className="h-10 rounded-lg" />
          </Form.Item>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="points" label="所需积分" rules={[{ required: true, message: '请输入' }]}>
              <InputNumber min={1} className="w-full h-10 rounded-lg" />
            </Form.Item>
            <Form.Item name="member_days" label="兑换天数" rules={[{ required: true, message: '请输入' }]}>
              <InputNumber min={1} className="w-full h-10 rounded-lg" suffix="天" />
            </Form.Item>
          </div>
          <Form.Item name="description" label="描述"><Input.TextArea rows={2} className="rounded-lg" placeholder="规则描述（可选）" /></Form.Item>
          <div className="grid grid-cols-2 gap-4">
            <Form.Item name="sort_order" label="排序" initialValue={0}><InputNumber min={0} className="w-full h-10 rounded-lg" /></Form.Item>
            <Form.Item name="enabled" label="状态" valuePropName="checked" initialValue={true}><Switch checkedChildren="启用" unCheckedChildren="禁用" /></Form.Item>
          </div>
        </Form>
      </Modal>

      {/* 生成积分卡弹窗 */}
      <Modal
        title="生成积分卡"
        open={cardModalOpen}
        onCancel={() => { setCardModalOpen(false); cardForm.resetFields() }}
        onOk={handleCreateCard}
        confirmLoading={createCardMutation.isPending}
        okText="生成"
      >
        <Form form={cardForm} layout="vertical" name="pointsCardForm" className="mt-4">
          <Form.Item name="points" label="积分面值" rules={[{ required: true, message: '请输入' }]}>
            <InputNumber min={1} max={100000} className="w-full h-10 rounded-lg" suffix="积分" placeholder="输入积分数量" />
          </Form.Item>
          <Form.Item name="quantity" label="生成数量" rules={[{ required: true, message: '请输入' }]}>
            <InputNumber min={1} max={1000} className="w-full h-10 rounded-lg" suffix="张" placeholder="输入生成数量" />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} className="rounded-lg" placeholder="备注信息（可选）" /></Form.Item>
        </Form>
      </Modal>

      {/* 导出卡密弹窗 */}
      <Modal
        title={`导出卡密 - ${exportBatchNo}`}
        open={exportModalOpen}
        onCancel={() => setExportModalOpen(false)}
        footer={[
          <Button key="copy" icon={<CopyOutlined />} onClick={copyAllCodes} className="rounded-lg">复制全部</Button>,
          <Button key="download" type="primary" icon={<ExportOutlined />} onClick={downloadCodes} className="rounded-lg">下载TXT</Button>,
        ]}
        width={520}
      >
        <div className="mt-4">
          <div className="text-gray-500 mb-3 flex items-center justify-between">
            <span>共 <span className="text-blue-600 font-medium">{exportedCodes.length}</span> 张卡密</span>
            <span className="text-xs text-gray-400">点击卡密可复制</span>
          </div>
          <div className="max-h-80 overflow-auto bg-gradient-to-b from-gray-50 to-white rounded-xl p-3 font-mono text-sm border border-gray-100">
            {exportedCodes.map((code, index) => (
              <div key={index} className="flex justify-between items-center py-2 hover:bg-blue-50 px-3 rounded-lg transition-colors cursor-pointer group" onClick={() => copyToClipboard(code)}>
                <span className="text-gray-700">{code}</span>
                <CopyOutlined className="text-gray-300 group-hover:text-blue-500 transition-colors" />
              </div>
            ))}
          </div>
        </div>
      </Modal>

      {/* 自动赠送规则弹窗 */}
      <Modal
        title={editingGiftRule ? '编辑自动赠送规则' : '添加自动赠送规则'}
        open={giftRuleModalOpen}
        onCancel={() => { setGiftRuleModalOpen(false); setEditingGiftRule(null); giftRuleForm.resetFields() }}
        onOk={handleGiftRuleSubmit}
        confirmLoading={createGiftRuleMutation.isPending || updateGiftRuleMutation.isPending}
        width={600}
      >
        <Form form={giftRuleForm} layout="vertical" name="pointsGiftRuleForm" className="mt-4">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="name" label="规则名称" rules={[{ required: true, message: '请输入规则名称' }]}>
                <Input placeholder="如：每日福利" className="h-10 rounded-lg" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="points" label="赠送积分" rules={[{ required: true, message: '请输入赠送积分' }]}>
                <InputNumber min={1} max={10000} className="w-full h-10 rounded-lg" placeholder="积分数量" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="rule_type" label="执行周期" rules={[{ required: true }]}>
                <Select className="h-10 rounded-lg">
                  <Select.Option value={1}>每日</Select.Option>
                  <Select.Option value={2}>每周</Select.Option>
                  <Select.Option value={3}>每月</Select.Option>
                  <Select.Option value={4}>每年</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="execute_time" label="执行时间" rules={[{ required: true }]}>
                <Input placeholder="08:00" className="h-10 rounded-lg" />
              </Form.Item>
            </Col>
            <Form.Item noStyle shouldUpdate={(prev, cur) => prev.rule_type !== cur.rule_type}>
              {({ getFieldValue }) => {
                const ruleType = getFieldValue('rule_type')
                if (ruleType === 2) {
                  return (
                    <Col span={8}>
                      <Form.Item name="execute_day" label="周几" rules={[{ required: true }]}>
                        <Select className="h-10 rounded-lg">
                          <Select.Option value={1}>周一</Select.Option>
                          <Select.Option value={2}>周二</Select.Option>
                          <Select.Option value={3}>周三</Select.Option>
                          <Select.Option value={4}>周四</Select.Option>
                          <Select.Option value={5}>周五</Select.Option>
                          <Select.Option value={6}>周六</Select.Option>
                          <Select.Option value={7}>周日</Select.Option>
                        </Select>
                      </Form.Item>
                    </Col>
                  )
                }
                if (ruleType === 3) {
                  return (
                    <Col span={8}>
                      <Form.Item name="execute_day" label="每月几号" rules={[{ required: true }]}>
                        <InputNumber min={1} max={28} className="w-full h-10 rounded-lg" placeholder="1-28" />
                      </Form.Item>
                    </Col>
                  )
                }
                if (ruleType === 4) {
                  return (
                    <>
                      <Col span={4}>
                        <Form.Item name="execute_month" label="月份" rules={[{ required: true }]}>
                          <Select className="h-10 rounded-lg">
                            {[1,2,3,4,5,6,7,8,9,10,11,12].map(m => (
                              <Select.Option key={m} value={m}>{m}月</Select.Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={4}>
                        <Form.Item name="execute_day" label="日期" rules={[{ required: true }]}>
                          <InputNumber min={1} max={28} className="w-full h-10 rounded-lg" />
                        </Form.Item>
                      </Col>
                    </>
                  )
                }
                return null
              }}
            </Form.Item>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="target_type" label="目标用户">
                <Select className="h-10 rounded-lg">
                  <Select.Option value="all">所有用户</Select.Option>
                  <Select.Option value="member">会员用户</Select.Option>
                  <Select.Option value="non_member">非会员用户</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Form.Item noStyle shouldUpdate={(prev, cur) => prev.target_type !== cur.target_type}>
              {({ getFieldValue }) => getFieldValue('target_type') === 'member' && (
                <Col span={12}>
                  <Form.Item name="member_level" label="会员等级">
                    <Select allowClear placeholder="全部会员" className="h-10 rounded-lg">
                      <Select.Option value={1}>月卡会员</Select.Option>
                      <Select.Option value={2}>年卡会员</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
              )}
            </Form.Item>
          </Row>

          <Divider className="my-3" />

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="send_notification" label="发送站内通知" valuePropName="checked">
                <Switch checkedChildren="是" unCheckedChildren="否" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="enabled" label="启用规则" valuePropName="checked">
                <Switch checkedChildren="启用" unCheckedChildren="禁用" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.send_notification !== cur.send_notification}>
            {({ getFieldValue }) => getFieldValue('send_notification') && (
              <>
                <Form.Item name="notification_title" label="通知标题" extra="留空使用默认标题">
                  <Input className="h-10 rounded-lg" placeholder="默认：🎁 积分到账通知" />
                </Form.Item>
                <Form.Item name="notification_body" label="通知内容" extra="留空使用默认内容">
                  <Input.TextArea rows={2} className="rounded-lg" placeholder="默认：恭喜您获得 X 积分！" />
                </Form.Item>
              </>
            )}
          </Form.Item>
        </Form>
      </Modal>

    </div>
  )
}

export default PointsManagement

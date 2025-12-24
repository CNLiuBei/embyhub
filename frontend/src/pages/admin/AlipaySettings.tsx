import { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, Form, Input, Switch, Button, App, Space, Table, Tag, Tooltip, Typography, Modal, InputNumber, Popconfirm, Tabs, Alert, Descriptions, Badge } from 'antd'
import { AlipayCircleOutlined, SaveOutlined, ApiOutlined, ReloadOutlined, EyeOutlined, EyeInvisibleOutlined, PlusOutlined, EditOutlined, DeleteOutlined, CrownOutlined, CloudOutlined, PlayCircleOutlined, PauseCircleOutlined, CopyOutlined, DownloadOutlined } from '@ant-design/icons'
import { alipayAdminApi, AlipayConfig, AlipayLog, VipPlan, tunnelApi, TunnelStatus } from '../../services/paymentApi'
import dayjs from 'dayjs'

const { TextArea } = Input
const { Text } = Typography

const AlipaySettings = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm()
  const [planForm] = Form.useForm()
  const [tunnelForm] = Form.useForm()
  const [showPrivateKey, setShowPrivateKey] = useState(false)
  const [logsPage, setLogsPage] = useState(1)
  const [planModalOpen, setPlanModalOpen] = useState(false)
  const [editingPlan, setEditingPlan] = useState<VipPlan | null>(null)
  const [tunnelModalOpen, setTunnelModalOpen] = useState(false)

  // 获取配置
  const { data: configData, isLoading: configLoading } = useQuery({
    queryKey: ['alipayConfig'],
    queryFn: async () => {
      const response = await alipayAdminApi.getConfig()
      return response.data.data as AlipayConfig
    },
  })

  // 获取日志
  const { data: logsData, isLoading: logsLoading } = useQuery({
    queryKey: ['alipayLogs', logsPage],
    queryFn: async () => {
      const response = await alipayAdminApi.getLogs({ page: logsPage, page_size: 10 })
      return response.data.data as { list: AlipayLog[]; total: number }
    },
  })

  // 获取VIP套餐
  const { data: plansData, isLoading: plansLoading } = useQuery({
    queryKey: ['vipPlans'],
    queryFn: async () => {
      const response = await alipayAdminApi.getVipPlans()
      return response.data.data as VipPlan[]
    },
  })

  // 获取隧道状态
  const { data: tunnelStatus, isLoading: tunnelLoading } = useQuery({
    queryKey: ['tunnelStatus'],
    queryFn: async () => {
      const response = await tunnelApi.getStatus()
      return response.data.data as TunnelStatus
    },
    refetchInterval: 3000, // 每3秒刷新一次状态
  })

  // 记录之前的配置状态
  const prevConfiguredRef = useRef<boolean | undefined>(undefined)
  
  // 监听隧道状态变化，如果配置成功则关闭弹窗
  useEffect(() => {
    if (tunnelModalOpen && tunnelStatus?.configured && prevConfiguredRef.current === false) {
      setTunnelModalOpen(false)
      tunnelForm.resetFields()
      message.success('隧道创建成功')
    }
    prevConfiguredRef.current = tunnelStatus?.configured
  }, [tunnelStatus?.configured, tunnelModalOpen, tunnelForm, message])

  // 保存配置
  const saveMutation = useMutation({
    mutationFn: (values: Parameters<typeof alipayAdminApi.saveConfig>[0]) =>
      alipayAdminApi.saveConfig(values),
    onSuccess: () => {
      message.success('保存成功')
      queryClient.invalidateQueries({ queryKey: ['alipayConfig'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '保存失败')
    },
  })

  // 测试连接
  const testMutation = useMutation({
    mutationFn: () => alipayAdminApi.testConnection(),
    onSuccess: () => {
      message.success('连接测试成功')
    },
    onError: (err: Error) => {
      message.error(err.message || '连接测试失败')
    },
  })

  // 创建套餐
  const createPlanMutation = useMutation({
    mutationFn: (data: { name: string; description: string; price: number; duration_days: number }) =>
      alipayAdminApi.createVipPlan(data),
    onSuccess: () => {
      message.success('创建成功')
      queryClient.invalidateQueries({ queryKey: ['vipPlans'] })
      setPlanModalOpen(false)
      planForm.resetFields()
    },
    onError: (err: Error) => {
      message.error(err.message || '创建失败')
    },
  })

  // 更新套餐
  const updatePlanMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { name: string; description: string; price: number; duration_days: number; is_active: boolean } }) =>
      alipayAdminApi.updateVipPlan(id, data),
    onSuccess: () => {
      message.success('更新成功')
      queryClient.invalidateQueries({ queryKey: ['vipPlans'] })
      setPlanModalOpen(false)
      setEditingPlan(null)
      planForm.resetFields()
    },
    onError: (err: Error) => {
      message.error(err.message || '更新失败')
    },
  })

  // 删除套餐
  const deletePlanMutation = useMutation({
    mutationFn: (id: number) => alipayAdminApi.deleteVipPlan(id),
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['vipPlans'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '删除失败')
    },
  })

  // 切换套餐状态
  const togglePlanMutation = useMutation({
    mutationFn: (id: number) => alipayAdminApi.toggleVipPlanStatus(id),
    onSuccess: () => {
      message.success('状态已更新')
      queryClient.invalidateQueries({ queryKey: ['vipPlans'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '操作失败')
    },
  })

  // 创建隧道（自动处理授权流程）
  const [pendingTunnelData, setPendingTunnelData] = useState<{ tunnel_name: string; domain: string; subdomain: string; local_host?: string; local_port?: number } | null>(null)
  const [isWaitingAuth, setIsWaitingAuth] = useState(false)

  const createTunnelMutation = useMutation({
    mutationFn: (data: { tunnel_name: string; domain: string; subdomain: string; local_host?: string; local_port?: number }) =>
      tunnelApi.createTunnel(data),
    onSuccess: (response) => {
      const data = response.data.data
      if (data.need_auth && data.auth_url) {
        // 需要授权，保存当前表单数据，打开授权页面
        setPendingTunnelData(tunnelForm.getFieldsValue())
        setIsWaitingAuth(true)
        window.open(data.auth_url, '_blank')
        message.info('请在新窗口中完成 Cloudflare 授权，授权完成后会自动创建隧道')
      } else {
        // 创建成功
        message.success('隧道创建成功')
        queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
        setTunnelModalOpen(false)
        tunnelForm.resetFields()
        setPendingTunnelData(null)
        setIsWaitingAuth(false)
      }
    },
    onError: (err: Error) => {
      // 即使报错也刷新状态，因为可能是超时但实际已成功
      queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
      message.error(err.message || '创建隧道失败')
      setIsWaitingAuth(false)
    },
    onSettled: () => {
      // 无论成功失败都刷新状态
      queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
    },
  })

  // 监听授权状态变化，授权完成后自动重试创建隧道
  useEffect(() => {
    if (isWaitingAuth && tunnelStatus?.logged_in && pendingTunnelData) {
      // 授权完成，自动重试创建隧道
      setIsWaitingAuth(false)
      message.info('授权成功，正在创建隧道...')
      createTunnelMutation.mutate(pendingTunnelData)
    }
  }, [tunnelStatus?.logged_in, isWaitingAuth, pendingTunnelData])

  // 启动隧道
  const startTunnelMutation = useMutation({
    mutationFn: () => tunnelApi.startTunnel(),
    onSuccess: () => {
      message.success('隧道已启动')
      queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '启动失败')
    },
  })

  // 停止隧道
  const stopTunnelMutation = useMutation({
    mutationFn: () => tunnelApi.stopTunnel(),
    onSuccess: () => {
      message.success('隧道已停止')
      queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '停止失败')
    },
  })

  // 删除隧道
  const deleteTunnelMutation = useMutation({
    mutationFn: () => tunnelApi.deleteTunnel(),
    onSuccess: () => {
      message.success('隧道已删除')
      queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '删除失败')
    },
  })

  // 下载 cloudflared
  const downloadCloudflaredMutation = useMutation({
    mutationFn: () => tunnelApi.downloadCloudflared(),
    onSuccess: () => {
      message.success('cloudflared 下载成功')
      queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
    },
    onError: (err: Error) => {
      message.error(err.message || '下载失败')
    },
  })

  // Cloudflare 授权
  const loginMutation = useMutation({
    mutationFn: () => tunnelApi.login(),
    onSuccess: (response) => {
      const data = response.data.data
      if (data.logged_in) {
        message.success('已完成 Cloudflare 授权')
        queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })
      } else if (data.auth_url) {
        // 在新窗口打开授权URL
        window.open(data.auth_url, '_blank')
        message.info('请在新窗口中完成 Cloudflare 授权，授权完成后点击刷新状态')
      }
    },
    onError: (err: Error) => {
      message.error(err.message || '获取授权链接失败')
    },
  })

  // 表单提交
  const handleSubmit = (values: AlipayConfig) => {
    saveMutation.mutate({
      app_id: values.app_id,
      app_public_key: values.app_public_key,
      app_private_key: values.app_private_key,
      alipay_public_key: values.alipay_public_key,
      notify_url: values.notify_url,
      enabled: values.enabled,
      is_production: values.is_production,
    })
  }

  // 套餐表单提交
  const handlePlanSubmit = (values: { name: string; description: string; price: number; duration_days: number; is_active?: boolean }) => {
    // 价格转换为分
    const priceInCents = Math.round(values.price * 100)
    
    if (editingPlan) {
      updatePlanMutation.mutate({
        id: editingPlan.id,
        data: {
          name: values.name,
          description: values.description || '',
          price: priceInCents,
          duration_days: values.duration_days,
          is_active: values.is_active ?? true,
        },
      })
    } else {
      createPlanMutation.mutate({
        name: values.name,
        description: values.description || '',
        price: priceInCents,
        duration_days: values.duration_days,
      })
    }
  }

  // 打开编辑弹窗
  const openEditModal = (plan: VipPlan) => {
    setEditingPlan(plan)
    planForm.setFieldsValue({
      name: plan.name,
      description: plan.description,
      price: plan.price / 100, // 分转元
      duration_days: plan.duration_days,
      is_active: plan.is_active,
    })
    setPlanModalOpen(true)
  }

  // 打开新建弹窗
  const openCreateModal = () => {
    setEditingPlan(null)
    planForm.resetFields()
    setPlanModalOpen(true)
  }

  // 日志表格列
  const logColumns = [
    {
      title: '订单号',
      dataIndex: 'order_no',
      key: 'order_no',
      width: 200,
      render: (text: string) => text ? <Text copyable={{ text }}>{text}</Text> : '-',
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: string) => {
        const actionMap: Record<string, { color: string; text: string }> = {
          create: { color: 'blue', text: '创建订单' },
          notify: { color: 'green', text: '异步通知' },
          query: { color: 'orange', text: '查询订单' },
        }
        const config = actionMap[action] || { color: 'default', text: action }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'success' ? 'green' : 'red'}>
          {status === 'success' ? '成功' : '失败'}
        </Tag>
      ),
    },
    {
      title: '错误信息',
      dataIndex: 'error_msg',
      key: 'error_msg',
      width: 200,
      ellipsis: true,
      render: (text: string) => text ? (
        <Tooltip title={text}>
          <Text type="danger">{text}</Text>
        </Tooltip>
      ) : '-',
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      width: 80,
      render: (ms: number) => `${ms}ms`,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  // 套餐表格列
  const planColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '套餐名称',
      dataIndex: 'name',
      key: 'name',
      width: 120,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      width: 200,
      ellipsis: true,
    },
    {
      title: '价格',
      dataIndex: 'price',
      key: 'price',
      width: 100,
      render: (price: number) => (
        <span className="text-red-500 font-bold">¥{(price / 100).toFixed(2)}</span>
      ),
    },
    {
      title: '会员天数',
      dataIndex: 'duration_days',
      key: 'duration_days',
      width: 100,
      render: (days: number) => <Tag color="blue">{days}天</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 80,
      render: (isActive: boolean, record: VipPlan) => (
        <Switch
          checked={isActive}
          onChange={() => togglePlanMutation.mutate(record.id)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
          size="small"
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: unknown, record: VipPlan) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此套餐？"
            onConfirm={() => deletePlanMutation.mutate(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const tabItems = [
    {
      key: 'config',
      label: (
        <Space>
          <AlipayCircleOutlined />
          支付配置
        </Space>
      ),
      children: (
        <Form
          form={form}
          layout="vertical"
          initialValues={configData}
          onFinish={handleSubmit}
          key={configData?.id}
        >
          <div className="grid grid-cols-2 gap-4">
            <Form.Item
              name="app_id"
              label="应用ID (AppID)"
              rules={[{ required: true, message: '请输入应用ID' }]}
            >
              <Input placeholder="请输入支付宝应用ID" />
            </Form.Item>

            <Form.Item
              name="notify_url"
              label="异步通知地址"
              rules={[{ required: true, message: '请输入异步通知地址' }]}
              extra="支付成功后支付宝会回调此地址"
            >
              <Input placeholder="https://your-domain.com/api/v1/payment/alipay/notify" />
            </Form.Item>
          </div>

          <Form.Item
            name="app_public_key"
            label="应用公钥"
          >
            <TextArea
              rows={4}
              placeholder="请输入应用公钥（可选）"
            />
          </Form.Item>

          <Form.Item
            name="app_private_key"
            label={
              <Space>
                <span>应用私钥</span>
                <Button
                  type="link"
                  size="small"
                  icon={showPrivateKey ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                  onClick={() => setShowPrivateKey(!showPrivateKey)}
                >
                  {showPrivateKey ? '隐藏' : '显示'}
                </Button>
              </Space>
            }
            extra="私钥将加密存储，更新时留空表示不修改"
          >
            <TextArea
              rows={4}
              placeholder="请输入应用私钥"
              style={{ fontFamily: showPrivateKey ? 'inherit' : 'monospace' }}
            />
          </Form.Item>

          <Form.Item
            name="alipay_public_key"
            label="支付宝公钥"
            rules={[{ required: true, message: '请输入支付宝公钥' }]}
            extra="用于验证支付宝回调签名"
          >
            <TextArea
              rows={4}
              placeholder="请输入支付宝公钥"
            />
          </Form.Item>

          <div className="grid grid-cols-2 gap-4">
            <Form.Item
              name="enabled"
              label="启用支付"
              valuePropName="checked"
            >
              <Switch checkedChildren="启用" unCheckedChildren="禁用" />
            </Form.Item>

            <Form.Item
              name="is_production"
              label="生产环境"
              valuePropName="checked"
              extra="关闭则使用沙箱环境"
            >
              <Switch checkedChildren="生产" unCheckedChildren="沙箱" />
            </Form.Item>
          </div>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SaveOutlined />}
                loading={saveMutation.isPending}
              >
                保存配置
              </Button>
              <Button
                icon={<ApiOutlined />}
                onClick={() => testMutation.mutate()}
                loading={testMutation.isPending}
              >
                测试连接
              </Button>
            </Space>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'plans',
      label: (
        <Space>
          <CrownOutlined />
          VIP套餐
        </Space>
      ),
      children: (
        <div>
          <div className="mb-4 flex justify-between items-center">
            <div className="text-gray-500">
              管理VIP会员套餐，用户可在会员中心选择套餐进行支付宝购买
            </div>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              添加套餐
            </Button>
          </div>
          <Table
            columns={planColumns}
            dataSource={plansData || []}
            rowKey="id"
            loading={plansLoading}
            pagination={false}
            size="small"
          />
        </div>
      ),
    },
    {
      key: 'logs',
      label: (
        <Space>
          <ApiOutlined />
          API日志
        </Space>
      ),
      children: (
        <div>
          <div className="mb-4 flex justify-end">
            <Button
              icon={<ReloadOutlined />}
              onClick={() => queryClient.invalidateQueries({ queryKey: ['alipayLogs'] })}
            >
              刷新
            </Button>
          </div>
          <Table
            columns={logColumns}
            dataSource={logsData?.list || []}
            rowKey="id"
            loading={logsLoading}
            pagination={{
              current: logsPage,
              pageSize: 10,
              total: logsData?.total || 0,
              onChange: setLogsPage,
              showSizeChanger: false,
              showTotal: (total) => `共 ${total} 条`,
            }}
            size="small"
            scroll={{ x: 900 }}
          />
        </div>
      ),
    },
    {
      key: 'tunnel',
      label: (
        <Space>
          <CloudOutlined />
          回调隧道
        </Space>
      ),
      children: (
        <div>
          {/* 隧道状态说明 */}
          <Alert
            message="Cloudflare 隧道说明"
            description="支付宝异步通知需要公网可访问的回调地址。通过 Cloudflare Tunnel 可以将本地服务暴露到公网，无需公网IP或端口映射。"
            type="info"
            showIcon
            className="mb-4"
          />

          {/* 环境检查 */}
          <Card title="环境状态" size="small" className="mb-4" loading={tunnelLoading}>
            <Descriptions column={2} size="small">
              <Descriptions.Item label="cloudflared">
                {tunnelStatus?.installed ? (
                  <Badge status="success" text={tunnelStatus.version || '已安装'} />
                ) : (
                  <Badge status="error" text="未安装" />
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Cloudflare 授权">
                {tunnelStatus?.logged_in ? (
                  <Badge status="success" text="已授权" />
                ) : (
                  <Space>
                    <Badge status="warning" text="未授权" />
                    <Button
                      type="link"
                      size="small"
                      onClick={() => loginMutation.mutate()}
                      loading={loginMutation.isPending}
                    >
                      点击授权
                    </Button>
                  </Space>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="隧道配置">
                {tunnelStatus?.configured ? (
                  <Badge status="success" text="已配置" />
                ) : (
                  <Badge status="default" text="未配置" />
                )}
              </Descriptions.Item>
              <Descriptions.Item label="隧道状态">
                {tunnelStatus?.running ? (
                  <Badge status="processing" text="运行中" />
                ) : (
                  <Badge status="default" text="已停止" />
                )}
              </Descriptions.Item>
              <Descriptions.Item label="连通性">
                {!tunnelStatus?.running ? (
                  <Badge status="default" text="未运行" />
                ) : tunnelStatus?.connected ? (
                  <Badge status="success" text="已连通" />
                ) : (
                  <Badge status="error" text="无法连通" />
                )}
              </Descriptions.Item>
            </Descriptions>

            {/* 操作按钮 */}
            <div className="mt-4 flex gap-2">
              {!tunnelStatus?.installed && (
                <Button
                  type="primary"
                  icon={<DownloadOutlined />}
                  onClick={() => downloadCloudflaredMutation.mutate()}
                  loading={downloadCloudflaredMutation.isPending}
                >
                  {downloadCloudflaredMutation.isPending ? '正在下载...' : '下载 cloudflared'}
                </Button>
              )}
              {tunnelStatus?.installed && !tunnelStatus?.logged_in && (
                <Button
                  type="primary"
                  onClick={() => loginMutation.mutate()}
                  loading={loginMutation.isPending}
                >
                  授权 Cloudflare
                </Button>
              )}
              <Button
                icon={<ReloadOutlined />}
                onClick={() => queryClient.invalidateQueries({ queryKey: ['tunnelStatus'] })}
              >
                刷新状态
              </Button>
            </div>
          </Card>

          {/* 隧道配置 */}
          {tunnelStatus?.configured && (
            <Card title="隧道信息" size="small" className="mb-4">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="隧道名称">{tunnelStatus.tunnel_name}</Descriptions.Item>
                <Descriptions.Item label="本地地址">{tunnelStatus.local_host || 'localhost'}:{tunnelStatus.local_port}</Descriptions.Item>
                <Descriptions.Item label="公网域名" span={2}>
                  <Space>
                    <a href={`https://${tunnelStatus.full_domain}`} target="_blank" rel="noopener noreferrer">
                      https://{tunnelStatus.full_domain}
                    </a>
                    <Button
                      type="link"
                      size="small"
                      icon={<CopyOutlined />}
                      onClick={() => {
                        navigator.clipboard.writeText(`https://${tunnelStatus.full_domain}`)
                        message.success('已复制')
                      }}
                    />
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="回调地址" span={2}>
                  <Space>
                    <Text copyable={{ text: tunnelStatus.notify_url }}>
                      {tunnelStatus.notify_url}
                    </Text>
                  </Space>
                </Descriptions.Item>
              </Descriptions>

              {tunnelStatus.error_msg && (
                <Alert message="错误信息" description={tunnelStatus.error_msg} type="error" className="mt-4" />
              )}

              <div className="mt-4 flex gap-2">
                {tunnelStatus.running ? (
                  <>
                    <Button
                      icon={<PauseCircleOutlined />}
                      onClick={() => stopTunnelMutation.mutate()}
                      loading={stopTunnelMutation.isPending}
                    >
                      停止隧道
                    </Button>
                    <Button
                      icon={<ReloadOutlined />}
                      onClick={() => startTunnelMutation.mutate()}
                      loading={startTunnelMutation.isPending}
                    >
                      重启隧道
                    </Button>
                  </>
                ) : (
                  <Button
                    type="primary"
                    icon={<PlayCircleOutlined />}
                    onClick={() => startTunnelMutation.mutate()}
                    loading={startTunnelMutation.isPending}
                  >
                    启动隧道
                  </Button>
                )}
                <Popconfirm
                  title="确定删除隧道？"
                  description="删除后需要重新配置"
                  onConfirm={() => deleteTunnelMutation.mutate()}
                >
                  <Button danger icon={<DeleteOutlined />} loading={deleteTunnelMutation.isPending}>
                    删除隧道
                  </Button>
                </Popconfirm>
              </div>
            </Card>
          )}

          {/* 创建隧道按钮 */}
          {tunnelStatus?.installed && !tunnelStatus?.configured && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setTunnelModalOpen(true)}>
              配置隧道
            </Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <Card
        title={
          <Space>
            <AlipayCircleOutlined style={{ color: '#1677ff', fontSize: 20 }} />
            <span>支付宝设置</span>
          </Space>
        }
        loading={configLoading}
      >
        <Tabs items={tabItems} />
      </Card>

      {/* 套餐编辑弹窗 */}
      <Modal
        title={editingPlan ? '编辑套餐' : '添加套餐'}
        open={planModalOpen}
        onCancel={() => {
          setPlanModalOpen(false)
          setEditingPlan(null)
          planForm.resetFields()
        }}
        footer={null}
        width={500}
      >
        <Form
          form={planForm}
          layout="vertical"
          onFinish={handlePlanSubmit}
        >
          <Form.Item
            name="name"
            label="套餐名称"
            rules={[{ required: true, message: '请输入套餐名称' }]}
          >
            <Input placeholder="如：月卡会员、年卡会员" />
          </Form.Item>

          <Form.Item
            name="description"
            label="套餐描述"
          >
            <Input placeholder="如：30天VIP会员" />
          </Form.Item>

          <div className="grid grid-cols-2 gap-4">
            <Form.Item
              name="price"
              label="价格（元）"
              rules={[{ required: true, message: '请输入价格' }]}
            >
              <InputNumber
                min={0.01}
                step={0.01}
                precision={2}
                placeholder="10.00"
                style={{ width: '100%' }}
                prefix="¥"
              />
            </Form.Item>

            <Form.Item
              name="duration_days"
              label="会员天数"
              rules={[{ required: true, message: '请输入会员天数' }]}
            >
              <InputNumber
                min={1}
                placeholder="30"
                style={{ width: '100%' }}
                suffix="天"
              />
            </Form.Item>
          </div>

          {editingPlan && (
            <Form.Item
              name="is_active"
              label="启用状态"
              valuePropName="checked"
            >
              <Switch checkedChildren="启用" unCheckedChildren="禁用" />
            </Form.Item>
          )}

          <Form.Item className="mb-0 mt-6">
            <Space className="w-full justify-end">
              <Button onClick={() => {
                setPlanModalOpen(false)
                setEditingPlan(null)
                planForm.resetFields()
              }}>
                取消
              </Button>
              <Button
                type="primary"
                htmlType="submit"
                loading={createPlanMutation.isPending || updatePlanMutation.isPending}
              >
                {editingPlan ? '保存' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 隧道配置弹窗 */}
      <Modal
        title="配置 Cloudflare 隧道"
        open={tunnelModalOpen}
        onCancel={() => {
          setTunnelModalOpen(false)
          tunnelForm.resetFields()
          setPendingTunnelData(null)
          setIsWaitingAuth(false)
        }}
        footer={null}
        width={500}
      >
        <Alert
          message="配置说明"
          description={
            <div>
              <p>1. 请确保您已在 Cloudflare 中添加了对应的域名</p>
              <p>2. 如果未授权 Cloudflare，创建隧道时会自动弹出授权页面</p>
              <p>3. 授权完成后会自动继续创建隧道</p>
            </div>
          }
          type="info"
          showIcon
          className="mb-4"
        />
        {isWaitingAuth && (
          <Alert
            message="等待授权"
            description="请在新窗口中完成 Cloudflare 授权，授权完成后会自动创建隧道"
            type="warning"
            showIcon
            className="mb-4"
          />
        )}
        <Form
          form={tunnelForm}
          layout="vertical"
          onFinish={(values) => createTunnelMutation.mutate(values)}
          initialValues={{ local_host: 'localhost', local_port: 54680 }}
        >
          <Form.Item
            name="tunnel_name"
            label="隧道名称"
            rules={[{ required: true, message: '请输入隧道名称' }]}
            extra="用于标识隧道，建议使用英文"
          >
            <Input placeholder="如：alipay-callback" disabled={isWaitingAuth} />
          </Form.Item>

          <Form.Item
            name="domain"
            label="主域名"
            rules={[{ required: true, message: '请输入主域名' }]}
            extra="您在 Cloudflare 中托管的域名"
          >
            <Input placeholder="如：example.com" disabled={isWaitingAuth} />
          </Form.Item>

          <Form.Item
            name="subdomain"
            label="子域名前缀"
            rules={[{ required: true, message: '请输入子域名前缀' }]}
            extra="最终域名将是: 子域名.主域名"
          >
            <Input placeholder="如：pay" addonAfter=".您的域名" disabled={isWaitingAuth} />
          </Form.Item>

          <Form.Item
            name="local_host"
            label="本地IP/主机"
            extra="后端服务运行的IP地址，默认 localhost"
          >
            <Input placeholder="localhost" disabled={isWaitingAuth} />
          </Form.Item>

          <Form.Item
            name="local_port"
            label="本地端口"
            rules={[{ required: true, message: '请输入本地端口' }]}
            extra="后端服务运行的端口"
          >
            <InputNumber min={1} max={65535} style={{ width: '100%' }} disabled={isWaitingAuth} />
          </Form.Item>

          <Form.Item className="mb-0 mt-6">
            <Space className="w-full justify-end">
              <Button onClick={() => {
                setTunnelModalOpen(false)
                tunnelForm.resetFields()
                setPendingTunnelData(null)
                setIsWaitingAuth(false)
              }}>
                取消
              </Button>
              <Button
                type="primary"
                htmlType="submit"
                loading={createTunnelMutation.isPending || isWaitingAuth}
              >
                {isWaitingAuth ? '等待授权...' : '创建隧道'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default AlipaySettings

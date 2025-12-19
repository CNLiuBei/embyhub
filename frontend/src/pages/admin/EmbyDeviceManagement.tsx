import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, Tag, Space, Popconfirm, App, Card, Statistic, Alert, Modal, Form, Input, Switch, Table, InputNumber, Tabs, Row, Col, Tooltip } from 'antd'
import { SyncOutlined, PlusOutlined, GlobalOutlined, LockOutlined, UnlockOutlined, DeleteOutlined, CheckCircleOutlined, CloseCircleOutlined, AppstoreOutlined, PlayCircleOutlined, ThunderboltOutlined, TeamOutlined, DesktopOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'

interface ClientWhitelistItem {
  name: string
  display_name: string
  enabled: boolean
}

interface ClientWhitelistSettings {
  enabled: boolean
  clients: ClientWhitelistItem[]
}

interface PlayLimitSettings {
  enabled: boolean
  max_playing: number
  speed_enabled: boolean
  speed_user: number
  speed_member: number
  speed_admin: number
}

const EmbyDeviceManagement = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [addClientModal, setAddClientModal] = useState(false)
  const [form] = Form.useForm()

  // 获取客户端白名单设置
  const { data: clientWhitelist, isLoading: whitelistLoading, refetch: refetchWhitelist } = useQuery<ClientWhitelistSettings>({
    queryKey: ['clientWhitelist'],
    queryFn: async () => {
      const response = await adminApi.getClientWhitelistSettings()
      return response.data.data as ClientWhitelistSettings
    },
  })

  // 保存客户端白名单设置
  const saveWhitelistMutation = useMutation({
    mutationFn: (data: ClientWhitelistSettings) => adminApi.saveClientWhitelistSettings(data),
    onSuccess: () => {
      message.success('设置已保存')
      queryClient.invalidateQueries({ queryKey: ['clientWhitelist'] })
    },
    onError: () => {
      message.error('保存失败')
    },
  })

  // 添加客户端到白名单
  const addClientMutation = useMutation({
    mutationFn: (data: { name: string; displayName: string }) => 
      adminApi.addClientToWhitelist(data.name, data.displayName),
    onSuccess: () => {
      message.success('客户端已添加')
      queryClient.invalidateQueries({ queryKey: ['clientWhitelist'] })
      setAddClientModal(false)
      form.resetFields()
    },
    onError: () => {
      message.error('添加失败')
    },
  })

  // 删除客户端
  const removeClientMutation = useMutation({
    mutationFn: (name: string) => adminApi.removeClientFromWhitelist(name),
    onSuccess: () => {
      message.success('客户端已删除')
      queryClient.invalidateQueries({ queryKey: ['clientWhitelist'] })
    },
    onError: () => {
      message.error('删除失败')
    },
  })

  // 更新客户端状态
  const updateClientStatusMutation = useMutation({
    mutationFn: (data: { name: string; enabled: boolean }) => 
      adminApi.updateClientStatus(data.name, data.enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clientWhitelist'] })
    },
    onError: () => {
      message.error('更新失败')
    },
  })

  // 获取播放限制设置
  const { data: playLimit, isLoading: playLimitLoading } = useQuery<PlayLimitSettings>({
    queryKey: ['playLimit'],
    queryFn: async () => {
      const response = await adminApi.getPlayLimitSettings()
      return response.data.data as PlayLimitSettings
    },
  })

  // 保存播放限制设置
  const savePlayLimitMutation = useMutation({
    mutationFn: (data: PlayLimitSettings) => adminApi.savePlayLimitSettings(data),
    onSuccess: () => {
      message.success('设置已保存')
      queryClient.invalidateQueries({ queryKey: ['playLimit'] })
    },
    onError: () => {
      message.error('保存失败')
    },
  })

  // 统计数据
  const totalClients = clientWhitelist?.clients?.length || 0
  const enabledClientsCount = clientWhitelist?.clients?.filter(c => c.enabled).length || 0

  // 客户端表格列
  const clientColumns = [
    {
      title: '客户端',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: ClientWhitelistItem) => (
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white">
            <DesktopOutlined />
          </div>
          <div>
            <div className="font-medium">{record.display_name || name}</div>
            <div className="text-xs text-gray-400">标识: {name}</div>
          </div>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled: boolean) => (
        enabled ? (
          <Tag color="success" icon={<CheckCircleOutlined />}>已授权</Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />}>未授权</Tag>
        )
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: ClientWhitelistItem) => (
        <Space>
          <Switch
            checked={record.enabled}
            loading={updateClientStatusMutation.isPending}
            onChange={(checked) => {
              updateClientStatusMutation.mutate({ name: record.name, enabled: checked })
            }}
            size="small"
          />
          <Popconfirm
            title="确定删除此客户端？"
            onConfirm={() => removeClientMutation.mutate(record.name)}
            okText="确定"
            cancelText="取消"
          >
            <Button 
              danger 
              type="text"
              icon={<DeleteOutlined />} 
              size="small"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // 客户端白名单 Tab
  const ClientWhitelistTab = () => (
    <div className="space-y-4">
      {/* 代理地址提示 */}
      <div className="p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-100">
        <div className="flex items-start gap-3">
          <GlobalOutlined className="text-2xl text-blue-500 mt-1" />
          <div className="flex-1">
            <div className="font-medium text-gray-800 mb-1">代理服务地址</div>
            <code className="text-lg font-mono bg-white px-3 py-1 rounded border">
              {window.location.hostname}:54682
            </code>
            <div className="text-xs text-gray-500 mt-2">
              用户在播放器中填写此地址即可，系统会自动转发到 Emby 服务器并进行客户端验证
            </div>
          </div>
        </div>
      </div>

      {/* 启用开关 */}
      <Card size="small" className="!rounded-xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {clientWhitelist?.enabled ? (
              <div className="w-10 h-10 rounded-lg bg-orange-100 flex items-center justify-center">
                <LockOutlined className="text-orange-500 text-lg" />
              </div>
            ) : (
              <div className="w-10 h-10 rounded-lg bg-green-100 flex items-center justify-center">
                <UnlockOutlined className="text-green-500 text-lg" />
              </div>
            )}
            <div>
              <div className="font-medium">客户端白名单</div>
              <div className="text-xs text-gray-400">
                {clientWhitelist?.enabled 
                  ? `仅允许 ${enabledClientsCount} 种授权客户端访问` 
                  : '允许所有客户端访问'}
              </div>
            </div>
          </div>
          <Switch
            checked={clientWhitelist?.enabled || false}
            loading={saveWhitelistMutation.isPending || whitelistLoading}
            onChange={(checked) => {
              if (clientWhitelist) {
                saveWhitelistMutation.mutate({ ...clientWhitelist, enabled: checked })
              }
            }}
          />
        </div>
      </Card>

      {/* 客户端列表 */}
      <Card 
        size="small" 
        className="!rounded-xl"
        title={
          <div className="flex items-center gap-2">
            <AppstoreOutlined />
            <span>客户端列表</span>
            <Tag>{totalClients}</Tag>
          </div>
        }
        extra={
          <Space>
            <Button icon={<SyncOutlined />} onClick={() => refetchWhitelist()} loading={whitelistLoading}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setAddClientModal(true)}>
              添加
            </Button>
          </Space>
        }
      >
        <Table
          columns={clientColumns}
          dataSource={clientWhitelist?.clients || []}
          rowKey="name"
          loading={whitelistLoading}
          pagination={false}
          size="small"
          rowClassName={(record) => record.enabled ? '' : 'opacity-50'}
        />
      </Card>

      {/* 常用客户端 */}
      <Card size="small" className="!rounded-xl" title={<><InfoCircleOutlined className="mr-2" />常用客户端标识</>}>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
          {[
            { name: 'VidHub', desc: 'VidHub' },
            { name: 'SenPlayer', desc: 'SenPlayer' },
            { name: 'Infuse', desc: 'Infuse / Infuse-Direct' },
            { name: 'Emby Web', desc: 'Emby Web' },
            { name: 'Emby Theater', desc: 'Emby Theater' },
            { name: 'Emby for iOS', desc: 'Emby for iOS' },
            { name: 'Emby for Android', desc: 'Emby for Android' },
            { name: 'Fileball', desc: 'Fileball' },
          ].map(item => (
            <div key={item.name} className="p-2 bg-gray-50 rounded-lg">
              <div className="font-medium text-gray-700">{item.name}</div>
              <div className="text-xs text-gray-400">{item.desc}</div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )

  // 播放限制 Tab
  const PlayLimitTab = () => (
    <div className="space-y-4">
      <Row gutter={16}>
        {/* 同时播放数限制 */}
        <Col xs={24} lg={12}>
          <Card size="small" className="!rounded-xl h-full">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-purple-100 flex items-center justify-center">
                  <TeamOutlined className="text-purple-500 text-lg" />
                </div>
                <div>
                  <div className="font-medium">同时播放数限制</div>
                  <div className="text-xs text-gray-400">限制每个用户同时播放的设备数量</div>
                </div>
              </div>
              <Switch
                checked={playLimit?.enabled || false}
                loading={savePlayLimitMutation.isPending || playLimitLoading}
                onChange={(checked) => {
                  savePlayLimitMutation.mutate({
                    enabled: checked,
                    max_playing: playLimit?.max_playing || 1,
                    speed_enabled: playLimit?.speed_enabled || false,
                    speed_user: playLimit?.speed_user || 10,
                    speed_member: playLimit?.speed_member || 30,
                    speed_admin: playLimit?.speed_admin || 0
                  })
                }}
              />
            </div>
            
            <div className="p-4 bg-gray-50 rounded-xl">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-sm text-gray-600">最大同时播放数</div>
                  <div className="text-xs text-gray-400">超出限制时新播放将被阻止</div>
                </div>
                <InputNumber
                  min={1}
                  max={10}
                  value={playLimit?.max_playing || 1}
                  disabled={!playLimit?.enabled}
                  size="large"
                  style={{ width: 100 }}
                  onChange={(value) => {
                    if (value) {
                      savePlayLimitMutation.mutate({
                        enabled: playLimit?.enabled || false,
                        max_playing: value,
                        speed_enabled: playLimit?.speed_enabled || false,
                        speed_user: playLimit?.speed_user || 10,
                        speed_member: playLimit?.speed_member || 30,
                        speed_admin: playLimit?.speed_admin || 0
                      })
                    }
                  }}
                />
              </div>
            </div>

            <Alert
              type="info"
              message="当同一账号在多个设备播放时，系统会统计所有设备的播放会话数"
              className="mt-4 !rounded-lg"
              showIcon
            />
          </Card>
        </Col>

        {/* 播放速率限制 */}
        <Col xs={24} lg={12}>
          <Card size="small" className="!rounded-xl h-full">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-orange-100 flex items-center justify-center">
                  <ThunderboltOutlined className="text-orange-500 text-lg" />
                </div>
                <div>
                  <div className="font-medium">播放速率限制</div>
                  <div className="text-xs text-gray-400">按用户角色限制视频流带宽</div>
                </div>
              </div>
              <Switch
                checked={playLimit?.speed_enabled || false}
                loading={savePlayLimitMutation.isPending || playLimitLoading}
                onChange={(checked) => {
                  savePlayLimitMutation.mutate({
                    enabled: playLimit?.enabled || false,
                    max_playing: playLimit?.max_playing || 1,
                    speed_enabled: checked,
                    speed_user: playLimit?.speed_user || 10,
                    speed_member: playLimit?.speed_member || 30,
                    speed_admin: playLimit?.speed_admin || 0
                  })
                }}
              />
            </div>

            <div className="space-y-3">
              {/* 普通用户 */}
              <div className="p-3 bg-gray-50 rounded-xl flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-gray-400"></span>
                  <div>
                    <div className="text-sm font-medium">普通用户</div>
                    <div className="text-xs text-gray-400">未开通会员</div>
                  </div>
                </div>
                <InputNumber
                  min={1}
                  max={100}
                  value={playLimit?.speed_user || 5}
                  disabled={!playLimit?.speed_enabled}
                  addonAfter="MB/s"
                  style={{ width: 130 }}
                  onChange={(value) => {
                    if (value) {
                      savePlayLimitMutation.mutate({
                        enabled: playLimit?.enabled || false,
                        max_playing: playLimit?.max_playing || 1,
                        speed_enabled: playLimit?.speed_enabled || false,
                        speed_user: value,
                        speed_member: playLimit?.speed_member || 10,
                        speed_admin: playLimit?.speed_admin || 0
                      })
                    }
                  }}
                />
              </div>

              {/* 会员用户 */}
              <div className="p-3 bg-yellow-50 rounded-xl flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-yellow-500"></span>
                  <div>
                    <div className="text-sm font-medium">会员用户</div>
                    <div className="text-xs text-gray-400">已开通会员</div>
                  </div>
                </div>
                <InputNumber
                  min={1}
                  max={100}
                  value={playLimit?.speed_member || 10}
                  disabled={!playLimit?.speed_enabled}
                  addonAfter="MB/s"
                  style={{ width: 130 }}
                  onChange={(value) => {
                    if (value) {
                      savePlayLimitMutation.mutate({
                        enabled: playLimit?.enabled || false,
                        max_playing: playLimit?.max_playing || 1,
                        speed_enabled: playLimit?.speed_enabled || false,
                        speed_user: playLimit?.speed_user || 5,
                        speed_member: value,
                        speed_admin: playLimit?.speed_admin || 0
                      })
                    }
                  }}
                />
              </div>

              {/* 管理员 */}
              <div className="p-3 bg-blue-50 rounded-xl flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-blue-500"></span>
                  <div>
                    <div className="text-sm font-medium">管理员</div>
                    <div className="text-xs text-gray-400">
                      <Tooltip title="设置为 0 表示不限制">
                        0 = 不限制 <InfoCircleOutlined />
                      </Tooltip>
                    </div>
                  </div>
                </div>
                <InputNumber
                  min={0}
                  max={100}
                  value={playLimit?.speed_admin || 0}
                  disabled={!playLimit?.speed_enabled}
                  addonAfter="MB/s"
                  style={{ width: 130 }}
                  onChange={(value) => {
                    if (value !== null && value !== undefined) {
                      savePlayLimitMutation.mutate({
                        enabled: playLimit?.enabled || false,
                        max_playing: playLimit?.max_playing || 1,
                        speed_enabled: playLimit?.speed_enabled || false,
                        speed_user: playLimit?.speed_user || 5,
                        speed_member: playLimit?.speed_member || 10,
                        speed_admin: value
                      })
                    }
                  }}
                />
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )

  return (
    <div className="flex flex-col h-full gap-4">
      {/* 顶部统计 */}
      <Row gutter={16}>
        <Col xs={12} sm={6}>
          <Card className="glass-card !rounded-xl">
            <Statistic 
              title="客户端总数" 
              value={totalClients} 
              prefix={<AppstoreOutlined className="text-blue-500" />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="glass-card !rounded-xl">
            <Statistic 
              title="已授权" 
              value={enabledClientsCount} 
              prefix={<CheckCircleOutlined className="text-green-500" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="glass-card !rounded-xl">
            <Statistic 
              title="播放限制" 
              value={playLimit?.enabled ? playLimit.max_playing : '无'} 
              suffix={playLimit?.enabled ? '个' : ''}
              prefix={<TeamOutlined className="text-purple-500" />}
              valueStyle={{ color: playLimit?.enabled ? '#722ed1' : undefined }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="glass-card !rounded-xl">
            <Statistic 
              title="速率限制" 
              value={playLimit?.speed_enabled ? '已启用' : '未启用'} 
              prefix={<ThunderboltOutlined className="text-orange-500" />}
              valueStyle={{ color: playLimit?.speed_enabled ? '#fa8c16' : undefined, fontSize: 20 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 主内容 Tabs */}
      <Card className="glass-card flex-1 !rounded-xl">
        <Tabs
          defaultActiveKey="clients"
          items={[
            {
              key: 'clients',
              label: (
                <span className="flex items-center gap-2">
                  <AppstoreOutlined />
                  客户端授权
                </span>
              ),
              children: <ClientWhitelistTab />,
            },
            {
              key: 'limits',
              label: (
                <span className="flex items-center gap-2">
                  <PlayCircleOutlined />
                  播放限制
                </span>
              ),
              children: <PlayLimitTab />,
            },
          ]}
        />
      </Card>

      {/* 添加客户端弹窗 */}
      <Modal
        title={
          <div className="flex items-center gap-2">
            <PlusOutlined />
            添加客户端
          </div>
        }
        open={addClientModal}
        onCancel={() => {
          setAddClientModal(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        confirmLoading={addClientMutation.isPending}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => {
            addClientMutation.mutate({
              name: values.name,
              displayName: values.displayName || values.name
            })
          }}
        >
          <Form.Item
            name="name"
            label="客户端标识"
            rules={[{ required: true, message: '请输入客户端标识' }]}
            extra="用于匹配客户端请求头中的 Client 字段"
          >
            <Input placeholder="如: VidHub, SenPlayer, Infuse" />
          </Form.Item>
          <Form.Item
            name="displayName"
            label="显示名称"
            extra="可选，用于在列表中显示的友好名称"
          >
            <Input placeholder="如: VidHub 播放器" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default EmbyDeviceManagement

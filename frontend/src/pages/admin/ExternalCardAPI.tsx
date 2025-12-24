import { useState, useEffect, useId } from 'react'
import { Card, Form, Input, InputNumber, Switch, Button, Table, Space, Alert, App, Tag, Tooltip, Typography, Modal, Popconfirm } from 'antd'
import { ApiOutlined, SaveOutlined, ReloadOutlined, CopyOutlined, KeyOutlined, HistoryOutlined, PlusOutlined, DeleteOutlined, EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons'
import { externalCardApi } from '../../services/api'

const { Text } = Typography

interface ExternalCardAPISettings {
  enabled: boolean
  api_key: string
  api_keys?: APIKeyItem[]  // 支持多个API密钥
  allowed_ips: string
  rate_limit: number
  default_type: number
  log_enabled: boolean
}

interface APIKeyItem {
  id: string
  name: string
  key: string
  enabled: boolean
  created_at: string
  last_used_at?: string
}

interface APILog {
  id: number
  ip: string
  method: string
  path: string
  params: string
  response: string
  status: number
  duration: number
  created_at: string
}

const ExternalCardAPI = () => {
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [logsLoading, setLogsLoading] = useState(false)
  const [logs, setLogs] = useState<APILog[]>([])
  const [logsTotal, setLogsTotal] = useState(0)
  const [logsPage, setLogsPage] = useState(1)
  const [form] = Form.useForm()
  const { message } = App.useApp()
  
  // 多API密钥管理
  const [apiKeys, setApiKeys] = useState<APIKeyItem[]>([])
  const [showApiKey, setShowApiKey] = useState<Record<string, boolean>>({})
  const [addKeyModalVisible, setAddKeyModalVisible] = useState(false)
  const [addKeyForm] = Form.useForm()
  const [currentApiKey, setCurrentApiKey] = useState('')  // 当前显示的API密钥
  
  // 使用唯一ID避免StrictMode双重渲染导致的重复ID警告
  const formId = useId()

  useEffect(() => {
    loadSettings()
    loadLogs()
  }, [])

  const loadSettings = async () => {
    try {
      setLoading(true)
      const response = await externalCardApi.getSettings()
      const data = response.data.data as ExternalCardAPISettings
      form.setFieldsValue(data)
      // 设置当前API密钥用于显示
      setCurrentApiKey(data.api_key || '')
      // 加载多API密钥列表
      if (data.api_keys) {
        setApiKeys(data.api_keys)
      }
    } catch (err) {
      message.error('加载设置失败')
    } finally {
      setLoading(false)
    }
  }

  const loadLogs = async (page = 1) => {
    try {
      setLogsLoading(true)
      const response = await externalCardApi.getLogs({ page, page_size: 10 })
      setLogs(response.data.data?.list || [])
      setLogsTotal(response.data.data?.total || 0)
      setLogsPage(page)
    } catch (err) {
      message.error('加载日志失败')
    } finally {
      setLogsLoading(false)
    }
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      setSaving(true)
      await externalCardApi.saveSettings(values)
      message.success('保存成功')
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(String(err) || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleGenerateKey = async () => {
    try {
      setGenerating(true)
      const response = await externalCardApi.generateAPIKey()
      const newKey = response.data.data?.api_key
      if (newKey) {
        // 更新表单值和当前显示的密钥
        form.setFieldValue('api_key', newKey)
        setCurrentApiKey(newKey)
        message.success('已生成新的API密钥')
      } else {
        message.error('生成失败：未返回密钥')
      }
    } catch (err) {
      message.error('生成失败')
    } finally {
      setGenerating(false)
    }
  }

  // 添加新的API密钥
  const handleAddApiKey = async () => {
    try {
      const values = await addKeyForm.validateFields()
      const response = await externalCardApi.generateAPIKey()
      const newKey = response.data.data?.api_key
      if (newKey) {
        const newKeyItem: APIKeyItem = {
          id: Date.now().toString(),
          name: values.name,
          key: newKey,
          enabled: true,
          created_at: new Date().toISOString(),
        }
        const newApiKeys = [...apiKeys, newKeyItem]
        setApiKeys(newApiKeys)
        // 保存到后端
        await externalCardApi.saveSettings({
          ...form.getFieldsValue(),
          api_keys: newApiKeys,
        })
        setAddKeyModalVisible(false)
        addKeyForm.resetFields()
        message.success('API密钥添加成功')
      }
    } catch (err) {
      message.error('添加失败')
    }
  }

  // 删除API密钥
  const handleDeleteApiKey = async (id: string) => {
    const newApiKeys = apiKeys.filter(k => k.id !== id)
    setApiKeys(newApiKeys)
    try {
      await externalCardApi.saveSettings({
        ...form.getFieldsValue(),
        api_keys: newApiKeys,
      })
      message.success('删除成功')
    } catch {
      message.error('删除失败')
    }
  }

  // 切换API密钥启用状态
  const handleToggleApiKey = async (id: string, enabled: boolean) => {
    const newApiKeys = apiKeys.map(k => k.id === id ? { ...k, enabled } : k)
    setApiKeys(newApiKeys)
    try {
      await externalCardApi.saveSettings({
        ...form.getFieldsValue(),
        api_keys: newApiKeys,
      })
      message.success(enabled ? '已启用' : '已禁用')
    } catch {
      message.error('操作失败')
    }
  }

  // 切换密钥显示/隐藏
  const toggleShowKey = (id: string) => {
    setShowApiKey(prev => ({ ...prev, [id]: !prev[id] }))
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
  }

  const logColumns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
      width: 140,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      width: 200,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => (
        <Tag color={status === 200 ? 'green' : status === 401 ? 'orange' : 'red'}>
          {status}
        </Tag>
      ),
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      width: 80,
      render: (ms: number) => `${ms}ms`,
    },
    {
      title: '响应',
      dataIndex: 'response',
      key: 'response',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip title={text}>
          <Text ellipsis style={{ maxWidth: 200 }}>{text}</Text>
        </Tooltip>
      ),
    },
  ]

  const cardTypeOptions = [
    { value: 1, label: '月卡' },
    { value: 2, label: '季卡' },
    { value: 3, label: '半年卡' },
    { value: 4, label: '年卡' },
  ]

  return (
    <div className="space-y-4">
      <Card
        title={
          <span>
            <ApiOutlined className="mr-2" />
            外部卡密API设置
          </span>
        }
        loading={loading}
      >
        <Alert
          message="功能说明"
          description={
            <div className="text-sm">
              <p>外部卡密API允许第三方系统（如闲鱼自动回复系统）调用本系统获取卡密，实现自动发货。</p>
              <p className="text-blue-600 mt-2">💡 第三方系统调用示例：</p>
              <pre className="bg-gray-100 p-2 rounded mt-1 text-xs overflow-x-auto">
{`GET /api/external/card/fetch?type=1&count=1
Authorization: Bearer YOUR_API_KEY`}
              </pre>
              <p className="text-orange-500 mt-2">⚠️ 请妥善保管API密钥，泄露可能导致卡密被盗取</p>
            </div>
          }
          type="info"
          showIcon
          className="mb-6"
        />

        <Form form={form} name={`externalCardAPIForm${formId}`} layout="vertical" className="max-w-2xl">
          <Form.Item name="enabled" label="启用外部API" valuePropName="checked">
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>

          <Form.Item 
            name="api_key" 
            label="主API密钥"
            extra="第三方系统调用时需要在请求头中携带此密钥"
          >
            <Space.Compact className="w-full">
              <Input 
                value={currentApiKey}
                onChange={(e) => {
                  setCurrentApiKey(e.target.value)
                  form.setFieldValue('api_key', e.target.value)
                }}
                placeholder="点击生成按钮生成API密钥" 
                className="flex-1"
                type={showApiKey['main'] ? 'text' : 'password'}
                autoComplete="off"
                suffix={
                  <span 
                    onClick={() => toggleShowKey('main')} 
                    style={{ cursor: 'pointer' }}
                  >
                    {showApiKey['main'] ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                  </span>
                }
              />
              <Button 
                icon={<KeyOutlined />} 
                onClick={handleGenerateKey}
                loading={generating}
              >
                生成新密钥
              </Button>
              <Button 
                icon={<CopyOutlined />} 
                onClick={() => copyToClipboard(currentApiKey || '')}
              >
                复制
              </Button>
            </Space.Compact>
          </Form.Item>

          <Form.Item 
            name="allowed_ips" 
            label="IP白名单"
            extra="允许调用API的IP地址，多个IP用逗号分隔，留空表示不限制"
          >
            <Input placeholder="如：10.10.10.68,192.168.1.100" />
          </Form.Item>

          <Form.Item 
            name="rate_limit" 
            label="请求频率限制"
            extra="每分钟最大请求次数，0表示不限制"
          >
            <InputNumber min={0} max={1000} placeholder="60" className="w-full" />
          </Form.Item>

          <Form.Item 
            name="default_type" 
            label="默认卡密类型"
            extra="当第三方系统未指定类型时使用的默认类型"
          >
            <select className="w-full border rounded px-3 py-2">
              {cardTypeOptions.map(opt => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </Form.Item>

          <Form.Item name="log_enabled" label="记录请求日志" valuePropName="checked">
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
              保存设置
            </Button>
          </Form.Item>
        </Form>

        {/* 多API密钥管理 */}
        <div className="mt-6 border-t pt-6">
          <div className="flex justify-between items-center mb-4">
            <h4 className="font-bold text-base">多API密钥管理</h4>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={() => setAddKeyModalVisible(true)}
            >
              添加API密钥
            </Button>
          </div>
          <Alert
            message="多密钥说明"
            description="您可以创建多个API密钥分配给不同的第三方系统使用，便于管理和追踪。每个密钥可以单独启用/禁用。"
            type="info"
            showIcon
            className="mb-4"
          />
          <Table
            dataSource={apiKeys}
            rowKey="id"
            size="small"
            pagination={false}
            columns={[
              {
                title: '名称',
                dataIndex: 'name',
                key: 'name',
                width: 150,
              },
              {
                title: 'API密钥',
                dataIndex: 'key',
                key: 'key',
                render: (key: string, record: APIKeyItem) => (
                  <Space>
                    <Text copyable={{ text: key }}>
                      {showApiKey[record.id] ? key : key.substring(0, 8) + '****' + key.substring(key.length - 4)}
                    </Text>
                    <Button 
                      type="text" 
                      size="small"
                      icon={showApiKey[record.id] ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                      onClick={() => toggleShowKey(record.id)}
                    />
                  </Space>
                ),
              },
              {
                title: '状态',
                dataIndex: 'enabled',
                key: 'enabled',
                width: 100,
                render: (enabled: boolean, record: APIKeyItem) => (
                  <Switch 
                    checked={enabled} 
                    onChange={(checked) => handleToggleApiKey(record.id, checked)}
                    checkedChildren="启用"
                    unCheckedChildren="禁用"
                  />
                ),
              },
              {
                title: '创建时间',
                dataIndex: 'created_at',
                key: 'created_at',
                width: 180,
                render: (text: string) => text ? new Date(text).toLocaleString() : '-',
              },
              {
                title: '操作',
                key: 'action',
                width: 80,
                render: (_: unknown, record: APIKeyItem) => (
                  <Popconfirm
                    title="确定删除此API密钥？"
                    onConfirm={() => handleDeleteApiKey(record.id)}
                    okText="确定"
                    cancelText="取消"
                  >
                    <Button type="text" danger icon={<DeleteOutlined />} size="small" />
                  </Popconfirm>
                ),
              },
            ]}
            locale={{ emptyText: '暂无额外的API密钥' }}
          />
        </div>
      </Card>

      <Card
        title={
          <span>
            <HistoryOutlined className="mr-2" />
            API调用日志
          </span>
        }
        extra={
          <Button icon={<ReloadOutlined />} onClick={() => loadLogs(logsPage)} loading={logsLoading}>
            刷新
          </Button>
        }
      >
        <Table
          columns={logColumns}
          dataSource={logs}
          rowKey="id"
          loading={logsLoading}
          pagination={{
            current: logsPage,
            total: logsTotal,
            pageSize: 10,
            onChange: (page) => loadLogs(page),
            showTotal: (total) => `共 ${total} 条`,
          }}
          scroll={{ x: 900 }}
          size="small"
        />
      </Card>

      <Card title="API接口文档">
        <div className="space-y-4 text-sm">
          <div>
            <h4 className="font-bold mb-2">1. 获取卡密</h4>
            <pre className="bg-gray-100 p-3 rounded overflow-x-auto">
{`GET /api/external/card/fetch
Headers:
  Authorization: Bearer YOUR_API_KEY

Query Parameters:
  type: 卡密类型（1月卡 2季卡 3半年卡 4年卡），可选
  count: 获取数量（1-10），可选，默认1

Response:
{
  "success": true,
  "message": "获取成功",
  "data": {
    "code": "True-XXXXXXXXXXXXXXXXXXXXXXXX",
    "type": 1,
    "type_name": "月卡",
    "duration": 30
  }
}`}
            </pre>
          </div>

          <div>
            <h4 className="font-bold mb-2">2. 查询库存</h4>
            <pre className="bg-gray-100 p-3 rounded overflow-x-auto">
{`GET /api/external/card/stock
Headers:
  Authorization: Bearer YOUR_API_KEY

Response:
{
  "success": true,
  "message": "获取成功",
  "data": {
    "月卡": 100,
    "季卡": 50,
    "半年卡": 30,
    "年卡": 20
  }
}`}
            </pre>
          </div>
        </div>
      </Card>

      {/* 添加API密钥弹窗 */}
      <Modal
        title="添加API密钥"
        open={addKeyModalVisible}
        onOk={handleAddApiKey}
        onCancel={() => {
          setAddKeyModalVisible(false)
          addKeyForm.resetFields()
        }}
        okText="生成并添加"
        cancelText="取消"
      >
        <Form form={addKeyForm} name={`addApiKeyForm${formId}`} layout="vertical">
          <Form.Item
            name="name"
            label="密钥名称"
            rules={[{ required: true, message: '请输入密钥名称' }]}
            extra="用于标识此密钥的用途，如：闲鱼自动发货、淘宝店铺等"
          >
            <Input placeholder="如：闲鱼自动发货" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ExternalCardAPI

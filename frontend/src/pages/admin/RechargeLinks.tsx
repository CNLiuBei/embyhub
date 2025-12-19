import { useState, useEffect } from 'react'
import { Card, Input, Switch, Button, Alert, App, Table, Tag, Tabs, InputNumber, Space } from 'antd'
import { LinkOutlined, SaveOutlined, ShoppingCartOutlined, TrophyOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'

interface RechargeLink {
  card_type: number
  name: string
  url: string
  enabled: boolean
}

interface PointsRechargeLink {
  points: number
  name: string
  url: string
  enabled: boolean
}

const cardTypeConfig: Record<number, { color: string; label: string }> = {
  1: { color: 'blue', label: '月卡' },
  2: { color: 'cyan', label: '季卡' },
  3: { color: 'purple', label: '半年卡' },
  4: { color: 'gold', label: '年卡' },
}

const RechargeLinks = () => {
  const { message } = App.useApp()
  const [activeTab, setActiveTab] = useState('member')
  
  // 会员卡链接
  const [memberLoading, setMemberLoading] = useState(false)
  const [memberSaving, setMemberSaving] = useState(false)
  const [memberLinks, setMemberLinks] = useState<RechargeLink[]>([
    { card_type: 1, name: '月卡', url: '', enabled: false },
    { card_type: 2, name: '季卡', url: '', enabled: false },
    { card_type: 3, name: '半年卡', url: '', enabled: false },
    { card_type: 4, name: '年卡', url: '', enabled: false },
  ])

  // 积分卡链接
  const [pointsLoading, setPointsLoading] = useState(false)
  const [pointsSaving, setPointsSaving] = useState(false)
  const [pointsLinks, setPointsLinks] = useState<PointsRechargeLink[]>([
    { points: 100, name: '100积分', url: '', enabled: false },
    { points: 500, name: '500积分', url: '', enabled: false },
    { points: 1000, name: '1000积分', url: '', enabled: false },
    { points: 5000, name: '5000积分', url: '', enabled: false },
  ])

  useEffect(() => {
    if (activeTab === 'member') {
      loadMemberSettings()
    } else {
      loadPointsSettings()
    }
  }, [activeTab])

  // 加载会员卡链接设置
  const loadMemberSettings = async () => {
    try {
      setMemberLoading(true)
      const response = await adminApi.getRechargeLinksSettings()
      const data = response.data.data as { links: RechargeLink[] }
      if (data.links && data.links.length > 0) {
        setMemberLinks(data.links)
      }
    } catch {
      message.error('加载设置失败')
    } finally {
      setMemberLoading(false)
    }
  }

  // 加载积分卡链接设置
  const loadPointsSettings = async () => {
    try {
      setPointsLoading(true)
      const response = await adminApi.getPointsRechargeLinksSettings()
      const data = response.data.data as { links: PointsRechargeLink[] }
      if (data.links && data.links.length > 0) {
        setPointsLinks(data.links)
      }
    } catch {
      message.error('加载设置失败')
    } finally {
      setPointsLoading(false)
    }
  }

  // 保存会员卡链接
  const handleMemberSave = async () => {
    try {
      setMemberSaving(true)
      await adminApi.saveRechargeLinksSettings({ links: memberLinks })
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    } finally {
      setMemberSaving(false)
    }
  }

  // 保存积分卡链接
  const handlePointsSave = async () => {
    try {
      setPointsSaving(true)
      await adminApi.savePointsRechargeLinksSettings({ links: pointsLinks })
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    } finally {
      setPointsSaving(false)
    }
  }

  const handleMemberLinkChange = (cardType: number, field: 'url' | 'enabled' | 'name', value: string | boolean) => {
    setMemberLinks(prev => prev.map(link => 
      link.card_type === cardType ? { ...link, [field]: value } : link
    ))
  }

  const handlePointsLinkChange = (index: number, field: 'url' | 'enabled' | 'name' | 'points', value: string | boolean | number) => {
    setPointsLinks(prev => prev.map((link, i) => 
      i === index ? { ...link, [field]: value } : link
    ))
  }

  const addPointsLink = () => {
    setPointsLinks(prev => [...prev, { points: 100, name: '100积分', url: '', enabled: false }])
  }

  const removePointsLink = (index: number) => {
    setPointsLinks(prev => prev.filter((_, i) => i !== index))
  }

  const memberColumns = [
    {
      title: '卡密类型',
      dataIndex: 'card_type',
      key: 'card_type',
      width: 120,
      render: (type: number) => (
        <Tag color={cardTypeConfig[type]?.color || 'default'}>
          {cardTypeConfig[type]?.label || '未知'}
        </Tag>
      ),
    },
    {
      title: '显示名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (name: string, record: RechargeLink) => (
        <Input
          value={name}
          onChange={(e) => handleMemberLinkChange(record.card_type, 'name', e.target.value)}
          placeholder="显示名称"
        />
      ),
    },
    {
      title: '购买链接',
      dataIndex: 'url',
      key: 'url',
      render: (url: string, record: RechargeLink) => (
        <Input
          value={url}
          onChange={(e) => handleMemberLinkChange(record.card_type, 'url', e.target.value)}
          placeholder="输入购买链接"
          prefix={<LinkOutlined className="text-gray-400" />}
        />
      ),
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (enabled: boolean, record: RechargeLink) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleMemberLinkChange(record.card_type, 'enabled', checked)}
          size="small"
        />
      ),
    },
    {
      title: '预览',
      key: 'preview',
      width: 80,
      render: (_: unknown, record: RechargeLink) => (
        record.url && record.enabled ? (
          <Button type="link" size="small" onClick={() => window.open(record.url, '_blank')}>
            测试
          </Button>
        ) : <span className="text-gray-400">-</span>
      ),
    },
  ]

  const pointsColumns = [
    {
      title: '积分数量',
      dataIndex: 'points',
      key: 'points',
      width: 120,
      render: (points: number, _: PointsRechargeLink, index: number) => (
        <InputNumber
          value={points}
          onChange={(val) => handlePointsLinkChange(index, 'points', val || 0)}
          min={1}
          className="w-full"
        />
      ),
    },
    {
      title: '显示名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (name: string, _: PointsRechargeLink, index: number) => (
        <Input
          value={name}
          onChange={(e) => handlePointsLinkChange(index, 'name', e.target.value)}
          placeholder="显示名称"
        />
      ),
    },
    {
      title: '购买链接',
      dataIndex: 'url',
      key: 'url',
      render: (url: string, _: PointsRechargeLink, index: number) => (
        <Input
          value={url}
          onChange={(e) => handlePointsLinkChange(index, 'url', e.target.value)}
          placeholder="输入购买链接"
          prefix={<LinkOutlined className="text-gray-400" />}
        />
      ),
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (enabled: boolean, _: PointsRechargeLink, index: number) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handlePointsLinkChange(index, 'enabled', checked)}
          size="small"
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: PointsRechargeLink, index: number) => (
        <Space>
          {record.url && record.enabled && (
            <Button type="link" size="small" onClick={() => window.open(record.url, '_blank')}>
              测试
            </Button>
          )}
          <Button type="link" size="small" danger onClick={() => removePointsLink(index)}>
            <DeleteOutlined />
          </Button>
        </Space>
      ),
    },
  ]

  const tabItems = [
    {
      key: 'member',
      label: <span><ShoppingCartOutlined className="mr-1" />会员卡购买</span>,
      children: (
        <div>
          <Alert
            message="会员卡购买链接"
            description="配置不同会员卡类型的购买链接，用户在会员中心可以看到对应的购买按钮。"
            type="info"
            showIcon
            className="mb-4"
          />
          <Table
            columns={memberColumns}
            dataSource={memberLinks}
            rowKey="card_type"
            pagination={false}
            loading={memberLoading}
            size="small"
          />
          <div className="mt-4">
            <Button type="primary" icon={<SaveOutlined />} onClick={handleMemberSave} loading={memberSaving}>
              保存会员卡链接
            </Button>
          </div>
        </div>
      ),
    },
    {
      key: 'points',
      label: <span><TrophyOutlined className="mr-1" />积分卡购买</span>,
      children: (
        <div>
          <Alert
            message="积分卡购买链接"
            description="配置不同积分数量的购买链接，用户在积分页面可以看到对应的购买按钮。"
            type="info"
            showIcon
            className="mb-4"
          />
          <Table
            columns={pointsColumns}
            dataSource={pointsLinks}
            rowKey={(_, index) => index?.toString() || '0'}
            pagination={false}
            loading={pointsLoading}
            size="small"
          />
          <div className="mt-4 flex gap-2">
            <Button icon={<PlusOutlined />} onClick={addPointsLink}>
              添加积分档位
            </Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handlePointsSave} loading={pointsSaving}>
              保存积分卡链接
            </Button>
          </div>
        </div>
      ),
    },
  ]

  return (
    <Card
      title={
        <span>
          <ShoppingCartOutlined className="mr-2" />
          充值链接管理
        </span>
      }
    >
      <Tabs items={tabItems} activeKey={activeTab} onChange={setActiveTab} />
    </Card>
  )
}

export default RechargeLinks

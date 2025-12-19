import { useQuery } from '@tanstack/react-query'
import { Card, Row, Col, Statistic, Tag, Spin, Button } from 'antd'
import { 
  CheckCircleOutlined, 
  CloseCircleOutlined, 
  ReloadOutlined,
  UserOutlined,
  CrownOutlined,
  CreditCardOutlined,
  LoginOutlined,
  NotificationOutlined,
  StopOutlined,
} from '@ant-design/icons'
import { adminApi } from '../../services/api'

interface HealthStatus {
  status: string
  uptime: string
  timestamp: string
  services: Record<string, string>
  system: {
    go_version: string
    num_goroutine: number
    num_cpu: number
    mem_alloc: string
    mem_sys: string
  }
}

interface SystemStats {
  total_users: number
  total_members: number
  total_cards: number
  used_cards: number
  today_logins: number
  today_registers: number
  announcements: number
  blocked_ips: number
}

const SystemMonitor = () => {
  const { data: health, isLoading: healthLoading, refetch: refetchHealth } = useQuery({
    queryKey: ['systemHealth'],
    queryFn: async () => {
      const response = await adminApi.getSystemHealth()
      return response.data.data as HealthStatus
    },
    refetchInterval: 30000, // 30秒刷新一次
  })

  const { data: stats, isLoading: statsLoading, refetch: refetchStats } = useQuery({
    queryKey: ['systemStats'],
    queryFn: async () => {
      const response = await adminApi.getSystemStats()
      return response.data.data as SystemStats
    },
    refetchInterval: 60000, // 60秒刷新一次
  })

  const handleRefresh = () => {
    refetchHealth()
    refetchStats()
  }

  const isHealthy = health?.status === 'healthy'

  return (
    <div className="flex flex-col gap-4">
      {/* 标题栏 */}
      <div className="glass-card p-4">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-semibold m-0">系统监控</h2>
            <p className="text-gray-500 text-sm m-0 mt-1">实时监控系统运行状态和资源使用情况</p>
          </div>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>刷新</Button>
        </div>
      </div>

      {/* 系统状态 */}
      <div className="glass-card p-4">
        <h3 className="text-base font-medium mb-4">系统状态</h3>
        {healthLoading ? (
          <div className="text-center py-8"><Spin /></div>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="系统状态"
                  value={isHealthy ? '正常' : '异常'}
                  prefix={isHealthy ? <CheckCircleOutlined style={{ color: '#52c41a' }} /> : <CloseCircleOutlined style={{ color: '#ff4d4f' }} />}
                  valueStyle={{ color: isHealthy ? '#52c41a' : '#ff4d4f' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic title="运行时间" value={health?.uptime || '-'} />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic title="Go版本" value={health?.system.go_version || '-'} />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic title="协程数" value={health?.system.num_goroutine || 0} />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic title="CPU核心" value={health?.system.num_cpu || 0} />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic title="内存使用" value={health?.system.mem_alloc || '-'} />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic title="系统内存" value={health?.system.mem_sys || '-'} />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <div className="mb-2 text-gray-500 text-sm">服务状态</div>
                {health?.services && Object.entries(health.services).map(([name, status]) => (
                  <Tag key={name} color={status === 'healthy' ? 'green' : 'red'} className="mb-1">
                    {name}: {status === 'healthy' ? '正常' : '异常'}
                  </Tag>
                ))}
              </Card>
            </Col>
          </Row>
        )}
      </div>

      {/* 业务统计 */}
      <div className="glass-card p-4">
        <h3 className="text-base font-medium mb-4">业务统计</h3>
        {statsLoading ? (
          <div className="text-center py-8"><Spin /></div>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="总用户数"
                  value={stats?.total_users || 0}
                  prefix={<UserOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="会员用户"
                  value={stats?.total_members || 0}
                  prefix={<CrownOutlined style={{ color: '#faad14' }} />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="总卡密数"
                  value={stats?.total_cards || 0}
                  prefix={<CreditCardOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="已使用卡密"
                  value={stats?.used_cards || 0}
                  suffix={`/ ${stats?.total_cards || 0}`}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="今日登录"
                  value={stats?.today_logins || 0}
                  prefix={<LoginOutlined style={{ color: '#1890ff' }} />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="今日注册"
                  value={stats?.today_registers || 0}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="发布公告"
                  value={stats?.announcements || 0}
                  prefix={<NotificationOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card size="small">
                <Statistic
                  title="封禁IP数"
                  value={stats?.blocked_ips || 0}
                  prefix={<StopOutlined style={{ color: '#ff4d4f' }} />}
                />
              </Card>
            </Col>
          </Row>
        )}
      </div>
    </div>
  )
}

export default SystemMonitor

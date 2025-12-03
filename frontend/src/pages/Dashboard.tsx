import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Table, message, Progress } from 'antd';
import { UserOutlined, CheckCircleOutlined, EyeOutlined, CrownOutlined, KeyOutlined, WarningOutlined, RiseOutlined } from '@ant-design/icons';
import SafeECharts from '@/components/SafeECharts';
import { getStatistics } from '@/api/statistics';
import type { Statistics } from '@/types';

// 统计卡片组件
const StatCard: React.FC<{
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
  bgColor: string;
}> = ({ title, value, icon, color, bgColor }) => (
  <div style={{
    background: 'rgba(255, 255, 255, 0.5)',
    backdropFilter: 'blur(20px) saturate(180%)',
    WebkitBackdropFilter: 'blur(20px) saturate(180%)',
    borderRadius: 16,
    padding: '24px',
    display: 'flex',
    alignItems: 'center',
    gap: 16,
    boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
    border: '1px solid rgba(255, 255, 255, 0.4)',
    transition: 'all 0.3s ease',
    cursor: 'pointer',
  }}
  onMouseEnter={(e) => {
    e.currentTarget.style.transform = 'translateY(-4px)';
    e.currentTarget.style.boxShadow = '0 8px 24px rgba(0,0,0,0.12)';
  }}
  onMouseLeave={(e) => {
    e.currentTarget.style.transform = 'translateY(0)';
    e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.06)';
  }}
  >
    <div style={{
      width: 56,
      height: 56,
      borderRadius: 14,
      background: bgColor,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: 24,
      color: color,
    }}>
      {icon}
    </div>
    <div>
      <div style={{ color: '#86868b', fontSize: 13, marginBottom: 4 }}>{title}</div>
      <div style={{ fontSize: 28, fontWeight: 700, color: '#1d1d1f' }}>{value.toLocaleString()}</div>
    </div>
  </div>
);

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [statistics, setStatistics] = useState<Statistics | null>(null);

  // 加载统计数据
  const loadStatistics = async () => {
    setLoading(true);
    try {
      const response = await getStatistics();
      if (response.code === 200 && response.data) {
        setStatistics(response.data);
      }
    } catch (error) {
      message.error('加载统计数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStatistics();
  }, []);

  // 访问趋势图表配置
  const getTrendChartOption = () => {
    const trend = statistics?.access_trend || [];
    return {
      tooltip: { 
        trigger: 'axis',
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#eee',
        borderWidth: 1,
        textStyle: { color: '#333' },
        boxShadow: '0 4px 12px rgba(0,0,0,0.1)'
      },
      grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
      xAxis: {
        type: 'category',
        data: trend.map(item => {
          const d = new Date(item.date);
          return `${d.getMonth() + 1}/${d.getDate()}`;
        }),
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { color: '#86868b', fontSize: 12 }
      },
      yAxis: {
        type: 'value',
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { lineStyle: { color: '#f5f5f7', type: 'dashed' } },
        axisLabel: { color: '#86868b', fontSize: 12 }
      },
      series: [{
        data: trend.map(item => item.count),
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        areaStyle: { 
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(0, 122, 255, 0.3)' },
              { offset: 1, color: 'rgba(0, 122, 255, 0.02)' }
            ]
          }
        },
        lineStyle: { color: '#007aff', width: 3 },
        itemStyle: { color: '#007aff', borderWidth: 2, borderColor: '#fff' }
      }]
    };
  };

  // 卡密使用情况饼图
  const getCardKeyChartOption = () => {
    const stats = statistics?.cardkey_stats;
    if (!stats) return {};
    return {
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      legend: { show: false },
      series: [{
        type: 'pie',
        radius: ['50%', '75%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 3 },
        label: { show: false },
        data: [
          { value: stats.unused_cards, name: '未使用', itemStyle: { color: '#34c759' } },
          { value: stats.used_cards, name: '已使用', itemStyle: { color: '#007aff' } },
          { value: stats.disabled_cards, name: '已禁用', itemStyle: { color: '#ff3b30' } },
        ]
      }]
    };
  };

  // Top用户表格列
  const topUsersColumns = [
    { 
      title: '排名', 
      key: 'rank', 
      width: 60, 
      render: (_: any, __: any, index: number) => (
        <div style={{
          width: 28,
          height: 28,
          borderRadius: 8,
          background: index === 0 ? 'linear-gradient(135deg, #ffd700, #ffb300)' : 
                     index === 1 ? 'linear-gradient(135deg, #c0c0c0, #a0a0a0)' : 
                     index === 2 ? 'linear-gradient(135deg, #cd7f32, #b8860b)' : '#f5f5f7',
          color: index < 3 ? '#fff' : '#86868b',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontWeight: 600,
          fontSize: 13,
        }}>
          {index + 1}
        </div>
      )
    },
    { 
      title: '用户名', 
      dataIndex: 'username', 
      key: 'username',
      render: (v: string) => <span style={{ fontWeight: 500 }}>{v}</span>
    },
    { 
      title: '访问次数', 
      dataIndex: 'access_count', 
      key: 'access_count', 
      render: (v: number) => <span style={{ color: '#007aff', fontWeight: 600 }}>{v}</span>
    },
  ];

  const vipStats = statistics?.vip_stats;
  const cardStats = statistics?.cardkey_stats;

  return (
    <div style={{ padding: '0 8px' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ 
          fontSize: 28, 
          fontWeight: 700, 
          color: '#1d1d1f', 
          margin: 0,
          letterSpacing: '-0.5px'
        }}>
          系统概览
        </h1>
        <p style={{ color: '#86868b', marginTop: 4, fontSize: 14 }}>
          实时数据统计与分析
        </p>
      </div>

      {/* 基础统计卡片 */}
      <Row gutter={[20, 20]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={12} md={6}>
          <StatCard
            title="总用户数"
            value={statistics?.total_users || 0}
            icon={<UserOutlined />}
            color="#007aff"
            bgColor="rgba(0, 122, 255, 0.1)"
          />
        </Col>
        <Col xs={12} sm={12} md={6}>
          <StatCard
            title="活跃用户"
            value={statistics?.active_users || 0}
            icon={<CheckCircleOutlined />}
            color="#34c759"
            bgColor="rgba(52, 199, 89, 0.1)"
          />
        </Col>
        <Col xs={12} sm={12} md={6}>
          <StatCard
            title="今日访问"
            value={statistics?.today_access || 0}
            icon={<EyeOutlined />}
            color="#ff9500"
            bgColor="rgba(255, 149, 0, 0.1)"
          />
        </Col>
        <Col xs={12} sm={12} md={6}>
          <StatCard
            title="VIP用户"
            value={vipStats?.total_vip || 0}
            icon={<CrownOutlined />}
            color="#af52de"
            bgColor="rgba(175, 82, 222, 0.1)"
          />
        </Col>
      </Row>

      {/* VIP和卡密统计 */}
      <Row gutter={[20, 20]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <CrownOutlined style={{ color: '#af52de', fontSize: 18 }} />
                <span style={{ fontWeight: 600 }}>VIP会员统计</span>
              </div>
            }
            variant="borderless"
            style={{ borderRadius: 16, boxShadow: '0 2px 8px rgba(0,0,0,0.06)', height: '100%' }}
            styles={{ body: { height: 'calc(100% - 57px)' } }}
          >
            <Row gutter={24}>
              <Col span={12}>
                <div style={{ 
                  background: 'linear-gradient(135deg, rgba(52, 199, 89, 0.1), rgba(52, 199, 89, 0.05))',
                  borderRadius: 12,
                  padding: 16,
                  marginBottom: 12
                }}>
                  <div style={{ color: '#86868b', fontSize: 13, marginBottom: 4 }}>有效VIP</div>
                  <div style={{ fontSize: 32, fontWeight: 700, color: '#34c759' }}>{vipStats?.total_vip || 0}</div>
                </div>
                <div style={{ 
                  background: '#f5f5f7',
                  borderRadius: 12,
                  padding: 16
                }}>
                  <div style={{ color: '#86868b', fontSize: 13, marginBottom: 4 }}>已过期</div>
                  <div style={{ fontSize: 32, fontWeight: 700, color: '#86868b' }}>{vipStats?.expired_vip || 0}</div>
                </div>
              </Col>
              <Col span={12}>
                <div style={{ 
                  background: '#fffbeb',
                  borderRadius: 12,
                  padding: 16,
                  marginBottom: 12
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                    <WarningOutlined style={{ color: '#ff9500' }} />
                    <span style={{ color: '#86868b', fontSize: 13 }}>3天内到期</span>
                  </div>
                  <div style={{ fontSize: 24, fontWeight: 700, color: '#ff9500' }}>{vipStats?.expiring_3_day || 0}</div>
                  <Progress 
                    percent={vipStats?.total_vip ? Math.round((vipStats.expiring_3_day || 0) / vipStats.total_vip * 100) : 0} 
                    size="small" 
                    showInfo={false} 
                    strokeColor="#ff9500"
                    trailColor="rgba(255, 149, 0, 0.2)"
                    style={{ marginTop: 8 }}
                  />
                </div>
                <div style={{ 
                  background: 'rgba(0, 122, 255, 0.05)',
                  borderRadius: 12,
                  padding: 16
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                    <RiseOutlined style={{ color: '#007aff' }} />
                    <span style={{ color: '#86868b', fontSize: 13 }}>7天内到期</span>
                  </div>
                  <div style={{ fontSize: 24, fontWeight: 700, color: '#007aff' }}>{vipStats?.expiring_7_day || 0}</div>
                  <Progress 
                    percent={vipStats?.total_vip ? Math.round((vipStats.expiring_7_day || 0) / vipStats.total_vip * 100) : 0} 
                    size="small" 
                    showInfo={false}
                    strokeColor="#007aff"
                    trailColor="rgba(0, 122, 255, 0.2)"
                    style={{ marginTop: 8 }}
                  />
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <KeyOutlined style={{ color: '#007aff', fontSize: 18 }} />
                <span style={{ fontWeight: 600 }}>卡密统计</span>
              </div>
            }
            variant="borderless"
            style={{ borderRadius: 16, boxShadow: '0 2px 8px rgba(0,0,0,0.06)', height: '100%' }}
            styles={{ body: { height: 'calc(100% - 57px)', display: 'flex', alignItems: 'center' } }}
          >
            <Row gutter={20} align="middle" style={{ width: '100%' }}>
              <Col span={9}>
                <SafeECharts option={getCardKeyChartOption()} style={{ height: 180 }} />
              </Col>
              <Col span={15}>
                {[
                  { label: '未使用', value: cardStats?.unused_cards || 0, color: '#34c759' },
                  { label: '已使用', value: cardStats?.used_cards || 0, color: '#007aff' },
                  { label: '已禁用', value: cardStats?.disabled_cards || 0, color: '#ff3b30' },
                ].map((item, i) => (
                  <div key={i} style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'space-between',
                    padding: '12px 16px',
                    background: i % 2 === 0 ? '#f5f5f7' : '#fff',
                    borderRadius: 10,
                    marginBottom: i < 2 ? 8 : 0
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <div style={{ 
                        width: 12, 
                        height: 12, 
                        borderRadius: 4, 
                        background: item.color 
                      }} />
                      <span style={{ color: '#1d1d1f' }}>{item.label}</span>
                    </div>
                    <span style={{ fontWeight: 700, color: item.color, fontSize: 18 }}>{item.value}</span>
                  </div>
                ))}
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Row gutter={[20, 20]}>
        <Col xs={24} lg={12}>
          <Card 
            title={<span style={{ fontWeight: 600 }}>📈 近7日访问趋势</span>}
            loading={loading} 
            variant="borderless"
            style={{ borderRadius: 16, boxShadow: '0 2px 8px rgba(0,0,0,0.06)', height: '100%' }}
          >
            <SafeECharts option={getTrendChartOption()} style={{ height: 260 }} />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title={<span style={{ fontWeight: 600 }}>🏆 今日访问排行</span>}
            loading={loading} 
            variant="borderless"
            style={{ borderRadius: 16, boxShadow: '0 2px 8px rgba(0,0,0,0.06)', height: '100%' }}
          >
            <Table
              columns={topUsersColumns}
              dataSource={statistics?.top_users || []}
              rowKey="user_id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;

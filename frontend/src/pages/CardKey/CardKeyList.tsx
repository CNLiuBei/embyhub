import React, { useState, useEffect } from 'react';
import { Table, Button, Space, Modal, Form, InputNumber, Select, Input, message, Tag, Popconfirm, Row, Col, Tooltip, Card } from 'antd';
import { PlusOutlined, ReloadOutlined, CopyOutlined, StopOutlined, CheckCircleOutlined, DeleteOutlined, DownloadOutlined, KeyOutlined, GiftOutlined, CloseCircleOutlined, CrownOutlined } from '@ant-design/icons';
import { getCardKeys, createCardKeys, disableCardKey, enableCardKey, deleteCardKey, getCardKeyStatistics, CardKey, CardKeyCreateRequest } from '@/api/cardKey';

const CardKeyList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [cardKeys, setCardKeys] = useState<CardKey[]>([]);
  const [total, setTotal] = useState(0);
  const [statistics, setStatistics] = useState<any>({});
  const [pagination, setPagination] = useState({ page: 1, page_size: 10 });
  const [filters, setFilters] = useState<{ status?: number; card_type?: number; keyword?: string }>({});
  const [modalVisible, setModalVisible] = useState(false);
  const [generatedKeys, setGeneratedKeys] = useState<CardKey[]>([]);
  const [resultModalVisible, setResultModalVisible] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [form] = Form.useForm();
  
  // 监听表单值变化以更新按钮状态
  const countValue = Form.useWatch('count', form);
  const durationValue = Form.useWatch('duration', form);

  // 加载卡密列表
  const loadCardKeys = async () => {
    setLoading(true);
    try {
      const response: any = await getCardKeys({ ...pagination, ...filters });
      if (response.code === 200 && response.data) {
        setCardKeys(response.data.list || []);
        setTotal(response.data.total || 0);
      }
    } catch (error) {
      message.error('加载卡密列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载统计数据
  const loadStatistics = async () => {
    try {
      const response: any = await getCardKeyStatistics();
      if (response.code === 200 && response.data) {
        setStatistics(response.data);
      }
    } catch (error) {
      console.error('加载统计失败');
    }
  };

  useEffect(() => {
    loadCardKeys();
    loadStatistics();
  }, [pagination.page, pagination.page_size, filters]);

  // 生成卡密
  const handleCreate = async (values: CardKeyCreateRequest) => {
    try {
      // 确保数值类型正确
      const payload = {
        ...values,
        count: Number(values.count),
        card_type: Number(values.card_type),
        duration: Number(values.duration),
      };
      const response: any = await createCardKeys(payload);
      if (response.code === 200 && response.data) {
        message.success(`成功生成 ${response.data.length} 个卡密`);
        setGeneratedKeys(response.data);
        setModalVisible(false);
        setResultModalVisible(true);
        loadCardKeys();
        loadStatistics();
      }
    } catch (error) {
      message.error('生成卡密失败');
    }
  };

  // 禁用卡密
  const handleDisable = async (id: number) => {
    try {
      await disableCardKey(id);
      message.success('禁用成功');
      loadCardKeys();
      loadStatistics();
    } catch (error) {
      message.error('禁用失败');
    }
  };

  // 启用卡密
  const handleEnable = async (id: number) => {
    try {
      await enableCardKey(id);
      message.success('启用成功');
      loadCardKeys();
      loadStatistics();
    } catch (error) {
      message.error('启用失败');
    }
  };

  // 删除卡密
  const handleDelete = async (id: number) => {
    try {
      await deleteCardKey(id);
      message.success('删除成功');
      loadCardKeys();
      loadStatistics();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 复制卡密
  const copyCardCode = (code: string) => {
    navigator.clipboard.writeText(code);
    message.success('已复制到剪贴板');
  };

  // 批量复制生成的卡密
  const copyAllCardCodes = () => {
    const codes = generatedKeys.map(k => k.card_code).join('\n');
    navigator.clipboard.writeText(codes);
    message.success('已复制所有卡密到剪贴板');
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的卡密');
      return;
    }
    let deleted = 0;
    for (const id of selectedRowKeys) {
      try {
        await deleteCardKey(id as number);
        deleted++;
      } catch (error) {
        // 忽略已使用的卡密删除失败
      }
    }
    message.success(`成功删除 ${deleted} 个卡密`);
    setSelectedRowKeys([]);
    loadCardKeys();
    loadStatistics();
  };

  // 导出卡密
  const handleExport = () => {
    const exportData = cardKeys.filter(k => 
      selectedRowKeys.length === 0 || selectedRowKeys.includes(k.id)
    );
    if (exportData.length === 0) {
      message.warning('没有可导出的卡密');
      return;
    }
    const content = exportData.map(k => 
      `${k.card_code}\tVIP会员码\t${k.duration}天\t${k.status === 1 ? '未使用' : k.status === 2 ? '已使用' : '已禁用'}`
    ).join('\n');
    const header = '卡密码\t类型\t有效期\t状态\n';
    const blob = new Blob([header + content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `卡密导出_${new Date().toLocaleDateString()}.txt`;
    a.click();
    URL.revokeObjectURL(url);
    message.success('导出成功');
  };

  // 搜索
  const handleSearch = () => {
    setFilters({ ...filters, keyword: searchKeyword });
  };

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
    getCheckboxProps: (record: CardKey) => ({
      disabled: record.status === 2, // 已使用的不可选
    }),
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '卡密码',
      dataIndex: 'card_code',
      key: 'card_code',
      render: (code: string) => (
        <Space>
          <code style={{ fontFamily: 'monospace', fontSize: 13 }}>{code}</code>
          <Tooltip title="复制">
            <Button type="text" size="small" icon={<CopyOutlined />} onClick={() => copyCardCode(code)} />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'card_type',
      key: 'card_type',
      width: 100,
      render: () => (
        <Tag color="purple" icon={<CrownOutlined />}>VIP会员码</Tag>
      ),
    },
    {
      title: '有效期',
      dataIndex: 'duration',
      key: 'duration',
      width: 80,
      render: (days: number) => `${days}天`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: number) => {
        const statusMap: { [key: number]: { color: string; text: string } } = {
          0: { color: 'default', text: '已禁用' },
          1: { color: 'green', text: '未使用' },
          2: { color: 'orange', text: '已使用' },
        };
        const s = statusMap[status] || { color: 'default', text: '未知' };
        return <Tag color={s.color}>{s.text}</Tag>;
      },
    },
    {
      title: '使用者',
      dataIndex: 'used_by_user',
      key: 'used_by_user',
      width: 100,
      render: (user: any) => user?.username || '-',
    },
    {
      title: '使用时间',
      dataIndex: 'used_at',
      key: 'used_at',
      width: 160,
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
    {
      title: '备注',
      dataIndex: 'remark',
      key: 'remark',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: CardKey) => (
        <Space size="small">
          {record.status === 1 && (
            <Button 
              type="link" 
              size="small" 
              icon={<StopOutlined />}
              onClick={() => handleDisable(record.id)}
            >
              禁用
            </Button>
          )}
          {record.status === 0 && (
            <Button 
              type="link" 
              size="small" 
              icon={<CheckCircleOutlined />}
              onClick={() => handleEnable(record.id)}
            >
              启用
            </Button>
          )}
          {record.status !== 2 && (
            <Popconfirm
              title="确定删除此卡密吗？"
              onConfirm={() => handleDelete(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  // 统计卡片组件
  const StatCard = ({ icon, title, value, color, bgColor }: { icon: React.ReactNode; title: string; value: number; color: string; bgColor: string }) => (
    <div style={{
      background: 'rgba(255, 255, 255, 0.5)',
      backdropFilter: 'blur(20px) saturate(180%)',
      WebkitBackdropFilter: 'blur(20px) saturate(180%)',
      borderRadius: 16,
      padding: 20,
      display: 'flex',
      alignItems: 'center',
      gap: 16,
      boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
      border: '1px solid rgba(255, 255, 255, 0.4)',
      transition: 'all 0.3s',
      cursor: 'default',
    }}>
      <div style={{
        width: 52,
        height: 52,
        borderRadius: 14,
        background: bgColor,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: 22,
        color: color,
      }}>
        {icon}
      </div>
      <div>
        <div style={{ fontSize: 28, fontWeight: 700, color: '#1d1d1f', lineHeight: 1.2 }}>{value}</div>
        <div style={{ color: '#86868b', fontSize: 13, marginTop: 2 }}>{title}</div>
      </div>
    </div>
  );

  return (
    <div style={{ padding: '0 4px' }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ fontSize: 28, fontWeight: 700, color: '#1d1d1f', margin: 0, letterSpacing: '-0.5px' }}>
          卡密管理
        </h1>
        <p style={{ color: '#86868b', marginTop: 4, fontSize: 14, margin: '4px 0 0' }}>
          管理VIP卡密的生成、使用和状态
        </p>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <StatCard
            icon={<KeyOutlined />}
            title="卡密总数"
            value={statistics.total || 0}
            color="#007aff"
            bgColor="rgba(0, 122, 255, 0.1)"
          />
        </Col>
        <Col xs={12} sm={6}>
          <StatCard
            icon={<GiftOutlined />}
            title="未使用"
            value={statistics.unused || 0}
            color="#34c759"
            bgColor="rgba(52, 199, 89, 0.1)"
          />
        </Col>
        <Col xs={12} sm={6}>
          <StatCard
            icon={<CheckCircleOutlined />}
            title="已使用"
            value={statistics.used || 0}
            color="#ff9500"
            bgColor="rgba(255, 149, 0, 0.1)"
          />
        </Col>
        <Col xs={12} sm={6}>
          <StatCard
            icon={<CloseCircleOutlined />}
            title="已禁用"
            value={statistics.disabled || 0}
            color="#8e8e93"
            bgColor="rgba(142, 142, 147, 0.1)"
          />
        </Col>
      </Row>

      {/* 操作栏 */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 20,
        flexWrap: 'wrap',
        gap: 12,
        background: 'rgba(255, 255, 255, 0.5)',
        backdropFilter: 'blur(20px) saturate(180%)',
        padding: '16px 20px',
        borderRadius: 12,
        boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
        border: '1px solid rgba(255, 255, 255, 0.4)',
      }}>
        <Space wrap>
          <Space.Compact>
            <Input
              placeholder="搜索卡密码/备注"
              style={{ width: 180 }}
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onPressEnter={handleSearch}
              allowClear
            />
            <Button type="primary" onClick={handleSearch}>搜索</Button>
          </Space.Compact>
          <Select
            placeholder="状态筛选"
            allowClear
            style={{ width: 110 }}
            onChange={(value) => setFilters({ ...filters, status: value })}
          >
            <Select.Option value={1}>未使用</Select.Option>
            <Select.Option value={2}>已使用</Select.Option>
            <Select.Option value={0}>已禁用</Select.Option>
          </Select>
        </Space>
        <Space wrap>
          {selectedRowKeys.length > 0 && (
            <Popconfirm
              title={`确定删除选中的 ${selectedRowKeys.length} 个卡密吗？`}
              onConfirm={handleBatchDelete}
              okText="确定"
              cancelText="取消"
            >
              <Button danger icon={<DeleteOutlined />}>
                批量删除 ({selectedRowKeys.length})
              </Button>
            </Popconfirm>
          )}
          <Button icon={<DownloadOutlined />} onClick={handleExport}>
            导出
          </Button>
          <Button icon={<ReloadOutlined />} onClick={loadCardKeys}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
            生成卡密
          </Button>
        </Space>
      </div>

      {/* 卡密列表 */}
      <Card 
        styles={{ body: { padding: 0 } }}
        style={{ borderRadius: 12, boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}
      >
        <Table
          rowSelection={rowSelection}
          columns={columns}
          dataSource={cardKeys}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.page,
            pageSize: pagination.page_size,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => {
              setPagination({ page, page_size: pageSize });
            },
          }}
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* 生成卡密弹窗 */}
      <Modal
        title="生成卡密"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText="生成"
        cancelText="取消"
      >
        <Form
          form={form}
          onFinish={handleCreate}
          layout="vertical"
          initialValues={{ count: 10, card_type: 1, duration: 30 }}
        >
          <Form.Item
            name="count"
            label="生成数量"
            rules={[{ required: true, message: '请输入数量' }]}
            extra="单次最多生成100个"
          >
            <Space.Compact>
              <InputNumber min={1} max={100} style={{ width: 100 }} />
              <Button 
                type={countValue === 10 ? 'primary' : 'default'}
                onClick={() => form.setFieldValue('count', 10)}
              >10个</Button>
              <Button 
                type={countValue === 50 ? 'primary' : 'default'}
                onClick={() => form.setFieldValue('count', 50)}
              >50个</Button>
              <Button 
                type={countValue === 100 ? 'primary' : 'default'}
                onClick={() => form.setFieldValue('count', 100)}
              >100个</Button>
            </Space.Compact>
          </Form.Item>
          <Form.Item
            name="card_type"
            label="卡密类型"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select>
              <Select.Option value={1}>
                <Space><CrownOutlined style={{ color: '#af52de' }} />VIP会员码 - 用于开通/续费会员</Space>
              </Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="duration"
            label="会员有效期（天）"
            rules={[{ required: true, message: '请输入有效期' }]}
            extra="用户使用卡密后的会员有效天数"
          >
            <Space.Compact>
              <InputNumber min={1} max={365} style={{ width: 100 }} />
              <Button 
                type={durationValue === 30 ? 'primary' : 'default'}
                onClick={() => form.setFieldValue('duration', 30)}
              >30天</Button>
              <Button 
                type={durationValue === 90 ? 'primary' : 'default'}
                onClick={() => form.setFieldValue('duration', 90)}
              >90天</Button>
              <Button 
                type={durationValue === 365 ? 'primary' : 'default'}
                onClick={() => form.setFieldValue('duration', 365)}
              >365天</Button>
            </Space.Compact>
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} placeholder="可选备注信息，如：活动卡密、测试卡密等" />
          </Form.Item>
        </Form>
        <div style={{ marginTop: 8, color: '#666', fontSize: 12 }}>
          💡 卡密格式：TL|XXXXXXXXXXXXXXXXXXXXXXXX（26位）
        </div>
      </Modal>

      {/* 生成结果弹窗 */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <CheckCircleOutlined style={{ color: '#34c759' }} />
            <span>生成成功</span>
            <Tag color="blue">{generatedKeys.length} 个</Tag>
          </div>
        }
        open={resultModalVisible}
        onCancel={() => setResultModalVisible(false)}
        footer={[
          <Button key="copy" type="primary" icon={<CopyOutlined />} onClick={copyAllCardCodes}>
            复制全部卡密
          </Button>,
          <Button key="close" onClick={() => setResultModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={520}
      >
        <div style={{ 
          maxHeight: 400, 
          overflow: 'auto', 
          background: '#f9f9f9', 
          borderRadius: 8, 
          padding: 12 
        }}>
          {generatedKeys.map((key, index) => (
            <div 
              key={key.id} 
              style={{ 
                padding: '10px 12px', 
                background: 'rgba(255, 255, 255, 0.6)', 
                borderRadius: 8, 
                marginBottom: 8,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
              }}
            >
              <Space>
                <span style={{ color: '#86868b', fontSize: 12, width: 24 }}>{index + 1}.</span>
                <code style={{ 
                  fontFamily: 'Monaco, monospace', 
                  fontSize: 13, 
                  color: '#1d1d1f',
                  background: '#f5f5f7',
                  padding: '4px 8px',
                  borderRadius: 4,
                }}>
                  {key.card_code}
                </code>
              </Space>
              <Tooltip title="复制">
                <Button 
                  type="text" 
                  size="small" 
                  icon={<CopyOutlined />} 
                  onClick={() => copyCardCode(key.card_code)}
                  style={{ color: '#007aff' }}
                />
              </Tooltip>
            </div>
          ))}
        </div>
      </Modal>
    </div>
  );
};

export default CardKeyList;

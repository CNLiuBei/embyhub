import React, { useState, useEffect } from 'react';
import {
  Card, Table, Button, Space, Modal, Form, Input, Select, Tag, DatePicker, Descriptions, Typography, App
} from 'antd';
import { SearchOutlined, ReloadOutlined, EyeOutlined } from '@ant-design/icons';
import { goofishApi, GoofishOrder, GoofishOrderCard } from '../../../services/goofishApi';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { Text, Paragraph } = Typography;

const statusMap: Record<number, { label: string; color: string }> = {
  10: { label: '处理中', color: 'processing' },
  20: { label: '已成功', color: 'success' },
  30: { label: '已失败', color: 'error' },
};

const GoofishOrdersPage: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [orders, setOrders] = useState<GoofishOrder[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<any>({});
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<GoofishOrder | null>(null);
  const [orderCards, setOrderCards] = useState<GoofishOrderCard[]>([]);
  const [form] = Form.useForm();

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const res = await goofishApi.getOrderList({
        page,
        page_size: pageSize,
        ...filters,
      });
      // 后端返回格式: {code: 200, data: {total, list}, message: "success"}
      const data = res.data.data;
      setOrders(data?.list || []);
      setTotal(data?.total || 0);
    } catch (error) {
      message.error('获取订单列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrders();
  }, [page, pageSize, filters]);

  const handleSearch = (values: any) => {
    const newFilters: any = {};
    if (values.order_no) newFilters.order_no = values.order_no;
    if (values.goods_no) newFilters.goods_no = values.goods_no;
    if (values.biz_order_no) newFilters.biz_order_no = values.biz_order_no;
    if (values.status !== undefined) newFilters.status = values.status;
    if (values.time_range && values.time_range.length === 2) {
      newFilters.start_time = values.time_range[0].format('YYYY-MM-DD 00:00:00');
      newFilters.end_time = values.time_range[1].format('YYYY-MM-DD 23:59:59');
    }
    setFilters(newFilters);
    setPage(1);
  };

  const handleReset = () => {
    form.resetFields();
    setFilters({});
    setPage(1);
  };

  const handleViewDetail = async (record: GoofishOrder) => {
    setSelectedOrder(record);
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await goofishApi.getOrderDetail(record.order_no);
      // 后端返回格式: {code: 200, data: {order, cards}, message: "success"}
      const data = res.data.data;
      setOrderCards(data?.cards || []);
    } catch (error) {
      message.error('获取订单详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  const columns = [
    {
      title: '闲管家订单号',
      dataIndex: 'order_no',
      key: 'order_no',
      width: 180,
      render: (text: string) => (
        <Paragraph copyable style={{ margin: 0 }}>{text}</Paragraph>
      ),
    },
    {
      title: '本系统订单号',
      dataIndex: 'out_order_no',
      key: 'out_order_no',
      width: 160,
    },
    {
      title: '业务订单号',
      dataIndex: 'biz_order_no',
      key: 'biz_order_no',
      width: 160,
      render: (text: string) => text || '-',
    },
    {
      title: '商品',
      key: 'goods',
      width: 150,
      render: (_: any, record: GoofishOrder) => (
        <div>
          <div>{record.goods_name}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.goods_no}</Text>
        </div>
      ),
    },
    {
      title: '数量',
      dataIndex: 'buy_quantity',
      key: 'buy_quantity',
      width: 60,
    },
    {
      title: '金额(元)',
      dataIndex: 'order_amount',
      key: 'order_amount',
      width: 100,
      render: (amount: number) => (amount / 100).toFixed(2),
    },
    {
      title: '状态',
      dataIndex: 'order_status',
      key: 'order_status',
      width: 80,
      render: (status: number) => {
        const info = statusMap[status];
        return info ? <Tag color={info.color}>{info.label}</Tag> : status;
      },
    },
    {
      title: '来源IP',
      dataIndex: 'client_ip',
      key: 'client_ip',
      width: 120,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: GoofishOrder) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)}>
          详情
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card title="订单管理">
        <Form form={form} layout="inline" onFinish={handleSearch} style={{ marginBottom: 16 }}>
          <Form.Item name="order_no">
            <Input placeholder="闲管家订单号" allowClear style={{ width: 180 }} />
          </Form.Item>
          <Form.Item name="goods_no">
            <Input placeholder="商品编码" allowClear style={{ width: 140 }} />
          </Form.Item>
          <Form.Item name="biz_order_no">
            <Input placeholder="业务订单号" allowClear style={{ width: 160 }} />
          </Form.Item>
          <Form.Item name="status">
            <Select placeholder="状态" allowClear style={{ width: 100 }}>
              <Option value={10}>处理中</Option>
              <Option value={20}>已成功</Option>
              <Option value={30}>已失败</Option>
            </Select>
          </Form.Item>
          <Form.Item name="time_range">
            <RangePicker />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
              <Button onClick={handleReset}>重置</Button>
              <Button icon={<ReloadOutlined />} onClick={fetchOrders}>刷新</Button>
            </Space>
          </Form.Item>
        </Form>

        <Table
          columns={columns}
          dataSource={orders}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1400 }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
        />
      </Card>

      <Modal
        title="订单详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={700}
      >
        {selectedOrder && (
          <div>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="闲管家订单号" span={2}>
                <Paragraph copyable style={{ margin: 0 }}>{selectedOrder.order_no}</Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label="本系统订单号">{selectedOrder.out_order_no}</Descriptions.Item>
              <Descriptions.Item label="业务订单号">{selectedOrder.biz_order_no || '-'}</Descriptions.Item>
              <Descriptions.Item label="商品名称">{selectedOrder.goods_name}</Descriptions.Item>
              <Descriptions.Item label="商品编码">{selectedOrder.goods_no}</Descriptions.Item>
              <Descriptions.Item label="购买数量">{selectedOrder.buy_quantity}</Descriptions.Item>
              <Descriptions.Item label="订单金额">{(selectedOrder.order_amount / 100).toFixed(2)} 元</Descriptions.Item>
              <Descriptions.Item label="订单状态">
                {(() => {
                  const info = statusMap[selectedOrder.order_status];
                  return info ? <Tag color={info.color}>{info.label}</Tag> : selectedOrder.order_status;
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="来源IP">{selectedOrder.client_ip}</Descriptions.Item>
              <Descriptions.Item label="创建时间" span={2}>
                {dayjs(selectedOrder.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              {selectedOrder.remark && (
                <Descriptions.Item label="备注" span={2}>{selectedOrder.remark}</Descriptions.Item>
              )}
            </Descriptions>

            <Card title="发货卡密" size="small" style={{ marginTop: 16 }} loading={detailLoading}>
              {orderCards.length > 0 ? (
                <Table
                  dataSource={orderCards}
                  rowKey="id"
                  size="small"
                  pagination={false}
                  columns={[
                    { title: '卡密ID', dataIndex: 'card_id', width: 80 },
                    {
                      title: '卡密码',
                      dataIndex: 'card_code',
                      render: (text: string) => (
                        <Text copyable={{ text: text }} style={{ wordBreak: 'break-all' }}>{text}</Text>
                      ),
                    },
                    {
                      title: '发货时间',
                      dataIndex: 'created_at',
                      width: 160,
                      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
                    },
                  ]}
                />
              ) : (
                <Text type="secondary">暂无发货记录</Text>
              )}
            </Card>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default GoofishOrdersPage;

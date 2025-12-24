import React, { useState, useEffect } from 'react';
import {
  Card, Table, Button, Space, Modal, Form, Input, InputNumber, Select, Switch, Tag, Popconfirm, Tooltip, App
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { goofishApi, GoofishGoods, CreateGoodsRequest, UpdateGoodsRequest } from '../../../services/goofishApi';

const { Option } = Select;

const cardTypeMap: Record<number, { label: string; color: string }> = {
  1: { label: '月卡', color: 'blue' },
  2: { label: '季卡', color: 'green' },
  3: { label: '半年卡', color: 'orange' },
  4: { label: '年卡', color: 'red' },
};

const statusMap: Record<number, { label: string; color: string }> = {
  1: { label: '在架', color: 'success' },
  2: { label: '下架', color: 'default' },
};

const GoofishGoodsPage: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [goods, setGoods] = useState<GoofishGoods[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingGoods, setEditingGoods] = useState<GoofishGoods | null>(null);
  const [autoGenerating, setAutoGenerating] = useState(false);
  const [form] = Form.useForm();

  const handleAutoGenerate = async () => {
    setAutoGenerating(true);
    try {
      const res = await goofishApi.autoGenerateGoods();
      const data = res.data.data;
      if (data.created > 0) {
        message.success(`成功创建 ${data.created} 个商品映射`);
        fetchGoods();
      } else if (data.skipped > 0) {
        message.info(`所有商品映射已存在，跳过 ${data.skipped} 个`);
      } else {
        message.info('没有需要创建的商品映射');
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '自动生成失败');
    } finally {
      setAutoGenerating(false);
    }
  };

  const fetchGoods = async () => {
    setLoading(true);
    try {
      const res = await goofishApi.getGoodsList({ page, page_size: pageSize });
      // 后端返回格式: {code: 200, data: {total, list}, message: "success"}
      const data = res.data.data;
      setGoods(data?.list || []);
      setTotal(data?.total || 0);
    } catch (error) {
      message.error('获取商品列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGoods();
  }, [page, pageSize]);

  const handleCreate = () => {
    setEditingGoods(null);
    form.resetFields();
    form.setFieldsValue({
      status: 1,
      auto_generate: false,
      card_prefix: 'GF',
      max_auto_generate: 10,
    });
    setModalVisible(true);
  };

  const handleEdit = (record: GoofishGoods) => {
    setEditingGoods(record);
    form.setFieldsValue({
      goods_no: record.goods_no,
      goods_name: record.goods_name,
      card_type: record.card_type,
      price: record.price / 100, // 转换为元
      status: record.status,
      auto_generate: record.auto_generate,
      card_prefix: record.card_prefix,
      duration: record.duration,
      max_auto_generate: record.max_auto_generate,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await goofishApi.deleteGoods(id);
      message.success('删除成功');
      fetchGoods();
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        price: Math.round(values.price * 100), // 转换为分
      };

      if (editingGoods) {
        await goofishApi.updateGoods(editingGoods.id, data as UpdateGoodsRequest);
        message.success('更新成功');
      } else {
        await goofishApi.createGoods(data as CreateGoodsRequest);
        message.success('创建成功');
      }
      setModalVisible(false);
      fetchGoods();
    } catch (error: any) {
      message.error(error.response?.data?.error || '操作失败');
    }
  };

  const columns = [
    {
      title: '商品编码',
      dataIndex: 'goods_no',
      key: 'goods_no',
      width: 120,
    },
    {
      title: '商品名称',
      dataIndex: 'goods_name',
      key: 'goods_name',
      width: 150,
    },
    {
      title: '卡密类型',
      dataIndex: 'card_type',
      key: 'card_type',
      width: 100,
      render: (type: number) => {
        const info = cardTypeMap[type];
        return info ? <Tag color={info.color}>{info.label}</Tag> : type;
      },
    },
    {
      title: '价格(元)',
      dataIndex: 'price',
      key: 'price',
      width: 100,
      render: (price: number) => (price / 100).toFixed(2),
    },
    {
      title: '库存',
      dataIndex: 'stock',
      key: 'stock',
      width: 80,
      render: (stock: number) => (
        <Tag color={stock > 10 ? 'green' : stock > 0 ? 'orange' : 'red'}>
          {stock}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        const info = statusMap[status];
        return info ? <Tag color={info.color}>{info.label}</Tag> : status;
      },
    },
    {
      title: '自动生成',
      dataIndex: 'auto_generate',
      key: 'auto_generate',
      width: 100,
      render: (auto: boolean, record: GoofishGoods) => (
        <Tooltip title={auto ? `前缀: ${record.card_prefix}, 最大: ${record.max_auto_generate}` : ''}>
          <Tag color={auto ? 'blue' : 'default'}>{auto ? '已启用' : '未启用'}</Tag>
        </Tooltip>
      ),
    },
    {
      title: '有效天数',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (d: number) => `${d}天`,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: GoofishGoods) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除此商品？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="商品映射管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchGoods}>刷新</Button>
            <Popconfirm
              title="自动生成商品映射"
              description="将根据系统卡密类型自动创建商品映射（月卡、季卡、半年卡、年卡）"
              onConfirm={handleAutoGenerate}
              okText="确定"
              cancelText="取消"
            >
              <Button icon={<ThunderboltOutlined />} loading={autoGenerating}>
                自动生成
              </Button>
            </Popconfirm>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              添加商品
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={goods}
          rowKey="id"
          loading={loading}
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
        title={editingGoods ? '编辑商品' : '添加商品'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="goods_no"
            label="商品编码"
            rules={[{ required: true, message: '请输入商品编码' }]}
            tooltip="唯一标识，用于闲管家下单时指定商品"
          >
            <Input placeholder="如: MONTH_CARD_001" disabled={!!editingGoods} />
          </Form.Item>

          <Form.Item
            name="goods_name"
            label="商品名称"
            rules={[{ required: true, message: '请输入商品名称' }]}
          >
            <Input placeholder="如: 月卡会员" />
          </Form.Item>

          <Form.Item
            name="card_type"
            label="卡密类型"
            rules={[{ required: true, message: '请选择卡密类型' }]}
            tooltip="对应本系统卡密的类型"
          >
            <Select placeholder="选择卡密类型">
              <Option value={1}>月卡 (30天)</Option>
              <Option value={2}>季卡 (90天)</Option>
              <Option value={3}>半年卡 (180天)</Option>
              <Option value={4}>年卡 (365天)</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="price"
            label="价格(元)"
            rules={[{ required: true, message: '请输入价格' }]}
            tooltip="单位为元，系统会自动转换为分"
          >
            <InputNumber min={0.01} step={0.01} precision={2} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item name="status" label="状态">
            <Select>
              <Option value={1}>在架</Option>
              <Option value={2}>下架</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="auto_generate"
            label="自动生成卡密"
            valuePropName="checked"
            tooltip="库存不足时自动生成新卡密"
          >
            <Switch />
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.auto_generate !== cur.auto_generate}>
            {({ getFieldValue }) =>
              getFieldValue('auto_generate') && (
                <>
                  <Form.Item name="card_prefix" label="卡密前缀">
                    <Input placeholder="如: GF" />
                  </Form.Item>
                  <Form.Item name="duration" label="有效天数">
                    <InputNumber min={1} style={{ width: '100%' }} />
                  </Form.Item>
                  <Form.Item name="max_auto_generate" label="单次最大生成数量">
                    <InputNumber min={1} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </>
              )
            }
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingGoods ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default GoofishGoodsPage;

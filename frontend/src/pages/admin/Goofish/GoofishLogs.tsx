import React, { useState, useEffect } from 'react';
import {
  Card, Table, Button, Space, Modal, Form, Input, Tag, DatePicker, Typography, Popconfirm, App
} from 'antd';
import { SearchOutlined, ReloadOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { goofishApi, GoofishLog } from '../../../services/goofishApi';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Text } = Typography;

const GoofishLogsPage: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<GoofishLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<any>({});
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedLog, setSelectedLog] = useState<GoofishLog | null>(null);
  const [form] = Form.useForm();

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const res = await goofishApi.getLogs({
        page,
        page_size: pageSize,
        ...filters,
      });
      // 后端返回格式: {code: 200, data: {total, list}, message: "success"}
      const data = res.data.data;
      setLogs(data?.list || []);
      setTotal(data?.total || 0);
    } catch (error) {
      message.error('获取日志列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [page, pageSize, filters]);

  const handleSearch = (values: any) => {
    const newFilters: any = {};
    if (values.endpoint) newFilters.endpoint = values.endpoint;
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

  const handleCleanOldLogs = async () => {
    try {
      const res = await goofishApi.cleanOldLogs();
      // 后端返回格式: {code: 200, data: {count}, message: "清理成功"}
      const data = res.data.data;
      message.success(`清理成功，共删除 ${data?.count || 0} 条日志`);
      fetchLogs();
    } catch (error) {
      message.error('清理失败');
    }
  };

  const handleViewDetail = (record: GoofishLog) => {
    setSelectedLog(record);
    setDetailVisible(true);
  };

  const formatJson = (str: string) => {
    try {
      return JSON.stringify(JSON.parse(str), null, 2);
    } catch {
      return str;
    }
  };

  const getResponseCodeColor = (code: number) => {
    if (code === 0) return 'success';
    if (code >= 400 && code < 500) return 'warning';
    if (code >= 500) return 'error';
    return 'default';
  };

  const columns = [
    {
      title: '接口路径',
      dataIndex: 'endpoint',
      key: 'endpoint',
      width: 200,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 70,
      render: (method: string) => <Tag>{method}</Tag>,
    },
    {
      title: '响应码',
      dataIndex: 'response_code',
      key: 'response_code',
      width: 80,
      render: (code: number) => (
        <Tag color={getResponseCodeColor(code)}>{code}</Tag>
      ),
    },
    {
      title: '耗时(ms)',
      dataIndex: 'duration',
      key: 'duration',
      width: 90,
      render: (d: number) => (
        <Tag color={d > 1000 ? 'red' : d > 500 ? 'orange' : 'green'}>{d}</Tag>
      ),
    },
    {
      title: '客户端IP',
      dataIndex: 'client_ip',
      key: 'client_ip',
      width: 120,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: GoofishLog) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)}>
          详情
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="API日志"
        extra={
          <Popconfirm title="确定清理30天前的日志？" onConfirm={handleCleanOldLogs}>
            <Button danger icon={<DeleteOutlined />}>清理旧日志</Button>
          </Popconfirm>
        }
      >
        <Form form={form} layout="inline" onFinish={handleSearch} style={{ marginBottom: 16 }}>
          <Form.Item name="endpoint">
            <Input placeholder="接口路径" allowClear style={{ width: 200 }} />
          </Form.Item>
          <Form.Item name="time_range">
            <RangePicker />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
              <Button onClick={handleReset}>重置</Button>
              <Button icon={<ReloadOutlined />} onClick={fetchLogs}>刷新</Button>
            </Space>
          </Form.Item>
        </Form>

        <Table
          columns={columns}
          dataSource={logs}
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
        title="日志详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={800}
      >
        {selectedLog && (
          <div>
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div>
                <Text strong>接口路径：</Text>
                <Text>{selectedLog.endpoint}</Text>
              </div>
              <div>
                <Text strong>请求方法：</Text>
                <Tag>{selectedLog.method}</Tag>
              </div>
              <div>
                <Text strong>响应码：</Text>
                <Tag color={getResponseCodeColor(selectedLog.response_code)}>
                  {selectedLog.response_code}
                </Tag>
              </div>
              <div>
                <Text strong>耗时：</Text>
                <Text>{selectedLog.duration} ms</Text>
              </div>
              <div>
                <Text strong>客户端IP：</Text>
                <Text>{selectedLog.client_ip}</Text>
              </div>
              <div>
                <Text strong>时间：</Text>
                <Text>{dayjs(selectedLog.created_at).format('YYYY-MM-DD HH:mm:ss')}</Text>
              </div>
              <div>
                <Text strong>请求内容：</Text>
                <pre style={{
                  background: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                  maxHeight: 200,
                  overflow: 'auto',
                  fontSize: 12,
                }}>
                  {formatJson(selectedLog.request_body)}
                </pre>
              </div>
              <div>
                <Text strong>响应内容：</Text>
                <pre style={{
                  background: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                  maxHeight: 200,
                  overflow: 'auto',
                  fontSize: 12,
                }}>
                  {formatJson(selectedLog.response_body)}
                </pre>
              </div>
            </Space>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default GoofishLogsPage;

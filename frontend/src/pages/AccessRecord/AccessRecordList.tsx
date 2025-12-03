import React, { useState, useEffect } from 'react';
import { Table, DatePicker, Space, Button, message } from 'antd';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { getAccessRecords } from '@/api/accessRecord';
import type { AccessRecord, AccessRecordQuery } from '@/types';

const { RangePicker } = DatePicker;

const AccessRecordList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [records, setRecords] = useState<AccessRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState<AccessRecordQuery>({
    page: 1,
    page_size: 10,
  });

  // 加载访问记录
  const loadRecords = async () => {
    setLoading(true);
    try {
      const response = await getAccessRecords(query);
      if (response.code === 200 && response.data) {
        setRecords(response.data.list || []);
        setTotal(response.data.total || 0);
      }
    } catch (error) {
      message.error('加载访问记录失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRecords();
  }, [query.page, query.page_size]);

  // 搜索
  const handleSearch = () => {
    setQuery({ ...query, page: 1 });
    loadRecords();
  };

  // 重置
  const handleReset = () => {
    setQuery({ page: 1, page_size: 10 });
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'record_id',
      key: 'record_id',
      width: 80,
    },
    {
      title: '用户名',
      dataIndex: 'user',
      key: 'user',
      render: (user: any) => user?.username || '-',
    },
    {
      title: '访问资源',
      dataIndex: 'resource',
      key: 'resource',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: '设备信息',
      dataIndex: 'device_info',
      key: 'device_info',
      ellipsis: true,
    },
    {
      title: '访问时间',
      dataIndex: 'access_time',
      key: 'access_time',
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
  ];

  return (
    <div>
      {/* 页面头部 */}
      <div className="page-header">
        <h1>访问记录</h1>
        <p>查看用户访问历史和系统使用情况</p>
      </div>

      {/* 筛选区域 */}
      <div className="filter-section" style={{ marginBottom: 20 }}>
        <Space wrap>
          <RangePicker
            placeholder={['开始时间', '结束时间']}
            onChange={(dates) => {
              if (dates) {
                setQuery({
                  ...query,
                  start_time: dates[0]?.format('YYYY-MM-DD HH:mm:ss'),
                  end_time: dates[1]?.format('YYYY-MM-DD HH:mm:ss'),
                });
              } else {
                const { start_time, end_time, ...rest } = query;
                setQuery(rest);
              }
            }}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
            搜索
          </Button>
          <Button onClick={handleReset}>重置</Button>
          <Button icon={<ReloadOutlined />} onClick={loadRecords}>
            刷新
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={records}
        rowKey="record_id"
        loading={loading}
        pagination={{
          current: query.page,
          pageSize: query.page_size,
          total: total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, pageSize) => {
            setQuery({ ...query, page, page_size: pageSize });
          },
        }}
      />
    </div>
  );
};

export default AccessRecordList;

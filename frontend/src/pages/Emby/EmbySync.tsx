import React, { useState } from 'react';
import { Card, Button, Space, Table, message, Alert } from 'antd';
import { SyncOutlined, ApiOutlined, ReloadOutlined } from '@ant-design/icons';
import { testConnection, syncUsers, getEmbyUsers } from '@/api/emby';
import type { EmbyUser } from '@/types';

const EmbySync: React.FC = () => {
  const [testing, setTesting] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [embyUsers, setEmbyUsers] = useState<EmbyUser[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<'success' | 'error' | null>(null);

  // 测试连接
  const handleTestConnection = async () => {
    setTesting(true);
    try {
      const response = await testConnection();
      if (response.code === 200) {
        message.success('连接测试成功');
        setConnectionStatus('success');
      } else {
        message.error('连接测试失败');
        setConnectionStatus('error');
      }
    } catch (error) {
      message.error('连接测试失败');
      setConnectionStatus('error');
    } finally {
      setTesting(false);
    }
  };

  // 同步用户
  const handleSyncUsers = async () => {
    setSyncing(true);
    try {
      const response = await syncUsers();
      if (response.code === 200) {
        message.success('用户同步成功');
        loadEmbyUsers();
      } else {
        message.error(response.message || '用户同步失败');
      }
    } catch (error) {
      message.error('用户同步失败');
    } finally {
      setSyncing(false);
    }
  };

  // 加载Emby用户列表
  const loadEmbyUsers = async () => {
    setLoading(true);
    try {
      const response = await getEmbyUsers();
      if (response.code === 200 && response.data) {
        // 如果是数组，直接使用；如果是对象，取list属性
        const userList = Array.isArray(response.data) 
          ? response.data 
          : ((response.data as any).list || []);
        setEmbyUsers(userList);
      }
    } catch (error) {
      message.error('加载Emby用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 表格列定义
  const columns = [
    {
      title: 'Emby用户ID',
      dataIndex: 'id',
      key: 'id',
    },
    {
      title: '用户名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '最后登录时间',
      dataIndex: 'last_login_date',
      key: 'last_login_date',
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
    {
      title: '最后活动时间',
      dataIndex: 'last_activity_date',
      key: 'last_activity_date',
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
  ];

  return (
    <div>
      {/* 页面头部 */}
      <div className="page-header">
        <h1>Emby 同步</h1>
        <p>同步Emby服务器用户数据和状态</p>
      </div>

      {connectionStatus && (
        <Alert
          message={connectionStatus === 'success' ? 'Emby服务器连接正常' : 'Emby服务器连接失败'}
          type={connectionStatus}
          showIcon
          closable
          style={{ marginBottom: 16 }}
        />
      )}

      <Card title="操作" style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ApiOutlined />}
            loading={testing}
            onClick={handleTestConnection}
          >
            测试连接
          </Button>
          <Button
            type="primary"
            icon={<SyncOutlined />}
            loading={syncing}
            onClick={handleSyncUsers}
          >
            同步用户
          </Button>
          <Button
            icon={<ReloadOutlined />}
            loading={loading}
            onClick={loadEmbyUsers}
          >
            刷新列表
          </Button>
        </Space>
      </Card>

      <Card title="Emby用户列表">
        <Table
          columns={columns}
          dataSource={embyUsers}
          rowKey="id"
          loading={loading}
          pagination={false}
        />
      </Card>
    </div>
  );
};

export default EmbySync;

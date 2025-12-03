import React, { useState, useEffect } from 'react';
import { Table, Tag, message } from 'antd';
import { getPermissions } from '@/api/permission';
import type { Permission } from '@/types';

const PermissionList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [permissions, setPermissions] = useState<Permission[]>([]);

  // 加载权限列表
  const loadPermissions = async () => {
    setLoading(true);
    try {
      const response = await getPermissions();
      if (response.code === 200 && response.data) {
        // 如果是数组，直接使用；如果是对象，取list属性
        const permissionList = Array.isArray(response.data) 
          ? response.data 
          : ((response.data as any).list || []);
        setPermissions(permissionList);
      }
    } catch (error) {
      message.error('加载权限列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPermissions();
  }, []);

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'permission_id',
      key: 'permission_id',
      width: 80,
    },
    {
      title: '权限名称',
      dataIndex: 'permission_name',
      key: 'permission_name',
    },
    {
      title: '权限标识',
      dataIndex: 'permission_key',
      key: 'permission_key',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
  ];

  return (
    <div>
      {/* 页面头部 */}
      <div className="page-header">
        <h1>权限管理</h1>
        <p>查看系统所有权限定义</p>
      </div>

      <Table
        columns={columns}
        dataSource={permissions}
        rowKey="permission_id"
        loading={loading}
        pagination={false}
      />
    </div>
  );
};

export default PermissionList;

import React, { useState, useEffect } from 'react';
import { Table, Button, Space, Modal, Form, Input, message, Popconfirm, Transfer } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, SafetyOutlined } from '@ant-design/icons';
import { getRoles, createRole, updateRole, deleteRole, assignPermissions } from '@/api/role';
import { getPermissions } from '@/api/permission';
import type { Role, Permission, RoleCreateRequest } from '@/types';

const RoleList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [permissionModalVisible, setPermissionModalVisible] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [selectedPermissions, setSelectedPermissions] = useState<number[]>([]);
  const [form] = Form.useForm();

  // 加载角色列表
  const loadRoles = async () => {
    setLoading(true);
    try {
      const response = await getRoles();
      if (response.code === 200 && response.data) {
        // 如果是数组，直接使用；如果是对象，取list属性
        const roleList = Array.isArray(response.data) 
          ? response.data 
          : ((response.data as any).list || []);
        setRoles(roleList);
      }
    } catch (error) {
      message.error('加载角色列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载权限列表
  const loadPermissions = async () => {
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
      console.error('加载权限列表失败:', error);
    }
  };

  useEffect(() => {
    loadRoles();
    loadPermissions();
  }, []);

  // 打开新增/编辑对话框
  const handleOpenModal = (role?: Role) => {
    setEditingRole(role || null);
    if (role) {
      form.setFieldsValue({
        role_name: role.role_name,
        description: role.description,
      });
    } else {
      form.resetFields();
    }
    setModalVisible(true);
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      if (editingRole) {
        // 更新角色
        await updateRole(editingRole.role_id, values);
        message.success('更新成功');
      } else {
        // 创建角色
        const createData: RoleCreateRequest = {
          role_name: values.role_name,
          description: values.description,
        };
        await createRole(createData);
        message.success('创建成功');
      }
      
      setModalVisible(false);
      loadRoles();
    } catch (error) {
      console.error('操作失败:', error);
    }
  };

  // 删除角色
  const handleDelete = async (roleId: number) => {
    try {
      await deleteRole(roleId);
      message.success('删除成功');
      loadRoles();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 打开权限分配对话框
  const handleOpenPermissionModal = (role: Role) => {
    setSelectedRole(role);
    const permissionIds = role.permissions?.map(p => p.permission_id) || [];
    setSelectedPermissions(permissionIds);
    setPermissionModalVisible(true);
  };

  // 提交权限分配
  const handleAssignPermissions = async () => {
    if (!selectedRole) return;
    
    try {
      await assignPermissions(selectedRole.role_id, selectedPermissions);
      message.success('权限分配成功');
      setPermissionModalVisible(false);
      loadRoles();
    } catch (error) {
      message.error('权限分配失败');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'role_id',
      key: 'role_id',
      width: 80,
    },
    {
      title: '角色名称',
      dataIndex: 'role_name',
      key: 'role_name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '权限数量',
      dataIndex: 'permissions',
      key: 'permissions',
      render: (permissions: Permission[]) => permissions?.length || 0,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 250,
      render: (_: any, record: Role) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<SafetyOutlined />}
            onClick={() => handleOpenPermissionModal(record)}
          >
            分配权限
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleOpenModal(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此角色吗？"
            onConfirm={() => handleDelete(record.role_id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 页面头部 */}
      <div className="page-header">
        <h1>角色管理</h1>
        <p>管理系统角色和权限分配</p>
      </div>

      {/* 操作栏 */}
      <div className="page-actions">
        <div className="page-actions-left" />
        <div className="page-actions-right">
          <Button icon={<ReloadOutlined />} onClick={loadRoles}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => handleOpenModal()}>
            新增角色
          </Button>
        </div>
      </div>

      <Table
        columns={columns}
        dataSource={roles}
        rowKey="role_id"
        loading={loading}
        pagination={false}
      />

      {/* 新增/编辑角色对话框 */}
      <Modal
        title={editingRole ? '编辑角色' : '新增角色'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="role_name"
            label="角色名称"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input placeholder="请输入角色名称" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="请输入角色描述" rows={4} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 权限分配对话框 */}
      <Modal
        title={`为角色 "${selectedRole?.role_name}" 分配权限`}
        open={permissionModalVisible}
        onOk={handleAssignPermissions}
        onCancel={() => setPermissionModalVisible(false)}
        width={700}
      >
        <Transfer
          dataSource={permissions.map(p => ({
            key: p.permission_id.toString(),
            title: p.permission_name,
            description: p.description,
          }))}
          targetKeys={selectedPermissions.map(id => id.toString())}
          onChange={(targetKeys) => {
            setSelectedPermissions(targetKeys.map(key => parseInt(String(key))));
          }}
          render={item => item.title}
          listStyle={{ width: 300, height: 400 }}
          titles={['可用权限', '已分配权限']}
        />
      </Modal>
    </div>
  );
};

export default RoleList;

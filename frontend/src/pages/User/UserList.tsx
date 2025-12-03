import React, { useState, useEffect } from 'react';
import { Table, Button, Space, Modal, Form, Input, Select, message, Tag, Tooltip, Popconfirm, Descriptions, Badge, InputNumber, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, EyeOutlined, SyncOutlined, CheckCircleOutlined, CloseCircleOutlined, CrownOutlined } from '@ant-design/icons';
import { getUsers, createUser, updateUser, deleteUser, resetPassword, setUserVip } from '@/api/user';
import { getRoles } from '@/api/role';
import { usePermission } from '@/hooks/usePermission';
import type { User, Role, UserCreateRequest, UserUpdateRequest, PaginationParams } from '@/types';

const UserList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [, setTotal] = useState(0);
  const [pagination, setPagination] = useState<PaginationParams>({ page: 1, page_size: 10 });
  const [searchKeyword, setSearchKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState<number | undefined>(undefined);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [vipModalVisible, setVipModalVisible] = useState(false);
  const [vipUser, setVipUser] = useState<User | null>(null);
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [vipForm] = Form.useForm();
  const { hasPermission } = usePermission();


  // 加载用户列表
  const loadUsers = async () => {
    setLoading(true);
    try {
      const response = await getUsers(pagination);
      if (response.code === 200 && response.data) {
        setUsers(response.data.list || []);
        setTotal(response.data.total || 0);
      }
    } catch (error) {
      message.error('加载用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载角色列表
  const loadRoles = async () => {
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
      console.error('加载角色列表失败:', error);
    }
  };

  useEffect(() => {
    loadUsers();
    loadRoles();
  }, [pagination.page, pagination.page_size]);

  // 打开新增/编辑对话框
  const handleOpenModal = (user?: User) => {
    setEditingUser(user || null);
    if (user) {
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        role_id: user.role_id,
        status: user.status,
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
      
      if (editingUser) {
        // 更新用户
        const updateData: UserUpdateRequest = {
          email: values.email,
          role_id: values.role_id,
          status: values.status,
        };
        await updateUser(editingUser.user_id, updateData);
        message.success('更新成功');
      } else {
        // 创建用户
        const createData: UserCreateRequest = {
          username: values.username,
          password: values.password,
          email: values.email,
          role_id: values.role_id,
        };
        await createUser(createData);
        message.success('创建成功');
      }
      
      setModalVisible(false);
      loadUsers();
    } catch (error) {
      console.error('操作失败:', error);
    }
  };

  // 删除用户
  const handleDelete = async (userId: number) => {
    try {
      await deleteUser(userId);
      message.success('删除成功，Emby账号已同步删除');
      loadUsers();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 查看用户详情
  const handleViewDetail = (user: User) => {
    setSelectedUser(user);
    setDetailModalVisible(true);
  };

  // 打开重置密码弹窗
  const handleOpenPasswordModal = (user: User) => {
    setSelectedUser(user);
    setPasswordModalVisible(true);
    passwordForm.resetFields();
  };

  // 重置密码
  const handleResetPassword = async (values: { password: string }) => {
    if (!selectedUser) return;
    try {
      await resetPassword(selectedUser.user_id, values.password);
      message.success('密码重置成功，Emby密码已同步更新');
      setPasswordModalVisible(false);
    } catch (error) {
      message.error('密码重置失败');
    }
  };

  // 打开VIP设置弹窗
  const handleOpenVipModal = (user: User) => {
    setVipUser(user);
    setVipModalVisible(true);
    vipForm.resetFields();
    vipForm.setFieldValue('days', 30);
  };

  // 设置VIP
  const handleSetVip = async (values: { days: number }) => {
    if (!vipUser) return;
    try {
      const response: any = await setUserVip(vipUser.user_id, values.days);
      if (response.code === 200) {
        message.success(`VIP设置成功，到期时间：${new Date(response.data.vip_expire_at).toLocaleDateString('zh-CN')}`);
        setVipModalVisible(false);
        loadUsers();
      } else {
        message.error(response.message || 'VIP设置失败');
      }
    } catch (error) {
      message.error('VIP设置失败');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 80,
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Emby关联',
      dataIndex: 'emby_user_id',
      key: 'emby_user_id',
      width: 150,
      render: (embyUserId: string) => (
        embyUserId ? (
          <Tooltip title={embyUserId}>
            <Tag color="green" icon={<CheckCircleOutlined />}>
              已关联
            </Tag>
          </Tooltip>
        ) : (
          <Tag color="default" icon={<CloseCircleOutlined />}>
            未关联
          </Tag>
        )
      ),
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: Role, record: User) => {
        // 优先使用role对象，否则根据role_id查找
        if (role?.role_name) {
          return <Tag color="blue">{role.role_name}</Tag>;
        }
        const foundRole = roles.find(r => r.role_id === record.role_id);
        return foundRole ? <Tag color="blue">{foundRole.role_name}</Tag> : '-';
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => (
        <Tag color={status === 1 ? 'green' : 'red'}>
          {status === 1 ? '正常' : '禁用'}
        </Tag>
      ),
    },
    {
      title: 'VIP',
      dataIndex: 'vip_level',
      key: 'vip_level',
      width: 100,
      render: (vipLevel: number, record: User) => {
        if (vipLevel === 1) {
          const expireAt = record.vip_expire_at ? new Date(record.vip_expire_at) : null;
          const isExpired = expireAt ? expireAt < new Date() : true;
          if (isExpired) {
            return <Tag color="default">已过期</Tag>;
          }
          return (
            <Tooltip title={`到期: ${expireAt?.toLocaleDateString('zh-CN')}`}>
              <Tag color="gold">VIP</Tag>
            </Tooltip>
          );
        }
        return <Tag color="default">普通</Tag>;
      },
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
      width: 280,
      render: (_: any, record: User) => (
        <Space size="small">
          <Button 
            type="link" 
            size="small" 
            icon={<EyeOutlined />} 
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
          {hasPermission('user:edit') && (
            <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleOpenModal(record)}>
              编辑
            </Button>
          )}
          {hasPermission('user:edit') && (
            <Button type="link" size="small" icon={<SyncOutlined />} onClick={() => handleOpenPasswordModal(record)}>
              重置密码
            </Button>
          )}
          {hasPermission('user:edit') && (
            <Button type="link" size="small" icon={<CrownOutlined />} style={{ color: '#faad14' }} onClick={() => handleOpenVipModal(record)}>
              VIP
            </Button>
          )}
          {hasPermission('user:delete') && (
            <Popconfirm
              title="确定删除此用户吗？"
              description="删除后将同时删除关联的Emby账号"
              onConfirm={() => handleDelete(record.user_id)}
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

  // 过滤用户列表
  const filteredUsers = users.filter(user => {
    const matchSearch = !searchKeyword || 
      user.username.toLowerCase().includes(searchKeyword.toLowerCase()) ||
      (user.email && user.email.toLowerCase().includes(searchKeyword.toLowerCase()));
    const matchStatus = statusFilter === undefined || user.status === statusFilter;
    return matchSearch && matchStatus;
  });

  return (
    <div style={{ padding: '0 4px' }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ fontSize: 28, fontWeight: 700, color: '#1d1d1f', margin: 0, letterSpacing: '-0.5px' }}>
          用户管理
        </h1>
        <p style={{ color: '#86868b', marginTop: 4, fontSize: 14, margin: '4px 0 0' }}>
          管理系统用户、角色分配和VIP权限
        </p>
      </div>

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
              placeholder="搜索用户名或邮箱"
              allowClear
              style={{ width: 180 }}
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onPressEnter={() => setSearchKeyword(searchKeyword)}
            />
            <Button type="primary" onClick={() => setSearchKeyword(searchKeyword)}>搜索</Button>
          </Space.Compact>
          <Select
            placeholder="状态筛选"
            allowClear
            style={{ width: 120 }}
            value={statusFilter}
            onChange={setStatusFilter}
          >
            <Select.Option value={1}>正常</Select.Option>
            <Select.Option value={0}>禁用</Select.Option>
          </Select>
        </Space>
        <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={loadUsers}>刷新</Button>
          {hasPermission('user:create') && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => handleOpenModal()}>
              新增用户
            </Button>
          )}
        </Space>
      </div>

      {/* 用户列表 */}
      <Card 
        styles={{ body: { padding: 0 } }}
        style={{ borderRadius: 12, boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}
      >
        <Table
          columns={columns}
          dataSource={filteredUsers}
          rowKey="user_id"
          loading={loading}
          pagination={{
            current: pagination.page,
            pageSize: pagination.page_size,
            total: filteredUsers.length,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (page, pageSize) => {
              setPagination({ page, page_size: pageSize });
            },
          }}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title={editingUser ? '编辑用户' : '新增用户'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' }
            ]}
          >
            <Input placeholder="请输入用户名" disabled={!!editingUser} />
          </Form.Item>

          {!editingUser && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少6个字符' }
              ]}
            >
              <Input.Password placeholder="请输入密码" autoComplete="new-password" />
            </Form.Item>
          )}

          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" type="email" />
          </Form.Item>

          <Form.Item
            name="role_id"
            label="角色"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select placeholder="请选择角色">
              {roles.map(role => (
                <Select.Option key={role.role_id} value={role.role_id}>
                  {role.role_name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          {editingUser && (
            <Form.Item name="status" label="状态" initialValue={1}>
              <Select>
                <Select.Option value={1}>正常</Select.Option>
                <Select.Option value={0}>禁用</Select.Option>
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>

      {/* 用户详情弹窗 */}
      <Modal
        title="用户详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={600}
      >
        {selectedUser && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="用户ID">{selectedUser.user_id}</Descriptions.Item>
            <Descriptions.Item label="用户名">{selectedUser.username}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{selectedUser.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="角色">{selectedUser.role?.role_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Badge status={selectedUser.status === 1 ? 'success' : 'error'} text={selectedUser.status === 1 ? '正常' : '禁用'} />
            </Descriptions.Item>
            <Descriptions.Item label="Emby关联状态">
              {selectedUser.emby_user_id ? (
                <Tag color="green" icon={<CheckCircleOutlined />}>已关联</Tag>
              ) : (
                <Tag color="default" icon={<CloseCircleOutlined />}>未关联</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="Emby用户ID">
              {selectedUser.emby_user_id || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {selectedUser.created_at ? new Date(selectedUser.created_at).toLocaleString('zh-CN') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {selectedUser.updated_at ? new Date(selectedUser.updated_at).toLocaleString('zh-CN') : '-'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* 重置密码弹窗 */}
      <Modal
        title={`重置密码 - ${selectedUser?.username}`}
        open={passwordModalVisible}
        onCancel={() => setPasswordModalVisible(false)}
        onOk={() => passwordForm.submit()}
        okText="确定"
        cancelText="取消"
      >
        <Form form={passwordForm} onFinish={handleResetPassword} layout="vertical">
          <Form.Item
            name="password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6个字符' }
            ]}
          >
            <Input.Password placeholder="请输入新密码（同时更新Emby密码）" autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认密码"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入密码" autoComplete="new-password" />
          </Form.Item>
        </Form>
        <div style={{ marginTop: 8, color: '#666', fontSize: 12 }}>
          💡 提示：密码将同步更新到Emby服务器
        </div>
      </Modal>

      {/* VIP设置弹窗 */}
      <Modal
        title={<><CrownOutlined style={{ color: '#faad14' }} /> 设置VIP - {vipUser?.username}</>}
        open={vipModalVisible}
        onCancel={() => setVipModalVisible(false)}
        onOk={() => vipForm.submit()}
        okText="确定"
        cancelText="取消"
      >
        <div style={{ marginBottom: 16, padding: 12, background: '#fffbe6', borderRadius: 4 }}>
          <p style={{ margin: 0, fontSize: 13 }}>
            💡 为用户增加VIP时长
          </p>
          <p style={{ margin: '8px 0 0', fontSize: 12, color: '#666' }}>
            • 如果用户已是VIP，时长会在原有基础上叠加<br/>
            • 如果用户不是VIP，将从现在开始计算
          </p>
        </div>
        <Form form={vipForm} onFinish={handleSetVip} layout="vertical">
          <Form.Item
            name="days"
            label="VIP天数"
            rules={[{ required: true, message: '请输入天数' }]}
          >
            <Space.Compact style={{ width: '100%' }}>
              <InputNumber 
                min={1} 
                max={3650} 
                style={{ flex: 1 }} 
                placeholder="请输入VIP天数"
              />
              <Button disabled style={{ pointerEvents: 'none' }}>天</Button>
            </Space.Compact>
          </Form.Item>
          <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
            <Button size="small" onClick={() => vipForm.setFieldValue('days', 30)}>30天</Button>
            <Button size="small" onClick={() => vipForm.setFieldValue('days', 90)}>90天</Button>
            <Button size="small" onClick={() => vipForm.setFieldValue('days', 180)}>180天</Button>
            <Button size="small" onClick={() => vipForm.setFieldValue('days', 365)}>365天</Button>
          </div>
        </Form>
        {vipUser?.vip_level === 1 && vipUser?.vip_expire_at && (
          <div style={{ color: '#666', fontSize: 12 }}>
            当前VIP到期时间：{new Date(vipUser.vip_expire_at).toLocaleDateString('zh-CN')}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default UserList;

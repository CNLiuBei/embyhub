import { useState, useMemo, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Space, Tag, Input, Select, Popconfirm, App, Upload, Modal, Tooltip, Progress, Collapse, Switch, InputNumber, Card } from 'antd'
import { SearchOutlined, ReloadOutlined, StopOutlined, CheckOutlined, UploadOutlined, DownloadOutlined, CheckCircleOutlined, CloseCircleOutlined, SyncOutlined, CrownOutlined, CloudUploadOutlined, CloudDownloadOutlined, DeleteOutlined, ClockCircleOutlined, SaveOutlined, GiftOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'
import UserAvatar from '../../components/UserAvatar'
import dayjs from 'dayjs'

interface UserItem {
  id: string
  username: string
  email: string
  nickname: string
  avatar: string
  status: number
  role: number
  member_level: number
  member_expire: string | null
  last_login_at: string | null
  created_at: string
  emby_user_id?: string  // Emby用户ID
}

interface EmbyUser {
  id: string
  name: string
  is_admin: boolean
  is_disabled: boolean
  has_password: boolean
  last_login_date: string
  last_activity_date: string
  enable_media_playback: boolean
  enable_remote_access: boolean
  enable_all_folders: boolean
}

// 合并后的用户类型
interface MergedUser extends UserItem {
  emby_synced: boolean  // 是否已同步到Emby
  emby_status?: 'synced' | 'not_synced' | 'status_mismatch' | 'emby_only'  // Emby同步状态
  emby_info?: EmbyUser  // Emby用户信息
  emby_user_id?: string // Emby用户ID（从后端返回）
}

interface UserCleanupSettings {
  enabled: boolean
  inactive_days: number
  expired_days: number
  delete_emby_account: boolean
}

const UserManagement = () => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [filters, setFilters] = useState<{ username?: string; email?: string; nickname?: string; status?: number; role?: number }>({})
  const [searchForm, setSearchForm] = useState<{ username?: string; email?: string; nickname?: string; status?: number; role?: number }>({})
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([])
  const [importModalOpen, setImportModalOpen] = useState(false)
  const [importLoading, setImportLoading] = useState(false)
  const [importResult, setImportResult] = useState<{ total: number; success: number; failed: number; errors?: string[] } | null>(null)
  const [syncLoading, setSyncLoading] = useState(false)
  
  // 批量续费
  const [renewModalOpen, setRenewModalOpen] = useState(false)
  const [renewDays, setRenewDays] = useState(30)
  const [renewLoading, setRenewLoading] = useState(false)
  
  // 用户清理设置
  const [cleanupSettings, setCleanupSettings] = useState<UserCleanupSettings>({
    enabled: false,
    inactive_days: 90,
    expired_days: 30,
    delete_emby_account: true
  })
  const [cleanupLoading, setCleanupLoading] = useState(false)
  const [cleanupSaving, setCleanupSaving] = useState(false)

  // 加载用户清理设置
  useEffect(() => {
    const loadCleanupSettings = async () => {
      try {
        setCleanupLoading(true)
        const res = await adminApi.getUserCleanupSettings()
        setCleanupSettings(res.data.data)
      } catch {
        // 使用默认值
      } finally {
        setCleanupLoading(false)
      }
    }
    loadCleanupSettings()
  }, [])

  // 保存用户清理设置
  const handleSaveCleanupSettings = async () => {
    try {
      setCleanupSaving(true)
      await adminApi.saveUserCleanupSettings(cleanupSettings)
      message.success('清理设置已保存')
    } catch {
      message.error('保存失败')
    } finally {
      setCleanupSaving(false)
    }
  }

  // 本地用户列表
  const { data, isLoading } = useQuery({
    queryKey: ['userList', page, pageSize, filters],
    queryFn: async () => {
      const response = await adminApi.getUserList({ page, page_size: pageSize, ...filters })
      return response.data.data as { list: UserItem[]; total: number }
    },
  })

  // Emby用户列表 - 始终加载用于合并显示
  const { data: embyUsers, isLoading: embyLoading, refetch: refetchEmby } = useQuery({
    queryKey: ['embyUsers'],
    queryFn: async () => {
      try {
        const response = await adminApi.getEmbyUsers()
        return response.data.data as EmbyUser[]
      } catch {
        return [] as EmbyUser[]
      }
    },
  })

  // 合并本地用户和Emby用户数据
  const mergedUsers = useMemo(() => {
    const result: MergedUser[] = []
    const embyUserMap = new Map<string, EmbyUser>()
    const localUserNames = new Set<string>()
    
    // 建立Emby用户映射
    if (embyUsers) {
      embyUsers.forEach(eu => embyUserMap.set(eu.name.toLowerCase(), eu))
    }
    
    // 处理本地用户
    if (data?.list) {
      data.list.forEach(user => {
        localUserNames.add(user.username.toLowerCase())
        const embyUser = embyUserMap.get(user.username.toLowerCase())
        let embyStatus: 'synced' | 'not_synced' | 'status_mismatch' = 'not_synced'
        
        if (embyUser) {
          const localEnabled = user.status === 1
          const embyEnabled = !embyUser.is_disabled
          embyStatus = localEnabled === embyEnabled ? 'synced' : 'status_mismatch'
        }
        
        result.push({
          ...user,
          emby_synced: !!embyUser,
          emby_status: embyStatus,
          emby_info: embyUser,
        } as MergedUser)
      })
    }
    
    // 添加仅存在于Emby的用户（本地不存在）
    if (embyUsers) {
      embyUsers.forEach(embyUser => {
        if (!localUserNames.has(embyUser.name.toLowerCase())) {
          result.push({
            id: `emby_${embyUser.id}`,
            username: embyUser.name,
            email: '',
            nickname: embyUser.name,
            avatar: '',
            status: embyUser.is_disabled ? 2 : 1,
            role: embyUser.is_admin ? 3 : 0,
            member_level: 0,
            member_expire: null,
            last_login_at: embyUser.last_login_date || null,
            created_at: '',
            emby_synced: true,
            emby_status: 'emby_only' as any,
            emby_info: embyUser,
          } as MergedUser)
        }
      })
    }
    
    return result
  }, [data?.list, embyUsers])

  // 处理搜索
  const handleSearch = () => {
    setFilters(searchForm)
    setPage(1)
  }

  // 处理重置
  const handleReset = () => {
    setSearchForm({})
    setFilters({})
    setPage(1)
  }

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: number }) => adminApi.updateUserStatus(id, status),
    onSuccess: () => {
      message.success('操作成功，Emby账号已同步')
      queryClient.invalidateQueries({ queryKey: ['userList'] })
      setTimeout(() => refetchEmby(), 500) // 延迟刷新Emby列表
    },
  })

  const batchStatusMutation = useMutation({
    mutationFn: ({ ids, status }: { ids: string[]; status: number }) => adminApi.batchUpdateStatus(ids, status),
    onSuccess: () => {
      message.success('批量操作成功')
      setSelectedRowKeys([])
      queryClient.invalidateQueries({ queryKey: ['userList'] })
      setTimeout(() => refetchEmby(), 500)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApi.deleteUser(id),
    onSuccess: () => {
      message.success('用户已删除，Emby账号已同步删除')
      queryClient.invalidateQueries({ queryKey: ['userList'] })
      setTimeout(() => refetchEmby(), 500)
    },
  })

  // 同步单个用户到Emby
  const syncToEmbyMutation = useMutation({
    mutationFn: (id: string) => adminApi.syncUserToEmby(id),
    onSuccess: (res) => {
      message.success(res.data.data?.message || '同步成功')
      queryClient.invalidateQueries({ queryKey: ['userList'] })
      setTimeout(() => refetchEmby(), 500)
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || '同步失败')
    },
  })

  // 同步用户状态
  const syncStatusMutation = useMutation({
    mutationFn: (id: string) => adminApi.syncUserStatus(id),
    onSuccess: (res) => {
      message.success(res.data.data?.message || '状态同步成功')
      setTimeout(() => refetchEmby(), 500)
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || '状态同步失败')
    },
  })

  // 导入Emby用户到本地
  const importEmbyUserMutation = useMutation({
    mutationFn: (username: string) => adminApi.importEmbyUser(username),
    onSuccess: (res) => {
      message.success(res.data.data?.message || '导入成功')
      queryClient.invalidateQueries({ queryKey: ['userList'] })
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || '导入失败')
    },
  })

  // 批量同步所有用户到Emby
  const handleSyncAllToEmby = async () => {
    setSyncLoading(true)
    try {
      const res = await adminApi.syncAllUsers()
      const result = res.data.data
      message.success(`同步完成: 成功${result.success}个, 新建${result.new_created}个, 失败${result.failed}个`)
      queryClient.invalidateQueries({ queryKey: ['userList'] })
      setTimeout(() => refetchEmby(), 500)
    } catch (err: any) {
      message.error(err.response?.data?.message || '批量同步失败')
    } finally {
      setSyncLoading(false)
    }
  }

  // 批量导入所有Emby用户
  const handleImportAllEmby = async () => {
    setSyncLoading(true)
    try {
      const res = await adminApi.importAllEmbyUsers()
      const result = res.data.data
      if (result.total === 0) {
        message.info('没有需要导入的Emby用户')
      } else {
        message.success(`导入完成: 成功${result.success}个, 失败${result.failed}个`)
      }
      queryClient.invalidateQueries({ queryKey: ['userList'] })
    } catch (err: any) {
      message.error(err.response?.data?.message || '批量导入失败')
    } finally {
      setSyncLoading(false)
    }
  }

  // 批量续费
  const handleBatchRenew = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要续费的用户')
      return
    }
    setRenewLoading(true)
    try {
      const res = await adminApi.batchSetMember(selectedRowKeys, renewDays)
      const result = res.data.data
      message.success(`续费完成: 成功 ${result.success} 个, 失败 ${result.failed} 个`)
      setRenewModalOpen(false)
      setSelectedRowKeys([])
      queryClient.invalidateQueries({ queryKey: ['userList'] })
    } catch (err: any) {
      message.error(err.response?.data?.message || '批量续费失败')
    } finally {
      setRenewLoading(false)
    }
  }

  const roleNames = ['普通用户', '会员用户', '管理员', '超级管理员']
  const memberLevelNames = ['非会员', '普通会员', '高级会员', '超级会员']
  const memberLevelColors = ['default', 'blue', 'gold', 'purple']

  // 计算会员剩余天数
  const getMemberDaysLeft = (expireDate: string | null) => {
    if (!expireDate) return 0
    const diff = dayjs(expireDate).diff(dayjs(), 'day')
    return diff > 0 ? diff : 0
  }

  const columns = [
    {
      title: '用户',
      key: 'user',
      width: 200,
      render: (_: unknown, record: MergedUser) => (
        <div className="flex items-center gap-3">
          <UserAvatar src={record.avatar} name={record.username} size={40} />
          <div className="min-w-0">
            <div className="font-medium truncate flex items-center gap-1">
              {record.username}
              {record.role >= 2 && <CrownOutlined className="text-yellow-500" />}
            </div>
            <div className="text-xs text-gray-400 truncate">{record.nickname || record.email}</div>
          </div>
        </div>
      ),
    },
    { 
      title: '邮箱', 
      dataIndex: 'email', 
      key: 'email', 
      width: 180, 
      ellipsis: true,
      render: (email: string) => (
        <Tooltip title={email}>
          <span className="text-gray-600">{email}</span>
        </Tooltip>
      ),
    },
    {
      title: 'Emby ID',
      key: 'emby_user_id',
      width: 120,
      ellipsis: true,
      render: (_: unknown, record: MergedUser) => {
        const embyId = record.emby_user_id || record.emby_info?.id
        return embyId ? (
          <Tooltip title={embyId}>
            <span className="text-xs text-gray-500 font-mono">{embyId.substring(0, 8)}...</span>
          </Tooltip>
        ) : (
          <span className="text-gray-300">-</span>
        )
      },
    },
    {
      title: '会员等级',
      key: 'member',
      width: 150,
      render: (_: unknown, record: MergedUser) => {
        const daysLeft = getMemberDaysLeft(record.member_expire)
        const isExpired = record.member_level > 0 && daysLeft === 0
        const isExpiringSoon = daysLeft > 0 && daysLeft <= 7
        
        return (
          <div className="space-y-1">
            <Tag color={memberLevelColors[record.member_level]} className="m-0">
              {memberLevelNames[record.member_level]}
            </Tag>
            {record.member_level > 0 && record.member_expire && (
              <div className="text-xs">
                {isExpired ? (
                  <span className="text-red-500">已过期</span>
                ) : isExpiringSoon ? (
                  <span className="text-orange-500">剩余 {daysLeft} 天</span>
                ) : (
                  <span className="text-gray-400">{dayjs(record.member_expire).format('YYYY-MM-DD')} 到期</span>
                )}
              </div>
            )}
          </div>
        )
      },
    },
    {
      title: '状态',
      key: 'status',
      width: 130,
      render: (_: unknown, record: MergedUser) => {
        const isEmbyOnly = record.emby_status === 'emby_only'
        return (
          <div className="space-y-1">
            <Tag color={record.status === 1 ? 'green' : 'red'}>
              {record.status === 1 ? '正常' : '禁用'}
            </Tag>
            {isEmbyOnly ? (
              <Tooltip title="仅存在于Emby，本地未注册">
                <Tag icon={<SyncOutlined />} color="purple" className="m-0">仅Emby</Tag>
              </Tooltip>
            ) : record.emby_synced ? (
              record.emby_status === 'status_mismatch' ? (
                <Tooltip title={`本地${record.status === 1 ? '启用' : '禁用'}, Emby${record.emby_info?.is_disabled ? '禁用' : '启用'}`}>
                  <Tag icon={<SyncOutlined spin />} color="warning" className="m-0">Emby异常</Tag>
                </Tooltip>
              ) : (
                <Tag icon={<CheckCircleOutlined />} color="success" className="m-0">Emby同步</Tag>
              )
            ) : (
              <Tag icon={<CloseCircleOutlined />} color="default" className="m-0">未同步</Tag>
            )}
          </div>
        )
      },
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 100,
      render: (role: number) => {
        const colors = ['default', 'cyan', 'blue', 'purple']
        return <Tag color={colors[role]}>{roleNames[role]}</Tag>
      },
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      width: 120,
      render: (time: string) => (
        <span className="text-gray-500 text-sm">
          {time ? dayjs(time).format('MM-DD HH:mm') : '-'}
        </span>
      ),
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 110,
      render: (time: string) => (
        <span className="text-gray-500 text-sm">
          {dayjs(time).format('YYYY-MM-DD')}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right' as const,
      render: (_: unknown, record: MergedUser) => {
        const isEmbyOnly = record.emby_status === 'emby_only'
        
        // 仅Emby用户显示导入按钮
        if (isEmbyOnly) {
          return (
            <Space size="small">
              <Tooltip title="导入此Emby用户到本地数据库">
                <Button 
                  type="link" 
                  size="small" 
                  icon={<CloudDownloadOutlined />}
                  loading={importEmbyUserMutation.isPending}
                  onClick={() => importEmbyUserMutation.mutate(record.username)}
                >
                  导入
                </Button>
              </Tooltip>
            </Space>
          )
        }
        
        return (
          <Space size="small">
            <Button type="link" size="small" onClick={() => navigate(`/admin/users/${record.id}`)}>
              详情
            </Button>
            {/* 同步按钮 */}
            {!record.emby_synced && (
              <Tooltip title="在Emby创建此用户账号">
                <Button 
                  type="link" 
                  size="small" 
                  icon={<CloudUploadOutlined />}
                  loading={syncToEmbyMutation.isPending}
                  onClick={() => syncToEmbyMutation.mutate(record.id)}
                >
                  同步
                </Button>
              </Tooltip>
            )}
            {record.emby_status === 'status_mismatch' && (
              <Tooltip title="修复本地与Emby状态不一致">
                <Button 
                  type="link" 
                  size="small" 
                  icon={<SyncOutlined />}
                  className="text-orange-500"
                  loading={syncStatusMutation.isPending}
                  onClick={() => syncStatusMutation.mutate(record.id)}
                >
                  修复
                </Button>
              </Tooltip>
            )}
            {record.role < 2 && (
              <>
                {record.status === 1 ? (
                  <Popconfirm 
                    title="确定禁用该用户?" 
                    description="同步禁用Emby账号" 
                    onConfirm={() => statusMutation.mutate({ id: record.id, status: 2 })}
                  >
                    <Button type="link" size="small" danger>禁用</Button>
                  </Popconfirm>
                ) : (
                  <Popconfirm 
                    title="确定启用该用户?" 
                    description="同步启用Emby账号" 
                    onConfirm={() => statusMutation.mutate({ id: record.id, status: 1 })}
                  >
                    <Button type="link" size="small" className="text-green-600">启用</Button>
                  </Popconfirm>
                )}
                <Popconfirm 
                  title="确定删除该用户?" 
                  description="同步删除Emby账号"
                  onConfirm={() => deleteMutation.mutate(record.id)}
                >
                  <Button type="link" size="small" danger>删除</Button>
                </Popconfirm>
              </>
            )}
          </Space>
        )
      },
    },
  ]

  // 统计信息
  const stats = useMemo(() => {
    if (!mergedUsers.length) return null
    const localTotal = data?.total || 0
    const embyOnlyCount = mergedUsers.filter(u => u.emby_status === 'emby_only').length
    const synced = mergedUsers.filter(u => u.emby_synced && u.emby_status !== 'emby_only').length
    const active = mergedUsers.filter(u => u.status === 1).length
    const members = mergedUsers.filter(u => u.member_level > 0).length
    const total = mergedUsers.length
    return { 
      total, 
      localTotal,
      embyOnlyCount,
      synced, 
      active, 
      members, 
      syncRate: localTotal > 0 ? Math.round((synced / localTotal) * 100) : 0 
    }
  }, [mergedUsers, data?.total])

  return (
    <div className="flex flex-col h-full gap-4">
      {/* 搜索栏 - 移动端优化 */}
      <div className="glass-card p-4">
        {/* 移动端搜索 */}
        <div className="md:hidden space-y-3">
          <Input
            placeholder="搜索账号/邮箱/昵称"
            prefix={<SearchOutlined />}
            allowClear
            value={searchForm.username}
            onChange={(e) => setSearchForm((f) => ({ ...f, username: e.target.value }))}
            onPressEnter={handleSearch}
          />
          <div className="flex gap-2">
            <Select
              placeholder="状态"
              allowClear
              value={searchForm.status}
              onChange={(v) => setSearchForm((f) => ({ ...f, status: v }))}
              className="flex-1"
              options={[{ label: '正常', value: 1 }, { label: '禁用', value: 2 }]}
            />
            <Button icon={<SearchOutlined />} type="primary" onClick={handleSearch}>搜索</Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset} />
          </div>
        </div>
        
        {/* 桌面端搜索 */}
        <Space wrap className="hidden md:flex">
          <Input
            placeholder="账号"
            prefix={<SearchOutlined />}
            allowClear
            value={searchForm.username}
            onChange={(e) => setSearchForm((f) => ({ ...f, username: e.target.value }))}
            onPressEnter={handleSearch}
            style={{ width: 150 }}
          />
          <Input
            placeholder="邮箱"
            allowClear
            value={searchForm.email}
            onChange={(e) => setSearchForm((f) => ({ ...f, email: e.target.value }))}
            onPressEnter={handleSearch}
            style={{ width: 180 }}
          />
          <Input
            placeholder="昵称"
            allowClear
            value={searchForm.nickname}
            onChange={(e) => setSearchForm((f) => ({ ...f, nickname: e.target.value }))}
            onPressEnter={handleSearch}
            style={{ width: 130 }}
          />
          <Select
            placeholder="角色"
            allowClear
            value={searchForm.role}
            onChange={(v) => setSearchForm((f) => ({ ...f, role: v }))}
            style={{ width: 130 }}
            options={roleNames.map((name: string, i: number) => ({ label: name, value: i }))}
          />
          <Select
            placeholder="状态"
            allowClear
            value={searchForm.status}
            onChange={(v) => setSearchForm((f) => ({ ...f, status: v }))}
            style={{ width: 100 }}
            options={[{ label: '正常', value: 1 }, { label: '禁用', value: 2 }]}
          />
          <Button icon={<SearchOutlined />} type="primary" onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
          <Button icon={<SyncOutlined />} loading={embyLoading} onClick={() => refetchEmby()}>
            刷新Emby
          </Button>
          <Popconfirm
            title="批量同步到Emby"
            description="将所有本地用户同步到Emby（创建不存在的账号，修复状态不一致）"
            onConfirm={handleSyncAllToEmby}
          >
            <Button icon={<CloudUploadOutlined />} loading={syncLoading}>
              全部同步到Emby
            </Button>
          </Popconfirm>
          <Popconfirm
            title="批量导入Emby用户"
            description="将所有仅存在于Emby的用户导入到本地数据库"
            onConfirm={handleImportAllEmby}
          >
            <Button icon={<CloudDownloadOutlined />} loading={syncLoading}>
              导入全部Emby用户
            </Button>
          </Popconfirm>
          <Button icon={<UploadOutlined />} onClick={() => setImportModalOpen(true)}>
            导入用户
          </Button>
        </Space>
      </div>

      {/* 用户清理设置 */}
      <Collapse 
        className="glass-card mb-0"
        items={[{
          key: 'cleanup',
          label: (
            <span className="flex items-center gap-2">
              <DeleteOutlined className="text-red-500" />
              <span>自动清理过期用户</span>
              {cleanupSettings.enabled && <Tag color="green">已启用</Tag>}
            </span>
          ),
          children: cleanupLoading ? (
            <div className="text-center py-4 text-gray-400">加载中...</div>
          ) : (
            <div className="space-y-4">
              <div className="text-gray-500 text-sm mb-4">
                自动清理长时间未登录且会员已过期的用户，每天凌晨4点执行。
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card size="small" className="!rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-sm font-medium">启用自动清理</div>
                      <div className="text-xs text-gray-400">开启后将自动执行</div>
                    </div>
                    <Switch
                      checked={cleanupSettings.enabled}
                      onChange={(checked) => setCleanupSettings(s => ({ ...s, enabled: checked }))}
                    />
                  </div>
                </Card>

                <Card size="small" className="!rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-sm font-medium flex items-center gap-1">
                        <ClockCircleOutlined className="text-blue-500" />
                        未登录天数
                      </div>
                      <div className="text-xs text-gray-400">超过此天数未登录</div>
                    </div>
                    <InputNumber
                      min={7}
                      max={365}
                      value={cleanupSettings.inactive_days}
                      onChange={(v) => setCleanupSettings(s => ({ ...s, inactive_days: v || 90 }))}
                      disabled={!cleanupSettings.enabled}
                      addonAfter="天"
                      style={{ width: 100 }}
                    />
                  </div>
                </Card>

                <Card size="small" className="!rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-sm font-medium flex items-center gap-1">
                        <ClockCircleOutlined className="text-orange-500" />
                        会员过期天数
                      </div>
                      <div className="text-xs text-gray-400">会员过期超过此天数</div>
                    </div>
                    <InputNumber
                      min={1}
                      max={365}
                      value={cleanupSettings.expired_days}
                      onChange={(v) => setCleanupSettings(s => ({ ...s, expired_days: v || 30 }))}
                      disabled={!cleanupSettings.enabled}
                      addonAfter="天"
                      style={{ width: 100 }}
                    />
                  </div>
                </Card>

                <Card size="small" className="!rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-sm font-medium">同时删除Emby账号</div>
                      <div className="text-xs text-gray-400">清理时一并删除</div>
                    </div>
                    <Switch
                      checked={cleanupSettings.delete_emby_account}
                      onChange={(checked) => setCleanupSettings(s => ({ ...s, delete_emby_account: checked }))}
                      disabled={!cleanupSettings.enabled}
                    />
                  </div>
                </Card>
              </div>

              <div className="flex items-center justify-between pt-2 border-t border-gray-100">
                <div className="text-xs text-gray-400">
                  清理条件：会员过期超过 <span className="text-orange-500 font-medium">{cleanupSettings.expired_days}</span> 天 
                  且 最后登录超过 <span className="text-blue-500 font-medium">{cleanupSettings.inactive_days}</span> 天的非管理员用户
                </div>
                <Button 
                  type="primary" 
                  icon={<SaveOutlined />} 
                  onClick={handleSaveCleanupSettings}
                  loading={cleanupSaving}
                >
                  保存设置
                </Button>
              </div>
            </div>
          ),
        }]}
      />

      <div className="glass-card p-4 flex-1 flex flex-col">
        {/* 统计信息 - 移动端优化 */}
        {stats && (
          <>
            {/* 移动端统计卡片 */}
            <div className="md:hidden grid grid-cols-4 gap-2 mb-4">
              <div className="bg-blue-50 rounded-lg p-2 text-center">
                <div className="text-lg font-bold text-blue-600">{stats.total}</div>
                <div className="text-xs text-gray-500">总计</div>
              </div>
              <div className="bg-green-50 rounded-lg p-2 text-center">
                <div className="text-lg font-bold text-green-600">{stats.active}</div>
                <div className="text-xs text-gray-500">正常</div>
              </div>
              <div className="bg-orange-50 rounded-lg p-2 text-center">
                <div className="text-lg font-bold text-orange-600">{stats.members}</div>
                <div className="text-xs text-gray-500">会员</div>
              </div>
              <div className="bg-purple-50 rounded-lg p-2 text-center">
                <div className="text-lg font-bold text-purple-600">{stats.syncRate}%</div>
                <div className="text-xs text-gray-500">同步</div>
              </div>
            </div>
            
            {/* 桌面端统计 */}
            <div className="hidden md:flex items-center gap-6 mb-4 pb-4 border-b border-gray-100">
              <div className="flex items-center gap-2">
                <span className="text-gray-500">总计</span>
                <span className="text-xl font-bold text-blue-600">{stats.total}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-gray-500">本地</span>
                <span className="text-lg font-semibold text-indigo-600">{stats.localTotal}</span>
              </div>
              {stats.embyOnlyCount > 0 && (
                <div className="flex items-center gap-2">
                  <span className="text-gray-500">仅Emby</span>
                  <span className="text-lg font-semibold text-purple-600">{stats.embyOnlyCount}</span>
                </div>
              )}
              <div className="flex items-center gap-2">
                <span className="text-gray-500">正常</span>
                <span className="text-lg font-semibold text-green-600">{stats.active}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-gray-500">会员</span>
                <span className="text-lg font-semibold text-orange-600">{stats.members}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-gray-500">Emby同步</span>
                <Tooltip title={`本地用户中 ${stats.synced}/${stats.localTotal} 已同步到Emby`}>
                  <Progress 
                    percent={stats.syncRate} 
                    size="small" 
                    style={{ width: 80 }}
                    strokeColor="#52c41a"
                  />
                </Tooltip>
              </div>
              {embyUsers && (
                <div className="flex items-center gap-2 ml-auto">
                  <Tag color="blue">Emby用户: {embyUsers.length}</Tag>
                </div>
              )}
            </div>
          </>
        )}

        <div className="flex justify-between items-center mb-4">
          <span className="font-semibold">用户列表</span>
          {selectedRowKeys.length > 0 && (
            <Space className="hidden md:flex">
              <span className="text-gray-500">已选 {selectedRowKeys.length} 项</span>
              <Button
                icon={<GiftOutlined />}
                type="primary"
                size="small"
                onClick={() => setRenewModalOpen(true)}
              >
                批量续费
              </Button>
              <Button
                icon={<StopOutlined />}
                danger
                size="small"
                onClick={() => batchStatusMutation.mutate({ ids: selectedRowKeys, status: 2 })}
              >
                批量禁用
              </Button>
              <Button
                icon={<CheckOutlined />}
                size="small"
                onClick={() => batchStatusMutation.mutate({ ids: selectedRowKeys, status: 1 })}
              >
                批量启用
              </Button>
            </Space>
          )}
        </div>
        
        {/* 移动端卡片列表 */}
        <div className="md:hidden space-y-3 flex-1 overflow-auto">
          {(isLoading || embyLoading) ? (
            <div className="text-center py-8 text-gray-400">加载中...</div>
          ) : mergedUsers.length === 0 ? (
            <div className="text-center py-8 text-gray-400">暂无数据</div>
          ) : (
            mergedUsers.map((user) => (
              <div 
                key={user.id} 
                className="bg-white rounded-xl p-3 shadow-sm border border-gray-100"
                onClick={() => user.emby_status !== 'emby_only' && navigate(`/admin/users/${user.id}`)}
              >
                <div className="flex items-center gap-3">
                  <UserAvatar src={user.avatar} name={user.username} size={48} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium truncate">{user.username}</span>
                      {user.role >= 2 && <CrownOutlined className="text-yellow-500" />}
                      <Tag color={user.status === 1 ? 'green' : 'red'} className="ml-auto">
                        {user.status === 1 ? '正常' : '禁用'}
                      </Tag>
                    </div>
                    <div className="text-xs text-gray-400 truncate">{user.email || user.nickname}</div>
                    <div className="flex items-center gap-2 mt-1">
                      <Tag color={memberLevelColors[user.member_level]} className="m-0 text-xs">
                        {memberLevelNames[user.member_level]}
                      </Tag>
                      {user.emby_synced ? (
                        <Tag icon={<CheckCircleOutlined />} color="success" className="m-0 text-xs">Emby</Tag>
                      ) : (
                        <Tag color="default" className="m-0 text-xs">未同步</Tag>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
          {/* 移动端分页 */}
          <div className="flex justify-center py-4">
            <Space>
              <Button 
                size="small" 
                disabled={page <= 1}
                onClick={() => setPage(p => Math.max(1, p - 1))}
              >
                上一页
              </Button>
              <span className="text-gray-500 text-sm">{page} / {Math.ceil((data?.total || 0) / pageSize)}</span>
              <Button 
                size="small"
                disabled={page >= Math.ceil((data?.total || 0) / pageSize)}
                onClick={() => setPage(p => p + 1)}
              >
                下一页
              </Button>
            </Space>
          </div>
        </div>
        
        {/* 桌面端表格 */}
        <div className="hidden md:block flex-1">
          <Table
            columns={columns}
            dataSource={mergedUsers}
            rowKey="id"
            loading={isLoading || embyLoading}
            scroll={{ x: 1200 }}
            rowSelection={{
              selectedRowKeys,
              onChange: (keys) => setSelectedRowKeys(keys as string[]),
              getCheckboxProps: (record: MergedUser) => ({
                disabled: record.role >= 2, // 管理员不可选
              }),
            }}
            pagination={{
              current: page,
              pageSize,
              total: data?.total || 0,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
              showTotal: (total) => `共 ${total} 条`,
              onChange: (p, ps) => { setPage(p); setPageSize(ps) },
            }}
            size="middle"
          />
        </div>
      </div>

      {/* 导入弹窗 */}
      <Modal
        title="批量导入用户"
        open={importModalOpen}
        onCancel={() => { setImportModalOpen(false); setImportResult(null) }}
        footer={importResult ? [
          <Button key="close" onClick={() => { setImportModalOpen(false); setImportResult(null) }}>
            关闭
          </Button>
        ] : null}
        width={500}
      >
        {importResult ? (
          <div className="py-4">
            <div className="text-center mb-4">
              <div className="text-lg font-semibold mb-2">导入完成</div>
              <div className="text-gray-500">
                总计 <span className="text-blue-500 font-bold">{importResult.total}</span> 条，
                成功 <span className="text-green-500 font-bold">{importResult.success}</span> 条，
                失败 <span className="text-red-500 font-bold">{importResult.failed}</span> 条
              </div>
            </div>
            {importResult.errors && importResult.errors.length > 0 && (
              <div className="mt-4">
                <div className="text-sm text-gray-500 mb-2">错误详情：</div>
                <div className="max-h-40 overflow-y-auto bg-gray-50 p-3 rounded text-sm">
                  {importResult.errors.map((err, i) => (
                    <div key={i} className="text-red-500">{err}</div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="py-4">
            <div className="mb-4">
              <p className="text-gray-500 mb-2">支持 CSV 或 Excel 文件，格式要求：</p>
              <div className="bg-gray-50 p-3 rounded text-sm">
                <div>第1列：用户名（必填）</div>
                <div>第2列：邮箱（必填）</div>
                <div>第3列：密码（必填，至少6位）</div>
                <div>第4列：昵称（可选）</div>
              </div>
            </div>
            <div className="flex gap-2 mb-4">
              <Button 
                icon={<DownloadOutlined />} 
                onClick={async () => {
                  const res = await adminApi.getImportTemplate()
                  const url = window.URL.createObjectURL(new Blob([res.data]))
                  const link = document.createElement('a')
                  link.href = url
                  link.download = 'user_import_template.xlsx'
                  link.click()
                }}
              >
                下载模板
              </Button>
            </div>
            <Upload.Dragger
              accept=".csv,.xlsx,.xls"
              showUploadList={false}
              beforeUpload={async (file) => {
                setImportLoading(true)
                try {
                  const res = await adminApi.importUsers(file)
                  setImportResult(res.data.data)
                  queryClient.invalidateQueries({ queryKey: ['userList'] })
                  message.success('导入完成')
                } catch (err: any) {
                  message.error(err.response?.data?.message || '导入失败')
                } finally {
                  setImportLoading(false)
                }
                return false
              }}
              disabled={importLoading}
            >
              <p className="text-4xl text-gray-300 mb-2">
                <UploadOutlined />
              </p>
              <p className="text-gray-500">
                {importLoading ? '导入中...' : '点击或拖拽文件到此处上传'}
              </p>
            </Upload.Dragger>
          </div>
        )}
      </Modal>

      {/* 批量续费弹窗 */}
      <Modal
        title={<span><GiftOutlined className="mr-2 text-orange-500" />批量续费会员</span>}
        open={renewModalOpen}
        onCancel={() => setRenewModalOpen(false)}
        onOk={handleBatchRenew}
        confirmLoading={renewLoading}
        okText="确认续费"
        cancelText="取消"
      >
        <div className="py-4">
          <div className="mb-4 p-3 bg-blue-50 rounded-lg">
            <div className="text-blue-600 font-medium mb-1">已选择 {selectedRowKeys.length} 个用户</div>
            <div className="text-gray-500 text-sm">续费将为所选用户增加会员时长，已有会员时间会累加</div>
          </div>
          
          <div className="mb-4">
            <div className="text-gray-600 mb-2">选择续费时长：</div>
            <div className="grid grid-cols-4 gap-2 mb-3">
              {[
                { label: '月卡', days: 30 },
                { label: '季卡', days: 90 },
                { label: '半年卡', days: 180 },
                { label: '年卡', days: 365 },
              ].map(item => (
                <Button
                  key={item.days}
                  type={renewDays === item.days ? 'primary' : 'default'}
                  onClick={() => setRenewDays(item.days)}
                  className="h-12"
                >
                  <div>
                    <div className="font-medium">{item.label}</div>
                    <div className="text-xs opacity-70">{item.days}天</div>
                  </div>
                </Button>
              ))}
            </div>
            <div className="flex items-center gap-2">
              <span className="text-gray-500">自定义天数：</span>
              <InputNumber
                min={1}
                max={3650}
                value={renewDays}
                onChange={(v) => setRenewDays(v || 30)}
                addonAfter="天"
                style={{ width: 150 }}
              />
            </div>
          </div>

          <div className="p-3 bg-orange-50 rounded-lg text-sm">
            <div className="text-orange-600 font-medium mb-1">续费说明</div>
            <ul className="text-gray-600 list-disc list-inside space-y-1">
              <li>续费会自动恢复被禁用的账户</li>
              <li>已有会员时间会累加，不会覆盖</li>
              <li>续费成功后会发送邮件通知用户</li>
            </ul>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default UserManagement

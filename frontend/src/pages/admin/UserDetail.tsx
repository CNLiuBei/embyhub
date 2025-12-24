import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Descriptions, Button, Tag, Table, Modal, Form, Input, InputNumber, Select, Space, Popconfirm, App, Tooltip, Tabs, Progress, Badge } from 'antd'
import { ArrowLeftOutlined, CheckCircleOutlined, CloseCircleOutlined, SyncOutlined, DesktopOutlined, PlayCircleOutlined, StopOutlined, DeleteOutlined, SettingOutlined, GiftOutlined } from '@ant-design/icons'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import { adminApi } from '../../services/api'
import dayjs from 'dayjs'

interface EmbyUserInfo {
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
  enable_live_tv_access: boolean
  enable_live_tv_management: boolean
  enable_content_deletion: boolean
  enable_content_downloading: boolean
  enable_subtitle_downloading: boolean
  enable_subtitle_management: boolean
  enable_playback_remuxing: boolean
  enable_video_playback_transcoding: boolean
  enable_audio_playback_transcoding: boolean
  enable_public_sharing: boolean
  enable_all_devices: boolean
  enable_all_channels: boolean
  simultaneous_stream_limit: number
  remote_client_bitrate_limit: number
  invalid_login_attempt_count: number
}

interface EmbySession {
  id: string
  user_id: string
  user_name: string
  client: string
  device_id: string
  device_name: string
  device_type: string
  app_version: string
  last_activity_date: string
  remote_end_point: string
  is_playing: boolean
  now_playing_item?: {
    id: string
    name: string
    type: string
    series_name?: string
    season_name?: string
    episode_number?: number
    season_number?: number
    run_time_ticks?: number
    production_year?: number
  }
  play_state?: {
    position_ticks: number
    is_paused: boolean
    is_muted: boolean
    play_method?: string
  }
  transcoding_info?: {
    video_codec?: string
    audio_codec?: string
    is_video_direct: boolean
    is_audio_direct: boolean
    bitrate?: number
    width?: number
    height?: number
    completion?: number
  }
}

interface EmbyDevice {
  id: string
  name: string
  app_name: string
  app_version: string
  last_user_id?: string
  last_user_name?: string
  last_activity_date: string
}

interface UserDetailData {
  id: string
  username: string
  email: string
  nickname: string
  avatar: string
  gender: number
  status: number
  role: number
  member_level: number
  member_expire: string | null
  last_login_at: string | null
  last_login_ip: string
  register_ip: string
  created_at: string
  points: number
}

interface LoginLogItem {
  id: number
  ip: string
  device: string
  status: number
  remark: string
  created_at: string
}

// 权限标签组件
const PermissionTag = ({ enabled, label }: { enabled: boolean; label: string }) => (
  <Tooltip title={enabled ? '已启用' : '已禁用'}>
    <Tag 
      icon={enabled ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
      color={enabled ? 'success' : 'default'}
      className="m-0"
    >
      {label}
    </Tag>
  </Tooltip>
)

const UserDetail = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const currentUser = useSelector((state: RootState) => state.auth.user)
  const [resetPasswordOpen, setResetPasswordOpen] = useState(false)
  const [setMemberOpen, setSetMemberOpen] = useState(false)
  const [form] = Form.useForm()
  const [memberForm] = Form.useForm()
  const { message } = App.useApp()

  const { data: user } = useQuery<UserDetailData>({
    queryKey: ['userDetail', id],
    queryFn: async () => {
      const response = await adminApi.getUserDetail(id!)
      return response.data.data as UserDetailData
    },
    enabled: !!id,
  })

  const { data: loginLogs } = useQuery({
    queryKey: ['loginLogs', id],
    queryFn: async () => {
      const response = await adminApi.getLoginLogs(id!, { page: 1, page_size: 20 })
      return response.data.data as { list: LoginLogItem[]; total: number }
    },
    enabled: !!id,
  })

  // 获取Emby用户信息
  const { data: embyUser, isLoading: embyLoading, refetch: refetchEmby } = useQuery<EmbyUserInfo | null>({
    queryKey: ['embyUser', user?.username],
    queryFn: async () => {
      if (!user?.username) return null
      try {
        const response = await adminApi.getEmbyUserByUsername(user.username)
        return response.data.data as EmbyUserInfo
      } catch {
        return null
      }
    },
    enabled: !!user?.username,
  })

  // 获取用户的Emby会话
  const { data: embySessions, isLoading: sessionsLoading, refetch: refetchSessions } = useQuery<EmbySession[]>({
    queryKey: ['embySessions', user?.username],
    queryFn: async () => {
      if (!user?.username) return []
      try {
        const response = await adminApi.getEmbySessionsByUsername(user.username)
        return response.data.data as EmbySession[]
      } catch {
        return []
      }
    },
    enabled: !!user?.username,
  })

  // 获取用户的Emby设备
  const { data: embyDevices, isLoading: devicesLoading, refetch: refetchDevices } = useQuery<EmbyDevice[]>({
    queryKey: ['embyDevices', user?.username],
    queryFn: async () => {
      if (!user?.username) return []
      try {
        const response = await adminApi.getEmbyDevicesByUsername(user.username)
        return response.data.data as EmbyDevice[]
      } catch {
        return []
      }
    },
    enabled: !!user?.username,
  })

  const statusMutation = useMutation({
    mutationFn: (status: number) => adminApi.updateUserStatus(id!, status),
    onSuccess: () => {
      message.success('操作成功')
      queryClient.invalidateQueries({ queryKey: ['userDetail', id] })
    },
  })

  const roleMutation = useMutation({
    mutationFn: (role: number) => adminApi.updateUserRole(id!, role),
    onSuccess: () => {
      message.success('角色修改成功')
      queryClient.invalidateQueries({ queryKey: ['userDetail', id] })
    },
  })

  const resetPasswordMutation = useMutation({
    mutationFn: (password: string) => adminApi.resetPassword(id!, password),
    onSuccess: () => {
      message.success('密码重置成功')
      setResetPasswordOpen(false)
      form.resetFields()
    },
  })

  const setMemberMutation = useMutation({
    mutationFn: (data: { level: number; days: number }) => adminApi.setMember(id!, data),
    onSuccess: () => {
      message.success('会员设置成功')
      setSetMemberOpen(false)
      memberForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['userDetail', id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => adminApi.deleteUser(id!),
    onSuccess: () => {
      message.success('用户已删除')
      navigate('/admin/users')
    },
  })

  // 终止会话
  const killSessionMutation = useMutation({
    mutationFn: (sessionId: string) => adminApi.killEmbySession(sessionId),
    onSuccess: () => {
      message.success('会话已终止')
      refetchSessions()
    },
    onError: () => {
      message.error('终止会话失败')
    },
  })

  // 删除设备
  const deleteDeviceMutation = useMutation({
    mutationFn: (deviceId: string) => adminApi.deleteEmbyDevice(deviceId),
    onSuccess: () => {
      message.success('设备已删除')
      refetchDevices()
    },
    onError: () => {
      message.error('删除设备失败')
    },
  })

  // 设置流数限制
  const setStreamLimitMutation = useMutation({
    mutationFn: (limit: number) => adminApi.setEmbyUserStreamLimit(user!.username, limit),
    onSuccess: () => {
      message.success('流数限制已更新')
      refetchEmby()
      setStreamLimitOpen(false)
    },
    onError: () => {
      message.error('设置流数限制失败')
    },
  })

  // 调整积分
  const adjustPointsMutation = useMutation({
    mutationFn: (data: { points: number; remark?: string }) => 
      adminApi.adjustPoints({ user_id: id!, points: data.points, remark: data.remark }),
    onSuccess: () => {
      message.success('积分调整成功')
      setPointsAdjustOpen(false)
      pointsForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['userDetail', id] })
    },
    onError: (err: Error) => {
      message.error(err.message || '积分调整失败')
    },
  })

  const [streamLimitOpen, setStreamLimitOpen] = useState(false)
  const [streamLimitForm] = Form.useForm()
  const [pointsAdjustOpen, setPointsAdjustOpen] = useState(false)
  const [pointsForm] = Form.useForm()

  const roleNames = ['普通用户', '会员用户', '管理员', '超级管理员']
  const isSuperAdmin = currentUser?.role === 3  // 只有超级管理员可以修改角色

  const loginLogColumns = [
    { title: 'IP', dataIndex: 'ip', key: 'ip' },
    { title: '设备', dataIndex: 'device', key: 'device', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => <Tag color={status === 1 ? 'green' : 'red'}>{status === 1 ? '成功' : '失败'}</Tag>,
    },
    { title: '备注', dataIndex: 'remark', key: 'remark' },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  if (!user) return null

  return (
    <div className="flex flex-col h-full gap-4">
      {/* 顶部操作栏 */}
      <div className="glass-card p-4 flex justify-between items-center">
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>
          返回列表
        </Button>
        <Space>
          {/* 超级管理员只能重置密码 */}
          {user.role === 3 ? (
            <Button onClick={() => setResetPasswordOpen(true)}>重置密码</Button>
          ) : (
            <>
              {user.status === 1 ? (
                <Button danger onClick={() => statusMutation.mutate(2)}>禁用账号</Button>
              ) : (
                <Button type="primary" onClick={() => statusMutation.mutate(1)}>启用账号</Button>
              )}
              <Button onClick={() => setResetPasswordOpen(true)}>重置密码</Button>
              {user.role < 2 && (
                <Button type="primary" onClick={() => setSetMemberOpen(true)}>
                  {user.member_level > 0 && user.member_expire ? '续费会员' : '升级会员'}
                </Button>
              )}
              <Popconfirm 
                title="确定删除该用户？" 
                description="删除后将同步删除Emby账号，此操作不可恢复"
                onConfirm={() => deleteMutation.mutate()}
                okText="确定删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
              >
                <Button danger>删除用户</Button>
              </Popconfirm>
            </>
          )}
        </Space>
      </div>

      {/* 用户详情 */}
      <div className="glass-card p-6">
        <div className="font-semibold mb-4">用户详情</div>
        <Descriptions column={2}>
          <Descriptions.Item label="用户ID">{user.id}</Descriptions.Item>
          <Descriptions.Item label="账号">{user.username}</Descriptions.Item>
          <Descriptions.Item label="邮箱">{user.email || '-'}</Descriptions.Item>
          <Descriptions.Item label="昵称">{user.nickname}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={user.status === 1 ? 'green' : 'red'}>{user.status === 1 ? '正常' : '禁用'}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="角色">
            {user.role === 3 ? (
              // 超级管理员不可修改角色
              <div>
                <Tag color="purple">{roleNames[user.role]}</Tag>
                <span className="text-xs text-gray-400 ml-2">系统唯一</span>
              </div>
            ) : isSuperAdmin ? (
              // 只有超级管理员可修改其他用户角色
              <Select
                value={user.role}
                onChange={(v) => roleMutation.mutate(v)}
                style={{ width: 120 }}
                options={roleNames.map((name: string, i: number) => ({ label: name, value: i }))}
              />
            ) : (
              <Tag color={['default', 'cyan', 'blue', 'purple'][user.role]}>{roleNames[user.role]}</Tag>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="会员到期">
            {user.role >= 2 ? '长期' : (user.member_expire ? dayjs(user.member_expire).format('YYYY-MM-DD') : '-')}
          </Descriptions.Item>
          <Descriptions.Item label="积分">
            <Space>
              <span className="text-orange-500 font-bold">{user.points || 0}</span>
              <Button 
                type="link" 
                size="small" 
                icon={<GiftOutlined />}
                onClick={() => setPointsAdjustOpen(true)}
              >
                调整
              </Button>
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="最后登录">
            {user.last_login_at ? dayjs(user.last_login_at).format('YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="最后登录IP">{user.last_login_ip || '-'}</Descriptions.Item>
          <Descriptions.Item label="注册时间">{dayjs(user.created_at).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>
          <Descriptions.Item label="注册IP">{user.register_ip || '-'}</Descriptions.Item>
        </Descriptions>
      </div>

      {/* Emby账号信息 */}
      <div className="glass-card p-6">
        <Tabs
          defaultActiveKey="info"
          items={[
            {
              key: 'info',
              label: '账号信息',
              children: embyUser ? (
                <div className="space-y-4">
                  <div className="flex justify-between items-center">
                    <div className="text-sm text-gray-500">基本信息</div>
                    <Space>
                      <Button 
                        icon={<SettingOutlined />} 
                        size="small"
                        onClick={() => {
                          streamLimitForm.setFieldValue('limit', embyUser.simultaneous_stream_limit)
                          setStreamLimitOpen(true)
                        }}
                      >
                        设置流数限制
                      </Button>
                      <Button 
                        icon={<SyncOutlined />} 
                        size="small" 
                        loading={embyLoading}
                        onClick={() => refetchEmby()}
                      >
                        刷新
                      </Button>
                    </Space>
                  </div>
                  <Descriptions column={3} size="small">
                    <Descriptions.Item label="Emby ID">{embyUser.id}</Descriptions.Item>
                    <Descriptions.Item label="用户名">{embyUser.name}</Descriptions.Item>
                    <Descriptions.Item label="状态">
                      <Tag color={embyUser.is_disabled ? 'red' : 'green'}>
                        {embyUser.is_disabled ? '已禁用' : '正常'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="管理员">
                      <Tag color={embyUser.is_admin ? 'purple' : 'default'}>
                        {embyUser.is_admin ? '是' : '否'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="已设密码">
                      {embyUser.has_password ? '是' : '否'}
                    </Descriptions.Item>
                    <Descriptions.Item label="最后登录">
                      {embyUser.last_login_date ? dayjs(embyUser.last_login_date).format('YYYY-MM-DD HH:mm') : '-'}
                    </Descriptions.Item>
                    <Descriptions.Item label="最后活动">
                      {embyUser.last_activity_date ? dayjs(embyUser.last_activity_date).format('YYYY-MM-DD HH:mm') : '-'}
                    </Descriptions.Item>
                    <Descriptions.Item label="同时播放限制">
                      <Tag color={embyUser.simultaneous_stream_limit > 0 ? 'blue' : 'default'}>
                        {embyUser.simultaneous_stream_limit > 0 ? `${embyUser.simultaneous_stream_limit} 个` : '无限制'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="远程码率限制">
                      {embyUser.remote_client_bitrate_limit > 0 ? `${Math.round(embyUser.remote_client_bitrate_limit / 1000000)} Mbps` : '无限制'}
                    </Descriptions.Item>
                  </Descriptions>

                  <div>
                    <div className="text-sm text-gray-500 mb-2">权限设置</div>
                    <div className="flex flex-wrap gap-2">
                      <PermissionTag enabled={embyUser.enable_media_playback} label="媒体播放" />
                      <PermissionTag enabled={embyUser.enable_remote_access} label="远程访问" />
                      <PermissionTag enabled={embyUser.enable_all_folders} label="所有媒体库" />
                      <PermissionTag enabled={embyUser.enable_all_devices} label="所有设备" />
                      <PermissionTag enabled={embyUser.enable_all_channels} label="所有频道" />
                      <PermissionTag enabled={embyUser.enable_live_tv_access} label="直播电视" />
                      <PermissionTag enabled={embyUser.enable_live_tv_management} label="直播管理" />
                      <PermissionTag enabled={embyUser.enable_content_deletion} label="删除内容" />
                      <PermissionTag enabled={embyUser.enable_content_downloading} label="下载内容" />
                      <PermissionTag enabled={embyUser.enable_subtitle_downloading} label="下载字幕" />
                      <PermissionTag enabled={embyUser.enable_subtitle_management} label="字幕管理" />
                      <PermissionTag enabled={embyUser.enable_playback_remuxing} label="播放Remux" />
                      <PermissionTag enabled={embyUser.enable_video_playback_transcoding} label="视频转码" />
                      <PermissionTag enabled={embyUser.enable_audio_playback_transcoding} label="音频转码" />
                      <PermissionTag enabled={embyUser.enable_public_sharing} label="公开分享" />
                    </div>
                  </div>

                  {embyUser.invalid_login_attempt_count > 0 && (
                    <div className="text-orange-500 text-sm">
                      ⚠️ 登录失败次数: {embyUser.invalid_login_attempt_count}
                    </div>
                  )}
                </div>
              ) : embyLoading ? (
                <div className="text-gray-400 text-center py-8">加载中...</div>
              ) : (
                <div className="text-gray-400 text-center py-8">
                  该用户未同步到Emby，或Emby服务未启用
                </div>
              ),
            },
            {
              key: 'sessions',
              label: (
                <Badge count={embySessions?.filter(s => s.is_playing).length || 0} size="small" offset={[8, 0]}>
                  <span>在线会话</span>
                </Badge>
              ),
              children: (
                <div>
                  <div className="flex justify-between items-center mb-4">
                    <div className="text-sm text-gray-500">
                      当前在线 {embySessions?.length || 0} 个会话
                      {embySessions && embySessions.filter(s => s.is_playing).length > 0 && (
                        <span className="text-green-500 ml-2">
                          ({embySessions.filter(s => s.is_playing).length} 个正在播放)
                        </span>
                      )}
                    </div>
                    <Button 
                      icon={<SyncOutlined />} 
                      size="small" 
                      loading={sessionsLoading}
                      onClick={() => refetchSessions()}
                    >
                      刷新
                    </Button>
                  </div>
                  {embySessions && embySessions.length > 0 ? (
                    <div className="space-y-3">
                      {embySessions.map(session => (
                        <div 
                          key={session.id} 
                          className={`p-4 rounded-lg border ${session.is_playing ? 'border-green-200 bg-green-50' : 'border-gray-200 bg-gray-50'}`}
                        >
                          <div className="flex justify-between items-start">
                            <div className="flex-1">
                              <div className="flex items-center gap-2 mb-2">
                                <DesktopOutlined className="text-lg" />
                                <span className="font-medium">{session.device_name}</span>
                                <Tag color="blue">{session.client}</Tag>
                                {session.is_playing && (
                                  <Tag color="green" icon={<PlayCircleOutlined />}>播放中</Tag>
                                )}
                              </div>
                              <div className="text-sm text-gray-500 space-y-1">
                                <div>IP: {session.remote_end_point || '-'}</div>
                                <div>版本: {session.app_version || '-'}</div>
                                <div>最后活动: {session.last_activity_date ? dayjs(session.last_activity_date).format('YYYY-MM-DD HH:mm:ss') : '-'}</div>
                              </div>
                              {session.now_playing_item && (
                                <div className="mt-3 p-3 bg-white rounded border">
                                  <div className="text-sm font-medium text-green-600 mb-1">
                                    正在播放: {session.now_playing_item.series_name ? `${session.now_playing_item.series_name} - ` : ''}{session.now_playing_item.name}
                                  </div>
                                  {session.play_state && session.now_playing_item.run_time_ticks && (
                                    <div className="flex items-center gap-2">
                                      <Progress 
                                        percent={Math.round((session.play_state.position_ticks / session.now_playing_item.run_time_ticks) * 100)} 
                                        size="small" 
                                        className="flex-1"
                                      />
                                      {session.play_state.is_paused && <Tag color="orange">已暂停</Tag>}
                                    </div>
                                  )}
                                  {session.transcoding_info && (
                                    <div className="text-xs text-gray-400 mt-1">
                                      {session.transcoding_info.is_video_direct ? '直接播放' : `转码: ${session.transcoding_info.video_codec || '-'}`}
                                      {session.transcoding_info.width && ` ${session.transcoding_info.width}x${session.transcoding_info.height}`}
                                      {session.transcoding_info.bitrate && ` ${Math.round(session.transcoding_info.bitrate / 1000000)}Mbps`}
                                    </div>
                                  )}
                                </div>
                              )}
                            </div>
                            <Popconfirm
                              title="确定终止此会话？"
                              description="用户将被强制下线"
                              onConfirm={() => killSessionMutation.mutate(session.id)}
                              okText="确定"
                              cancelText="取消"
                            >
                              <Button 
                                danger 
                                icon={<StopOutlined />} 
                                size="small"
                                loading={killSessionMutation.isPending}
                              >
                                终止
                              </Button>
                            </Popconfirm>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : sessionsLoading ? (
                    <div className="text-gray-400 text-center py-8">加载中...</div>
                  ) : (
                    <div className="text-gray-400 text-center py-8">暂无在线会话</div>
                  )}
                </div>
              ),
            },
            {
              key: 'devices',
              label: `登录设备 (${embyDevices?.length || 0})`,
              children: (
                <div>
                  <div className="flex justify-between items-center mb-4">
                    <div className="text-sm text-gray-500">
                      共 {embyDevices?.length || 0} 个设备
                    </div>
                    <Button 
                      icon={<SyncOutlined />} 
                      size="small" 
                      loading={devicesLoading}
                      onClick={() => refetchDevices()}
                    >
                      刷新
                    </Button>
                  </div>
                  {embyDevices && embyDevices.length > 0 ? (
                    <Table
                      dataSource={embyDevices}
                      rowKey="id"
                      size="small"
                      pagination={false}
                      columns={[
                        {
                          title: '设备名称',
                          dataIndex: 'name',
                          key: 'name',
                          render: (name: string, record: EmbyDevice) => (
                            <div>
                              <div className="font-medium">{name}</div>
                              <div className="text-xs text-gray-400">{record.app_name}</div>
                            </div>
                          ),
                        },
                        {
                          title: '版本',
                          dataIndex: 'app_version',
                          key: 'app_version',
                        },
                        {
                          title: '最后活动',
                          dataIndex: 'last_activity_date',
                          key: 'last_activity_date',
                          render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
                        },
                        {
                          title: '操作',
                          key: 'action',
                          width: 80,
                          render: (_: unknown, record: EmbyDevice) => (
                            <Popconfirm
                              title="确定删除此设备？"
                              description="删除后该设备需要重新登录"
                              onConfirm={() => deleteDeviceMutation.mutate(record.id)}
                              okText="确定"
                              cancelText="取消"
                            >
                              <Button 
                                danger 
                                icon={<DeleteOutlined />} 
                                size="small"
                                loading={deleteDeviceMutation.isPending}
                              />
                            </Popconfirm>
                          ),
                        },
                      ]}
                    />
                  ) : devicesLoading ? (
                    <div className="text-gray-400 text-center py-8">加载中...</div>
                  ) : (
                    <div className="text-gray-400 text-center py-8">暂无登录设备</div>
                  )}
                </div>
              ),
            },

          ]}
        />
      </div>

      {/* 登录日志 */}
      <div className="glass-card p-6 flex-1">
        <div className="font-semibold mb-4">登录日志</div>
        <Table
          columns={loginLogColumns}
          dataSource={loginLogs?.list || []}
          rowKey="id"
          size="small"
          pagination={false}
        />
      </div>

      {/* 重置密码弹窗 */}
      <Modal
        title="重置密码"
        open={resetPasswordOpen}
        onCancel={() => setResetPasswordOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={resetPasswordMutation.isPending}
      >
        <Form form={form} name="resetPasswordForm" onFinish={(v) => resetPasswordMutation.mutate(v.password)}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 8 }]}>
            <Input.Password placeholder="请输入新密码(至少8位)" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 升级/续费会员弹窗 */}
      <Modal
        title={user?.member_level > 0 && user?.member_expire ? '续费会员' : '升级会员'}
        open={setMemberOpen}
        onCancel={() => { setSetMemberOpen(false); memberForm.resetFields() }}
        onOk={() => memberForm.submit()}
        confirmLoading={setMemberMutation.isPending}
      >
        <Form form={memberForm} name="setMemberForm" onFinish={(v) => setMemberMutation.mutate(v)} layout="vertical">
          <Form.Item label="快捷选择">
            <Space wrap>
              <Button size="small" onClick={() => memberForm.setFieldValue('days', 7)}>7天</Button>
              <Button size="small" onClick={() => memberForm.setFieldValue('days', 30)}>30天(月卡)</Button>
              <Button size="small" onClick={() => memberForm.setFieldValue('days', 90)}>90天</Button>
              <Button size="small" onClick={() => memberForm.setFieldValue('days', 180)}>180天</Button>
              <Button size="small" onClick={() => memberForm.setFieldValue('days', 365)}>365天(年卡)</Button>
            </Space>
          </Form.Item>
          <Form.Item name="days" label="会员天数" rules={[{ required: true, message: '请输入或选择会员天数' }]}>
            <Space.Compact style={{ width: '100%' }}>
              <InputNumber min={1} max={3650} placeholder="请输入会员天数" style={{ flex: 1 }} />
              <Button disabled>天</Button>
            </Space.Compact>
          </Form.Item>
          <Form.Item noStyle shouldUpdate>
            {() => {
              const days = memberForm.getFieldValue('days')
              if (!days) return null
              // 如果用户已是会员且未过期，从当前到期时间累加；否则从今天开始
              let baseDate = dayjs()
              if (user?.member_expire && dayjs(user.member_expire).isAfter(dayjs())) {
                baseDate = dayjs(user.member_expire)
              }
              const expireDate = baseDate.add(days, 'day').format('YYYY-MM-DD')
              return (
                <div className="bg-blue-50 p-3 rounded-lg mb-4">
                  <div className="text-blue-600 text-sm">
                    <div>预计到期时间：<strong>{expireDate}</strong></div>
                    <div className="text-xs text-blue-400 mt-1">
                      {days >= 365 ? '年卡会员' : '月卡会员'} · 
                      {user?.status !== 1 && ' 账户将自动恢复正常 ·'} 
                      角色将升级为会员用户
                    </div>
                  </div>
                </div>
              )
            }}
          </Form.Item>
        </Form>
      </Modal>

      {/* 设置流数限制弹窗 */}
      <Modal
        title="设置同时播放限制"
        open={streamLimitOpen}
        onCancel={() => { setStreamLimitOpen(false); streamLimitForm.resetFields() }}
        onOk={() => streamLimitForm.submit()}
        confirmLoading={setStreamLimitMutation.isPending}
      >
        <Form 
          form={streamLimitForm} 
          name="streamLimitForm"
          onFinish={(v) => setStreamLimitMutation.mutate(v.limit)} 
          layout="vertical"
        >
          <Form.Item label="快捷选择" className="mb-2">
            <Space wrap>
              <Button size="small" onClick={() => streamLimitForm.setFieldValue('limit', 0)}>无限制</Button>
              <Button size="small" onClick={() => streamLimitForm.setFieldValue('limit', 1)}>1个</Button>
              <Button size="small" onClick={() => streamLimitForm.setFieldValue('limit', 2)}>2个</Button>
              <Button size="small" onClick={() => streamLimitForm.setFieldValue('limit', 3)}>3个</Button>
              <Button size="small" onClick={() => streamLimitForm.setFieldValue('limit', 5)}>5个</Button>
            </Space>
          </Form.Item>
          <Form.Item 
            name="limit" 
            label="同时播放数量" 
            rules={[{ required: true, message: '请输入数量' }]}
            extra="设置为0表示无限制"
          >
            <InputNumber min={0} max={100} placeholder="请输入数量" style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 积分调整弹窗 */}
      <Modal
        title="调整积分"
        open={pointsAdjustOpen}
        onCancel={() => { setPointsAdjustOpen(false); pointsForm.resetFields() }}
        onOk={() => pointsForm.submit()}
        confirmLoading={adjustPointsMutation.isPending}
      >
        <Form 
          form={pointsForm} 
          name="adjustPointsForm"
          onFinish={(v) => adjustPointsMutation.mutate(v)} 
          layout="vertical"
        >
          <div className="mb-4 p-3 bg-gray-50 rounded-lg">
            <div className="text-gray-500 text-sm">当前积分</div>
            <div className="text-2xl font-bold text-orange-500">{user?.points || 0}</div>
          </div>
          <Form.Item label="快捷选择" className="mb-2">
            <Space wrap>
              <Button size="small" type="primary" ghost onClick={() => pointsForm.setFieldValue('points', 10)}>+10</Button>
              <Button size="small" type="primary" ghost onClick={() => pointsForm.setFieldValue('points', 50)}>+50</Button>
              <Button size="small" type="primary" ghost onClick={() => pointsForm.setFieldValue('points', 100)}>+100</Button>
              <Button size="small" type="primary" ghost onClick={() => pointsForm.setFieldValue('points', 500)}>+500</Button>
              <Button size="small" danger ghost onClick={() => pointsForm.setFieldValue('points', -10)}>-10</Button>
              <Button size="small" danger ghost onClick={() => pointsForm.setFieldValue('points', -50)}>-50</Button>
              <Button size="small" danger ghost onClick={() => pointsForm.setFieldValue('points', -100)}>-100</Button>
            </Space>
          </Form.Item>
          <Form.Item 
            name="points" 
            label="调整积分" 
            rules={[{ required: true, message: '请输入调整积分' }]}
            extra="正数为增加，负数为扣除"
          >
            <InputNumber placeholder="请输入调整积分" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} placeholder="调整原因（可选）" />
          </Form.Item>
          <Form.Item noStyle shouldUpdate>
            {() => {
              const points = pointsForm.getFieldValue('points')
              if (!points) return null
              const newPoints = (user?.points || 0) + points
              return (
                <div className={`p-3 rounded-lg ${points > 0 ? 'bg-green-50' : 'bg-red-50'}`}>
                  <div className={`text-sm ${points > 0 ? 'text-green-600' : 'text-red-600'}`}>
                    调整后积分：<strong>{newPoints < 0 ? 0 : newPoints}</strong>
                    {newPoints < 0 && <span className="text-red-500 ml-2">（积分不足，将扣除至0）</span>}
                  </div>
                </div>
              )
            }}
          </Form.Item>
        </Form>
      </Modal>

    </div>
  )
}

export default UserDetail

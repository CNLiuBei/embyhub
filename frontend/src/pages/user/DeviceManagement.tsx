import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { List, Button, Tag, App, Popconfirm, Empty, Spin } from 'antd'
import { MobileOutlined, DesktopOutlined, TabletOutlined, DeleteOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { userApi } from '../../services/api'

interface DeviceInfo {
  id: number
  device_id: string
  device_name: string
  device_type: string
  last_ip: string
  last_active: string
  is_current: boolean
}

const DeviceManagement = () => {
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data: devices, isLoading } = useQuery({
    queryKey: ['userDevices'],
    queryFn: async () => {
      const response = await userApi.getDevices()
      return response.data.data as DeviceInfo[]
    },
  })

  const removeDeviceMutation = useMutation({
    mutationFn: (deviceId: string) => userApi.removeDevice(deviceId),
    onSuccess: () => {
      message.success('设备已移除')
      queryClient.invalidateQueries({ queryKey: ['userDevices'] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '移除失败'),
  })

  const removeAllMutation = useMutation({
    mutationFn: () => userApi.removeAllDevices(),
    onSuccess: (res: any) => {
      const count = res.data.data?.removed_count || 0
      message.success(`已移除 ${count} 个设备`)
      queryClient.invalidateQueries({ queryKey: ['userDevices'] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '移除失败'),
  })

  const getDeviceIcon = (type: string) => {
    switch (type) {
      case 'mobile': return <MobileOutlined className="text-xl" />
      case 'tablet': return <TabletOutlined className="text-xl" />
      default: return <DesktopOutlined className="text-xl" />
    }
  }

  const otherDevicesCount = devices?.filter(d => !d.is_current).length || 0

  return (
    <div className="glass-card p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold m-0">登录设备</h3>
        {otherDevicesCount > 0 && (
          <Popconfirm
            title="确定移除所有其他设备？"
            description="移除后其他设备需要重新登录"
            onConfirm={() => removeAllMutation.mutate()}
          >
            <Button danger size="small" loading={removeAllMutation.isPending}>
              移除其他设备
            </Button>
          </Popconfirm>
        )}
      </div>

      {isLoading ? (
        <div className="text-center py-8"><Spin /></div>
      ) : !devices?.length ? (
        <Empty description="暂无设备记录" />
      ) : (
        <List
          dataSource={devices}
          renderItem={(device) => (
            <List.Item
              className={device.is_current ? 'bg-green-50 rounded-lg px-4' : 'px-4'}
              actions={[
                device.is_current ? (
                  <Tag color="green" icon={<CheckCircleOutlined />}>当前设备</Tag>
                ) : (
                  <Popconfirm
                    title="确定移除此设备？"
                    onConfirm={() => removeDeviceMutation.mutate(device.device_id)}
                  >
                    <Button 
                      type="link" 
                      danger 
                      size="small" 
                      icon={<DeleteOutlined />}
                      loading={removeDeviceMutation.isPending}
                    >
                      移除
                    </Button>
                  </Popconfirm>
                )
              ]}
            >
              <List.Item.Meta
                avatar={<div className="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center">{getDeviceIcon(device.device_type)}</div>}
                title={<span className="font-medium">{device.device_name || '未知设备'}</span>}
                description={
                  <div className="text-gray-500 text-sm">
                    <span>IP: {device.last_ip || '未知'}</span>
                    <span className="mx-2">·</span>
                    <span>最后活动: {device.last_active}</span>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}
    </div>
  )
}

export default DeviceManagement

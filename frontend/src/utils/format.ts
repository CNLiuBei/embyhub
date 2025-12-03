import dayjs from 'dayjs'

// 格式化日期时间
export const formatDateTime = (date: string | Date, format: string = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!date) return '-'
  return dayjs(date).format(format)
}

// 格式化日期
export const formatDate = (date: string | Date): string => {
  return formatDateTime(date, 'YYYY-MM-DD')
}

// 格式化时间
export const formatTime = (date: string | Date): string => {
  return formatDateTime(date, 'HH:mm:ss')
}

// 格式化相对时间
export const formatRelativeTime = (date: string | Date): string => {
  if (!date) return '-'
  const now = dayjs()
  const target = dayjs(date)
  const diff = now.diff(target, 'second')
  
  if (diff < 60) {
    return '刚刚'
  } else if (diff < 3600) {
    return `${Math.floor(diff / 60)}分钟前`
  } else if (diff < 86400) {
    return `${Math.floor(diff / 3600)}小时前`
  } else if (diff < 2592000) {
    return `${Math.floor(diff / 86400)}天前`
  } else {
    return formatDate(date)
  }
}

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

// 格式化数字（千分位）
export const formatNumber = (num: number): string => {
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

// 格式化百分比
export const formatPercent = (num: number, decimals: number = 2): string => {
  return `${(num * 100).toFixed(decimals)}%`
}

// 状态文本映射
export const getStatusText = (status: number): string => {
  return status === 1 ? '启用' : '禁用'
}

// 状态标签类型映射
export const getStatusTagType = (status: number): 'success' | 'error' => {
  return status === 1 ? 'success' : 'error'
}

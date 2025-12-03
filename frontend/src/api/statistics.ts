// 统计数据API
import { get } from '@/utils/request'
import type { Statistics } from '@/types'

// 获取统计数据
export const getStatistics = () => {
  return get<Statistics>('/statistics')
}

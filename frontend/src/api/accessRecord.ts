// 访问记录API
import { get, post } from '@/utils/request'
import type { AccessRecord, AccessRecordQuery, PaginationResponse } from '@/types'

// 获取访问记录列表
export const getAccessRecords = (params: AccessRecordQuery) => {
  return get<PaginationResponse<AccessRecord>>('/access-records', params)
}

// 创建访问记录
export const createAccessRecord = (data: Partial<AccessRecord>) => {
  return post<AccessRecord>('/access-records', data)
}

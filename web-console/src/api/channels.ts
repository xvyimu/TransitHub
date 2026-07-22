import { http } from './http'
import type { ApiResponse, ChannelListData } from '@/types/api'

export interface ChannelListQuery {
  p?: number
  page_size?: number
  status?: string
  type?: string
  group?: string
  keyword?: string
  id_sort?: boolean
}

/** Read-only channel list — GET /api/channel/ or /api/channel/search */
export async function listChannels(q: ChannelListQuery = {}) {
  const params: Record<string, string | number | boolean> = {
    p: q.p ?? 1,
    page_size: q.page_size ?? 20,
  }
  if (q.status !== undefined && q.status !== '') params.status = q.status
  if (q.type !== undefined && q.type !== '') params.type = q.type
  if (q.group) params.group = q.group
  if (q.id_sort) params.id_sort = true

  const keyword = (q.keyword || '').trim()
  if (keyword) {
    params.keyword = keyword
    const res = await http.get<ApiResponse<ChannelListData | unknown>>('/api/channel/search', {
      params,
    })
    return res.data
  }

  const res = await http.get<ApiResponse<ChannelListData>>('/api/channel/', { params })
  return res.data
}

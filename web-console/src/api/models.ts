import { http } from './http'
import type { ApiResponse, ModelListData } from '@/types/api'

export interface ModelListQuery {
  p?: number
  page_size?: number
  /** Filter status; empty / omit = all */
  status?: string
  /** Vendor id filter for search path */
  vendor?: string
  keyword?: string
  sync_official?: string
}

/** Read-only model meta list — GET /api/models/ or /api/models/search */
export async function listModels(q: ModelListQuery = {}) {
  const params: Record<string, string | number> = {
    p: q.p ?? 1,
    page_size: q.page_size ?? 20,
  }
  if (q.status !== undefined && q.status !== '') params.status = q.status
  if (q.vendor?.trim()) params.vendor = q.vendor.trim()
  if (q.sync_official !== undefined && q.sync_official !== '') {
    params.sync_official = q.sync_official
  }

  const keyword = (q.keyword || '').trim()
  if (keyword) {
    params.keyword = keyword
    const res = await http.get<ApiResponse<ModelListData | unknown>>('/api/models/search', {
      params,
    })
    return res.data
  }

  const res = await http.get<ApiResponse<ModelListData>>('/api/models/', { params })
  return res.data
}

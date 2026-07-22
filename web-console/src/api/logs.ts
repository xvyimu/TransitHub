import { http } from './http'
import type { ApiResponse, LogListData } from '@/types/api'

/** Admin role threshold — matches common.RoleAdminUser / React ROLE.ADMIN. */
export const ADMIN_ROLE = 10

export interface LogListQuery {
  p?: number
  page_size?: number
  type?: number | string
  model_name?: string
  username?: string
  token_name?: string
  channel?: number | string
  group?: string
  request_id?: string
  upstream_request_id?: string
  trace_id?: string
  start_timestamp?: number
  end_timestamp?: number
  /** Prefer admin path when true; otherwise GET /api/log/self. */
  isAdmin?: boolean
}

function buildParams(q: LogListQuery): Record<string, string | number> {
  const params: Record<string, string | number> = {
    p: q.p ?? 1,
    page_size: q.page_size ?? 20,
  }
  if (q.type !== undefined && q.type !== '' && q.type !== 'all') {
    params.type = typeof q.type === 'number' ? q.type : Number(q.type)
  }
  if (q.model_name?.trim()) params.model_name = q.model_name.trim()
  if (q.username?.trim()) params.username = q.username.trim()
  if (q.token_name?.trim()) params.token_name = q.token_name.trim()
  if (q.channel !== undefined && q.channel !== '') params.channel = q.channel
  if (q.group?.trim()) params.group = q.group.trim()
  if (q.request_id?.trim()) params.request_id = q.request_id.trim()
  if (q.upstream_request_id?.trim()) {
    params.upstream_request_id = q.upstream_request_id.trim()
  }
  if (q.trace_id?.trim()) params.trace_id = q.trace_id.trim()
  if (q.start_timestamp) params.start_timestamp = q.start_timestamp
  if (q.end_timestamp) params.end_timestamp = q.end_timestamp
  return params
}

/**
 * Read-only usage logs.
 * Admin: GET /api/log/  User: GET /api/log/self
 * On 403 from admin path, falls back to self.
 */
export async function listLogs(q: LogListQuery = {}) {
  const params = buildParams(q)
  const preferAdmin = q.isAdmin !== false

  if (preferAdmin) {
    try {
      const res = await http.get<ApiResponse<LogListData>>('/api/log/', { params })
      return res.data
    } catch (err: unknown) {
      const status =
        err && typeof err === 'object' && 'response' in err
          ? (err as { response?: { status?: number } }).response?.status
          : undefined
      if (status !== 403 && status !== 401) throw err
      // fall through to self
    }
  }

  const res = await http.get<ApiResponse<LogListData>>('/api/log/self', { params })
  return res.data
}

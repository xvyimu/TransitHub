/** Shared API response shell — aligns with Go common.Api* / gin.H success shell. */

export interface ApiResponse<T = unknown> {
  success: boolean
  message: string
  data?: T
}

export interface LoginPayload {
  username: string
  password: string
  turnstile?: string
}

export interface LoginData {
  require_2fa?: boolean
  id?: number
}

export interface UserSelf {
  id: number
  username: string
  display_name?: string
  role?: number
  status?: number
  email?: string
  group?: string
  quota?: number
  used_quota?: number
  permissions?: Record<string, unknown>
  [key: string]: unknown
}

/** Subset of GET /api/status fields used by MVP Health cards. */
export interface StatusData {
  version?: string
  system_name?: string
  start_time?: number
  register_enabled?: boolean
  password_login_enabled?: boolean
  turnstile_check?: boolean
  turnstile_site_key?: string
  setup?: boolean
  [key: string]: unknown
}

export type ProbeName = 'healthz' | 'livez' | 'readyz' | 'frontend-healthz'

export interface ProbeResult {
  name: ProbeName
  ok: boolean
  status: number | null
  body: unknown
  error?: string
}

/** Admin channel list item (key omitted by backend). */
export interface ChannelItem {
  id: number
  type?: number
  name?: string
  status?: number
  group?: string
  tag?: string
  models?: string
  priority?: number
  weight?: number
  created_time?: number
  test_time?: number
  response_time?: number
  balance?: number
  used_quota?: number
  [key: string]: unknown
}

export interface ChannelListData {
  items: ChannelItem[]
  total: number
  page: number
  page_size: number
  type_counts?: Record<string, number>
}

/** Usage log row — aligns with model.Log JSON tags. */
export interface LogItem {
  id: number
  user_id?: number
  created_at?: number
  type?: number
  content?: string
  username?: string
  token_name?: string
  model_name?: string
  quota?: number
  prompt_tokens?: number
  completion_tokens?: number
  use_time?: number
  is_stream?: boolean
  channel?: number
  channel_name?: string
  group?: string
  request_id?: string
  upstream_request_id?: string
  trace_id?: string
  other?: string
  [key: string]: unknown
}

export interface LogListData {
  items: LogItem[]
  total: number
  page: number
  page_size: number
  has_more?: boolean
  next_cursor?: string
}

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

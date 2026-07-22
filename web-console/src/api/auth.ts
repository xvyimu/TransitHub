import { http } from './http'
import type { ApiResponse, LoginData, LoginPayload, UserSelf } from '@/types/api'

export async function login(payload: LoginPayload) {
  const turnstile = payload.turnstile ?? ''
  const res = await http.post<ApiResponse<LoginData>>(
    `/api/user/login?turnstile=${encodeURIComponent(turnstile)}`,
    {
      username: payload.username,
      password: payload.password,
    },
  )
  return res.data
}

export async function logout() {
  const res = await http.get<ApiResponse>(`/api/user/logout`, {
    skipAuthRedirect: true,
  })
  return res.data
}

export async function getSelf(opts?: { skipAuthRedirect?: boolean }) {
  const res = await http.get<ApiResponse<UserSelf>>(`/api/user/self`, {
    skipAuthRedirect: opts?.skipAuthRedirect,
  })
  return res.data
}

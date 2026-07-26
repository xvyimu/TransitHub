import { http } from './http'
import type { ApiResponse, TokenListData } from '@/types/api'
import type { TokenPayload } from './tokenForm'

export interface TokenListQuery {
  p?: number
  page_size?: number
  keyword?: string
}

/**
 * Per-user API tokens — GET /api/token/ or /api/token/search (UserAuth).
 * Backend masks the key on list/get; full key is a separate Critical route.
 */
export async function listTokens(q: TokenListQuery = {}) {
  const params: Record<string, string | number> = {
    p: q.p ?? 1,
    page_size: q.page_size ?? 20,
  }
  const keyword = (q.keyword || '').trim()
  if (keyword) {
    params.keyword = keyword
    const res = await http.get<ApiResponse<TokenListData | unknown>>('/api/token/search', {
      params,
    })
    return res.data
  }
  const res = await http.get<ApiResponse<TokenListData>>('/api/token/', { params })
  return res.data
}

/** Create a token — POST /api/token/. Backend generates the key. */
export async function createToken(payload: TokenPayload) {
  const res = await http.post<ApiResponse>('/api/token/', payload)
  return res.data
}

/** Update a token — PUT /api/token/. Requires payload.id. */
export async function updateToken(payload: TokenPayload) {
  const res = await http.put<ApiResponse>('/api/token/', payload)
  return res.data
}

/** Toggle enable/disable only — PUT /api/token/?status_only=1. */
export async function setTokenStatus(id: number, status: number) {
  const res = await http.put<ApiResponse>('/api/token/?status_only=1', { id, status })
  return res.data
}

/** Delete a token — DELETE /api/token/:id. */
export async function deleteToken(id: number) {
  const res = await http.delete<ApiResponse>(`/api/token/${id}`)
  return res.data
}

/** Reveal the full (unmasked) key — POST /api/token/:id/key (Critical route). */
export async function getTokenKey(id: number) {
  const res = await http.post<ApiResponse<{ key: string }>>(`/api/token/${id}/key`)
  return res.data
}

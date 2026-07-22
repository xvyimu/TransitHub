import axios, { type AxiosError, type AxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types/api'

declare module 'axios' {
  export interface AxiosRequestConfig {
    /** Skip redirect-to-login on 401 (e.g. bootstrap self probe). */
    skipAuthRedirect?: boolean
  }
}

/** Same-origin base; production Nginx proxies /api and probes. */
export const http = axios.create({
  baseURL: '',
  withCredentials: true,
  headers: {
    'Cache-Control': 'no-store',
  },
})

let onUnauthorized: (() => void) | null = null

/** Register a handler (auth store clear + router) for HTTP 401. */
export function setUnauthorizedHandler(handler: (() => void) | null) {
  onUnauthorized = handler
}

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    const config = error.config as AxiosRequestConfig | undefined
    if (error.response?.status === 401 && !config?.skipAuthRedirect) {
      onUnauthorized?.()
    }
    return Promise.reject(error)
  },
)

export function isApiSuccess<T>(body: ApiResponse<T> | undefined | null): body is ApiResponse<T> & {
  success: true
} {
  return !!body && body.success === true
}

export function apiMessage(err: unknown, fallback = 'Request failed'): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as ApiResponse | undefined
    if (data?.message) return data.message
    if (err.message) return err.message
  }
  if (err instanceof Error && err.message) return err.message
  return fallback
}

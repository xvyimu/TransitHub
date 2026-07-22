import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as authApi from '@/api/auth'
import { apiMessage, isApiSuccess } from '@/api/http'
import type { LoginPayload, UserSelf } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserSelf | null>(null)
  const bootstrapped = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value)
  const displayName = computed(
    () => user.value?.display_name || user.value?.username || '',
  )

  function clearSession() {
    user.value = null
    error.value = null
  }

  async function bootstrap() {
    loading.value = true
    error.value = null
    try {
      const body = await authApi.getSelf({ skipAuthRedirect: true })
      if (isApiSuccess(body) && body.data) {
        user.value = body.data
      } else {
        user.value = null
      }
    } catch {
      user.value = null
    } finally {
      bootstrapped.value = true
      loading.value = false
    }
  }

  async function login(payload: LoginPayload) {
    loading.value = true
    error.value = null
    try {
      const body = await authApi.login(payload)
      if (!isApiSuccess(body)) {
        error.value = body.message || 'Login failed'
        return { ok: false as const, require2fa: false }
      }
      if (body.data?.require_2fa) {
        error.value =
          body.message ||
          'Two-factor authentication is required (not in Phase1 MVP). Use the legacy console or complete 2FA later.'
        return { ok: false as const, require2fa: true }
      }
      const self = await authApi.getSelf()
      if (!isApiSuccess(self) || !self.data) {
        error.value = self.message || 'Failed to load profile after login'
        return { ok: false as const, require2fa: false }
      }
      user.value = self.data
      return { ok: true as const, require2fa: false }
    } catch (e) {
      error.value = apiMessage(e, 'Login failed')
      return { ok: false as const, require2fa: false }
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    loading.value = true
    try {
      await authApi.logout()
    } catch {
      // best-effort; clear local session either way
    } finally {
      clearSession()
      loading.value = false
    }
  }

  return {
    user,
    bootstrapped,
    loading,
    error,
    isAuthenticated,
    displayName,
    clearSession,
    bootstrap,
    login,
    logout,
  }
})

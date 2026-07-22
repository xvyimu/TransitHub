import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('@/layouts/ConsoleLayout.vue'),
      children: [
        { path: '', redirect: '/health' },
        {
          path: 'health',
          name: 'health',
          component: () => import('@/views/HealthView.vue'),
        },
        {
          path: 'channels',
          name: 'channels',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'channels' },
        },
        {
          path: 'models',
          name: 'models',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'models' },
        },
        {
          path: 'keys',
          name: 'keys',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'keys' },
        },
        {
          path: 'logs',
          name: 'logs',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'logs' },
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'users' },
        },
        {
          path: 'billing',
          name: 'billing',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'billing' },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'settings' },
        },
        {
          path: 'system',
          name: 'system',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'system' },
        },
        {
          path: 'playground',
          name: 'playground',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'playground' },
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { domain: 'profile' },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { public: true },
    },
  ],
})

/** Open-redirect safe: only same-origin relative paths. */
export function safeRedirect(raw: unknown): string {
  if (typeof raw !== 'string' || !raw.startsWith('/') || raw.startsWith('//')) {
    return '/health'
  }
  if (raw.startsWith('/login')) return '/health'
  return raw
}

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.bootstrapped) {
    await auth.bootstrap()
  }
  if (to.meta.public) {
    if (to.name === 'login' && auth.isAuthenticated) {
      return safeRedirect(to.query.redirect)
    }
    return true
  }
  if (!auth.isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }
  return true
})

export default router

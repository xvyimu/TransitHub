<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const collapsed = ref(false)

const activeKey = computed(() => {
  const name = route.name
  return typeof name === 'string' ? name : 'health'
})

const menuItems = computed(() => [
  { key: 'health', label: t('nav.health') },
  { key: 'channels', label: t('nav.channels') },
  { key: 'models', label: t('nav.models') },
  { key: 'keys', label: t('nav.keys') },
  { key: 'logs', label: t('nav.logs') },
  { key: 'users', label: t('nav.users') },
  { key: 'billing', label: t('nav.billing') },
  { key: 'settings', label: t('nav.settings') },
  { key: 'system', label: t('nav.system') },
  { key: 'playground', label: t('nav.playground') },
  { key: 'profile', label: t('nav.profile') },
])

function onMenuClick(info: { key: string }) {
  void router.push({ name: info.key })
}

async function onLogout() {
  await auth.logout()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <a-layout class="shell">
    <a-layout-sider
      v-model:collapsed="collapsed"
      collapsible
      :width="220"
      :collapsed-width="64"
      class="sider"
    >
      <div class="brand">
        <span v-if="!collapsed" class="brand-title">{{ t('app.title') }}</span>
        <span v-else class="brand-title-collapsed">TH</span>
        <a-tag v-if="!collapsed" color="orange" style="margin-top: 6px">Phase1 MVP</a-tag>
      </div>
      <a-menu
        :selected-keys="[activeKey]"
        :items="menuItems"
        mode="inline"
        @click="onMenuClick"
      />
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="header">
        <div class="header-inner">
          <span class="legacy-notice">{{ t('app.legacyNotice') }}</span>
          <a-space>
            <span class="user-name">{{ auth.displayName }}</span>
            <a-button size="small" @click="onLogout">{{ t('nav.logout') }}</a-button>
          </a-space>
        </div>
      </a-layout-header>
      <a-layout-content class="content">
        <RouterView />
      </a-layout-content>
      <a-layout-footer class="notice-footer">
        <span>{{ t('app.noticeAttribution') }}</span>
        <a
          href="https://github.com/QuantumNous/new-api"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ t('app.originalProject') }}
        </a>
        <a
          href="https://github.com/xvyimu/TransitHub"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ t('app.transitHubSource') }}
        </a>
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.shell {
  min-height: 100vh;
}
.sider {
  background: #ffffff;
  border-right: 1px solid #f0f0f0;
}
.brand {
  padding: 16px 16px 8px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  border-bottom: 1px solid #f5f5f5;
  margin-bottom: 4px;
}
.brand-title {
  font-size: 16px;
  font-weight: 600;
  color: #18181b;
}
.brand-title-collapsed {
  font-size: 18px;
  font-weight: 700;
  color: #18181b;
}
.header {
  height: 56px;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
}
.header-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.legacy-notice {
  font-size: 12px;
  color: #a3a3a3;
}
.user-name {
  font-size: 14px;
  color: #525252;
}
.content {
  padding: 24px;
  min-height: calc(100vh - 56px - 48px);
  background: #fafafa;
}
.notice-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 12px;
  padding: 12px 24px;
  font-size: 12px;
  color: #a3a3a3;
  background: #ffffff;
  border-top: 1px solid #f0f0f0;
}
.notice-footer a {
  color: #525252;
}
.notice-footer a:hover {
  color: #18181b;
}
</style>
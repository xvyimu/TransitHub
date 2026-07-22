<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, RouterView } from 'vue-router'
import {
  NLayout,
  NLayoutHeader,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NButton,
  NSpace,
  NText,
  NTag,
  type MenuOption,
} from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeKey = computed(() => {
  const name = route.name
  return typeof name === 'string' ? name : 'health'
})

const menuOptions = computed<MenuOption[]>(() => [
  { label: t('nav.health'), key: 'health' },
  { label: t('nav.channels'), key: 'channels' },
  { label: t('nav.models'), key: 'models' },
  { label: t('nav.keys'), key: 'keys' },
  { label: t('nav.logs'), key: 'logs' },
  { label: t('nav.users'), key: 'users' },
  { label: t('nav.billing'), key: 'billing' },
  { label: t('nav.settings'), key: 'settings' },
  { label: t('nav.system'), key: 'system' },
  { label: t('nav.playground'), key: 'playground' },
  { label: t('nav.profile'), key: 'profile' },
])

function onMenuUpdate(key: string | number) {
  void router.push({ name: String(key) })
}

async function onLogout() {
  await auth.logout()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <NLayout has-sider class="shell">
    <NLayoutSider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      show-trigger
      content-style="padding-top: 8px"
    >
      <div class="brand">
        <NText strong>{{ t('app.title') }}</NText>
        <NTag size="tiny" type="warning" style="margin-top: 6px">Phase1 MVP</NTag>
      </div>
      <NMenu
        :value="activeKey"
        :options="menuOptions"
        @update:value="onMenuUpdate"
      />
    </NLayoutSider>
    <NLayout>
      <NLayoutHeader bordered class="header">
        <NSpace justify="space-between" align="center" style="width: 100%">
          <NText depth="3" style="font-size: 12px">{{ t('app.legacyNotice') }}</NText>
          <NSpace align="center">
            <NText>{{ auth.displayName }}</NText>
            <NButton size="small" @click="onLogout">{{ t('nav.logout') }}</NButton>
          </NSpace>
        </NSpace>
      </NLayoutHeader>
      <NLayoutContent content-style="padding: 20px; min-height: calc(100vh - 56px)">
        <RouterView />
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>

<style scoped>
.shell {
  min-height: 100vh;
}
.brand {
  padding: 12px 16px 8px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}
.header {
  height: 56px;
  padding: 0 16px;
  display: flex;
  align-items: center;
}
</style>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { NResult, NButton } from 'naive-ui'
import { useRouter } from 'vue-router'

const { t } = useI18n()
const router = useRouter()
</script>

<template>
  <div class="wrap">
    <NResult status="404" title="404" description="Not found">
      <template #footer>
        <NButton type="primary" @click="router.push('/health')">
          {{ t('nav.health') }}
        </NButton>
      </template>
    </NResult>
  </div>
</template>

<style scoped>
.wrap {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>

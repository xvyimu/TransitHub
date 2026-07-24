<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const { t } = useI18n()
const router = useRouter()
</script>

<template>
  <div class="wrap">
    <a-result status="404" title="404" sub-title="Not found">
      <template #extra>
        <a-button type="primary" @click="router.push('/health')">
          {{ t('nav.health') }}
        </a-button>
      </template>
    </a-result>
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
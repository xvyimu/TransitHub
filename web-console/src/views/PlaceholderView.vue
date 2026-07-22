<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { NResult } from 'naive-ui'

const { t } = useI18n()
</script>

<template>
  <div class="placeholder">
    <NResult status="info" :title="t('placeholder.title')" :description="t('placeholder.body')" />
  </div>
</template>

<style scoped>
.placeholder {
  padding: 48px 16px;
}
</style>

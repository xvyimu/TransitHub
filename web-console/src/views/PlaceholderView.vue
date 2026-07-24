<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <div class="placeholder">
    <a-result status="info" :title="t('placeholder.title')" :sub-title="t('placeholder.body')" />
  </div>
</template>

<style scoped>
.placeholder {
  padding: 48px 16px;
}
</style>
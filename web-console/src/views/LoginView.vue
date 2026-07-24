<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { safeRedirect } from '@/router'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const localError = ref<string | null>(null)

const errorMessage = computed(() => localError.value || auth.error)

async function onSubmit() {
  localError.value = null
  if (!username.value.trim() || !password.value) {
    localError.value = t('login.required')
    return
  }
  const result = await auth.login({
    username: username.value.trim(),
    password: password.value,
  })
  if (result.ok) {
    await router.replace(safeRedirect(route.query.redirect))
  }
}
</script>

<template>
  <div class="login-page">
    <a-card class="login-card" :title="t('login.title')" :bordered="true">
      <template #extra>
        <span class="subtitle">{{ t('app.legacyNotice') }}</span>
      </template>
      <a-alert
        v-if="errorMessage"
        type="error"
        :message="t('common.error')"
        :description="errorMessage"
        show-icon
        style="margin-bottom: 16px"
      />
      <a-form layout="vertical" @submit.prevent="onSubmit">
        <a-form-item :label="t('login.username')">
          <a-input
            v-model:value="username"
            autocomplete="username"
            :disabled="auth.loading"
          />
        </a-form-item>
        <a-form-item :label="t('login.password')">
          <a-input
            v-model:value="password"
            type="password"
            autocomplete="current-password"
            :disabled="auth.loading"
            @keydown.enter="onSubmit"
          />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" :loading="auth.loading" html-type="submit" block>
            {{ t('login.submit') }}
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: #fafafa;
}
.login-card {
  width: 100%;
  max-width: 400px;
}
.subtitle {
  display: block;
  font-size: 11px;
  color: #a3a3a3;
  line-height: 1.4;
}
</style>
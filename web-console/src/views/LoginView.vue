<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NAlert,
  NSpace,
  NText,
} from 'naive-ui'
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
    <NCard class="login-card" :title="t('login.title')">
      <NText depth="3" class="subtitle">{{ t('app.legacyNotice') }}</NText>
      <NAlert
        v-if="errorMessage"
        type="error"
        :title="t('common.error')"
        style="margin: 12px 0"
      >
        {{ errorMessage }}
      </NAlert>
      <NForm @submit.prevent="onSubmit">
        <NFormItem :label="t('login.username')">
          <NInput
            v-model:value="username"
            autocomplete="username"
            :disabled="auth.loading"
          />
        </NFormItem>
        <NFormItem :label="t('login.password')">
          <NInput
            v-model:value="password"
            type="password"
            show-password-on="click"
            autocomplete="current-password"
            :disabled="auth.loading"
            @keydown.enter="onSubmit"
          />
        </NFormItem>
        <NSpace justify="end">
          <NButton type="primary" :loading="auth.loading" @click="onSubmit">
            {{ t('login.submit') }}
          </NButton>
        </NSpace>
      </NForm>
    </NCard>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(160deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
}
.login-card {
  width: 100%;
  max-width: 400px;
}
.subtitle {
  display: block;
  margin-bottom: 8px;
  font-size: 12px;
  line-height: 1.4;
}
</style>

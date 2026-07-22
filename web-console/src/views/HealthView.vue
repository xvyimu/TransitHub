<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NGrid,
  NGi,
  NSpace,
  NTag,
  NSpin,
  NDescriptions,
  NDescriptionsItem,
  NAlert,
  NCode,
  NText,
} from 'naive-ui'
import { fetchProbes, getStatus } from '@/api/status'
import { apiMessage, isApiSuccess } from '@/api/http'
import type { ProbeResult, StatusData } from '@/types/api'

const { t } = useI18n()

const loading = ref(false)
const error = ref<string | null>(null)
const status = ref<StatusData | null>(null)
const probes = ref<ProbeResult[]>([])

function boolLabel(v: boolean | undefined) {
  if (v === true) return t('health.yes')
  if (v === false) return t('health.no')
  return t('health.unknown')
}

function unwrapStatus(body: {
  success?: boolean
  data?: StatusData
  version?: string
  system_name?: string
  [key: string]: unknown
}): StatusData | null {
  if (body.data && typeof body.data === 'object') {
    return body.data
  }
  // Tolerate non-nested shapes if ever returned.
  if (body.version || body.system_name) {
    return body as StatusData
  }
  return null
}

async function refresh() {
  loading.value = true
  error.value = null
  try {
    const [statusBody, probeList] = await Promise.all([getStatus(), fetchProbes()])
    probes.value = probeList
    if (isApiSuccess(statusBody)) {
      status.value = unwrapStatus(statusBody as never)
    } else {
      error.value = statusBody.message || 'status failed'
      status.value = null
    }
  } catch (e) {
    error.value = apiMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="health">
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 class="title">{{ t('health.title') }}</h2>
      <NButton type="primary" :loading="loading" @click="refresh">
        {{ t('health.refresh') }}
      </NButton>
    </NSpace>

    <NAlert v-if="error" type="warning" style="margin-bottom: 16px">
      {{ error }}
    </NAlert>

    <NSpin :show="loading">
      <NGrid cols="1 s:2" :x-gap="16" :y-gap="16" responsive="screen">
        <NGi>
          <NCard :title="t('health.probes')" size="small">
            <NSpace vertical>
              <div v-for="p in probes" :key="p.name" class="probe-row">
                <span class="probe-name">{{ p.name }}</span>
                <NTag :type="p.ok ? 'success' : 'error'" size="small">
                  {{ p.ok ? t('common.ok') : t('common.down') }}
                </NTag>
                <span class="probe-meta">
                  HTTP {{ p.status ?? '—' }}
                  <template v-if="p.error"> · {{ p.error }}</template>
                </span>
                <NCode
                  v-if="p.body && typeof p.body === 'object'"
                  class="probe-body"
                  :code="JSON.stringify(p.body)"
                  language="json"
                />
              </div>
              <NText v-if="!probes.length" depth="3">{{ t('common.loading') }}</NText>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard :title="t('health.status')" size="small">
            <NDescriptions v-if="status" label-placement="left" :column="1" size="small">
              <NDescriptionsItem :label="t('health.version')">
                {{ status.version ?? t('health.unknown') }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('health.systemName')">
                {{ status.system_name ?? t('health.unknown') }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('health.passwordLogin')">
                {{ boolLabel(status.password_login_enabled as boolean | undefined) }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('health.register')">
                {{ boolLabel(status.register_enabled as boolean | undefined) }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('health.turnstile')">
                {{ boolLabel(status.turnstile_check as boolean | undefined) }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('health.setup')">
                {{ boolLabel(status.setup as boolean | undefined) }}
              </NDescriptionsItem>
            </NDescriptions>
            <span v-else class="muted">{{ t('health.unknown') }}</span>
          </NCard>
        </NGi>
      </NGrid>
    </NSpin>
  </div>
</template>

<style scoped>
.title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
}
.probe-row {
  display: grid;
  grid-template-columns: 140px auto 1fr;
  gap: 8px 12px;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.probe-name {
  font-family: ui-monospace, monospace;
  font-size: 13px;
}
.probe-meta {
  grid-column: 1 / -1;
  font-size: 12px;
  opacity: 0.7;
}
.probe-body {
  grid-column: 1 / -1;
  font-size: 11px;
  max-height: 80px;
  overflow: auto;
}
.muted {
  opacity: 0.6;
}
</style>

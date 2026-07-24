<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
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
    <div class="page-header">
      <h2 class="title">{{ t('health.title') }}</h2>
      <a-button type="primary" :loading="loading" @click="refresh">
        {{ t('health.refresh') }}
      </a-button>
    </div>

    <a-alert
      v-if="error"
      type="warning"
      show-icon
      :message="t('common.error')"
      :description="error"
      style="margin-bottom: 16px"
    />

    <a-spin :spinning="loading">
      <a-row :gutter="[16, 16]">
        <a-col :xs="24" :md="12">
          <a-card :title="t('health.probes')" size="small">
            <div v-for="p in probes" :key="p.name" class="probe-row">
              <span class="probe-name">{{ p.name }}</span>
              <a-tag :color="p.ok ? 'green' : 'red'" size="small">
                {{ p.ok ? t('common.ok') : t('common.down') }}
              </a-tag>
              <span class="probe-meta">
                HTTP {{ p.status ?? '—' }}
                <template v-if="p.error"> · {{ p.error }}</template>
              </span>
              <pre
                v-if="p.body && typeof p.body === 'object'"
                class="probe-body"
              >{{ JSON.stringify(p.body, null, 2) }}</pre>
            </div>
            <span v-if="!probes.length" class="muted">{{ t('common.loading') }}</span>
          </a-card>
        </a-col>
        <a-col :xs="24" :md="12">
          <a-card :title="t('health.status')" size="small">
            <a-descriptions v-if="status" :column="1" size="small" bordered>
              <a-descriptions-item :label="t('health.version')">
                {{ status.version ?? t('health.unknown') }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('health.systemName')">
                {{ status.system_name ?? t('health.unknown') }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('health.passwordLogin')">
                {{ boolLabel(status.password_login_enabled as boolean | undefined) }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('health.register')">
                {{ boolLabel(status.register_enabled as boolean | undefined) }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('health.turnstile')">
                {{ boolLabel(status.turnstile_check as boolean | undefined) }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('health.setup')">
                {{ boolLabel(status.setup as boolean | undefined) }}
              </a-descriptions-item>
            </a-descriptions>
            <span v-else class="muted">{{ t('health.unknown') }}</span>
          </a-card>
        </a-col>
      </a-row>
    </a-spin>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
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
  border-bottom: 1px solid #f5f5f5;
}
.probe-row:last-child {
  border-bottom: none;
}
.probe-name {
  font-family: ui-monospace, monospace;
  font-size: 13px;
}
.probe-meta {
  grid-column: 1 / -1;
  font-size: 12px;
  color: #a3a3a3;
}
.probe-body {
  grid-column: 1 / -1;
  font-size: 11px;
  max-height: 80px;
  overflow: auto;
  margin: 0;
  padding: 8px;
  background: #fafafa;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
}
.muted {
  color: #a3a3a3;
}
</style>
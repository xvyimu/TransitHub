<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ADMIN_ROLE, listLogs } from '@/api/logs'
import { apiMessage, isApiSuccess } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import type { LogItem } from '@/types/api'

const { t } = useI18n()
const auth = useAuthStore()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<LogItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const typeFilter = ref<string>('all')
const modelName = ref('')
const username = ref('')
const requestId = ref('')

const isAdmin = computed(() => (auth.user?.role ?? 0) >= ADMIN_ROLE)

const typeOptions = computed(() => [
  { label: t('logs.typeAll'), value: 'all' },
  { label: t('logs.typeUnknown'), value: '0' },
  { label: t('logs.typeTopup'), value: '1' },
  { label: t('logs.typeConsume'), value: '2' },
  { label: t('logs.typeManage'), value: '3' },
  { label: t('logs.typeSystem'), value: '4' },
  { label: t('logs.typeError'), value: '5' },
  { label: t('logs.typeRefund'), value: '6' },
  { label: t('logs.typeLogin'), value: '7' },
])

function typeLabel(type: number | undefined) {
  switch (type) {
    case 1: return t('logs.typeTopup')
    case 2: return t('logs.typeConsume')
    case 3: return t('logs.typeManage')
    case 4: return t('logs.typeSystem')
    case 5: return t('logs.typeError')
    case 6: return t('logs.typeRefund')
    case 7: return t('logs.typeLogin')
    case 0: return t('logs.typeUnknown')
    default: return t('health.unknown')
  }
}

function typeTagColor(type: number | undefined) {
  if (type === 2) return 'blue'
  if (type === 1 || type === 6) return 'green'
  if (type === 5) return 'red'
  if (type === 3 || type === 4) return 'orange'
  return 'default'
}

function formatTime(ts: number | undefined) {
  if (!ts || ts <= 0) return t('health.unknown')
  const ms = ts > 1e12 ? ts : ts * 1000
  try {
    return new Date(ms).toLocaleString()
  } catch {
    return String(ts)
  }
}

function truncateId(id: string | undefined, max = 12) {
  if (!id) return t('health.unknown')
  if (id.length <= max) return id
  return `${id.slice(0, max)}…`
}

function hRequestId(id: string | undefined) {
  if (!id) return t('health.unknown')
  return h(
    'a-tooltip',
    { title: id },
    { default: () => h('span', { style: 'cursor: default; font-size: 12px; font-family: monospace' }, truncateId(id, 14)) },
  )
}

const columns = computed(() => [
  {
    title: t('logs.colTime'),
    dataIndex: 'created_at',
    key: 'created_at',
    width: 160,
    customRender: ({ text }: { text: number | undefined }) => formatTime(text),
  },
  {
    title: t('logs.colType'),
    dataIndex: 'type',
    key: 'type',
    width: 100,
    customRender: ({ text }: { text: number | undefined }) =>
      h('a-tag', { color: typeTagColor(text), size: 'small' }, { default: () => typeLabel(text) }),
  },
  {
    title: t('logs.colUsername'),
    dataIndex: 'username',
    key: 'username',
    width: 110,
    ellipsis: true,
  },
  {
    title: t('logs.colModel'),
    dataIndex: 'model_name',
    key: 'model_name',
    ellipsis: true,
  },
  {
    title: t('logs.colQuota'),
    dataIndex: 'quota',
    key: 'quota',
    width: 90,
  },
  {
    title: t('logs.colTokens'),
    key: 'tokens',
    width: 120,
    customRender: ({ record }: { record: LogItem }) => {
      const p = record.prompt_tokens ?? 0
      const c = record.completion_tokens ?? 0
      return `${p} / ${c}`
    },
  },
  {
    title: t('logs.colChannel'),
    key: 'channel',
    width: 120,
    ellipsis: true,
    customRender: ({ record }: { record: LogItem }) => {
      if (record.channel_name) return record.channel_name
      if (record.channel) return String(record.channel)
      return t('health.unknown')
    },
  },
  {
    title: t('logs.colRequestId'),
    dataIndex: 'request_id',
    key: 'request_id',
    width: 140,
    customRender: ({ text }: { text: string | undefined }) => hRequestId(text),
  },
])

function normalizeListBody(body: unknown): { items: LogItem[]; total: number } {
  if (!body || typeof body !== 'object') return { items: [], total: 0 }
  const b = body as Record<string, unknown>
  const data = (b.data && typeof b.data === 'object' ? b.data : b) as Record<string, unknown>
  const rawItems = data.items
  const list = Array.isArray(rawItems) ? (rawItems as LogItem[]) : []
  const tot = typeof data.total === 'number' ? data.total : list.length
  return { items: list, total: tot }
}

let refreshSeq = 0

async function refresh() {
  const seq = ++refreshSeq
  loading.value = true
  error.value = null
  try {
    const body = await listLogs({
      p: page.value,
      page_size: pageSize.value,
      type: typeFilter.value === 'all' ? undefined : typeFilter.value,
      model_name: modelName.value,
      username: isAdmin.value ? username.value : undefined,
      request_id: requestId.value,
      isAdmin: isAdmin.value === true,
    })
    if (seq !== refreshSeq) return
    if (!isApiSuccess(body)) {
      error.value = body.message || 'list failed'
      items.value = []
      total.value = 0
      return
    }
    const { items: list, total: tot } = normalizeListBody(body)
    items.value = list
    total.value = tot
  } catch (e) {
    if (seq !== refreshSeq) return
    error.value = apiMessage(e)
    items.value = []
    total.value = 0
  } finally {
    if (seq === refreshSeq) loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  void refresh()
}

function onSearch() {
  page.value = 1
  void refresh()
}

watch(typeFilter, () => {
  page.value = 1
  void refresh()
})

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="logs">
    <div class="page-header">
      <div>
        <h2 class="title">{{ t('logs.title') }}</h2>
        <span class="hint">{{ t('logs.readonlyHint') }}</span>
      </div>
      <a-button type="primary" :loading="loading" @click="refresh">{{ t('health.refresh') }}</a-button>
    </div>

    <div class="toolbar">
      <a-select v-model:value="typeFilter" :options="typeOptions" style="width: 140px" />
      <a-input
        v-model:value="modelName"
        allow-clear
        :placeholder="t('logs.modelPlaceholder')"
        style="width: 180px"
        @keyup.enter="onSearch"
      />
      <a-input
        v-if="isAdmin"
        v-model:value="username"
        allow-clear
        :placeholder="t('logs.usernamePlaceholder')"
        style="width: 140px"
        @keyup.enter="onSearch"
      />
      <a-input
        v-model:value="requestId"
        allow-clear
        :placeholder="t('logs.requestIdPlaceholder')"
        style="width: 200px"
        @keyup.enter="onSearch"
      />
      <a-button @click="onSearch">{{ t('logs.search') }}</a-button>
    </div>

    <a-alert
      v-if="error"
      type="error"
      show-icon
      :message="t('common.error')"
      :description="error"
      style="margin-bottom: 12px"
    />

    <a-table
      :columns="columns"
      :data-source="items"
      :loading="loading"
      :pagination="{
        current: page,
        pageSize,
        total,
        showSizeChanger: false,
        onChange: onPageChange,
      }"
      size="small"
      :row-key="(record: LogItem) => record.id"
    />
  </div>
</template>

<style scoped>
.logs {
  max-width: 1200px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}
.title {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 600;
}
.hint {
  font-size: 12px;
  color: #a3a3a3;
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
</style>
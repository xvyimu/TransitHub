<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NAlert,
  NButton,
  NDataTable,
  NInput,
  NSelect,
  NSpace,
  NTag,
  NText,
  NTooltip,
  type DataTableColumns,
  type SelectOption,
} from 'naive-ui'
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

const typeOptions = computed<SelectOption[]>(() => [
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
    case 1:
      return t('logs.typeTopup')
    case 2:
      return t('logs.typeConsume')
    case 3:
      return t('logs.typeManage')
    case 4:
      return t('logs.typeSystem')
    case 5:
      return t('logs.typeError')
    case 6:
      return t('logs.typeRefund')
    case 7:
      return t('logs.typeLogin')
    case 0:
      return t('logs.typeUnknown')
    default:
      return t('health.unknown')
  }
}

function typeTagType(type: number | undefined) {
  if (type === 2) return 'info' as const
  if (type === 1 || type === 6) return 'success' as const
  if (type === 5) return 'error' as const
  if (type === 3 || type === 4) return 'warning' as const
  return 'default' as const
}

function formatTime(ts: number | undefined) {
  if (!ts || ts <= 0) return t('health.unknown')
  // Backend stores seconds; tolerate ms if ever passed.
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
  if (!id) return h(NText, { depth: 3 }, { default: () => t('health.unknown') })
  return h(
    NTooltip,
    { trigger: 'hover' },
    {
      trigger: () =>
        h(
          NText,
          { code: true, style: 'cursor: default; font-size: 12px' },
          { default: () => truncateId(id, 14) },
        ),
      default: () => id,
    },
  )
}

const columns = computed<DataTableColumns<LogItem>>(() => [
  {
    title: t('logs.colTime'),
    key: 'created_at',
    width: 160,
    render: (row) => formatTime(row.created_at),
  },
  {
    title: t('logs.colType'),
    key: 'type',
    width: 100,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: typeTagType(row.type), bordered: false },
        { default: () => typeLabel(row.type) },
      ),
  },
  {
    title: t('logs.colUsername'),
    key: 'username',
    width: 110,
    ellipsis: { tooltip: true },
  },
  {
    title: t('logs.colModel'),
    key: 'model_name',
    ellipsis: { tooltip: true },
  },
  {
    title: t('logs.colQuota'),
    key: 'quota',
    width: 90,
  },
  {
    title: t('logs.colTokens'),
    key: 'tokens',
    width: 120,
    render: (row) => {
      const p = row.prompt_tokens ?? 0
      const c = row.completion_tokens ?? 0
      return `${p} / ${c}`
    },
  },
  {
    title: t('logs.colChannel'),
    key: 'channel',
    width: 120,
    ellipsis: { tooltip: true },
    render: (row) => {
      if (row.channel_name) return row.channel_name
      if (row.channel) return String(row.channel)
      return t('health.unknown')
    },
  },
  {
    title: t('logs.colRequestId'),
    key: 'request_id',
    width: 140,
    render: (row) => hRequestId(row.request_id),
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

async function refresh() {
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
      isAdmin: isAdmin.value,
    })
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
    error.value = apiMessage(e)
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
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
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <div>
        <h2 class="title">{{ t('logs.title') }}</h2>
        <NText depth="3" style="font-size: 12px">{{ t('logs.readonlyHint') }}</NText>
      </div>
      <NButton type="primary" :loading="loading" @click="refresh">{{ t('health.refresh') }}</NButton>
    </NSpace>

    <NSpace style="margin-bottom: 12px" wrap>
      <NSelect v-model:value="typeFilter" :options="typeOptions" style="width: 140px" />
      <NInput
        v-model:value="modelName"
        clearable
        :placeholder="t('logs.modelPlaceholder')"
        style="width: 180px"
        @keyup.enter="onSearch"
      />
      <NInput
        v-if="isAdmin"
        v-model:value="username"
        clearable
        :placeholder="t('logs.usernamePlaceholder')"
        style="width: 140px"
        @keyup.enter="onSearch"
      />
      <NInput
        v-model:value="requestId"
        clearable
        :placeholder="t('logs.requestIdPlaceholder')"
        style="width: 200px"
        @keyup.enter="onSearch"
      />
      <NButton @click="onSearch">{{ t('logs.search') }}</NButton>
    </NSpace>

    <NAlert v-if="error" type="error" style="margin-bottom: 12px" :title="t('common.error')">
      {{ error }}
    </NAlert>

    <NDataTable
      :columns="columns"
      :data="items"
      :loading="loading"
      :bordered="false"
      :single-line="false"
      size="small"
      :pagination="{
        page,
        pageSize,
        itemCount: total,
        showSizePicker: false,
        onChange: onPageChange,
      }"
    />
  </div>
</template>

<style scoped>
.logs {
  max-width: 1200px;
}
.title {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 600;
}
</style>

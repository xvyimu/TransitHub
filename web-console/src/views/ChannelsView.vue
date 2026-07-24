<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { listChannels } from '@/api/channels'
import { apiMessage, isApiSuccess } from '@/api/http'
import type { ChannelItem } from '@/types/api'

const { t } = useI18n()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<ChannelItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const statusFilter = ref<string>('all')

const statusOptions = computed(() => [
  { label: t('channels.statusAll'), value: 'all' },
  { label: t('channels.statusEnabled'), value: '1' },
  { label: t('channels.statusDisabled'), value: '0' },
])

function statusColor(s: number | undefined) {
  if (s === 1) return 'green'
  if (s === 2) return 'orange'
  return 'default'
}

function statusLabel(s: number | undefined) {
  if (s === 1) return t('channels.statusEnabled')
  if (s === 2) return t('channels.statusAutoDisabled')
  if (s === 0) return t('channels.statusDisabled')
  return t('health.unknown')
}

function hTag(status: number | undefined) {
  return h(
    'a-tag',
    { color: statusColor(status), size: 'small' },
    { default: () => statusLabel(status) },
  )
}

const columns = computed(() => [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 72 },
  {
    title: t('channels.colName'),
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: t('channels.colType'),
    dataIndex: 'type',
    key: 'type',
    width: 80,
  },
  {
    title: t('channels.colStatus'),
    dataIndex: 'status',
    key: 'status',
    width: 110,
    customRender: ({ text }: { text: number | undefined }) => hTag(text),
  },
  {
    title: t('channels.colGroup'),
    dataIndex: 'group',
    key: 'group',
    ellipsis: true,
    width: 120,
  },
  {
    title: t('channels.colTag'),
    dataIndex: 'tag',
    key: 'tag',
    ellipsis: true,
    width: 100,
  },
  {
    title: t('channels.colPriority'),
    dataIndex: 'priority',
    key: 'priority',
    width: 80,
  },
])

function normalizeListBody(body: unknown): { items: ChannelItem[]; total: number } {
  if (!body || typeof body !== 'object') return { items: [], total: 0 }
  const b = body as Record<string, unknown>
  const data = (b.data && typeof b.data === 'object' ? b.data : b) as Record<string, unknown>
  const rawItems = data.items
  const list = Array.isArray(rawItems) ? (rawItems as ChannelItem[]) : []
  const tot = typeof data.total === 'number' ? data.total : list.length
  return { items: list, total: tot }
}

async function refresh() {
  loading.value = true
  error.value = null
  try {
    const body = await listChannels({
      p: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
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

watch(statusFilter, () => {
  page.value = 1
  void refresh()
})

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="channels">
    <div class="page-header">
      <div>
        <h2 class="title">{{ t('channels.title') }}</h2>
        <span class="hint">{{ t('channels.readonlyHint') }}</span>
      </div>
      <a-button type="primary" :loading="loading" @click="refresh">{{ t('health.refresh') }}</a-button>
    </div>

    <div class="toolbar">
      <a-input
        v-model:value="keyword"
        allow-clear
        :placeholder="t('channels.searchPlaceholder')"
        style="width: 240px"
        @keyup.enter="onSearch"
      />
      <a-select v-model:value="statusFilter" :options="statusOptions" style="width: 140px" />
      <a-button @click="onSearch">{{ t('channels.search') }}</a-button>
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
      :row-key="(record: ChannelItem) => record.id"
    />
  </div>
</template>

<style scoped>
.channels {
  max-width: 1100px;
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
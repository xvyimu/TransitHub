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
  type DataTableColumns,
} from 'naive-ui'
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

function statusType(s: number | undefined) {
  if (s === 1) return 'success' as const
  if (s === 2) return 'warning' as const
  return 'default' as const
}

function statusLabel(s: number | undefined) {
  if (s === 1) return t('channels.statusEnabled')
  if (s === 2) return t('channels.statusAutoDisabled')
  if (s === 0) return t('channels.statusDisabled')
  return t('health.unknown')
}

function hTag(status: number | undefined) {
  return h(
    NTag,
    { size: 'small', type: statusType(status), bordered: false },
    { default: () => statusLabel(status) },
  )
}

const columns = computed<DataTableColumns<ChannelItem>>(() => [
  { title: 'ID', key: 'id', width: 72 },
  {
    title: t('channels.colName'),
    key: 'name',
    ellipsis: { tooltip: true },
  },
  {
    title: t('channels.colType'),
    key: 'type',
    width: 80,
  },
  {
    title: t('channels.colStatus'),
    key: 'status',
    width: 110,
    render: (row) => hTag(row.status),
  },
  {
    title: t('channels.colGroup'),
    key: 'group',
    ellipsis: { tooltip: true },
    width: 120,
  },
  {
    title: t('channels.colTag'),
    key: 'tag',
    ellipsis: { tooltip: true },
    width: 100,
  },
  {
    title: t('channels.colPriority'),
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
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <div>
        <h2 class="title">{{ t('channels.title') }}</h2>
        <NText depth="3" style="font-size: 12px">{{ t('channels.readonlyHint') }}</NText>
      </div>
      <NButton type="primary" :loading="loading" @click="refresh">{{ t('health.refresh') }}</NButton>
    </NSpace>

    <NSpace style="margin-bottom: 12px" wrap>
      <NInput
        v-model:value="keyword"
        clearable
        :placeholder="t('channels.searchPlaceholder')"
        style="width: 240px"
        @keyup.enter="onSearch"
      />
      <NSelect v-model:value="statusFilter" :options="statusOptions" style="width: 140px" />
      <NButton @click="onSearch">{{ t('channels.search') }}</NButton>
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
.channels {
  max-width: 1100px;
}
.title {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 600;
}
</style>

<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { listModels } from '@/api/models'
import { apiMessage, isApiSuccess } from '@/api/http'
import type { ModelItem } from '@/types/api'

const { t } = useI18n()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<ModelItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const statusFilter = ref<string>('all')

const statusOptions = computed(() => [
  { label: t('models.statusAll'), value: 'all' },
  { label: t('models.statusEnabled'), value: '1' },
  { label: t('models.statusDisabled'), value: '0' },
])

function statusColor(s: number | undefined) {
  if (s === 1) return 'green'
  return 'default'
}

function statusLabel(s: number | undefined) {
  if (s === 1) return t('models.statusEnabled')
  if (s === 0) return t('models.statusDisabled')
  return t('health.unknown')
}

function nameRuleLabel(rule: number | undefined) {
  switch (rule) {
    case 0: return t('models.nameRuleExact')
    case 1: return t('models.nameRulePrefix')
    case 2: return t('models.nameRuleContains')
    case 3: return t('models.nameRuleSuffix')
    default: return t('health.unknown')
  }
}

function hStatus(status: number | undefined) {
  return h('a-tag', { color: statusColor(status), size: 'small' }, { default: () => statusLabel(status) })
}

function boundChannelsText(row: ModelItem) {
  const chs = row.bound_channels
  if (!chs || chs.length === 0) return t('health.unknown')
  return chs.map((c) => c.name || String(c.type)).join(', ')
}

function groupsText(row: ModelItem) {
  const g = row.enable_groups
  if (!g || g.length === 0) return t('health.unknown')
  return g.join(', ')
}

const columns = computed(() => [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 72 },
  {
    title: t('models.colName'),
    dataIndex: 'model_name',
    key: 'model_name',
    ellipsis: true,
  },
  {
    title: t('models.colStatus'),
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: number | undefined }) => hStatus(text),
  },
  {
    title: t('models.colVendor'),
    dataIndex: 'vendor_id',
    key: 'vendor_id',
    width: 90,
    customRender: ({ text }: { text: number | undefined | null }) =>
      text !== undefined && text !== null ? String(text) : t('health.unknown'),
  },
  {
    title: t('models.colNameRule'),
    dataIndex: 'name_rule',
    key: 'name_rule',
    width: 100,
    customRender: ({ text }: { text: number | undefined }) => nameRuleLabel(text),
  },
  {
    title: t('models.colTags'),
    dataIndex: 'tags',
    key: 'tags',
    width: 140,
    ellipsis: true,
    customRender: ({ text }: { text: string | undefined }) => text || t('health.unknown'),
  },
  {
    title: t('models.colGroups'),
    key: 'enable_groups',
    width: 140,
    ellipsis: true,
    customRender: ({ record }: { record: ModelItem }) => groupsText(record),
  },
  {
    title: t('models.colBoundChannels'),
    key: 'bound_channels',
    ellipsis: true,
    customRender: ({ record }: { record: ModelItem }) => boundChannelsText(record),
  },
])

function normalizeListBody(body: unknown): { items: ModelItem[]; total: number } {
  if (!body || typeof body !== 'object') return { items: [], total: 0 }
  const b = body as Record<string, unknown>
  const data = (b.data && typeof b.data === 'object' ? b.data : b) as Record<string, unknown>
  const rawItems = data.items
  const list = Array.isArray(rawItems) ? (rawItems as ModelItem[]) : []
  const tot = typeof data.total === 'number' ? data.total : list.length
  return { items: list, total: tot }
}

let refreshSeq = 0

async function refresh() {
  const seq = ++refreshSeq
  loading.value = true
  error.value = null
  try {
    const body = await listModels({
      p: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
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

watch(statusFilter, () => {
  page.value = 1
  void refresh()
})

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="models">
    <div class="page-header">
      <div>
        <h2 class="title">{{ t('models.title') }}</h2>
        <span class="hint">{{ t('models.readonlyHint') }}</span>
      </div>
      <a-button type="primary" :loading="loading" @click="refresh">{{ t('health.refresh') }}</a-button>
    </div>

    <div class="toolbar">
      <a-input
        v-model:value="keyword"
        allow-clear
        :placeholder="t('models.searchPlaceholder')"
        style="width: 240px"
        @keyup.enter="onSearch"
      />
      <a-select v-model:value="statusFilter" :options="statusOptions" style="width: 140px" />
      <a-button @click="onSearch">{{ t('models.search') }}</a-button>
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
      :row-key="(record: ModelItem) => record.id"
    />
  </div>
</template>

<style scoped>
.models {
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
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

function statusType(s: number | undefined) {
  if (s === 1) return 'success' as const
  return 'default' as const
}

function statusLabel(s: number | undefined) {
  if (s === 1) return t('models.statusEnabled')
  if (s === 0) return t('models.statusDisabled')
  return t('health.unknown')
}

function nameRuleLabel(rule: number | undefined) {
  switch (rule) {
    case 0:
      return t('models.nameRuleExact')
    case 1:
      return t('models.nameRulePrefix')
    case 2:
      return t('models.nameRuleContains')
    case 3:
      return t('models.nameRuleSuffix')
    default:
      return t('health.unknown')
  }
}

function hStatus(status: number | undefined) {
  return h(
    NTag,
    { size: 'small', type: statusType(status), bordered: false },
    { default: () => statusLabel(status) },
  )
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

const columns = computed<DataTableColumns<ModelItem>>(() => [
  { title: 'ID', key: 'id', width: 72 },
  {
    title: t('models.colName'),
    key: 'model_name',
    ellipsis: { tooltip: true },
  },
  {
    title: t('models.colStatus'),
    key: 'status',
    width: 100,
    render: (row) => hStatus(row.status),
  },
  {
    title: t('models.colVendor'),
    key: 'vendor_id',
    width: 90,
    render: (row) =>
      row.vendor_id !== undefined && row.vendor_id !== null
        ? String(row.vendor_id)
        : t('health.unknown'),
  },
  {
    title: t('models.colNameRule'),
    key: 'name_rule',
    width: 100,
    render: (row) => nameRuleLabel(row.name_rule),
  },
  {
    title: t('models.colTags'),
    key: 'tags',
    width: 140,
    ellipsis: { tooltip: true },
    render: (row) => row.tags || t('health.unknown'),
  },
  {
    title: t('models.colGroups'),
    key: 'enable_groups',
    width: 140,
    ellipsis: { tooltip: true },
    render: (row) => groupsText(row),
  },
  {
    title: t('models.colBoundChannels'),
    key: 'bound_channels',
    ellipsis: { tooltip: true },
    render: (row) => boundChannelsText(row),
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

/** Drop stale responses when filters/pages change rapidly. */
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
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <div>
        <h2 class="title">{{ t('models.title') }}</h2>
        <NText depth="3" style="font-size: 12px">{{ t('models.readonlyHint') }}</NText>
      </div>
      <NButton type="primary" :loading="loading" @click="refresh">{{ t('health.refresh') }}</NButton>
    </NSpace>

    <NSpace style="margin-bottom: 12px" wrap>
      <NInput
        v-model:value="keyword"
        clearable
        :placeholder="t('models.searchPlaceholder')"
        style="width: 240px"
        @keyup.enter="onSearch"
      />
      <NSelect v-model:value="statusFilter" :options="statusOptions" style="width: 140px" />
      <NButton @click="onSearch">{{ t('models.search') }}</NButton>
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
.models {
  max-width: 1200px;
}
.title {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 600;
}
</style>

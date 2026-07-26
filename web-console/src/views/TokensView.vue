<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { Tag, message as antMessage } from 'antdv-next'
import { useI18n } from 'vue-i18n'
import {
  listTokens,
  createToken,
  updateToken,
  deleteToken,
  setTokenStatus,
  getTokenKey,
} from '@/api/tokens'
import { apiMessage, isApiSuccess } from '@/api/http'
import {
  TOKEN_STATUS_ENABLED,
  TOKEN_STATUS_DISABLED,
  TOKEN_STATUS_EXPIRED,
  TOKEN_STATUS_EXHAUSTED,
  blankTokenForm,
  tokenToForm,
  tokenFormToPayload,
  validateTokenForm,
  type TokenFormModel,
} from '@/api/tokenForm'
import type { TokenItem } from '@/types/api'

const { t } = useI18n()

// ── list state ──
const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<TokenItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')

// ── modal state ──
const showForm = ref(false)
const formSaving = ref(false)
const formError = ref<string | null>(null)
const formModel = ref<TokenFormModel>(blankTokenForm())
const editingId = ref<number | null>(null)

// ── reveal-key state ──
const showKeyModal = ref(false)
const revealedKey = ref('')

const isEditing = computed(() => editingId.value != null)

function statusColor(s: number | undefined) {
  if (s === TOKEN_STATUS_ENABLED) return 'green'
  if (s === TOKEN_STATUS_DISABLED) return 'default'
  if (s === TOKEN_STATUS_EXPIRED) return 'orange'
  if (s === TOKEN_STATUS_EXHAUSTED) return 'red'
  return 'default'
}

function statusLabel(s: number | undefined) {
  if (s === TOKEN_STATUS_ENABLED) return t('tokens.statusEnabled')
  if (s === TOKEN_STATUS_DISABLED) return t('tokens.statusDisabled')
  if (s === TOKEN_STATUS_EXPIRED) return t('tokens.statusExpired')
  if (s === TOKEN_STATUS_EXHAUSTED) return t('tokens.statusExhausted')
  return t('health.unknown')
}

function hTag(status: number | undefined) {
  return h(Tag, { color: statusColor(status) }, { default: () => statusLabel(status) })
}

/** Only manual enabled/disabled tokens can be toggled from the list. */
function statusCanToggle(s: number | undefined) {
  return s === TOKEN_STATUS_ENABLED || s === TOKEN_STATUS_DISABLED
}

const columns = computed(() => [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 64 },
  { title: t('tokens.colName'), dataIndex: 'name', key: 'name', ellipsis: true },
  { title: t('tokens.colKey'), key: 'key', width: 170 },
  { title: t('tokens.colStatus'), dataIndex: 'status', key: 'status', width: 110 },
  { title: t('tokens.colQuota'), key: 'quota', width: 150 },
  { title: t('tokens.colGroup'), dataIndex: 'group', key: 'group', width: 100, ellipsis: true },
  { title: t('tokens.colActions'), key: 'actions', width: 280 },
])

function quotaText(record: TokenItem) {
  if (record.unlimited_quota) return t('tokens.quotaUnlimited')
  return `${record.remain_quota ?? 0} / ${record.used_quota ?? 0}`
}

function normalizeListBody(body: unknown): { items: TokenItem[]; total: number } {
  if (!body || typeof body !== 'object') return { items: [], total: 0 }
  const b = body as Record<string, unknown>
  const data = (b.data && typeof b.data === 'object' ? b.data : b) as Record<string, unknown>
  const rawItems = data.items
  const list = Array.isArray(rawItems) ? (rawItems as TokenItem[]) : []
  const tot = typeof data.total === 'number' ? data.total : list.length
  return { items: list, total: tot }
}

let refreshSeq = 0

async function refresh() {
  const seq = ++refreshSeq
  loading.value = true
  error.value = null
  try {
    const body = await listTokens({
      p: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
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

// ── create / edit ──
function openCreate() {
  editingId.value = null
  formModel.value = blankTokenForm()
  formError.value = null
  showForm.value = true
}

function openEdit(record: TokenItem) {
  editingId.value = record.id
  formModel.value = tokenToForm(record)
  formError.value = null
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  formError.value = null
  formSaving.value = false
}

async function saveForm() {
  const err = validateTokenForm(formModel.value)
  if (err) {
    formError.value = t(`tokens.err${err.charAt(0).toUpperCase()}${err.slice(1)}`)
    return
  }
  formError.value = null
  formSaving.value = true
  try {
    const payload = tokenFormToPayload(formModel.value)
    const body = isEditing.value ? await updateToken(payload) : await createToken(payload)
    if (!isApiSuccess(body)) {
      formError.value = body.message || t('tokens.saveFailed')
      return
    }
    antMessage.success(isEditing.value ? t('tokens.updated') : t('tokens.created'))
    closeForm()
    void refresh()
  } catch (e) {
    formError.value = apiMessage(e, t('tokens.saveFailed'))
  } finally {
    formSaving.value = false
  }
}

async function toggleStatus(record: TokenItem) {
  const next =
    record.status === TOKEN_STATUS_ENABLED ? TOKEN_STATUS_DISABLED : TOKEN_STATUS_ENABLED
  try {
    const body = await setTokenStatus(record.id, next)
    if (!isApiSuccess(body)) {
      antMessage.error(body.message || t('tokens.statusFailed'))
      return
    }
    antMessage.success(t('tokens.statusUpdated'))
    void refresh()
  } catch (e) {
    antMessage.error(apiMessage(e, t('tokens.statusFailed')))
  }
}

async function doDelete(record: TokenItem) {
  try {
    const body = await deleteToken(record.id)
    if (!isApiSuccess(body)) {
      antMessage.error(body.message || t('tokens.deleteFailed'))
      return
    }
    antMessage.success(t('tokens.deleted'))
    // Deleting the last row on a trailing page: step back so we don't land empty.
    if (page.value > 1 && items.value.length === 1) {
      page.value = page.value - 1
    }
    void refresh()
  } catch (e) {
    antMessage.error(apiMessage(e, t('tokens.deleteFailed')))
  }
}

// ── reveal full key ──
async function revealKey(record: TokenItem) {
  try {
    const body = await getTokenKey(record.id)
    if (!isApiSuccess(body)) {
      antMessage.error(body.message || t('tokens.keyFetchFailed'))
      return
    }
    const data = (body.data ?? {}) as Record<string, unknown>
    revealedKey.value = typeof data.key === 'string' ? data.key : ''
    showKeyModal.value = true
  } catch (e) {
    antMessage.error(apiMessage(e, t('tokens.keyFetchFailed')))
  }
}

watch(keyword, (v, old) => {
  // Clearing the box should reset to the full list.
  if (v === '' && old !== '') {
    page.value = 1
    void refresh()
  }
})

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="tokens">
    <div class="page-header">
      <div>
        <h2 class="title">{{ t('tokens.title') }}</h2>
        <span class="hint">{{ t('tokens.subtitle') }}</span>
      </div>
      <a-space>
        <a-button :loading="loading" @click="refresh">{{ t('health.refresh') }}</a-button>
        <a-button type="primary" @click="openCreate">{{ t('tokens.create') }}</a-button>
      </a-space>
    </div>

    <div class="toolbar">
      <a-input
        v-model:value="keyword"
        allow-clear
        :placeholder="t('tokens.searchPlaceholder')"
        style="width: 240px"
        @keyup.enter="onSearch"
      />
      <a-button @click="onSearch">{{ t('tokens.search') }}</a-button>
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
      :row-key="(record: TokenItem) => record.id"
      :scroll="{ x: 960 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'key'">
          <span class="masked-key">{{ (record as TokenItem).key || t('tokens.keyHidden') }}</span>
        </template>
        <template v-else-if="column.key === 'status'">
          <component :is="hTag((record as TokenItem).status)" />
        </template>
        <template v-else-if="column.key === 'quota'">
          {{ quotaText(record as TokenItem) }}
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-space :size="4">
            <a-button size="small" @click="revealKey(record as TokenItem)">
              {{ t('tokens.viewKey') }}
            </a-button>
            <a-button
              v-if="statusCanToggle((record as TokenItem).status)"
              size="small"
              @click="toggleStatus(record as TokenItem)"
            >
              {{
                (record as TokenItem).status === TOKEN_STATUS_ENABLED
                  ? t('tokens.disable')
                  : t('tokens.enable')
              }}
            </a-button>
            <a-button size="small" @click="openEdit(record as TokenItem)">
              {{ t('tokens.edit') }}
            </a-button>
            <a-popconfirm
              :title="t('tokens.deleteConfirm')"
              :ok-text="t('tokens.delete')"
              :cancel-text="t('tokens.cancel')"
              @confirm="doDelete(record as TokenItem)"
            >
              <a-button size="small" danger>{{ t('tokens.delete') }}</a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- Create / edit modal -->
    <a-modal
      v-model:open="showForm"
      :title="isEditing ? t('tokens.editTitle') : t('tokens.createTitle')"
      :confirm-loading="formSaving"
      :ok-text="isEditing ? t('tokens.save') : t('tokens.create')"
      :cancel-text="t('tokens.cancel')"
      @ok="saveForm"
      @cancel="closeForm"
    >
      <a-form layout="vertical">
        <a-form-item :label="t('tokens.formName')" required>
          <a-input
            v-model:value="formModel.name"
            :maxlength="50"
            :placeholder="t('tokens.formNamePlaceholder')"
          />
        </a-form-item>

        <a-form-item :label="t('tokens.formGroup')">
          <a-input
            v-model:value="formModel.group"
            :placeholder="t('tokens.formGroupPlaceholder')"
          />
        </a-form-item>

        <a-form-item :label="t('tokens.formUnlimited')">
          <a-switch v-model:checked="formModel.unlimitedQuota" />
        </a-form-item>

        <a-form-item v-if="!formModel.unlimitedQuota" :label="t('tokens.formQuota')">
          <a-input-number
            v-model:value="formModel.remainQuota"
            :min="0"
            style="width: 100%"
            :placeholder="t('tokens.formQuotaPlaceholder')"
          />
        </a-form-item>

        <a-form-item :label="t('tokens.formNeverExpires')">
          <a-switch v-model:checked="formModel.neverExpires" />
        </a-form-item>

        <a-form-item v-if="!formModel.neverExpires" :label="t('tokens.formExpiresAt')">
          <input
            v-model="formModel.expiresAt"
            type="datetime-local"
            class="dt-input"
          />
        </a-form-item>

        <a-form-item :label="t('tokens.formAllowIps')">
          <a-textarea
            v-model:value="formModel.allowIps"
            :rows="2"
            :placeholder="t('tokens.formAllowIpsPlaceholder')"
          />
        </a-form-item>
      </a-form>

      <a-alert
        v-if="formError"
        type="error"
        show-icon
        :message="formError"
        style="margin-top: 4px"
      />
    </a-modal>

    <!-- Reveal-key modal -->
    <a-modal
      v-model:open="showKeyModal"
      :title="t('tokens.keyTitle')"
      :footer="null"
      :width="520"
    >
      <div class="key-display">
        <a-typography-paragraph
          v-if="revealedKey"
          code
          copyable
          :content="revealedKey"
        />
        <span v-else class="hint">{{ t('tokens.keyEmpty') }}</span>
      </div>
      <a-alert type="warning" show-icon :message="t('tokens.keyWarning')" />
    </a-modal>
  </div>
</template>

<style scoped>
.tokens {
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
.masked-key {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
}
.dt-input {
  width: 100%;
  padding: 4px 11px;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.5;
}
.dt-input:focus {
  outline: none;
  border-color: #1677ff;
}
.key-display {
  margin-bottom: 12px;
}
</style>

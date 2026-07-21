/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import type { ChannelHealthMetrics } from '../types'

export const CHANNEL_HEALTH_METRICS_QUERY_KEY = [
  'channel-health-metrics',
  'channels-page',
] as const

export type ChannelFailureReasonParts = {
  openCircuit: boolean
  consecutiveFailure?: number
  lastStatus?: number
  lastModel?: string
  lastError?: string
}

export type ChannelFailureTopError = {
  channel_id: number
  count: number
  last_status?: number
  last_model?: string
  last_at_unix?: number
  reasonParts: ChannelFailureReasonParts
}

export type ChannelFailureOpenCircuit = {
  channel_id: number
  consecutive_failure: number
  last_error: string
}

export type ChannelCallSignal = 'normal' | 'abnormal' | 'unknown'

export type ChannelFailureViewModel = {
  isColdStart: boolean
  relayOk: number
  relayFail: number
  openCircuits: ChannelFailureOpenCircuit[]
  topErrors: ChannelFailureTopError[]
  /** channel_id → recent error count from top_error_channels */
  errorCountByChannel: Record<number, number>
  /** channel_id → structured failure reason (localize in UI) */
  reasonPartsByChannel: Record<number, ChannelFailureReasonParts>
  /** channel_ids currently in open circuit state */
  openCircuitChannelIds: number[]
}

export type ChannelErrorLogsSearch = {
  channel: string
  type: ['5']
}

const SECRETISH =
  /(sk-[a-zA-Z0-9_-]{8,}|Bearer\s+\S+|api[_-]?key\s*[:=]\s*\S+|eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]+\.)/i

/** Truncate and scrub credential-like fragments from upstream error text. */
export function sanitizeFailureReason(raw: string | null | undefined): string {
  if (!raw) return ''
  let text = raw.replace(/\s+/g, ' ').trim()
  text = text.replace(SECRETISH, '[redacted]')
  if (text.length > 120) {
    text = `${text.slice(0, 117)}...`
  }
  return text
}

export function buildChannelFailureReasonParts(input: {
  last_status?: number
  last_model?: string
  circuit_last_error?: string
  open_circuit?: boolean
  consecutive_failure?: number
}): ChannelFailureReasonParts {
  return {
    openCircuit: Boolean(input.open_circuit),
    consecutiveFailure:
      input.consecutive_failure && input.consecutive_failure > 0
        ? input.consecutive_failure
        : undefined,
    lastStatus:
      typeof input.last_status === 'number' && input.last_status > 0
        ? input.last_status
        : undefined,
    lastModel: input.last_model?.trim() || undefined,
    lastError: sanitizeFailureReason(input.circuit_last_error) || undefined,
  }
}

export function hasFailureReasonParts(
  parts: ChannelFailureReasonParts | undefined
): boolean {
  if (!parts) return false
  return Boolean(
    parts.openCircuit ||
      parts.lastStatus ||
      parts.lastModel ||
      parts.lastError
  )
}

/**
 * Localize structured failure parts. Pass i18n `t` from react-i18next.
 */
export function formatChannelFailureReasonLocalized(
  parts: ChannelFailureReasonParts | undefined,
  t: (key: string, options?: Record<string, unknown>) => string
): string {
  if (!parts || !hasFailureReasonParts(parts)) return ''
  const bits: string[] = []
  if (parts.openCircuit) {
    bits.push(t('Circuit open'))
    if (parts.consecutiveFailure) {
      bits.push(t('Consecutive failures: {{count}}', {
        count: parts.consecutiveFailure,
      }))
    }
  }
  if (parts.lastStatus) {
    bits.push(t('HTTP status {{code}}', { code: parts.lastStatus }))
  }
  if (parts.lastModel) {
    bits.push(t('Model {{name}}', { name: parts.lastModel }))
  }
  if (parts.lastError) {
    bits.push(parts.lastError)
  }
  return bits.join(' · ')
}

/**
 * Build a presentational view-model from in-process health metrics.
 * Does not include secrets; channel_id only.
 */
export function buildChannelFailureViewModel(
  metrics: ChannelHealthMetrics | null | undefined,
  options: { topErrorLimit?: number } = {}
): ChannelFailureViewModel {
  const topErrorLimit = options.topErrorLimit ?? 5
  if (!metrics) {
    return {
      isColdStart: false,
      relayOk: 0,
      relayFail: 0,
      openCircuits: [],
      topErrors: [],
      errorCountByChannel: {},
      reasonPartsByChannel: {},
      openCircuitChannelIds: [],
    }
  }

  const relayOk = metrics.relay_success ?? 0
  const relayFail = metrics.relay_fail ?? 0
  const topSource = metrics.top_error_channels ?? []

  const openCircuits = (metrics.circuits ?? [])
    .filter((c) => c && c.state === 'open')
    .map((c) => ({
      channel_id: c.channel_id,
      consecutive_failure: c.consecutive_failure ?? 0,
      last_error: sanitizeFailureReason(c.last_error),
    }))

  const openById = new Map(
    openCircuits.map((c) => [c.channel_id, c] as const)
  )
  const openCircuitChannelIds = openCircuits.map((c) => c.channel_id)

  const errorCountByChannel: Record<number, number> = {}
  const reasonPartsByChannel: Record<number, ChannelFailureReasonParts> = {}

  for (const e of topSource) {
    if (!e || typeof e.channel_id !== 'number' || !(e.count > 0)) continue
    errorCountByChannel[e.channel_id] = e.count
    const open = openById.get(e.channel_id)
    reasonPartsByChannel[e.channel_id] = buildChannelFailureReasonParts({
      last_status: e.last_status,
      last_model: e.last_model,
      open_circuit: Boolean(open),
      consecutive_failure: open?.consecutive_failure,
      circuit_last_error: open?.last_error,
    })
  }

  for (const c of openCircuits) {
    if (!reasonPartsByChannel[c.channel_id]) {
      reasonPartsByChannel[c.channel_id] = buildChannelFailureReasonParts({
        open_circuit: true,
        consecutive_failure: c.consecutive_failure,
        circuit_last_error: c.last_error,
      })
    }
  }

  const topErrors = topSource
    .filter((e) => e && typeof e.channel_id === 'number' && e.count > 0)
    .slice(0, topErrorLimit)
    .map((e) => ({
      channel_id: e.channel_id,
      count: e.count,
      last_status: e.last_status,
      last_model: e.last_model,
      last_at_unix: e.last_at_unix,
      reasonParts:
        reasonPartsByChannel[e.channel_id] ||
        buildChannelFailureReasonParts({
          last_status: e.last_status,
          last_model: e.last_model,
        }),
    }))

  const isColdStart =
    relayOk === 0 &&
    relayFail === 0 &&
    openCircuits.length === 0 &&
    topErrors.length === 0 &&
    (metrics.shadow?.samples ?? 0) === 0

  return {
    isColdStart,
    relayOk,
    relayFail,
    openCircuits,
    topErrors,
    errorCountByChannel,
    reasonPartsByChannel,
    openCircuitChannelIds,
  }
}

export function channelErrorLogsSearch(
  channelId: number
): ChannelErrorLogsSearch {
  return {
    channel: String(channelId),
    type: ['5'],
  }
}

export function channelHasFailureSignal(
  channelId: number,
  vm: ChannelFailureViewModel
): boolean {
  return (
    (vm.errorCountByChannel[channelId] ?? 0) > 0 ||
    vm.openCircuitChannelIds.includes(channelId)
  )
}

/**
 * Per-row call health badge for the current process window.
 * - abnormal: appears in top errors or open circuit
 * - normal: metrics loaded, not cold start, and no failure signal
 * - unknown: metrics missing or cold start (no evidence yet)
 */
export function channelCallSignal(
  channelId: number,
  vm: ChannelFailureViewModel,
  options: { metricsLoaded?: boolean } = {}
): ChannelCallSignal {
  if (options.metricsLoaded === false) return 'unknown'
  if (vm.isColdStart) return 'unknown'
  if (channelHasFailureSignal(channelId, vm)) return 'abnormal'
  if (vm.relayOk + vm.relayFail > 0) return 'normal'
  return 'unknown'
}

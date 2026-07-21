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
import assert from 'node:assert/strict'
import { describe, it } from 'vitest'

import type { ChannelHealthMetrics } from '../types'
import {
  buildChannelFailureReasonParts,
  buildChannelFailureViewModel,
  channelCallSignal,
  channelErrorLogsSearch,
  channelHasFailureSignal,
  formatChannelFailureReasonLocalized,
  hasFailureReasonParts,
  sanitizeFailureReason,
} from './channel-failure-visibility'

function baseMetrics(
  partial: Partial<ChannelHealthMetrics> = {}
): ChannelHealthMetrics {
  return {
    generated_at: 1,
    relay_success: 0,
    relay_fail: 0,
    retry_histogram: {},
    top_error_channels: [],
    circuits: [],
    shadow: { samples: 0, agree: 0, agree_rate: 0 },
    refund_intents: {},
    ...partial,
  }
}

const tIdentity = (key: string, options?: Record<string, unknown>) => {
  if (key === 'HTTP status {{code}}') return `HTTP status ${options?.code}`
  if (key === 'Model {{name}}') return `Model ${options?.name}`
  if (key === 'Consecutive failures: {{count}}')
    return `Consecutive failures: ${options?.count}`
  return key
}

describe('buildChannelFailureViewModel', () => {
  it('returns empty non-cold model when metrics missing', () => {
    const vm = buildChannelFailureViewModel(undefined)
    assert.equal(vm.isColdStart, false)
    assert.equal(vm.topErrors.length, 0)
    assert.equal(vm.openCircuits.length, 0)
  })

  it('marks cold start when all in-process counters are zero', () => {
    const vm = buildChannelFailureViewModel(baseMetrics())
    assert.equal(vm.isColdStart, true)
    assert.equal(vm.relayOk, 0)
    assert.equal(vm.relayFail, 0)
  })

  it('is not cold start when relay traffic exists', () => {
    const vm = buildChannelFailureViewModel(
      baseMetrics({
        relay_success: 11,
        relay_fail: 4,
        shadow: { samples: 11, agree: 3, agree_rate: 0.27 },
      })
    )
    assert.equal(vm.isColdStart, false)
    assert.equal(vm.relayOk, 11)
    assert.equal(vm.relayFail, 4)
  })

  it('truncates top errors and maps counts + reasons', () => {
    const top = Array.from({ length: 8 }, (_, i) => ({
      channel_id: i + 1,
      count: 10 - i,
      last_status: 402,
      last_model: 'glm-5',
    }))
    const vm = buildChannelFailureViewModel(
      baseMetrics({
        relay_fail: 1,
        shadow: { samples: 1, agree: 0, agree_rate: 0 },
        top_error_channels: top,
      }),
      { topErrorLimit: 5 }
    )
    assert.equal(vm.topErrors.length, 5)
    assert.equal(vm.topErrors[0]?.channel_id, 1)
    assert.equal(vm.errorCountByChannel[3], 8)
    assert.equal(vm.reasonPartsByChannel[1]?.lastStatus, 402)
    assert.equal(vm.topErrors[0]?.reasonParts.lastModel, 'glm-5')
    assert.equal(channelHasFailureSignal(1, vm), true)
    assert.equal(channelHasFailureSignal(99, vm), false)
    assert.equal(channelCallSignal(1, vm, { metricsLoaded: true }), 'abnormal')
    assert.equal(channelCallSignal(99, vm, { metricsLoaded: true }), 'normal')
  })

  it('keeps only open circuits and exposes circuit reason', () => {
    const vm = buildChannelFailureViewModel(
      baseMetrics({
        relay_fail: 1,
        shadow: { samples: 1, agree: 0, agree_rate: 0 },
        circuits: [
          {
            channel_id: 84,
            state: 'open',
            consecutive_failure: 3,
            open_until_unix: 0,
            last_error: 'upstream balance empty',
          },
          {
            channel_id: 2,
            state: 'closed',
            consecutive_failure: 0,
            open_until_unix: 0,
            last_error: '',
          },
        ],
      })
    )
    assert.equal(vm.openCircuits.length, 1)
    assert.equal(vm.openCircuits[0]?.channel_id, 84)
    assert.deepEqual(vm.openCircuitChannelIds, [84])
    assert.equal(channelHasFailureSignal(84, vm), true)
    assert.equal(vm.reasonPartsByChannel[84]?.openCircuit, true)
    assert.equal(
      vm.reasonPartsByChannel[84]?.lastError,
      'upstream balance empty'
    )
  })
})

describe('channelErrorLogsSearch', () => {
  it('deep-links error logs for a channel id', () => {
    assert.deepEqual(channelErrorLogsSearch(84), {
      channel: '84',
      type: ['5'],
    })
  })
})

describe('sanitizeFailureReason', () => {
  it('redacts secret-like fragments', () => {
    const out = sanitizeFailureReason('auth failed sk-abc123456789 token')
    assert.doesNotMatch(out, /sk-abc/)
    assert.match(out, /redacted/)
  })
})

describe('formatChannelFailureReasonLocalized', () => {
  it('joins status model and circuit state via t()', () => {
    const parts = buildChannelFailureReasonParts({
      open_circuit: true,
      consecutive_failure: 3,
      last_status: 429,
      last_model: 'deepseek',
      circuit_last_error: 'rate limited',
    })
    assert.equal(hasFailureReasonParts(parts), true)
    const reason = formatChannelFailureReasonLocalized(parts, tIdentity)
    assert.match(reason, /Circuit open/)
    assert.match(reason, /HTTP status 429/)
    assert.match(reason, /Model deepseek/)
    assert.match(reason, /rate limited/)
  })
})

describe('channelCallSignal', () => {
  it('returns unknown on cold start', () => {
    const vm = buildChannelFailureViewModel(baseMetrics())
    assert.equal(channelCallSignal(1, vm, { metricsLoaded: true }), 'unknown')
  })
})

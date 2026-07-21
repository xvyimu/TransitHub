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
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { AlertTriangle, CheckCircle2, RadioTower } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import {
  ADMIN_PERMISSION_ACTIONS,
  ADMIN_PERMISSION_RESOURCES,
  hasPermission,
} from '@/lib/admin-permissions'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { getChannelHealthMetrics } from '../api'
import {
  CHANNEL_HEALTH_METRICS_QUERY_KEY,
  buildChannelFailureViewModel,
  channelErrorLogsSearch,
  formatChannelFailureReasonLocalized,
  hasFailureReasonParts,
} from '../lib/channel-failure-visibility'

/**
 * Channels-page strip: surface in-process failure signals + deep links.
 * Reuses health-metrics; never renders secrets.
 */
export function ChannelFailureStrip({ className }: { className?: string }) {
  const { t } = useTranslation()
  const currentUser = useAuthStore((s) => s.auth.user)
  const canViewOps = hasPermission(
    currentUser,
    ADMIN_PERMISSION_RESOURCES.CHANNEL,
    ADMIN_PERMISSION_ACTIONS.OPERATE
  )

  const healthQuery = useQuery({
    queryKey: CHANNEL_HEALTH_METRICS_QUERY_KEY,
    queryFn: getChannelHealthMetrics,
    enabled: canViewOps,
    staleTime: 30_000,
    retry: false,
  })

  const vm = useMemo(
    () => buildChannelFailureViewModel(healthQuery.data?.data),
    [healthQuery.data?.data]
  )

  if (!canViewOps) return null

  const loaded = healthQuery.isSuccess
  const failed = healthQuery.isError
  const hasTraffic = vm.relayOk + vm.relayFail > 0

  return (
    <section
      className={cn(
        'bg-muted/30 mb-3 rounded-2xl border p-3 shadow-xs',
        className
      )}
      aria-label={t('Channel call health')}
    >
      <div className='mb-2 flex flex-wrap items-center gap-2'>
        <AlertTriangle
          className='text-warning size-4 shrink-0'
          aria-hidden
        />
        <h2 className='text-sm font-semibold'>{t('Channel call health')}</h2>
        <span className='text-muted-foreground text-xs'>
          {t('In-process metrics')}
        </span>
        {loaded && vm.isColdStart ? (
          <span className='text-muted-foreground border-border rounded-md border border-dashed px-1.5 py-0.5 text-[11px]'>
            {t('Cold start')}
          </span>
        ) : null}
      </div>

      {failed ? (
        <p className='text-muted-foreground text-xs'>
          {t('Could not load channel health metrics')}
        </p>
      ) : null}

      {loaded && vm.isColdStart ? (
        <p className='text-muted-foreground mb-2 text-xs leading-relaxed'>
          {t(
            'Metrics accumulate since process start; zeros after deploy are expected'
          )}
        </p>
      ) : null}

      {loaded && !vm.isColdStart ? (
        <div className='flex flex-wrap items-center gap-2 text-xs'>
          <span
            className={cn(
              'inline-flex items-center gap-1 rounded-md border px-2 py-0.5 font-medium',
              hasTraffic && vm.relayOk > 0
                ? 'border-success/30 bg-success/10 text-success'
                : 'border-border text-muted-foreground'
            )}
            title={t('Successful relay calls in this process')}
          >
            <CheckCircle2 className='size-3.5' aria-hidden />
            {t('Normal calls')}: {vm.relayOk}
          </span>
          <span
            className={cn(
              'inline-flex items-center gap-1 rounded-md border px-2 py-0.5 font-medium',
              vm.relayFail > 0
                ? 'border-destructive/30 bg-destructive/10 text-destructive'
                : 'border-border text-muted-foreground'
            )}
            title={t('Failed relay calls in this process')}
          >
            <AlertTriangle className='size-3.5' aria-hidden />
            {t('Abnormal calls')}: {vm.relayFail}
          </span>
          {vm.openCircuits.length > 0 ? (
            <span className='text-warning border-warning/30 bg-warning/10 inline-flex items-center gap-1 rounded-md border px-2 py-0.5 font-medium'>
              {t('Open circuits')}: {vm.openCircuits.length}
            </span>
          ) : (
            <span className='text-muted-foreground border-border rounded-md border px-2 py-0.5'>
              {t('Open circuits')}: 0
            </span>
          )}
        </div>
      ) : null}

      {loaded && vm.topErrors.length > 0 ? (
        <ul className='mt-2 flex flex-col gap-1.5'>
          {vm.topErrors.map((e) => {
            const reason = formatChannelFailureReasonLocalized(
              e.reasonParts,
              t
            )
            return (
              <li key={e.channel_id}>
                <Link
                  to='/usage-logs/$section'
                  params={{ section: 'common' }}
                  search={channelErrorLogsSearch(e.channel_id)}
                  className='border-border bg-background hover:bg-muted flex flex-wrap items-center gap-x-2 gap-y-1 rounded-md border px-2 py-1.5 text-xs'
                >
                  <span className='text-destructive inline-flex items-center gap-1 font-medium'>
                    <RadioTower className='size-3.5' aria-hidden />
                    #{e.channel_id}
                  </span>
                  <span className='text-muted-foreground'>
                    {t('Failures')}: {e.count}
                  </span>
                  {hasFailureReasonParts(e.reasonParts) && reason ? (
                    <span
                      className='text-foreground/90 min-w-0 flex-1 truncate'
                      title={reason}
                    >
                      {t('Failure reason')}: {reason}
                    </span>
                  ) : (
                    <span className='text-muted-foreground'>
                      {t('Failure reason unavailable')}
                    </span>
                  )}
                </Link>
              </li>
            )
          })}
        </ul>
      ) : null}

      {loaded &&
      !vm.isColdStart &&
      vm.topErrors.length === 0 &&
      vm.openCircuits.length === 0 ? (
        <p className='text-muted-foreground mt-2 text-xs'>
          {t('No top error channels in the current process window')}
        </p>
      ) : null}
    </section>
  )
}

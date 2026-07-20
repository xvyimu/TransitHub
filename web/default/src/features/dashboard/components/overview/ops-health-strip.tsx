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
import {
  Activity,
  AlertTriangle,
  GitMerge,
  RadioTower,
  ScrollText,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  getChannelOps,
  getChannelHealthMetrics,
  getDuplicateChannels,
} from '@/features/channels/api'
import {
  ADMIN_PERMISSION_ACTIONS,
  ADMIN_PERMISSION_RESOURCES,
  hasPermission,
} from '@/lib/admin-permissions'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

/**
 * Compact ops strip for LOCAL-ONLY gateway admins: channel health shortcuts,
 * duplicate groups, and deep links into channels / error logs.
 */
export function OpsHealthStrip() {
  const { t } = useTranslation()
  const currentUser = useAuthStore((s) => s.auth.user)
  const canViewOps = hasPermission(
    currentUser,
    ADMIN_PERMISSION_RESOURCES.CHANNEL,
    ADMIN_PERMISSION_ACTIONS.OPERATE
  )

  const opsQuery = useQuery({
    queryKey: ['channel-ops', 'overview-strip'],
    queryFn: getChannelOps,
    enabled: canViewOps,
    staleTime: 60_000,
    retry: false,
  })

  const dupQuery = useQuery({
    queryKey: ['channel-duplicates', 'overview-strip'],
    queryFn: getDuplicateChannels,
    enabled: canViewOps,
    staleTime: 60_000,
    retry: false,
  })

  const healthQuery = useQuery({
    queryKey: ['channel-health-metrics', 'overview-strip'],
    queryFn: getChannelHealthMetrics,
    enabled: canViewOps,
    staleTime: 30_000,
    retry: false,
  })

  if (!canViewOps) return null

  const retryTimes = opsQuery.data?.data?.retry_times
  const dupCount = dupQuery.data?.data?.groups?.length ?? 0
  const health = healthQuery.data?.data
  const openCircuits =
    health?.circuits?.filter((c) => c.state === 'open').length ?? 0
  const topErr = health?.top_error_channels?.[0]
  const shadowRate = health?.shadow?.agree_rate
  const relayFail = health?.relay_fail ?? 0
  const relayOk = health?.relay_success ?? 0

  return (
    <section
      className='bg-muted/30 rounded-2xl border p-4 shadow-xs'
      aria-label={t('Ops health')}
    >
      <div className='mb-3 flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-2'>
          <Activity className='text-primary size-4' aria-hidden />
          <h2 className='text-sm font-semibold'>{t('Ops health')}</h2>
          <span className='text-muted-foreground text-xs'>
            {t('Local gateway')}
          </span>
        </div>
        <div className='text-muted-foreground flex flex-wrap items-center gap-3 text-xs'>
          {typeof retryTimes === 'number' && (
            <span>
              {t('Max Retries')}: {retryTimes}
            </span>
          )}
          {(relayOk > 0 || relayFail > 0) && (
            <span>
              {t('Relay')}: {relayOk}/{relayFail}
            </span>
          )}
          {openCircuits > 0 && (
            <span className='text-warning font-medium'>
              {t('Open circuits')}: {openCircuits}
            </span>
          )}
          {typeof shadowRate === 'number' && health?.shadow?.samples ? (
            <span>
              {t('Shadow agree')}: {(shadowRate * 100).toFixed(0)}%
            </span>
          ) : null}
          {topErr ? (
            <span>
              {t('Top error ch')}: #{topErr.channel_id} ({topErr.count})
            </span>
          ) : null}
          {dupCount > 0 && (
            <span className='text-warning flex items-center gap-1 font-medium'>
              <GitMerge className='size-3.5' />
              {t('{{count}} duplicate group(s)', { count: dupCount })}
            </span>
          )}
        </div>
      </div>

      <div className='grid gap-2 sm:grid-cols-3'>
        <Button
          variant='outline'
          className={cn('h-auto justify-start rounded-xl px-3 py-3 text-left')}
          render={<Link to='/channels' />}
        >
          <span className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
            <RadioTower className='size-4' />
          </span>
          <span className='flex min-w-0 flex-col gap-0.5'>
            <span className='text-sm font-medium'>{t('Channels')}</span>
            <span className='text-muted-foreground line-clamp-1 text-xs'>
              {t('Manage routes')}
            </span>
          </span>
        </Button>

        <Button
          variant='outline'
          className={cn('h-auto justify-start rounded-xl px-3 py-3 text-left')}
          render={
            <Link to='/channels' search={{ status: ['disabled'] }} />
          }
        >
          <span className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
            <AlertTriangle className='size-4' />
          </span>
          <span className='flex min-w-0 flex-col gap-0.5'>
            <span className='text-sm font-medium'>{t('Problems')}</span>
            <span className='text-muted-foreground line-clamp-1 text-xs'>
              {t('Disabled / auto-disabled')}
            </span>
          </span>
        </Button>

        <Button
          variant='outline'
          className={cn('h-auto justify-start rounded-xl px-3 py-3 text-left')}
          render={
            <Link
              to='/usage-logs/$section'
              params={{ section: 'common' }}
              search={{ type: ['5'] }}
            />
          }
        >
          <span className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
            <ScrollText className='size-4' />
          </span>
          <span className='flex min-w-0 flex-col gap-0.5'>
            <span className='text-sm font-medium'>{t('Error logs')}</span>
            <span className='text-muted-foreground line-clamp-1 text-xs'>
              {t('Recent failures')}
            </span>
          </span>
        </Button>
      </div>
    </section>
  )
}

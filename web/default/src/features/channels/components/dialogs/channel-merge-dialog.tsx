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
import { useQueryClient } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import {
  fetchChannelMergePreview,
  fetchDuplicateChannelGroups,
  handleMergeChannels,
} from '../../lib'
import type {
  ChannelMergePreview,
  DuplicateChannelGroup,
} from '../../types'

type ChannelMergeDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** When set, skip discovery and preview/merge these ids directly. */
  selectedIds?: number[]
}

export function ChannelMergeDialog({
  open,
  onOpenChange,
  selectedIds,
}: ChannelMergeDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [loading, setLoading] = useState(false)
  const [merging, setMerging] = useState(false)
  const [groups, setGroups] = useState<DuplicateChannelGroup[]>([])
  const [selectedGroupKey, setSelectedGroupKey] = useState<string>('')
  const [primaryId, setPrimaryId] = useState<number>(0)
  const [preview, setPreview] = useState<ChannelMergePreview | null>(null)
  const [error, setError] = useState<string>('')

  const isSelectionMode = (selectedIds?.length ?? 0) >= 2

  const activeIds = useMemo(() => {
    if (isSelectionMode) {
      return selectedIds ?? []
    }
    const group = groups.find((g) => g.group_key === selectedGroupKey)
    return group?.channels.map((c) => c.id) ?? []
  }, [isSelectionMode, selectedIds, groups, selectedGroupKey])

  const loadDiscovery = useCallback(async () => {
    setLoading(true)
    setError('')
    setPreview(null)
    try {
      const list = await fetchDuplicateChannelGroups()
      setGroups(list)
      if (list.length > 0) {
        setSelectedGroupKey(list[0].group_key)
        setPrimaryId(list[0].suggested_primary_id)
      } else {
        setSelectedGroupKey('')
        setPrimaryId(0)
      }
    } catch (e) {
      const msg =
        e instanceof Error
          ? e.message
          : t('Failed to load duplicate channels')
      setError(msg)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }, [t])

  const loadPreview = useCallback(
    async (ids: number[], primary: number) => {
      if (ids.length < 2) {
        setPreview(null)
        return
      }
      setLoading(true)
      setError('')
      try {
        const p = await fetchChannelMergePreview({
          ids,
          primary_id: primary || undefined,
        })
        setPreview(p)
        setPrimaryId(p.primary_id)
      } catch (e) {
        const msg =
          e instanceof Error
            ? e.message
            : t('Failed to preview channel merge')
        setError(msg)
        setPreview(null)
      } finally {
        setLoading(false)
      }
    },
    [t]
  )

  useEffect(() => {
    if (!open) {
      setGroups([])
      setSelectedGroupKey('')
      setPrimaryId(0)
      setPreview(null)
      setError('')
      setMerging(false)
      return
    }
    if (isSelectionMode) {
      void loadPreview(selectedIds ?? [], 0)
    } else {
      void loadDiscovery()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, isSelectionMode])

  useEffect(() => {
    if (!open || isSelectionMode) return
    if (!selectedGroupKey) return
    const group = groups.find((g) => g.group_key === selectedGroupKey)
    if (!group) return
    const ids = group.channels.map((c) => c.id)
    const primary = primaryId || group.suggested_primary_id
    void loadPreview(ids, primary)
    // only re-preview when group changes; primary changes handled below
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedGroupKey, groups, open, isSelectionMode])

  const handlePrimaryChange = (value: string | null) => {
    if (!value) return
    const id = Number(value)
    setPrimaryId(id)
    if (activeIds.length >= 2) {
      void loadPreview(activeIds, id)
    }
  }

  const handleGroupChange = (value: string | null) => {
    if (!value) return
    setSelectedGroupKey(value)
    const group = groups.find((g) => g.group_key === value)
    if (group) {
      setPrimaryId(group.suggested_primary_id)
    }
  }

  const handleConfirm = async () => {
    if (activeIds.length < 2) return
    setMerging(true)
    await handleMergeChannels(
      { ids: activeIds, primary_id: primaryId || undefined },
      queryClient,
      () => {
        onOpenChange(false)
      }
    )
    setMerging(false)
  }

  const primaryOptions = preview?.channels ?? []

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Merge Duplicate Channels')}
      description={
        isSelectionMode
          ? t(
              'Merge the selected channels into one multi-key channel. Others will be permanently deleted.'
            )
          : t(
              'Find channels with the same name and host, then merge their keys into one multi-key channel.'
            )
      }
      contentHeight='auto'
      bodyClassName='space-y-4'
      footer={
        <>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={merging}
          >
            {t('Cancel')}
          </Button>
          <Button
            variant='destructive'
            onClick={handleConfirm}
            disabled={merging || loading || !preview || activeIds.length < 2}
          >
            {merging && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {merging ? t('Merging...') : t('Confirm Merge')}
          </Button>
        </>
      }
    >
      <div className='space-y-4 py-2'>
        {loading && !preview && (
          <div className='text-muted-foreground flex items-center gap-2 text-sm'>
            <Loader2 className='h-4 w-4 animate-spin' />
            {t('Loading...')}
          </div>
        )}

        {error && (
          <p className='text-destructive text-sm' role='alert'>
            {error}
          </p>
        )}

        {!isSelectionMode && !loading && groups.length === 0 && !error && (
          <p className='text-muted-foreground text-sm'>
            {t('No duplicate channel groups found.')}
          </p>
        )}

        {!isSelectionMode && groups.length > 0 && (
          <div className='space-y-2'>
            <Label>{t('Duplicate group')}</Label>
            <Select value={selectedGroupKey} onValueChange={handleGroupChange}>
              <SelectTrigger>
                <SelectValue placeholder={t('Select a group')} />
              </SelectTrigger>
              <SelectContent>
                {groups.map((g) => (
                  <SelectItem key={g.group_key} value={g.group_key}>
                    {g.name} @ {g.host} ({g.count})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {preview && (
          <>
            <div className='space-y-2'>
              <Label>{t('Primary channel')}</Label>
              <Select
                value={String(primaryId || preview.primary_id)}
                onValueChange={handlePrimaryChange}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {primaryOptions.map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      #{c.id} · {c.name} · keys={c.key_count} · prio=
                      {c.priority} · w={c.weight}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='bg-muted/40 space-y-1 rounded-md border p-3 text-sm'>
              <p>
                <strong>{t('Host')}:</strong> {preview.host}
              </p>
              <p>
                <strong>{t('Merged keys')}:</strong> {preview.merged_key_count}
              </p>
              <p>
                <strong>{t('Models')}:</strong> {preview.models_count}
              </p>
              <p>
                <strong>{t('Groups')}:</strong> {preview.groups || '-'}
              </p>
              <p>
                <strong>{t('Priority / Weight')}:</strong> {preview.priority} /{' '}
                {preview.weight}
              </p>
              <p>
                <strong>{t('Will delete')}:</strong>{' '}
                {preview.delete_ids.map((id) => `#${id}`).join(', ') || '-'}
              </p>
            </div>

            <p className='text-muted-foreground text-xs'>
              {t(
                'This will merge {{count}} channels into #{{primary}} as multi-key and permanently delete the rest.',
                {
                  count: preview.channels.length,
                  primary: preview.primary_id,
                }
              )}
            </p>
          </>
        )}
      </div>
    </Dialog>
  )
}

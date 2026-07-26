/**
 * Pure helpers for token create/update — extracted for unit tests.
 * Mirrors backend invariants in controller/token.go (AddToken/UpdateToken):
 *  - name length <= 50
 *  - when not unlimited: remain_quota in [0, maxQuota]
 *  - expired_time: -1 means never expires, otherwise unix seconds
 * Runtime marshaling lives in tokens.ts.
 */
import type { TokenItem } from '@/types/api'

export const TOKEN_NAME_MAX = 50

/** common.QuotaPerUnit (500000) * 1e9 — controller/token.go maxQuotaValue. */
export const TOKEN_QUOTA_PER_UNIT = 500000
export const TOKEN_MAX_QUOTA = 1000000000 * TOKEN_QUOTA_PER_UNIT

/** common.TokenStatus* — 0 is never used by the backend. */
export const TOKEN_STATUS_ENABLED = 1
export const TOKEN_STATUS_DISABLED = 2
export const TOKEN_STATUS_EXPIRED = 3
export const TOKEN_STATUS_EXHAUSTED = 4

export const NEVER_EXPIRES = -1

export interface TokenFormModel {
  id?: number
  name: string
  unlimitedQuota: boolean
  /** Raw quota units (as stored); ignored when unlimitedQuota. */
  remainQuota: number
  neverExpires: boolean
  /** datetime-local string (YYYY-MM-DDTHH:mm); ignored when neverExpires. */
  expiresAt: string
  group: string
  allowIps: string
}

/** Payload shape sent to POST/PUT /api/token/ (subset of model.Token). */
export interface TokenPayload {
  id?: number
  name: string
  unlimited_quota: boolean
  remain_quota: number
  expired_time: number
  group: string
  allow_ips: string
}

/**
 * datetime-local string -> unix seconds. Returns NaN for blank/invalid input
 * so validateTokenForm can reject it. Interpreted in the browser's local zone,
 * matching how the value is produced by unixToDatetimeLocal.
 */
export function datetimeLocalToUnix(s: string): number {
  const trimmed = (s || '').trim()
  if (!trimmed) return NaN
  const ms = new Date(trimmed).getTime()
  if (!Number.isFinite(ms)) return NaN
  return Math.floor(ms / 1000)
}

/** unix seconds -> datetime-local string in local zone; '' when never/invalid. */
export function unixToDatetimeLocal(sec: number): string {
  if (!Number.isFinite(sec) || sec <= 0) return ''
  const d = new Date(sec * 1000)
  if (!Number.isFinite(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}` +
    `T${pad(d.getHours())}:${pad(d.getMinutes())}`
  )
}

export function blankTokenForm(): TokenFormModel {
  return {
    name: '',
    unlimitedQuota: false,
    remainQuota: 0,
    neverExpires: true,
    expiresAt: '',
    group: '',
    allowIps: '',
  }
}

/** Build an editable form model from an existing token row. */
export function tokenToForm(t: TokenItem): TokenFormModel {
  const expired = typeof t.expired_time === 'number' ? t.expired_time : NEVER_EXPIRES
  const never = expired === NEVER_EXPIRES
  return {
    id: t.id,
    name: t.name ?? '',
    unlimitedQuota: t.unlimited_quota === true,
    remainQuota: typeof t.remain_quota === 'number' ? t.remain_quota : 0,
    neverExpires: never,
    expiresAt: never ? '' : unixToDatetimeLocal(expired),
    group: t.group ?? '',
    allowIps: typeof t.allow_ips === 'string' ? t.allow_ips : '',
  }
}

export type TokenFormError =
  | 'nameRequired'
  | 'nameTooLong'
  | 'quotaNegative'
  | 'quotaExceedMax'
  | 'expiryInvalid'

/** Returns the first invariant violation, or null when the form is valid. */
export function validateTokenForm(f: TokenFormModel): TokenFormError | null {
  const name = f.name.trim()
  if (!name) return 'nameRequired'
  if (name.length > TOKEN_NAME_MAX) return 'nameTooLong'
  if (!f.unlimitedQuota) {
    if (!Number.isFinite(f.remainQuota) || f.remainQuota < 0) return 'quotaNegative'
    if (f.remainQuota > TOKEN_MAX_QUOTA) return 'quotaExceedMax'
  }
  if (!f.neverExpires) {
    if (!Number.isFinite(datetimeLocalToUnix(f.expiresAt))) return 'expiryInvalid'
  }
  return null
}

/** Normalize a validated form into the backend payload. Call validate first. */
export function tokenFormToPayload(f: TokenFormModel): TokenPayload {
  const payload: TokenPayload = {
    name: f.name.trim(),
    unlimited_quota: f.unlimitedQuota,
    remain_quota: f.unlimitedQuota ? 0 : Math.floor(f.remainQuota),
    expired_time: f.neverExpires ? NEVER_EXPIRES : datetimeLocalToUnix(f.expiresAt),
    group: f.group.trim(),
    allow_ips: f.allowIps,
  }
  if (f.id != null) payload.id = f.id
  return payload
}

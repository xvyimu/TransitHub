import { describe, expect, it } from 'vitest'
import {
  NEVER_EXPIRES,
  TOKEN_MAX_QUOTA,
  blankTokenForm,
  datetimeLocalToUnix,
  tokenFormToPayload,
  tokenToForm,
  unixToDatetimeLocal,
  validateTokenForm,
  type TokenFormModel,
} from './tokenForm'

function form(overrides: Partial<TokenFormModel> = {}): TokenFormModel {
  return { ...blankTokenForm(), name: 'ok', ...overrides }
}

describe('validateTokenForm', () => {
  it('accepts a minimal valid form', () => {
    expect(validateTokenForm(form())).toBeNull()
  })

  it('requires a non-empty name', () => {
    expect(validateTokenForm(form({ name: '   ' }))).toBe('nameRequired')
  })

  it('rejects names longer than 50 (backend MsgTokenNameTooLong)', () => {
    expect(validateTokenForm(form({ name: 'a'.repeat(51) }))).toBe('nameTooLong')
    expect(validateTokenForm(form({ name: 'a'.repeat(50) }))).toBeNull()
  })

  it('rejects negative quota when not unlimited', () => {
    expect(validateTokenForm(form({ remainQuota: -1 }))).toBe('quotaNegative')
  })

  it('rejects quota above the backend max', () => {
    expect(validateTokenForm(form({ remainQuota: TOKEN_MAX_QUOTA + 1 }))).toBe(
      'quotaExceedMax',
    )
    expect(validateTokenForm(form({ remainQuota: TOKEN_MAX_QUOTA }))).toBeNull()
  })

  it('ignores quota bounds when unlimited', () => {
    expect(
      validateTokenForm(form({ unlimitedQuota: true, remainQuota: -999 })),
    ).toBeNull()
  })

  it('rejects a blank or invalid expiry when not never-expires', () => {
    expect(validateTokenForm(form({ neverExpires: false, expiresAt: '' }))).toBe(
      'expiryInvalid',
    )
    expect(
      validateTokenForm(form({ neverExpires: false, expiresAt: 'not-a-date' })),
    ).toBe('expiryInvalid')
  })

  it('accepts a valid expiry when not never-expires', () => {
    expect(
      validateTokenForm(form({ neverExpires: false, expiresAt: '2030-01-01T00:00' })),
    ).toBeNull()
  })
})

describe('tokenFormToPayload', () => {
  it('maps camelCase form to snake_case backend fields', () => {
    const p = tokenFormToPayload(
      form({ name: '  key  ', remainQuota: 500, group: ' vip ', allowIps: '1.2.3.4' }),
    )
    expect(p).toMatchObject({
      name: 'key',
      unlimited_quota: false,
      remain_quota: 500,
      expired_time: NEVER_EXPIRES,
      group: 'vip',
      allow_ips: '1.2.3.4',
    })
    expect(p.id).toBeUndefined()
  })

  it('zeroes quota and preserves id when unlimited on edit', () => {
    const p = tokenFormToPayload(form({ id: 7, unlimitedQuota: true, remainQuota: 42 }))
    expect(p.id).toBe(7)
    expect(p.unlimited_quota).toBe(true)
    expect(p.remain_quota).toBe(0)
  })

  it('sends never-expires sentinel when neverExpires', () => {
    expect(tokenFormToPayload(form({ neverExpires: true })).expired_time).toBe(
      NEVER_EXPIRES,
    )
  })

  it('truncates fractional quota (backend stores int)', () => {
    expect(tokenFormToPayload(form({ remainQuota: 10.9 })).remain_quota).toBe(10)
  })
})

describe('datetime round-trip', () => {
  it('unixToDatetimeLocal returns empty for never/invalid', () => {
    expect(unixToDatetimeLocal(NEVER_EXPIRES)).toBe('')
    expect(unixToDatetimeLocal(0)).toBe('')
    expect(unixToDatetimeLocal(Number.NaN)).toBe('')
  })

  it('round-trips a local datetime string through unix seconds', () => {
    const s = '2030-06-15T13:45'
    const unix = datetimeLocalToUnix(s)
    expect(Number.isFinite(unix)).toBe(true)
    expect(unixToDatetimeLocal(unix)).toBe(s)
  })

  it('datetimeLocalToUnix returns NaN for blank/invalid', () => {
    expect(Number.isNaN(datetimeLocalToUnix(''))).toBe(true)
    expect(Number.isNaN(datetimeLocalToUnix('nope'))).toBe(true)
  })
})

describe('tokenToForm', () => {
  it('marks never-expires when expired_time is -1', () => {
    const f = tokenToForm({ id: 1, name: 'x', expired_time: NEVER_EXPIRES })
    expect(f.neverExpires).toBe(true)
    expect(f.expiresAt).toBe('')
  })

  it('fills expiresAt from a real expiry', () => {
    const unix = datetimeLocalToUnix('2031-03-04T09:30')
    const f = tokenToForm({ id: 1, name: 'x', expired_time: unix })
    expect(f.neverExpires).toBe(false)
    expect(f.expiresAt).toBe('2031-03-04T09:30')
  })
})

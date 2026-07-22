import { describe, expect, it } from 'vitest'
import {
  isAdminLogBodyOk,
  parseLogTypeParam,
  shouldPreferAdmin,
} from './logQuery'

describe('parseLogTypeParam', () => {
  it('drops all/empty', () => {
    expect(parseLogTypeParam('all')).toBeUndefined()
    expect(parseLogTypeParam('')).toBeUndefined()
    expect(parseLogTypeParam(undefined)).toBeUndefined()
  })
  it('parses finite numbers', () => {
    expect(parseLogTypeParam(2)).toBe(2)
    expect(parseLogTypeParam('7')).toBe(7)
  })
  it('rejects NaN strings', () => {
    expect(parseLogTypeParam('nope')).toBeUndefined()
  })
})

describe('shouldPreferAdmin', () => {
  it('only true when explicitly true', () => {
    expect(shouldPreferAdmin(true)).toBe(true)
    expect(shouldPreferAdmin(false)).toBe(false)
    expect(shouldPreferAdmin(undefined)).toBe(false)
  })
})

describe('isAdminLogBodyOk', () => {
  it('requires success true', () => {
    expect(isAdminLogBodyOk({ success: true })).toBe(true)
    expect(isAdminLogBodyOk({ success: false })).toBe(false)
    expect(isAdminLogBodyOk(null)).toBe(false)
  })
})

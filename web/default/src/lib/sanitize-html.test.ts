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
import { JSDOM } from 'jsdom'
import { describe, it } from 'vitest'

import {
  createHtmlSanitizer,
  normalizeAnchorRel,
  sanitizeIframeSrc,
} from './sanitize-html'

const sanitizeHtml = createHtmlSanitizer(new JSDOM('').window as unknown as Window)

describe('normalizeAnchorRel', () => {
  it('forces noopener and noreferrer for blank targets', () => {
    assert.equal(
      normalizeAnchorRel('_blank', 'opener external'),
      'external noopener noreferrer'
    )
  })

  it('leaves non-blank links without forced rel tokens', () => {
    assert.equal(normalizeAnchorRel('_self', 'external'), 'external')
  })
})

describe('sanitizeHtml (DOMPurify path)', () => {
  it('removes scripts, event handlers, and javascript URLs', () => {
    const dirty = [
      '<script>window.pwned = true</script>',
      '<img src="x" onerror="window.pwned = true">',
      '<a href="javascript:window.pwned=true">click</a>',
    ].join('')

    const clean = sanitizeHtml(dirty)

    assert.ok(!clean.toLowerCase().includes('<script'))
    assert.ok(!clean.toLowerCase().includes('onerror'))
    assert.ok(!clean.toLowerCase().includes('javascript:'))
  })

  it('preserves ordinary markup', () => {
    assert.equal(
      sanitizeHtml('<h2>Title</h2><p><strong>Body</strong></p>'),
      '<h2>Title</h2><p><strong>Body</strong></p>'
    )
  })

  it('removes inline styles and embedded documents', () => {
    const clean = sanitizeHtml(
      '<p style="background:url(https://tracker.example)">Body</p><iframe src="https://example.com"></iframe>'
    )
    assert.equal(clean, '<p>Body</p>')
  })
})

describe('sanitizeIframeSrc', () => {
  it('allows http(s) only', () => {
    assert.equal(
      sanitizeIframeSrc('https://example.com/x'),
      'https://example.com/x'
    )
    assert.equal(sanitizeIframeSrc('javascript:alert(1)'), null)
  })
})

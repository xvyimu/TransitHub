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

/**
 * Admin-configured HTML sanitizer (footer, notices, etc.).
 * Single path: DOMPurify — aligned with HtmlContent and classic sanitizeHtml.
 */

import createDOMPurify, { type WindowLike } from 'dompurify'

const SANITIZE_OPTIONS = {
  USE_PROFILES: { html: true },
  FORBID_TAGS: [
    'style',
    'iframe',
    'object',
    'embed',
    'form',
    'script',
    'link',
    'meta',
    'base',
  ],
  FORBID_ATTR: ['style', 'srcdoc', 'srcset'],
} as const

type PurifyInstance = ReturnType<typeof createDOMPurify>

let browserPurify: PurifyInstance | null = null

function getBrowserPurify(): PurifyInstance | null {
  if (typeof window === 'undefined') return null
  if (!browserPurify) {
    browserPurify = createDOMPurify(window)
  }
  return browserPurify
}

/** Test / non-browser: build a purifier bound to an explicit window (e.g. jsdom). */
export function createHtmlSanitizer(
  windowObject: WindowLike
): (dirty: string) => string {
  const purifier = createDOMPurify(windowObject)
  return (dirty: string) =>
    purifier.sanitize(String(dirty ?? ''), { ...SANITIZE_OPTIONS })
}

/**
 * Normalize anchor rel when target is blank-ish.
 * Kept for call sites / tests that only need rel policy.
 */
export function normalizeAnchorRel(
  target: string | null,
  rel: string | null
): string | null {
  const tokens = new Set(
    (rel ?? '')
      .split(/\s+/)
      .map((token) => token.trim().toLowerCase())
      .filter((token) => token && token !== 'opener')
  )
  if (target?.toLowerCase() === '_blank') {
    tokens.add('noopener')
    tokens.add('noreferrer')
  }
  return tokens.size > 0 ? [...tokens].sort().join(' ') : null
}

function hardenAnchors(html: string): string {
  if (typeof document === 'undefined') return html
  try {
    const template = document.createElement('template')
    template.innerHTML = html
    template.content.querySelectorAll('a[target="_blank"]').forEach((link) => {
      const next = normalizeAnchorRel(
        link.getAttribute('target'),
        link.getAttribute('rel')
      )
      if (next) link.setAttribute('rel', next)
      else link.removeAttribute('rel')
    })
    return template.innerHTML
  } catch {
    return html
  }
}

/** Sanitize admin-configured HTML before dangerouslySetInnerHTML. */
export function sanitizeHtml(dirty: string): string {
  if (!dirty) return ''
  const purifier = getBrowserPurify()
  if (!purifier) {
    // SSR / node without window: strip the worst patterns; real path is browser.
    return String(dirty)
      .replaceAll(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
      .replaceAll(/on\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '')
      .replaceAll(/javascript\s*:/gi, '')
  }
  const cleaned = purifier.sanitize(String(dirty), { ...SANITIZE_OPTIONS })
  return hardenAnchors(cleaned)
}

/** Allow only http(s) iframe sources. */
export function sanitizeIframeSrc(url: string): string | null {
  try {
    const u = new URL(url)
    if (u.protocol === 'https:' || u.protocol === 'http:') return u.toString()
  } catch {
    /* empty */
  }
  return null
}

/**
 * Pure helpers extracted for unit tests of log query building / success shell.
 * Runtime path remains in logs.ts listLogs.
 */
export function parseLogTypeParam(
  type: number | string | undefined | null,
): number | undefined {
  if (type === undefined || type === null || type === '' || type === 'all') {
    return undefined
  }
  const n = typeof type === 'number' ? type : Number(type)
  return Number.isFinite(n) ? n : undefined
}

export function shouldPreferAdmin(isAdmin: boolean | undefined): boolean {
  return isAdmin === true
}

/** True when admin list body is usable; false → fall back to /self. */
export function isAdminLogBodyOk(body: { success?: boolean } | null | undefined): boolean {
  return body?.success === true
}

export function parsePositiveUint32RouteQueryParameter(value: unknown): number | null {
  const parsed = typeof value === 'string' && /^[1-9]\d*$/.test(value) ? Number(value) : NaN
  return Number.isSafeInteger(parsed) && parsed <= 4294967295 ? parsed : null
}

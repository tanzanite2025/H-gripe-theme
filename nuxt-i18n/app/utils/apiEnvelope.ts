export type ApiEnvelope<T> = T | { data?: T | { data?: T } }

export const unwrapApiEnvelope = <T>(payload: ApiEnvelope<T> | null | undefined): T | null => {
  let current: unknown = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (current === null || current === undefined) return null
    if (typeof current !== 'object' || Array.isArray(current)) return current as T
    if (!Object.prototype.hasOwnProperty.call(current, 'data')) return current as T
    current = (current as { data?: unknown }).data
  }
  return null
}

export const createIdempotencyKey = (prefix = 'checkout') => {
  let token: string
  if (typeof globalThis.crypto?.randomUUID === 'function') {
    token = globalThis.crypto.randomUUID()
  } else if (typeof globalThis.crypto?.getRandomValues === 'function') {
    const bytes = new Uint8Array(16)
    globalThis.crypto.getRandomValues(bytes)
    bytes[6] = (bytes[6]! & 0x0f) | 0x40
    bytes[8] = (bytes[8]! & 0x3f) | 0x80
    token = Array.from(bytes, value => value.toString(16).padStart(2, '0'))
      .join('')
      .replace(/^(.{8})(.{4})(.{4})(.{4})(.{12})$/, '$1-$2-$3-$4-$5')
  } else {
    token = `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`
  }
  return `${prefix}-${token}`
}

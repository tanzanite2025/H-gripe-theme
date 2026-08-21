const DEVICE_FINGERPRINT_HEADER = 'X-Device-Fingerprint'

let deviceFingerprintPromise: Promise<string> | null = null

type WebGLDebugRendererInfoLike = {
  UNMASKED_RENDERER_WEBGL: number
}

const toHex = (bytes: ArrayBuffer) =>
  Array.from(new Uint8Array(bytes), byte => byte.toString(16).padStart(2, '0')).join('')

const readWebGlRenderer = (): string => {
  if (typeof document === 'undefined') return ''

  const canvas = document.createElement('canvas')
  const context = canvas.getContext('webgl') || canvas.getContext('experimental-webgl')
  if (!context) return ''

  const gl = context as WebGLRenderingContext
  const debugInfo = gl.getExtension('WEBGL_debug_renderer_info') as WebGLDebugRendererInfoLike | null
  if (debugInfo) {
    const renderer = gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL)
    if (typeof renderer === 'string' && renderer.trim()) {
      return renderer.trim()
    }
  }

  const fallbackRenderer = gl.getParameter(gl.RENDERER)
  return typeof fallbackRenderer === 'string' ? fallbackRenderer.trim() : String(fallbackRenderer || '').trim()
}

const collectDeviceFingerprintSignals = (): string => {
  if (typeof window === 'undefined') return ''

  const colorDepth = typeof window.screen?.colorDepth === 'number'
    ? String(window.screen.colorDepth)
    : ''
  const timezoneOffset = String(new Date().getTimezoneOffset())
  const renderer = readWebGlRenderer()

  return [colorDepth, timezoneOffset, renderer]
    .map(value => value.trim())
    .filter(Boolean)
    .join('|')
}

const hashFingerprint = async (value: string): Promise<string> => {
  if (!value || !globalThis.crypto?.subtle) {
    return ''
  }

  try {
    const digest = await globalThis.crypto.subtle.digest('SHA-256', new TextEncoder().encode(value))
    return toHex(digest)
  } catch {
    return ''
  }
}

export const resolveDeviceFingerprint = async (): Promise<string> => {
  if (typeof window === 'undefined') {
    return ''
  }

  if (!deviceFingerprintPromise) {
    deviceFingerprintPromise = (async () => {
      const signals = collectDeviceFingerprintSignals()
      if (!signals) {
        return ''
      }
      return await hashFingerprint(signals)
    })().catch(() => '')
  }

  return deviceFingerprintPromise
}

export const deviceFingerprintHeaderName = DEVICE_FINGERPRINT_HEADER

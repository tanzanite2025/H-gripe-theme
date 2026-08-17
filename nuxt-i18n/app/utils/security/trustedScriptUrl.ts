const POLICY_NAME = 'tanzanite-script-url'

interface TrustedScriptURLPolicy {
  createScriptURL: (value: string) => unknown
}

interface TrustedTypesFactory {
  createPolicy: (
    name: string,
    rules: {
      createScriptURL: (value: string) => string
    },
  ) => TrustedScriptURLPolicy
}

interface TrustedExternalScriptOptions {
  id: string
  url: string
  async?: boolean
  defer?: boolean
  dataset?: Record<string, string>
}

declare global {
  interface Window {
    trustedTypes?: TrustedTypesFactory
    __tanzaniteTrustedScriptUrlPolicy?: TrustedScriptURLPolicy
  }
}

const trustedExternalScriptUrls = new Set([
  'https://accounts.google.com/gsi/client',
  'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit',
  'https://js.stripe.com/dahlia/stripe.js',
])

const googleScriptUrl = (value: string): string | null => {
  try {
    const url = new URL(value)
    if (url.origin !== 'https://www.googletagmanager.com') return null
    if (!['/gtag/js', '/gtm.js'].includes(url.pathname)) return null

    const identifier = url.searchParams.get('id') || ''
    const isGoogleAnalyticsId = /^G-[A-Z0-9]{4,}$/i.test(identifier)
    const isGoogleTagManagerId = /^GTM-[A-Z0-9]+$/i.test(identifier)
    if (!isGoogleAnalyticsId && !isGoogleTagManagerId) return null
    if ([...url.searchParams.keys()].some(key => key !== 'id')) return null

    return url.href
  } catch {
    return null
  }
}

const allowedScriptUrl = (value: string): string => {
  const normalized = value.trim()
  if (trustedExternalScriptUrls.has(normalized)) return normalized

  const googleUrl = googleScriptUrl(normalized)
  if (googleUrl) return googleUrl

  throw new TypeError(`Refused an unapproved external script URL: ${normalized}`)
}

const policy = (): TrustedScriptURLPolicy | null => {
  if (!import.meta.client || !window.trustedTypes) return null
  if (window.__tanzaniteTrustedScriptUrlPolicy) return window.__tanzaniteTrustedScriptUrlPolicy

  const trustedTypesPolicy = window.trustedTypes.createPolicy(POLICY_NAME, {
    createScriptURL: allowedScriptUrl,
  })
  window.__tanzaniteTrustedScriptUrlPolicy = trustedTypesPolicy
  return trustedTypesPolicy
}

const assignTrustedScriptUrl = (script: HTMLScriptElement, url: string): void => {
  const trustedPolicy = policy()
  if (!trustedPolicy) {
    script.src = allowedScriptUrl(url)
    return
  }

  script.src = trustedPolicy.createScriptURL(url) as string
}

const waitForScript = (script: HTMLScriptElement): Promise<HTMLScriptElement> => {
  if (script.dataset.tanzaniteScriptState === 'loaded') return Promise.resolve(script)

  return new Promise((resolve, reject) => {
    script.addEventListener('load', () => {
      script.dataset.tanzaniteScriptState = 'loaded'
      resolve(script)
    }, { once: true })
    script.addEventListener('error', () => {
      script.dataset.tanzaniteScriptState = 'failed'
      reject(new Error(`Failed to load ${script.src}`))
    }, { once: true })
  })
}

export const initializeTrustedTypes = (): void => {
  void policy()
}

export const loadTrustedExternalScript = (
  options: TrustedExternalScriptOptions,
): Promise<HTMLScriptElement> => {
  if (!import.meta.client) {
    return Promise.reject(new Error('External scripts can only load in the browser'))
  }

  const existing = document.getElementById(options.id)
  if (existing instanceof HTMLScriptElement) return waitForScript(existing)

  const script = document.createElement('script')
  script.id = options.id
  script.async = options.async ?? true
  script.defer = options.defer ?? false
  for (const [name, value] of Object.entries(options.dataset || {})) {
    script.dataset[name] = value
  }

  const loaded = waitForScript(script)
  assignTrustedScriptUrl(script, options.url)
  document.head.appendChild(script)
  return loaded
}

export const loadGoogleIdentityScript = (): Promise<HTMLScriptElement> => (
  loadTrustedExternalScript({
    id: 'commerce-platform-google-identity',
    url: 'https://accounts.google.com/gsi/client',
    defer: true,
  })
)

export const loadTurnstileScript = (): Promise<HTMLScriptElement> => (
  loadTrustedExternalScript({
    id: 'commerce-platform-turnstile',
    url: 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit',
    defer: true,
    dataset: {
      commercePlatformTurnstile: 'true',
    },
  })
)

export const loadStripeScript = (): Promise<HTMLScriptElement> => (
  loadTrustedExternalScript({
    id: 'commerce-platform-stripe-js',
    url: 'https://js.stripe.com/dahlia/stripe.js',
  })
)

export const loadGoogleAnalyticsScript = (measurementId: string): Promise<HTMLScriptElement> => (
  loadTrustedExternalScript({
    id: 'commerce-platform-google-analytics',
    url: `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(measurementId)}`,
  })
)

export const loadGoogleTagManagerScript = (containerId: string): Promise<HTMLScriptElement> => (
  loadTrustedExternalScript({
    id: 'commerce-platform-google-tag-manager',
    url: `https://www.googletagmanager.com/gtm.js?id=${encodeURIComponent(containerId)}`,
  })
)

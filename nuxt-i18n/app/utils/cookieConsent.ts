export const COOKIE_CONSENT_KEY = 'learn_gripe_cookie_consent'
export const COOKIE_CONSENT_UPDATED_EVENT = 'cookie-consent-updated'

export interface CookieConsentPreferences {
  essential: boolean
  performance: boolean
  preference: boolean
  advertising: boolean
  timestamp: number
}

export function readCookieConsent(): CookieConsentPreferences | null {
  if (!import.meta.client) return null

  const stored = localStorage.getItem(COOKIE_CONSENT_KEY)
  if (!stored) return null

  try {
    return JSON.parse(stored) as CookieConsentPreferences
  } catch {
    return null
  }
}

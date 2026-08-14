import { useRouter } from '#imports'
import {
  fetchAnalyticsSettings,
  type AnalyticsSettings,
} from '~/composables/useAnalyticsSettings'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import {
  COOKIE_CONSENT_UPDATED_EVENT,
  readCookieConsent,
  type CookieConsentPreferences,
} from '~/utils/cookieConsent'

const GOOGLE_ANALYTICS_SCRIPT_ID = 'commerce-platform-google-analytics'
const GOOGLE_TAG_MANAGER_SCRIPT_ID = 'commerce-platform-google-tag-manager'

type Gtag = (...args: unknown[]) => void

type AnalyticsWindow = Window & {
  dataLayer?: unknown[]
  gtag?: Gtag
}

const currentPage = () => ({
      page_path: `${window.location.pathname}${window.location.search}`,
  page_location: window.location.href,
  page_title: document.title,
})

const canLoadGoogleTagManager = (consent: CookieConsentPreferences | null) => {
  return Boolean(consent?.advertising)
}

const updateGoogleConsent = (consent: CookieConsentPreferences | null) => {
  const analyticsWindow = window as AnalyticsWindow
  if (typeof analyticsWindow.gtag !== 'function') return

  analyticsWindow.gtag('consent', 'update', {
    analytics_storage: consent?.performance ? 'granted' : 'denied',
    ad_storage: consent?.advertising ? 'granted' : 'denied',
    ad_user_data: consent?.advertising ? 'granted' : 'denied',
    ad_personalization: consent?.advertising ? 'granted' : 'denied',
  })
}

const ensureGoogleAnalytics = (measurementId: string, consent: CookieConsentPreferences) => {
  const analyticsWindow = window as AnalyticsWindow
  analyticsWindow.dataLayer = analyticsWindow.dataLayer || []
  analyticsWindow.gtag = analyticsWindow.gtag || ((...args: unknown[]) => {
    analyticsWindow.dataLayer?.push(args)
  })

  analyticsWindow.gtag('consent', 'default', {
    analytics_storage: consent.performance ? 'granted' : 'denied',
    ad_storage: consent.advertising ? 'granted' : 'denied',
    ad_user_data: consent.advertising ? 'granted' : 'denied',
    ad_personalization: consent.advertising ? 'granted' : 'denied',
    wait_for_update: 500,
  })
  analyticsWindow.gtag('js', new Date())
  analyticsWindow.gtag('config', measurementId, { send_page_view: false })

  if (!document.getElementById(GOOGLE_ANALYTICS_SCRIPT_ID)) {
    const script = document.createElement('script')
    script.id = GOOGLE_ANALYTICS_SCRIPT_ID
    script.async = true
    script.src = `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(measurementId)}`
    document.head.appendChild(script)
  }
}

const ensureGoogleTagManager = (containerId: string) => {
	const analyticsWindow = window as AnalyticsWindow
	analyticsWindow.dataLayer = analyticsWindow.dataLayer || []

	if (!document.getElementById(GOOGLE_TAG_MANAGER_SCRIPT_ID)) {
		analyticsWindow.dataLayer.push({
			'gtm.start': new Date().getTime(),
			event: 'gtm.js',
		})
		const script = document.createElement('script')
		script.id = GOOGLE_TAG_MANAGER_SCRIPT_ID
    script.async = true
    script.src = `https://www.googletagmanager.com/gtm.js?id=${encodeURIComponent(containerId)}`
    document.head.appendChild(script)
  }
}

const trackPageView = () => {
  const analyticsWindow = window as AnalyticsWindow
  const page = currentPage()

  if (typeof analyticsWindow.gtag === 'function') {
    analyticsWindow.gtag('event', 'page_view', page)
  }
  analyticsWindow.dataLayer?.push({ event: 'page_view', ...page })
}

export default defineNuxtPlugin(async () => {
  const apiBase = usePublicApiBase().value
  const router = useRouter()

  let settings: AnalyticsSettings
  try {
    settings = await fetchAnalyticsSettings(apiBase)
  } catch (error) {
    console.warn('Failed to initialize analytics settings:', error)
    return
  }

  const applyConsent = (consent: CookieConsentPreferences | null) => {
    if (!consent) return

    if (settings.googleAnalytics && consent.performance) {
      ensureGoogleAnalytics(settings.googleAnalytics, consent)
    }
    if (settings.googleTagManager && canLoadGoogleTagManager(consent)) {
      ensureGoogleTagManager(settings.googleTagManager)
    }

    updateGoogleConsent(consent)
    if (consent.performance || canLoadGoogleTagManager(consent)) {
      trackPageView()
    }
  }

  applyConsent(readCookieConsent())

  const handleConsentUpdate = (event: Event) => {
    const detail = (event as CustomEvent<CookieConsentPreferences>).detail
    applyConsent(detail || readCookieConsent())
  }

  window.addEventListener(COOKIE_CONSENT_UPDATED_EVENT, handleConsentUpdate)

  router.afterEach(() => {
    const consent = readCookieConsent()
    if (!consent) return
    if (settings.googleAnalytics && consent.performance) {
      trackPageView()
      return
    }
    if (settings.googleTagManager && canLoadGoogleTagManager(consent)) {
      trackPageView()
    }
  })
})

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
import {
  loadGoogleAnalyticsScript,
  loadGoogleTagManagerScript,
} from '~/utils/security/trustedScriptUrl'

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

  void loadGoogleAnalyticsScript(measurementId).catch((error: unknown) => {
    console.warn('Failed to load Google Analytics:', error)
  })
}

const ensureGoogleTagManager = (containerId: string) => {
	const analyticsWindow = window as AnalyticsWindow
	analyticsWindow.dataLayer = analyticsWindow.dataLayer || []

  if (document.getElementById('commerce-platform-google-tag-manager')) return

  analyticsWindow.dataLayer.push({
    'gtm.start': new Date().getTime(),
    event: 'gtm.js',
  })
  void loadGoogleTagManagerScript(containerId).catch((error: unknown) => {
    console.warn('Failed to load Google Tag Manager:', error)
  })
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

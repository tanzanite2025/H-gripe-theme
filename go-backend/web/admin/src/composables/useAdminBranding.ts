import axios from 'axios'
import { computed, reactive } from 'vue'

interface AdminBrandingState {
  loaded: boolean
  loading: boolean
  brandName: string
  brandInitial: string
  panelLabel: string
  loginTitle: string
  footerText: string
  htmlTitle: string
}

type SiteSettingsPayload = Record<string, unknown>

const publicApi = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000,
  withCredentials: true
})

const asString = (value: unknown): string => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return String(value)
}

const branding = reactive<AdminBrandingState>({
  loaded: false,
  loading: false,
  brandName: '',
  brandInitial: '',
  panelLabel: '',
  loginTitle: '',
  footerText: '',
  htmlTitle: ''
})

let loadPromise: Promise<AdminBrandingState> | null = null

const applySettings = (raw: SiteSettingsPayload = {}): void => {
  branding.brandName = asString(raw.adminBrandName ?? raw.admin_brand_name).trim()
  branding.brandInitial = asString(raw.adminBrandInitial ?? raw.admin_brand_initial).trim()
  branding.panelLabel = asString(raw.adminPanelLabel ?? raw.admin_panel_label).trim()
  branding.loginTitle = asString(raw.adminLoginTitle ?? raw.admin_login_title).trim()
  branding.footerText = asString(raw.adminFooterText ?? raw.admin_footer_text).trim()
  branding.htmlTitle = asString(raw.adminHTMLTitle ?? raw.admin_html_title).trim()
}

export const loadAdminBranding = async (force = false): Promise<AdminBrandingState> => {
  if (branding.loaded && !force) return branding
  if (loadPromise && !force) return loadPromise

  branding.loading = true
  loadPromise = publicApi
    .get<SiteSettingsPayload>('/api/v1/settings/site', {
      params: { locale: 'en' },
      headers: { accept: 'application/json' }
    })
    .then((response) => {
      applySettings(response.data || {})
      branding.loaded = true
      return branding
    })
    .catch((error) => {
      console.warn('Failed to load admin branding:', error)
      branding.loaded = true
      return branding
    })
    .finally(() => {
      branding.loading = false
      loadPromise = null
    })

  return loadPromise
}

export const getAdminDocumentTitle = (routeTitle = ''): string => {
  const section = String(routeTitle || '').trim()
  const appTitle = branding.htmlTitle.trim()
  if (appTitle && section && section !== appTitle) return `${section} - ${appTitle}`
  return appTitle || ''
}

export const setAdminDocumentTitle = (routeTitle = ''): void => {
  if (typeof document === 'undefined') return
  document.title = getAdminDocumentTitle(routeTitle)
}

export const useAdminBranding = () => {
  const brandName = computed(() => branding.brandName)
  const brandInitial = computed(() => branding.brandInitial)
  const panelLabel = computed(() => branding.panelLabel)
  const loginTitle = computed(() => branding.loginTitle)
  const footerText = computed(() => branding.footerText)

  return {
    branding,
    brandName,
    brandInitial,
    panelLabel,
    loginTitle,
    footerText,
    loadAdminBranding,
    setAdminDocumentTitle
  }
}

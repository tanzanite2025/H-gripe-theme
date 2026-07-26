import axios from 'axios'
import { computed, reactive } from 'vue'

const publicApi = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000,
  withCredentials: true
})

const asString = (value) => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return String(value)
}

const branding = reactive({
  loaded: false,
  loading: false,
  brandName: '',
  brandInitial: '',
  panelLabel: '',
  loginTitle: '',
  footerText: '',
  htmlTitle: '',
  siteUrl: ''
})

let loadPromise = null

const applySettings = (raw = {}) => {
  branding.brandName = asString(raw.adminBrandName ?? raw.admin_brand_name).trim()
  branding.brandInitial = asString(raw.adminBrandInitial ?? raw.admin_brand_initial).trim()
  branding.panelLabel = asString(raw.adminPanelLabel ?? raw.admin_panel_label).trim()
  branding.loginTitle = asString(raw.adminLoginTitle ?? raw.admin_login_title).trim()
  branding.footerText = asString(raw.adminFooterText ?? raw.admin_footer_text).trim()
  branding.htmlTitle = asString(raw.adminHTMLTitle ?? raw.admin_html_title).trim()
  branding.siteUrl = asString(raw.siteUrl ?? raw.site_url).trim().replace(/\/+$/, '')
}

export const loadAdminBranding = async (force = false) => {
  if (branding.loaded && !force) return branding
  if (loadPromise && !force) return loadPromise

  branding.loading = true
  loadPromise = publicApi
    .get('/api/v1/settings/site', {
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

export const getAdminDocumentTitle = (routeTitle = '') => {
  const section = String(routeTitle || '').trim()
  const appTitle = branding.htmlTitle.trim()
  if (appTitle && section && section !== appTitle) return `${section} - ${appTitle}`
  return appTitle || ''
}

export const setAdminDocumentTitle = (routeTitle = '') => {
  if (typeof document === 'undefined') return
  document.title = getAdminDocumentTitle(routeTitle)
}

export const useAdminBranding = () => {
  const brandName = computed(() => branding.brandName)
  const brandInitial = computed(() => branding.brandInitial)
  const panelLabel = computed(() => branding.panelLabel)
  const loginTitle = computed(() => branding.loginTitle)
  const footerText = computed(() => branding.footerText)
  const siteUrl = computed(() => branding.siteUrl)

  return {
    branding,
    brandName,
    brandInitial,
    panelLabel,
    loginTitle,
    footerText,
    siteUrl,
    loadAdminBranding,
    setAdminDocumentTitle
  }
}

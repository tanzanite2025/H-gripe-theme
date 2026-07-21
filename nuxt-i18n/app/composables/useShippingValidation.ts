import { ref } from 'vue'
import { COUNTRIES, getCountryByCode, getZipFormatHint, validateZipFormat } from '~/data/countries'

export interface ShippingRule {
  type?: string
  min?: number | null
  max?: number | null
  min_value?: number | null
  max_value?: number | null
  fee: number
  priority?: number
  free_over?: number | null
  region?: string
  regions?: string[]
  zip_ranges?: string[]
  eta_min_days?: number | null
  eta_max_days?: number | null
  service?: string
  service_label?: string
}

export interface ShippingTemplate {
  id: number
  name?: string
  template_name?: string
  type?: string
  description?: string
  enabled?: boolean
  is_active?: boolean
  free_shipping?: boolean
  free_threshold?: number
  default_fee?: number
  rules: ShippingRule[]
  meta?: {
    carrier?: string
    currency?: string
  }
}

export interface ShippingValidationResult {
  isShippable: boolean
  matchedRule: ShippingRule | null
  matchedTemplate: ShippingTemplate | null
  reason?: string
}

type ApiResponse<T> = T | { data?: T | { data?: T } }

const allCountryCodes = () => COUNTRIES.map(country => country.code.toUpperCase())

const unwrapTemplateList = (payload: ApiResponse<ShippingTemplate[]> | null | undefined): ShippingTemplate[] => {
  let current: unknown = payload

  for (let depth = 0; depth < 3; depth += 1) {
    if (Array.isArray(current)) return current as ShippingTemplate[]
    if (!current || typeof current !== 'object') return []
    current = (current as Record<string, unknown>).data
  }

  return []
}

const isTemplateActive = (template: ShippingTemplate): boolean => {
  return template.enabled !== false && template.is_active !== false
}

const normalizeRegionList = (regions: string[]): string[] => {
  return regions.map(region => String(region).trim().toUpperCase()).filter(Boolean)
}

const splitRegions = (region: string): string[] => {
  const trimmed = region.trim()
  if (!trimmed) return []

  try {
    const parsed = JSON.parse(trimmed)
    if (Array.isArray(parsed)) {
      return normalizeRegionList(parsed)
    }
  } catch {
    // Use delimiter parsing below.
  }

  return normalizeRegionList(trimmed.split(/[,;|\s]+/))
}

const ruleRegions = (rule: ShippingRule): string[] => {
  if (Array.isArray(rule.regions)) {
    return normalizeRegionList(rule.regions)
  }
  return splitRegions(rule.region || '')
}

const ruleMatchesCountry = (rule: ShippingRule, countryCode: string): boolean => {
  const normalizedCountry = countryCode.toUpperCase()
  const regions = ruleRegions(rule)
  if (regions.length === 0) return true
  if (regions.some(region => ['*', 'ALL', 'GLOBAL', 'WORLDWIDE'].includes(region))) return true
  return regions.includes(normalizedCountry)
}

export function isZipInRanges(zipCode: string, zipRanges: string[]): boolean {
  if (!zipRanges || zipRanges.length === 0) {
    return true
  }

  const normalizedZip = zipCode.replace(/\s/g, '').toUpperCase()

  for (const range of zipRanges) {
    if (range.includes('-')) {
      const parts = range.split('-').map(item => item.trim().toUpperCase())
      const start = parts[0] || ''
      const end = parts[1] || ''

      if (!start || !end) continue

      if (/^\d+$/.test(normalizedZip) && /^\d+$/.test(start) && /^\d+$/.test(end)) {
        const zipNum = parseInt(normalizedZip)
        const startNum = parseInt(start)
        const endNum = parseInt(end)
        if (zipNum >= startNum && zipNum <= endNum) {
          return true
        }
      } else if (normalizedZip >= start && normalizedZip <= end) {
        return true
      }
    } else if (normalizedZip === range.trim().toUpperCase()) {
      return true
    }
  }

  return false
}

function findMatchingRule(
  countryCode: string,
  zipCode: string,
  templates: ShippingTemplate[]
): { rule: ShippingRule | null; template: ShippingTemplate | null } {
  const countryRules: Array<{ rule: ShippingRule; template: ShippingTemplate }> = []

  for (const template of templates) {
    if (!isTemplateActive(template)) continue

    if (!template.rules?.length) {
      countryRules.push({
        rule: {
          type: template.type || 'weight',
          fee: Number(template.default_fee || 0),
          region: 'ALL',
          free_over: template.free_shipping ? template.free_threshold || 0 : null,
        },
        template,
      })
      continue
    }

    for (const rule of template.rules || []) {
      if (ruleMatchesCountry(rule, countryCode)) {
        countryRules.push({ rule, template })
      }
    }
  }

  if (countryRules.length === 0) {
    return { rule: null, template: null }
  }

  if (zipCode) {
    for (const { rule, template } of countryRules) {
      const zipRanges = rule.zip_ranges || []
      if (zipRanges.length > 0 && isZipInRanges(zipCode, zipRanges)) {
        return { rule, template }
      }
    }
  }

  for (const { rule, template } of countryRules) {
    const zipRanges = rule.zip_ranges || []
    if (zipRanges.length === 0) {
      return { rule, template }
    }
  }

  return { rule: null, template: null }
}

export function useShippingValidation() {
  const shippingTemplates = ref<ShippingTemplate[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function loadShippingTemplates(): Promise<void> {
    if (shippingTemplates.value.length > 0) {
      return
    }

    isLoading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
      const response = await $fetch<ApiResponse<ShippingTemplate[]>>(`${base}/shipping/templates`)
      shippingTemplates.value = unwrapTemplateList(response)
    } catch (e) {
      console.error('Failed to load shipping templates:', e)
      error.value = 'Failed to load shipping information'
    } finally {
      isLoading.value = false
    }
  }

  function isCountryShippable(countryCode: string): boolean {
    if (!countryCode) return false

    for (const template of shippingTemplates.value) {
      if (!isTemplateActive(template)) continue

      if (!template.rules?.length) {
        return true
      }

      for (const rule of template.rules || []) {
        if (ruleMatchesCountry(rule, countryCode)) {
          return true
        }
      }
    }

    return false
  }

  function getShippableCountries(): string[] {
    const countries = new Set<string>()

    for (const template of shippingTemplates.value) {
      if (!isTemplateActive(template)) continue

      if (!template.rules?.length) {
        allCountryCodes().forEach(code => countries.add(code))
        continue
      }

      for (const rule of template.rules || []) {
        const regions = ruleRegions(rule)
        if (regions.length === 0) {
          allCountryCodes().forEach(code => countries.add(code))
          continue
        }
        if (regions.some(region => ['*', 'ALL', 'GLOBAL', 'WORLDWIDE'].includes(region))) {
          allCountryCodes().forEach(code => countries.add(code))
          continue
        }
        regions.forEach(region => countries.add(region))
      }
    }

    return Array.from(countries)
  }

  function getSortedCountries() {
    const shippable = getShippableCountries()

    return [...COUNTRIES].sort((a, b) => {
      const aShippable = shippable.includes(a.code)
      const bShippable = shippable.includes(b.code)

      if (aShippable && !bShippable) return -1
      if (!aShippable && bShippable) return 1
      return a.name.localeCompare(b.name)
    })
  }

  function validateShipping(countryCode: string, zipCode: string = ''): ShippingValidationResult {
    const hasTemplates = shippingTemplates.value.length > 0

    if (!countryCode) {
      return {
        isShippable: false,
        matchedRule: null,
        matchedTemplate: null,
        reason: 'Please select a country',
      }
    }

    if (!hasTemplates && import.meta.dev) {
      return {
        isShippable: true,
        matchedRule: null,
        matchedTemplate: null,
        reason: undefined,
      }
    }

    const { rule, template } = findMatchingRule(countryCode, zipCode, shippingTemplates.value)

    if (!rule) {
      const country = getCountryByCode(countryCode)
      const countryName = country?.name || countryCode

      return {
        isShippable: false,
        matchedRule: null,
        matchedTemplate: null,
        reason: zipCode
          ? `Sorry, we don't ship to ${countryName} (${zipCode})`
          : `Sorry, we don't ship to ${countryName}`,
      }
    }

    return {
      isShippable: true,
      matchedRule: rule,
      matchedTemplate: template,
    }
  }

  function getEstimatedDeliveryText(rule: ShippingRule | null): string | null {
    if (!rule) return null

    const minDays = rule.eta_min_days
    const maxDays = rule.eta_max_days

    if (minDays != null && maxDays != null) {
      return `${minDays}-${maxDays} business days`
    }
    if (minDays != null) {
      return `${minDays}+ business days`
    }
    if (maxDays != null) {
      return `Up to ${maxDays} business days`
    }

    return null
  }

  return {
    shippingTemplates,
    isLoading,
    error,
    loadShippingTemplates,
    isCountryShippable,
    getShippableCountries,
    getSortedCountries,
    validateShipping,
    getEstimatedDeliveryText,
    getZipFormatHint,
    validateZipFormat,
    getCountryByCode,
  }
}

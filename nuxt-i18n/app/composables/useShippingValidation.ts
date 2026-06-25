/**
 * 配送验证 Composable
 * 用于验证用户所在地区是否可配送
 */

import { ref, computed } from 'vue'
import { COUNTRIES, getCountryByCode, getZipFormatHint, validateZipFormat } from '~/data/countries'

/**
 * 运费规则接口
 */
export interface ShippingRule {
  type: string
  min: number | null
  max: number | null
  fee: number
  priority?: number
  free_over?: number | null
  regions?: string[]
  zip_ranges?: string[]
  eta_min_days?: number | null
  eta_max_days?: number | null
  service?: string
  service_label?: string
}

/**
 * 运费模板接口
 */
export interface ShippingTemplate {
  id: number
  template_name: string
  description?: string
  is_active: boolean
  rules: ShippingRule[]
  meta?: {
    carrier?: string
    currency?: string
  }
}

/**
 * 配送验证结果
 */
export interface ShippingValidationResult {
  isShippable: boolean
  matchedRule: ShippingRule | null
  matchedTemplate: ShippingTemplate | null
  reason?: string
}

/**
 * 检查邮编是否在指定范围内
 * @export 供其他模块复用
 */
export function isZipInRanges(zipCode: string, zipRanges: string[]): boolean {
  if (!zipRanges || zipRanges.length === 0) {
    return true // 空数组表示该国家的兜底规则（国家已匹配的前提下）
  }

  const normalizedZip = zipCode.replace(/\s/g, '').toUpperCase()

  for (const range of zipRanges) {
    if (range.includes('-')) {
      const parts = range.split('-').map(s => s.trim().toUpperCase())
      const start = parts[0] || ''
      const end = parts[1] || ''
      
      if (!start || !end) continue
      
      // 数字邮编比较
      if (/^\d+$/.test(normalizedZip) && /^\d+$/.test(start) && /^\d+$/.test(end)) {
        const zipNum = parseInt(normalizedZip)
        const startNum = parseInt(start)
        const endNum = parseInt(end)
        if (zipNum >= startNum && zipNum <= endNum) {
          return true
        }
      }
      // 字母数字邮编比较（如英国 SW1A 1AA）
      else if (normalizedZip >= start && normalizedZip <= end) {
        return true
      }
    } else {
      // 单个邮编精确匹配
      if (normalizedZip === range.trim().toUpperCase()) {
        return true
      }
    }
  }

  return false
}

/**
 * 查找匹配的配送规则
 */
function findMatchingRule(
  countryCode: string,
  zipCode: string,
  templates: ShippingTemplate[]
): { rule: ShippingRule | null; template: ShippingTemplate | null } {
  const normalizedCountry = countryCode.toUpperCase()

  // 收集所有匹配国家的规则
  const countryRules: Array<{ rule: ShippingRule; template: ShippingTemplate }> = []

  for (const template of templates) {
    if (!template.is_active) continue

    for (const rule of template.rules || []) {
      // 检查国家匹配 — regions 为空或不包含用户国家则跳过
      const regions = rule.regions || []
      if (regions.length === 0 || !regions.map(r => r.toUpperCase()).includes(normalizedCountry)) {
        continue
      }

      countryRules.push({ rule, template })
    }
  }

  if (countryRules.length === 0) {
    return { rule: null, template: null }
  }

  // 如果有邮编，优先匹配有邮编范围的规则
  if (zipCode) {
    for (const { rule, template } of countryRules) {
      if (rule.zip_ranges && rule.zip_ranges.length > 0) {
        if (isZipInRanges(zipCode, rule.zip_ranges)) {
          return { rule, template }
        }
      }
    }
  }

  // 其次匹配兜底规则（zip_ranges 为空）
  for (const { rule, template } of countryRules) {
    if (!rule.zip_ranges || rule.zip_ranges.length === 0) {
      return { rule, template }
    }
  }

  return { rule: null, template: null }
}

/**
 * 配送验证 Composable
 */
export function useShippingValidation() {
  const shippingTemplates = ref<ShippingTemplate[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  /**
   * 加载运费模板
   */
  async function loadShippingTemplates(): Promise<void> {
    if (shippingTemplates.value.length > 0) {
      return // 已加载，不重复请求
    }

    isLoading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
      const response = await $fetch<{ data: ShippingTemplate[] }>(
        `${base}/shipping/templates`
      )

      if (Array.isArray(response?.data)) {
        shippingTemplates.value = response.data
      }
    } catch (e) {
      console.error('Failed to load shipping templates:', e)
      error.value = 'Failed to load shipping information'
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 检查国家是否可配送
   */
  function isCountryShippable(countryCode: string): boolean {
    if (!countryCode) return false

    const normalizedCountry = countryCode.toUpperCase()

    for (const template of shippingTemplates.value) {
      if (!template.is_active) continue

      for (const rule of template.rules || []) {
        const regions = rule.regions || []
        // regions 为空表示不可配送
        if (regions.length > 0 && regions.map(r => r.toUpperCase()).includes(normalizedCountry)) {
          return true
        }
      }
    }

    return false
  }

  /**
   * 获取所有可配送的国家代码
   */
  function getShippableCountries(): string[] {
    const countries = new Set<string>()

    for (const template of shippingTemplates.value) {
      if (!template.is_active) continue

      for (const rule of template.rules || []) {
        const regions = rule.regions || []
        for (const region of regions) {
          countries.add(region.toUpperCase())
        }
      }
    }

    return Array.from(countries)
  }

  /**
   * 获取排序后的国家列表（可配送的在前）
   */
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

  /**
   * 验证配送可用性
   */
  function validateShipping(countryCode: string, zipCode: string = ''): ShippingValidationResult {
    const hasTemplates = shippingTemplates.value.length > 0

    if (!countryCode) {
      return {
        isShippable: false,
        matchedRule: null,
        matchedTemplate: null,
        reason: 'Please select a country'
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

      if (zipCode) {
        return {
          isShippable: false,
          matchedRule: null,
          matchedTemplate: null,
          reason: `Sorry, we don't ship to ${countryName} (${zipCode})`
        }
      }

      return {
        isShippable: false,
        matchedRule: null,
        matchedTemplate: null,
        reason: `Sorry, we don't ship to ${countryName}`
      }
    }

    return {
      isShippable: true,
      matchedRule: rule,
      matchedTemplate: template
    }
  }

  /**
   * 获取预计送达时间文本
   */
  function getEstimatedDeliveryText(rule: ShippingRule | null): string | null {
    if (!rule) return null

    const minDays = rule.eta_min_days
    const maxDays = rule.eta_max_days

    if (minDays != null && maxDays != null) {
      return `${minDays}-${maxDays} business days`
    } else if (minDays != null) {
      return `${minDays}+ business days`
    } else if (maxDays != null) {
      return `Up to ${maxDays} business days`
    }

    return null
  }

  return {
    // State
    shippingTemplates,
    isLoading,
    error,

    // Methods
    loadShippingTemplates,
    isCountryShippable,
    getShippableCountries,
    getSortedCountries,
    validateShipping,
    getEstimatedDeliveryText,

    // Re-export utilities
    getZipFormatHint,
    validateZipFormat,
    getCountryByCode
  }
}

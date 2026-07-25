export const bindingScopeLabel = (scope: string) => {
  const labels: Record<string, string> = {
    default: '默认',
    product_type: '产品类型',
    product: '产品',
    variant: 'SKU / 变体',
  }
  return labels[scope] || scope || '-'
}

export const bindingTargetLabel = (binding: any) => {
  if (binding.scope === 'default') return '全局默认'
  if (binding.scope === 'product_type') return `product_type_id=${binding.product_type_id || '-'}`
  if (binding.scope === 'product') return `product_id=${binding.product_id || '-'}`
  if (binding.scope === 'variant') return `variant_id=${binding.variant_id || '-'}`
  return '-'
}

export const bindingTemplateName = (binding: any, templates: any[] = []) => {
  if (binding.template?.name) return binding.template.name
  const template = templates.find((item) => Number(item.id) === Number(binding.template_id))
  return template?.name || '未知模板'
}

export const templateTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    weight: '按重量',
    quantity: '按数量',
    price: '按金额',
  }
  return labels[type] || type || '-'
}

export const formatMoney = (value: any) => Number(value || 0).toFixed(2)
export const formatDate = (value: any) => value ? new Date(value).toLocaleString('zh-CN') : '-'

export const formatRuleSummary = (rules: any[] = []) => {
  if (!Array.isArray(rules) || rules.length === 0) return '无规则，使用默认运费'
  return rules
    .slice(0, 4)
    .map((rule) => `${rule.region || '-'} ${Number(rule.min_value || 0)}-${Number(rule.max_value || 0) || '∞'}: ${formatMoney(rule.fee)}`)
    .join('；')
}

export const compactListLabel = (value: any) => {
  if (!value) return '-'
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) {
      return parsed.length ? parsed.join(', ') : '-'
    }
  } catch {
    // Keep raw value.
  }
  return String(value)
}

export const serviceAreaLabel = (value: any) => {
  if (!value) return '-'
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) return parsed.join(', ')
  } catch {
    // Keep raw value.
  }
  return String(value)
}

export const carrierServiceCarrierName = (service: any, carriers: any[] = []) => {
  if (service.carrier?.name) return service.carrier.name
  const carrier = carriers.find((item) => Number(item.id) === Number(service.carrier_id))
  return carrier?.name || `Carrier #${service.carrier_id || '-'}`
}

export const carrierServiceTemplateName = (service: any, templates: any[] = []) => {
  if (service.template?.name) return service.template.name
  if (!service.template_id) return '未绑定模板'
  const template = templates.find((item) => Number(item.id) === Number(service.template_id))
  return template?.name || `Template #${service.template_id}`
}

export const billingModeLabel = (mode: string) => {
  const labels: Record<string, string> = {
    actual_weight: '实重计费',
    volumetric_weight: '体积重计费',
    greater_of_actual_and_volumetric: '实重/体积重取大',
  }
  return labels[mode] || mode || '-'
}

export const trackingEnvironmentLabel = (environment: string) => {
  const labels: Record<string, string> = {
    production: 'Production',
    sandbox: 'Sandbox',
  }
  return labels[environment] || environment || '-'
}

const apiBaseUrl = () => {
  const configured = String(import.meta.env.VITE_API_BASE_URL || '').trim().replace(/\/+$/, '')
  if (/^https?:\/\//i.test(configured)) {
    try {
      return new URL(configured).origin
    } catch {
      return configured
    }
  }
  if (typeof window !== 'undefined' && window.location?.origin) return window.location.origin
  return ''
}

export const trackingWebhookUrl = (provider: any) => {
  const providerCode = String(provider?.provider_code || '').trim()
  if (!providerCode) return ''

  const path = `/api/v1/shipping/webhook/${encodeURIComponent(providerCode)}`
  const base = apiBaseUrl()
  return base ? `${base}${path}` : path
}

export const trackingProviderHasApiKey = (provider: any) => provider?.api_key_configured === true || Boolean(provider?.api_key)
export const trackingProviderHasWebhookSecret = (provider: any) => provider?.webhook_secret_configured === true || Boolean(provider?.webhook_secret)

export const formatTrackingSyncPolicy = (provider: any) => {
  const policies: string[] = []
  if (provider.auto_register) policies.push('自动注册追踪号')
  if (provider.webhook_enabled) policies.push('Webhook 推送')
  if (provider.polling_enabled) policies.push(`轮询 ${Number(provider.polling_interval_minutes || 60)} 分钟`)
  if (Number(provider.request_timeout_seconds || 0) > 0) policies.push(`超时 ${Number(provider.request_timeout_seconds)} 秒`)
  return policies.length ? policies.join(' / ') : '暂未启用同步策略'
}

export const trackingProviderName = (mapping: any, trackingProviders: any[] = []) => {
  if (mapping.provider?.provider_name) return mapping.provider.provider_name
  const provider = trackingProviders.find((item) => Number(item.id) === Number(mapping.provider_id))
  return provider?.provider_name || `Provider #${mapping.provider_id || '-'}`
}

export const trackingMappingScopeLabel = (scope: string) => {
  const labels: Record<string, string> = {
    carrier: '承运商映射',
    carrier_service: '线路服务映射',
  }
  return labels[scope] || scope || '-'
}

export const trackingMappingLocalTargetLabel = (mapping: any, carriers: any[] = [], carrierServices: any[] = []) => {
  if (mapping.scope === 'carrier_service') {
    if (mapping.carrier_service?.service_name) return mapping.carrier_service.service_name
    const service = carrierServices.find((item) => Number(item.id) === Number(mapping.carrier_service_id))
    return service?.service_name || `Carrier service #${mapping.carrier_service_id || '-'}`
  }

  if (mapping.carrier?.name) return mapping.carrier.name
  const carrier = carriers.find((item) => Number(item.id) === Number(mapping.carrier_id))
  return carrier?.name || `Carrier #${mapping.carrier_id || '-'}`
}

export const formatGrams = (value: any) => `${Number(value || 0).toLocaleString()} g`

export const formatServiceWeightStep = (service: any) => {
  const first = Number(service.first_weight_grams || 0)
  const additional = Number(service.additional_weight_grams || 0)
  const min = Number(service.min_charge_weight_grams || 0)
  const parts: string[] = []
  if (first > 0) parts.push(`首 ${formatGrams(first)}`)
  if (additional > 0) parts.push(`续 ${formatGrams(additional)}`)
  if (min > 0) parts.push(`最低 ${formatGrams(min)}`)
  return parts.length ? parts.join(' / ') : '未设置'
}

export const formatVolumetricDivisor = (service: any) => {
  const divisor = Number(service.volumetric_divisor || 0)
  const surcharge = Number(service.fuel_surcharge_percent || 0)
  const remote = Number(service.remote_surcharge || 0)
  const parts: string[] = []
  if (divisor > 0) parts.push(`÷${divisor}`)
  if (surcharge > 0) parts.push(`燃油 ${surcharge.toFixed(3)}%`)
  if (remote > 0) parts.push(`偏远 ${formatMoney(remote)}`)
  return parts.length ? parts.join(' / ') : '未设置'
}

export const formatEta = (service: any) => {
  const min = Number(service.eta_min_days || 0)
  const max = Number(service.eta_max_days || 0)
  if (min > 0 && max > 0) return `${min}-${max} 天`
  if (min > 0) return `${min}+ 天`
  if (max > 0) return `≤ ${max} 天`
  return '-'
}

export const formatWeight = (value: any) => `${Number(value || 0).toFixed(3)} kg`

export const formatDimensions = (rule: any) => {
  const length = Number(rule.box_length || 0).toFixed(2)
  const width = Number(rule.box_width || 0).toFixed(2)
  const height = Number(rule.box_height || 0).toFixed(2)
  return `${length} × ${width} × ${height} cm`
}

export const appliesCount = (rule: any) => Array.isArray(rule.applies) ? rule.applies.length : 0

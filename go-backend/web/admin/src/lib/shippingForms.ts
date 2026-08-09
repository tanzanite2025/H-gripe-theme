const firstId = (items: any[] = []) => items[0]?.id ? String(items[0].id) : ''

export const defaultShippingCarrierForm = () => ({
  id: null,
  name: '',
  code: '',
  tracking_url: '',
  contact: '',
  phone: '',
  email: '',
  service_area: '',
  enabled: true,
  sort_order: 0,
})

export const defaultShippingCarrierServiceForm = (carriers: any[] = []) => ({
  id: null,
  carrier_id: firstId(carriers),
  template_id: 'none',
  service_code: '',
  service_name: '',
  route_name: '',
  countries: '[]',
  currency: '',
  billing_mode: 'actual_weight',
  first_weight_grams: 0,
  additional_weight_grams: 0,
  min_charge_weight_grams: 0,
  volumetric_divisor: 6000,
  fuel_surcharge_percent: 0,
  remote_surcharge: 0,
  eta_min_days: 0,
  eta_max_days: 0,
  enabled: true,
  sort_order: 0,
  description: '',
})

export const defaultShippingTrackingProviderForm = () => ({
  id: null,
  provider_code: '17TRACK',
  provider_name: '17TRACK',
  environment: 'production',
  base_url: '',
  api_key: '',
  webhook_secret: '',
  webhook_enabled: false,
  auto_register: false,
  polling_enabled: false,
  polling_interval_minutes: 60,
  request_timeout_seconds: 15,
  enabled: true,
  sort_order: 0,
  description: '',
})

export const defaultShippingTrackingCarrierMappingForm = (
  trackingProviders: any[] = [],
  carriers: any[] = [],
  carrierServices: any[] = []
) => ({
  id: null,
  provider_id: firstId(trackingProviders),
  scope: 'carrier',
  carrier_id: firstId(carriers),
  carrier_service_id: firstId(carrierServices),
  provider_carrier_code: '',
  provider_carrier_name: '',
  enabled: true,
  priority: 0,
  description: '',
})

export const defaultShippingTemplateForm = () => ({
  id: null,
  name: '',
  type: 'weight',
  free_shipping: false,
  free_threshold: 0,
  default_fee: 0,
  display_price_snapshots: {},
  description: '',
  enabled: true,
  rules: [],
})

export const defaultShippingZoneForm = () => ({
  id: null,
  name: '',
  countries: '[]',
  states: '[]',
  postal_codes: '[]',
  enabled: true,
})

export const defaultShippingPackagingForm = () => ({
  id: null,
  rule_name: '',
  description: '',
  box_weight: 0,
  box_length: 0,
  box_width: 0,
  box_height: 0,
  max_weight: 0,
  is_active: true,
})

export const resetReactive = (target: Record<string, any>, defaults: Record<string, any>) => {
  Object.keys(target).forEach((key) => delete target[key])
  Object.assign(target, defaults)
}

export const clearErrors = (errors: Record<string, any>) => {
  Object.keys(errors).forEach((key) => delete errors[key])
}

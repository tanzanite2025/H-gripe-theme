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
  template_kind: 'carrier',
  is_system_managed: false,
  free_shipping_countries: '[]',
  currency: '',
  free_shipping: false,
  free_threshold_minor: 0,
  default_fee_minor: 0,
  display_price_snapshots: {},
  description: '',
  enabled: true,
  route_group_id: null,
  customer_billing_mode: 'actual_weight',
  customer_first_weight_grams: 0,
  customer_additional_weight_grams: 0,
  customer_min_charge_weight_grams: 0,
  customer_volumetric_divisor: 6000,
  customer_fuel_surcharge_percent_decimal: '0',
  customer_remote_surcharge_minor: 0,
  customer_remote_postal_codes: '[]',
  customer_eta_min_days: 0,
  customer_eta_max_days: 0,
  rules: [],
})

export const defaultShippingRouteAggregateForm = () => ({
  id: null,
  name: '',
  description: '',
  countries: '[]',
  enabled: true,
  sort_order: 0,
  carrier_service_ids: [],
  published_carrier_services: [],
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

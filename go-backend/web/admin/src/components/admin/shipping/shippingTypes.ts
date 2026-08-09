export type ShippingID = string | number
export type ShippingDialogMode = 'create' | 'edit'
export type ShippingErrorMap = Record<string, string | undefined>

export interface ShippingDisplayPrice {
  amount?: number | string | null
  currency?: string | null
  quote_currency?: string | null
  rate?: number | string | null
  source?: string | null
  converted?: boolean
  fallback_reason?: string | null
}

export type ShippingDisplayPriceSnapshots = Record<string, ShippingDisplayPrice[]>

export interface ShippingTemplateRule {
  id?: ShippingID | null
  region?: string | null
  min_value?: number | string | null
  max_value?: number | string | null
  fee?: number | string | null
  additional?: number | string | null
  display_price_snapshots?: ShippingDisplayPriceSnapshots
}

export interface ShippingTemplate {
  id: ShippingID
  name?: string | null
  type?: string | null
  free_shipping?: boolean
  free_threshold?: number | string | null
  default_fee?: number | string | null
  display_price_snapshots?: ShippingDisplayPriceSnapshots
  description?: string | null
  enabled?: boolean
  rules?: ShippingTemplateRule[]
}

export interface ShippingTemplateForm {
  id: ShippingID | null
  name: string
  type: string
  free_shipping: boolean
  free_threshold: number
  default_fee: number
  display_price_snapshots: ShippingDisplayPriceSnapshots
  description: string
  enabled: boolean
  rules: ShippingTemplateRule[]
}

export interface ShippingZone {
  id: ShippingID
  name?: string | null
  countries?: unknown
  states?: unknown
  postal_codes?: unknown
  enabled?: boolean
}

export interface ShippingZoneForm {
  id: ShippingID | null
  name: string
  countries: string
  states: string
  postal_codes: string
  enabled: boolean
}

export interface ShippingCarrier {
  id: ShippingID
  code?: string | null
  name?: string | null
  tracking_url?: string | null
  contact?: string | null
  phone?: string | null
  email?: string | null
  service_area?: unknown
  enabled?: boolean
  sort_order?: number | string | null
}

export interface ShippingCarrierForm {
  id: ShippingID | null
  name: string
  code: string
  tracking_url: string
  contact: string
  phone: string
  email: string
  service_area: string
  enabled: boolean
  sort_order: number
}

export interface ShippingCarrierService {
  id: ShippingID
  carrier_id?: ShippingID | null
  service_code?: string | null
  service_name?: string | null
  route_name?: string | null
  countries?: unknown
  template_id?: ShippingID | null
  billing_mode?: string | null
  first_weight_grams?: number | string | null
  additional_weight_grams?: number | string | null
  min_charge_weight_grams?: number | string | null
  volumetric_divisor?: number | string | null
  fuel_surcharge_percent?: number | string | null
  remote_surcharge?: number | string | null
  eta_min_days?: number | string | null
  eta_max_days?: number | string | null
  enabled?: boolean
}

export interface ShippingCarrierServiceForm {
  id: ShippingID | null
  carrier_id: string
  template_id: string
  service_code: string
  service_name: string
  route_name: string
  countries: string
  currency: string
  billing_mode: string
  first_weight_grams: number
  additional_weight_grams: number
  min_charge_weight_grams: number
  volumetric_divisor: number
  fuel_surcharge_percent: number
  remote_surcharge: number
  eta_min_days: number
  eta_max_days: number
  enabled: boolean
  sort_order: number
  description: string
}

export interface TrackingProvider {
  id: ShippingID
  provider_code?: string | null
  provider_name?: string | null
  environment?: string | null
  base_url?: string | null
  api_key?: string | null
  api_key_configured?: boolean
  webhook_secret?: string | null
  webhook_secret_configured?: boolean
  webhook_enabled?: boolean
  auto_register?: boolean
  polling_enabled?: boolean
  polling_interval_minutes?: number | string | null
  request_timeout_seconds?: number | string | null
  enabled?: boolean
  sort_order?: number | string | null
  description?: string | null
}

export interface TrackingProviderForm {
  id: ShippingID | null
  provider_code: string
  provider_name: string
  environment: string
  base_url: string
  api_key: string
  webhook_secret: string
  webhook_enabled: boolean
  auto_register: boolean
  polling_enabled: boolean
  polling_interval_minutes: number
  request_timeout_seconds: number
  enabled: boolean
  sort_order: number
  description: string
  api_key_configured?: boolean
  webhook_secret_configured?: boolean
}

export interface TrackingCarrierMapping {
  id: ShippingID
  provider_id?: ShippingID | null
  provider?: TrackingProvider | null
  scope?: string | null
  carrier_id?: ShippingID | null
  carrier_service_id?: ShippingID | null
  provider_carrier_code?: string | null
  provider_carrier_name?: string | null
  enabled?: boolean
  priority?: number | string | null
  description?: string | null
}

export interface TrackingCarrierMappingForm {
  id: ShippingID | null
  provider_id: string
  scope: string
  carrier_id: string
  carrier_service_id: string
  provider_carrier_code: string
  provider_carrier_name: string
  enabled: boolean
  priority: number
  description: string
}

export interface PackagingRule {
  id: ShippingID
  rule_name?: string | null
  description?: string | null
  box_weight?: number | string | null
  box_length?: number | string | null
  box_width?: number | string | null
  box_height?: number | string | null
  max_weight?: number | string | null
  is_active?: boolean
  applies?: unknown[]
}

export interface PackagingRuleForm {
  id: ShippingID | null
  rule_name: string
  description: string
  box_weight: number
  box_length: number
  box_width: number
  box_height: number
  max_weight: number
  is_active: boolean
}

export interface ShippingQuoteItemInput {
  product_id: number | string
  variant_id: number | string | null
  quantity: number | string
}

export interface ShippingQuoteForm {
  country: string
  currency: string
  items: ShippingQuoteItemInput[]
}

export interface ShippingQuoteOption {
  carrier_service_id?: ShippingID | null
  service_name?: string | null
  carrier_name?: string | null
  service_code?: string | null
  template_id?: ShippingID | null
  billing_mode?: string | null
  actual_weight_grams?: number | string | null
  volumetric_weight_grams?: number | string | null
  billable_weight_grams?: number | string | null
  base_fee?: number | string | null
  fuel_surcharge?: number | string | null
  remote_surcharge?: number | string | null
  shipping_fee?: number | string | null
  eta_min_days?: number | string | null
  eta_max_days?: number | string | null
}

export interface ShippingQuoteItemResult {
  product_id?: ShippingID | null
  variant_id?: ShippingID | null
  template_id?: ShippingID | null
  template_name?: string | null
  quantity?: number | string | null
  unit_price?: number | string | null
  weight_grams?: number | string | null
  packaging_rule_id?: ShippingID | null
  packaging_rule_name?: string | null
  packaging_weight_grams?: number | string | null
  charge_weight_grams?: number | string | null
  shipping_fee?: number | string | null
}

export interface ShippingQuoteResult {
  shipping_fee?: number | string | null
  currency?: string | null
  free_shipping?: boolean
  source?: string | null
  selected_option?: ShippingQuoteOption | null
  options?: ShippingQuoteOption[]
  items?: ShippingQuoteItemResult[]
}

export interface TrackingShipment {
  id?: ShippingID | null
  order_id?: ShippingID | null
  tracking_number?: string | null
  provider_carrier_code?: string | null
  tracking_provider_id?: ShippingID | null
  carrier_id?: ShippingID | null
  carrier_service_id?: ShippingID | null
  provider?: TrackingProvider | null
  carrier?: ShippingCarrier | null
  carrier_service?: ShippingCarrierService | null
  sync_status?: string | null
  registration_status?: string | null
  event_count?: number | string | null
  last_synced_at?: string | number | Date | null
  next_sync_at?: string | number | Date | null
  last_error?: string | null
  enabled?: boolean
}

export interface TrackingEvent {
  id?: ShippingID | null
  tracking_number?: string | null
  provider_carrier_code?: string | null
  event_time?: string | number | Date | null
  status?: string | null
  location?: string | null
  description?: string | null
  created_at?: string | number | Date | null
}

export interface TrackingShipmentFilters {
  sync_status: string
  registration_status: string
  provider_id: string
  carrier_id: string
  carrier_service_id: string
  enabled: string
  due_only: string
  keyword: string
  limit: string
}

export interface TrackingPollingError {
  order_id?: ShippingID | null
  tracking_number?: string | null
  error?: string | null
}

export interface TrackingPollingState {
  enabled?: boolean
  running?: boolean
  interval_seconds?: number | string | null
  interval?: string | null
  batch_limit?: number | string | null
  last_started_at?: string | number | Date | null
  last_finished_at?: string | number | Date | null
  last_matched?: number | string | null
  last_synced?: number | string | null
  last_failed?: number | string | null
  last_duration_ms?: number | string | null
  last_error?: string | null
  last_errors?: TrackingPollingError[]
}

export interface TrackingWebhookState {
  last_accepted?: boolean
  last_error?: string | null
  last_received_at?: string | number | Date | null
  last_http_status?: number | string | null
  last_duration_ms?: number | string | null
  last_signature_checked?: boolean
  last_signature_valid?: boolean
  last_provider_code?: string | null
  last_tracking_number?: string | null
  last_order_id?: ShippingID | null
  last_carrier_code?: string | null
  last_event_count?: number | string | null
}

export interface ShippingLoadingState {
  templates: boolean
  zones: boolean
  carriers: boolean
  services: boolean
  tracking: boolean
  trackingMappings: boolean
  trackingShipments: boolean
  packaging: boolean
}

export type ShippingResource =
  | ShippingTemplate
  | ShippingZone
  | ShippingCarrier
  | ShippingCarrierService
  | TrackingProvider
  | TrackingCarrierMapping
  | PackagingRule

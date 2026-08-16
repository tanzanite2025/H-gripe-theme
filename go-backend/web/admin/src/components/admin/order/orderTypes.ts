import type { Component } from 'vue'

export type OrderID = number | string
export type OrderStatusTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'
export type OrderSelectionState = boolean | 'indeterminate'

export interface OrderStatusOption {
  label: string
  value: string
}

export interface OrderFilters {
  search: string
  status: string
  payment_status: string
  shipping_status: string
  start_date: string
  end_date: string
}

export interface OrderPagination {
  page: number
  pageSize: number
  total: number
}

export interface OrderShippingAddress {
  first_name?: string | null
  last_name?: string | null
  phone?: string | null
  email?: string | null
  address_1?: string | null
  address_2?: string | null
  city?: string | null
  state?: string | null
  postal_code?: string | null
  country?: string | null
}

export interface OrderItem {
  id?: OrderID | null
  product_name?: string | null
  sku?: string | null
  price?: number | string | null
  quantity?: number | string | null
  total?: number | string | null
  hs_code?: string | null
  cn_code?: string | null
  country_of_origin?: string | null
  customs_description?: string | null
  declared_value?: number | string | null
  declared_value_confirmed?: boolean | null
}

export interface OrderRecord {
  id: OrderID
  order_number?: string | null
  status?: string | null
  payment_status?: string | null
  shipping_status?: string | null
  payment_method?: string | null
  shipping_method?: string | null
  tracking_number?: string | null
  tracking_provider_id?: OrderID | null
  carrier_id?: OrderID | null
  carrier_service_id?: OrderID | null
  provider_carrier_code?: string | null
  created_at?: string | null
  paid_at?: string | null
  shipping_address?: OrderShippingAddress | null
  items?: OrderItem[]
  subtotal_amount?: number | string | null
  shipping_fee?: number | string | null
  tax_amount?: number | string | null
  discount_amount?: number | string | null
  total_amount?: number | string | null
  currency?: string | null
  customer_note?: string | null
  admin_note?: string | null
}

export interface OrderStats {
  total?: number
  today?: number
  total_revenue?: number | string | null
  today_revenue?: number | string | null
}

export interface OrderStatItem {
  key: string
  label: string
  value: string | number
  icon: Component
  tone: OrderStatusTone
}

export interface OrderStatusForm {
  id: OrderID | null
  order_number: string
  status: string
  shipping_status: string
  tracking_number: string
  tracking_provider_id: string
  carrier_id: string
  carrier_service_id: string
}

export interface ShippingCarrier {
  id: OrderID
  name?: string | null
  code?: string | null
}

export interface ShippingCarrierService {
  id: OrderID
  carrier_id?: OrderID | null
  service_name?: string | null
  service_code?: string | null
}

export interface TrackingProvider {
  id: OrderID
  provider_name?: string | null
  provider_code?: string | null
}

export interface TrackingCarrierMapping {
  id?: OrderID | null
  provider_id?: OrderID | null
  scope?: string | null
  carrier_id?: OrderID | null
  carrier_service_id?: OrderID | null
  provider_carrier_code?: string | null
  provider_carrier_name?: string | null
}

export interface TrackingEvent {
  id?: OrderID | null
  tracking_number?: string | null
  event_time?: string | null
  status?: string | null
  location?: string | null
  description?: string | null
}

export interface TrackingShipment {
  sync_status?: string | null
  registration_status?: string | null
  event_count?: number | null
  last_synced_at?: string | null
  next_sync_at?: string | null
  last_error?: string | null
}

export interface OrderConfirmation {
  open: boolean
  type: '' | 'delete' | 'batch-status'
  target: OrderRecord | OrderRecord[] | null
  status: string
  title: string
  description: string
  confirmLabel: string
  destructive: boolean
}

export interface OrderDisputeEvidenceSummary {
  complete?: boolean
  ready_count?: number
  total_count?: number
  missing_count?: number
  manual_required_count?: number
  unavailable_count?: number
  blocker_count?: number
  missing_items?: string[]
  manual_items?: string[]
  last_evaluated_at?: string | null
}

export interface OrderDisputeMistakeAssessment {
  level?: string
  label?: string
  reason?: string
  signals?: string[]
}

export interface OrderDisputeContactDraft {
  can_send?: boolean
  to?: string
  subject?: string
  body?: string
  mailto_url?: string
}

export interface OrderDisputeCase {
  provider: 'stripe' | 'paypal' | string
  dispute_id: OrderID
  provider_dispute_id?: string | null
  provider_payment_id?: string | null
  order_id?: OrderID | null
  order_number?: string | null
  customer_name?: string | null
  customer_email?: string | null
  order_status?: string | null
  payment_status?: string | null
  shipping_status?: string | null
  tracking_number?: string | null
  amount?: number | string | null
  currency?: string | null
  reason?: string | null
  status?: string | null
  state?: string | null
  life_cycle_stage?: string | null
  evidence_due_at?: string | null
  evidence_submitted_at?: string | null
  created_at?: string | null
  updated_at?: string | null
  has_delivered_event?: boolean
  delivered_at?: string | null
  needs_response?: boolean
  evidence_summary?: OrderDisputeEvidenceSummary
  submission_ready?: boolean
  submission_blockers?: string[]
  warnings?: string[]
  mistake_assessment?: OrderDisputeMistakeAssessment
  suggested_action?: string | null
  contact_draft?: OrderDisputeContactDraft
}

export interface OrderDisputeAnalysisSummary {
  total?: number
  needs_response?: number
  evidence_blocked?: number
  evidence_submitted?: number
  likely_mistake?: number
  missing_email?: number
}

export interface OrderDisputeAnalysis {
  order?: OrderRecord | null
  disputes?: OrderDisputeCase[]
  summary?: OrderDisputeAnalysisSummary
  contact_draft?: OrderDisputeContactDraft | null
}

export interface OrderDisputeEmailForm {
  provider: string
  dispute_id: OrderID | null
  to: string
  subject: string
  body: string
}

export type OrderStatusNameResolver = (status?: string | null) => string
export type OrderStatusToneResolver = (status?: string | null) => OrderStatusTone
export type OrderMoneyFormatter = (amount?: number | string | null, currency?: string | null) => string
export type OrderDateFormatter = (value?: string | number | Date | null) => string
export type OrderShippingNameResolver = (address?: OrderShippingAddress | null) => string
export type OrderShippingAddressLineResolver = (address?: OrderShippingAddress | null) => string
export type OrderCarrierLabelResolver = (order?: OrderRecord | null) => string

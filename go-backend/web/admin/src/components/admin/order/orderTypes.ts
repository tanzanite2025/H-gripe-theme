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

export type OrderStatusNameResolver = (status?: string | null) => string
export type OrderStatusToneResolver = (status?: string | null) => OrderStatusTone
export type OrderMoneyFormatter = (amount?: number | string | null) => string
export type OrderDateFormatter = (value?: string | number | Date | null) => string
export type OrderShippingNameResolver = (address?: OrderShippingAddress | null) => string
export type OrderShippingAddressLineResolver = (address?: OrderShippingAddress | null) => string
export type OrderCarrierLabelResolver = (order?: OrderRecord | null) => string

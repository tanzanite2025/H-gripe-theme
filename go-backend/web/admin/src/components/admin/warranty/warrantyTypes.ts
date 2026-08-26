import type {
  APIID,
  WarrantyClaim as ApiWarrantyClaim,
  WarrantyOrderItem as ApiWarrantyOrderItem,
  WarrantyServiceRecord as ApiWarrantyServiceRecord,
  ShipmentRecord as ApiShipmentRecord,
  ShipmentItemSnapshot as ApiShipmentItemSnapshot
} from '@/api/warranty'

export type WarrantyID = APIID

export interface WarrantyStatusOption {
  value: string
  label: string
}

export interface WarrantyUser {
  id?: WarrantyID | null
  first_name?: string | null
  last_name?: string | null
  username?: string | null
  email?: string | null
}

export interface WarrantyProduct {
  id?: WarrantyID | null
  name?: string | null
  sku?: string | null
}

export interface WarrantyOrderItem extends ApiWarrantyOrderItem {
  id?: WarrantyID | null
  product_id?: WarrantyID | null
  product_name?: string | null
  sku?: string | null
  variant_id?: WarrantyID | null
  quantity?: number | string | null
}

export interface WarrantyServiceRecord extends ApiWarrantyServiceRecord {
  id?: WarrantyID | null
  service_type?: string | null
  status?: string | null
  summary?: string | null
  cost_amount?: number | string | null
  currency?: string | null
  performed_at?: string | null
  created_at?: string | null
}

export interface WarrantyShipmentItem extends ApiShipmentItemSnapshot {
  id?: WarrantyID | null
  product_id?: WarrantyID | null
  product_name?: string | null
  sku?: string | null
  variant_id?: WarrantyID | null
  quantity?: number | string | null
}

export interface WarrantyShipmentRecord extends ApiShipmentRecord {
  id?: WarrantyID | null
  order_id?: WarrantyID | null
  order_number?: string | null
  user_id?: WarrantyID | null
  customer_name?: string | null
  customer_email?: string | null
  tracking_number?: string | null
  shipped_at?: string | null
  items_snapshot?: WarrantyShipmentItem[]
  product_codes?: string[]
  shipping_note?: string | null
  shipping_images?: string[]
  warranty_months?: number | null
  warranty_start_at?: string | null
  warranty_expires?: string | null
  status?: string | null
  record_bound?: boolean
  order_status?: string | null
  shipping_status?: string | null
}

export interface WarrantyShipmentDraft {
  shippingNote: string
  shippingImages: string[]
  productCodes: string[]
  warrantyMonths: number
  warrantyStart: string
}

export interface WarrantyClaim extends ApiWarrantyClaim {
  id?: WarrantyID | null
  status?: string | null
  email?: string | null
  order_number?: string | null
  issue_type?: string | null
  description?: string | null
  tire_pressure?: string | number | null
  is_tubeless?: boolean | null
  images?: string | null
  video_url?: string | null
  created_at?: string | null
  processed_at?: string | null
  resolution?: string | null
  order_item_id?: WarrantyID | null
  order_item?: WarrantyOrderItem | null
  service_records?: WarrantyServiceRecord[]
}

export interface WarrantyFilters {
  status: string
  keyword?: string
}

export interface WarrantyPagination {
  page: number
  pageSize: number
  total: number
}

export interface WarrantyStatusUpdating {
  claim: WarrantyID | null
}

export interface WarrantyServiceRecordForm {
  serviceType: string
  status: string
  summary: string
  costAmount: string
  currency: string
  performedAt: string
}

import type {
  APIID,
  WarrantyClaim as ApiWarrantyClaim,
  WarrantyOrderItem as ApiWarrantyOrderItem,
  WarrantyRegistration as ApiWarrantyRegistration,
  WarrantyServiceRecord as ApiWarrantyServiceRecord
} from '@/api/registrations'

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

export interface WarrantyRegistration extends ApiWarrantyRegistration {
  id?: WarrantyID | null
  status?: string | null
  product_id?: WarrantyID | null
  product?: WarrantyProduct | null
  serial_number?: string | null
  user?: WarrantyUser | null
  purchase_date?: string | null
  warranty_expires?: string | null
  purchase_proof?: string | null
  created_at?: string | null
  updated_at?: string | null
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

export interface WarrantyClaimRegistration {
  id?: WarrantyID | null
  product?: WarrantyProduct | null
  user?: WarrantyUser | null
}

export interface WarrantyClaim extends ApiWarrantyClaim {
  id?: WarrantyID | null
  status?: string | null
  registration_id?: WarrantyID | null
  registration?: WarrantyClaimRegistration | null
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
}

export interface WarrantyPagination {
  page: number
  pageSize: number
  total: number
}

export interface WarrantyStatusUpdating {
  registration: WarrantyID | null
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

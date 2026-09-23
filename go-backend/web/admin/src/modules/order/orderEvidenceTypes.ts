import type { OrderID } from '@/modules/order/orderTypes'

export type OrderEvidencePackageStatus = 'incomplete' | 'ready' | 'locked' | 'superseded' | string
export type OrderEvidenceItemStatus = 'missing' | 'draft' | 'complete' | 'waived' | string

export type OrderEvidenceItemType =
  | 'configuration_confirmation'
  | 'product_identity'
  | 'outbound_weight_packaging'
  | 'signed_pod'
  | 'spoke_qc_tension'
  | string

export interface OrderEvidenceAttachment {
  id: OrderID
  evidence_item_id?: OrderID | null
  storage_key?: string | null
  original_filename?: string | null
  mime_type?: string | null
  size_bytes?: number | null
  sha256?: string | null
  uploaded_by?: OrderID | null
  created_at?: string | null
}

export interface OrderEvidenceItem {
  id: OrderID
  package_id?: OrderID | null
  order_id?: OrderID | null
  order_item_id?: OrderID | null
  snapshot_id?: OrderID | null
  item_type: OrderEvidenceItemType
  status: OrderEvidenceItemStatus
  required_reason: string
  data_json?: unknown
  captured_at?: string | null
  captured_by?: OrderID | null
  content_sha256?: string | null
  attachments?: OrderEvidenceAttachment[]
  created_at?: string | null
  updated_at?: string | null
}

export interface OrderEvidencePackage {
  id: OrderID
  order_id: OrderID
  snapshot_id: OrderID
  package_version: number
  status: OrderEvidencePackageStatus
  order_total_usd_snapshot: number
  is_high_value: boolean
  has_spoke_tension_qc: boolean
  schema_version: number
  created_by?: OrderID | null
  locked_at?: string | null
  created_at?: string | null
  updated_at?: string | null
  items?: OrderEvidenceItem[]
}

export interface OrderEvidenceCompleteness {
  total: number
  complete: number
  waived: number
  pending: number
  satisfied: number
  percent: number
  ready: boolean
}

export interface OrderEvidenceTrackingShipment {
  id: OrderID
  tracking_number: string
  provider_carrier_code: string
  provider_code?: string | null
  provider_name?: string | null
  carrier_name?: string | null
  carrier_service_name?: string | null
  registration_status: string
  sync_status: string
  event_count: number
  last_event_at?: string | null
  last_synced_at?: string | null
  enabled: boolean
}

export interface OrderEvidenceDeliveryEvent {
  id: OrderID
  tracking_number: string
  status: string
  location?: string | null
  description?: string | null
  recipient_signature_name?: string | null
  proof_of_delivery_url?: string | null
  event_time: string
}

export interface OrderEvidenceManualPOD {
  item_id: OrderID
  status: OrderEvidenceItemStatus
  tracking_number?: string | null
  tracking_association_status: 'matched' | 'mismatch' | 'not_recorded' | 'shipment_unavailable' | string
  captured_at?: string | null
  captured_by?: OrderID | null
  attachment_count: number
}

export interface OrderEvidenceTrackingContext {
  shipments?: OrderEvidenceTrackingShipment[]
  latest_delivery_events?: OrderEvidenceDeliveryEvent[]
  provider_pod_urls?: string[]
  manual_pod?: OrderEvidenceManualPOD | null
}

export interface OrderEvidenceSourceReference {
  source_type: string
  source_id: OrderID
  order_item_id?: OrderID | null
  evidence_item_type?: string | null
  status?: string | null
  content_sha256?: string | null
  captured_at?: string | null
}

export interface OrderEvidencePackageResult {
  package: OrderEvidencePackage
  completeness: OrderEvidenceCompleteness
  tracking_context?: OrderEvidenceTrackingContext | null
  sources?: OrderEvidenceSourceReference[]
  warnings?: string[]
}

export interface OrderEvidenceOrderSummary {
  order_id: OrderID
  order_number: string
  customer_first_name?: string | null
  customer_last_name?: string | null
  customer_email?: string | null
  order_status?: string | null
  payment_status?: string | null
  shipping_status?: string | null
  total_amount?: number | string | null
  currency?: string | null
  created_at?: string | null
  package_id?: OrderID | null
  package_version?: number | null
  package_status?: OrderEvidencePackageStatus | null
  order_total_usd_snapshot?: number | string | null
  is_high_value?: boolean
  has_spoke_tension_qc?: boolean
  total_evidence_items?: number
  complete_evidence_items?: number
  waived_evidence_items?: number
  pending_evidence_items?: number
}

export interface OrderEvidenceListParams {
  page: number
  page_size: number
  search?: string
  package_status?: string
  high_value?: string
  spoke_tension_qc?: string
}

export interface OrderEvidenceListResult {
  data: OrderEvidenceOrderSummary[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface OrderEvidenceItemUpdateInput {
  status: OrderEvidenceItemStatus
  data_json: unknown
  captured_at: string | null
}

export interface OrderFulfillmentEvidenceDraft {
  item_id: OrderID
  data_json: unknown
  captured_at: string
  files: File[]
}

export interface OrderFulfillmentSubmitInput {
  tracking_number: string
  tracking_provider_id: OrderID
  carrier_id: OrderID | null
  carrier_service_id: OrderID | null
  signature_confirmed: boolean
  evidence: OrderFulfillmentEvidenceDraft[]
}

export interface AssignableAgent {
  id?: string | number
  user_id?: string | number
  name?: string
  email?: string
}

export interface CustomerServiceFiltersState {
  search: string
  view: string
  status: string
  identity: string
  unread: string
  [key: string]: string
}

export interface CustomerMemberTier {
  name?: string
  icon?: string
  color?: string
  total_points?: number | string
}

export interface CustomerSummary {
  identity?: string
  type?: string
  display_name?: string
  identity_label?: string
  region_label?: string
  member_tier?: CustomerMemberTier | null
}

export interface CustomerConversation {
  id: string | number
  customer_name?: string
  customer_summary?: CustomerSummary
  visitor_anonymous?: boolean
  display_status?: string
  status?: string
  status_version?: number
  inbox_archived?: boolean
  archived_at?: string | number | Date | null
  conversation_id?: string
  ticket_number?: string | number
  assigned_to?: string | number | null
  unread_count?: number
  last_message?: string
}

export type CustomerConversationStatus = 'open' | 'in_progress' | 'resolved' | 'closed'

export interface CustomerConversationStatusResult {
  id: string | number
  status: CustomerConversationStatus
  status_version: number
}

export interface CustomerTypingState {
  active: boolean
  displayName?: string
}

export type CustomerTypingByConversation = Record<string, CustomerTypingState>

export interface MessageMetadata {
  url?: string
  title?: string
  [key: string]: unknown
}

export interface CustomerConversationMessage {
  id: string | number
  is_agent?: boolean
  sender_name?: string
  created_at?: string | number | Date | null
  message_type?: string
  content?: string
  message?: string
  metadata?: MessageMetadata | null
  attachments?: unknown[]
  attachment_url?: string | null
}

export interface CustomerServiceSendMessagePayload {
  message: string
  messageType?: string
  metadata?: unknown
  attachmentUrl?: string
  attachments?: string[]
  toastLabel?: string
  clearReplyMessage?: boolean
}

export interface CustomerPagination {
  page: number
  pageSize: number
  total: number
}

export interface CustomerAccount {
  id: string | number
  display_name?: string
  username?: string
  email?: string
  member_tier?: CustomerMemberTier | null
  locale?: string
  status?: string
  created_at?: string | number | Date | null
}

export interface CustomerAnonymous {
  note?: string
  visitor_hash_preview?: string
}

export interface CustomerContact {
  email?: string
  email_source?: string
  locale?: string
  locale_source?: string
  timezone?: string
  timezone_source?: string
}

export interface CustomerCartItem {
  id: string | number
  image?: string
  name?: string
  sku?: string
  variant_name?: string
  quantity?: number | string
  currency?: string
  price?: number | string
  line_total?: number | string
  inventory_snapshot?: string
}

export interface CustomerCart {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  item_count?: number | string
  currency?: string
  total?: number | string
  items?: CustomerCartItem[]
}

export interface CustomerWishlistItem {
  id: string | number
  image?: string
  name?: string
  sku?: string
  product_id?: string | number
}

export interface CustomerWishlist {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  count?: number | string
  items?: CustomerWishlistItem[]
}

export interface CustomerOrderItem {
  id: string | number
  order_number?: string
  total_amount?: number | string
  status?: string
  payment_status?: string
  shipping_status?: string
  created_at?: string | number | Date | null
  currency?: string
  url?: string
  thumbnail?: string
  item_count?: number | string
  discounted_subtotal_amount?: number | string
  subtotal_amount?: number | string
  discount_amount?: number | string
  shipping_fee?: number | string
  tax_amount?: number | string
  items?: CustomerOrderLine[]
  shipments?: CustomerOrderShipment[]
}

export interface CustomerOrderLine {
  id: string | number
  product_id?: string | number
  variant_id?: string | number | null
  name?: string
  sku?: string
  thumbnail?: string
  quantity?: number | string
  currency?: string
  unit_price?: number | string
  subtotal_amount?: number | string
  discount_amount?: number | string
  total_amount?: number | string
  fulfillment_mode?: string
  inventory_snapshot?: string
  pricing_snapshot?: string
}

export interface CustomerOrderShipment {
  id: string | number
  carrier?: string
  carrier_service?: string
  tracking_number?: string
  registration_status?: string
  sync_status?: string
  last_event_at?: string | number | Date | null
}

export interface CustomerOrders {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  total?: number | string
  items?: CustomerOrderItem[]
  fulfillment_status?: CustomerContextFactStatus
  fulfillment_reason?: string
}

export type CustomerContextFactStatus = 'available' | 'unavailable' | 'permission_denied' | 'error' | string

export interface CustomerShippingAddress {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  source_order_id?: string | number
  source_order_number?: string
  recipient_name?: string
  address_line?: string
  city?: string
  state?: string
  postal_code?: string
  country?: string
  phone_present?: boolean
}

export interface CustomerAfterSalesItem {
  id: string | number
  order_id?: string | number
  order_number?: string
  type?: string
  status?: string
  reason?: string
  item_summary?: string
  created_at?: string | number | Date | null
  updated_at?: string | number | Date | null
  current_handler_id?: string | number
  current_handler_name?: string
  resolution?: string
}

export interface CustomerRefund {
  id: string | number
  order_id?: string | number
  order_number?: string
  status?: string
  reason?: string
  amount?: number | string
  currency?: string
  created_at?: string | number | Date | null
  completed_at?: string | number | Date | null
}

export interface CustomerWarrantyClaim {
  id: string | number
  order_item_id?: string | number | null
  order_number?: string
  issue_type?: string
  status?: string
  tire_pressure?: string
  is_tubeless?: boolean
  resolution?: string
  processed_by?: string | number
  created_at?: string | number | Date | null
  updated_at?: string | number | Date | null
}

export interface CustomerAfterSales {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  items?: CustomerAfterSalesItem[]
  refund_status?: CustomerContextFactStatus
  refund_reason?: string
  refunds?: CustomerRefund[]
  warranty_status?: CustomerContextFactStatus
  warranty_reason?: string
  warranty_claims?: CustomerWarrantyClaim[]
}

export interface CustomerPaymentDisputeItem {
  provider?: string
  order_id?: string | number
  order_number?: string
  status?: string
  reason?: string
  amount?: number | string
  currency?: string
  evidence_due_at?: string | number | Date | null
  evidence_submitted_at?: string | number | Date | null
}

export interface CustomerPaymentDisputes {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  items?: CustomerPaymentDisputeItem[]
}

export interface CustomerBrowsingItem {
  product_id: string | number
  name?: string
  sku?: string
  thumbnail?: string
  currency?: string
  price?: number | string
  view_count?: number | string
  last_viewed_at?: string | number | Date | null
}

export interface CustomerBrowsing {
  available: boolean
  status?: CustomerContextFactStatus
  reason?: string
  count?: number | string
  items?: CustomerBrowsingItem[]
}

export interface CustomerSignal {
  status?: string
  value?: string
  reason?: string
}

export interface CustomerContext {
  customer?: {
    account?: CustomerAccount | null
    anonymous?: CustomerAnonymous | null
  }
  contact?: CustomerContact
  cart?: CustomerCart
  wishlist?: CustomerWishlist
  orders?: CustomerOrders
  shipping_address?: CustomerShippingAddress
  after_sales?: CustomerAfterSales
  payment_disputes?: CustomerPaymentDisputes
  browsing?: CustomerBrowsing
  signals?: Record<string, CustomerSignal>
}

export interface CustomerServiceAnalyticsRegion {
  region_label: string
  member_count?: number | string
  visitor_count?: number | string
  count: number | string
  percent?: number | string
}

export interface CustomerServiceAnalytics {
  date?: string
  total_conversations?: number | string
  known_region_count?: number | string
  unknown_region_count?: number | string
  member_customer_count?: number | string
  converted_member_customer_count?: number | string
  member_conversion_rate?: number | string
  average_reply_interval_seconds?: number | string
  reply_interval_count?: number | string
  unanswered_customer_turns?: number | string
  regions?: CustomerServiceAnalyticsRegion[]
}

export interface FAQItem {
  id: string | number
  question: string
  answer?: string
  answer_image_url?: string | null
}

export interface FAQPage {
  page_id: string
  locale: string
  title?: string
  route_path?: string
  items?: FAQItem[]
}

export interface FAQSelection {
  page: FAQPage
  faq: FAQItem
}

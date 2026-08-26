export interface AssignableAgent {
  id?: string | number
  user_id?: string | number
  name?: string
  email?: string
}

export interface AssignableGroup {
  id: string | number
  name: string
}

export interface CustomerServiceFiltersState {
  search: string
  groupId: string
  status: string
  identity: string
  assignedTo: string
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
  conversation_id?: string
  ticket_number?: string | number
  assigned_to?: string | number | null
  unread_count?: number
  last_message?: string
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
  line_total?: number | string
}

export interface CustomerCart {
  available: boolean
  reason?: string
  item_count?: number | string
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
}

export interface CustomerOrders {
  available: boolean
  reason?: string
  total?: number | string
  items?: CustomerOrderItem[]
}

export interface CustomerBrowsingItem {
  product_id: string | number
  view_count?: number | string
  last_viewed_at?: string | number | Date | null
}

export interface CustomerBrowsing {
  available: boolean
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

export interface FAQCategory {
  category_key: string
  name?: string
  faqs?: FAQItem[]
}

export interface FAQPage {
  page_id: string
  locale: string
  title?: string
  route_path?: string
  categories?: FAQCategory[]
}

export interface FAQSelection {
  page: FAQPage
  category: FAQCategory
  faq: FAQItem
}

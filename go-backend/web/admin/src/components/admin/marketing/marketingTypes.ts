import type { AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'

export interface MarketingPagination {
  page: number
  pageSize: number
  total: number
}

export interface CouponRecord {
  id: string | number
  code: string
  type: string
  value?: number | string
  description?: string
  min_amount?: number | string
  used_count?: number | string
  usage_limit?: number | string
  start_date?: string
  end_date?: string
  enabled?: boolean
}

export interface CouponFilters {
  status: string
}

export interface MarketingStatusDisplay {
  label: string
  tone: AdminStatusTone
}

export interface GiftCardRecord {
  id: string | number
  code: string
  initial_value?: number | string
  balance?: number | string
  currency?: string
  recipient_name?: string
  recipient_email?: string
  status?: string
  expires_at?: string
  created_at?: string
}

export interface GiftCardFilters {
  status: string
}

export interface LoyaltyTransaction {
  id: string | number
  type?: string | null
  points?: number | string | null
  balance?: number | string | null
  source?: string | null
  source_id?: string | number | null
  description?: string | null
  created_at?: string | number | Date | null
}

export interface LoyaltyFilters {
  user_id: string | number
}

export interface LoyaltyAdjustmentForm {
  user_id: string | number
  points: number | string
  description: string
}

export type LoyaltyErrors = Partial<Record<keyof LoyaltyAdjustmentForm, string>>

export interface LoyaltySettings {
  tz_loyalty_purchase_earn_points_per_currency_unit: number | string
  tz_loyalty_referral_referrer_points: number | string
  tz_loyalty_referral_referee_points: number | string
  tz_loyalty_checkin_base_points: number | string
  tz_loyalty_checkin_streak_interval_days: number | string
  tz_loyalty_checkin_streak_bonus_points: number | string
  tz_loyalty_checkin_max_points: number | string
}

export interface GiftCardRedeemOption {
  key?: string
  value: number | string
  currency: string
  stock_quantity: number | string
  redeemed_quantity?: number | string
  remaining_quantity?: number | string
}

export interface GiftCardRedeemSettings {
  tz_redeem_enabled: boolean
  tz_redeem_currency: string
  tz_redeem_exchange_rate: number | string
  tz_redeem_min_points: number | string
  tz_redeem_max_value_per_day: number | string
  tz_redeem_card_expiry_days: number | string
  options?: GiftCardRedeemOption[]
}

export interface MemberLevel {
  id?: string | number | null
  name?: string
  color?: string
  min_points?: number | string
  max_points?: number | string
  discount_rate?: number | string
  benefits?: string
  sort_order?: number | string
}

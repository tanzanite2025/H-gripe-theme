export interface Coupon {
  id?: number
  code: string
  type: 'percentage' | 'fixed'
  value_minor?: number
  value_rate_decimal?: string
  currency?: string
  min_amount_minor?: number
  max_discount_minor?: number
  discount_minor?: number
}

export interface CouponValidationResponse {
  valid: boolean
  coupon: Coupon
  discount_minor: number
}

export interface CartItem {
  product_id?: number
  variant_id?: number | null
  product_specification_template_id?: number | null
  price_minor: number
  currency?: string
  quantity: number
  weight?: number
  weight_grams?: number
}

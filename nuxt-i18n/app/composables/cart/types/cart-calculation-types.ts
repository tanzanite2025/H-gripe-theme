/**
 * 购物车计算系统 - 类型定义
 */

// 会员等级配置
export interface MemberTier {
  name: string
  min: number
  max: number | null
  discount: number // 折扣百分比
}

// 运费模板（扩展版）
export interface CartShippingTemplate {
  id: number
  name: string
  template_name?: string
  type: 'weight' | 'quantity' | 'volume' | 'amount' | 'items' | 'price'
  base_fee?: number
  default_fee?: number
  free_threshold?: number
  free_shipping?: boolean
  enabled?: boolean
  is_active?: boolean
  rules: Array<{
    type?: string
    min: number | null
    max: number | null
    min_value?: number | null
    max_value?: number | null
    fee: number
    region?: string
    free_over?: number | null
    regions?: string[]
    zip_ranges?: string[]
    eta_min_days?: number | null
    eta_max_days?: number | null
    service?: string
    service_label?: string
  }>
}

// 配送地址
export interface ShippingAddressInfo {
  country: string
  zip?: string
  city?: string
}

// 税率配置
export interface TaxRate {
  id: number
  name: string
  rate: number // 百分比
  region?: string
  is_active: boolean
}

// 用户积分信息
export interface UserPoints {
  total: number
  available: number
  tier: string
}

// 优惠券
export interface Coupon {
  id?: number
  code: string
  type: 'percentage' | 'fixed' | 'points'
  value_minor?: number
  value_rate_decimal?: string
  currency?: string
  min_amount_minor?: number
  max_discount_minor?: number
}

export interface CouponValidationResponse {
  valid: boolean
  coupon: Coupon
  discount_minor: number
}

// 购物车商品
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

// 运费计算结果
export interface ShippingCalculationResult {
  fee_minor: number
  rule: CartShippingTemplate['rules'][0] | null
  template: CartShippingTemplate | null
}

// 总价计算结果
export interface TotalCalculationResult {
  subtotal_minor: number
  member_discount_minor: number
  memberTier: MemberTier
  coupon_discount_minor: number
  points_discount_minor: number
  discounted_subtotal_minor: number
  shipping_minor: number
  tax_minor: number
  total_minor: number
  breakdown: {
    original_subtotal_minor: number
    total_discount_minor: number
    shipping_fee_minor: number
    tax_fee_minor: number
    final_total_minor: number
  }
}

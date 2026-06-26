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
  type: 'weight' | 'quantity' | 'volume' | 'amount' | 'items'
  base_fee: number
  free_threshold?: number
  is_active?: boolean
  rules: Array<{
    type?: string
    min: number | null
    max: number | null
    fee: number
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

// 优惠券/礼品卡
export interface Coupon {
  code: string
  type: 'percentage' | 'fixed' | 'points'
  value: number
  min_amount?: number
}

// 购物车商品
export interface CartItem {
  price: number
  quantity: number
  weight?: number
}

// 运费计算结果
export interface ShippingCalculationResult {
  fee: number
  rule: CartShippingTemplate['rules'][0] | null
  template: CartShippingTemplate | null
}

// 总价计算结果
export interface TotalCalculationResult {
  subtotal: number
  memberDiscount: number
  memberTier: MemberTier
  couponDiscount: number
  pointsDiscount: number
  discountedSubtotal: number
  shipping: number
  tax: number
  total: number
  breakdown: {
    originalSubtotal: number
    totalDiscount: number
    shippingFee: number
    taxFee: number
    finalTotal: number
  }
}

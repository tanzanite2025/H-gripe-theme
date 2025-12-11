export interface ShippingRule {
  type?: 'weight' | 'quantity' | 'amount' | 'items' | string
  min?: number | null
  max?: number | null
  fee: number
  priority?: number
  free_over?: number | null
  regions?: string[]
  zip_ranges?: string[]
  eta_min_days?: number | null
  eta_max_days?: number | null
  service?: string
  service_label?: string
}

export interface ShippingTemplate {
  id: number
  name?: string
  template_name?: string
  description?: string
  type?: 'weight' | 'quantity' | 'volume' | 'amount' | 'items' | string
  base_fee?: number
  free_threshold?: number
  is_active?: boolean
  rules: ShippingRule[]
  meta?: {
    carrier?: string
    currency?: string
  }
}

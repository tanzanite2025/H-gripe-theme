export interface CartItem {
  id: number | string
  product_id?: number
  variant_id?: number | null
  title: string
  price: number
  currency?: string
  quantity: number
  slug?: string
  name?: string
  sku?: string
  sale_price?: number | null
  thumbnail?: string
  image?: string
  weight?: number
  weight_grams?: number
  product_specification_template_id?: number | null
  category?: string
  categories?: unknown[]
  tags?: string[]
  fulfillment_mode?: 'stock' | 'made_to_order'
  selected_options?: CartSelectedOption[]
  configuration_hash?: string
}

export interface CartSelectedOption {
  group_slug: string
  value_keys: string[]
}

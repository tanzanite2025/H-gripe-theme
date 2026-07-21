export interface CartItem {
  id: number
  product_id?: number
  variant_id?: number | null
  title: string
  price: number
  quantity: number
  slug?: string
  name?: string
  sku?: string
  sale_price?: number | null
  thumbnail?: string
  image?: string
  maxStock?: number
  stock?: number
  weight?: number
  weight_grams?: number
  product_type_id?: number | null
  category?: string
  categories?: unknown[]
  tags?: string[]
}

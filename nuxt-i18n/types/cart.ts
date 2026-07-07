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
  category?: string
  categories?: unknown[]
  tags?: string[]
}

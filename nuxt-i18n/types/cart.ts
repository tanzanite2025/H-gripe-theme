export interface CartItem {
  id: number
  title: string
  price: number
  quantity: number
  slug?: string
  name?: string
  sku?: string
  thumbnail?: string
  maxStock?: number
  weight?: number
  category?: string
  tags?: string[]
}

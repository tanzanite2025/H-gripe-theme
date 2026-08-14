export interface QuickBuySelectedProduct {
  productId: number
  stepKey: string
  variantId: number | null
  title: string
  slug: string
  sku?: string
  thumbnail: string
  quantity: number
  weightGrams: number
  unitPrice: number
  currency: string
}

export interface QuickBuySelectedProductStepSlot {
  slotKey: string
  index: number
  stepKey: string
  stepLabel: string
  item: QuickBuySelectedProduct | null
  additionalItemCount: number
}

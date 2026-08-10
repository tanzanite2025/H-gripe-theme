export interface QuickBuyStep {
  id: number
  slug: string
  name: string
}

export interface QuickBuyConfig {
  steps?: unknown[]
  storeApiBase?: string
  cartUrl?: string
  checkoutUrl?: string
  taxonomy?: string
  buttonText?: string
  enabled?: boolean
  successMessage?: string
  requireLogin?: boolean
}

export type QuickBuyResolvedConfig = Omit<QuickBuyConfig, 'steps'> & {
  steps: QuickBuyStep[]
}

export interface QuickBuyStep {
  id: number
  slug: string
  name: string
  stepKey?: string
  sortOrder?: number
  productTypes?: QuickBuyProductType[]
}

export interface QuickBuyProductType {
  id: number
  slug: string
  name: string
  imageUrl?: string
  primary?: boolean
}

export interface QuickBuyConfig {
  steps?: QuickBuyStep[]
  flowHelpText?: string
}

export interface QuickBuyFlowVersion {
  id: number
  versionNumber: number
  status: string
  publishedAt?: string
  startsAt?: string
  endsAt?: string
}

export interface QuickBuyFlow {
  id: number
  slug: string
  name: string
  description?: string
  helpText?: string
  translations?: QuickBuyFlowTranslation[]
  entrySurface?: string
  isEnabled?: boolean
  version?: QuickBuyFlowVersion
  steps: QuickBuyStep[]
}

export interface QuickBuyFlowTranslation {
  id?: number
  locale: string
  helpText?: string
}

export interface QuickBuySessionSelectionInput {
  stepKey: string
  productId: number
  variantId?: number | null
  quantity: number
}

export interface QuickBuySessionValidationIssue {
  severity: 'error' | 'warning' | 'info' | string
  code: string
  message: string
  stepKey?: string
  productId?: number
  variantId?: number | null
}

export interface QuickBuySessionValidationResult {
  valid: boolean
  issues: QuickBuySessionValidationIssue[]
}

export interface QuickBuySessionItem {
  id: number
  stepKey: string
  productId: number
  variantId?: number | null
  quantity: number
  unitPriceSnapshot: number
  currencySnapshot: string
  weightSnapshotG: number
  productSnapshot?: Record<string, unknown>
  variantSnapshot?: Record<string, unknown>
  sortOrder: number
}

export interface QuickBuySession {
  sessionToken: string
  flowId: number
  flowVersionId: number
  locale: string
  marketCountry: string
  currency: string
  status: string
  validationStatus: string
  subtotalSnapshot: number
  weightSnapshotG: number
  expiresAt?: string
  flow?: QuickBuyFlow
  items: QuickBuySessionItem[]
  validation?: QuickBuySessionValidationResult
}

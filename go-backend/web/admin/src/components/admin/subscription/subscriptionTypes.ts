export type SubscriptionStatusTone = 'green' | 'gray'
export type SubscriptionSelectionState = boolean | 'indeterminate'
export type SubscriptionConfirmationType = '' | 'status' | 'delete' | 'batch-delete'

export type SubscriptionBooleanResolver = (email: string) => boolean
export type SubscriptionLabelResolver = (value?: string | null) => string
export type SubscriptionToneResolver = (status?: string | null) => SubscriptionStatusTone
export type SubscriptionDateFormatter = (value?: string | null) => string

export interface SubscriptionRecord {
  id?: number | string | null
  email: string
  status?: string | null
  locale?: string | null
  source?: string | null
  tags?: string | null
  subscribed_at?: string | null
  unsubscribed_at?: string | null
}

export interface SubscriptionFilters {
  search: string
  status: string
}

export interface SubscriptionPagination {
  page: number
  pageSize: number
  total: number
}

export interface SubscriptionStats {
  total_count?: number
  total?: number
  active_count?: number
  active?: number
  unsubscribed_count?: number
  cancelled?: number
  monthly_count?: number
  today?: number
}

export interface SubscriptionListResponse {
  subscriptions?: SubscriptionRecord[]
  pagination?: {
    total?: number
  }
  total?: number
}

export interface SubscriptionEmailsResponse {
  emails?: string[]
}

export interface SubscriptionBatchDeleteResponse {
  deleted?: number
}

export interface SubscriptionConfirmation {
  open: boolean
  type: SubscriptionConfirmationType
  target: SubscriptionRecord | SubscriptionRecord[] | null
  status: string
  title: string
  description: string
  confirmLabel: string
  destructive: boolean
}

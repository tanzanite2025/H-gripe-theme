export interface PaymentRiskSnapshot {
  level?: string
  window_start?: string | number | Date | null
  window_end?: string | number | Date | null
  window_days?: number | string
  successful_payment_count?: number | string
  dispute_activity_rate?: number | string
  early_fraud_warning_rate?: number | string
  refund_rate?: number | string
  dispute_count?: number | string
  refund_count?: number | string
}

export interface PaymentRiskProviderReport {
  snapshot?: PaymentRiskSnapshot | null
  reasons?: string[]
}

export type PaymentRiskReports = Record<string, PaymentRiskProviderReport | undefined>

export type PaymentProtectionAction = 'force_3ds' | 'pause_payment' | string
export type PaymentProtectionScopeType = 'global' | 'provider' | 'country' | 'payment_method' | string

export interface PaymentProtectionPolicy {
  max_control_duration_hours: number | string
  max_pause_payment_duration_hours: number | string
  max_global_pause_payment_duration_hours: number | string
  [key: string]: number | string
}

export interface PaymentProtectionControl {
  id: string | number
  action: PaymentProtectionAction
  scope_type: PaymentProtectionScopeType
  scope_value?: string
  status?: string
  expires_at?: string | number | Date | null
  reason?: string
  active?: boolean
}

export interface PaymentProtectionControlPayload {
  action: PaymentProtectionAction
  scope_type: PaymentProtectionScopeType
  scope_value: string
  reason: string
  expires_at: string
  confirm: boolean
}

export interface PaymentProtectionAuditLog {
  id: string | number
  action?: string
  username?: string
  user_id?: string | number
  ip_address?: string
  created_at?: string | number | Date | null
}

export interface PaymentRefundRecommendation {
  id: string | number
  status?: string
  provider?: string
  recommended_amount?: number | string
  currency?: string
  order_id?: string | number | null
  linked_refund_id?: string | number | null
  reason?: string
  provider_payment_id?: string
  payment_intent_id?: string
  charge_id?: string
  refund_created_at?: string | number | Date | null
  transaction_id?: string | number | null
}

export interface PaymentRefundDraftPayload {
  amount: number
  reason: string
  decision_notes: string
  confirm: boolean
}

export interface PaymentRefundExecutionPayload {
  refund_id: string | number
  confirm: boolean
}

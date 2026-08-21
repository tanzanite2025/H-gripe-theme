export interface RiskStrategySnapshot {
  level?: string
  window_start?: string | number | Date | null
  window_end?: string | number | Date | null
  window_days?: number | string
  successful_payment_count?: number | string
  successful_payment_amount?: number | string
  dispute_activity_rate?: number | string
  early_fraud_warning_rate?: number | string
  refund_rate?: number | string
  three_ds_upgrade_rate?: number | string
  dispute_count?: number | string
  dispute_amount?: number | string
  early_fraud_warning_count?: number | string
  refund_count?: number | string
  refund_amount?: number | string
  checkout_attempt_count?: number | string
  three_ds_upgrade_count?: number | string
  three_ds_challenge_count?: number | string
  three_ds_exemption_count?: number | string
  recommended_action?: string
  computed_at?: string | number | Date | null
}

export interface RiskStrategyProviderReport {
  snapshot?: RiskStrategySnapshot | null
  reasons?: string[]
}

export type RiskStrategyReports = Record<string, RiskStrategyProviderReport | undefined>

export interface RiskStrategyMonitoringPolicy {
  window_days?: number | string
  minimum_successful_payments?: number | string
  warning_dispute_activity_rate?: number | string
  critical_dispute_activity_rate?: number | string
  warning_early_fraud_rate?: number | string
  critical_early_fraud_rate?: number | string
  warning_refund_rate?: number | string
  critical_refund_rate?: number | string
  auto_step_up_enabled?: boolean
  alerting_enabled?: boolean
}

export interface RiskStrategyMonitoringConfiguration extends RiskStrategyMonitoringPolicy {
  worker_enabled?: boolean
  worker_interval_seconds?: number | string
}

export interface RiskStrategyConfiguration {
  monitoring?: RiskStrategyMonitoringConfiguration
  three_ds?: {
    enabled?: boolean
    runtime_available?: boolean
    adaptive_enabled?: boolean
    low_risk_max_amount?: number | string
    avs_billing_shipping_mismatch_high_value_threshold_usd?: number | string
    trusted_paid_orders?: number | string
    visitor_risk_lookback_days?: number | string
    step_up_risk_score?: number | string
    challenge_risk_score?: number | string
  }
  payment_risk?: {
    enabled?: boolean
    failure_window_seconds?: number | string
    failure_threshold?: number | string
    delay_seconds?: number | string
    high_risk_score?: number | string
  }
  bin_rate_limit?: {
    enabled?: boolean
    window_seconds?: number | string
    failure_threshold?: number | string
    block_duration_seconds?: number | string
  }
  gateway_circuit_breaker?: {
    enabled?: boolean
    window_seconds?: number | string
    failure_rate_threshold?: number | string
    minimum_sample_count?: number | string
    open_duration_seconds?: number | string
  }
  protection?: {
    enabled?: boolean
    max_control_duration_hours?: number | string
    max_pause_payment_duration_hours?: number | string
    max_global_pause_payment_duration_hours?: number | string
  }
  anti_abuse?: {
    turnstile_required?: boolean
    turnstile_configured?: boolean
    verification_ip_window_seconds?: number | string
    verification_destination_window_seconds?: number | string
    verification_daily_limit?: number | string
    verification_global_window_seconds?: number | string
    verification_global_limit?: number | string
    verification_circuit_seconds?: number | string
  }
  order_abuse?: {
    enabled?: boolean
    order_create_window_seconds?: number | string
    max_order_creations_per_user?: number | string
    max_order_creations_per_session?: number | string
    max_order_creations_per_ip?: number | string
  }
  visitor_risk?: {
    enabled?: boolean
    flush_interval_seconds?: number | string
    max_pending_facts?: number | string
    sample_path_limit?: number | string
    retention_days?: number | string
  }
}

export interface PaymentGatewayHealth {
  provider?: string
  enabled?: boolean
  allowed?: boolean
  circuit_open?: boolean
  failure_rate?: number | string
  sample_count?: number | string
  failure_count?: number | string
  retry_after_seconds?: number | string
  error?: string
}

export type PaymentGatewayHealthMap = Record<string, PaymentGatewayHealth | undefined>

export type RiskStrategyGatewayHealth = PaymentGatewayHealth
export type RiskStrategyGatewayHealthMap = PaymentGatewayHealthMap

export interface PaymentGatewayRuntimeStatus {
  provider: string
  label?: string
  environment?: string
  three_ds_mode?: string
  runtime_source?: string
  configured?: boolean
  webhook_configured?: boolean
  webhook_supported?: boolean
  production_ready?: boolean
  callback_url?: string
  required_fields?: string[]
  configured_fields?: string[]
  missing?: string[]
  blockers?: string[]
  warnings?: string[]
  documentation_url?: string
  documentation_label?: string
  admin_config_configured?: boolean
  admin_config_readable?: boolean
}

export interface PaymentGatewayRuntime {
  runtime_source?: string
  secret_store_configured?: boolean
  gateways?: PaymentGatewayRuntimeStatus[]
}

export type RiskStrategyGatewayRuntimeStatus = PaymentGatewayRuntimeStatus
export type RiskStrategyGatewayRuntime = PaymentGatewayRuntime

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

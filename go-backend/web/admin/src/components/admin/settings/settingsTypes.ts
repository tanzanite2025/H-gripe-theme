export interface MarketLanguageOption {
  code: string
  name: string
}

export interface MarketCurrencyOption {
  code: string
  name: string
}

export interface StorefrontMarket {
  id?: number | string | null
  code: string
  name?: string
  countries?: string[]
  supported_locales: string[]
  default_locale: string
  display_currencies: string[]
  default_currency: string
  payment_method_policy?: string
  logistics_policy?: string
  tax_policy?: string
  enabled: boolean
  priority?: number | string
}

export interface StorefrontMarketForm {
  code: string
  name: string
  supported_locales: string[]
  default_locale: string
  display_currencies: string[]
  default_currency: string
  payment_method_policy: string
  logistics_policy: string
  tax_policy: string
  enabled: boolean
  priority: number | string
}

export interface PaymentMethodRecord {
  id: number | string
  name: string
  code: string
  icon?: string
  description?: string
  fee_type?: string
  fee_value?: number | string
  min_amount?: number | string
  max_amount?: number | string
  enabled?: boolean
  sort_order?: number | string
  settings?: string
}

export interface PaymentMethodForm {
  id: number | string | null
  name: string
  code: string
  icon: string
  description: string
  fee_type: string
  fee_value: number | string
  min_amount: number | string
  max_amount: number | string
  enabled: boolean
  sort_order: number | string
  settings: string
}

export interface PublicChatGroup {
  id: number | string
  name: string
  code?: string
  description?: string
  sort_order?: number | string
  status?: string
}

export interface PublicChatAgentGroup {
  id: number | string
  name: string
}

export interface PublicChatAgent {
  id: number | string
  avatar?: string
  display_name?: string
  username?: string
  email?: string
  whatsapp?: string
  agent_id?: string
  user_id?: number | string
  raw_role?: string
  normalized_role?: string
  user_status?: string
  profile_status?: string
  online_status?: string
  groups?: PublicChatAgentGroup[]
  exposed?: boolean
}

export interface PublicChatSummary {
  profile_count?: number | string
  exposed_agents?: number | string
}

export interface PaymentGatewayRuntimeStatus {
  provider: string
  environment?: string
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

export interface PaymentGatewayOption {
  value: string
  label: string
  description: string
  officialDashboardURL?: string
}

export interface PaymentCallbackCheckResult {
  transport_reachable?: boolean
  route_reachable?: boolean
  expected_signature_failure?: boolean
  status_code?: number | string
  method?: string
  duration_ms?: number | string
  error?: string
}

export interface PaymentGatewayCredentialField {
  key: string
  label: string
  placeholder: string
  description?: string
  multiline?: boolean
}

export interface CommercialCrawlerRule {
  provider?: string
  user_agent: string
}

export interface CommercialCrawlerEnforcementLayer {
  layer: string
  status?: string
}

export interface CommercialCrawlerIntelligenceSeed {
  id: string | number
  category?: string
  name: string
  action?: string
  enforcement?: string
  identification?: string
  aliases?: string[]
  detection_signals?: string[]
  threshold?: string
}

export interface OrderNumberProtection {
  configured?: boolean
  format?: string
  verification?: string
  key_rotation?: string
  public_id_policy?: string[]
  internal_id_policy?: string[]
}

export interface CommercialCrawlerRobotsTxtPolicy {
  path?: string
  source?: string
  wildcard_user_agent?: string
  disallow?: string
  blocked_user_agents?: string[]
  block_status?: number | string
}

export interface CommercialCrawlerProtection {
  enabled?: boolean
  response_status?: number | string
  rules?: CommercialCrawlerRule[]
  enforcement?: CommercialCrawlerEnforcementLayer[]
  intelligence_seeds?: CommercialCrawlerIntelligenceSeed[]
  order_number_protection?: OrderNumberProtection | null
  robots_txt?: CommercialCrawlerRobotsTxtPolicy | null
}

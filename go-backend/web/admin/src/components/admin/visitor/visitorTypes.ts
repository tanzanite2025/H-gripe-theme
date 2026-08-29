export interface VisitorPagination {
  page: number
  pageSize: number
  total: number
}

export interface VisitorProfile {
  id: number | string
  identity?: string
  user_id?: number | string | null
  profile_status?: string
  profile_quality_score?: number | string
  last_meaningful_action?: string
  email?: string
  email_source?: string
  region_label?: string
  ip_address?: string
  ip_block?: VisitorIPBlockRule | null
  ip_block_rules?: VisitorIPBlockRule[]
  ip_block_match?: VisitorIPBlockRule | null
  locale?: string
  has_customer_service_visitor?: boolean
  has_cart_session?: boolean
  has_email?: boolean
  customer_service_visitor_hash_preview?: string
  cart_session_id?: string
  locale_source?: string
  country_code?: string
  timezone?: string
  region?: string
  city?: string
  has_ip_fingerprint?: boolean
  has_user_agent_fingerprint?: boolean
  last_seen_at?: string | number | Date | null
  last_meaningful_seen_at?: string | number | Date | null
  first_meaningful_seen_at?: string | number | Date | null
  retention_until?: string | number | Date | null
  created_at?: string | number | Date | null
  updated_at?: string | number | Date | null
}

export interface VisitorIPBlockRule {
  id: number | string
  cidr?: string
  source?: string
  source_reference?: string
  reason?: string
  expires_at?: string | number | Date | null
  enabled?: boolean
  status?: string
  created_by?: number | string | null
  disabled_by?: number | string | null
  disabled_at?: string | number | Date | null
  created_at?: string | number | Date | null
  updated_at?: string | number | Date | null
}

export interface VisitorIPBlockPayload {
  reason: string
  expires_at: string | null
}

export interface VisitorGlobalIPBlockPayload extends VisitorIPBlockPayload {
  cidr: string
}

export interface VisitorRiskDecision {
  action?: string
  reason?: string
  expires_at?: string | number | Date | null
}

export interface VisitorRiskFact {
  id: number | string
  day?: string | number | Date | null
  risk_level?: string
  risk_score?: number | string
  decision?: VisitorRiskDecision | null
  ip_hash_preview?: string
  device_fingerprint_hash_preview?: string
  user_agent_hash_preview?: string
  country_code?: string
  request_count?: number | string
  no_cookie_request_count?: number | string
  invalid_request_count?: number | string
  auth_failure_count?: number | string
  checkout_failure_count?: number | string
  bot_like_user_agent_count?: number | string
  unique_path_count?: number | string
  unique_anonymous_count?: number | string
  unique_session_count?: number | string
  meaningful_action_count?: number | string
  sample_paths?: string[]
  last_seen_at?: string | number | Date | null
}

export interface VisitorRiskDecisionPayload {
  action: string
  reason: string
  expires_at: string | null
}

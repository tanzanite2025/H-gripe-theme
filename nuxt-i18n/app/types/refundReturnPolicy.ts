export interface RefundReturnPolicyImage {
  url: string
  alt: string
  caption?: string
}

export interface RefundReturnPolicySection {
  id: string
  title: string
  body?: string
  bullets?: string[]
  image?: RefundReturnPolicyImage | null
}

export interface RefundReturnPolicy {
  title: string
  intro?: string
  sections: RefundReturnPolicySection[]
  contact_label?: string
  contact_url?: string
  updated_at?: string
}

export interface RefundReturnPolicyResponse {
  policy: RefundReturnPolicy
  locale?: string
  requested_locale?: string
  fallback?: boolean
}

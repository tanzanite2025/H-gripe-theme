export interface RefundCancellationPolicyImage {
  url: string
  alt: string
  caption?: string
}

export interface RefundCancellationPolicySection {
  id: string
  title: string
  body?: string
  bullets?: string[]
  image?: RefundCancellationPolicyImage | null
}

export interface RefundCancellationPolicy {
  title: string
  intro?: string
  sections: RefundCancellationPolicySection[]
  contact_label?: string
  contact_url?: string
  updated_at?: string
}

export interface RefundCancellationPolicyResponse {
  policy: RefundCancellationPolicy
  locale?: string
  requested_locale?: string
  fallback?: boolean
}

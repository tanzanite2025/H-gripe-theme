import axios from '@/utils/axios'

export interface RefundReturnPolicyImage {
  url: string
  alt: string
  caption: string
}

export interface RefundReturnPolicySection {
  id: string
  title: string
  body: string
  bullets: string[]
  image?: RefundReturnPolicyImage | null
}

export interface RefundReturnPolicy {
  title: string
  intro: string
  sections: RefundReturnPolicySection[]
  contact_label: string
  contact_url: string
  updated_at: string
}

export interface RefundReturnPolicyEditorSection extends Omit<RefundReturnPolicySection, 'image'> {
  bulletsText: string
  image: RefundReturnPolicyImage
}

export interface RefundReturnPolicyEditor extends Omit<RefundReturnPolicy, 'sections'> {
  sections: RefundReturnPolicyEditorSection[]
}

export interface RefundReturnPolicyResponse {
  policy: RefundReturnPolicy
  locale: string
  requested_locale: string
  fallback: boolean
}

const endpoint = '/api/admin/settings/refund-return-policy'

export const refundReturnPolicyApi = {
  async get(locale: string): Promise<RefundReturnPolicyResponse> {
    const response = await axios.get<RefundReturnPolicyResponse>(endpoint, { params: { locale } })
    return response.data
  },

  async update(locale: string, policy: RefundReturnPolicy): Promise<RefundReturnPolicyResponse> {
    const response = await axios.put<RefundReturnPolicyResponse>(endpoint, { locale, policy })
    return response.data
  },
}

export default refundReturnPolicyApi

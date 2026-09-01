import axios from '@/utils/axios'
import type { MediaAsset } from '@/api/media'
import {
  requireApiObject,
  requireApiObjectField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

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

const endpoint = '/api/admin/content/refund-return-policy'

export const refundReturnPolicyApi = {
  async get(locale: string): Promise<RefundReturnPolicyResponse> {
    const response = await axios.get<RefundReturnPolicyResponse>(endpoint, { params: { locale } })
    return response.data
  },

  async update(locale: string, policy: RefundReturnPolicy): Promise<RefundReturnPolicyResponse> {
    const response = await axios.put<RefundReturnPolicyResponse>(endpoint, { locale, policy })
    return response.data
  },

  async uploadImage(formData: FormData): Promise<MediaAsset> {
    const path = `${endpoint}/assets`
    const payload = requireApiObject(unwrapApiPayload(await axios.post(path, formData), path), path)
    return requireApiObjectField<MediaAsset>(payload, 'asset', path)
  },
}

export default refundReturnPolicyApi

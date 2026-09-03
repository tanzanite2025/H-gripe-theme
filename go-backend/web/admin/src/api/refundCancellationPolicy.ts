import axios from '@/utils/axios'
import type { MediaAsset } from '@/api/media'
import {
  requireApiObject,
  requireApiObjectField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export interface RefundCancellationPolicyImage {
  url: string
  alt: string
  caption: string
}

export interface RefundCancellationPolicySection {
  id: string
  title: string
  body: string
  bullets: string[]
  image?: RefundCancellationPolicyImage | null
}

export interface RefundCancellationPolicy {
  title: string
  intro: string
  sections: RefundCancellationPolicySection[]
  contact_label: string
  contact_url: string
  updated_at: string
}

export interface RefundCancellationPolicyEditorSection extends Omit<RefundCancellationPolicySection, 'image'> {
  bulletsText: string
  image: RefundCancellationPolicyImage
}

export interface RefundCancellationPolicyEditor extends Omit<RefundCancellationPolicy, 'sections'> {
  sections: RefundCancellationPolicyEditorSection[]
}

export interface RefundCancellationPolicyResponse {
  policy: RefundCancellationPolicy
  locale: string
  requested_locale: string
  fallback: boolean
}

const endpoint = '/api/admin/content/refund-cancellation-policy'

export const refundCancellationPolicyApi = {
  async get(locale: string): Promise<RefundCancellationPolicyResponse> {
    const response = await axios.get<RefundCancellationPolicyResponse>(endpoint, { params: { locale } })
    return response.data
  },

  async update(locale: string, policy: RefundCancellationPolicy): Promise<RefundCancellationPolicyResponse> {
    const response = await axios.put<RefundCancellationPolicyResponse>(endpoint, { locale, policy })
    return response.data
  },

  async uploadImage(formData: FormData): Promise<MediaAsset> {
    const path = `${endpoint}/assets`
    const payload = requireApiObject(unwrapApiPayload(await axios.post(path, formData), path), path)
    return requireApiObjectField<MediaAsset>(payload, 'asset', path)
  },
}

export default refundCancellationPolicyApi

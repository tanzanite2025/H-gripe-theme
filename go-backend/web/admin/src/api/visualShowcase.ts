import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type {
  VisualShowcaseAdministrationImageUploadResponse,
  VisualShowcaseAdministrationItemApiRecord,
  VisualShowcaseAdministrationItemSavePayload,
  VisualShowcaseAdministrationResponse,
} from '@/components/admin/visual-showcase/visualShowcaseTypes'

const visualShowcaseAdministrationEndpoint = (showcaseKey: string): string => (
  `/api/admin/content/visual-showcases/${encodeURIComponent(showcaseKey)}`
)

const visualShowcaseAdministrationUploadEndpoint = (showcaseKey: string): string => (
  `/api/admin/content/visual-showcases/${encodeURIComponent(showcaseKey)}/assets`
)

const readVisualShowcaseAdministrationResponse = (
  response: unknown,
  endpoint: string,
): VisualShowcaseAdministrationResponse => {
  const payload = requireApiObject(unwrapApiPayload(response, endpoint), endpoint)
  return {
    showcase_key: requireApiStringField(payload, 'showcase_key', endpoint),
    locale: requireApiStringField(payload, 'locale', endpoint),
    items: requireApiArrayField<VisualShowcaseAdministrationItemApiRecord>(payload, 'items', endpoint),
  }
}

const readVisualShowcaseUploadResponse = (
  response: unknown,
  endpoint: string,
): VisualShowcaseAdministrationImageUploadResponse => {
  const payload = requireApiObject(unwrapApiPayload(response, endpoint), endpoint)
  return requireApiObjectField<VisualShowcaseAdministrationImageUploadResponse>(payload, 'asset', endpoint)
}

export const visualShowcaseApi = {
  async getItems(showcaseKey: string, locale: string): Promise<VisualShowcaseAdministrationResponse> {
    const endpoint = visualShowcaseAdministrationEndpoint(showcaseKey)
    return readVisualShowcaseAdministrationResponse(
      await axios.get(endpoint, { params: { locale } }),
      endpoint,
    )
  },

  async uploadImage(
    showcaseKey: string,
    locale: string,
    file: File,
  ): Promise<VisualShowcaseAdministrationImageUploadResponse> {
    const endpoint = visualShowcaseAdministrationUploadEndpoint(showcaseKey)
    const formData = new FormData()
    formData.append('file', file)
    formData.append('locale', locale)
    return readVisualShowcaseUploadResponse(
      await axios.post(endpoint, formData),
      endpoint,
    )
  },

  async replaceItems(
    showcaseKey: string,
    locale: string,
    items: VisualShowcaseAdministrationItemSavePayload[],
  ): Promise<VisualShowcaseAdministrationResponse> {
    const endpoint = visualShowcaseAdministrationEndpoint(showcaseKey)
    return readVisualShowcaseAdministrationResponse(
      await axios.put(endpoint, { locale, items }),
      endpoint,
    )
  },
}

export default visualShowcaseApi

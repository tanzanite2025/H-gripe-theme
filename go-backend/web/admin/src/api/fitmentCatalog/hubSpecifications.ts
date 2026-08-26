import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  readApiBody,
} from '@/utils/apiResponse'
import { readHubSpecification } from './readers'
import type {
  HubSpecification,
  HubSpecificationListPayload,
  HubSpecificationPayload,
} from './types'

export const fitmentHubSpecificationsApi = {
  async listHubSpecifications(params: Record<string, unknown> = {}): Promise<HubSpecificationListPayload> {
    const endpoint = '/api/admin/fitment-catalog/hub-specifications'
    const body = requireApiObject(readApiBody(await axios.get(endpoint, { params }), endpoint), endpoint, 'response body')
    const specifications = requireApiArrayField<HubSpecification>(body, 'specifications', endpoint)
      .map((specification) => readHubSpecification(specification, endpoint))
    return {
      specifications,
      pagination: requireApiPagination(body, body, endpoint),
    }
  },

  async getHubSpecification(id: number | string): Promise<HubSpecification> {
    const endpoint = `/api/admin/fitment-catalog/hub-specifications/${id}`
    const body = requireApiObject(readApiBody(await axios.get(endpoint), endpoint), endpoint, 'response body')
    return readHubSpecification(requireApiObjectField(body, 'specification', endpoint), endpoint)
  },

  async createHubSpecification(payload: HubSpecificationPayload): Promise<HubSpecification> {
    const endpoint = '/api/admin/fitment-catalog/hub-specifications'
    const body = requireApiObject(readApiBody(await axios.post(endpoint, payload), endpoint), endpoint, 'response body')
    return readHubSpecification(requireApiObjectField(body, 'specification', endpoint), endpoint)
  },

  async updateHubSpecification(id: number | string, payload: HubSpecificationPayload): Promise<HubSpecification> {
    const endpoint = `/api/admin/fitment-catalog/hub-specifications/${id}`
    const body = requireApiObject(readApiBody(await axios.put(endpoint, payload), endpoint), endpoint, 'response body')
    return readHubSpecification(requireApiObjectField(body, 'specification', endpoint), endpoint)
  },

  async updateHubSpecificationStatus(id: number | string, isEnabled: boolean): Promise<HubSpecification> {
    const endpoint = `/api/admin/fitment-catalog/hub-specifications/${id}/status`
    const body = requireApiObject(
      readApiBody(await axios.patch(endpoint, { is_enabled: isEnabled }), endpoint),
      endpoint,
      'response body',
    )
    return readHubSpecification(requireApiObjectField(body, 'specification', endpoint), endpoint)
  },

  async removeHubSpecification(id: number | string) {
    const endpoint = `/api/admin/fitment-catalog/hub-specifications/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default fitmentHubSpecificationsApi

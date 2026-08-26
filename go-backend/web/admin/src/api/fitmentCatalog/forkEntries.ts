import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  readApiBody,
} from '@/utils/apiResponse'
import { readForkFitmentEntry } from './readers'
import type {
  ForkFitmentEntry,
  ForkFitmentEntryListPayload,
  ForkFitmentEntryPayload,
} from './types'

export const forkFitmentEntriesApi = {
  async listForkEntries(params: Record<string, unknown> = {}): Promise<ForkFitmentEntryListPayload> {
    const endpoint = '/api/admin/fitment-catalog/fork-entries'
    const body = requireApiObject(readApiBody(await axios.get(endpoint, { params }), endpoint), endpoint, 'response body')
    const entries = requireApiArrayField<ForkFitmentEntry>(body, 'entries', endpoint)
      .map((entry) => readForkFitmentEntry(entry, endpoint))
    return {
      entries,
      pagination: requireApiPagination(body, body, endpoint),
    }
  },

  async getForkEntry(id: number | string): Promise<ForkFitmentEntry> {
    const endpoint = `/api/admin/fitment-catalog/fork-entries/${id}`
    const body = requireApiObject(readApiBody(await axios.get(endpoint), endpoint), endpoint, 'response body')
    return readForkFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async createForkEntry(payload: ForkFitmentEntryPayload): Promise<ForkFitmentEntry> {
    const endpoint = '/api/admin/fitment-catalog/fork-entries'
    const body = requireApiObject(readApiBody(await axios.post(endpoint, payload), endpoint), endpoint, 'response body')
    return readForkFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async updateForkEntry(id: number | string, payload: ForkFitmentEntryPayload): Promise<ForkFitmentEntry> {
    const endpoint = `/api/admin/fitment-catalog/fork-entries/${id}`
    const body = requireApiObject(readApiBody(await axios.put(endpoint, payload), endpoint), endpoint, 'response body')
    return readForkFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async updateForkEntryStatus(id: number | string, isEnabled: boolean): Promise<ForkFitmentEntry> {
    const endpoint = `/api/admin/fitment-catalog/fork-entries/${id}/status`
    const body = requireApiObject(
      readApiBody(await axios.patch(endpoint, { is_enabled: isEnabled }), endpoint),
      endpoint,
      'response body',
    )
    return readForkFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async removeForkEntry(id: number | string) {
    const endpoint = `/api/admin/fitment-catalog/fork-entries/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default forkFitmentEntriesApi

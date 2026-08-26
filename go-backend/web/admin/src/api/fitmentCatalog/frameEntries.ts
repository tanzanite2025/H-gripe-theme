import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  readApiBody,
} from '@/utils/apiResponse'
import { readFrameFitmentEntry } from './readers'
import type {
  FrameFitmentEntry,
  FrameFitmentEntryListPayload,
  FrameFitmentEntryPayload,
} from './types'

export const frameFitmentEntriesApi = {
  async listFrameEntries(params: Record<string, unknown> = {}): Promise<FrameFitmentEntryListPayload> {
    const endpoint = '/api/admin/fitment-catalog/frame-entries'
    const body = requireApiObject(readApiBody(await axios.get(endpoint, { params }), endpoint), endpoint, 'response body')
    const entries = requireApiArrayField<FrameFitmentEntry>(body, 'entries', endpoint)
      .map((entry) => readFrameFitmentEntry(entry, endpoint))
    return {
      entries,
      pagination: requireApiPagination(body, body, endpoint),
    }
  },

  async getFrameEntry(id: number | string): Promise<FrameFitmentEntry> {
    const endpoint = `/api/admin/fitment-catalog/frame-entries/${id}`
    const body = requireApiObject(readApiBody(await axios.get(endpoint), endpoint), endpoint, 'response body')
    return readFrameFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async createFrameEntry(payload: FrameFitmentEntryPayload): Promise<FrameFitmentEntry> {
    const endpoint = '/api/admin/fitment-catalog/frame-entries'
    const body = requireApiObject(readApiBody(await axios.post(endpoint, payload), endpoint), endpoint, 'response body')
    return readFrameFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async updateFrameEntry(id: number | string, payload: FrameFitmentEntryPayload): Promise<FrameFitmentEntry> {
    const endpoint = `/api/admin/fitment-catalog/frame-entries/${id}`
    const body = requireApiObject(readApiBody(await axios.put(endpoint, payload), endpoint), endpoint, 'response body')
    return readFrameFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async updateFrameEntryStatus(id: number | string, isEnabled: boolean): Promise<FrameFitmentEntry> {
    const endpoint = `/api/admin/fitment-catalog/frame-entries/${id}/status`
    const body = requireApiObject(
      readApiBody(await axios.patch(endpoint, { is_enabled: isEnabled }), endpoint),
      endpoint,
      'response body',
    )
    return readFrameFitmentEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async removeFrameEntry(id: number | string) {
    const endpoint = `/api/admin/fitment-catalog/frame-entries/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default frameFitmentEntriesApi

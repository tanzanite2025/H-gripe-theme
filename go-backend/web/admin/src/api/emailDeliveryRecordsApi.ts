import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type EmailDeliveryRecordStatus = 'sending' | 'sent' | 'failed' | 'unknown'

export interface EmailDeliveryRecord {
  id: number
  event_type: string
  template_code: string
  locale: string
  template_version: number
  recipient_email: string
  subject: string
  reference_type?: string
  reference_number?: string
  status: EmailDeliveryRecordStatus
  attempt_count: number
  provider_code?: string
  last_error?: string
  first_attempt_at: string
  last_attempt_at: string
  sent_at?: string
  created_at: string
  updated_at: string
}

export interface EmailDeliveryRecordPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface EmailDeliveryRecordListResult {
  records: EmailDeliveryRecord[]
  pagination: EmailDeliveryRecordPagination
}

export interface EmailDeliveryRecordListFilters {
  page: number
  page_size: number
  status?: string
  search?: string
}

const emailDeliveryRecordsEndpoint = '/api/admin/email/deliveries'

const readEmailDeliveryRecordListResult = (response: unknown): EmailDeliveryRecordListResult => {
  const payload = requireApiObject(
    unwrapApiPayload(response, emailDeliveryRecordsEndpoint),
    emailDeliveryRecordsEndpoint,
  )
  const pagination = requireApiObjectField(
    payload,
    'pagination',
    emailDeliveryRecordsEndpoint,
  )
  return {
    records: requireApiArrayField<EmailDeliveryRecord>(payload, 'records', emailDeliveryRecordsEndpoint),
    pagination: {
      page: requireApiNumberField(pagination, 'page', emailDeliveryRecordsEndpoint),
      page_size: requireApiNumberField(pagination, 'page_size', emailDeliveryRecordsEndpoint),
      total: requireApiNumberField(pagination, 'total', emailDeliveryRecordsEndpoint),
      total_pages: requireApiNumberField(pagination, 'total_pages', emailDeliveryRecordsEndpoint),
    },
  }
}

export const emailDeliveryRecordsApi = {
  async listEmailDeliveryRecords(filters: EmailDeliveryRecordListFilters): Promise<EmailDeliveryRecordListResult> {
    return readEmailDeliveryRecordListResult(
      await axios.get(emailDeliveryRecordsEndpoint, { params: filters }),
    )
  },
}

export default emailDeliveryRecordsApi

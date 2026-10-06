import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiObject,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type EmailProviderEncryptionType = 'none' | 'starttls' | 'tls'

export type EmailProviderTestStatus = 'healthy' | 'failed' | ''

export interface EmailProviderSettingsRecord {
  id: number
  code: string
  name: string
  driver: string
  host: string
  port: number
  username: string
  has_password: boolean
  from_name: string
  from_email: string
  reply_to?: string
  encryption_type: EmailProviderEncryptionType
  is_active: boolean
  is_default: boolean
  last_tested_at?: string
  last_test_status?: EmailProviderTestStatus
  last_test_error?: string
  created_at: string
  updated_at: string
}

export interface SaveEmailProviderSettingsInput {
  code: string
  name: string
  driver?: string
  host: string
  port: number
  username: string
  password?: string
  from_name: string
  from_email: string
  reply_to?: string
  encryption_type: EmailProviderEncryptionType
  is_active: boolean
  is_default: boolean
}

const emailProviderSettingsEndpoint = '/api/admin/email/providers'

const readEmailProviderSettingsRecord = (
  response: unknown,
  endpoint: string,
): EmailProviderSettingsRecord => (
  requireApiObject(unwrapApiPayload(response, endpoint), endpoint) as EmailProviderSettingsRecord
)

const readEmailProviderSettingsList = (
  response: unknown,
  endpoint: string,
): EmailProviderSettingsRecord[] => {
  const payload = requireApiObject(unwrapApiPayload(response, endpoint), endpoint)
  return requireApiArrayField<EmailProviderSettingsRecord>(payload, 'providers', endpoint)
}

const readEmailProviderSettingsErrorMessage = (error: unknown): string => {
  if (!error || typeof error !== 'object') return '邮件通道操作失败'
  const response = (error as { response?: { data?: unknown } }).response
  const responseData = response?.data
  if (!responseData || typeof responseData !== 'object') return '邮件通道操作失败'

  const payload = responseData as { error?: unknown; message?: unknown }
  if (typeof payload.error === 'string' && payload.error.trim()) return payload.error
  if (typeof payload.message === 'string' && payload.message.trim()) return payload.message
  return '邮件通道操作失败'
}

export const getEmailProviderSettingsErrorMessage = readEmailProviderSettingsErrorMessage

export const emailProviderSettingsApi = {
  async listEmailProviderSettings(): Promise<EmailProviderSettingsRecord[]> {
    return readEmailProviderSettingsList(
      await axios.get(emailProviderSettingsEndpoint),
      emailProviderSettingsEndpoint,
    )
  },

  async createEmailProviderSettings(input: SaveEmailProviderSettingsInput): Promise<EmailProviderSettingsRecord> {
    return readEmailProviderSettingsRecord(
      await axios.post(emailProviderSettingsEndpoint, input),
      emailProviderSettingsEndpoint,
    )
  },

  async updateEmailProviderSettings(id: number, input: SaveEmailProviderSettingsInput): Promise<EmailProviderSettingsRecord> {
    const endpoint = `${emailProviderSettingsEndpoint}/${id}`
    return readEmailProviderSettingsRecord(await axios.put(endpoint, input), endpoint)
  },

  async setEmailProviderSettingsActive(id: number, active: boolean): Promise<EmailProviderSettingsRecord> {
    const endpoint = `${emailProviderSettingsEndpoint}/${id}/active`
    return readEmailProviderSettingsRecord(await axios.patch(endpoint, { active }), endpoint)
  },

  async setDefaultEmailProviderSettings(id: number): Promise<EmailProviderSettingsRecord> {
    const endpoint = `${emailProviderSettingsEndpoint}/${id}/default`
    return readEmailProviderSettingsRecord(await axios.post(endpoint), endpoint)
  },

  async sendEmailProviderSettingsTestEmail(id: number, targetEmail: string): Promise<EmailProviderSettingsRecord> {
    const endpoint = `${emailProviderSettingsEndpoint}/${id}/test`
    return readEmailProviderSettingsRecord(
      await axios.post(endpoint, { target_email: targetEmail }),
      endpoint,
    )
  },
}

export default emailProviderSettingsApi

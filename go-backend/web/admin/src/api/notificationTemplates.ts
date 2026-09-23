import axios from '@/utils/axios'
import { requireApiArrayField, requireApiObject, requireApiObjectField, unwrapApiPayload } from '@/utils/apiResponse'

export interface NotificationTemplate {
  id: number
  code: string
  locale: string
  category: string
  name: string
  subject_template: string
  body_html: string
  body_text: string
  allowed_variables: string[]
  required_variables: string[]
  is_enabled: boolean
  version: number
  updated_at: string
}

export interface NotificationTemplateVersion extends NotificationTemplate {
  template_id: number
  changed_by_user_id?: number
  change_reason?: string
  created_at: string
}

export interface NotificationTemplatePreview {
  subject: string
  html: string
  text: string
}

const endpoint = '/api/admin/email/templates'

const payload = (response: unknown, path: string) => unwrapApiPayload(response, path)

export const notificationTemplatesApi = {
  async list(): Promise<NotificationTemplate[]> {
    const path = endpoint
    const data = requireApiObject(payload(await axios.get(path), path), path)
    return requireApiArrayField<NotificationTemplate>(data, 'templates', path)
  },
  async get(id: number): Promise<NotificationTemplate> {
    const path = `${endpoint}/${id}`
    return requireApiObject(payload(await axios.get(path), path), path) as NotificationTemplate
  },
  async versions(id: number): Promise<NotificationTemplateVersion[]> {
    const path = `${endpoint}/${id}/versions`
    const data = requireApiObject(payload(await axios.get(path), path), path)
    return requireApiArrayField<NotificationTemplateVersion>(data, 'versions', path)
  },
  async save(input: Partial<NotificationTemplate> & { change_reason?: string }): Promise<NotificationTemplate> {
    const path = `${endpoint}/${input.id}`
    return requireApiObject(payload(await axios.put(path, input), path), path) as NotificationTemplate
  },
  async preview(input: { code: string; locale: string; variables: Record<string, string> }): Promise<NotificationTemplatePreview> {
    const path = `${endpoint}/preview`
    return requireApiObject(payload(await axios.post(path, input), path), path) as NotificationTemplatePreview
  },
  async rollback(id: number, version: number, changeReason?: string): Promise<NotificationTemplate> {
    const path = `${endpoint}/${id}/versions/${version}/rollback`
    return requireApiObject(payload(await axios.post(path, { change_reason: changeReason || '' }), path), path) as NotificationTemplate
  },
}

export default notificationTemplatesApi

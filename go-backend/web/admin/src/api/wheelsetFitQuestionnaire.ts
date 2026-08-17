import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export type WheelsetFitQuestionnaireVersionStatus = 'draft' | 'published' | 'archived' | string

export interface WheelsetFitQuestionnaire {
  id: number
  slug: string
  product_category_slug: string
  source_locale: string
  is_enabled: boolean
}

export interface WheelsetFitQuestionTranslation {
  id?: number
  locale: string
  prompt: string
  help_title: string
  help_body: string
  source_revision: number
  translated_revision: number
}

export interface WheelsetFitQuestionOptionTranslation {
  id?: number
  locale: string
  label: string
  description: string
  source_revision: number
  translated_revision: number
}

export interface WheelsetFitQuestionOption {
  id: number
  question_id: number
  option_key: string
  answer_value: string
  sort_order: number
  is_unknown: boolean
  is_enabled: boolean
  product_filter_effects: Record<string, unknown>
  source_revision: number
  translations: WheelsetFitQuestionOptionTranslation[]
}

export interface WheelsetFitQuestion {
  id: number
  questionnaire_version_id: number
  question_key: string
  answer_key: string
  sort_order: number
  input_mode: 'single_choice' | string
  is_required: boolean
  allow_unknown: boolean
  is_enabled: boolean
  source_revision: number
  translations: WheelsetFitQuestionTranslation[]
  options: WheelsetFitQuestionOption[]
}

export interface WheelsetFitQuestionnaireVersion {
  id: number
  questionnaire_id: number
  version_number: number
  status: WheelsetFitQuestionnaireVersionStatus
  published_at?: string | null
  questionnaire?: WheelsetFitQuestionnaire
  questions: WheelsetFitQuestion[]
}

export interface WheelsetFitQuestionTranslationPayload {
  locale: string
  prompt: string
  help_title: string
  help_body: string
}

export interface WheelsetFitQuestionOptionTranslationPayload {
  locale: string
  label: string
  description: string
}

export interface WheelsetFitQuestionOptionPayload {
  option_key: string
  answer_value: string
  sort_order: number
  is_unknown: boolean
  is_enabled: boolean
  product_filter_effects: Record<string, unknown>
  translations: WheelsetFitQuestionOptionTranslationPayload[]
}

export interface WheelsetFitQuestionPayload {
  question_key: string
  answer_key: string
  sort_order: number
  input_mode: 'single_choice'
  is_required: boolean
  allow_unknown: boolean
  is_enabled: boolean
  translations: WheelsetFitQuestionTranslationPayload[]
  options: WheelsetFitQuestionOptionPayload[]
}

export interface WheelsetFitQuestionnaireValidationIssue {
  severity: 'error' | 'warning' | 'info' | string
  code: string
  message: string
  question_id?: number
  question_key?: string
  option_key?: string
  locale?: string
}

export interface WheelsetFitQuestionnaireValidationResult {
  valid: boolean
  issues: WheelsetFitQuestionnaireValidationIssue[]
}

const baseEndpoint = '/api/admin/wheelset-fit-questionnaire'

const readVersion = (response: unknown, endpoint: string): WheelsetFitQuestionnaireVersion => {
  const envelope = requireApiObject(unwrapApiPayload(response, endpoint), endpoint, 'response payload')
  const version = requireApiObjectField(envelope, 'data', endpoint)
  requireApiNumberField(version, 'id', endpoint)
  requireApiNumberField(version, 'version_number', endpoint)
  requireApiStringField(version, 'status', endpoint)
  requireApiArrayField(version, 'questions', endpoint)
  return version as unknown as WheelsetFitQuestionnaireVersion
}

const readValidation = (response: unknown, endpoint: string): WheelsetFitQuestionnaireValidationResult => {
  const envelope = requireApiObject(unwrapApiPayload(response, endpoint), endpoint, 'response payload')
  const result = requireApiObjectField(envelope, 'data', endpoint)
  requireApiArrayField(result, 'issues', endpoint)
  return result as unknown as WheelsetFitQuestionnaireValidationResult
}

export const wheelsetFitQuestionnaireApi = {
  async getCurrentVersion(): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/current`
    return readVersion(await axios.get(endpoint), endpoint)
  },

  async createDraft(): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/draft`
    return readVersion(await axios.post(endpoint), endpoint)
  },

  async createQuestion(payload: WheelsetFitQuestionPayload): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/questions`
    return readVersion(await axios.post(endpoint, payload), endpoint)
  },

  async updateQuestion(id: number, payload: WheelsetFitQuestionPayload): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/questions/${id}`
    return readVersion(await axios.put(endpoint, payload), endpoint)
  },

  async deleteQuestion(id: number): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/questions/${id}`
    return readVersion(await axios.delete(endpoint), endpoint)
  },

  async reorderQuestions(questionIDs: number[]): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/questions/order`
    return readVersion(await axios.put(endpoint, { question_ids: questionIDs }), endpoint)
  },

  async validateVersion(versionID: number): Promise<WheelsetFitQuestionnaireValidationResult> {
    const endpoint = `${baseEndpoint}/versions/${versionID}/validate`
    return readValidation(await axios.post(endpoint), endpoint)
  },

  async publishVersion(versionID: number): Promise<WheelsetFitQuestionnaireVersion> {
    const endpoint = `${baseEndpoint}/versions/${versionID}/publish`
    return readVersion(await axios.post(endpoint), endpoint)
  },
}

export default wheelsetFitQuestionnaireApi

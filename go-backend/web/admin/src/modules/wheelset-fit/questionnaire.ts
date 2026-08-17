import type {
  WheelsetFitQuestionOptionTranslationPayload,
  WheelsetFitQuestionTranslationPayload,
} from '@/api/wheelsetFitQuestionnaire'

export interface WheelsetFitQuestionOptionForm {
  client_id: number
  option_key: string
  answer_value: string
  sort_order: number
  is_unknown: boolean
  is_enabled: boolean
  product_filter_effects_json: string
  translations: WheelsetFitQuestionOptionTranslationPayload[]
}

export interface WheelsetFitQuestionForm {
  id?: number
  question_key: string
  answer_key: string
  sort_order: number
  input_mode: 'single_choice'
  is_required: boolean
  allow_unknown: boolean
  is_enabled: boolean
  translations: WheelsetFitQuestionTranslationPayload[]
  options: WheelsetFitQuestionOptionForm[]
}

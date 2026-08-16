export const WHEELSET_SELECTION_REQUEST_MESSAGE_TYPE = 'wheelset_selection_request' as const
export const WHEELSET_PRODUCT_CATEGORY_SLUG = 'wheelset' as const
export const WHEELSET_SELECTION_ASSISTANT_SLUG = 'wheelset-fit-helper' as const

export type WheelsetSelectionRequestMessageType = typeof WHEELSET_SELECTION_REQUEST_MESSAGE_TYPE

export type WheelsetSelectionAssistantSource =
  | 'guides/wheelset-buyers'
  | 'quick-buy/wheelset-selection-assistant'
  | 'chat'
  | string

export type WheelsetSelectionAssistantStepKey =
  | 'start'
  | 'knowledge'
  | 'bike-basics'
  | 'hub'
  | 'rim'
  | 'spoke'
  | 'nipple'
  | 'summary'

export type WheelsetSelectionKnownSectionKey = 'hub' | 'rim' | 'spoke' | 'nipple'

export type WheelsetSelectionCompletionStatus = 'draft' | 'partial' | 'complete'

export type WheelsetSelectionRecommendedNextAction =
  | 'support_review'
  | 'quick_buy_direct_select'
  | 'product_match'
  | 'ai_recommendation'

export type WheelsetSelectionAnswerValue = string | number | boolean | string[] | null

export type WheelsetSelectionAnswers = Record<string, WheelsetSelectionAnswerValue>

export type WheelsetSelectionKnownSections = Partial<Record<WheelsetSelectionKnownSectionKey, boolean>>

export interface WheelsetSelectionLocalizedText {
  [locale: string]: string
}

export interface WheelsetSelectionGraphQueryEffects {
  keyword?: string
  spec_filters?: Record<string, string[]>
}

export interface WheelsetSelectionGraphOption {
  key: string
  label?: WheelsetSelectionLocalizedText
  description?: WheelsetSelectionLocalizedText
  answer_effects?: Record<string, string>
  query_effects?: WheelsetSelectionGraphQueryEffects
  next_node_key?: string
}

export interface WheelsetSelectionGraphNode {
  key: string
  type: 'question' | 'terminal' | 'support'
  prompt?: WheelsetSelectionLocalizedText
  helper?: WheelsetSelectionLocalizedText
  options?: WheelsetSelectionGraphOption[]
  editor?: {
    x: number
    y: number
  }
}

export interface WheelsetSelectionAssistantConfig {
  kind: string
  schema_version: number
  entry_node_key: string
  base_product_query: {
    category_slug: string
  }
  nodes: WheelsetSelectionGraphNode[]
}

export interface WheelsetSelectionAssistantFlow {
  id: number
  slug: string
  name: string
  description: string
  product_category_slug: string
  is_enabled: boolean
  sort_order: number
  version: {
    id: number
    version_number: number
    status: string
    config: WheelsetSelectionAssistantConfig
    published_at?: string | null
  }
}

export interface WheelsetSelectionProductQuery {
  category_slug: string
  keyword?: string
  spec_filters: Record<string, string[]>
}

export interface WheelsetSelectionPathEntry {
  node_key: string
  option_key: string
  label: string
  next_node_key: string
}

export interface WheelsetSelectionRequestMetadata {
  kind: WheelsetSelectionRequestMessageType
  version: 1
  source: WheelsetSelectionAssistantSource
  flow_slug?: string
  flow_version?: number
  completion_status: WheelsetSelectionCompletionStatus
  known_sections: WheelsetSelectionKnownSections
  answers: WheelsetSelectionAnswers
  answer_labels?: Record<string, string>
  selected_path?: WheelsetSelectionPathEntry[]
  product_query?: WheelsetSelectionProductQuery
  unknowns: string[]
  recommended_next_action: WheelsetSelectionRecommendedNextAction
}

export interface WheelsetSelectionRequestDraft {
  message: string
  message_type: WheelsetSelectionRequestMessageType
  metadata: WheelsetSelectionRequestMetadata
}

export interface WheelsetSelectionAssistantShellStep {
  key: WheelsetSelectionAssistantStepKey | string
  label: string
}

export interface WheelsetSelectionOption {
  value: string
  label: string
  description?: string
}

export interface WheelsetSelectionQuestion {
  key: string
  prompt: string
  helper?: string
  options: WheelsetSelectionOption[]
}

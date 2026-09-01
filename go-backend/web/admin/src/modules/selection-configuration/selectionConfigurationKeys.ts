export const selectionConfigurationKeyKindQuestionKey = 'question_key' as const
export const selectionConfigurationKeyKindAnswerKey = 'answer_key' as const

export type SelectionConfigurationKeyKind = typeof selectionConfigurationKeyKindQuestionKey | typeof selectionConfigurationKeyKindAnswerKey

export const selectionConfigurationKeyKindOptions: Array<{ label: string, value: SelectionConfigurationKeyKind }> = [
  { label: '问题 KEY', value: selectionConfigurationKeyKindQuestionKey },
  { label: '回答 KEY', value: selectionConfigurationKeyKindAnswerKey },
]

export const selectionConfigurationKeyKindLabels: Record<SelectionConfigurationKeyKind, string> = {
  [selectionConfigurationKeyKindQuestionKey]: '问题 KEY',
  [selectionConfigurationKeyKindAnswerKey]: '回答 KEY',
}

export const selectionConfigurationKeyKindDescriptions: Record<SelectionConfigurationKeyKind, string> = {
  [selectionConfigurationKeyKindQuestionKey]: '问卷中的问题主键，只允许从注册表里选。',
  [selectionConfigurationKeyKindAnswerKey]: '问卷中的回答主键，只允许从注册表里选。',
}

export interface SelectionConfigurationKeyEditorForm {
  id?: number
  kind: SelectionConfigurationKeyKind
  code: string
  display_label: string
  description: string
  is_enabled: boolean
  sort_order: number
}

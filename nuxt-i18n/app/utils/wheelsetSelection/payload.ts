import type {
  WheelsetSelectionAssistantFlow,
  WheelsetSelectionAssistantSource,
  WheelsetSelectionAnswers,
  WheelsetSelectionCompletionStatus,
  WheelsetSelectionKnownSections,
  WheelsetSelectionPathEntry,
  WheelsetSelectionProductQuery,
  WheelsetSelectionRequestDraft,
} from '~/types/wheelsetSelectionAssistant'
import { WHEELSET_SELECTION_REQUEST_MESSAGE_TYPE } from '~/types/wheelsetSelectionAssistant'

interface WheelsetSelectionPayloadInput {
  source: WheelsetSelectionAssistantSource
  flow: WheelsetSelectionAssistantFlow | null
  answers: WheelsetSelectionAnswers
  answerLabels: Record<string, string>
  selectedPath: WheelsetSelectionPathEntry[]
  productQuery: WheelsetSelectionProductQuery
  knownSections?: WheelsetSelectionKnownSections
  unknowns?: string[]
  completionStatus?: WheelsetSelectionCompletionStatus
}

export const buildWheelsetSelectionRequestDraft = (
  input: WheelsetSelectionPayloadInput,
): WheelsetSelectionRequestDraft => {
  const answerLines = input.selectedPath.length
    ? input.selectedPath.map((entry, index) => `${index + 1}. ${entry.label}`).join('\n')
    : 'No fit answers selected yet.'
  const filterLines = Object.entries(input.productQuery.spec_filters)
    .map(([key, values]) => `${key}: ${values.join(', ')}`)
  const message = [
    'Wheelset fit request',
    answerLines,
    filterLines.length ? `Product filters:\n${filterLines.join('\n')}` : 'Product filters: wheelset category only',
  ].join('\n')

  return {
    message,
    message_type: WHEELSET_SELECTION_REQUEST_MESSAGE_TYPE,
    metadata: {
      kind: WHEELSET_SELECTION_REQUEST_MESSAGE_TYPE,
      version: 1,
      source: input.source,
      flow_slug: input.flow?.slug,
      flow_version: input.flow?.version?.version_number,
      completion_status: input.completionStatus || 'partial',
      known_sections: input.knownSections || {},
      answers: input.answers,
      answer_labels: input.answerLabels,
      selected_path: input.selectedPath,
      product_query: input.productQuery,
      unknowns: input.unknowns || [],
      recommended_next_action: 'support_review',
    },
  }
}

import type { InjectionKey, Ref } from 'vue'
import { inject } from 'vue'

export interface WheelsetSelectionAssistantQuestionPaginationContext {
  total: Ref<number>
  activeIndex: Ref<number>
  reachableIndex: Ref<number>
  registerJumpToIndexHandler: (handler: ((questionIndex: number) => void) | null) => void
}

export const wheelsetSelectionAssistantQuestionPaginationKey: InjectionKey<WheelsetSelectionAssistantQuestionPaginationContext> = Symbol('wheelset-selection-assistant-question-pagination')

export const useWheelsetSelectionAssistantQuestionPagination = () => {
  return inject(wheelsetSelectionAssistantQuestionPaginationKey, null)
}

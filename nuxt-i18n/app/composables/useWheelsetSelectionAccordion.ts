import { computed, inject, provide, ref, type ComputedRef, type InjectionKey, type Ref } from 'vue'

export type WheelsetSelectionAccordionSection = 'question' | 'results'

type WheelsetSelectionAccordionContext = {
  activeSection: Ref<WheelsetSelectionAccordionSection>
  isExpanded: (section: WheelsetSelectionAccordionSection) => ComputedRef<boolean>
  toggle: (section: WheelsetSelectionAccordionSection) => void
}

const wheelsetSelectionAccordionKey: InjectionKey<WheelsetSelectionAccordionContext> = Symbol(
  'wheelset-selection-accordion',
)

export const provideWheelsetSelectionAccordion = (
  defaultSection: WheelsetSelectionAccordionSection = 'question',
) => {
  const activeSection = ref<WheelsetSelectionAccordionSection>(defaultSection)
  const context: WheelsetSelectionAccordionContext = {
    activeSection,
    isExpanded: section => computed(() => activeSection.value === section),
    toggle: section => {
      if (activeSection.value !== section) {
        activeSection.value = section
      }
    },
  }

  provide(wheelsetSelectionAccordionKey, context)
  return context
}

export const useWheelsetSelectionAccordion = () => inject(wheelsetSelectionAccordionKey, null)

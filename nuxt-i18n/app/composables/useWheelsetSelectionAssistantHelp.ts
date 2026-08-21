import { inject, provide, ref, type InjectionKey, type Ref } from 'vue'

type WheelsetSelectionAssistantHelp = {
  title: Ref<string>
  content: Ref<string>
  setHelp: (help: { title?: string; content?: string }) => void
  clearHelp: () => void
}

const wheelsetSelectionAssistantHelpKey: InjectionKey<WheelsetSelectionAssistantHelp> = Symbol(
  'wheelset-selection-assistant-help',
)

export const provideWheelsetSelectionAssistantHelp = () => {
  const title = ref('')
  const content = ref('')
  const context: WheelsetSelectionAssistantHelp = {
    title,
    content,
    setHelp: help => {
      title.value = String(help.title || '')
      content.value = String(help.content || '').trim()
    },
    clearHelp: () => {
      title.value = ''
      content.value = ''
    },
  }

  provide(wheelsetSelectionAssistantHelpKey, context)
  return context
}

export const useWheelsetSelectionAssistantHelp = () => inject(wheelsetSelectionAssistantHelpKey, null)

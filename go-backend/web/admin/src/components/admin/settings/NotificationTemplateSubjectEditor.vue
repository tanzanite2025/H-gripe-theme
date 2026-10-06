<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
  templateVariableDisplayValues?: Record<string, string>
}>(), {
  modelValue: '',
  placeholder: '请输入邮件标题',
  disabled: false,
  templateVariableDisplayValues: () => ({}),
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = ref<HTMLElement | null>(null)
const syncing = ref(false)
let lastEmittedModelValue: string | null = null

const renderTemplateVariablesAsDisplayValues = (source: string): string => {
  if (!source || typeof document === 'undefined') return source

  const container = document.createElement('div')
  container.textContent = source
  const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT)
  const textNodes: Text[] = []
  let currentNode = walker.nextNode()
  while (currentNode) {
    textNodes.push(currentNode as Text)
    currentNode = walker.nextNode()
  }

  const variablePattern = /\{\{\s*([a-z][a-z0-9_]*)\s*\}\}/gi
  textNodes.forEach((textNode) => {
    const value = textNode.nodeValue || ''
    if (!variablePattern.test(value)) {
      variablePattern.lastIndex = 0
      return
    }
    variablePattern.lastIndex = 0

    const fragment = document.createDocumentFragment()
    let lastIndex = 0
    value.replace(variablePattern, (match, variableName: string, offset: number) => {
      if (offset > lastIndex) fragment.append(value.slice(lastIndex, offset))
      const displayNode = document.createElement('span')
      displayNode.dataset.notificationTemplateVariable = variableName
      displayNode.contentEditable = 'false'
      displayNode.className = 'notification-template-subject-editor__template-variable'
      displayNode.textContent = props.templateVariableDisplayValues[variableName] || '示例内容'
      fragment.append(displayNode)
      lastIndex = offset + match.length
      return match
    })
    if (lastIndex < value.length) fragment.append(value.slice(lastIndex))
    textNode.replaceWith(fragment)
  })

  return container.innerHTML
}

const serializeTemplateVariablesToPlainText = (source: HTMLElement): string => {
  const clonedSource = source.cloneNode(true) as HTMLElement
  clonedSource.querySelectorAll<HTMLElement>('[data-notification-template-variable]').forEach((displayNode) => {
    const variableName = displayNode.dataset.notificationTemplateVariable
    if (variableName) displayNode.replaceWith(`{{${variableName}}}`)
  })
  return (clonedSource.innerText || clonedSource.textContent || '')
    .replace(/\u00a0/g, ' ')
    .replace(/[\r\n]+/g, ' ')
    .replace(/\s{2,}/g, ' ')
    .trim()
}

const syncEditor = async (): Promise<void> => {
  if (!editor.value) return
  if (editor.value === document.activeElement && props.modelValue === lastEmittedModelValue) return
  syncing.value = true
  await nextTick()
  if (editor.value) editor.value.innerHTML = renderTemplateVariablesAsDisplayValues(props.modelValue || '')
  lastEmittedModelValue = props.modelValue
  syncing.value = false
}

const emitContent = (): void => {
  if (!editor.value || syncing.value || props.disabled) return
  const serializedContent = serializeTemplateVariablesToPlainText(editor.value)
  lastEmittedModelValue = serializedContent
  emit('update:modelValue', serializedContent)
}

onMounted(syncEditor)
watch(() => props.modelValue, syncEditor)
</script>

<template>
  <div
    ref="editor"
    :contenteditable="disabled ? 'false' : 'true'"
    role="textbox"
    aria-multiline="false"
    class="notification-template-subject-editor w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm outline-none focus:ring-2 focus:ring-ring"
    :class="{ 'pointer-events-none opacity-60': disabled }"
    :data-placeholder="placeholder"
    @input="emitContent"
    @blur="emitContent"
    @keydown.enter.prevent
  />
</template>

<style scoped>
.notification-template-subject-editor:empty::before {
  color: hsl(var(--muted-foreground));
  content: attr(data-placeholder);
  pointer-events: none;
}

.notification-template-subject-editor :deep(.notification-template-subject-editor__template-variable) {
  background: hsl(var(--muted));
  border-radius: 0.25rem;
  color: hsl(var(--foreground));
  padding: 0.05rem 0.25rem;
  white-space: nowrap;
}
</style>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import {
  Bold,
  Heading2,
  Heading3,
  Italic,
  Link2,
  List,
  ListOrdered,
  Pilcrow,
  Quote,
  Redo2,
  Undo2,
} from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const props = withDefaults(defineProps<{
  modelValue: string
  disabled?: boolean
  templateVariableDisplayValues?: Record<string, string>
}>(), {
  modelValue: '',
  disabled: false,
  templateVariableDisplayValues: () => ({}),
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = ref<HTMLElement | null>(null)
const linkURL = ref('')
const syncing = ref(false)
let lastEmittedModelValue: string | null = null
let savedRange: Range | null = null

const renderTemplateVariablesAsDisplayValues = (source: string): string => {
  if (!source || typeof document === 'undefined') return source

  const container = document.createElement('div')
  container.innerHTML = source
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
      const displayValue = props.templateVariableDisplayValues[variableName] || '示例内容'
      const displayNode = document.createElement('span')
      displayNode.dataset.notificationTemplateVariable = variableName
      displayNode.contentEditable = 'false'
      displayNode.className = 'rich-text-editor__template-variable'
      displayNode.textContent = displayValue
      fragment.append(displayNode)
      lastIndex = offset + match.length
      return match
    })
    if (lastIndex < value.length) fragment.append(value.slice(lastIndex))
    textNode.replaceWith(fragment)
  })

  return container.innerHTML
}

const serializeTemplateVariablesFromEditor = (source: HTMLElement): string => {
  const clonedSource = source.cloneNode(true) as HTMLElement
  clonedSource.querySelectorAll<HTMLElement>('[data-notification-template-variable]').forEach((displayNode) => {
    const variableName = displayNode.dataset.notificationTemplateVariable
    if (variableName) displayNode.replaceWith(`{{${variableName}}}`)
  })
  return clonedSource.innerHTML
}

const isSelectionInsideEditor = (): boolean => {
  if (!editor.value) return false
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return false
  const node = selection.anchorNode
  return Boolean(node && editor.value.contains(node))
}

const saveSelection = (): void => {
  if (!isSelectionInsideEditor()) return
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  savedRange = selection.getRangeAt(0).cloneRange()
}

const restoreSelection = (): void => {
  editor.value?.focus()
  if (!savedRange) return
  const selection = window.getSelection()
  selection?.removeAllRanges()
  selection?.addRange(savedRange)
}

const syncEditor = async (): Promise<void> => {
  if (!editor.value) return
  if (editor.value === document.activeElement && props.modelValue === lastEmittedModelValue) return
  syncing.value = true
  await nextTick()
  editor.value.innerHTML = renderTemplateVariablesAsDisplayValues(props.modelValue || '')
  lastEmittedModelValue = props.modelValue
  syncing.value = false
}

onMounted(syncEditor)
watch(() => props.modelValue, syncEditor)

const emitContent = (): void => {
  if (!editor.value || syncing.value || props.disabled) return
  const serializedContent = serializeTemplateVariablesFromEditor(editor.value)
  lastEmittedModelValue = serializedContent
  emit('update:modelValue', serializedContent)
  saveSelection()
}

const exec = (command: string, value?: string): void => {
  restoreSelection()
  document.execCommand(command, false, value)
  emitContent()
}

const setBlock = (tag: 'P' | 'H2' | 'H3' | 'BLOCKQUOTE'): void => {
  exec('formatBlock', tag)
}

const escapeAttribute = (value: string): string => value
  .replace(/&/g, '&amp;')
  .replace(/"/g, '&quot;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')

const insertLink = (): void => {
  const value = linkURL.value.trim()
  if (!value || !/^(https?:\/\/|mailto:|\/(?!\/))/i.test(value)) return

  restoreSelection()
  const selection = window.getSelection()
  if (selection && selection.rangeCount > 0 && !selection.getRangeAt(0).collapsed) {
    document.execCommand('createLink', false, value)
    emitContent()
  } else {
    const safeURL = escapeAttribute(value)
    document.execCommand('insertHTML', false, `<a href="${safeURL}">${safeURL}</a>`)
    emitContent()
  }
  linkURL.value = ''
}
</script>

<template>
  <div class="flex flex-col rounded-lg border bg-muted/15 p-2" :class="{ 'pointer-events-none opacity-60': disabled }">
    <div class="flex shrink-0 flex-wrap items-center gap-1 border-b pb-2">
      <Button type="button" variant="ghost" size="icon-sm" title="撤销" @click="exec('undo')">
        <Undo2 class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="重做" @click="exec('redo')">
        <Redo2 class="size-4" />
      </Button>
      <span class="mx-1 h-5 w-px bg-border" aria-hidden="true" />
      <Button type="button" variant="ghost" size="icon-sm" title="段落" @click="setBlock('P')">
        <Pilcrow class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="二级标题" @click="setBlock('H2')">
        <Heading2 class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="三级标题" @click="setBlock('H3')">
        <Heading3 class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="引用" @click="setBlock('BLOCKQUOTE')">
        <Quote class="size-4" />
      </Button>
      <span class="mx-1 h-5 w-px bg-border" aria-hidden="true" />
      <Button type="button" variant="ghost" size="icon-sm" title="加粗" @click="exec('bold')">
        <Bold class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="斜体" @click="exec('italic')">
        <Italic class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="无序列表" @click="exec('insertUnorderedList')">
        <List class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="有序列表" @click="exec('insertOrderedList')">
        <ListOrdered class="size-4" />
      </Button>
      <div class="flex min-w-[12rem] flex-1 items-center gap-1">
        <Input
          v-model="linkURL"
          class="h-7 min-w-0 text-xs"
          placeholder="链接 URL"
          @keydown.enter.prevent="insertLink"
        />
        <Button type="button" variant="outline" size="icon-sm" title="插入链接" @click="insertLink">
          <Link2 class="size-4" />
        </Button>
      </div>
    </div>

    <div
      ref="editor"
      :contenteditable="disabled ? 'false' : 'true'"
      role="textbox"
      aria-multiline="true"
      class="rich-text-editor__canvas mt-2 min-h-48 flex-1 overflow-y-auto rounded-md bg-background px-4 py-3 text-sm leading-6 outline-none focus:ring-2 focus:ring-ring"
      data-placeholder="请输入正文，可使用段落、标题、加粗、斜体、列表、引用和链接"
      @focus="saveSelection"
      @keyup="saveSelection"
      @mouseup="saveSelection"
      @input="emitContent"
      @blur="emitContent"
    />
  </div>
</template>

<style scoped>
.rich-text-editor__canvas:empty::before {
  color: hsl(var(--muted-foreground));
  content: attr(data-placeholder);
  pointer-events: none;
}

.rich-text-editor__canvas :deep(h2),
.rich-text-editor__canvas :deep(h3) {
  margin: 0.85rem 0 0.45rem;
  font-weight: 800;
}

.rich-text-editor__canvas :deep(p),
.rich-text-editor__canvas :deep(ul),
.rich-text-editor__canvas :deep(ol),
.rich-text-editor__canvas :deep(blockquote) {
  margin: 0 0 0.75rem;
}

.rich-text-editor__canvas :deep(ul),
.rich-text-editor__canvas :deep(ol) {
  padding-left: 1.3rem;
  list-style-position: outside;
}

.rich-text-editor__canvas :deep(ul) {
  list-style-type: disc;
}

.rich-text-editor__canvas :deep(ol) {
  list-style-type: decimal;
}

.rich-text-editor__canvas :deep(blockquote) {
  border-left: 3px solid hsl(var(--primary));
  color: hsl(var(--muted-foreground));
  padding-left: 0.85rem;
}

.rich-text-editor__canvas :deep(a) {
  color: hsl(var(--primary));
  text-decoration: underline;
}

.rich-text-editor__canvas :deep(.rich-text-editor__template-variable) {
  background: hsl(var(--muted));
  border-radius: 0.25rem;
  color: hsl(var(--foreground));
  padding: 0.05rem 0.25rem;
  white-space: nowrap;
}
</style>

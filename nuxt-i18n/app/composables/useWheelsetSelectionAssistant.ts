import { computed, ref, watch } from 'vue'
import { useI18n, useRuntimeConfig } from '#imports'
import type {
  WheelsetSelectionAnswers,
  WheelsetSelectionAssistantConfig,
  WheelsetSelectionAssistantFlow,
  WheelsetSelectionGraphNode,
  WheelsetSelectionPathEntry,
  WheelsetSelectionProductQuery,
  WheelsetSelectionQuestion,
} from '~/types/wheelsetSelectionAssistant'
import {
  WHEELSET_PRODUCT_CATEGORY_SLUG,
  WHEELSET_SELECTION_ASSISTANT_SLUG,
} from '~/types/wheelsetSelectionAssistant'

interface AssistantSelectionState {
  nodeKey: string
  optionKey: string
  label: string
  nextNodeKey: string
  answerEffects: Record<string, string>
  queryEffects: {
    keyword?: string
    spec_filters?: Record<string, string[]>
  }
}

const normalizeLocale = (value: unknown) => String(value || '').trim().replace(/-/g, '_').toLowerCase()

const localizedText = (
  value: Record<string, string> | undefined,
  locale: string,
): string => {
  if (!value) return ''
  const normalized = normalizeLocale(locale)
  const base = normalized.split('_')[0] || ''
  if (normalized === 'zh_cn' || base === 'zh') {
    return String(value[normalized] || value[base] || value.zh_cn || value.en || '')
  }
  return String(value[normalized] || value[base] || value.en || '')
}

const cloneFilters = (filters: Record<string, string[]>): Record<string, string[]> => (
  Object.fromEntries(
    Object.entries(filters).map(([key, values]) => [key, [...values]]),
  )
)

const mergeQueryEffects = (
  target: Record<string, string[]>,
  effects: AssistantSelectionState['queryEffects'],
) => {
  for (const [key, values] of Object.entries(effects.spec_filters || {})) {
    const current = target[key] || []
    target[key] = Array.from(new Set([...current, ...values.map(value => String(value).trim()).filter(Boolean)]))
  }
}

export const useWheelsetSelectionAssistant = (
  flowSlug = WHEELSET_SELECTION_ASSISTANT_SLUG,
) => {
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  const publicBaseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  const flow = ref<WheelsetSelectionAssistantFlow | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const currentNodeKey = ref('')
  const path = ref<AssistantSelectionState[]>([])

  const assistantConfig = computed<WheelsetSelectionAssistantConfig | null>(() => flow.value?.version?.config || null)
  const nodesByKey = computed(() => new Map(
    (assistantConfig.value?.nodes || []).map(node => [node.key, node]),
  ))
  const currentNode = computed<WheelsetSelectionGraphNode | null>(() => (
    nodesByKey.value.get(currentNodeKey.value) || null
  ))
  const currentQuestion = computed<WheelsetSelectionQuestion | null>(() => {
    const node = currentNode.value
    if (!node || node.type !== 'question') return null
    return {
      key: node.key,
      prompt: localizedText(node.prompt, String(locale.value)),
      helpTitle: localizedText(node.help_title, String(locale.value)) || undefined,
      helpBody: localizedText(node.help_body, String(locale.value)) || undefined,
      options: (node.options || []).map(option => ({
        value: option.key,
        label: localizedText(option.label, String(locale.value)) || option.key,
        description: localizedText(option.description, String(locale.value)) || undefined,
      })),
    }
  })

  const answers = computed<WheelsetSelectionAnswers>(() => {
    const result: WheelsetSelectionAnswers = {}
    for (const entry of path.value) {
      Object.assign(result, entry.answerEffects)
    }
    return result
  })

  const answerLabels = computed<Record<string, string>>(() => {
    const result: Record<string, string> = {}
    for (const entry of path.value) {
      for (const key of Object.keys(entry.answerEffects)) {
        result[key] = entry.label
      }
    }
    return result
  })

  const specFilters = computed(() => {
    const result: Record<string, string[]> = {}
    for (const entry of path.value) {
      mergeQueryEffects(result, entry.queryEffects)
    }
    return result
  })

  const keyword = computed(() => {
    const values = path.value
      .map(entry => String(entry.queryEffects.keyword || '').trim())
      .filter(Boolean)
    return Array.from(new Set(values)).join(' ')
  })

  const productQuery = computed<WheelsetSelectionProductQuery>(() => ({
    category_slug: assistantConfig.value?.base_product_query?.category_slug
      || flow.value?.product_category_slug
      || WHEELSET_PRODUCT_CATEGORY_SLUG,
    keyword: keyword.value || undefined,
    spec_filters: cloneFilters(specFilters.value),
  }))

  const selectedPath = computed<WheelsetSelectionPathEntry[]>(() => path.value.map(entry => ({
    node_key: entry.nodeKey,
    option_key: entry.optionKey,
    label: entry.label,
    next_node_key: entry.nextNodeKey,
  })))

  const isComplete = computed(() => (
    Boolean(currentNode.value && currentNode.value.type !== 'question')
    || Boolean(currentNodeKey.value && !currentNode.value)
  ))
  const canGoBack = computed(() => path.value.length > 0)

  const rebuildFromPath = () => {
    const entryNodeKey = assistantConfig.value?.entry_node_key || ''
    const lastEntry = path.value[path.value.length - 1]
    currentNodeKey.value = lastEntry?.nextNodeKey || entryNodeKey
  }

  const reset = () => {
    path.value = []
    currentNodeKey.value = assistantConfig.value?.entry_node_key || ''
  }

  const selectOption = (optionKey: string) => {
    const node = currentNode.value
    const option = node?.options?.find(item => item.key === optionKey)
    if (!node || node.type !== 'question' || !option) return

    const effects = option.query_effects || {}
    path.value.push({
      nodeKey: node.key,
      optionKey: option.key,
      label: localizedText(option.label, String(locale.value)) || option.key,
      nextNodeKey: String(option.next_node_key || ''),
      answerEffects: { ...(option.answer_effects || {}) },
      queryEffects: {
        keyword: effects.keyword,
        spec_filters: cloneFilters(effects.spec_filters || {}),
      },
    })
    rebuildFromPath()
  }

  const goBack = () => {
    if (!path.value.length) return
    path.value = path.value.slice(0, -1)
    rebuildFromPath()
  }

  const jumpToPathIndex = (targetIndex: number) => {
    const normalizedTargetIndex = Math.max(
      0,
      Math.min(Number.isFinite(targetIndex) ? Math.trunc(targetIndex) : 0, path.value.length),
    )

    if (normalizedTargetIndex === path.value.length) {
      return
    }

    path.value = path.value.slice(0, normalizedTargetIndex)
    rebuildFromPath()
  }

  const load = async () => {
    loading.value = true
    error.value = null
    try {
      const endpoint = flowSlug === WHEELSET_SELECTION_ASSISTANT_SLUG
        ? `${publicBaseURL}/wheelset-fit-questionnaire/current`
        : `${publicBaseURL}/selection-assistant/flows/${encodeURIComponent(flowSlug)}`
      const response = await $fetch<{ data: WheelsetSelectionAssistantFlow }>(
        endpoint,
      )
      const nextFlow = response?.data
      if (!nextFlow?.version?.config?.nodes?.length) {
        throw new Error('Published selection assistant is empty.')
      }
      flow.value = nextFlow
      reset()
    } catch (cause: any) {
      flow.value = null
      currentNodeKey.value = ''
      error.value = cause?.data?.message || cause?.message || 'Unable to load the wheelset selection assistant.'
    } finally {
      loading.value = false
    }
  }

  watch(() => flowSlug, () => {
    void load()
  }, { immediate: true })

  return {
    flow,
    config: assistantConfig,
    loading,
    error,
    currentNode,
    currentQuestion,
    currentNodeKey,
    answers,
    answerLabels,
    path: selectedPath,
    productQuery,
    isComplete,
    canGoBack,
    selectOption,
    goBack,
    jumpToPathIndex,
    reset,
    reload: load,
  }
}

export type WheelsetSelectionAssistantState = ReturnType<typeof useWheelsetSelectionAssistant>

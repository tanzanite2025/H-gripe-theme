<template>
  <WheelsetSelectionAssistantTwoColumnLayout>
    <template #question>
      <div v-if="assistant.loading.value" class="wheelset-selection-assistant-flow__state">
        <Icon name="lucide:loader-circle" class="h-5 w-5 animate-spin" />
        <span>{{ t('wheelsetSelectionAssistant.states.loading') }}</span>
      </div>
      <div v-else-if="assistant.error.value" class="wheelset-selection-assistant-flow__state wheelset-selection-assistant-flow__state--error">
        <p>{{ assistant.error.value }}</p>
        <button type="button" class="wheelset-selection-assistant-flow__retry" @click="assistant.reload">
          {{ t('wheelsetSelectionAssistant.states.retry') }}
        </button>
      </div>
      <WheelsetSelectionAssistantOutcomePanel
        v-else-if="assistant.isComplete.value && assistant.currentNode.value"
        :prompt="localizedNodePrompt"
        :helper="localizedNodeHelper"
        :can-go-back="assistant.canGoBack.value"
        @back="assistant.goBack"
        @contact-support="handleContactSupport"
      />
      <WheelsetSelectionQuestionPanel
        v-else-if="assistant.currentQuestion.value"
        :question="assistant.currentQuestion.value"
        :can-go-back="assistant.canGoBack.value"
        @select="assistant.selectOption"
        @back="assistant.goBack"
      >
        <template #after-options>
          <WheelsetSelectionSupportCta @contact-support="handleContactSupport" />
        </template>
      </WheelsetSelectionQuestionPanel>
      <div v-else class="wheelset-selection-assistant-flow__state">
        <span>{{ t('wheelsetSelectionAssistant.states.empty') }}</span>
      </div>
    </template>

    <template #results>
      <WheelsetSelectionProductResultsPanel
        :category-slug="assistant.productQuery.value.category_slug"
        :selected-label="selectedAnswerLabel"
        :products="products.products.value"
        :loading="products.loading.value"
        :error="products.error.value"
        :page="products.page.value"
        :has-more="products.hasMore.value"
        @previous-page="products.setPage(products.page.value - 1)"
        @next-page="products.setPage(products.page.value + 1)"
      />
    </template>
  </WheelsetSelectionAssistantTwoColumnLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watchEffect } from 'vue'
import { useI18n } from '#imports'
import WheelsetSelectionAssistantTwoColumnLayout from '~/components/wheelset-selection/WheelsetSelectionAssistantTwoColumnLayout.vue'
import WheelsetSelectionAssistantOutcomePanel from '~/components/wheelset-selection/WheelsetSelectionAssistantOutcomePanel.vue'
import WheelsetSelectionProductResultsPanel from '~/components/wheelset-selection/WheelsetSelectionProductResultsPanel.vue'
import WheelsetSelectionQuestionPanel from '~/components/wheelset-selection/WheelsetSelectionQuestionPanel.vue'
import WheelsetSelectionSupportCta from '~/components/wheelset-selection/WheelsetSelectionSupportCta.vue'
import { useWheelsetSelectionAssistantQuestionPagination } from '~/composables/useWheelsetSelectionAssistantQuestionPagination'
import {
  type WheelsetSelectionAssistantSource,
  type WheelsetSelectionRequestDraft,
} from '~/types/wheelsetSelectionAssistant'
import { useWheelsetSelectionAssistant } from '~/composables/useWheelsetSelectionAssistant'
import { useWheelsetSelectionProducts } from '~/composables/useWheelsetSelectionProducts'
import { buildWheelsetSelectionRequestDraft } from '~/utils/wheelsetSelection/payload'
import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'

const props = withDefaults(defineProps<{
  source?: WheelsetSelectionAssistantSource
}>(), {
  source: 'guides/wheelset-buyers',
})
const emit = defineEmits<{
  contactSupport: [draft: WheelsetSelectionRequestDraft]
}>()

const { t, locale } = useI18n()
const assistant = useWheelsetSelectionAssistant()
const products = useWheelsetSelectionProducts(assistant.productQuery)
const questionPagination = useWheelsetSelectionAssistantQuestionPagination()

const selectedAnswerLabel = computed(() => (
  assistant.path.value.map(entry => entry.label).join(' / ')
))
const localizedGraphText = (value?: Record<string, string>) => {
  if (!value) return ''

  const normalizedLocale = normalizeStorefrontLocaleCode(locale.value) || 'en'
  const baseLocale = normalizedLocale.split('_')[0] || ''

  return String(value[normalizedLocale] || value[baseLocale] || value.en || '')
}

const localizedNodePrompt = computed(() => (
  localizedGraphText(assistant.currentNode.value?.prompt)
    || t('wheelsetSelectionAssistant.outcome.fallbackTitle')
))
const localizedNodeHelper = computed(() => localizedGraphText(assistant.currentNode.value?.helper))

const orderedQuestionNodeKeys = computed(() => (
  (assistant.config.value?.nodes || [])
    .filter(node => node.type === 'question')
    .sort((left, right) => {
      const leftX = Number(left.editor?.x ?? 0)
      const rightX = Number(right.editor?.x ?? 0)
      if (leftX !== rightX) return leftX - rightX
      return String(left.key).localeCompare(String(right.key))
    })
    .map(node => node.key)
))

const activeQuestionIndex = computed(() => {
  const currentKey = assistant.currentQuestion.value?.key || assistant.currentNode.value?.key || ''
  const directIndex = orderedQuestionNodeKeys.value.findIndex(key => key === currentKey)
  if (directIndex >= 0) return directIndex
  return Math.max(0, Math.min(assistant.path.value.length, Math.max(orderedQuestionNodeKeys.value.length - 1, 0)))
})

const reachableQuestionIndex = computed(() => (
  Math.max(
    0,
    Math.min(
      Math.max(activeQuestionIndex.value, assistant.path.value.length),
      Math.max(orderedQuestionNodeKeys.value.length - 1, 0),
    ),
  )
))

const jumpToQuestionIndex = (questionIndex: number) => {
  const normalizedQuestionIndex = Math.max(
    0,
    Math.min(Number.isFinite(questionIndex) ? Math.trunc(questionIndex) : 0, orderedQuestionNodeKeys.value.length - 1),
  )
  const targetQuestionKey = orderedQuestionNodeKeys.value[normalizedQuestionIndex]
  if (!targetQuestionKey) return

  if (normalizedQuestionIndex <= assistant.path.value.length) {
    assistant.jumpToPathIndex(normalizedQuestionIndex)
    return
  }

  assistant.currentNodeKey.value = targetQuestionKey
}

watchEffect(() => {
  if (!questionPagination) return

  questionPagination.total.value = orderedQuestionNodeKeys.value.length
  questionPagination.activeIndex.value = activeQuestionIndex.value
  questionPagination.reachableIndex.value = reachableQuestionIndex.value
  questionPagination.registerJumpToIndexHandler(jumpToQuestionIndex)
})

onBeforeUnmount(() => {
  if (!questionPagination) return

  questionPagination.total.value = 0
  questionPagination.activeIndex.value = 0
  questionPagination.reachableIndex.value = 0
  questionPagination.registerJumpToIndexHandler(null)
})

const handleContactSupport = () => {
  emit('contactSupport', buildWheelsetSelectionRequestDraft({
    source: props.source,
    flow: assistant.flow.value,
    answers: assistant.answers.value,
    answerLabels: assistant.answerLabels.value,
    selectedPath: assistant.path.value,
    productQuery: assistant.productQuery.value,
    completionStatus: assistant.isComplete.value ? 'complete' : 'partial',
  }))
}
</script>

<style scoped>
.wheelset-selection-assistant-flow__state {
  display: grid;
  min-height: 100%;
  place-items: center;
  gap: 0.65rem;
  padding: 2rem;
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.875rem;
  text-align: center;
}

.wheelset-selection-assistant-flow__state--error {
  align-content: center;
  color: #fca5a5;
}

.wheelset-selection-assistant-flow__state p {
  max-width: 28rem;
}

.wheelset-selection-assistant-flow__retry {
  min-height: 2.25rem;
  border-radius: 0.65rem;
  background: var(--tz-brand-primary, #b5ff6d);
  padding: 0 0.9rem;
  color: #101014;
  font-size: 0.8rem;
  font-weight: 800;
}
</style>

<template>
  <WheelsetSelectionAssistantTwoColumnLayout>
    <template #question>
      <nav
        v-if="hasQuestionPagination"
        class="wheelset-selection-assistant-flow__question-pagination"
      >
        <div
          class="wheelset-selection-assistant-flow__question-pagination-dots tz-carousel-pagination"
          :aria-label="t('wheelsetSelectionAssistant.questionPagination.label', 'Question cards')"
        >
          <button
            v-for="questionNumber in questionPagination?.total.value || 0"
            :key="questionNumber"
            type="button"
            class="tz-carousel-pagination__dot wheelset-selection-assistant-flow__question-pagination-dot"
            :class="{
              'is-active': questionNumber - 1 === questionPagination?.activeIndex.value,
              'is-future': questionNumber - 1 > (questionPagination?.reachableIndex.value ?? 0),
            }"
            :aria-current="questionNumber - 1 === questionPagination?.activeIndex.value ? 'step' : undefined"
            :aria-label="t('wheelsetSelectionAssistant.questionPagination.goToQuestion', `Show question ${questionNumber}`)"
            :title="t('wheelsetSelectionAssistant.questionPagination.goToQuestion', `Show question ${questionNumber}`)"
            :disabled="questionNumber - 1 > (questionPagination?.reachableIndex.value ?? 0)"
            @click="handleQuestionPaginationClick(questionNumber - 1)"
          />
        </div>
      </nav>

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
      <Transition
        v-else
        name="wheelset-selection-assistant-question"
        mode="out-in"
        @after-enter="handleQuestionTransitionFinished"
      >
        <WheelsetSelectionAssistantOutcomePanel
          v-if="assistant.isComplete.value && assistant.currentNode.value"
          key="outcome"
          :prompt="localizedNodePrompt"
          :helper="localizedNodeHelper"
          :can-go-back="assistant.canGoBack.value"
          @back="assistant.goBack"
          @contact-support="handleContactSupport"
        />
        <WheelsetSelectionQuestionPanel
          v-else-if="assistant.currentQuestion.value"
          :key="assistant.currentQuestion.value.key"
          :question="assistant.currentQuestion.value"
          :selected-value="selectedQuestionOptionKey"
          :can-go-back="assistant.canGoBack.value"
          :is-selecting="isQuestionTransitioning"
          @select="handleQuestionSelect"
          @back="assistant.goBack"
        />
        <div v-else key="empty" class="wheelset-selection-assistant-flow__state">
          <span>{{ t('wheelsetSelectionAssistant.states.empty') }}</span>
        </div>
      </Transition>
    </template>

    <template #results>
      <WheelsetSelectionProductResultsPanel
        :category-slug="assistant.productQuery.value.category_slug"
        :selected-label="selectedAnswerLabel"
        :products="products.products.value"
        :total="products.total.value"
        :loading="products.loading.value"
        :error="products.error.value"
        :page="products.page.value"
        :has-more="products.hasMore.value"
        @previous-page="products.setPage(products.page.value - 1)"
        @next-page="products.setPage(products.page.value + 1)"
        @retry="products.reload"
      />
    </template>
  </WheelsetSelectionAssistantTwoColumnLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watchEffect } from 'vue'
import { useI18n } from '#imports'
import WheelsetSelectionAssistantTwoColumnLayout from '~/components/wheelset-selection/WheelsetSelectionAssistantTwoColumnLayout.vue'
import WheelsetSelectionAssistantOutcomePanel from '~/components/wheelset-selection/WheelsetSelectionAssistantOutcomePanel.vue'
import WheelsetSelectionProductResultsPanel from '~/components/wheelset-selection/WheelsetSelectionProductResultsPanel.vue'
import WheelsetSelectionQuestionPanel from '~/components/wheelset-selection/WheelsetSelectionQuestionPanel.vue'
import { useWheelsetSelectionAssistantHelp } from '~/composables/useWheelsetSelectionAssistantHelp'
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
const products = useWheelsetSelectionProducts(
  assistant.productQuery,
  computed(() => Boolean(assistant.config.value)),
)
const assistantHelp = useWheelsetSelectionAssistantHelp()
const questionPagination = useWheelsetSelectionAssistantQuestionPagination()
const selectedQuestionOptionKey = ref<string | undefined>()
const isQuestionTransitioning = ref(false)
let pendingQuestionSelectionTimer: ReturnType<typeof setTimeout> | null = null
const hasQuestionPagination = computed(() => Boolean(
  questionPagination && questionPagination.total.value > 1,
))

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

const handleQuestionSelect = (optionKey: string) => {
  if (isQuestionTransitioning.value) return

  selectedQuestionOptionKey.value = optionKey
  isQuestionTransitioning.value = true
  pendingQuestionSelectionTimer = setTimeout(() => {
    pendingQuestionSelectionTimer = null
    assistant.selectOption(optionKey)
    selectedQuestionOptionKey.value = undefined
  }, 260)
}

const handleQuestionTransitionFinished = () => {
  isQuestionTransitioning.value = false
}

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

  if (normalizedQuestionIndex > reachableQuestionIndex.value) return
  assistant.jumpToPathIndex(normalizedQuestionIndex)
}

const handleQuestionPaginationClick = (questionIndex: number) => {
  if (!questionPagination || questionIndex > questionPagination.reachableIndex.value) return
  jumpToQuestionIndex(questionIndex)
}

watchEffect(() => {
  if (assistantHelp) {
    const question = assistant.currentQuestion.value
    assistantHelp.setHelp({
      title: question?.helpTitle || t('quickBuy.help.title', 'Help'),
      content: question?.helpBody || '',
    })
  }

  if (!questionPagination) return

  questionPagination.total.value = orderedQuestionNodeKeys.value.length
  questionPagination.activeIndex.value = activeQuestionIndex.value
  questionPagination.reachableIndex.value = reachableQuestionIndex.value
})

onBeforeUnmount(() => {
  if (pendingQuestionSelectionTimer) {
    clearTimeout(pendingQuestionSelectionTimer)
    pendingQuestionSelectionTimer = null
  }

  assistantHelp?.clearHelp()

  if (!questionPagination) return

  questionPagination.total.value = 0
  questionPagination.activeIndex.value = 0
  questionPagination.reachableIndex.value = 0
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

.wheelset-selection-assistant-flow__question-pagination {
  display: flex;
  flex: 0 0 auto;
  justify-content: center;
  padding: 0.1rem 0 0;
}

.wheelset-selection-assistant-flow__question-pagination-dots {
  display: inline-flex;
  justify-content: center;
  gap: 0.5rem;
}

.wheelset-selection-assistant-flow__question-pagination-dot {
  width: 2rem;
  height: 2rem;
  min-width: 2rem;
  min-height: 2rem;
  --tz-carousel-pagination-dot-width: 0.5rem;
  --tz-carousel-pagination-dot-height: 0.5rem;
}

.wheelset-selection-assistant-flow__question-pagination-dot.is-future {
  opacity: 0.7;
}

.wheelset-selection-assistant-flow__question-pagination-dot.is-future::before {
  background: var(--tz-border-strong);
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
  background: var(--tz-site-accent, #059669);
  padding: 0 0.9rem;
  color: #ffffff;
  font-size: 0.8rem;
  font-weight: 800;
}

.wheelset-selection-assistant-question-enter-active,
.wheelset-selection-assistant-question-leave-active {
  transition:
    opacity 260ms ease,
    transform 260ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: opacity, transform;
}

.wheelset-selection-assistant-question-enter-from {
  opacity: 0;
  transform: translateX(1.5rem);
}

.wheelset-selection-assistant-question-leave-to {
  opacity: 0;
  transform: translateX(-1.5rem);
}

@media (prefers-reduced-motion: reduce) {
  .wheelset-selection-assistant-question-enter-active,
  .wheelset-selection-assistant-question-leave-active {
    transition: opacity 120ms ease;
  }

  .wheelset-selection-assistant-question-enter-from,
  .wheelset-selection-assistant-question-leave-to {
    transform: none;
  }
}
</style>

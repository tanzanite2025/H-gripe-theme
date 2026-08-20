<template>
  <section id="home-buying-path" class="bg-transparent py-8 text-white sm:py-10 lg:py-12">
      <div class="page-content-shell px-0 md:px-6">
        <div class="flex flex-col gap-4 sm:gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl">
            <span
              class="inline-flex items-center gap-2 rounded-full border border-[#B5FF6D]/30 bg-[#B5FF6D]/10 px-3 py-1 text-[11px] font-medium uppercase tracking-[0.16em] text-[#B5FF6D]"
            >
              <Icon name="lucide:route" class="h-3.5 w-3.5" aria-hidden="true" />
              {{ section.eyebrow }}
            </span>
            <h2 class="mt-3 text-2xl font-semibold leading-tight text-white sm:text-3xl">
              {{ section.title }}
            </h2>
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="card in section.cards"
            :key="card.id"
            class="group flex min-h-[218px] flex-col justify-between rounded-2xl premium-card p-5 transition-transform duration-200 hover:-translate-y-1"
          >
            <div>
              <div class="flex items-center gap-3">
                <div
                  class="grid h-11 w-11 shrink-0 place-items-center text-[#B5FF6D]"
                  aria-hidden="true"
                >
                  <Icon :name="card.icon" class="h-6 w-6" />
                </div>
                <h3 class="min-w-0 text-lg font-semibold leading-tight text-white">{{ card.title }}</h3>
              </div>
              <p class="mt-4 text-sm leading-relaxed tz-text-secondary">
                {{ card.description }}
              </p>
              <div
                v-if="card.highlights?.length"
                class="mt-4 flex flex-wrap gap-2"
                :aria-label="`${card.title} topics`"
              >
                <span
                  v-for="highlight in card.highlights"
                  :key="highlight"
                  class="rounded-full border border-white/10 bg-white/[0.06] px-2.5 py-1 text-[11px] font-semibold uppercase leading-none tracking-[0.12em] text-white/90"
                >
                  {{ highlight }}
                </span>
              </div>
            </div>

            <div class="mt-6">
              <NuxtLink
                v-if="card.kind === 'route'"
                :to="localePath(card.route || '/')"
                class="premium-button w-full justify-center"
              >
                <Icon name="lucide:arrow-right" class="mr-2 h-4 w-4" aria-hidden="true" />
                {{ card.actionLabel }}
              </NuxtLink>

              <div
                v-else-if="card.kind === 'quickbuy-options'"
                class="grid gap-2"
              >
                <button
                  v-for="action in card.quickBuyActions"
                  :key="action.id"
                  type="button"
                  class="premium-button w-full justify-center"
                  @click="handleQuickBuyAction(action.id)"
                >
                  <Icon :name="action.icon" class="mr-2 h-4 w-4" aria-hidden="true" />
                  {{ action.label }}
                </button>
              </div>

              <button
                v-else
                type="button"
                class="premium-button w-full justify-center"
                @click="handleCardAction(card)"
              >
                <Icon name="lucide:message-circle" class="mr-2 h-4 w-4" aria-hidden="true" />
                {{ card.actionLabel }}
              </button>
            </div>
          </article>
        </div>
      </div>
  </section>

    <QuickBuyContactServiceModal
      v-if="contactServiceOpen"
      @close="contactServiceOpen = false"
    />

    <QuickBuyModal
      v-if="quickBuyDirectSelectOpen"
      :config="quickBuyConfig"
      @close="quickBuyDirectSelectOpen = false"
    />

  <WheelsetSelectionAssistantModal
    v-if="quickBuyWheelsetSelectionAssistantOpen"
    :model-value="true"
    source="quick-buy/wheelset-selection-assistant"
    description=""
    @update:model-value="handleQuickBuyAssistantModelUpdate"
    @close="quickBuyWheelsetSelectionAssistantOpen = false"
  >
    <WheelsetSelectionAssistantFlow
      source="quick-buy/wheelset-selection-assistant"
      @contact-support="openWheelsetSelectionSupportChat"
    />
  </WheelsetSelectionAssistantModal>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useLocalePath } from '#imports'
import QuickBuyModal from '~/components/QuickBuy.vue'
import QuickBuyContactServiceModal from '~/components/quick-buy/QuickBuyContactServiceModal.vue'
import WheelsetSelectionAssistantModal from '~/components/WheelsetSelectionAssistantModal.vue'
import WheelsetSelectionAssistantFlow from '~/components/wheelset-selection/WheelsetSelectionAssistantFlow.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useQuickBuyFlow } from '~/composables/useQuickBuyFlow'
import {
  homePurchasePathSection,
  type HomePurchasePathCard,
  type HomePurchasePathQuickBuyAction,
} from '~/utils/homePurchasePath'
import type { WheelsetSelectionRequestDraft } from '~/types/wheelsetSelectionAssistant'

const localePath = useLocalePath()
const contactServiceOpen = ref(false)
const quickBuyDirectSelectOpen = ref(false)
const quickBuyWheelsetSelectionAssistantOpen = ref(false)
const section = homePurchasePathSection
const { quickBuyFlowConfig } = useQuickBuyFlow('dock')
const quickBuyConfig = computed(() => quickBuyFlowConfig.value)
const { openChat } = useChatWidget()

const handleCardAction = (card: HomePurchasePathCard) => {
  if (card.kind === 'contact-service') {
    quickBuyDirectSelectOpen.value = false
    quickBuyWheelsetSelectionAssistantOpen.value = false
    contactServiceOpen.value = true
  }
}

const handleQuickBuyAction = (actionId: HomePurchasePathQuickBuyAction['id']) => {
  contactServiceOpen.value = false
  quickBuyDirectSelectOpen.value = actionId === 'direct-select'
  quickBuyWheelsetSelectionAssistantOpen.value = actionId === 'wheelset-selection-assistant'
}

const handleQuickBuyAssistantModelUpdate = (value: boolean) => {
  if (!value) {
    quickBuyWheelsetSelectionAssistantOpen.value = false
  }
}

const openWheelsetSelectionSupportChat = async (draft?: WheelsetSelectionRequestDraft) => {
  quickBuyWheelsetSelectionAssistantOpen.value = false
  openChat({
    showAgentList: true,
    source: 'wheelset-selection-assistant',
    pendingSelectionRequest: draft || null,
  })
  await nextTick()
}
</script>

<style scoped>
#home-buying-path {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}
</style>

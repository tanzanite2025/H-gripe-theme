<template>
  <div class="grid gap-2">
    <button
      v-for="action in actions"
      :key="action.id"
      type="button"
      class="premium-button w-full justify-center"
      @click="handleQuickBuyAction(action.id)"
    >
      <Icon :name="action.icon" class="mr-2 h-4 w-4" aria-hidden="true" />
      {{ action.label }}
    </button>
  </div>

  <LazyQuickBuyContactServiceModal
    v-if="contactServiceOpen"
    @close="contactServiceOpen = false"
  />

  <LazyQuickBuy
    v-if="quickBuyDirectSelectOpen"
    :config="quickBuyConfig"
    @close="quickBuyDirectSelectOpen = false"
  />

  <LazyWheelsetSelectionAssistantModal
    v-if="quickBuyWheelsetSelectionAssistantOpen"
    :model-value="true"
    source="quick-buy/wheelset-selection-assistant"
    description=""
    @update:model-value="handleQuickBuyAssistantModelUpdate"
    @close="quickBuyWheelsetSelectionAssistantOpen = false"
  >
    <LazyWheelsetSelectionAssistantFlow
      source="quick-buy/wheelset-selection-assistant"
      @contact-support="openWheelsetSelectionSupportChat"
    />
    <template #footer>
      <WheelsetSelectionSupportCta @contact-support="openWheelsetSelectionSupportChat" />
    </template>
  </LazyWheelsetSelectionAssistantModal>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useQuickBuyFlow } from '~/composables/useQuickBuyFlow'
import WheelsetSelectionSupportCta from '~/components/wheelset-selection/WheelsetSelectionSupportCta.vue'
import type { HomePurchasePathQuickBuyAction } from '~/utils/homePurchasePath'
import type { WheelsetSelectionRequestDraft } from '~/types/wheelsetSelectionAssistant'

defineProps<{
  actions: HomePurchasePathQuickBuyAction[]
}>()

const contactServiceOpen = ref(false)
const quickBuyDirectSelectOpen = ref(false)
const quickBuyWheelsetSelectionAssistantOpen = ref(false)
const { quickBuyFlowConfig, refresh: refreshQuickBuyFlow } = useQuickBuyFlow('dock', { immediate: false })
const quickBuyConfig = computed(() => quickBuyFlowConfig.value)
const { openChat } = useChatWidget()
let quickBuyFlowWarmup: Promise<void> | null = null

const warmQuickBuyFlow = async () => {
  if (quickBuyConfig.value) return
  if (!quickBuyFlowWarmup) {
    quickBuyFlowWarmup = refreshQuickBuyFlow().finally(() => {
      quickBuyFlowWarmup = null
    })
  }
  await quickBuyFlowWarmup
}

const handleQuickBuyAction = async (actionId: HomePurchasePathQuickBuyAction['id']) => {
  contactServiceOpen.value = false
  quickBuyDirectSelectOpen.value = false
  quickBuyWheelsetSelectionAssistantOpen.value = false

  if (actionId === 'direct-select') {
    await warmQuickBuyFlow()
    quickBuyDirectSelectOpen.value = true
    return
  }

  if (actionId === 'wheelset-selection-assistant') {
    await warmQuickBuyFlow()
    quickBuyWheelsetSelectionAssistantOpen.value = true
  }
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

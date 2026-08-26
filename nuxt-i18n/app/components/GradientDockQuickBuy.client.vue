<template>
  <button
    ref="quickBuyAnchorRef"
    class="dock-icon-button dock-quick-buy-button h-11 md:h-12 tz-text-secondary hover:text-[#059669] transition-colors"
    :class="{ 'dock-quick-buy-button--active': quickActive }"
    @click="openQuick()"
    aria-haspopup="dialog"
    :aria-expanded="quickActive"
    :aria-label="$t('dockMenu.quickBuy')"
  >
    <span class="dock-icon-slot dock-quick-buy-frame">
      <svg
        class="dock-quick-buy-icon w-full h-full transition-all"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <title>Honeybadger</title>
        <path
          d="M11.999 0c-.346 0-.691.131-.955.395L.394 11.045a1.35 1.35 0 0 0 0 1.91l6.243 6.24.915-1.95L2.306 12l9.693-9.693 1.158 1.157 1.432-1.432L12.954.395A1.346 1.346 0 0 0 11.999 0Zm5.54 1.106a.331.331 0 0 0-.218.102l-1.777 1.778-1.432 1.432-8.393 8.392h4.726l-3.76 9.26c-.139.34.29.626.55.366l1.321-1.32v-.001l1.432-1.432h.001l8.56-8.561h-4.727l2.083-4.91v.001l.854-2.012 1.112-2.623c.108-.256-.108-.485-.333-.472Zm.25 4.125-.853 2.012 4.756 4.756L12 21.693l-1.056-1.055-1.432 1.432 1.533 1.534a1.35 1.35 0 0 0 1.91 0l10.65-10.65a1.35 1.35 0 0 0 0-1.91z"
          fill="currentColor"
        />
      </svg>
    </span>
  </button>

  <LazyQuickBuyEntryRouterPopover
    v-if="quickOpen"
    :config="quickBuyConfig"
    :anchor="quickBuyAnchorRef"
    @close="closeQuickEntry"
    @direct-select="openQuickDirectSelect"
    @contact-service="openQuickContactService"
    @wheelset-selection-assistant="openQuickWheelsetSelectionAssistant"
  />

  <LazyQuickBuy
    v-if="quickDirectSelectOpen"
    :config="quickBuyConfig"
    @close="closeQuickDirectSelect"
  />

  <LazyQuickBuyContactServiceModal
    v-if="quickContactServiceOpen"
    @close="closeQuickContactService"
  />

  <LazyWheelsetSelectionAssistantModal
    v-if="quickWheelsetSelectionAssistantOpen"
    :model-value="true"
    source="quick-buy/wheelset-selection-assistant"
    description=""
    @update:model-value="handleQuickWheelsetSelectionAssistantModelUpdate"
    @close="closeQuickWheelsetSelectionAssistant"
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
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from '#imports'
import { useChatWidget } from '~/composables/useChatWidget'
import { useQuickBuyFlow } from '~/composables/useQuickBuyFlow'
import WheelsetSelectionSupportCta from '~/components/wheelset-selection/WheelsetSelectionSupportCta.vue'
import type { WheelsetSelectionRequestDraft } from '~/types/wheelsetSelectionAssistant'

const quickOpen = ref(false)
const quickDirectSelectOpen = ref(false)
const quickContactServiceOpen = ref(false)
const quickWheelsetSelectionAssistantOpen = ref(false)
const quickBuyAnchorRef = ref<HTMLElement | null>(null)
const { openChat } = useChatWidget()
const { quickBuyFlowConfig } = useQuickBuyFlow('dock')
const quickBuyConfig = computed(() => quickBuyFlowConfig.value)
const { t: $t } = useI18n()
const emit = defineEmits<{
  open: []
}>()

const quickActive = computed(() =>
  quickOpen.value || quickDirectSelectOpen.value || quickContactServiceOpen.value || quickWheelsetSelectionAssistantOpen.value,
)

const closeAll = () => {
  quickOpen.value = false
  quickDirectSelectOpen.value = false
  quickContactServiceOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = false
}

const openQuickEntry = () => {
  emit('open')
  closeAll()
  quickOpen.value = true
}

const openQuick = () => {
  if (quickActive.value) {
    closeAll()
    return
  }

  openQuickEntry()
}

const openQuickFromGlobalEvent = () => {
  openQuickEntry()
}

const closeQuickFromGlobalEvent = () => {
  closeAll()
}

const closeQuickEntry = () => {
  quickOpen.value = false
}

const openQuickDirectSelect = () => {
  quickOpen.value = false
  quickContactServiceOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = false
  quickDirectSelectOpen.value = true
}

const closeQuickDirectSelect = () => {
  quickDirectSelectOpen.value = false
}

const openQuickContactService = () => {
  quickOpen.value = false
  quickDirectSelectOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = false
  quickContactServiceOpen.value = true
}

const closeQuickContactService = () => {
  quickContactServiceOpen.value = false
}

const openQuickWheelsetSelectionAssistant = () => {
  quickOpen.value = false
  quickDirectSelectOpen.value = false
  quickContactServiceOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = true
}

const closeQuickWheelsetSelectionAssistant = () => {
  quickWheelsetSelectionAssistantOpen.value = false
}

const handleQuickWheelsetSelectionAssistantModelUpdate = (value: boolean) => {
  if (!value) {
    closeQuickWheelsetSelectionAssistant()
  }
}

const openWheelsetSelectionSupportChat = async (draft?: WheelsetSelectionRequestDraft) => {
  closeQuickWheelsetSelectionAssistant()
  openChat({
    showAgentList: true,
    source: 'wheelset-selection-assistant',
    pendingSelectionRequest: draft || null,
  })
  await nextTick()
}

onMounted(() => {
  window.addEventListener('quickbuy:open-entry', openQuickFromGlobalEvent)
  window.addEventListener('quickbuy:close-all', closeQuickFromGlobalEvent)
})

onBeforeUnmount(() => {
  window.removeEventListener('quickbuy:open-entry', openQuickFromGlobalEvent)
  window.removeEventListener('quickbuy:close-all', closeQuickFromGlobalEvent)
})
</script>

<style scoped>
.dock-icon-button {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
}

.dock-icon-slot {
  display: inline-flex;
  flex: 0 0 32px;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
}

.dock-quick-buy-frame {
  border-radius: 999px;
  transition:
    background-color 180ms ease,
    box-shadow 180ms ease,
    color 180ms ease,
    transform 180ms ease;
}

.dock-quick-buy-button--active {
  color: var(--tz-site-accent, #059669);
}

.dock-quick-buy-button--active .dock-quick-buy-frame {
  background: rgba(5, 150, 105, 0.08);
  box-shadow:
    inset 0 0 0 1px color-mix(in srgb, var(--tz-site-accent, #059669) 74%, transparent),
    0 0 0 4px rgba(5, 150, 105, 0.075);
}

.dock-quick-buy-button:hover .dock-quick-buy-frame {
  transform: translateY(-0.0625rem);
}

@media (max-width: 767px) {
  .dock-icon-slot {
    flex-basis: 28px;
    width: 28px;
    height: 28px;
  }
}
</style>

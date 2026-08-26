<template>
  <Transition name="chat-attachment-hub">
    <div
      v-if="open"
      ref="hubElement"
      class="absolute bottom-[calc(100%+0.75rem)] right-0 z-30 w-[min(20rem,calc(100vw-1.5rem))] md:w-80"
      role="dialog"
      :aria-label="t('chatModal.attachments.title')"
      @click.stop
    >
      <div class="rounded-2xl border tz-border-subtle tz-surface-card p-3 shadow-[0_18px_44px_rgba(20,32,43,0.16)] backdrop-blur-xl">
        <div class="mb-2 flex items-center justify-between gap-3 px-1">
          <span class="text-sm font-semibold tz-text-primary">
            {{ t('chatModal.attachments.title') }}
          </span>
          <button
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded-full tz-text-primary/55 transition-colors hover:tz-surface-subtle hover:tz-text-primary"
            :aria-label="t('chatModal.actions.close')"
            @click="$emit('close')"
          >
            <Icon name="lucide:x" class="h-4 w-4" />
          </button>
        </div>

        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="action in actions"
            :key="action.id"
            type="button"
            class="flex min-h-20 flex-col items-center justify-center gap-2 rounded-xl border tz-border-subtle tz-surface-subtle px-2 py-3 text-center tz-text-primary transition-colors hover:border-[#059669]/60 hover:bg-[#059669]/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#059669]"
            @click="$emit('select', action.id)"
          >
            <span class="flex h-9 w-9 items-center justify-center rounded-full tz-surface-subtle tz-text-primary">
              <Icon :name="action.icon" class="h-5 w-5" />
            </span>
            <span class="text-xs font-semibold leading-4">
              {{ t(action.labelKey) }}
            </span>
          </button>
        </div>
        <p class="mt-2 px-1 text-[11px] leading-4 tz-text-muted">
          {{ uploadSpecHint('customer_service_attachment') }}
        </p>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import {
  CHAT_ATTACHMENT_ACTIONS,
  type ChatAttachmentActionId,
} from '~/composables/chat/useChatAttachmentActions'
import { uploadSpecHint } from '~/utils/uploadSpecs'

defineProps<{
  open: boolean
}>()

defineEmits<{
  close: []
  select: [action: ChatAttachmentActionId]
}>()

const { t } = useI18n()
const hubElement = ref<HTMLElement | null>(null)
const actions = computed(() => CHAT_ATTACHMENT_ACTIONS)

defineExpose({
  hubElement,
})
</script>

<style scoped>
.chat-attachment-hub-enter-active,
.chat-attachment-hub-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
  transform-origin: bottom right;
}

.chat-attachment-hub-enter-from,
.chat-attachment-hub-leave-to {
  opacity: 0;
  transform: translateY(0.5rem) scale(0.98);
}
</style>

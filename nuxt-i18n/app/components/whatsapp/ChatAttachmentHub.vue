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
      <div class="rounded-2xl border border-white/15 bg-black/95 p-3 shadow-[0_18px_44px_rgba(0,0,0,0.62)] backdrop-blur-xl">
        <div class="mb-2 flex items-center justify-between gap-3 px-1">
          <span class="text-sm font-semibold text-white">
            {{ t('chatModal.attachments.title') }}
          </span>
          <button
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded-full text-white/55 transition-colors hover:bg-white/10 hover:text-white"
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
            class="flex min-h-20 flex-col items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/[0.04] px-2 py-3 text-center text-white transition-colors hover:border-[#B5FF6D]/60 hover:bg-[#B5FF6D]/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#B5FF6D]"
            @click="$emit('select', action.id)"
          >
            <span class="flex h-9 w-9 items-center justify-center rounded-full bg-white/[0.08] text-white">
              <Icon :name="action.icon" class="h-5 w-5" />
            </span>
            <span class="text-xs font-semibold leading-4">
              {{ t(action.labelKey) }}
            </span>
          </button>
        </div>
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

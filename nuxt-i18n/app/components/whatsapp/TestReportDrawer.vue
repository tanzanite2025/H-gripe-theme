<template>
  <Teleport to="body">
    <Transition name="whatsapp-product-drawer">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[10001] flex items-end justify-center p-0 md:p-4 pointer-events-none"
      >
        <!-- Mobile Backdrop -->
        <div 
          class="absolute inset-0 bg-black/60 backdrop-blur-sm md:hidden pointer-events-auto transition-opacity"
          @click="handleClose"
        />

        <div
          class="pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]
                 rounded-2xl border-2 border-[#6b73ff]/40
                 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))]
                 backdrop-blur-md shadow-[0_0_30px_rgba(107,115,255,0.6)]
                 flex flex-col overflow-hidden"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b border-white/10 flex-shrink-0">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="text-sm font-semibold text-white/90 truncate">
                Test Report
                <span v-if="agent" class="text-xs text-white/60 ml-1">({{ agent.name }})</span>
              </div>
            </div>
            <button
              type="button"
              class="w-8 h-8 rounded-full border border-white/40 text-white flex items-center justify-center hover:bg-white/10 transition-colors"
              @click="handleClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <!-- Content -->
          <div class="flex-1 min-h-0 overflow-y-auto p-4 md:p-6 custom-scrollbar">
            <TestReportContent :sync-with-url="false" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import TestReportContent from '~/components/TestReportContent.vue'

const props = defineProps<{
  modelValue: boolean
  agent?: any | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
}>()

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}
</script>

<style scoped>
.whatsapp-product-drawer-enter-active,
.whatsapp-product-drawer-leave-active {
  transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}

.whatsapp-product-drawer-enter-from,
.whatsapp-product-drawer-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.whatsapp-product-drawer-enter-to,
.whatsapp-product-drawer-leave-from {
  transform: translateY(0%);
  opacity: 1;
}

/* 自定义滚动条样式，匹配 WhatsAppChatModal 中的风格 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>

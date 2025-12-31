<template>
  <Teleport to="body">
    <Transition name="wa-drawer">
      <div
        v-if="modelValue"
        class="wa-drawer-mask"
      >
        <!-- Mobile Backdrop -->
        <div 
          class="wa-drawer-backdrop md:hidden"
          @click="handleClose"
        />

        <div class="wa-drawer-shell">
          <!-- Header -->
          <div class="wa-drawer-header">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="wa-drawer-title">
                Test Report
                <span v-if="agent" class="text-xs text-white/60 ml-1">({{ agent.name }})</span>
              </div>
            </div>
            <button
              type="button"
              class="wa-drawer-close-btn"
              @click="handleClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <!-- Content -->
          <div class="wa-drawer-content custom-scrollbar">
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
/* 自定义滚动条样式 - 暂保留在此，或后续通过全局类处理 */
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

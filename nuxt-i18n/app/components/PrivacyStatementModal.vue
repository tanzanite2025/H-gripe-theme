<template>
  <div
    class="wa-drawer-shell"
  >
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-3 md:py-4 border-b border-white/10">
      <h2
        class="text-xl md:text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff]"
      >
        {{ $t('privacy.title', 'Privacy statement') }}
      </h2>
      <button
        type="button"
        @click="$emit('close')"
        class="w-10 h-10 rounded-full bg-white/5 hover:bg-white/10 border border-white/20 flex items-center justify-center transition-all duration-200 tz-text-secondary hover:text-white"
        aria-label="Close privacy statement"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-6 py-6">
      <PrivacyStatementContent />
    </div>
  </div>
</template>

<script setup lang="ts">
import PrivacyStatementContent from '~/components/PrivacyStatementContent.vue'

const emit = defineEmits<{ (e: 'close'): void }>()
</script>



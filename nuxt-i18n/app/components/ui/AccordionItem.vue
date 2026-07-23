<template>
  <div :id="`accordion-item-${id}`" class="accordion-item flex flex-col rounded-xl overflow-hidden border border-white/5 transition-colors duration-3000" :class="{ 'bg-[#11151e]': isActive, 'bg-slate-800/20': !isActive }">
    
    <!-- Header -->
    <button
      type="button"
      class="accordion-header w-full text-left p-3 md:p-4 flex items-center justify-between group transition-colors duration-200"
      :class="{ 'text-sky-400': isActive, 'tz-text-secondary hover:bg-white/5': !isActive }"
      @click="toggleItem(id)"
    >
      <div class="flex flex-col">
        <span class="text-base font-bold">{{ title }}</span>
        <span v-if="subtitle" class="text-xs tz-text-muted font-normal mt-0.5">{{ subtitle }}</span>
      </div>
      
      <!-- Icon Indicator -->
      <span 
        class="shrink-0 ml-4 w-8 h-8 rounded-full flex items-center justify-center transition-all duration-300"
        :class="{ 'bg-sky-500/10 rotate-180': isActive, 'bg-slate-700/50': !isActive }"
      >
        <svg 
          xmlns="http://www.w3.org/2000/svg" 
          width="16" 
          height="16" 
          viewBox="0 0 24 24" 
          fill="none" 
          stroke="currentColor" 
          stroke-width="2.5" 
          stroke-linecap="round" 
          stroke-linejoin="round"
          class="transition-colors duration-200"
          :class="{ 'text-sky-400': isActive, 'tz-text-muted': !isActive }"
        >
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </span>
    </button>

    <!-- Content -->
    <div
      v-show="isActive"
      :id="`accordion-content-${id}`"
      class="border-t border-slate-700/50 bg-slate-900/20 p-4"
    >
      <slot />
    </div>

  </div>
</template>

<script setup lang="ts">
import { inject, computed } from 'vue'

const props = defineProps<{
  id: string
  title: string
  subtitle?: string
}>()

// Inject context from parent
const accordion = inject('accordion') as {
  activeId: { value: string | null }
  toggleItem: (id: string) => void
}

const isActive = computed(() => accordion.activeId.value === props.id)
const toggleItem = accordion.toggleItem
</script>

<style scoped>
/* Optional slide animation can be added here if v-show transition is not enough */
.accordion-content {
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

<template>
  <div class="smart-accordion space-y-3">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, provide, nextTick, onMounted, watch } from 'vue'

const props = defineProps<{
  defaultId?: string
}>()

// Store the active item ID
const activeId = ref<string | null>(props.defaultId || null)

// Provide context to children
provide('accordion', {
  activeId,
  toggleItem: (id: string) => {
    // If clicking the already active item, do nothing (mutually exclusive & persistent)
    // Or allow toggle off? For this guide, "Accordion" usually keeps one open.
    // Let's allow toggling off if needed, or keep strictly one open.
    // User requested "only one open", so switching activeId is enough.
    
    // Toggle off if clicking the currently active item
    if (activeId.value === id) {
      activeId.value = null
      return
    }

    activeId.value = id
    
    // Auto-scroll logic
    nextTick(() => {
      const el = document.getElementById(`accordion-item-${id}`)
      if (el) {
        // Offset for sticky header if exists (usually ~60-80px)
        // Using scrollIntoView with block: 'start' usually puts it at the very top.
        // We might want a small margin-top.
        const y = el.getBoundingClientRect().top + window.pageYOffset - 120 // 120px offset for sticky header
        window.scrollTo({ top: y, behavior: 'smooth' })
      }
    })
  }
})

// Update local state if defaultId changes
watch(() => props.defaultId, (newId) => {
  if (newId) activeId.value = newId
})
</script>

<style scoped>
.smart-accordion {
  width: 100%;
}
</style>

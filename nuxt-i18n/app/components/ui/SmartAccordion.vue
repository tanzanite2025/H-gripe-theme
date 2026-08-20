<template>
  <div class="smart-accordion space-y-3">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, provide, watch } from 'vue'

const props = defineProps<{
  defaultId?: string
}>()

const activeId = ref<string | null>(props.defaultId || null)

provide('accordion', {
  activeId,
  toggleItem: (id: string) => {
    if (activeId.value === id) {
      activeId.value = null
      return
    }

    activeId.value = id
  }
})

watch(() => props.defaultId, (newId) => {
  if (newId) activeId.value = newId
})
</script>

<style scoped>
.smart-accordion {
  width: 100%;
}
</style>

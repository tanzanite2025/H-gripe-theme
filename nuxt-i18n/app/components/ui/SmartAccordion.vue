<template>
  <div class="smart-accordion space-y-3">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, provide, watch } from 'vue'

const props = defineProps<{
  defaultId?: string
  activeId?: string | null
}>()
const emit = defineEmits<{
  'update:activeId': [value: string | null]
}>()

const activeId = ref<string | null>(props.activeId ?? props.defaultId ?? null)

provide('accordion', {
  activeId,
  toggleItem: (id: string) => {
    if (activeId.value === id) {
      activeId.value = null
      emit('update:activeId', null)
      return
    }

    activeId.value = id
    emit('update:activeId', id)
  }
})

watch(() => props.activeId, (newId) => {
  if (newId !== undefined && newId !== activeId.value) {
    activeId.value = newId
  }
})

watch(() => props.defaultId, (newId) => {
  if (newId && props.activeId === undefined) activeId.value = newId
})
</script>

<style scoped>
.smart-accordion {
  width: 100%;
}
</style>

<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { useVModel } from '@vueuse/core'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
  defaultValue?: string | number
  modelValue?: string | number
}>()

const emits = defineEmits<{
  (e: 'update:modelValue', payload: string | number): void
}>()

const modelValue = useVModel(props, 'modelValue', emits, {
  passive: true,
  defaultValue: props.defaultValue,
})
</script>

<template>
  <textarea
    v-model="modelValue"
    data-slot="textarea"
    :class="cn('bg-muted/50 border-none focus-visible:ring-ring/50 rounded-2xl px-3 py-2 text-sm font-bold transition-all focus-visible:ring-2 flex field-sizing-content min-h-16 w-full outline-none placeholder:text-muted-foreground/60 disabled:cursor-not-allowed disabled:opacity-50', props.class)"
  />
</template>

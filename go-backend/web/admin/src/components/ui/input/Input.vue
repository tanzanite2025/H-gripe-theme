<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { useVModel } from '@vueuse/core'
import { cn } from '@/lib/utils'

const props = defineProps<{
  defaultValue?: string | number
  modelValue?: string | number
  class?: HTMLAttributes['class']
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
  <input
    v-model="modelValue"
    data-slot="input"
    :class="cn(
      'bg-muted/50 border-none focus-visible:ring-ring/50 h-9 rounded-xl px-3 py-1.5 text-sm font-bold transition-all focus-visible:ring-2 w-full min-w-0 outline-none placeholder:text-muted-foreground/60 disabled:opacity-50',
      props.class,
    )"
  >
</template>

<template>
  <select
    :value="modelValue"
    class="h-9 rounded-md border bg-background px-3 text-sm"
    :aria-label="ariaLabel"
    :disabled="disabled"
    @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
  >
    <option v-if="includeAll" value="">全部环境</option>
    <option v-for="option in options" :key="option.value" :value="option.value">
      {{ option.label }}
    </option>
  </select>
</template>

<script setup lang="ts">
interface EnvironmentOption {
  value: string
  label: string
}

withDefaults(defineProps<{
  modelValue: string
  options: readonly EnvironmentOption[]
  ariaLabel?: string
  disabled?: boolean
  includeAll?: boolean
}>(), {
  disabled: false,
  includeAll: true,
  ariaLabel: '环境筛选',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

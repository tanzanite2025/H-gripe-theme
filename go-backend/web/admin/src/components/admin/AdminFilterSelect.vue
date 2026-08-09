<template>
  <label class="block space-y-1">
    <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ label }}</span>
    <Select :model-value="modelValue" @update:model-value="handleValueChange">
      <SelectTrigger class="h-9 w-full">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem v-for="option in options" :key="option.value" :value="option.value">
          {{ option.label }}
        </SelectItem>
      </SelectContent>
    </Select>
  </label>
</template>

<script setup lang="ts">
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface AdminFilterOption {
  label: string
  value: string
}

defineProps<{
  modelValue: string
  label: string
  options: AdminFilterOption[]
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

const handleValueChange = (value: unknown): void => {
  emit('update:modelValue', String(value))
}
</script>

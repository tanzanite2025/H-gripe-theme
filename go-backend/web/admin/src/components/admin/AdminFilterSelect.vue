<template>
  <label class="block space-y-1">
    <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ label }}</span>
    <Select :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)">
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

<script setup>
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

defineProps({
  modelValue: { type: String, required: true },
  label: { type: String, required: true },
  options: { type: Array, required: true }
})

const emit = defineEmits(['update:modelValue'])
</script>

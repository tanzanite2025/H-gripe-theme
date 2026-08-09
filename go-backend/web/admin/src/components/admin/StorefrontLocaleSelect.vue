<template>
  <Select
    :model-value="modelValue"
    :disabled="disabled || loading"
    @update:model-value="emit('update:modelValue', String($event || ''))"
  >
    <SelectTrigger class="w-full">
      <LockKeyhole
        v-if="locked"
        class="size-3.5 shrink-0 text-muted-foreground"
        :title="lockedTitle"
      />
      <SelectValue :placeholder="placeholder" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="language in languageOptions" :key="language.value" :value="language.value">
        {{ language.label }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>

<script setup lang="ts">
import { LockKeyhole } from '@lucide/vue'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { LanguageOption } from '@/lib/languages'

withDefaults(defineProps<{
  modelValue?: string
  languageOptions: LanguageOption[]
  disabled?: boolean
  loading?: boolean
  locked?: boolean
  lockedTitle?: string
  placeholder?: string
}>(), {
  modelValue: '',
  disabled: false,
  loading: false,
  locked: false,
  lockedTitle: '语言已锁定',
  placeholder: '选择语言',
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()
</script>

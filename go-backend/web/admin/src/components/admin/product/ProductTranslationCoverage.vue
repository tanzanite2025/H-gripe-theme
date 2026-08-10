<template>
  <div
    v-if="group"
    class="min-w-24 space-y-1"
    :title="missingTitle"
  >
    <div class="flex items-center gap-1.5 text-xs font-bold">
      <Languages class="size-3.5 text-admin-selected" />
      <span>{{ group.translations.length }}/{{ totalLanguages }}</span>
    </div>
    <div v-if="group.missing_locales.length > 0" class="flex items-center gap-1 text-[10px] font-medium text-amber-600 dark:text-amber-400">
      <CircleAlert class="size-3" />
      缺 {{ group.missing_locales.length }} 语言
    </div>
    <div v-else class="flex items-center gap-1 text-[10px] font-medium text-emerald-600 dark:text-emerald-400">
      <CircleCheck class="size-3" />
      已覆盖
    </div>
  </div>
  <span v-else class="text-muted-foreground/50">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CircleAlert, CircleCheck, Languages } from '@lucide/vue'
import type { ProductTranslationGroup } from './productEditorTypes'

const props = withDefaults(defineProps<{
  group?: ProductTranslationGroup | null
  localeName: (locale?: string | null) => string
  totalLanguages?: number
}>(), {
  group: null,
  totalLanguages: 20
})

const missingTitle = computed(() => {
  if (!props.group || props.group.missing_locales.length === 0) return '翻译组已覆盖所有启用语言'
  return `缺少语言：${props.group.missing_locales.map((locale) => props.localeName(locale)).join('、')}`
})
</script>

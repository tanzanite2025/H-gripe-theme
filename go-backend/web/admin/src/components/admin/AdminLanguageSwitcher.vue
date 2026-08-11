<template>
  <div
    class="inline-flex items-center gap-1 rounded-full border border-border/80 bg-background/80 p-1 shadow-sm"
    :aria-label="t('language.label')"
  >
    <Languages class="ml-1.5 size-3.5 text-muted-foreground" aria-hidden="true" />
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="min-w-9 rounded-full px-2 py-1 text-[10px] font-black transition-colors"
      :class="locale === option.value ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
      :aria-label="option.value === 'zh-CN' ? t('language.switchToChinese') : t('language.switchToEnglish')"
      :aria-pressed="locale === option.value"
      @click="setLocale(option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { Languages } from '@lucide/vue'
import { useAdminI18n, type AdminLocale } from '@/i18n'

const { locale, setLocale, t } = useAdminI18n()
const options: Array<{ value: AdminLocale; label: string }> = [
  { value: 'zh-CN', label: '中' },
  { value: 'en-US', label: 'EN' },
]
</script>

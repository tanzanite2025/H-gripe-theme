<template>
  <AdminFilterPanel>
    <div class="flex flex-col gap-3 xl:flex-row xl:items-start xl:justify-between">
      <div class="min-w-0">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
          URL Search Control
        </p>
        <h2 class="mt-1 text-sm font-black">
          URL 搜索管理筛选
        </h2>
        <p class="mt-1 text-xs text-muted-foreground">
          先按语言、配置状态和关键词缩小范围，再编辑每条 URL 的搜索词和展示文案。
        </p>
      </div>

      <div class="shrink-0 text-left text-[10px] font-mono text-muted-foreground xl:text-right">
        <p>SHOWING / {{ paginationTotal }} 条</p>
        <p class="mt-1">LANG / {{ currentLocaleLabel }}</p>
        <p class="mt-1">CONFIG / {{ currentProfileStatusLabel }}</p>
      </div>
    </div>

    <form
      class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(220px,1.6fr)_repeat(4,minmax(150px,1fr))_auto]"
      @submit.prevent="emit('apply')"
    >
      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
          SEARCH / 搜索
        </span>
        <Input
          v-model="filters.search"
          autocomplete="off"
          placeholder="标题、路径、关键词或来源键"
        />
      </label>

      <AdminFilterSelect
        v-model="filters.locale"
        label="LANG / 语言"
        :options="localeOptions"
      />

      <AdminFilterSelect
        v-model="filters.source_type"
        label="SOURCE / 来源"
        :options="sourceOptions"
      />

      <AdminFilterSelect
        v-model="filters.searchable"
        label="SEARCHABLE / 搜索"
        :options="searchableOptions"
      />

      <AdminFilterSelect
        v-model="filters.search_profile_status"
        label="PROFILE / 配置"
        :options="profileStatusOptions"
      />

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
          ALIASES / 兼容路由
        </span>
        <div class="flex h-9 items-center gap-2 rounded-md border border-input bg-background px-3">
          <Switch v-model="filters.includeAliases" size="sm" />
          <span class="text-xs text-muted-foreground">显示 /faq、旧路径等兼容链接</span>
        </div>
      </label>

      <div class="flex items-end gap-2 md:col-span-2 xl:col-span-6 xl:justify-end">
        <Button
          type="submit"
          class="h-9 px-3 text-xs font-black uppercase tracking-wider"
          :disabled="loading"
        >
          <Search class="size-3.5" />
          查询
        </Button>
        <Button
          type="button"
          variant="outline"
          class="h-9 px-3 text-xs font-black uppercase tracking-wider"
          :disabled="loading"
          @click="emit('reset')"
        >
          重置
        </Button>
      </div>
    </form>
  </AdminFilterPanel>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import type { StorefrontRouteCatalogFilters } from '@/composables/url-management/useStorefrontRouteCatalog'

const props = defineProps<{
  filters: StorefrontRouteCatalogFilters
  paginationTotal: number
  localeFilterOptions: Array<{ value: string, label: string }>
  loading?: boolean
}>()

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()

const localeOptions = computed(() => [
  { value: 'all', label: '全部语言' },
  ...props.localeFilterOptions,
])

const sourceOptions = [
  { value: 'all', label: '全部来源' },
  { value: 'static', label: '静态页面' },
  { value: 'product', label: '产品' },
  { value: 'blog', label: 'Blog' },
  { value: 'alias', label: '兼容路由' },
]

const searchableOptions = [
  { value: 'all', label: '全部搜索属性' },
  { value: 'true', label: '可被全局搜索' },
  { value: 'false', label: '不进入全局搜索' },
]

const profileStatusOptions = [
  { value: 'all', label: '全部配置状态' },
  { value: 'configured', label: '已配置' },
  { value: 'unconfigured', label: '未配置' },
]

const currentLocaleLabel = computed(() => {
  const selectedLocale = props.filters.locale
  if (selectedLocale === 'all') return '全部语言'
  return localeOptions.value.find(option => option.value === selectedLocale)?.label || selectedLocale
})

const currentProfileStatusLabel = computed(() => {
  const selectedStatus = props.filters.search_profile_status
  return profileStatusOptions.find(option => option.value === selectedStatus)?.label || selectedStatus
})
</script>

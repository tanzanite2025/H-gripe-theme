<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_repeat(3,minmax(140px,0.7fr))_auto]" @submit.prevent="emit('apply')">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="商品名称、SKU 或描述" />
        </div>
      </label>

      <AdminFilterSelect v-model="filters.status" label="状态" :options="statusFilterOptions" />
      <AdminFilterSelect v-model="filters.locale" label="语言" :options="localeFilterOptions" />
      <AdminFilterSelect v-model="filters.featured" label="精选" :options="featuredFilterOptions" />

      <label class="space-y-1 block">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION / 操作</span>
        <div class="flex items-center gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="emit('reset')">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </label>
    </form>
  </AdminFilterPanel>
</template>

<script setup lang="ts">
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface ProductFilters {
  search: string
  status: string
  locale: string
  featured: string
  [key: string]: unknown
}

interface FilterOption {
  label: string
  value: string
}

const props = withDefaults(defineProps<{
  filters: ProductFilters
  statusFilterOptions?: FilterOption[]
  localeFilterOptions?: FilterOption[]
  featuredFilterOptions?: FilterOption[]
}>(), {
  statusFilterOptions: () => [],
  localeFilterOptions: () => [],
  featuredFilterOptions: () => [],
})

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()
</script>

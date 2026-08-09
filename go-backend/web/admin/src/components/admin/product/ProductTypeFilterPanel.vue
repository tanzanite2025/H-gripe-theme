<template>
  <AdminFilterPanel>
    <div class="grid gap-3 md:grid-cols-[minmax(240px,1fr)_180px_auto]">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="名称或标识" />
        </div>
      </label>

      <AdminFilterSelect v-model="filters.status" label="STATUS / 状态" :options="statusOptions" />

      <div class="flex items-end">
        <Button variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="emit('reset')">
          <RotateCcw class="size-3.5" />
          重置
        </Button>
      </div>
    </div>
  </AdminFilterPanel>
</template>

<script setup lang="ts">
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { ProductTypeFilters } from './productTypeTypes'

interface ProductTypeFilterOption {
  label: string
  value: string
}

defineProps<{
  filters: ProductTypeFilters
}>()

const emit = defineEmits<{
  (event: 'reset'): void
}>()

const statusOptions: ProductTypeFilterOption[] = [
  { label: '全部状态', value: 'all' },
  { label: '已启用', value: 'enabled' },
  { label: '已停用', value: 'disabled' },
]
</script>

<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_repeat(2,minmax(140px,0.7fr))_auto]" @submit.prevent="emit('apply')">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="主题、单号或回复" />
        </div>
      </label>

      <AdminFilterSelect :model-value="filters.status" label="状态" :options="statusFilterOptions" @update:model-value="updateFilter('status', $event)" />
      <AdminFilterSelect :model-value="filters.priority" label="优先级" :options="priorityFilterOptions" @update:model-value="updateFilter('priority', $event)" />

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
import type { LanguageOption } from '@/lib/languages'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { TicketFilters } from './ticketTypes'

type TicketFilterField = 'status' | 'priority'

const props = withDefaults(defineProps<{
  filters: TicketFilters
  statusFilterOptions?: LanguageOption[]
  priorityFilterOptions?: LanguageOption[]
}>(), {
  statusFilterOptions: () => [],
  priorityFilterOptions: () => []
})

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()

const updateFilter = (field: TicketFilterField, value: string): void => {
  props.filters[field] = value
  emit('apply')
}
</script>

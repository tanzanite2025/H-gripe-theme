<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_repeat(4,minmax(140px,0.7fr))_auto]" @submit.prevent="$emit('apply')">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="问题或回答内容" />
        </div>
      </label>

      <label v-for="filter in selectFilters" :key="filter.key" class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">{{ filter.label }}</span>
        <Select v-model="filters[filter.key]">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem v-for="option in filter.options" :key="option.value" :value="option.value">
              {{ option.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="space-y-1 block">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION / 操作</span>
        <div class="flex items-center gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="$emit('reset')">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </label>
    </form>
  </AdminFilterPanel>
</template>

<script setup>
import { computed } from 'vue'
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const props = defineProps({
  filters: { type: Object, required: true },
  pageFilterOptions: { type: Array, required: true },
  categoryFilterOptions: { type: Array, required: true },
  statusFilterOptions: { type: Array, required: true },
  localeFilterOptions: { type: Array, required: true }
})

defineEmits(['apply', 'reset'])

const selectFilters = computed(() => [
  { key: 'page_id', label: '页面', options: props.pageFilterOptions },
  { key: 'category', label: '分类', options: props.categoryFilterOptions },
  { key: 'status', label: '状态', options: props.statusFilterOptions },
  { key: 'locale', label: '语言', options: props.localeFilterOptions }
])
</script>

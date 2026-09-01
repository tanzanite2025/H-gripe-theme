<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_minmax(140px,0.7fr)_auto]" @submit.prevent="emit('apply')">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="搜索 Email" />
        </div>
      </label>

      <AdminFilterSelect v-model="filters.status" label="STATUS / 状态" :options="statusOptions" />

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
import type { SubscriptionFilters } from '@/modules/subscription/subscriptionTypes'

interface SubscriptionFilterOption {
  label: string
  value: string
}

defineProps<{
  filters: SubscriptionFilters
}>()

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()

const statusOptions: SubscriptionFilterOption[] = [
  { label: '全部状态', value: 'all' },
  { label: '已订阅', value: 'active' }
]
</script>


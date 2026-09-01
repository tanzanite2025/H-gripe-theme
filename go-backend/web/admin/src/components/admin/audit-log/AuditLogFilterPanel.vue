<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4" @submit.prevent="emit('apply')">
      <label class="space-y-1 block xl:col-span-2">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">KEYWORD / 关键词</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.keyword" class="h-9 pl-9" placeholder="用户、操作、资源、路径或错误信息" />
        </div>
      </label>
      <AdminFilterSelect v-model="filters.action" label="操作" :options="actionFilterOptions" />
      <AdminFilterSelect v-model="filters.resource" label="资源" :options="resourceFilterOptions" />
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">USER ID / 用户 ID</span>
        <Input v-model="filters.user_id" type="number" min="1" class="h-9" placeholder="全部用户" />
      </label>
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">IP ADDRESS / IP 地址</span>
        <Input v-model="filters.ip_address" class="h-9 font-mono" placeholder="全部地址" />
      </label>
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">START DATE / 开始日期</span>
        <Input v-model="filters.start_date" type="date" class="h-9" />
      </label>
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">END DATE / 结束日期</span>
        <Input v-model="filters.end_date" type="date" class="h-9" />
      </label>
      <label class="space-y-1 block xl:col-span-4">
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
import type { AuditLogFilters } from '@/modules/audit-log/auditLogTypes'

interface AuditLogFilterOption {
  label: string
  value: string
}

defineProps<{
  filters: AuditLogFilters
}>()

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()

const actionFilterOptions: AuditLogFilterOption[] = [
  { label: '全部操作', value: 'all' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '查看', value: 'view' }
]
const resourceFilterOptions: AuditLogFilterOption[] = [
  { label: '全部资源', value: 'all' },
  { label: '用户', value: 'user' },
  { label: '商品', value: 'product' },
  { label: '订单', value: 'order' },
  { label: '文章', value: 'post' },
  { label: '工单', value: 'ticket' },
  { label: 'FAQ', value: 'faq' },
  { label: '图库', value: 'gallery' },
  { label: '订阅', value: 'subscription' },
  { label: '营销', value: 'marketing' },
  { label: '设置', value: 'setting' }
]
</script>


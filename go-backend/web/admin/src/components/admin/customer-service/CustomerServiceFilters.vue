<template>
  <AdminFilterPanel>
    <form class="grid gap-3 lg:grid-cols-4 lg:gap-2 xl:grid-cols-[minmax(240px,1.4fr)_repeat(4,minmax(120px,1fr))_auto_auto] xl:gap-3" @submit.prevent="emit('apply')">
      <label class="block min-w-0 space-y-1 lg:col-span-2">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 w-full pl-9" placeholder="客户、邮箱、会话 ID、消息内容" />
        </div>
      </label>

      <label class="block space-y-1 min-w-0">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 状态</span>
        <Select v-model="filters.status">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部</SelectItem>
            <SelectItem value="pending">待处理</SelectItem>
            <SelectItem value="active">进行中</SelectItem>
            <SelectItem value="closed">已关闭</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1 min-w-0">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">IDENTITY / 身份</span>
        <Select v-model="filters.identity">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部客户</SelectItem>
            <SelectItem value="account">会员</SelectItem>
            <SelectItem value="anonymous">匿名访客</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1 min-w-0">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">UNREAD / 未读</span>
        <Select v-model="filters.unread">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部</SelectItem>
            <SelectItem value="unread">只看未读</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1 min-w-0">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">QUERY</span>
        <Button type="submit" class="h-9 w-full justify-center rounded-full px-3 font-black text-xs uppercase tracking-wider" :disabled="loading">
          <Search class="size-3.5" />
          查询
        </Button>
      </label>

      <label class="block space-y-1 min-w-0">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION</span>
        <Button type="button" variant="outline" class="h-9 w-full justify-center rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="emit('reset')">
          <RotateCcw class="size-3.5" />
          重置
        </Button>
      </label>
    </form>
  </AdminFilterPanel>
</template>

<script setup lang="ts">
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface CustomerServiceFiltersState {
  search: string
  status: string
  identity: string
  unread: string
  [key: string]: string
}

withDefaults(defineProps<{
  filters: CustomerServiceFiltersState
  loading?: boolean
}>(), {
  loading: false,
})

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()
</script>

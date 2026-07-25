<template>
  <AdminFilterPanel>
    <form class="grid gap-3 lg:grid-cols-[minmax(220px,1.2fr)_140px_140px_180px_130px_auto_auto]" @submit.prevent="emit('apply')">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="客户、邮箱、会话 ID、消息内容" />
        </div>
      </label>

      <label class="space-y-1 block">
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

      <label class="space-y-1 block">
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

      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">ASSIGNEE / 负责人</span>
        <Select v-model="filters.assignedTo">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部客服</SelectItem>
            <SelectItem v-for="agent in assignableAgents" :key="agent.user_id || agent.id" :value="String(agent.user_id || agent.id)">
              {{ agent.name || agent.email || `用户 ${agent.user_id || agent.id}` }}
            </SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">UNREAD / 未读</span>
        <Select v-model="filters.unread">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部</SelectItem>
            <SelectItem value="unread">只看未读</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="space-y-1 block">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">QUERY</span>
        <Button type="submit" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" :disabled="loading">
          <Search class="size-3.5" />
          查询
        </Button>
      </label>

      <label class="space-y-1 block">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION</span>
        <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="emit('reset')">
          <RotateCcw class="size-3.5" />
          重置
        </Button>
      </label>
    </form>
  </AdminFilterPanel>
</template>

<script setup>
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

defineProps({
  filters: { type: Object, required: true },
  assignableAgents: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['apply', 'reset'])
</script>

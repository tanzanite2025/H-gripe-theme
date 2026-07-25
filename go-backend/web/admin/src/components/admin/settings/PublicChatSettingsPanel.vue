<template>
  <div class="space-y-4">
    <Alert>
      <Info class="size-4" />
      <AlertTitle>公开客服状态</AlertTitle>
      <AlertDescription>公开客服需绑定活跃用户，且角色为 admin、manager 或 support。</AlertDescription>
    </Alert>

    <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      <p class="text-xs text-muted-foreground">从已有后台用户中选择客服账号，生成或更新 Public Chat 对外 Profile。</p>
      <div class="flex flex-wrap justify-end gap-2">
        <Button
          v-if="canEdit"
          size="sm"
          :disabled="loadingCandidates"
          @click="emit('open-agent-dialog')"
        >
          <Plus class="size-3.5" />
          添加客服 Profile
        </Button>
        <Button variant="outline" size="sm" :disabled="loadingAgents" @click="emit('refresh')">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loadingAgents }]" />
          刷新概览
        </Button>
      </div>
    </div>

    <section class="grid gap-3 sm:grid-cols-2" aria-label="Public Chat 客服统计">
      <div class="rounded-[20px] border border-dashed border-border/80 bg-muted/20 p-3">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">可分配 Agent Profile</span>
        <strong class="mt-1 block text-2xl font-black italic tracking-tighter tabular-nums">{{ summary.profile_count || 0 }}</strong>
      </div>
      <div class="rounded-[20px] border border-dashed border-border/80 bg-muted/20 p-3">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">目前生效（只读）</span>
        <strong class="mt-1 block text-2xl font-black italic tracking-tighter tabular-nums">{{ summary.exposed_agents || 0 }}</strong>
      </div>
    </section>

    <Alert v-for="warning in warnings" :key="warning" class="border-amber-200 bg-amber-50 text-amber-900">
      <TriangleAlert class="size-4" />
      <AlertTitle>配置提醒</AlertTitle>
      <AlertDescription>{{ warning }}</AlertDescription>
    </Alert>

    <AdminTablePanel :loading="loadingAgents">
      <Table class="min-w-[1280px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>客服</TableHead>
            <TableHead class="w-32">Agent ID</TableHead>
            <TableHead class="w-20">User ID</TableHead>
            <TableHead class="w-36">原始角色</TableHead>
            <TableHead class="w-28">Go 角色</TableHead>
            <TableHead class="w-28">用户状态</TableHead>
            <TableHead class="w-28">Profile</TableHead>
            <TableHead class="w-24">在线状态</TableHead>
            <TableHead class="w-24">公开</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="agents.length === 0" :colspan="10">
            <div class="flex flex-col items-center text-muted-foreground">
              <Headset class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无 Public Chat 客服 Profile</span>
            </div>
          </TableEmpty>
          <TableRow v-for="agent in agents" :key="agent.id">
            <TableCell class="font-mono text-xs text-muted-foreground">{{ agent.id }}</TableCell>
            <TableCell>
              <div class="flex items-center gap-2.5">
                <Avatar class="size-8">
                  <AvatarImage v-if="agent.avatar" :src="agent.avatar" :alt="agent.display_name" />
                  <AvatarFallback>{{ agentInitials(agent) }}</AvatarFallback>
                </Avatar>
                <div class="min-w-0">
                  <span class="block truncate font-bold text-xs">{{ agent.display_name || agent.username || '-' }}</span>
                  <span class="block truncate text-xs text-muted-foreground">{{ agent.email || agent.whatsapp || '-' }}</span>
                </div>
              </div>
            </TableCell>
            <TableCell class="font-mono text-xs">{{ agent.agent_id || '-' }}</TableCell>
            <TableCell class="font-mono text-xs">{{ agent.user_id || '-' }}</TableCell>
            <TableCell>{{ agent.raw_role || '-' }}</TableCell>
            <TableCell>{{ agent.normalized_role || '-' }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="agent.user_status === 'active' ? 'green' : 'gray'">{{ agent.user_status || '-' }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="agent.profile_status === 'active' ? 'green' : 'gray'">{{ agent.profile_status || '-' }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="agent.online_status === 'online' ? 'green' : 'gray'">{{ agent.online_status || '-' }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="agent.exposed ? 'green' : 'coral'">{{ agent.exposed ? '是' : '否' }}</AdminStatusBadge>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>
  </div>
</template>

<script setup>
import { Headset, Info, Plus, RefreshCw, TriangleAlert } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps({
  loadingAgents: { type: Boolean, default: false },
  loadingCandidates: { type: Boolean, default: false },
  summary: { type: Object, default: () => ({}) },
  agents: { type: Array, default: () => [] },
  warnings: { type: Array, default: () => [] },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits(['open-agent-dialog', 'refresh'])

const agentInitials = (agent) => {
  const name = agent.display_name || agent.username || agent.email || '?'
  return name.slice(0, 2).toUpperCase()
}
</script>

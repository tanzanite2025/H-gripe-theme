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
        <Button v-if="canEdit" variant="outline" size="sm" @click="emit('open-group-dialog')">
          <Plus class="size-3.5" />
          添加客服组
        </Button>
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

    <AdminTablePanel :loading="loadingGroups">
      <div class="flex items-center justify-between gap-3 border-b px-4 py-3">
        <div>
          <h3 class="text-sm font-black">客服分组</h3>
          <p class="mt-1 text-xs text-muted-foreground">分组只作为前台客服卡片的展示标签，不参与会话筛选或转接目标选择。</p>
        </div>
        <span class="text-xs font-bold text-muted-foreground">{{ groups.length }} 个组</span>
      </div>
      <Table class="min-w-[920px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead class="w-48">组名称</TableHead>
            <TableHead class="w-40">Code</TableHead>
            <TableHead>说明</TableHead>
            <TableHead class="w-24">排序</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-36 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="groups.length === 0" :colspan="7">
            <span class="text-xs">暂无客服组</span>
          </TableEmpty>
          <TableRow v-for="group in groups" :key="group.id">
            <TableCell class="font-mono text-xs text-muted-foreground">{{ group.id }}</TableCell>
            <TableCell class="font-bold">{{ group.name }}</TableCell>
            <TableCell class="font-mono text-xs">{{ group.code }}</TableCell>
 <TableCell class="max-w-[360px] truncate text-xs text-muted-foreground">{{ group.description || '-'}}</TableCell>
            <TableCell class="font-mono text-xs">{{ group.sort_order || 0 }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="group.status === 'active' ? 'green' : 'gray'">{{ group.status }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <div v-if="canEdit" class="flex justify-end gap-2">
                <Button variant="outline" size="xs" @click="emit('edit-group', group)">
                  <Pencil class="size-3" />
                  编辑
                </Button>
                <Button variant="outline" size="xs" @click="emit('delete-group', group)">
                  <Trash2 class="size-3" />
                  删除
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <section class="grid gap-3 sm:grid-cols-2" aria-label="Public Chat 客服统计">
      <div class="rounded-[20px] border border-dashed border-border/80 bg-muted/20 p-3">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">可分配 Agent Profile</span>
        <strong class="mt-1 block text-2xl font-black tracking-tighter tabular-nums">{{ summary.profile_count || 0 }}</strong>
      </div>
      <div class="rounded-[20px] border border-dashed border-border/80 bg-muted/20 p-3">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">目前生效（只读）</span>
        <strong class="mt-1 block text-2xl font-black tracking-tighter tabular-nums">{{ summary.exposed_agents || 0 }}</strong>
      </div>
    </section>

    <Alert v-for="warning in warnings" :key="warning" class="border-amber-200 bg-amber-50 text-amber-900">
      <TriangleAlert class="size-4" />
      <AlertTitle>配置提醒</AlertTitle>
      <AlertDescription>{{ warning }}</AlertDescription>
    </Alert>

    <AdminTablePanel :loading="loadingAgents">
      <Table class="min-w-[1440px]">
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
            <TableHead class="w-52">展示标签</TableHead>
            <TableHead class="w-24">公开</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="agents.length === 0" :colspan="11">
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
 <span class="block truncate font-bold text-xs">{{ agent.display_name || agent.username || '-'}}</span>
 <span class="block truncate text-xs text-muted-foreground">{{ agent.email || agent.whatsapp || '-'}}</span>
                </div>
              </div>
            </TableCell>
 <TableCell class="font-mono text-xs">{{ agent.agent_id || '-'}}</TableCell>
 <TableCell class="font-mono text-xs">{{ agent.user_id || '-'}}</TableCell>
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
              <div v-if="agent.groups?.length" class="flex max-w-52 flex-wrap gap-1">
                <span
                  v-for="group in agent.groups"
                  :key="group.id"
                  class="rounded-full border border-border bg-muted/50 px-2 py-0.5 text-[10px] font-bold"
                >
                  {{ group.name }}
                </span>
              </div>
              <span v-else class="text-xs text-muted-foreground">未分组</span>
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

<script setup lang="ts">
import { Headset, Info, Pencil, Plus, RefreshCw, Trash2, TriangleAlert } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { PublicChatAgent, PublicChatGroup, PublicChatSummary } from './settingsTypes'

withDefaults(defineProps<{
  loadingAgents?: boolean
  loadingGroups?: boolean
  loadingCandidates?: boolean
  summary?: PublicChatSummary
  agents?: PublicChatAgent[]
  groups?: PublicChatGroup[]
  warnings?: string[]
  canEdit?: boolean
}>(), {
  loadingAgents: false,
  loadingGroups: false,
  loadingCandidates: false,
  summary: () => ({}),
  agents: () => [],
  groups: () => [],
  warnings: () => [],
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'open-agent-dialog'): void
  (event: 'open-group-dialog'): void
  (event: 'edit-group', group: PublicChatGroup): void
  (event: 'delete-group', group: PublicChatGroup): void
  (event: 'refresh'): void
}>()

const agentInitials = (agent: PublicChatAgent): string => {
  const name = agent.display_name || agent.username || agent.email || '?'
  return name.slice(0, 2).toUpperCase()
}
</script>

<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
      <CardTitle>会话列表</CardTitle>
      <CardDescription>按客户会话隔离，每个卡片对应一个 conversation</CardDescription>
    </CardHeader>

    <div class="min-h-0 flex-1 overflow-y-auto p-3">
      <div v-if="loading" class="flex h-52 items-center justify-center text-muted-foreground">
        <LoaderCircle class="size-5 animate-spin" />
      </div>
      <div v-else-if="conversations.length === 0" class="flex h-52 flex-col items-center justify-center text-muted-foreground">
        <MessageCircleOff class="mb-2 size-7 opacity-55" />
        <span class="text-xs">暂无客服对话</span>
      </div>
      <div v-else class="space-y-2">
        <button
          v-for="conversation in conversations"
          :key="conversation.id"
          type="button"
          class="group w-full rounded-2xl border border-dashed p-3 text-left transition-all hover:border-admin-selected-border hover:bg-muted/45"
          :class="selectedConversation?.id === conversation.id ? 'border-admin-selected-border bg-admin-selected-soft shadow-[var(--admin-control-selected-surface-shadow)]' : 'border-border/80 bg-card'"
          @click="emit('select', conversation)"
        >
          <div class="flex items-start gap-3">
            <div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/10 font-black text-primary">
              {{ conversationInitials(conversation) }}
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <strong class="truncate text-xs font-black">{{ conversationDisplayName(conversation) }}</strong>
                <AdminStatusBadge :tone="statusTone(conversation.display_status)">
                  {{ statusLabel(conversation.display_status) }}
                </AdminStatusBadge>
              </div>
              <div class="mt-1 flex flex-wrap items-center gap-1.5 text-[10px] font-black">
                <span :class="identityPillClass(conversation)">
                  <UserCheck v-if="conversationIsMember(conversation)" class="size-3" />
                  <UserRound v-else class="size-3" />
                  {{ customerIdentityLabel(conversation) }}
                </span>
                <span
                  v-if="memberTier(conversation)"
                  class="inline-flex h-5 items-center gap-1 rounded-full border border-amber-500/20 bg-amber-500/10 px-2 text-amber-700"
                  :style="memberTierStyle(conversation)"
                  :title="`${memberTierName(conversation)} · ${Number(memberTier(conversation)?.total_points || 0)} 积分`"
                >
                  <span v-if="memberTierIcon(conversation)" class="leading-none">{{ memberTierIcon(conversation) }}</span>
                  <span class="max-w-20 truncate">{{ memberTierName(conversation) }}</span>
                </span>
                <span class="inline-flex h-5 max-w-full items-center gap-1 rounded-full border border-border bg-muted/55 px-2 text-muted-foreground">
                  <MapPin class="size-3" />
                  <span class="truncate">{{ customerRegionLabel(conversation) }}</span>
                </span>
              </div>
              <p class="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                {{ conversation.last_message || '暂无消息' }}
              </p>
              <div class="mt-2 flex flex-wrap items-center gap-2 text-[10px] font-bold uppercase tracking-wider text-muted-foreground/70">
                <span>#{{ conversation.ticket_number || conversation.id }}</span>
                <span>·</span>
                <span>{{ assigneeName(conversation.assigned_to, assignableAgents) }}</span>
                <span v-if="conversation.unread_count > 0" class="rounded-full bg-rose-500/10 px-2 py-0.5 text-rose-600">
                  {{ conversation.unread_count }} 未读
                </span>
                <span v-if="customerTypingByConversation[conversation.id]?.active" class="rounded-full bg-emerald-500/10 px-2 py-0.5 text-emerald-600">
                  正在输入
                </span>
              </div>
            </div>
          </div>
        </button>
      </div>
    </div>

    <CardFooter class="shrink-0 justify-between gap-3 text-xs text-muted-foreground">
      <span>共 {{ pagination.total }} 条</span>
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" :disabled="pagination.page <= 1 || loading" @click="emit('change-page', pagination.page - 1)">上一页</Button>
        <span class="font-mono text-[10px]">{{ pagination.page }} / {{ totalPages }}</span>
        <Button variant="outline" size="sm" :disabled="pagination.page >= totalPages || loading" @click="emit('change-page', pagination.page + 1)">下一页</Button>
      </div>
    </CardFooter>
  </Card>
</template>

<script setup lang="ts">
import { LoaderCircle, MapPin, MessageCircleOff, UserCheck, UserRound } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import {
  assigneeName,
  conversationDisplayName,
  conversationInitials,
  conversationIsMember,
  customerIdentityLabel,
  customerRegionLabel,
  identityPillClass,
  memberTier,
  memberTierIcon,
  memberTierName,
  memberTierStyle,
  statusLabel,
  statusTone,
} from '@/lib/customerServicePresentation'
import type {
  AssignableAgent,
  CustomerConversation,
  CustomerPagination,
  CustomerTypingByConversation,
} from './customerServiceTypes'

withDefaults(defineProps<{
  conversations?: CustomerConversation[]
  selectedConversation?: CustomerConversation | null
  assignableAgents?: AssignableAgent[]
  customerTypingByConversation?: CustomerTypingByConversation
  pagination: CustomerPagination
  totalPages?: number
  loading?: boolean
}>(), {
  conversations: () => [],
  selectedConversation: null,
  assignableAgents: () => [],
  customerTypingByConversation: () => ({}),
  totalPages: 1,
  loading: false,
})

const emit = defineEmits<{
  (event: 'select', conversation: CustomerConversation): void
  (event: 'change-page', page: number): void
}>()
</script>

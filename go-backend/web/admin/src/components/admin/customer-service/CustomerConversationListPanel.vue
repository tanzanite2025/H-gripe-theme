<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
      <CardTitle>会话列表</CardTitle>
      <CardDescription>按客户会话隔离；归档只清理收件箱，不删除聊天记录</CardDescription>
      <div v-if="canEdit && selectableConversations.length > 0" class="mt-2 flex items-center justify-between gap-2 rounded-xl border bg-background/80 px-3 py-2">
        <label class="flex min-w-0 items-center gap-2 text-xs font-bold">
          <Checkbox
            :model-value="pageSelectionState"
            aria-label="选择当前页可归档会话"
            @update:model-value="emit('toggle-page-selection', selectableConversations, $event === true)"
          />
          <span>{{ selectedConversationIds.length > 0 ? `已选 ${selectedConversationIds.length} 条` : '选择当前页' }}</span>
        </label>
        <Button
          v-if="selectedConversationIds.length > 0"
          variant="outline"
          size="sm"
          :disabled="batchArchiving"
          @click="emit('bulk-archive')"
        >
          <LoaderCircle v-if="batchArchiving" class="size-3.5 animate-spin" />
          <Archive v-else class="size-3.5" />
          批量归档
        </Button>
      </div>
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
        <div
          v-for="conversation in conversations"
          :key="conversation.id"
          role="button"
          tabindex="0"
          class="group w-full rounded-2xl border border-dashed p-3 text-left transition-all hover:border-admin-selected-border hover:bg-muted/45"
 :class="selectedConversation?.id === conversation.id ? 'border-admin-selected-border bg-admin-selected-soft shadow-[var(--admin-control-selected-surface-shadow)]': 'border-border/80 bg-card'"
          @click="emit('select', conversation)"
          @keydown.enter="emit('select', conversation)"
          @keydown.space.prevent="emit('select', conversation)"
        >
          <div class="flex items-start gap-3">
            <div v-if="canEdit && !conversation.inbox_archived" class="mt-2 shrink-0" @click.stop @keydown.stop>
              <Checkbox
                :model-value="isConversationSelected(conversation.id)"
                :aria-label="`选择会话 ${conversationDisplayName(conversation)}`"
                @update:model-value="emit('toggle-selection', conversation, $event === true)"
              />
            </div>
            <div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/10 font-black text-primary">
              {{ conversationInitials(conversation) }}
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <strong class="truncate text-xs font-black">{{ conversationDisplayName(conversation) }}</strong>
                <div class="flex items-center gap-1">
                  <AdminStatusBadge :tone="statusTone(conversation.display_status)">
                    {{ statusLabel(conversation.display_status) }}
                  </AdminStatusBadge>
                  <DropdownMenu v-if="canEdit">
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon" class="size-7 rounded-full" @click.stop>
                        <MoreHorizontal class="size-4" />
                        <span class="sr-only">会话操作</span>
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-40">
                      <DropdownMenuItem v-if="canCloseConversation(conversation)" @select="emit('close-and-archive', conversation)">
                        <CircleCheckBig class="size-4" />
                        关闭并归档
                      </DropdownMenuItem>
                      <DropdownMenuItem v-if="canResolveConversation(conversation)" @select="emit('resolve', conversation)">
                        <CircleCheckBig class="size-4" />
                        标记已解决
                      </DropdownMenuItem>
                      <DropdownMenuItem v-if="canReopenConversation(conversation)" @select="emit('reopen', conversation)">
                        <RotateCcw class="size-4" />
                        重新打开
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem v-if="conversation.inbox_archived" @select="emit('restore', conversation)">
                        <Inbox class="size-4" />
                        恢复到收件箱
                      </DropdownMenuItem>
                      <DropdownMenuItem v-else @select="emit('archive', conversation)">
                        <Archive class="size-4" />
                        归档会话
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
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
              <p
                class="mt-1 truncate text-xs leading-5 text-muted-foreground"
                :title="conversation.last_message || '暂无消息'"
              >
                {{ conversation.last_message || '暂无消息' }}
              </p>
              <div class="mt-2 flex flex-wrap items-center gap-2 text-[10px] font-bold uppercase tracking-wider text-muted-foreground/70">
                <span>#{{ conversation.ticket_number || conversation.id }}</span>
                <span>·</span>
                <span class="max-w-28 truncate normal-case">{{ assigneeName(conversation.assigned_to, assignableAgents, authStore.user) }}</span>
                <span v-if="conversation.unread_count > 0" class="rounded-full bg-rose-500/10 px-2 py-0.5 text-rose-600">
                  {{ conversation.unread_count }} 未读
                </span>
                <span v-if="customerTypingByConversation[conversation.id]?.active" class="rounded-full bg-emerald-500/10 px-2 py-0.5 text-emerald-600">
                  正在输入
                </span>
              </div>
            </div>
          </div>
        </div>
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
import { computed } from 'vue'
import { Archive, CircleCheckBig, Inbox, LoaderCircle, MapPin, MessageCircleOff, MoreHorizontal, RotateCcw, UserCheck, UserRound } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { useAuthStore } from '@/stores/auth'
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
} from '@/modules/customer-service/customerServiceTypes'

const props = withDefaults(defineProps<{
  conversations?: CustomerConversation[]
  selectedConversation?: CustomerConversation | null
  assignableAgents?: AssignableAgent[]
  customerTypingByConversation?: CustomerTypingByConversation
  pagination: CustomerPagination
  totalPages?: number
  loading?: boolean
  canEdit?: boolean
  selectedConversationIds?: Array<string | number>
  batchArchiving?: boolean
}>(), {
  conversations: () => [],
  selectedConversation: null,
  assignableAgents: () => [],
  customerTypingByConversation: () => ({}),
  totalPages: 1,
  loading: false,
  canEdit: false,
  selectedConversationIds: () => [],
  batchArchiving: false,
})

const emit = defineEmits<{
  (event: 'select', conversation: CustomerConversation): void
  (event: 'archive', conversation: CustomerConversation): void
  (event: 'restore', conversation: CustomerConversation): void
  (event: 'close-and-archive', conversation: CustomerConversation): void
  (event: 'resolve', conversation: CustomerConversation): void
  (event: 'reopen', conversation: CustomerConversation): void
  (event: 'toggle-selection', conversation: CustomerConversation, selected: boolean): void
  (event: 'toggle-page-selection', conversations: CustomerConversation[], selected: boolean): void
  (event: 'bulk-archive'): void
  (event: 'change-page', page: number): void
}>()

const authStore = useAuthStore()

const normalizeConversationID = (id: string | number): string => String(id)
const selectedConversationIDSet = computed(() => new Set(props.selectedConversationIds.map(normalizeConversationID)))
const selectableConversations = computed(() => props.conversations.filter((conversation) => !conversation.inbox_archived))
const selectedOnPageCount = computed(() => selectableConversations.value.filter((conversation) => selectedConversationIDSet.value.has(normalizeConversationID(conversation.id))).length)
const pageSelectionState = computed(() => {
  if (selectableConversations.value.length === 0 || selectedOnPageCount.value === 0) return false
  return selectedOnPageCount.value === selectableConversations.value.length ? true : 'indeterminate'
})

const isConversationSelected = (id: string | number): boolean => selectedConversationIDSet.value.has(normalizeConversationID(id))
const lifecycleStatus = (conversation: CustomerConversation): string => String(conversation.status || '').trim().toLowerCase()
const canCloseConversation = (conversation: CustomerConversation): boolean => !['resolved', 'closed'].includes(lifecycleStatus(conversation))
const canResolveConversation = (conversation: CustomerConversation): boolean => ['open', 'in_progress', ''].includes(lifecycleStatus(conversation))
const canReopenConversation = (conversation: CustomerConversation): boolean => ['resolved', 'closed'].includes(lifecycleStatus(conversation))
</script>


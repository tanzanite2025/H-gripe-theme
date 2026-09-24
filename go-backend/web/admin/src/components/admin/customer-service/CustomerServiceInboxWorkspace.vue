<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
    <div class="flex shrink-0 items-center justify-between gap-2">
      <div class="min-w-0">
        <p class="truncate text-sm font-black">客服工作台</p>
      </div>
      <Button
        variant="outline"
        size="sm"
        class="shrink-0 rounded-full"
        @click="notificationSettingsOpen = true"
      >
        <Bell class="size-3.5" />
        通知设置
      </Button>
    </div>

    <div class="flex shrink-0 items-center justify-between gap-2 lg:hidden">
      <div class="inline-flex min-w-0 flex-1 items-center rounded-lg border bg-muted/35 p-1">
        <Button
          variant="ghost"
          size="sm"
          class="min-w-0 flex-1 rounded-md text-xs"
          :class="{ 'bg-background shadow-sm': mobileView === 'conversations' }"
          @click="mobileView = 'conversations'"
        >
          <List class="size-3.5" />
          会话
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="min-w-0 flex-1 rounded-md text-xs"
 :class="{ 'bg-background shadow-sm': mobileView === 'messages'}"
          :disabled="!selectedConversation"
          @click="mobileView = 'messages'"
        >
          <MessageSquare class="size-3.5" />
          对话
        </Button>
      </div>
      <Button
        variant="outline"
        size="icon"
        class="shrink-0 rounded-full"
        aria-label="筛选客服会话"
        @click="mobileFiltersOpen = true"
      >
        <SlidersHorizontal class="size-4" />
      </Button>
    </div>

    <div class="hidden shrink-0 lg:block">
      <CustomerServiceFilters
        :filters="filters"
        :loading="loading"
        @apply="emit('apply')"
        @reset="emit('reset')"
      />
    </div>

    <section class="hidden min-h-0 flex-1 gap-4 overflow-hidden lg:grid lg:grid-cols-[minmax(300px,320px)_minmax(0,1fr)] lg:grid-rows-[minmax(0,1fr)] xl:grid-cols-[320px_minmax(0,1fr)] 2xl:grid-cols-[360px_minmax(0,1fr)]">
      <CustomerConversationListPanel
        :conversations="conversations"
        :selected-conversation="selectedConversation"
        :assignable-agents="assignableAgents"
        :customer-typing-by-conversation="customerTypingByConversation"
        :pagination="pagination"
        :total-pages="totalPages"
        :loading="loading"
        :can-edit="canEdit"
        :selected-conversation-ids="selectedConversationIds"
        :batch-archiving="batchArchiving"
        @select="emit('select', $event)"
        @archive="emit('archive', $event)"
        @restore="emit('restore', $event)"
        @close-and-archive="emit('close-and-archive', $event)"
        @resolve="emit('resolve', $event)"
        @reopen="emit('reopen', $event)"
        @toggle-selection="forwardToggleSelection"
        @toggle-page-selection="forwardTogglePageSelection"
        @bulk-archive="emit('bulk-archive')"
        @change-page="emit('change-page', $event)"
      />

      <CustomerConversationDetailPanel
        :transfer-to="transferTo"
        :reply-message="replyMessage"
        :selected-conversation="selectedConversation"
        :customer-context="customerContext"
        :messages="messages"
        :messages-loading="messagesLoading"
        :selected-customer-typing="selectedCustomerTyping"
        :assignable-agents="assignableAgents"
        :transferring="transferring"
        :replying="replying"
        :can-edit="canEdit"
        @update:transfer-to="emit('update:transferTo', $event)"
        @update:reply-message="emit('update:replyMessage', $event)"
        @transfer="emit('transfer')"
        @send-reply="emit('send-reply')"
        @send-message="emit('send-message', $event)"
        @open-context="openContextDialog"
        @typing-input="emit('typing-input')"
      />
    </section>

    <section class="min-h-0 flex-1 lg:hidden">
      <CustomerConversationListPanel
        v-if="mobileView === 'conversations'"
        :conversations="conversations"
        :selected-conversation="selectedConversation"
        :assignable-agents="assignableAgents"
        :customer-typing-by-conversation="customerTypingByConversation"
        :pagination="pagination"
        :total-pages="totalPages"
        :loading="loading"
        :can-edit="canEdit"
        :selected-conversation-ids="selectedConversationIds"
        :batch-archiving="batchArchiving"
        @select="selectMobileConversation"
        @archive="emit('archive', $event)"
        @restore="emit('restore', $event)"
        @close-and-archive="emit('close-and-archive', $event)"
        @resolve="emit('resolve', $event)"
        @reopen="emit('reopen', $event)"
        @toggle-selection="forwardToggleSelection"
        @toggle-page-selection="forwardTogglePageSelection"
        @bulk-archive="emit('bulk-archive')"
        @change-page="emit('change-page', $event)"
      />

      <CustomerConversationDetailPanel
        v-else-if="mobileView === 'messages'"
        :transfer-to="transferTo"
        :reply-message="replyMessage"
        :selected-conversation="selectedConversation"
        :customer-context="customerContext"
        :messages="messages"
        :messages-loading="messagesLoading"
        :selected-customer-typing="selectedCustomerTyping"
        :assignable-agents="assignableAgents"
        :transferring="transferring"
        :replying="replying"
        :can-edit="canEdit"
        @update:transfer-to="emit('update:transferTo', $event)"
        @update:reply-message="emit('update:replyMessage', $event)"
        @transfer="emit('transfer')"
        @send-reply="emit('send-reply')"
        @send-message="emit('send-message', $event)"
        @open-context="openContextDialog"
        @typing-input="emit('typing-input')"
      />
    </section>
  </div>

  <Sheet v-model:open="mobileFiltersOpen">
    <SheetContent side="bottom" class="max-h-[86dvh] rounded-t-2xl p-0" @open-auto-focus.prevent>
      <SheetHeader class="border-b pr-12">
        <SheetTitle>筛选客服会话</SheetTitle>
        <SheetDescription>按客户、状态和未读消息筛选。</SheetDescription>
      </SheetHeader>
      <div class="overflow-y-auto p-4">
        <CustomerServiceFilters
          :filters="filters"
          :loading="loading"
          @apply="applyMobileFilters"
          @reset="resetMobileFilters"
        />
      </div>
    </SheetContent>
  </Sheet>

  <CustomerServiceContextDialog
    v-model:open="contextDialogOpen"
    :selected-conversation="selectedConversation"
    :customer-context="customerContext"
    :loading="contextLoading"
    :context-error="contextError"
    :context-last-updated-at="contextLastUpdatedAt"
  />

  <Dialog v-model:open="notificationSettingsOpen">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle>通知设置</DialogTitle>
        <DialogDescription>新客户消息到达时的提醒方式。</DialogDescription>
      </DialogHeader>
      <div class="space-y-4 py-2">
        <label class="flex items-center justify-between gap-4 rounded-xl border p-3">
          <span class="flex min-w-0 items-center gap-2 text-sm font-bold"><Volume2 class="size-4 text-primary" />提示音</span>
          <Switch :checked="soundEnabled" aria-label="启用客服消息提示音" @update:checked="setSoundEnabled" />
        </label>
        <div class="rounded-xl border p-3">
          <div class="mb-2 flex items-center justify-between gap-3 text-sm">
            <label for="customer-service-notification-volume" class="flex items-center gap-2 font-bold">
              <Volume1 class="size-4 text-primary" />音量
            </label>
            <span class="font-mono text-xs text-muted-foreground">{{ Math.round(soundVolume * 100) }}%</span>
          </div>
          <input
            id="customer-service-notification-volume"
            class="h-2 w-full cursor-pointer accent-primary"
            type="range"
            min="0"
            max="1"
            step="0.01"
            :value="soundVolume"
            :disabled="!soundEnabled"
            aria-label="客服消息提示音音量"
            @input="setSoundVolume(Number(($event.target as HTMLInputElement).value))"
          />
        </div>
        <label class="flex items-center justify-between gap-4 rounded-xl border p-3">
          <span class="flex min-w-0 items-center gap-2 text-sm font-bold"><Monitor class="size-4 text-primary" />桌面通知</span>
          <Switch :checked="desktopEnabled" aria-label="启用客服桌面通知" @update:checked="setDesktopEnabled" />
        </label>
        <div class="flex items-center justify-between gap-3 rounded-xl bg-muted/40 px-3 py-2 text-xs">
          <span class="text-muted-foreground">浏览器授权：{{ notificationPermissionLabel }}</span>
          <Button v-if="notificationPermission === 'default'" variant="outline" size="sm" class="rounded-full" @click="requestPermission">授权</Button>
          <span v-else-if="notificationPermission === 'denied'" class="text-[11px] text-muted-foreground">请在浏览器设置中允许</span>
        </div>
        <label class="flex items-start justify-between gap-4 rounded-xl border p-3">
          <span class="min-w-0">
            <span class="flex items-center gap-2 text-sm font-bold"><MessageSquare class="size-4 text-primary" />聚焦会话时静默桌面通知</span>
            <span class="mt-1 block text-[11px] leading-5 text-muted-foreground">当前会话已在前台打开时，仅保留站内未读和提示音。</span>
          </span>
          <Switch :checked="suppressDesktopWhenFocused" aria-label="聚焦会话时静默桌面通知" @update:checked="setSuppressDesktopWhenFocused" />
        </label>
        <div class="rounded-xl border p-3">
          <label class="flex items-center justify-between gap-4">
            <span class="text-sm font-bold">夜间免打扰</span>
            <Switch :checked="quietHoursEnabled" aria-label="启用夜间免打扰" @update:checked="setQuietHoursEnabled" />
          </label>
          <div class="mt-3 grid grid-cols-2 gap-3">
            <label class="text-xs text-muted-foreground">开始<input class="mt-1 w-full rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="time" :value="quietHoursStart" :disabled="!quietHoursEnabled" @change="setQuietHoursStart(($event.target as HTMLInputElement).value)" /></label>
            <label class="text-xs text-muted-foreground">结束<input class="mt-1 w-full rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="time" :value="quietHoursEnd" :disabled="!quietHoursEnabled" @change="setQuietHoursEnd(($event.target as HTMLInputElement).value)" /></label>
          </div>
          <p class="mt-2 text-[11px] leading-5 text-muted-foreground">按当前客服浏览器本地时间静默提示音和桌面通知，不影响未读状态。</p>
        </div>
        <Button type="button" variant="outline" class="w-full rounded-full" :disabled="!soundEnabled" @click="testNotificationSound">
          <Volume2 class="size-4" />
          测试提示音
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Bell, List, MessageSquare, Monitor, SlidersHorizontal, Volume1, Volume2 } from '@lucide/vue'
import CustomerServiceContextDialog from '@/components/admin/customer-service/CustomerServiceContextDialog.vue'
import CustomerConversationDetailPanel from '@/components/admin/customer-service/CustomerConversationDetailPanel.vue'
import CustomerConversationListPanel from '@/components/admin/customer-service/CustomerConversationListPanel.vue'
import CustomerServiceFilters from '@/components/admin/customer-service/CustomerServiceFilters.vue'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Switch } from '@/components/ui/switch'
import { useCustomerServiceNotificationSettings } from '@/composables/customerService/useCustomerServiceRealtime'
import type {
  AssignableAgent,
  CustomerContext,
  CustomerConversation,
  CustomerConversationMessage,
  CustomerPagination,
  CustomerServiceFiltersState,
  CustomerServiceSendMessagePayload,
  CustomerTypingByConversation,
  CustomerTypingState,
} from '@/modules/customer-service/customerServiceTypes'

const props = withDefaults(defineProps<{
  filters: CustomerServiceFiltersState
  conversations?: CustomerConversation[]
  selectedConversation?: CustomerConversation | null
  customerTypingByConversation?: CustomerTypingByConversation
  pagination: CustomerPagination
  totalPages?: number
  loading?: boolean
  transferTo?: string
  replyMessage?: string
  messages?: CustomerConversationMessage[]
  messagesLoading?: boolean
  selectedCustomerTyping?: CustomerTypingState | null
  assignableAgents?: AssignableAgent[]
  transferring?: boolean
  replying?: boolean
  canEdit?: boolean
  customerContext?: CustomerContext | null
  contextLoading?: boolean
  contextError?: string | null
  contextLastUpdatedAt?: Date | null
  selectedConversationIds?: Array<string | number>
  batchArchiving?: boolean
}>(), {
  conversations: () => [],
  selectedConversation: null,
  customerTypingByConversation: () => ({}),
  totalPages: 1,
  loading: false,
  transferTo: '',
  replyMessage: '',
  messages: () => [],
  messagesLoading: false,
  selectedCustomerTyping: null,
  assignableAgents: () => [],
  transferring: false,
  replying: false,
  canEdit: false,
  customerContext: null,
  contextLoading: false,
  contextError: null,
  contextLastUpdatedAt: null,
  selectedConversationIds: () => [],
  batchArchiving: false,
})

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
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
  (event: 'update:transferTo', value: string): void
  (event: 'update:replyMessage', value: string): void
  (event: 'transfer'): void
  (event: 'send-reply'): void
  (event: 'send-message', payload: CustomerServiceSendMessagePayload): void
  (event: 'refresh-context'): void
  (event: 'typing-input'): void
}>()

type MobileInboxView = 'conversations' | 'messages'

const mobileView = ref<MobileInboxView>('conversations')
const mobileFiltersOpen = ref(false)
const contextDialogOpen = ref(false)
const notificationSettingsOpen = ref(false)
const {
  soundEnabled,
  soundVolume,
  desktopEnabled,
  suppressDesktopWhenFocused,
  quietHoursEnabled,
  quietHoursStart,
  quietHoursEnd,
  notificationPermission,
  setSoundEnabled,
  setSoundVolume,
  setDesktopEnabled,
  setSuppressDesktopWhenFocused,
  setQuietHoursEnabled,
  setQuietHoursStart,
  setQuietHoursEnd,
  testNotificationSound,
  requestPermission,
} = useCustomerServiceNotificationSettings()

const notificationPermissionLabel = computed(() => {
  if (notificationPermission.value === 'granted') return '已允许'
  if (notificationPermission.value === 'denied') return '已阻止'
  if (notificationPermission.value === 'default') return '未设置'
  return '不可用'
})

const selectMobileConversation = (conversation: CustomerConversation) => {
  mobileView.value = 'messages'
  emit('select', conversation)
}

const openContextDialog = () => {
  if (!props.selectedConversation) return
  contextDialogOpen.value = true
  emit('refresh-context')
}

const forwardToggleSelection = (conversation: CustomerConversation, selected: boolean) => {
  emit('toggle-selection', conversation, selected)
}

const forwardTogglePageSelection = (conversations: CustomerConversation[], selected: boolean) => {
  emit('toggle-page-selection', conversations, selected)
}

const applyMobileFilters = () => {
  mobileFiltersOpen.value = false
  emit('apply')
}

const resetMobileFilters = () => {
  mobileFiltersOpen.value = false
  emit('reset')
}
</script>


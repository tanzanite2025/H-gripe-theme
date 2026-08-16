<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <button
        type="button"
        class="fixed bottom-5 right-5 z-40 grid size-14 place-items-center rounded-full border border-border/80 bg-background text-foreground shadow-[0_16px_38px_rgba(15,23,42,0.24)] transition-transform duration-200 hover:-translate-y-1 hover:shadow-[0_20px_46px_rgba(15,23,42,0.32)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 sm:bottom-6 sm:right-6"
        :class="{ 'ring-4 ring-rose-500/10': unreadConversationCount > 0 }"
        :aria-label="inboxAriaLabel"
        aria-haspopup="dialog"
        :aria-expanded="inboxOpen"
        @click="inboxOpen = true"
      >
        <Headset class="size-6" aria-hidden="true" />
        <span
          v-if="unreadConversationCount > 0"
          class="absolute -right-1 -top-1 grid min-h-5 min-w-5 place-items-center rounded-full border-2 border-background bg-rose-500 px-1 text-[10px] font-black leading-none text-white"
          aria-hidden="true"
        >
          {{ unreadBadgeLabel }}
        </span>
      </button>
    </TooltipTrigger>
    <TooltipContent side="left" :side-offset="12">
      {{ inboxTooltipLabel }}
    </TooltipContent>
  </Tooltip>

  <CustomerServiceInboxDialog v-model:open="inboxOpen" />
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Headset } from '@lucide/vue'
import { toast } from 'vue-sonner'
import { useRoute, useRouter } from 'vue-router'
import customerServiceApi from '@/api/customerService'
import CustomerServiceInboxDialog from '@/components/admin/customer-service/CustomerServiceInboxDialog.vue'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useCustomerServiceRealtime } from '@/composables/customerService/useCustomerServiceRealtime'

const route = useRoute()
const router = useRouter()
const inboxOpen = ref(false)
const unreadConversationCount = ref(0)

let unreadRefreshTimer: number | null = null

const unreadBadgeLabel = computed(() => (
  unreadConversationCount.value > 99 ? '99+' : String(unreadConversationCount.value)
))

const inboxTooltipLabel = computed(() => (
  unreadConversationCount.value > 0
    ? `客服会话 (${unreadConversationCount.value} 个未读)`
    : '客服会话'
))

const inboxAriaLabel = computed(() => (
  unreadConversationCount.value > 0
    ? `打开客服会话，${unreadConversationCount.value} 个未读会话`
    : '打开客服会话'
))

const refreshUnreadConversationCount = async () => {
  try {
    const data = await customerServiceApi.listConversations({
      page: 1,
      page_size: 1,
      unread: 'true',
    })
    unreadConversationCount.value = Math.max(0, Number(data.pagination?.total || 0))
  } catch (error) {
    console.error('Failed to refresh customer-service unread summary:', error)
  }
}

const scheduleUnreadRefresh = () => {
  if (unreadRefreshTimer) {
    window.clearTimeout(unreadRefreshTimer)
  }

  unreadRefreshTimer = window.setTimeout(() => {
    unreadRefreshTimer = null
    void refreshUnreadConversationCount()
  }, 250)
}

const isCustomerMessageEvent = (event: Record<string, any>) => (
  event.type === 'conversation.message.created' && event.actor?.kind === 'customer'
)

const {
  connectCustomerServiceRealtime,
  closeCustomerServiceRealtime,
} = useCustomerServiceRealtime({
  buildWebSocketUrl: (lastEventId: string) => customerServiceApi.buildWebSocketUrl('inbox', undefined, lastEventId),
  connectionKey: () => 'inbox',
  onConnected: refreshUnreadConversationCount,
  onRefresh: async (event: Record<string, any>) => {
    scheduleUnreadRefresh()

    if (isCustomerMessageEvent(event) && !inboxOpen.value) {
      toast.info('收到新的客户消息', {
        id: `customer-service-message-${event.event_id || event.ticket_id || Date.now()}`,
        action: {
          label: '打开',
          onClick: () => {
            inboxOpen.value = true
          },
        },
      })
    }
  },
})

const removeInboxQuery = async () => {
  if (route.query.inbox !== 'customer-service') return

  const { inbox, ...query } = route.query
  await router.replace({ query })
}

watch(
  () => route.query.inbox,
  (inbox) => {
    if (inbox === 'customer-service') {
      inboxOpen.value = true
    }
  },
  { immediate: true },
)

watch(inboxOpen, (open) => {
  if (open) {
    closeCustomerServiceRealtime()
    void refreshUnreadConversationCount()
    return
  }
  void nextTick(async () => {
    if (!inboxOpen.value) {
      void refreshUnreadConversationCount()
      connectCustomerServiceRealtime()
    }
    await removeInboxQuery()
  })
})

onMounted(() => {
  void refreshUnreadConversationCount()
  if (!inboxOpen.value) {
    connectCustomerServiceRealtime()
  }
})

onBeforeUnmount(() => {
  closeCustomerServiceRealtime()
  if (unreadRefreshTimer) {
    window.clearTimeout(unreadRefreshTimer)
    unreadRefreshTimer = null
  }
})
</script>

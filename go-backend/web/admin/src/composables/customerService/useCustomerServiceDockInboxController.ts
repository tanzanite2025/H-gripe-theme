import { computed, nextTick, onScopeDispose, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { useAdminI18n } from '@/i18n'
import customerServiceApi from '@/api/customerService'
import { useCustomerServiceRealtime } from './useCustomerServiceRealtime'
import { requestCustomerServiceNotificationPermission } from './useCustomerServiceNotificationSettingsAndQuietHours'
import { useCustomerServiceUnreadConversationSummary } from './useCustomerServiceUnreadConversationSummary'

export function useCustomerServiceDockInboxController() {
  const route = useRoute()
  const router = useRouter()
  const { t } = useAdminI18n()
  const inboxOpen = ref(false)
  let disposed = false
  const {
    unreadConversationCount,
    refreshCustomerServiceUnreadConversationCount,
    scheduleCustomerServiceUnreadConversationCountRefresh,
  } = useCustomerServiceUnreadConversationSummary()

  const unreadBadgeLabel = computed(() => unreadConversationCount.value > 99 ? '99+' : String(unreadConversationCount.value))
  const inboxTooltipLabel = computed(() => unreadConversationCount.value > 0
    ? t('customerServiceLauncher.unreadTooltip', { count: unreadConversationCount.value })
    : t('customerServiceLauncher.tooltip'))
  const inboxAriaLabel = computed(() => unreadConversationCount.value > 0
    ? t('customerServiceLauncher.openUnread', { count: unreadConversationCount.value })
    : t('customerServiceLauncher.open'))

  function openCustomerServiceInboxFromDock(): void {
    if (disposed) return
    void requestCustomerServiceNotificationPermission()
    inboxOpen.value = true
  }

  const { connectCustomerServiceRealtime, closeCustomerServiceRealtime } = useCustomerServiceRealtime({
    buildWebSocketUrl: (lastEventId: string) => customerServiceApi.buildWebSocketUrl('inbox', undefined, lastEventId),
    connectionKey: () => 'inbox',
    onConnected: refreshCustomerServiceUnreadConversationCount,
    onRefresh: (event: Record<string, any>) => {
      if (disposed) return
      scheduleCustomerServiceUnreadConversationCountRefresh()
      if (event.type === 'conversation.message.created' && event.actor?.kind === 'customer' && !inboxOpen.value) {
        toast.info(t('customerServiceLauncher.newCustomerMessage'), {
          id: `customer-service-message-${event.event_id || event.ticket_id || Date.now()}`,
          action: { label: t('customerServiceLauncher.openNotificationAction'), onClick: openCustomerServiceInboxFromDock },
        })
      }
    },
  })

  async function removeCustomerServiceInboxQueryAfterClosing(): Promise<void> {
    if (disposed || inboxOpen.value || route.query.inbox !== 'customer-service') return
    const { inbox, ...query } = route.query
    try {
      await router.replace({ query })
    } catch (error) {
      if (!disposed) console.error('Failed to clear customer-service inbox query:', error)
    }
  }

  watch(() => route.query.inbox, (inbox) => {
    if (inbox === 'customer-service') inboxOpen.value = true
  }, { immediate: true })

  watch(inboxOpen, (open, wasOpen) => {
    if (open) {
      closeCustomerServiceRealtime()
      void refreshCustomerServiceUnreadConversationCount()
      return
    }
    void nextTick(() => {
      if (disposed || inboxOpen.value) return
      void refreshCustomerServiceUnreadConversationCount()
      connectCustomerServiceRealtime()
      if (wasOpen) void removeCustomerServiceInboxQueryAfterClosing()
    })
  }, { immediate: true, flush: 'sync' })

  onScopeDispose(() => {
    disposed = true
    closeCustomerServiceRealtime()
  })

  return { inboxOpen, unreadConversationCount, unreadBadgeLabel, inboxTooltipLabel, inboxAriaLabel, openCustomerServiceInboxFromDock }
}

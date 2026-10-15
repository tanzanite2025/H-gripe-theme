<template>
  <CustomerServiceInboxWorkbench show-header show-stats :conversation-navigation-target="conversationNavigationTarget" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import CustomerServiceInboxWorkbench from '@/components/admin/customer-service/CustomerServiceInboxWorkbench.vue'
import { parsePositiveUint32RouteQueryParameter } from '@/lib/parsePositiveUint32RouteQueryParameter'
import type { CustomerServiceConversationNavigationTarget } from '@/modules/customer-service/customerServiceTypes'

const route = useRoute()
const conversationNavigationTarget = computed<CustomerServiceConversationNavigationTarget | null>(() => {
  const rawId = route.query.conversation_id
  if (!rawId) return null
  const conversationId = parsePositiveUint32RouteQueryParameter(rawId)
  return {
    conversationId,
    invalidId: conversationId === null,
    searchTerm: typeof route.query.search === 'string' ? route.query.search : '',
  }
})
</script>

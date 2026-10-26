<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <AdminBottomUtilityDockActionButton
        appearance="support"
        :badge-label="unreadConversationCount > 0 ? unreadBadgeLabel : ''"
        :aria-label="inboxAriaLabel"
        aria-haspopup="dialog"
        :aria-expanded="inboxOpen"
        @click="openCustomerServiceInboxFromDock"
      >
        <Headset class="size-6" aria-hidden="true" />
      </AdminBottomUtilityDockActionButton>
    </TooltipTrigger>
    <TooltipContent side="top" :side-offset="12">
      {{ inboxTooltipLabel }}
    </TooltipContent>
  </Tooltip>

  <CustomerServiceInboxDialog v-model:open="inboxOpen" />
</template>

<script setup lang="ts">
import { Headset } from '@lucide/vue'
import CustomerServiceInboxDialog from '@/components/admin/customer-service/CustomerServiceInboxDialog.vue'
import AdminBottomUtilityDockActionButton from '@/components/admin/AdminBottomUtilityDockActionButton.vue'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useCustomerServiceDockInboxController } from '@/composables/customerService/useCustomerServiceDockInboxController'

const {
  inboxOpen,
  unreadConversationCount,
  unreadBadgeLabel,
  inboxTooltipLabel,
  inboxAriaLabel,
  openCustomerServiceInboxFromDock,
} = useCustomerServiceDockInboxController()
</script>

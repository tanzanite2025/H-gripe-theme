<template>
  <AdminBottomUtilityDock>
    <AdminGlobalSearchDockAction :navigation-items="navigationItems" />
    <CustomerServiceDockInboxAction v-if="canViewCustomerService" />
  </AdminBottomUtilityDock>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import AdminBottomUtilityDock from './AdminBottomUtilityDock.vue'
import AdminGlobalSearchDockAction from './search/AdminGlobalSearchDockAction.vue'
import type { AdminNavigationItem } from '@/lib/adminNavigation'

defineProps<{
  navigationItems: AdminNavigationItem[]
  canViewCustomerService: boolean
}>()

const CustomerServiceDockInboxAction = defineAsyncComponent(
  () => import('@/components/admin/customer-service/CustomerServiceDockInboxAction.vue'),
)
</script>

<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="保修管理"
      description="统一查看产品注册、保修申请和即将到期记录；当前只消费 Go 后端保修事实源。"
    >
      <template #actions>
        <Button variant="outline" :disabled="refreshing" @click="refreshCurrent">
          <RefreshCw :class="['size-3.5', { 'animate-spin': refreshing }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <Tabs v-model="activeTab" class="gap-4">
      <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
        <TabsTrigger value="registrations" class="h-9 flex-none px-3">
          <ShieldCheck class="size-4" />
          注册记录
        </TabsTrigger>
        <TabsTrigger value="claims" class="h-9 flex-none px-3">
          <FileWarning class="size-4" />
          保修申请
        </TabsTrigger>
        <TabsTrigger value="expiring" class="h-9 flex-none px-3">
          <Clock3 class="size-4" />
          即将到期
        </TabsTrigger>
        <TabsTrigger value="boundary" class="h-9 flex-none px-3">
          <GitBranch class="size-4" />
          数据边界
        </TabsTrigger>
      </TabsList>

      <WarrantyRegistrationsTab
        :registrations="registrations"
        :loading="loading.registrations"
        :filters="registrationFilters"
        :pagination="registrationPagination"
        :status-updating="statusUpdating"
        :status-options="registrationStatusOptions"
        :can-edit="canEdit"
        @update-status="updateRegistrationStatus"
        @update-page="updateRegistrationPage"
        @update-page-size="updateRegistrationPageSize"
      />

      <WarrantyClaimsTab
        :claims="claims"
        :loading="loading.claims"
        :detail-loading="loading.claimDetail"
        :resolution-saving="loading.claimResolution"
        :order-items-loading="loading.claimOrderItems"
        :order-item-binding="loading.claimOrderItemBinding"
        :service-records-loading="loading.serviceRecords"
        :service-record-creating="loading.serviceRecordCreating"
        :filters="claimFilters"
        :pagination="claimPagination"
        :status-updating="statusUpdating"
        :status-options="claimStatusOptions"
        :can-edit="canEdit"
        :selected-claim="selectedClaim"
        :resolution-draft="claimResolutionDraft"
        :order-items="claimOrderItems"
        :order-item-selection="orderItemSelection"
        :service-records="serviceRecords"
        :service-record-form="serviceRecordForm"
        :service-type-options="serviceTypeOptions"
        :service-status-options="serviceStatusOptions"
        @update-status="updateClaimStatus"
        @select-claim="selectClaim"
        @update-order-item-selection="updateOrderItemSelection"
        @bind-order-item="bindClaimOrderItem"
        @update-resolution-draft="updateClaimResolutionDraft"
        @save-resolution="saveClaimResolution"
        @update-service-record-form="updateServiceRecordForm"
        @create-service-record="createServiceRecord"
        @update-page="updateClaimPage"
        @update-page-size="updateClaimPageSize"
      />

      <WarrantyExpiringTab
        :expiring="expiring"
        :loading="loading.expiring"
      />

      <WarrantyBoundaryTab />
    </Tabs>
  </div>
</template>

<script setup>
import {
  Clock3,
  FileWarning,
  GitBranch,
  RefreshCw,
  ShieldCheck
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import WarrantyBoundaryTab from '@/components/admin/warranty/WarrantyBoundaryTab.vue'
import WarrantyClaimsTab from '@/components/admin/warranty/WarrantyClaimsTab.vue'
import WarrantyExpiringTab from '@/components/admin/warranty/WarrantyExpiringTab.vue'
import WarrantyRegistrationsTab from '@/components/admin/warranty/WarrantyRegistrationsTab.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useWarrantyAdmin } from '@/composables/warranty/useWarrantyAdmin'

const {
  activeTab,
  refreshing,
  registrations,
  claims,
  expiring,
  selectedClaim,
  claimResolutionDraft,
  claimOrderItems,
  serviceRecords,
  orderItemSelection,
  serviceRecordForm,
  loading,
  statusUpdating,
  registrationFilters,
  claimFilters,
  registrationPagination,
  claimPagination,
  canEdit,
  statItems,
  registrationStatusOptions,
  claimStatusOptions,
  serviceTypeOptions,
  serviceStatusOptions,
  refreshCurrent,
  updateRegistrationStatus,
  updateClaimStatus,
  selectClaim,
  updateOrderItemSelection,
  bindClaimOrderItem,
  updateClaimResolutionDraft,
  saveClaimResolution,
  updateServiceRecordForm,
  createServiceRecord,
  updateRegistrationPage,
  updateRegistrationPageSize,
  updateClaimPage,
  updateClaimPageSize
} = useWarrantyAdmin()
</script>

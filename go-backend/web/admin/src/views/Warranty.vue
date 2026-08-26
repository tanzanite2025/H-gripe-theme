<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="保修管理"
      description="从已发货订单补充售后凭据，并处理保修申请和服务记录。"
    >
      <template #actions>
        <Button variant="outline" :disabled="refreshing" @click="refreshCurrent">
          <RefreshCw :class="['size-3.5', { 'animate-spin': refreshing }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <Tabs :model-value="activeTab" class="gap-4">
      <WarrantyShipmentsTab
        :shipments="shipments"
        :loading="loading.shipments"
        :detail-loading="loading.shipmentDetail"
        :saving="loading.shipmentSaving"
        :uploading="loading.shipmentUploading"
        :filters="shipmentFilters"
        :pagination="shipmentPagination"
        :selected-shipment="selectedShipment"
        :draft="shipmentDraft"
        :can-edit="canEdit"
        @select-shipment="selectShipment"
        @update-search="shipmentFilters.keyword = $event"
        @update-draft="updateShipmentDraft"
        @remove-image="removeShipmentImage"
        @save="saveShipment"
        @upload-images="uploadShipmentImages"
        @update-page="updateShipmentPage"
        @update-page-size="updateShipmentPageSize"
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

      <WarrantyBoundaryTab />
    </Tabs>
  </div>
</template>

<script setup lang="ts">
import {
  RefreshCw,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import WarrantyBoundaryTab from '@/components/admin/warranty/WarrantyBoundaryTab.vue'
import WarrantyClaimsTab from '@/components/admin/warranty/WarrantyClaimsTab.vue'
import WarrantyShipmentsTab from '@/components/admin/warranty/WarrantyShipmentsTab.vue'
import { Button } from '@/components/ui/button'
import { Tabs } from '@/components/ui/tabs'
import { useWarrantyAdmin } from '@/composables/warranty/useWarrantyAdmin'

const {
  activeTab,
  refreshing,
  shipments,
  claims,
  selectedShipment,
  shipmentDraft,
  selectedClaim,
  claimResolutionDraft,
  claimOrderItems,
  serviceRecords,
  orderItemSelection,
  serviceRecordForm,
  loading,
  statusUpdating,
  shipmentFilters,
  claimFilters,
  shipmentPagination,
  claimPagination,
  canEdit,
  statItems,
  claimStatusOptions,
  serviceTypeOptions,
  serviceStatusOptions,
  refreshCurrent,
  selectShipment,
  updateShipmentDraft,
  removeShipmentImage,
  saveShipment,
  uploadShipmentImages,
  updateClaimStatus,
  selectClaim,
  updateOrderItemSelection,
  bindClaimOrderItem,
  updateClaimResolutionDraft,
  saveClaimResolution,
  updateServiceRecordForm,
  createServiceRecord,
  updateShipmentPage,
  updateShipmentPageSize,
  updateClaimPage,
  updateClaimPageSize
} = useWarrantyAdmin()
</script>

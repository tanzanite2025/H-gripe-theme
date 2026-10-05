<template>
  <div class="space-y-4">
    <AdminPageHeader title="物流管理" description="管理承运商、包装规则、运费模板和 17TRACK 追踪配置。">
      <template #actions>
        <Button variant="outline" :disabled="refreshing" @click="refreshCurrentTab">
          <RefreshCw :class="['size-3.5', { 'animate-spin': refreshing }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <ShippingTabsPanel
      ref="trackingShipmentsPanelRef"
      :active-tab="activeTab"
      :templates="templates"
      :carriers="carriers"
      :carrier-services="carrierServices"
      :fpx-channels="fpxChannels"
      :tracking-providers="trackingProviders"
      :tracking-carrier-mappings="trackingCarrierMappings"
      :packaging-rules="packagingRules"
      :loading="loading"
      :can-create="hasPermission('shipping:create')"
      :can-edit="hasPermission('shipping:edit')"
      :can-delete="hasPermission('shipping:delete')"
      @create-template="showCreateTemplateDialog"
      @edit-template="showEditTemplateDialog"
      @create-mapping="showCreateTrackingCarrierMappingDialog"
      @edit-mapping="showEditTrackingCarrierMappingDialog"
      @create-carrier="showCreateCarrierDialog"
      @edit-carrier="showEditCarrierDialog"
      @create-carrier-service="showCreateCarrierServiceDialog"
      @edit-carrier-service="showEditCarrierServiceDialog"
      @create-packaging="showCreatePackagingDialog"
      @edit-packaging="showEditPackagingDialog"
      @show-packaging-applies="showPackagingAppliesDialog"
      @create-tracking-provider="showCreateTrackingProviderDialog"
      @edit-tracking-provider="showEditTrackingProviderDialog"
      @copy-webhook="copyTrackingWebhookUrl"
      @count-change="handleTrackingShipmentsCountChange"
      @delete="requestDelete"
    />

    <ShippingDialogsPanel
      v-model:template-open="templateDialogOpen"
      v-model:carrier-open="carrierDialogOpen"
      v-model:carrier-service-open="carrierServiceDialogOpen"
      v-model:tracking-provider-open="trackingProviderDialogOpen"
      v-model:tracking-carrier-mapping-open="trackingCarrierMappingDialogOpen"
      v-model:packaging-open="packagingDialogOpen"
      v-model:packaging-applies-open="packagingAppliesDialogOpen"
      v-model:delete-open="deleteDialogOpen"
      :templates="templates"
      :carriers="carriers"
      :carrier-services="carrierServices"
      :fpx-channels="fpxChannels"
      :yanwen-published-channels="yanwenPublishedChannels"
      :tracking-providers="trackingProviders"
      :template-mode="templateDialogMode"
      :template-form="templateForm"
      :template-errors="templateErrors"
      :template-submitting="templateSubmitting"
      :carrier-mode="carrierDialogMode"
      :carrier-form="carrierForm"
      :carrier-errors="carrierErrors"
      :carrier-submitting="carrierSubmitting"
      :carrier-service-mode="carrierServiceDialogMode"
      :carrier-service-form="carrierServiceForm"
      :carrier-service-errors="carrierServiceErrors"
      :carrier-service-submitting="carrierServiceSubmitting"
      :tracking-provider-mode="trackingProviderDialogMode"
      :tracking-provider-form="trackingProviderForm"
      :tracking-provider-errors="trackingProviderErrors"
      :tracking-provider-webhook-url="trackingWebhookUrl(trackingProviderForm)"
      :tracking-provider-submitting="trackingProviderSubmitting"
      :tracking-carrier-mapping-mode="trackingCarrierMappingDialogMode"
      :tracking-carrier-mapping-form="trackingCarrierMappingForm"
      :tracking-carrier-mapping-errors="trackingCarrierMappingErrors"
      :tracking-carrier-mapping-submitting="trackingCarrierMappingSubmitting"
      :packaging-mode="packagingDialogMode"
      :packaging-form="packagingForm"
      :packaging-errors="packagingErrors"
      :packaging-submitting="packagingSubmitting"
      :packaging-applies-rule="packagingAppliesRule"
      :delete-title="deleteDialogTitle"
      :delete-description="deleteDialogDescription"
      @save-template="saveTemplate"
      @save-carrier="saveCarrier"
      @save-carrier-service="saveCarrierService"
      @save-tracking-provider="saveTrackingProvider"
      @save-tracking-carrier-mapping="saveTrackingCarrierMapping"
      @save-packaging-rule="savePackagingRule"
      @clear-template-error="clearTemplateError"
      @clear-carrier-error="clearCarrierError"
      @clear-carrier-service-error="clearCarrierServiceError"
      @clear-tracking-provider-error="clearTrackingProviderError"
      @clear-tracking-carrier-mapping-error="clearTrackingCarrierMappingError"
      @clear-packaging-error="clearPackagingError"
      @packaging-applies-updated="fetchPackagingRules"
      @confirm-delete="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Calculator,
  CircleCheck,
  Link2,
  Radar,
  RefreshCw,
  Truck,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import ShippingDialogsPanel from '@/components/admin/shipping/ShippingDialogsPanel.vue'
import ShippingTabsPanel from '@/components/admin/shipping/ShippingTabsPanel.vue'
import { Button } from '@/components/ui/button'
import { useRouteTab } from '@/composables/useRouteTab'
import { useShippingCarrierManager } from '@/composables/shipping/useShippingCarrierManager'
import { useShippingDeleteManager } from '@/composables/shipping/useShippingDeleteManager'
import { useShippingPackagingManager } from '@/composables/shipping/useShippingPackagingManager'
import { useShippingResources } from '@/composables/shipping/useShippingResources'
import { useShippingTemplateManager } from '@/composables/shipping/useShippingTemplateManager'
import { useShippingTrackingManager } from '@/composables/shipping/useShippingTrackingManager'
import { trackingWebhookUrl } from '@/lib/shippingPresentation'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const activeTab = useRouteTab({
  defaultValue: 'templates',
  values: ['templates', 'carriers', 'services', 'packaging', 'tracking', 'trackingShipments'],
  routes: {
    templates: 'ShippingTemplates',
    carriers: 'ShippingCarriers',
    services: 'ShippingServices',
    packaging: 'ShippingPackaging',
    tracking: 'ShippingTracking',
    trackingShipments: 'ShippingTrackingShipments',
  },
})
const trackingShipmentsPanelRef = ref(null)
const {
  templates,
  carriers,
  carrierServices,
  fpxChannels,
  yanwenPublishedChannels,
  trackingProviders,
  trackingCarrierMappings,
  trackingShipmentsCount,
  packagingRules,
  refreshing,
  loading,
  handleTrackingShipmentsCountChange,
  fetchTemplates,
  fetchCarriers,
  fetchCarrierServices,
  fetchTrackingProviders,
  fetchTrackingCarrierMappings,
  fetchPackagingRules,
  refreshCurrentTab: refreshShippingTab,
  fetchAllShippingResources,
} = useShippingResources()

const {
  deleteDialogOpen,
  deleteDialogTitle,
  deleteDialogDescription,
  requestDelete,
  confirmDelete,
} = useShippingDeleteManager({
  fetchTemplates,
  fetchCarriers,
  fetchCarrierServices,
  fetchTrackingProviders,
  fetchTrackingCarrierMappings,
  fetchPackagingRules,
})

const {
  templateDialogOpen,
  templateDialogMode,
  templateSubmitting,
  templateErrors,
  templateForm,
  clearTemplateError,
  showCreateTemplateDialog,
  showEditTemplateDialog,
  saveTemplate,
} = useShippingTemplateManager({
  fetchTemplates,
  carrierServices,
  fetchCarrierServices,
})

const {
  packagingDialogOpen,
  packagingDialogMode,
  packagingSubmitting,
  packagingErrors,
  packagingForm,
  packagingAppliesDialogOpen,
  packagingAppliesRule,
  clearPackagingError,
  showCreatePackagingDialog,
  showEditPackagingDialog,
  showPackagingAppliesDialog,
  savePackagingRule,
} = useShippingPackagingManager({ fetchPackagingRules })

const {
  carrierDialogOpen,
  carrierDialogMode,
  carrierSubmitting,
  carrierErrors,
  carrierForm,
  carrierServiceDialogOpen,
  carrierServiceDialogMode,
  carrierServiceSubmitting,
  carrierServiceErrors,
  carrierServiceForm,
  clearCarrierError,
  clearCarrierServiceError,
  showCreateCarrierDialog,
  showEditCarrierDialog,
  saveCarrier,
  showCreateCarrierServiceDialog,
  showEditCarrierServiceDialog,
  saveCarrierService,
} = useShippingCarrierManager({
  carriers,
  fetchCarriers,
  fetchCarrierServices,
})

const {
  trackingProviderDialogOpen,
  trackingProviderDialogMode,
  trackingProviderSubmitting,
  trackingProviderErrors,
  trackingProviderForm,
  trackingCarrierMappingDialogOpen,
  trackingCarrierMappingDialogMode,
  trackingCarrierMappingSubmitting,
  trackingCarrierMappingErrors,
  trackingCarrierMappingForm,
  clearTrackingProviderError,
  clearTrackingCarrierMappingError,
  showCreateTrackingProviderDialog,
  showEditTrackingProviderDialog,
  saveTrackingProvider,
  showCreateTrackingCarrierMappingDialog,
  showEditTrackingCarrierMappingDialog,
  saveTrackingCarrierMapping,
} = useShippingTrackingManager({
  trackingProviders,
  carriers,
  carrierServices,
  fetchTrackingProviders,
  fetchTrackingCarrierMappings,
})

const hasPermission = (permission: string) => authStore.hasPermission(permission)

const refreshCurrentTab = () => refreshShippingTab(activeTab.value, trackingShipmentsPanelRef)

const statItems = computed(() => [
  {
    key: 'templates',
    label: '运费模板',
    value: templates.value.length,
    icon: Calculator,
    tone: 'blue',
  },
  {
    key: 'carriers',
    label: '承运商',
    value: carriers.value.length,
    icon: Truck,
    tone: 'blue',
  },
  {
    key: 'trackingProviders',
    label: '追踪配置',
    value: trackingProviders.value.length,
    icon: Radar,
    tone: 'amber',
  },
  {
    key: 'trackingMappings',
    label: '承运商映射',
    value: trackingCarrierMappings.value.length,
    icon: Link2,
    tone: 'blue',
  },
  {
    key: 'trackingShipments',
    label: '追踪任务',
    value: trackingShipmentsCount.value,
    icon: RefreshCw,
    tone: 'amber',
  },
  {
    key: 'activePackaging',
    label: '启用包装规则',
    value: packagingRules.value.filter((rule) => rule.is_active).length,
    icon: CircleCheck,
    tone: 'green',
  },
])

const copyTrackingWebhookUrl = async (provider: any) => {
  const url = trackingWebhookUrl(provider)
  if (!url) return

  try {
    await navigator.clipboard.writeText(url)
    toast.success('Webhook 地址已复制')
  } catch (error) {
    console.error('Failed to copy tracking webhook URL:', error)
    toast.error('复制失败，请手动复制')
  }
}

onMounted(fetchAllShippingResources)
</script>

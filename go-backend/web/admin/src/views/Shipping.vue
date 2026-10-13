<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="activeShippingManagementTabPageHeader.title"
      :description="activeShippingManagementTabPageHeader.description"
    >
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
      :route-aggregates="routeAggregates"
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
      @create-route-aggregate="showCreateRouteAggregateDialog"
      @edit-route-aggregate="showEditRouteAggregateDialog"
      @create-template="showCreateTemplateDialog"
      @edit-template="showEditTemplateDialog"
      @create-mapping="showCreateTrackingCarrierMappingDialog"
      @edit-mapping="showEditTrackingCarrierMappingDialog"
      @create-carrier="showCreateCarrierDialog"
      @edit-carrier="showEditCarrierDialog"
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
      :route-aggregates="routeAggregates"
      v-model:route-aggregate-open="routeAggregateDialogOpen"
      :route-aggregate-mode="routeAggregateDialogMode" :route-aggregate-form="routeAggregateForm"
      :route-aggregate-errors="routeAggregateErrors" :route-aggregate-submitting="routeAggregateSubmitting"
      @save-route-aggregate="saveRouteAggregate" @clear-route-aggregate-error="clearRouteAggregateError"
      v-model:template-open="templateDialogOpen"
      v-model:carrier-open="carrierDialogOpen"
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
      @save-tracking-provider="saveTrackingProvider"
      @save-tracking-carrier-mapping="saveTrackingCarrierMapping"
      @save-packaging-rule="savePackagingRule"
      @clear-template-error="clearTemplateError"
      @clear-carrier-error="clearCarrierError"
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
import { useShippingRouteAggregateManager } from '@/composables/shipping/useShippingRouteAggregateManager'
import { useShippingTemplateManager } from '@/composables/shipping/useShippingTemplateManager'
import { useShippingTrackingManager } from '@/composables/shipping/useShippingTrackingManager'
import { trackingWebhookUrl } from '@/lib/shippingPresentation'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const shippingManagementTabPageHeaders = {
  routeAggregates: {
    title: '聚合线路',
    description: '管理承运商已发布的线路、聚合关系和覆盖国家。',
  },
  templates: {
    title: '线路计费模板',
    description: '设置向客户收取的运费、首续重、附加费及免邮条件。',
  },
  carriers: {
    title: '承运商',
    description: '维护物流公司的基础资料、官方查询链接和服务区域。',
  },
  packaging: {
    title: '包装规则',
    description: '管理包装箱重量、尺寸、最大承重和适用商品。',
  },
  tracking: {
    title: '追踪配置',
    description: '配置追踪服务的 API 凭证、Webhook、同步策略和承运商映射。',
  },
  trackingShipments: {
    title: '追踪任务',
    description: '查看运单追踪状态、同步进度和异常，处理到期同步任务。',
  },
}
const activeTab = useRouteTab<keyof typeof shippingManagementTabPageHeaders>({
  defaultValue: 'routeAggregates',
  values: ['routeAggregates', 'templates', 'carriers', 'packaging', 'tracking', 'trackingShipments'],
  routes: {
    routeAggregates: 'ShippingRouteAggregates',
    templates: 'ShippingCustomerBillingTemplates',
    carriers: 'ShippingCarriers',
    packaging: 'ShippingPackaging',
    tracking: 'ShippingTracking',
    trackingShipments: 'ShippingTrackingShipments',
  },
})
const activeShippingManagementTabPageHeader = computed(() => (
  shippingManagementTabPageHeaders[activeTab.value]
))
const trackingShipmentsPanelRef = ref(null)
const {
  routeAggregates,
  fetchRouteAggregates,
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
  fetchRouteAggregates,
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
} = useShippingTemplateManager({ fetchTemplates })

const { routeAggregateDialogOpen, routeAggregateDialogMode, routeAggregateSubmitting, routeAggregateErrors, routeAggregateForm, clearRouteAggregateError, showCreateRouteAggregateDialog, showEditRouteAggregateDialog, saveRouteAggregate } = useShippingRouteAggregateManager({ fetchRouteAggregates })

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
  clearCarrierError,
  showCreateCarrierDialog,
  showEditCarrierDialog,
  saveCarrier,
} = useShippingCarrierManager({
  fetchCarriers,
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
  { key: 'routeAggregates', label: '聚合线路', value: routeAggregates.value.length, icon: Truck, tone: 'blue' },
  {
    key: 'templates',
    label: '线路计费模板',
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

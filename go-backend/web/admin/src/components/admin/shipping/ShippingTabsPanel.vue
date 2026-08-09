<template>
  <Tabs :model-value="activeTab" class="gap-4">
    <ShippingTemplatesPanel
      :templates="templates"
      :tracking-carrier-mappings="trackingCarrierMappings"
      :tracking-providers="trackingProviders"
      :carriers="carriers"
      :carrier-services="carrierServices"
      :loading-templates="loading.templates"
      :loading-tracking-mappings="loading.trackingMappings"
      :can-create="canCreate"
      :can-edit="canEdit"
      :can-delete="canDelete"
      @create-template="emit('create-template')"
      @edit-template="emit('edit-template', $event)"
      @delete-template="emit('delete', 'template', $event)"
      @create-mapping="emit('create-mapping')"
      @edit-mapping="emit('edit-mapping', $event)"
      @delete-mapping="emit('delete', 'trackingCarrierMapping', $event)"
    />

    <ShippingZonesPanel
      :zones="zones"
      :loading="loading.zones"
      :can-create="canCreate"
      :can-edit="canEdit"
      :can-delete="canDelete"
      @create="emit('create-zone')"
      @edit="emit('edit-zone', $event)"
      @delete="emit('delete', 'zone', $event)"
    />

    <ShippingCarriersPanel
      :carriers="carriers"
      :loading="loading.carriers"
      :can-create="canCreate"
      :can-edit="canEdit"
      :can-delete="canDelete"
      @create="emit('create-carrier')"
      @edit="emit('edit-carrier', $event)"
      @delete="emit('delete', 'carrier', $event)"
    />

    <ShippingCarrierServicesPanel
      :carrier-services="carrierServices"
      :carriers="carriers"
      :templates="templates"
      :loading="loading.services"
      :can-create="canCreate"
      :can-edit="canEdit"
      :can-delete="canDelete"
      @create="emit('create-carrier-service')"
      @edit="emit('edit-carrier-service', $event)"
      @delete="emit('delete', 'carrierService', $event)"
    />

    <TabsContent value="quote" class="space-y-3">
      <ShippingQuoteCalculator />
    </TabsContent>

    <ShippingPackagingPanel
      :packaging-rules="packagingRules"
      :loading="loading.packaging"
      :can-create="canCreate"
      :can-edit="canEdit"
      :can-delete="canDelete"
      @create="emit('create-packaging')"
      @edit="emit('edit-packaging', $event)"
      @delete="emit('delete', 'packaging', $event)"
      @show-applies="emit('show-packaging-applies', $event)"
    />

    <TrackingProvidersPanel
      :tracking-providers="trackingProviders"
      :loading="loading.tracking"
      :can-create="canCreate"
      :can-edit="canEdit"
      :can-delete="canDelete"
      @create="emit('create-tracking-provider')"
      @edit="emit('edit-tracking-provider', $event)"
      @delete="emit('delete', 'trackingProvider', $event)"
      @copy-webhook="emit('copy-webhook', $event)"
    />

    <TabsContent value="trackingShipments">
      <TrackingShipmentsPanel
        ref="trackingShipmentsPanelRef"
        :tracking-providers="trackingProviders"
        :carriers="carriers"
        :carrier-services="carrierServices"
        :can-edit="canEdit"
        @count-change="emit('count-change', $event)"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ShippingCarrierServicesPanel from '@/components/admin/shipping/ShippingCarrierServicesPanel.vue'
import ShippingCarriersPanel from '@/components/admin/shipping/ShippingCarriersPanel.vue'
import ShippingPackagingPanel from '@/components/admin/shipping/ShippingPackagingPanel.vue'
import ShippingQuoteCalculator from '@/components/admin/shipping/ShippingQuoteCalculator.vue'
import ShippingTemplatesPanel from '@/components/admin/shipping/ShippingTemplatesPanel.vue'
import ShippingZonesPanel from '@/components/admin/shipping/ShippingZonesPanel.vue'
import TrackingProvidersPanel from '@/components/admin/shipping/TrackingProvidersPanel.vue'
import TrackingShipmentsPanel from '@/components/admin/shipping/TrackingShipmentsPanel.vue'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import type {
  PackagingRule,
  ShippingCarrier,
  ShippingCarrierService,
  ShippingLoadingState,
  ShippingResource,
  ShippingTemplate,
  ShippingZone,
  TrackingCarrierMapping,
  TrackingProvider
} from './shippingTypes'

withDefaults(defineProps<{
  activeTab?: string
  templates?: ShippingTemplate[]
  zones?: ShippingZone[]
  carriers?: ShippingCarrier[]
  carrierServices?: ShippingCarrierService[]
  trackingProviders?: TrackingProvider[]
  trackingCarrierMappings?: TrackingCarrierMapping[]
  packagingRules?: PackagingRule[]
  loading?: ShippingLoadingState
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  activeTab: 'templates',
  templates: () => [],
  zones: () => [],
  carriers: () => [],
  carrierServices: () => [],
  trackingProviders: () => [],
  trackingCarrierMappings: () => [],
  packagingRules: () => [],
  loading: () => ({
    templates: false,
    zones: false,
    carriers: false,
    services: false,
    tracking: false,
    trackingMappings: false,
    trackingShipments: false,
    packaging: false
  }),
  canCreate: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'create-template'): void
  (event: 'edit-template', template: ShippingTemplate): void
  (event: 'create-mapping'): void
  (event: 'edit-mapping', mapping: TrackingCarrierMapping): void
  (event: 'create-zone'): void
  (event: 'edit-zone', zone: ShippingZone): void
  (event: 'create-carrier'): void
  (event: 'edit-carrier', carrier: ShippingCarrier): void
  (event: 'create-carrier-service'): void
  (event: 'edit-carrier-service', service: ShippingCarrierService): void
  (event: 'create-packaging'): void
  (event: 'edit-packaging', rule: PackagingRule): void
  (event: 'show-packaging-applies', rule: PackagingRule): void
  (event: 'create-tracking-provider'): void
  (event: 'edit-tracking-provider', provider: TrackingProvider): void
  (event: 'copy-webhook', provider: TrackingProvider): void
  (event: 'count-change', count: number): void
  (event: 'delete', resourceType: string, resource: ShippingResource): void
}>()

const trackingShipmentsPanelRef = ref<{ refresh?: () => Promise<void> | void } | null>(null)
const refresh = () => trackingShipmentsPanelRef.value?.refresh?.()

defineExpose({ refresh })
</script>

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

<script setup>
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

defineProps({
  activeTab: { type: String, default: 'templates' },
  templates: { type: Array, default: () => [] },
  zones: { type: Array, default: () => [] },
  carriers: { type: Array, default: () => [] },
  carrierServices: { type: Array, default: () => [] },
  trackingProviders: { type: Array, default: () => [] },
  trackingCarrierMappings: { type: Array, default: () => [] },
  packagingRules: { type: Array, default: () => [] },
  loading: { type: Object, default: () => ({}) },
  canCreate: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
})

const emit = defineEmits([
  'create-template',
  'edit-template',
  'create-mapping',
  'edit-mapping',
  'create-zone',
  'edit-zone',
  'create-carrier',
  'edit-carrier',
  'create-carrier-service',
  'edit-carrier-service',
  'create-packaging',
  'edit-packaging',
  'show-packaging-applies',
  'create-tracking-provider',
  'edit-tracking-provider',
  'copy-webhook',
  'count-change',
  'delete',
])

const trackingShipmentsPanelRef = ref(null)
const refresh = () => trackingShipmentsPanelRef.value?.refresh?.()

defineExpose({ refresh })
</script>

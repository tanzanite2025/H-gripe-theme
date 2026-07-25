<template>
  <ShippingTemplateEditorDialog
    :open="templateOpen"
    :mode="templateMode"
    :form="templateForm"
    :errors="templateErrors"
    :submitting="templateSubmitting"
    @update:open="emit('update:templateOpen', $event)"
    @submit="emit('save-template')"
    @clear-error="emit('clear-template-error', $event)"
  />

  <ShippingZoneEditorDialog
    :open="zoneOpen"
    :mode="zoneMode"
    :form="zoneForm"
    :errors="zoneErrors"
    :submitting="zoneSubmitting"
    @update:open="emit('update:zoneOpen', $event)"
    @submit="emit('save-zone')"
    @clear-error="emit('clear-zone-error', $event)"
  />

  <ShippingTemplateBindingEditorDialog
    :open="bindingOpen"
    :mode="bindingMode"
    :form="bindingForm"
    :errors="bindingErrors"
    :templates="templates"
    :submitting="bindingSubmitting"
    @update:open="emit('update:bindingOpen', $event)"
    @submit="emit('save-binding')"
    @clear-error="emit('clear-binding-error', $event)"
  />

  <CarrierEditorDialog
    :open="carrierOpen"
    :mode="carrierMode"
    :form="carrierForm"
    :errors="carrierErrors"
    :submitting="carrierSubmitting"
    @update:open="emit('update:carrierOpen', $event)"
    @submit="emit('save-carrier')"
    @clear-error="emit('clear-carrier-error', $event)"
  />

  <CarrierServiceEditorDialog
    :open="carrierServiceOpen"
    :mode="carrierServiceMode"
    :form="carrierServiceForm"
    :errors="carrierServiceErrors"
    :carriers="carriers"
    :templates="templates"
    :submitting="carrierServiceSubmitting"
    @update:open="emit('update:carrierServiceOpen', $event)"
    @submit="emit('save-carrier-service')"
    @clear-error="emit('clear-carrier-service-error', $event)"
  />

  <TrackingProviderEditorDialog
    :open="trackingProviderOpen"
    :mode="trackingProviderMode"
    :form="trackingProviderForm"
    :errors="trackingProviderErrors"
    :webhook-url="trackingProviderWebhookUrl"
    :submitting="trackingProviderSubmitting"
    @update:open="emit('update:trackingProviderOpen', $event)"
    @submit="emit('save-tracking-provider')"
    @clear-error="emit('clear-tracking-provider-error', $event)"
  />

  <TrackingCarrierMappingEditorDialog
    :open="trackingCarrierMappingOpen"
    :mode="trackingCarrierMappingMode"
    :form="trackingCarrierMappingForm"
    :errors="trackingCarrierMappingErrors"
    :providers="trackingProviders"
    :carriers="carriers"
    :carrier-services="carrierServices"
    :submitting="trackingCarrierMappingSubmitting"
    @update:open="emit('update:trackingCarrierMappingOpen', $event)"
    @submit="emit('save-tracking-carrier-mapping')"
    @clear-error="emit('clear-tracking-carrier-mapping-error', $event)"
  />

  <PackagingRuleEditorDialog
    :open="packagingOpen"
    :mode="packagingMode"
    :form="packagingForm"
    :errors="packagingErrors"
    :submitting="packagingSubmitting"
    @update:open="emit('update:packagingOpen', $event)"
    @submit="emit('save-packaging-rule')"
    @clear-error="emit('clear-packaging-error', $event)"
  />

  <PackagingRuleAppliesDialog
    :open="packagingAppliesOpen"
    :rule="packagingAppliesRule"
    @update:open="emit('update:packagingAppliesOpen', $event)"
    @updated="emit('packaging-applies-updated')"
  />

  <AdminConfirmDialog
    :open="deleteOpen"
    :title="deleteTitle"
    :description="deleteDescription"
    confirm-label="确认删除"
    destructive
    @update:open="emit('update:deleteOpen', $event)"
    @confirm="emit('confirm-delete')"
  />
</template>

<script setup>
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import CarrierEditorDialog from '@/components/admin/shipping/CarrierEditorDialog.vue'
import CarrierServiceEditorDialog from '@/components/admin/shipping/CarrierServiceEditorDialog.vue'
import PackagingRuleAppliesDialog from '@/components/admin/shipping/PackagingRuleAppliesDialog.vue'
import PackagingRuleEditorDialog from '@/components/admin/shipping/PackagingRuleEditorDialog.vue'
import ShippingTemplateBindingEditorDialog from '@/components/admin/shipping/ShippingTemplateBindingEditorDialog.vue'
import ShippingTemplateEditorDialog from '@/components/admin/shipping/ShippingTemplateEditorDialog.vue'
import ShippingZoneEditorDialog from '@/components/admin/shipping/ShippingZoneEditorDialog.vue'
import TrackingCarrierMappingEditorDialog from '@/components/admin/shipping/TrackingCarrierMappingEditorDialog.vue'
import TrackingProviderEditorDialog from '@/components/admin/shipping/TrackingProviderEditorDialog.vue'

defineProps({
  templates: { type: Array, default: () => [] },
  carriers: { type: Array, default: () => [] },
  carrierServices: { type: Array, default: () => [] },
  trackingProviders: { type: Array, default: () => [] },

  templateOpen: { type: Boolean, default: false },
  templateMode: { type: String, default: 'create' },
  templateSubmitting: { type: Boolean, default: false },
  templateErrors: { type: Object, required: true },
  templateForm: { type: Object, required: true },

  zoneOpen: { type: Boolean, default: false },
  zoneMode: { type: String, default: 'create' },
  zoneSubmitting: { type: Boolean, default: false },
  zoneErrors: { type: Object, required: true },
  zoneForm: { type: Object, required: true },

  bindingOpen: { type: Boolean, default: false },
  bindingMode: { type: String, default: 'create' },
  bindingSubmitting: { type: Boolean, default: false },
  bindingErrors: { type: Object, required: true },
  bindingForm: { type: Object, required: true },

  carrierOpen: { type: Boolean, default: false },
  carrierMode: { type: String, default: 'create' },
  carrierSubmitting: { type: Boolean, default: false },
  carrierErrors: { type: Object, required: true },
  carrierForm: { type: Object, required: true },

  carrierServiceOpen: { type: Boolean, default: false },
  carrierServiceMode: { type: String, default: 'create' },
  carrierServiceSubmitting: { type: Boolean, default: false },
  carrierServiceErrors: { type: Object, required: true },
  carrierServiceForm: { type: Object, required: true },

  trackingProviderOpen: { type: Boolean, default: false },
  trackingProviderMode: { type: String, default: 'create' },
  trackingProviderSubmitting: { type: Boolean, default: false },
  trackingProviderErrors: { type: Object, required: true },
  trackingProviderForm: { type: Object, required: true },
  trackingProviderWebhookUrl: { type: String, default: '' },

  trackingCarrierMappingOpen: { type: Boolean, default: false },
  trackingCarrierMappingMode: { type: String, default: 'create' },
  trackingCarrierMappingSubmitting: { type: Boolean, default: false },
  trackingCarrierMappingErrors: { type: Object, required: true },
  trackingCarrierMappingForm: { type: Object, required: true },

  packagingOpen: { type: Boolean, default: false },
  packagingMode: { type: String, default: 'create' },
  packagingSubmitting: { type: Boolean, default: false },
  packagingErrors: { type: Object, required: true },
  packagingForm: { type: Object, required: true },

  packagingAppliesOpen: { type: Boolean, default: false },
  packagingAppliesRule: { type: Object, default: null },

  deleteOpen: { type: Boolean, default: false },
  deleteTitle: { type: String, default: '' },
  deleteDescription: { type: String, default: '' },
})

const emit = defineEmits([
  'update:templateOpen',
  'update:zoneOpen',
  'update:bindingOpen',
  'update:carrierOpen',
  'update:carrierServiceOpen',
  'update:trackingProviderOpen',
  'update:trackingCarrierMappingOpen',
  'update:packagingOpen',
  'update:packagingAppliesOpen',
  'update:deleteOpen',
  'save-template',
  'save-zone',
  'save-binding',
  'save-carrier',
  'save-carrier-service',
  'save-tracking-provider',
  'save-tracking-carrier-mapping',
  'save-packaging-rule',
  'clear-template-error',
  'clear-zone-error',
  'clear-binding-error',
  'clear-carrier-error',
  'clear-carrier-service-error',
  'clear-tracking-provider-error',
  'clear-tracking-carrier-mapping-error',
  'clear-packaging-error',
  'packaging-applies-updated',
  'confirm-delete',
])
</script>

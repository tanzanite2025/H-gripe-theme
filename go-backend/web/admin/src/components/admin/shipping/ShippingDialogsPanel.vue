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

<script setup lang="ts">
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import CarrierEditorDialog from '@/components/admin/shipping/CarrierEditorDialog.vue'
import CarrierServiceEditorDialog from '@/components/admin/shipping/CarrierServiceEditorDialog.vue'
import PackagingRuleAppliesDialog from '@/components/admin/shipping/PackagingRuleAppliesDialog.vue'
import PackagingRuleEditorDialog from '@/components/admin/shipping/PackagingRuleEditorDialog.vue'
import ShippingTemplateEditorDialog from '@/components/admin/shipping/ShippingTemplateEditorDialog.vue'
import ShippingZoneEditorDialog from '@/components/admin/shipping/ShippingZoneEditorDialog.vue'
import TrackingCarrierMappingEditorDialog from '@/components/admin/shipping/TrackingCarrierMappingEditorDialog.vue'
import TrackingProviderEditorDialog from '@/components/admin/shipping/TrackingProviderEditorDialog.vue'
import type {
  ShippingCarrier,
  ShippingCarrierForm,
  ShippingCarrierService,
  ShippingCarrierServiceForm,
  ShippingDialogMode,
  ShippingErrorMap,
  ShippingTemplate,
  ShippingTemplateForm,
  ShippingZone,
  ShippingZoneForm,
  PackagingRule,
  PackagingRuleForm,
  TrackingCarrierMapping,
  TrackingCarrierMappingForm,
  TrackingProvider,
  TrackingProviderForm
} from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  templates?: ShippingTemplate[]
  carriers?: ShippingCarrier[]
  carrierServices?: ShippingCarrierService[]
  trackingProviders?: TrackingProvider[]
  templateOpen?: boolean
  templateMode?: ShippingDialogMode
  templateSubmitting?: boolean
  templateErrors: ShippingErrorMap
  templateForm: ShippingTemplateForm
  zoneOpen?: boolean
  zoneMode?: ShippingDialogMode
  zoneSubmitting?: boolean
  zoneErrors: ShippingErrorMap
  zoneForm: ShippingZoneForm
  carrierOpen?: boolean
  carrierMode?: ShippingDialogMode
  carrierSubmitting?: boolean
  carrierErrors: ShippingErrorMap
  carrierForm: ShippingCarrierForm
  carrierServiceOpen?: boolean
  carrierServiceMode?: ShippingDialogMode
  carrierServiceSubmitting?: boolean
  carrierServiceErrors: ShippingErrorMap
  carrierServiceForm: ShippingCarrierServiceForm
  trackingProviderOpen?: boolean
  trackingProviderMode?: ShippingDialogMode
  trackingProviderSubmitting?: boolean
  trackingProviderErrors: ShippingErrorMap
  trackingProviderForm: TrackingProviderForm
  trackingProviderWebhookUrl?: string
  trackingCarrierMappingOpen?: boolean
  trackingCarrierMappingMode?: ShippingDialogMode
  trackingCarrierMappingSubmitting?: boolean
  trackingCarrierMappingErrors: ShippingErrorMap
  trackingCarrierMappingForm: TrackingCarrierMappingForm
  packagingOpen?: boolean
  packagingMode?: ShippingDialogMode
  packagingSubmitting?: boolean
  packagingErrors: ShippingErrorMap
  packagingForm: PackagingRuleForm
  packagingAppliesOpen?: boolean
  packagingAppliesRule?: PackagingRule | null
  deleteOpen?: boolean
  deleteTitle?: string
  deleteDescription?: string
}>(), {
  templates: () => [],
  carriers: () => [],
  carrierServices: () => [],
  trackingProviders: () => [],
  templateOpen: false,
  templateMode: 'create',
  templateSubmitting: false,
  zoneOpen: false,
  zoneMode: 'create',
  zoneSubmitting: false,
  carrierOpen: false,
  carrierMode: 'create',
  carrierSubmitting: false,
  carrierServiceOpen: false,
  carrierServiceMode: 'create',
  carrierServiceSubmitting: false,
  trackingProviderOpen: false,
  trackingProviderMode: 'create',
  trackingProviderSubmitting: false,
  trackingProviderWebhookUrl: '',
  trackingCarrierMappingOpen: false,
  trackingCarrierMappingMode: 'create',
  trackingCarrierMappingSubmitting: false,
  packagingOpen: false,
  packagingMode: 'create',
  packagingSubmitting: false,
  packagingAppliesOpen: false,
  packagingAppliesRule: null,
  deleteOpen: false,
  deleteTitle: '',
  deleteDescription: '',
})

const emit = defineEmits<{
  (event: 'update:templateOpen', value: boolean): void
  (event: 'update:zoneOpen', value: boolean): void
  (event: 'update:carrierOpen', value: boolean): void
  (event: 'update:carrierServiceOpen', value: boolean): void
  (event: 'update:trackingProviderOpen', value: boolean): void
  (event: 'update:trackingCarrierMappingOpen', value: boolean): void
  (event: 'update:packagingOpen', value: boolean): void
  (event: 'update:packagingAppliesOpen', value: boolean): void
  (event: 'update:deleteOpen', value: boolean): void
  (event: 'save-template'): void
  (event: 'save-zone'): void
  (event: 'save-carrier'): void
  (event: 'save-carrier-service'): void
  (event: 'save-tracking-provider'): void
  (event: 'save-tracking-carrier-mapping'): void
  (event: 'save-packaging-rule'): void
  (event: 'clear-template-error', field: unknown): void
  (event: 'clear-zone-error', field: unknown): void
  (event: 'clear-carrier-error', field: unknown): void
  (event: 'clear-carrier-service-error', field: unknown): void
  (event: 'clear-tracking-provider-error', field: unknown): void
  (event: 'clear-tracking-carrier-mapping-error', field: unknown): void
  (event: 'clear-packaging-error', field: unknown): void
  (event: 'packaging-applies-updated'): void
  (event: 'confirm-delete'): void
}>()
</script>


<template>
  <ShippingTemplateEditorDialog
    :open="templateOpen"
    :mode="templateMode"
    :form="templateForm"
    :errors="templateErrors"
    :submitting="templateSubmitting"
    :route-aggregates="routeAggregates"
    @update:open="emit('update:templateOpen', $event)"
    @submit="emit('save-template')"
    @clear-error="emit('clear-template-error', $event)"
  />

  <ShippingRouteAggregateEditorDialog
    :open="routeAggregateOpen" :mode="routeAggregateMode" :form="routeAggregateForm" :errors="routeAggregateErrors"
    :carrier-services="carrierServices" :carriers="carriers" :fpx-channels="fpxChannels" :yanwen-published-channels="yanwenPublishedChannels"
    :submitting="routeAggregateSubmitting" @update:open="emit('update:routeAggregateOpen', $event)"
    @submit="emit('save-route-aggregate')" @clear-error="emit('clear-route-aggregate-error', $event)"
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
import ShippingRouteAggregateEditorDialog from '@/components/admin/shipping/ShippingRouteAggregateEditorDialog.vue'
import PackagingRuleAppliesDialog from '@/components/admin/shipping/PackagingRuleAppliesDialog.vue'
import PackagingRuleEditorDialog from '@/components/admin/shipping/PackagingRuleEditorDialog.vue'
import ShippingTemplateEditorDialog from '@/components/admin/shipping/ShippingTemplateEditorDialog.vue'
import TrackingCarrierMappingEditorDialog from '@/components/admin/shipping/TrackingCarrierMappingEditorDialog.vue'
import TrackingProviderEditorDialog from '@/components/admin/shipping/TrackingProviderEditorDialog.vue'
import type {
  FpxPublishedCollectionReference,
  YanwenPublishedCollectionReference,
} from '@/api/shippingServiceCollectionReferenceApi'
import type {
  ShippingCarrier,
  ShippingCarrierForm,
  ShippingCarrierService,
  ShippingDialogMode,
  ShippingErrorMap,
  ShippingTemplate,
  ShippingRouteAggregate,
  ShippingRouteAggregateForm,
  ShippingTemplateForm,
  PackagingRule,
  PackagingRuleForm,
  TrackingCarrierMapping,
  TrackingCarrierMappingForm,
  TrackingProvider,
  TrackingProviderForm
} from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  templates?: ShippingTemplate[]
  routeAggregates?: ShippingRouteAggregate[]
  routeAggregateOpen?: boolean
  routeAggregateMode?: ShippingDialogMode
  routeAggregateSubmitting?: boolean
  routeAggregateForm: ShippingRouteAggregateForm
  routeAggregateErrors: ShippingErrorMap
  carriers?: ShippingCarrier[]
  fpxChannels?: FpxPublishedCollectionReference[]
  yanwenPublishedChannels?: YanwenPublishedCollectionReference[]
  carrierServices?: ShippingCarrierService[]
  trackingProviders?: TrackingProvider[]
  templateOpen?: boolean
  templateMode?: ShippingDialogMode
  templateSubmitting?: boolean
  templateErrors: ShippingErrorMap
  templateForm: ShippingTemplateForm
  carrierOpen?: boolean
  carrierMode?: ShippingDialogMode
  carrierSubmitting?: boolean
  carrierErrors: ShippingErrorMap
  carrierForm: ShippingCarrierForm
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
  routeAggregates: () => [],
  routeAggregateOpen: false,
  routeAggregateMode: 'create',
  routeAggregateSubmitting: false,
  carriers: () => [],
  fpxChannels: () => [],
  yanwenPublishedChannels: () => [],
  carrierServices: () => [],
  trackingProviders: () => [],
  templateOpen: false,
  templateMode: 'create',
  templateSubmitting: false,
  carrierOpen: false,
  carrierMode: 'create',
  carrierSubmitting: false,
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
  (event: 'update:routeAggregateOpen', value: boolean): void
  (event: 'save-route-aggregate'): void
  (event: 'clear-route-aggregate-error', field: unknown): void
  (event: 'update:templateOpen', value: boolean): void
  (event: 'update:carrierOpen', value: boolean): void
  (event: 'update:trackingProviderOpen', value: boolean): void
  (event: 'update:trackingCarrierMappingOpen', value: boolean): void
  (event: 'update:packagingOpen', value: boolean): void
  (event: 'update:packagingAppliesOpen', value: boolean): void
  (event: 'update:deleteOpen', value: boolean): void
  (event: 'save-template'): void
  (event: 'save-carrier'): void
  (event: 'save-tracking-provider'): void
  (event: 'save-tracking-carrier-mapping'): void
  (event: 'save-packaging-rule'): void
  (event: 'clear-template-error', field: unknown): void
  (event: 'clear-carrier-error', field: unknown): void
  (event: 'clear-tracking-provider-error', field: unknown): void
  (event: 'clear-tracking-carrier-mapping-error', field: unknown): void
  (event: 'clear-packaging-error', field: unknown): void
  (event: 'packaging-applies-updated'): void
  (event: 'confirm-delete'): void
}>()
</script>

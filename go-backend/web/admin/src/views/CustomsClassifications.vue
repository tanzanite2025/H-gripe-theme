<template>
  <div class="space-y-4">
    <AdminPageHeader title="清关资料中心" description="集中检查商品清关资料完整度，并维护可复用的分类资料模板。">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">商品管理</RouterLink>
        </Button>
        <Button variant="outline" @click="lookupDialogOpen = true">
          <FileSearch class="size-4" />
          编码查询
        </Button>
        <Button v-if="hasPermission('product:create')" @click="openTemplateCreate">
          <Plus class="size-4" />
          新建清关模板
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <CustomsProductCoveragePanel
      :product-rows="productRows"
      :product-loading="productLoading"
      :product-page="productPage"
      :product-page-size="productPageSize"
      :product-total="productTotal"
      :filters="productFilters"
      :product-spec-templates="productSpecTemplates"
      @refresh="refreshWorkbench"
      @apply="applyProductFilters"
      @update:page="updateProductPage"
      @update:page-size="updateProductPageSize"
      @edit-product="openProductEditor"
    />

    <CustomsTemplatePanel
      :templates="templates"
      :loading="templateLoading"
      :can-create="hasPermission('product:create')"
      :can-edit="hasPermission('product:edit')"
      :can-delete="hasPermission('product:delete')"
      @create="openTemplateCreate"
      @edit="openTemplateEdit"
      @delete="removeTemplate"
    />

    <CustomsLookupDialog
      :open="lookupDialogOpen"
      :provider="lookupProvider"
      :query="lookupQuery"
      :loading="lookupLoading"
      :completed="lookupCompleted"
      :candidates="lookupCandidates"
      :can-create="hasPermission('product:create')"
      @update:open="lookupDialogOpen = $event"
      @update:provider="lookupProvider = $event"
      @update:query="lookupQuery = $event"
      @run="runLookup"
      @create-template="handleCreateTemplateFromCandidate"
    />

    <CustomsClassificationEditorDialog
      :open="templateDialogOpen"
      :form="templateForm"
      :saving="templateSaving"
      :product-spec-templates="productSpecTemplates"
      @update:open="templateDialogOpen = $event"
      @update:product-spec-template="setTemplateProductSpecTemplate"
      @save="saveTemplate"
    />

    <ProductEditorDialog
      v-model:open="dialogVisible"
      :mode="dialogMode"
      :submitting="submitting"
      :form="productForm"
      :errors="formErrors"
      :product-spec-templates="productSpecTemplates"
      :brands="brands"
      :selected-product-spec-template="selectedProductSpecTemplate"
      :brand-select-value="brandSelectValue"
      :selected-spec-definitions="selectedSpecDefinitions"
      :variant-spec-definitions="variantSpecDefinitions"
      :default-variant-index="defaultVariantIndex"
      :product-spec-template-select-value="productSpecTemplateSelectValue"
      :shipping-template-select-value="shippingTemplateSelectValue"
      :shipping-templates="shippingTemplates"
      :after-sales-template-select-value="afterSalesTemplateSelectValue"
      :packaging-template-select-value="packagingTemplateSelectValue"
      :after-sales-templates="editorAfterSalesTemplates"
      :packaging-templates="editorPackagingTemplates"
      :customs-classifications="activeCustomsClassifications"
      :customs-classification-select-value="customsClassificationSelectValue"
      :template-scoped-values-touched="templateScopedValuesTouched"
      :uploading-media="uploadingMedia"
      :parse-spec-options="parseSpecOptions"
      :format-spec-option="formatSpecOption"
      :get-spec-label="getSpecLabel"
      :spec-select-value="specSelectValue"
      :language-options="languageOptions"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @product-spec-template-select="handleProductSpecTemplateSelect"
      @product-brand-select="setProductBrand"
      @product-shipping-template-select="setProductShippingTemplate"
      @product-information-template-select="setProductInformationTemplate"
      @customs-classification-select="handleCustomsClassificationSelect"
      @customs-classification-manual-edit="clearCustomsClassification"
      @set-spec-select-value="setSpecSelectValue"
      @add-variant="addVariant"
      @remove-variant="removeVariant"
      @set-default-variant="setDefaultVariant"
      @set-variant-active="setVariantActive"
      @upload-media="handleMediaUpload"
      @add-media-url="addMediaUrl"
      @set-primary-media="setPrimaryMedia"
      @move-media="moveMedia"
      @remove-media="removeMedia"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  CheckCircle2,
  CircleAlert,
  FileSearch,
  Plus,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import CustomsClassificationEditorDialog from '@/components/admin/customs/CustomsClassificationEditorDialog.vue'
import CustomsLookupDialog from '@/components/admin/customs/CustomsLookupDialog.vue'
import CustomsProductCoveragePanel from '@/components/admin/customs/CustomsProductCoveragePanel.vue'
import CustomsTemplatePanel from '@/components/admin/customs/CustomsTemplatePanel.vue'
import ProductEditorDialog from '@/components/admin/product/ProductEditorDialog.vue'
import { Button } from '@/components/ui/button'
import { useCustomsClassificationCenter } from '@/composables/customs/useCustomsClassificationCenter'
import { useProductEditor } from '@/composables/product/useProductEditor'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const hasPermission = (permission: string) => authStore.hasPermission(permission)
const lookupDialogOpen = ref(false)

const {
  productRows,
  productLoading,
  productPage,
  productPageSize,
  productTotal,
  productFilters,
  templates,
  templateLoading,
  templateDialogOpen,
  templateSaving,
  templateForm,
  lookupProvider,
  lookupQuery,
  lookupLoading,
  lookupCompleted,
  lookupCandidates,
  catalogTotal,
  incompleteTotal,
  completeTotal,
  missingHSTotal,
  brands,
  shippingTemplates,
  informationTemplates,
  activeTemplates: activeCustomsTemplates,
  refreshWorkbench,
  fetchProductRows,
  fetchStats,
  fetchTemplates,
  fetchEditorResources,
  applyProductFilters,
  updateProductPage,
  updateProductPageSize,
  openTemplateCreate,
  openTemplateEdit,
  setTemplateProductSpecTemplate,
  saveTemplate,
  removeTemplate,
  runLookup,
  openTemplateFromCandidate,
} = useCustomsClassificationCenter({
  locale: supportedLanguages.defaultLocale,
})

const statItems = computed(() => [
  { key: 'total', label: '商品总数', value: catalogTotal.value, icon: FileSearch, tone: 'gray' },
  { key: 'incomplete', label: '资料不完整', value: incompleteTotal.value, icon: CircleAlert, tone: 'amber' },
  { key: 'complete', label: '资料完整', value: completeTotal.value, icon: CheckCircle2, tone: 'green' },
  { key: 'missing-hs', label: '缺 HS Code', value: missingHSTotal.value, icon: CircleAlert, tone: 'coral' },
])

const {
  productSpecTemplates,
  dialogVisible,
  dialogMode,
  submitting,
  formErrors,
  productForm,
  uploadingMedia,
  selectedProductSpecTemplate,
  selectedSpecDefinitions,
  variantSpecDefinitions,
  defaultVariantIndex,
  productSpecTemplateSelectValue,
  brandSelectValue,
  shippingTemplateSelectValue,
  afterSalesTemplateSelectValue,
  packagingTemplateSelectValue,
  templateScopedValuesTouched,
  parseSpecOptions,
  formatSpecOption,
  getSpecLabel,
  specSelectValue,
  setSpecSelectValue,
  setProductShippingTemplate,
  setProductBrand,
  setProductInformationTemplate,
  applyCustomsClassification,
  clearCustomsClassification,
  clearFieldError,
  addMediaUrl,
  handleMediaUpload,
  setPrimaryMedia,
  moveMedia,
  removeMedia,
  addVariant,
  removeVariant,
  setDefaultVariant,
  setVariantActive,
  handleProductSpecTemplateSelect,
  fetchProductSpecTemplates: fetchEditorProductSpecTemplates,
  showEditDialog,
  submitForm,
} = useProductEditor({
  refreshProducts: refreshWorkbench,
  defaultLocale: supportedLanguages.defaultLocale,
})

const activeCustomsClassifications = computed(() => activeCustomsTemplates.value.filter((profile) => (
  String(profile.id) === String(productForm.customs_classification_profile_id || '')
  || (
    !profile.product_specification_template_id
    || (productForm.product_specification_template_id != null && String(profile.product_specification_template_id) === String(productForm.product_specification_template_id))
  )
)))

const customsClassificationSelectValue = computed(() => (
  productForm.customs_classification_profile_id
    ? String(productForm.customs_classification_profile_id)
    : '__none__'
))

const handleCustomsClassificationSelect = (value: string) => {
  if (value === '__none__') {
    clearCustomsClassification()
    return
  }
  const profile = activeCustomsClassifications.value.find((item) => String(item.id) === value)
  if (!profile) return
  applyCustomsClassification(profile)
  toast.success('已套用清关资料模板')
}

const editorAfterSalesTemplates = computed(() => informationTemplates.value.filter((item: any) => (
  item.kind === 'after_sales' && (item.is_enabled !== false || item.id === productForm.after_sales_template_id)
)))
const editorPackagingTemplates = computed(() => informationTemplates.value.filter((item: any) => (
  item.kind === 'packaging' && (item.is_enabled !== false || item.id === productForm.packaging_template_id)
)))

const openProductEditor = (product: Record<string, any>) => {
  void showEditDialog(product)
}

const handleCreateTemplateFromCandidate = (candidate: Parameters<typeof openTemplateFromCandidate>[0]) => {
  lookupDialogOpen.value = false
  openTemplateFromCandidate(candidate)
}

onMounted(async () => {
  await supportedLanguages.fetchLanguages()
  await Promise.all([
    fetchEditorProductSpecTemplates(),
    fetchEditorResources(),
    fetchTemplates(),
    fetchProductRows(),
    fetchStats(),
  ])
})
</script>

<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品管理" description="管理商品资料、规格、SKU 变体和库存状态">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/brands">
            <Tag class="size-4" />
            商品品牌
          </RouterLink>
        </Button>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/templates">
            <Tags class="size-4" />
            商品规格模板
          </RouterLink>
        </Button>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/customs-classifications">
            <Tags class="size-4" />
            清关资料中心
          </RouterLink>
        </Button>
        <Button v-if="hasPermission('product:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加商品
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <ProductFilterPanel
      :filters="filters"
      :status-filter-options="statusFilterOptions"
      :locale-filter-options="localeFilterOptions"
      :featured-filter-options="featuredFilterOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <ProductTablePanel
      :loading="loading"
      :products="products"
      :selected-products="selectedProducts"
      :pagination="pagination"
      :selection-state="selectionState"
      :can-edit="hasPermission('product:edit')"
      :can-delete="hasPermission('product:delete')"
      :can-manage-translations="hasPermission('product:edit') || hasPermission('product:create')"
      :can-sync-google="hasPermission('merchant:edit')"
      :locale-name="supportedLanguages.localeName"
      @batch-status="requestBatchStatus"
      @batch-delete="requestBatchDelete"
      @toggle-all-products="toggleAllProducts"
      @toggle-product="toggleProduct"
      @edit="showEditDialog"
      @translations="showTranslationsDialog"
      @sync-google="openGoogleSync"
      @toggle-status="requestToggleStatus"
      @delete="requestDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <ProductTranslationGroupDialog
      v-model:open="translationDialogVisible"
      :current-product="translationProduct"
      :translation-group="translationGroup"
      :translations-loading="translationLoading"
      :copying-locale="copyingLocale"
      :can-edit="hasPermission('product:edit')"
      :can-create-translation="hasPermission('product:create')"
      :locale-name="supportedLanguages.localeName"
      :total-languages="supportedLanguages.enabledLanguages.value.length"
      @copy="copyTranslation"
      @edit="editTranslation"
    />

    <ProductEditorDialog
      v-model:open="dialogVisible"
      :mode="dialogMode"
      :submitting="submitting"
      :form="productForm"
      :errors="formErrors"
      :product-spec-templates="productSpecTemplates"
      :brands="brands"
      :product-categories="productCategories"
      :selected-product-spec-template="selectedProductSpecTemplate"
      :brand-select-value="brandSelectValue"
      :selected-spec-definitions="selectedSpecDefinitions"
      :variant-spec-definitions="variantSpecDefinitions"
      :default-variant-index="defaultVariantIndex"
      :product-spec-template-select-value="productSpecTemplateSelectValue"
      :product-category-select-value="productCategorySelectValue"
      :shipping-template-select-value="shippingTemplateSelectValue"
      :shipping-templates="shippingTemplates"
      :after-sales-template-select-value="afterSalesTemplateSelectValue"
      :packaging-template-select-value="packagingTemplateSelectValue"
      :after-sales-templates="afterSalesTemplates"
      :packaging-templates="packagingTemplates"
      :customs-classifications="availableCustomsClassifications"
      :customs-classification-select-value="customsClassificationSelectValue"
      :template-scoped-values-touched="templateScopedValuesTouched"
      :uploading-media="uploadingMedia"
      :procurement-visible="canViewProcurement"
      :procurement-can-edit="canEditProcurement"
      :procurement-loading="procurementLoading"
      :procurement-saving="procurementSaving"
      :procurement-pending="procurementPending"
      :procurement-error="procurementError"
      :procurement-last-saved-at="procurementLastSavedAt"
      :procurement-drafts="procurementDraftRows"
      :parse-spec-options="parseSpecOptions"
      :format-spec-option="formatSpecOption"
      :get-spec-label="getSpecLabel"
      :spec-select-value="specSelectValue"
      :language-options="languageOptions"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @product-spec-template-select="handleProductSpecTemplateSelect"
      @product-category-select="setProductCategory"
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
      @retry-procurement="retryProcurement"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :destructive="confirmation.destructive"
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Boxes,
  CircleCheck,
  PackageOpen,
  Plus,
  Tag,
  Tags,
  TriangleAlert,
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import ProductEditorDialog from '@/components/admin/product/ProductEditorDialog.vue'
import ProductFilterPanel from '@/components/admin/product/ProductFilterPanel.vue'
import ProductTablePanel from '@/components/admin/product/ProductTablePanel.vue'
import ProductTranslationGroupDialog from '@/components/admin/product/ProductTranslationGroupDialog.vue'
import productApi, { productBrandApi, productInformationTemplateApi } from '@/api/products'
import { customsClassificationApi } from '@/api/customsClassifications'
import productCategoryApi, { type ProductCategoryRecord } from '@/api/productCategories'
import shippingApi from '@/api/shipping'
import { useProductCatalog } from '@/composables/product/useProductCatalog'
import { useProductEditor } from '@/composables/product/useProductEditor'
import { useProcurementProfitDraft } from '@/composables/product/useProcurementProfitDraft'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import type { ProductTranslation, ProductTranslationGroup } from '@/components/admin/product/productEditorTypes'

const authStore = useAuthStore()
const router = useRouter()
const canViewProcurement = computed(() => authStore.hasPermission('procurement:view'))
const canEditProcurement = computed(() => authStore.hasPermission('procurement:edit'))
const {
  loading: procurementLoading,
  saving: procurementSaving,
  pending: procurementPending,
  loadError: procurementLoadError,
  saveError: procurementSaveError,
  lastSavedAt: procurementLastSavedAt,
  rowsForVariants: procurementRowsForVariants,
  loadForProduct: loadProcurementForProduct,
  saveForProduct: saveProcurementForProduct,
  retryPending: retryProcurementPending,
} = useProcurementProfitDraft()
const procurementError = computed(() => procurementLoadError.value || procurementSaveError.value)
interface ProductInformationTemplateRecord {
  id: number
  kind: 'after_sales' | 'packaging'
  name: string
  slug?: string
  locale?: string
  is_enabled?: boolean
}

interface CustomsClassificationRecord {
  id: number
  product_specification_template_id?: number | null
  name: string
  hs_code: string
  cn_code?: string
  country_of_origin?: string
  customs_description?: string
  status?: string
}

type AdminProductRecord = Record<string, any>

interface ConfirmationState {
  open: boolean
  type: '' | 'status' | 'delete' | 'batch-status' | 'batch-delete'
  target: AdminProductRecord | AdminProductRecord[] | null
  status: string
  title: string
  description: string
  confirmLabel: string
  destructive: boolean
}

const shippingTemplates = ref<any[]>([])
const brands = ref<any[]>([])
const productCategories = ref<ProductCategoryRecord[]>([])
const informationTemplates = ref<ProductInformationTemplateRecord[]>([])
const customsClassifications = ref<CustomsClassificationRecord[]>([])
const translationDialogVisible = ref(false)
const translationLoading = ref(false)
const copyingLocale = ref('')
const translationProduct = ref<AdminProductRecord | null>(null)
const translationGroup = ref<ProductTranslationGroup | null>(null)
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const {
  loading,
  products,
  selectedProducts,
  stats,
  filters,
  pagination,
  selectionState,
  fetchStats,
  fetchProducts,
  refreshProducts,
  applyFilters,
  resetFilters,
  updatePage,
  updatePageSize,
  toggleAllProducts,
  toggleProduct
} = useProductCatalog()

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
  productCategorySelectValue,
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
  setProductCategory,
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
  fetchProductSpecTemplates,
  showCreateDialog,
  showEditDialog,
  submitForm
} = useProductEditor({
  refreshProducts,
  defaultLocale: supportedLanguages.defaultLocale,
  afterProductLoaded: async (product) => {
    if (!canViewProcurement.value) return
    await loadProcurementForProduct(product)
  },
  afterProductSaved: async (savedProduct: AdminProductRecord) => {
    if (!canEditProcurement.value) return {}
    if (procurementLoadError.value) {
      toast.warning('商品已保存，但成本资料尚未加载完成')
      return { keepDialogOpen: true }
    }
    const result = await saveProcurementForProduct(savedProduct)
    if (!result.success) {
      toast.warning('商品已保存，成本与利润资料待重试')
      return { keepDialogOpen: true }
    }
    if (result.result?.skipped?.length) {
      toast.warning('商品已保存，部分 SKU 尚未生成利润快照')
    } else {
      toast.success('成本与利润资料已保存')
    }
    return {}
  },
})

const procurementDraftRows = computed(() => procurementRowsForVariants(
  productForm.variants,
  productForm.name,
  productForm.currency,
))

const retryProcurement = async () => {
  if (!canEditProcurement.value) return
  const result = await retryProcurementPending()
  if (result.success) toast.success('成本与利润资料已重试保存')
  else toast.error('成本与利润资料重试失败')
}

const fetchShippingTemplates = async () => {
  try {
    shippingTemplates.value = await shippingApi.listTemplates()
  } catch (error) {
    console.error('Failed to fetch shipping templates:', error)
  }
}

const fetchBrands = async () => {
  try {
    brands.value = await productBrandApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product brands:', error)
  }
}

const fetchProductCategories = async () => {
  try {
    const payload = await productCategoryApi.list({ include_disabled: true })
    const flatten = (items: ProductCategoryRecord[], result: ProductCategoryRecord[] = []) => {
      items.forEach((item) => {
        result.push(item)
        if (item.children?.length) flatten(item.children, result)
      })
      return result
    }
    productCategories.value = flatten(payload.tree)
  } catch (error) {
    console.error('Failed to fetch product categories:', error)
  }
}

const afterSalesTemplates = computed(() => informationTemplates.value.filter((item) =>
  item.kind === 'after_sales' && (item.is_enabled !== false || item.id === productForm.after_sales_template_id)
))
const packagingTemplates = computed(() => informationTemplates.value.filter((item) =>
  item.kind === 'packaging' && (item.is_enabled !== false || item.id === productForm.packaging_template_id)
))

const customsClassificationSelectValue = computed(() => (
  productForm.customs_classification_profile_id
    ? String(productForm.customs_classification_profile_id)
    : '__none__'
))

const availableCustomsClassifications = computed(() => customsClassifications.value.filter((profile) => (
  String(profile.id) === String(productForm.customs_classification_profile_id || '')
  || (
    profile.status === 'active'
    && (
      !profile.product_specification_template_id
      || (productForm.product_specification_template_id != null && String(profile.product_specification_template_id) === String(productForm.product_specification_template_id))
    )
  )
)))

const fetchInformationTemplates = async () => {
  try {
    informationTemplates.value = await productInformationTemplateApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product information templates:', error)
  }
}

const fetchCustomsClassifications = async () => {
  try {
    customsClassifications.value = await customsClassificationApi.list({
      include_paused: true,
    })
  } catch (error) {
    console.error('Failed to fetch customs classifications:', error)
  }
}

const handleCustomsClassificationSelect = (value: string) => {
  if (value === '__none__') {
    clearCustomsClassification()
    return
  }
  const profile = customsClassifications.value.find((item) => String(item.id) === value)
  if (!profile) return
  applyCustomsClassification(profile)
  toast.success('已套用清关资料模板')
}

const confirmation = reactive<ConfirmationState>({
  open: false,
  type: '',
  target: null,
  status: '',
  title: '',
  description: '',
  confirmLabel: '确定',
  destructive: false
})

const statusFilterOptions = [
  { label: '全部状态', value: 'all' },
  { label: '在售', value: 'active' },
  { label: '下架', value: 'inactive' },
  { label: '缺货', value: 'out_of_stock' }
]
const localeFilterOptions = supportedLanguages.localeFilterOptions
const featuredFilterOptions = [
  { label: '全部商品', value: 'all' },
  { label: '仅精选', value: 'true' },
  { label: '非精选', value: 'false' }
]

const statItems = computed(() => [
  { key: 'total', label: '总商品数', value: stats.value.total || 0, icon: Boxes, tone: 'gray' },
  { key: 'active', label: '在售商品', value: stats.value.active || 0, icon: CircleCheck, tone: 'green' },
  { key: 'low-stock', label: '低库存', value: stats.value.low_stock || 0, icon: TriangleAlert, tone: 'amber' },
  { key: 'out-of-stock', label: '缺货商品', value: stats.value.out_of_stock || 0, icon: PackageOpen, tone: 'coral' }
])

const hasPermission = (permission: string) => authStore.hasPermission(permission)
const openGoogleSync = (product: AdminProductRecord) => {
  if (!hasPermission('merchant:edit')) return
  router.push({ name: 'GoogleMerchant', query: { product_id: product.id } })
}

const showTranslationsDialog = async (product: AdminProductRecord) => {
  translationProduct.value = product
  translationGroup.value = null
  translationDialogVisible.value = true
  translationLoading.value = true
  try {
    translationGroup.value = await productApi.translations(product.id)
  } catch (error) {
    console.error('Failed to fetch product translations:', error)
    toast.error('获取商品翻译组失败')
    translationDialogVisible.value = false
  } finally {
    translationLoading.value = false
  }
}

const copyTranslation = async (locale: string) => {
  if (!translationProduct.value || copyingLocale.value) return
  copyingLocale.value = locale
  try {
    const payload = await productApi.copyTranslation(translationProduct.value.id, locale)
    translationGroup.value = payload.translation_group || await productApi.translations(translationProduct.value.id)
    toast.success(`商品已复制到${supportedLanguages.localeName(locale)}`)
    await refreshProducts()
  } catch (error) {
    console.error('Failed to copy product translation:', error)
    toast.error('复制商品翻译失败')
  } finally {
    copyingLocale.value = ''
  }
}

const editTranslation = (translation: ProductTranslation) => {
  translationDialogVisible.value = false
  void showEditDialog(translation)
}

const setConfirmation = (values: Partial<ConfirmationState>) => Object.assign(confirmation, {
  open: true,
  type: '',
  target: null,
  status: '',
  confirmLabel: '确定',
  destructive: false,
  ...values
})
const requestToggleStatus = (product: AdminProductRecord) => {
  const status = product.status === 'active' ? 'inactive' : 'active'
  const action = status === 'active' ? '上架' : '下架'
  setConfirmation({
    type: 'status', target: product, status, title: `${action}商品？`,
    description: `商品“${product.name}”将被${action}。`, confirmLabel: action
  })
}
const requestDelete = (product: AdminProductRecord) => setConfirmation({
  type: 'delete', target: product, title: '删除商品？',
  description: `商品“${product.name}”将被永久删除，此操作不可恢复。`, confirmLabel: '删除', destructive: true
})
const requestBatchStatus = (status: string) => {
  const action = status === 'active' ? '上架' : '下架'
  setConfirmation({
    type: 'batch-status', target: [...selectedProducts.value], status, title: `批量${action}商品？`,
    description: `将 ${selectedProducts.value.length} 个商品批量${action}。`, confirmLabel: `批量${action}`
  })
}
const requestBatchDelete = () => setConfirmation({
  type: 'batch-delete', target: [...selectedProducts.value], title: '批量删除商品？',
  description: `${selectedProducts.value.length} 个商品将被永久删除，此操作不可恢复。`, confirmLabel: '批量删除', destructive: true
})
const executeConfirmedAction = async () => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'status') {
      if (!target || Array.isArray(target)) return
      await productApi.updateStatus(target.id, status)
      toast.success(status === 'active' ? '商品已上架' : '商品已下架')
    } else if (type === 'delete') {
      if (!target || Array.isArray(target)) return
      await productApi.deleteProduct(target.id)
      toast.success('商品已删除')
    } else if (type === 'batch-status') {
      if (!Array.isArray(target)) return
      await productApi.batchUpdateStatus(target.map((product) => product.id), status)
      toast.success(status === 'active' ? '商品已批量上架' : '商品已批量下架')
    } else if (type === 'batch-delete') {
      if (!Array.isArray(target)) return
      await productApi.batchDelete(target.map((product) => product.id))
      toast.success('商品已批量删除')
    }
    await refreshProducts()
  } catch (error) {
    console.error('Failed to update products:', error)
  }
}

onMounted(() => Promise.all([
  supportedLanguages.fetchLanguages(),
  fetchProductSpecTemplates(),
  fetchBrands(),
  fetchProductCategories(),
  fetchShippingTemplates(),
  fetchInformationTemplates(),
  fetchCustomsClassifications(),
  fetchStats(),
  fetchProducts()
]))
</script>

<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品管理" description="管理商品资料、规格、SKU 变体和库存状态">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/templates">
            <Tags class="size-4" />
            产品模板
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
      :can-sync-google="hasPermission('merchant:edit')"
      :locale-name="supportedLanguages.localeName"
      @batch-status="requestBatchStatus"
      @batch-delete="requestBatchDelete"
      @toggle-all-products="toggleAllProducts"
      @toggle-product="toggleProduct"
      @edit="showEditDialog"
      @sync-google="openGoogleSync"
      @toggle-status="requestToggleStatus"
      @delete="requestDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <ProductEditorDialog
      v-model:open="dialogVisible"
      :mode="dialogMode"
      :submitting="submitting"
      :form="productForm"
      :errors="formErrors"
      :product-types="productTypes"
      :selected-product-type="selectedProductType"
      :selected-spec-definitions="selectedSpecDefinitions"
      :variant-spec-definitions="variantSpecDefinitions"
      :default-variant-index="defaultVariantIndex"
      :product-type-select-value="productTypeSelectValue"
      :shipping-template-select-value="shippingTemplateSelectValue"
      :shipping-templates="shippingTemplates"
      :after-sales-template-select-value="afterSalesTemplateSelectValue"
      :packaging-template-select-value="packagingTemplateSelectValue"
      :after-sales-templates="afterSalesTemplates"
      :packaging-templates="packagingTemplates"
      :template-scoped-values-touched="templateScopedValuesTouched"
      :uploading-media="uploadingMedia"
      :parse-spec-options="parseSpecOptions"
      :format-spec-option="formatSpecOption"
      :get-spec-label="getSpecLabel"
      :spec-select-value="specSelectValue"
      :language-options="languageOptions"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @product-type-select="handleProductTypeSelect"
      @product-shipping-template-select="setProductShippingTemplate"
      @product-information-template-select="setProductInformationTemplate"
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
  Tags,
  TriangleAlert,
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import ProductEditorDialog from '@/components/admin/product/ProductEditorDialog.vue'
import ProductFilterPanel from '@/components/admin/product/ProductFilterPanel.vue'
import ProductTablePanel from '@/components/admin/product/ProductTablePanel.vue'
import productApi, { productInformationTemplateApi } from '@/api/products'
import shippingApi from '@/api/shipping'
import { useProductCatalog } from '@/composables/product/useProductCatalog'
import { useProductEditor } from '@/composables/product/useProductEditor'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
interface ProductInformationTemplateRecord {
  id: number
  kind: 'after_sales' | 'packaging'
  name: string
  slug?: string
  locale?: string
  is_enabled?: boolean
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
const informationTemplates = ref<ProductInformationTemplateRecord[]>([])
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
  productTypes,
  dialogVisible,
  dialogMode,
  submitting,
  formErrors,
  productForm,
  uploadingMedia,
  selectedProductType,
  selectedSpecDefinitions,
  variantSpecDefinitions,
  defaultVariantIndex,
  productTypeSelectValue,
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
  setProductInformationTemplate,
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
  handleProductTypeSelect,
  fetchProductTypes,
  showCreateDialog,
  showEditDialog,
  submitForm
} = useProductEditor({ refreshProducts, defaultLocale: supportedLanguages.defaultLocale })

const fetchShippingTemplates = async () => {
  try {
    shippingTemplates.value = await shippingApi.listTemplates()
  } catch (error) {
    console.error('Failed to fetch shipping templates:', error)
  }
}

const afterSalesTemplates = computed(() => informationTemplates.value.filter((item) =>
  item.kind === 'after_sales' && (item.is_enabled !== false || item.id === productForm.after_sales_template_id)
))
const packagingTemplates = computed(() => informationTemplates.value.filter((item) =>
  item.kind === 'packaging' && (item.is_enabled !== false || item.id === productForm.packaging_template_id)
))

const fetchInformationTemplates = async () => {
  try {
    informationTemplates.value = await productInformationTemplateApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product information templates:', error)
  }
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
  fetchProductTypes(),
  fetchShippingTemplates(),
  fetchInformationTemplates(),
  fetchStats(),
  fetchProducts()
]))
</script>

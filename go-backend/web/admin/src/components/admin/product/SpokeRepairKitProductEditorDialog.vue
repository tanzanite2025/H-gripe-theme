<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      size="full"
      data-spoke-repair-kit-editor-dialog
      class="!flex h-[94dvh] max-h-[calc(100dvh-1rem)] !w-[95dvw] !max-w-[95dvw] flex-col gap-0 overflow-hidden p-0"
      style="overflow: hidden;"
      @open-auto-focus="handleSpokeRepairKitDialogOpenAutoFocus"
    >
      <form class="flex min-h-0 min-w-0 flex-1 flex-col" @submit.prevent="submitSpokeRepairKitProductForm">
        <DialogHeader class="shrink-0 border-b px-5 py-4 pr-12">
          <DialogTitle>{{ mode === 'create' ? '新建辐条修补件' : '编辑辐条修补件' }}</DialogTitle>
          <DialogDescription>这里沿用商品的描述、清关、成本、媒体和发布设置，只去掉商品规格模板链路；适配轮组型号是这个商品专用的购买选项。</DialogDescription>
        </DialogHeader>

        <div ref="scrollContainer" class="min-h-0 min-w-0 flex-1 space-y-4 overflow-x-hidden overflow-y-auto overscroll-contain px-5 pb-8 pt-4 [scrollbar-gutter:stable]" @wheel.stop @touchmove.stop>
          <ol class="grid gap-1.5 rounded-lg border border-dashed bg-muted/20 p-2 text-xs text-muted-foreground sm:grid-cols-2 xl:grid-cols-5">
            <li v-for="step in editorSteps" :key="step.no" class="flex min-w-0 items-center gap-2 rounded-md bg-background/70 px-2.5 py-1.5">
              <span class="font-mono text-[10px] font-black text-primary">{{ step.no }}</span>
              <strong class="min-w-0 truncate text-foreground">{{ step.label }}</strong>
            </li>
          </ol>

          <AdminFormSection title="基础信息" description="商品基础识别信息直接写入商品；分类固定为“轮组配件 / 辐条修补件”。">
            <div class="grid gap-4 md:grid-cols-3">
              <AdminFormField label="商品名称" required :error="errors.name">
                <Input v-model="form.name" placeholder="请输入商品名称" @input="clearSpokeRepairKitFormError('name')" />
              </AdminFormField>
              <AdminFormField label="商品品牌" description="品牌会用于前台详情、SEO 和 Merchant 数据。">
                <Select :model-value="form.brand_id || '__none__'" @update:model-value="setSpokeRepairKitProductBrand($event)">
                  <SelectTrigger class="w-full"><SelectValue placeholder="未设置品牌" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">未设置品牌</SelectItem>
                    <SelectItem v-for="brand in brands" :key="brand.id" :value="String(brand.id)" :disabled="brand.is_enabled === false && String(form.brand_id) !== String(brand.id)">
                      {{ brand.name }}{{ brand.is_enabled === false ? '（停用）' : '' }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <AdminFormField label="Slug" required :error="errors.slug">
                <Input v-model="form.slug" class="font-mono" placeholder="例如 dt-swiss-spoke-repair-kit" @input="clearSpokeRepairKitFormError('slug')" />
              </AdminFormField>
              <AdminFormField label="语言" required :error="errors.locale" :description="mode === 'edit' ? '编辑商品时语言已锁定；如需其他语言，请新建对应语种商品。' : ''">
                <StorefrontLocaleSelect v-model="form.locale" :language-options="languageOptions" :disabled="mode === 'edit'" :locked="mode === 'edit'" locked-title="商品语言已锁定" />
              </AdminFormField>
              <AdminFormField label="主基准币种" required :error="errors.currency" description="商品和 SKU 的金额使用这个基准币种。">
                <Input v-model="form.currency" class="font-mono uppercase" disabled />
              </AdminFormField>
              <AdminFormField label="简短描述" class="md:col-span-2">
                <Textarea v-model="form.short_description" class="min-h-20" placeholder="用于列表和摘要展示" />
              </AdminFormField>
              <AdminFormField label="详细描述" class="md:col-span-3" :error="errors.description">
                <ProductDescriptionEditor v-model="form.description" @update:model-value="clearSpokeRepairKitFormError('description')" />
              </AdminFormField>
            </div>
          </AdminFormSection>

          <ProductCustomsSection
            :form="form"
            :errors="errors"
            :customs-classifications="availableSpokeRepairKitCustomsClassifications"
            customs-description-placeholder="例如 Bicycle spoke repair kit"
            description="清关字段和资料模板沿用现有商品字段，申报价值按订单处理。"
            @customs-classification-select="handleSpokeRepairKitCustomsClassificationSelect"
            @customs-classification-manual-edit="handleSpokeRepairKitCustomsManualEdit"
            @clear-error="clearSpokeRepairKitFormError"
          />

          <AdminFormSection title="适配轮组型号" description="只能从已核验的轮组型号目录中多选；这些型号会直接成为前台买家的单选项。">
            <SpokeRepairKitModelPicker
              :models="models"
              :selected-keys="form.model_keys"
              :loading="modelsLoading"
              :error="modelsError"
              @update:selected-keys="form.model_keys = $event; clearSpokeRepairKitFormError('model_keys')"
              @retry="loadSpokeRepairKitWheelsetModels"
            />
            <p v-if="errors.model_keys" class="mt-2 text-xs font-medium text-destructive">{{ errors.model_keys }}</p>
          </AdminFormSection>

          <AdminFormSection title="销售 SKU" description="修补包使用一个销售 SKU；价格、促销价、重量和库存仍按现有 SKU 字段保存。">
            <div class="grid gap-4 md:grid-cols-3 lg:grid-cols-5">
              <AdminFormField label="SKU" required :error="errors.variants">
                <Input v-model="defaultSpokeRepairKitSalesVariant.sku" class="font-mono" placeholder="例如 DT-SPOKE-KIT" @input="clearSpokeRepairKitFormError('variants')" />
              </AdminFormField>
              <AdminFormField label="价格" required :error="errors.variants">
                <Input v-model="defaultSpokeRepairKitSalesVariant.price" inputmode="decimal" type="text" placeholder="例如 29.90" @input="clearSpokeRepairKitFormError('variants')" />
              </AdminFormField>
              <AdminFormField label="促销价">
                <Input v-model="defaultSpokeRepairKitSalesVariant.sale_price" inputmode="decimal" type="text" placeholder="可选" />
              </AdminFormField>
              <AdminFormField label="重量（克）">
                <Input v-model.number="defaultSpokeRepairKitSalesVariant.weight_grams" type="number" min="0" step="1" />
              </AdminFormField>
              <AdminFormField label="库存" required :error="errors.variants">
                <Input v-model.number="defaultSpokeRepairKitSalesVariant.stock" type="number" min="0" step="1" @input="clearSpokeRepairKitFormError('variants')" />
              </AdminFormField>
            </div>
          </AdminFormSection>

          <ProductProfitabilitySection
            v-if="supplierCostVisible"
            :variants="form.variants"
            :drafts="supplierCostDrafts"
            :currency="form.currency"
            :can-edit="supplierCostCanEdit"
            :loading="supplierCostLoading"
            :saving="supplierCostSaving"
            :pending="supplierCostPending"
            :error="supplierCostError"
            :last-saved-at="supplierCostLastSavedAt"
            @retry="retrySpokeRepairKitSupplierCostSave"
          />

          <ProductMediaSection
            :media-items="form.media"
            :variants="form.variants"
            :variant-option-values="form.variant_option_values"
            :spec-definitions="[]"
            :uploading="uploadingMedia"
            :error="errors.media"
            @upload="(...args) => handleMediaUpload(...args)"
            @add-url="addMediaUrl"
            @clear-error="clearSpokeRepairKitFormError('media')"
            @set-primary="setPrimaryMedia"
            @move="moveMedia"
            @remove="removeMedia"
          />

          <ProductPublishingSection
            :form="form"
            :shipping-templates="shippingTemplates"
            :after-sales-templates="afterSalesTemplates"
            :packaging-templates="packagingTemplates"
            id-prefix="spoke-repair-kit"
            description="控制商品的公开状态、物流资料、售后资料和前台精选展示。"
            @product-shipping-template-select="setSpokeRepairKitProductShippingTemplate"
            @product-information-template-select="(...args) => setSpokeRepairKitInformationTemplate(...args)"
          />
        </div>

        <DialogFooter class="mx-0 mb-0 w-full min-w-0 shrink-0 flex-wrap border-t px-5 py-3 overflow-visible">
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting || supplierCostPending">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ supplierCostPending ? '请先处理成本资料' : submitting ? '保存中' : (mode === 'create' ? '创建修补件商品' : '保存修补件商品') }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import { toast } from 'vue-sonner'
import productApi from '@/api/products'
import spokeRepairKitCatalogApi, { type SpokeRepairKitCatalogModel } from '@/api/spokeRepairKitCatalog'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import ProductCustomsSection from '@/components/admin/product/ProductCustomsSection.vue'
import ProductDescriptionEditor from '@/components/admin/product/ProductDescriptionEditor.vue'
import ProductMediaSection from '@/components/admin/product/ProductMediaSection.vue'
import ProductProfitabilitySection from '@/components/admin/product/ProductProfitabilitySection.vue'
import ProductPublishingSection from '@/components/admin/product/ProductPublishingSection.vue'
import SpokeRepairKitModelPicker from '@/components/admin/product/SpokeRepairKitModelPicker.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { buildProductMediaFormValues } from '@/lib/productMedia'
import { useProductMediaManager } from '@/composables/product/useProductMediaManager'
import { useProductSupplierCostProfitDraft } from '@/composables/product/useProductSupplierCostProfitDraft'

interface BrandRecord { id: number | string; name: string; is_enabled?: boolean }
interface LanguageOption { value: string; label: string }
interface TemplateRecord { id: number | string; name: string; locale?: string; is_enabled?: boolean; enabled?: boolean }
interface CustomsClassificationRecord {
  id: number
  name: string
  hs_code: string
  cn_code?: string
  country_of_origin?: string
  customs_description?: string
  verified_at?: string | null
  review_due_at?: string | null
  status?: string
}

const editorSteps = [
  { no: '01', label: '基础识别' },
  { no: '02', label: '清关资料' },
  { no: '03', label: '适配型号' },
  { no: '04', label: 'SKU 与成本' },
  { no: '05', label: '媒体与发布' },
]

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: 'create' | 'edit'
  product?: Record<string, any> | null
  categoryId?: number | string | null
  brands?: BrandRecord[]
  defaultLocale?: string
  currency?: string
  languageOptions?: LanguageOption[]
  shippingTemplates?: TemplateRecord[]
  afterSalesTemplates?: TemplateRecord[]
  packagingTemplates?: TemplateRecord[]
  customsClassifications?: CustomsClassificationRecord[]
  supplierCostVisible?: boolean
  supplierCostCanEdit?: boolean
}>(), {
  open: false,
  mode: 'create',
  product: null,
  categoryId: null,
  brands: () => [],
  defaultLocale: 'en',
  currency: 'USD',
  languageOptions: () => [],
  shippingTemplates: () => [],
  afterSalesTemplates: () => [],
  packagingTemplates: () => [],
  customsClassifications: () => [],
  supplierCostVisible: false,
  supplierCostCanEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'saved', value: any): void
}>()

const scrollContainer = ref<HTMLElement | null>(null)
const models = ref<SpokeRepairKitCatalogModel[]>([])
const modelsLoading = ref(false)
const modelsError = ref('')
const submitting = ref(false)
const errors = reactive<Record<string, string>>({})

const createEmptySpokeRepairKitSalesVariant = (overrides: Record<string, any> = {}): Record<string, any> => ({
  id: null,
  shipping_template_id: null,
  sku: '',
  title: '',
  option_values: {},
  currency: props.currency,
  price: '',
  price_minor: 0,
  sale_price: '',
  sale_price_minor: null,
  stock: 0,
  weight_grams: 0,
  is_default: true,
  is_active: true,
  sort_order: 0,
  option_group_rules: [],
  option_value_rules: [],
  ...overrides,
})

const form = reactive<any>({
  id: null,
  product_specification_template_id: null,
  product_category_id: null,
  brand_id: null,
  shipping_template_id: null,
  after_sales_template_id: null,
  packaging_template_id: null,
  customs_classification_profile_id: null,
  hs_code: '',
  cn_code: '',
  country_of_origin: '',
  customs_description: '',
  name: '',
  slug: '',
  description: '',
  short_description: '',
  currency: props.currency,
  status: 'active',
  locale: props.defaultLocale,
  featured: false,
  media: [],
  model_keys: [],
  variant_option_values: [],
  option_value_relations: [],
  variants: [createEmptySpokeRepairKitSalesVariant()],
})

const clearSpokeRepairKitFormErrors = (): void => Object.keys(errors).forEach((key) => delete errors[key])
const clearSpokeRepairKitFormError = (key: string): void => { delete errors[key] }
const defaultSpokeRepairKitSalesVariant = computed(() => form.variants[0] || (form.variants[0] = createEmptySpokeRepairKitSalesVariant()))

const getSpokeRepairKitCurrencyMinorUnitCount = (value: unknown): number => {
  const code = String(value || '').trim().toUpperCase()
  if (['BHD', 'IQD', 'JOD', 'KWD', 'LYD', 'OMR', 'TND'].includes(code)) return 3
  if (['BIF', 'CLP', 'DJF', 'GNF', 'JPY', 'KMF', 'KRW', 'MGA', 'PYG', 'RWF', 'UGX', 'VND', 'VUV', 'XAF', 'XOF', 'XPF'].includes(code)) return 0
  return 2
}
const convertSpokeRepairKitMinorAmountToMajorDisplayValue = (value: unknown, currency: unknown): string => {
  const minor = Number(value)
  if (!Number.isFinite(minor)) return ''
  const units = getSpokeRepairKitCurrencyMinorUnitCount(currency)
  return (minor / (10 ** units)).toFixed(units)
}
const convertSpokeRepairKitMajorAmountToMinorUnits = (value: unknown, currency: unknown): number => {
  const major = Number(value)
  return Number.isFinite(major) ? Math.round(major * (10 ** getSpokeRepairKitCurrencyMinorUnitCount(currency))) : 0
}
const readSpokeRepairKitVariantPriceForForm = (variant: any, product: any): string => {
  const explicit = variant?.price_decimal || variant?.price
  if (explicit !== undefined && explicit !== null && String(explicit).trim() !== '') return String(explicit)
  return variant?.price_minor == null
    ? String(product?.price_decimal || '')
    : convertSpokeRepairKitMinorAmountToMajorDisplayValue(variant.price_minor, variant?.currency || product?.currency || props.currency)
}
const readSpokeRepairKitVariantSalePriceForForm = (variant: any, product: any): string => {
  const explicit = variant?.sale_price_decimal || variant?.sale_price
  if (explicit !== undefined && explicit !== null && String(explicit).trim() !== '') return String(explicit)
  return variant?.sale_price_minor == null
    ? String(product?.sale_price_decimal || '')
    : convertSpokeRepairKitMinorAmountToMajorDisplayValue(variant.sale_price_minor, variant?.currency || product?.currency || props.currency)
}
const parseSpokeRepairKitVariantOptionValues = (value: unknown): Record<string, any> => {
  if (value && typeof value === 'object') return { ...(value as Record<string, any>) }
  try {
    const parsed = JSON.parse(String(value || '{}'))
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}
const buildSpokeRepairKitWheelsetModelKeyFromRecord = (item: any): string => `${item?.brand_slug || item?.brandSlug || ''}:${item?.wheelset_model_slug || item?.slug || ''}`.toLowerCase()

const populateSpokeRepairKitProductForm = (product: Record<string, any> | null): void => {
  clearSpokeRepairKitFormErrors()
  const rawVariant = Array.isArray(product?.variants) ? product.variants.find((item: any) => item?.is_default) || product.variants[0] : null
  const currency = String(product?.currency || props.currency || 'USD').toUpperCase()
  const variant = rawVariant || createEmptySpokeRepairKitSalesVariant({ currency })
  Object.assign(form, {
    id: product?.id || null,
    product_specification_template_id: null,
    product_category_id: props.categoryId || null,
    brand_id: product?.brand_id ?? product?.brand?.id ?? null,
    shipping_template_id: product?.shipping_template_id ?? null,
    after_sales_template_id: product?.after_sales_template_id ?? product?.after_sales_template?.id ?? null,
    packaging_template_id: product?.packaging_template_id ?? product?.packaging_template?.id ?? null,
    customs_classification_profile_id: product?.customs_classification_profile_id ?? product?.customs_classification_profile?.id ?? null,
    hs_code: String(product?.hs_code || ''),
    cn_code: String(product?.cn_code || ''),
    country_of_origin: String(product?.country_of_origin || ''),
    customs_description: String(product?.customs_description || ''),
    name: String(product?.name || ''),
    slug: String(product?.slug || ''),
    description: String(product?.description || ''),
    short_description: String(product?.short_description || product?.short_desc || ''),
    currency,
    status: String(product?.status || 'active'),
    locale: String(product?.locale || props.defaultLocale || 'en'),
    featured: Boolean(product?.featured),
    model_keys: Array.isArray(product?.spoke_repair_kit_models)
      ? product.spoke_repair_kit_models.map(buildSpokeRepairKitWheelsetModelKeyFromRecord).filter((key: string) => key !== ':')
      : [],
    variant_option_values: [],
    option_value_relations: [],
    media: buildProductMediaFormValues(product),
    variants: [createEmptySpokeRepairKitSalesVariant({
      id: variant?.id || null,
      shipping_template_id: variant?.shipping_template_id ?? null,
      sku: String(variant?.sku || product?.sku || ''),
      title: String(variant?.title || ''),
      option_values: parseSpokeRepairKitVariantOptionValues(variant?.option_values),
      currency: String(variant?.currency || currency).toUpperCase(),
      price: readSpokeRepairKitVariantPriceForForm(variant, product),
      price_minor: Number(variant?.price_minor || 0),
      sale_price: readSpokeRepairKitVariantSalePriceForForm(variant, product),
      sale_price_minor: variant?.sale_price_minor == null ? null : Number(variant.sale_price_minor),
      stock: Number(variant?.stock ?? product?.stock ?? 0),
      weight_grams: Number(variant?.weight_grams ?? variant?.weight ?? 0),
      is_default: true,
      is_active: variant?.is_active !== false,
    })],
  })
}

const loadSpokeRepairKitWheelsetModels = async (): Promise<void> => {
  modelsLoading.value = true
  modelsError.value = ''
  try {
    models.value = await spokeRepairKitCatalogApi.listSpokeRepairKitWheelsetModels()
  } catch (error) {
    console.error('Failed to load spoke repair-kit wheelset models:', error)
    modelsError.value = '读取轮组型号目录失败，请重试。'
  } finally {
    modelsLoading.value = false
  }
}

const mediaManager = useProductMediaManager(form, { clearFieldError: clearSpokeRepairKitFormError })
const { uploadingMedia, addMediaUrl, handleMediaUpload, setPrimaryMedia, moveMedia, removeMedia, normalizeFormMedia } = mediaManager

const costManager = useProductSupplierCostProfitDraft()
const {
  loading: supplierCostLoading,
  saving: supplierCostSaving,
  pending: supplierCostPending,
  loadError: supplierCostLoadError,
  saveError: supplierCostSaveError,
  lastSavedAt: supplierCostLastSavedAt,
  rowsForVariants,
  loadForProduct: loadSupplierCostForProduct,
  saveForProduct: saveSupplierCostForProduct,
  retryPending: retrySupplierCostPending,
} = costManager
const supplierCostError = computed(() => supplierCostLoadError.value || supplierCostSaveError.value)
const supplierCostDrafts = computed(() => rowsForVariants(form.variants, form.name, form.currency))

const loadSpokeRepairKitSupplierCostData = async (): Promise<void> => {
  if (!props.supplierCostVisible) {
    costManager.reset()
    return
  }
  if (props.mode === 'edit' && props.product) await loadSupplierCostForProduct(form)
  else costManager.reset()
}

const availableSpokeRepairKitCustomsClassifications = computed(() => props.customsClassifications.filter((profile) => (
  String(profile.id) === String(form.customs_classification_profile_id || '')
  || profile.status === 'active'
)))

const setSpokeRepairKitProductBrand = (value: unknown): void => { form.brand_id = value === '__none__' ? null : Number(value) }
const setSpokeRepairKitProductShippingTemplate = (value: unknown): void => { form.shipping_template_id = value === '__none__' ? null : Number(value) }
const setSpokeRepairKitInformationTemplate = (field: 'after_sales_template_id' | 'packaging_template_id', value: unknown): void => { form[field] = value === '__none__' ? null : Number(value) }
const handleSpokeRepairKitCustomsClassificationSelect = (value: unknown): void => {
  if (value === '__none__') {
    form.customs_classification_profile_id = null
    return
  }
  const profile = props.customsClassifications.find((item) => String(item.id) === String(value))
  if (!profile) return
  form.customs_classification_profile_id = profile.id
  form.hs_code = String(profile.hs_code || '')
  form.cn_code = String(profile.cn_code || '')
  form.country_of_origin = String(profile.country_of_origin || '').toUpperCase()
  form.customs_description = String(profile.customs_description || '')
}
const handleSpokeRepairKitCustomsManualEdit = (field: string): void => {
  form.customs_classification_profile_id = null
  clearSpokeRepairKitFormError(field)
}

const buildSpokeRepairKitSalesVariantPayload = (): Record<string, any> => {
  const variant = defaultSpokeRepairKitSalesVariant.value
  const salePriceValue = String(variant.sale_price ?? '').trim()
  return {
    ...(variant.id ? { id: variant.id } : {}),
    shipping_template_id: variant.shipping_template_id || null,
    sku: String(variant.sku || '').trim(),
    title: String(variant.title || '').trim(),
    option_values: {},
    currency: form.currency,
    price_minor: convertSpokeRepairKitMajorAmountToMinorUnits(variant.price, form.currency),
    sale_price_minor: salePriceValue ? convertSpokeRepairKitMajorAmountToMinorUnits(salePriceValue, form.currency) : null,
    stock: Number(variant.stock || 0),
    weight_grams: Number(variant.weight_grams || 0),
    is_default: true,
    is_active: variant.is_active !== false,
    sort_order: 0,
    option_group_rules: [],
    option_value_rules: [],
  }
}

const validateSpokeRepairKitProductForm = (): boolean => {
  clearSpokeRepairKitFormErrors()
  const variant = defaultSpokeRepairKitSalesVariant.value
  if (!form.name.trim()) errors.name = '请输入商品名称'
  if (!/^[a-z0-9]+(?:[_-][a-z0-9]+)*$/.test(form.slug.trim())) errors.slug = '请输入小写字母、数字、下划线或短横线组成的 slug'
  if (!form.locale) errors.locale = '请选择语言'
  if (!/^[A-Z]{3}$/.test(form.currency)) errors.currency = '商品币种无效'
  if (!form.model_keys.length) errors.model_keys = '至少选择一个适配轮组型号'
  if (!variant.sku.trim()) errors.variants = '请输入 SKU'
  if (convertSpokeRepairKitMajorAmountToMinorUnits(variant.price, form.currency) <= 0) errors.variants = '价格必须大于 0'
  if (String(variant.sale_price || '').trim() && convertSpokeRepairKitMajorAmountToMinorUnits(variant.sale_price, form.currency) < 0) errors.variants = '促销价不能为负数'
  if (!Number.isFinite(Number(variant.stock)) || Number(variant.stock) < 0) errors.variants = '库存不能为负数'
  if (form.hs_code && !/^\d{6}$/.test(form.hs_code)) errors.hs_code = 'HS Code 必须是 6 位数字'
  if (form.cn_code && !/^\d{8}$/.test(form.cn_code)) errors.cn_code = 'CN Code 必须是 8 位数字'
  if (form.country_of_origin && !/^[A-Z]{2}$/.test(String(form.country_of_origin).toUpperCase())) errors.country_of_origin = '请输入 2 位国家代码，例如 CN'
  if (String(form.customs_description || '').length > 255) errors.customs_description = '英文报关品名不能超过 255 个字符'
  const media = normalizeFormMedia()
  if (form.media.some((item: any) => !String(item.url || '').trim())) errors.media = '媒体条目必须填写 URL，空条目请删除'
  else if (media.filter((item: any) => item.media_type === 'image' && item.is_primary).length > 1) errors.media = '商品主图只能设置一张'
  if (!props.categoryId) errors.category = '辐条修补件分类尚未加载'
  if (Object.keys(errors).length) {
    toast.error('请检查修补件商品表单中的必填项')
    return false
  }
  return true
}

const submitSpokeRepairKitProductForm = async (): Promise<void> => {
  if (!validateSpokeRepairKitProductForm() || submitting.value || supplierCostPending.value) return
  submitting.value = true
  const payload = {
    product_specification_template_id: null,
    product_category_id: Number(props.categoryId),
    brand_id: form.brand_id ? Number(form.brand_id) : null,
    shipping_template_id: form.shipping_template_id,
    after_sales_template_id: form.after_sales_template_id,
    packaging_template_id: form.packaging_template_id,
    customs_classification_profile_id: form.customs_classification_profile_id || null,
    hs_code: String(form.hs_code || '').trim(),
    cn_code: String(form.cn_code || '').trim(),
    country_of_origin: String(form.country_of_origin || '').trim().toUpperCase(),
    customs_description: String(form.customs_description || '').trim(),
    name: form.name.trim(),
    slug: form.slug.trim().toLowerCase(),
    description: form.description,
    short_description: form.short_description,
    currency: form.currency,
    fulfillment_mode: 'stock',
    status: form.status,
    locale: form.locale,
    featured: Boolean(form.featured),
    specs: {},
    variants: [buildSpokeRepairKitSalesVariantPayload()],
    variant_option_values: [],
    option_value_relations: [],
    media: normalizeFormMedia(),
    spoke_repair_kit_model_keys: [...form.model_keys],
  }

  try {
    if (props.supplierCostVisible && supplierCostLoadError.value) {
      toast.warning('成本资料尚未加载完成，请处理后再保存')
      return
    }
    const saved = props.mode === 'edit' && form.id
      ? await productApi.update(form.id, payload)
      : await productApi.create(payload)
    emit('saved', saved)

    if (props.supplierCostVisible && props.supplierCostCanEdit) {
      const costResult = await saveSupplierCostForProduct(saved)
      if (!costResult.success) {
        toast.warning('商品已保存，成本与利润资料待重试')
        return
      }
    }

    toast.success(props.mode === 'edit' ? '辐条修补件已保存' : '辐条修补件已创建')
    emit('update:open', false)
  } catch (error: any) {
    console.error('Failed to save spoke repair-kit product:', error)
    toast.error(error?.response?.data?.error || '辐条修补件保存失败')
  } finally {
    submitting.value = false
  }
}

const retrySpokeRepairKitSupplierCostSave = async (): Promise<void> => {
  if (!props.supplierCostCanEdit) return
  const result = await retrySupplierCostPending()
  if (result.success) toast.success('成本与利润资料已重试保存')
  else toast.error('成本与利润资料重试失败')
}

const handleSpokeRepairKitDialogOpenAutoFocus = (event: Event): void => {
  event.preventDefault()
  void nextTick(() => requestAnimationFrame(() => scrollContainer.value?.scrollTo({ top: 0, behavior: 'auto' })))
}

watch(() => props.open, (open) => {
  if (!open) return
  populateSpokeRepairKitProductForm(props.product)
  void loadSpokeRepairKitWheelsetModels()
  void loadSpokeRepairKitSupplierCostData()
}, { immediate: true })
watch(() => props.product, (product) => {
  if (!props.open) return
  populateSpokeRepairKitProductForm(product)
  void loadSpokeRepairKitSupplierCostData()
})
</script>

<style scoped>
:global([data-spoke-repair-kit-editor-dialog]) {
  display: flex !important;
  overflow: hidden !important;
}

:global([data-spoke-repair-kit-editor-dialog] > form) {
  min-height: 0;
}
</style>

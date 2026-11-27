<template>
  <Dialog :open="open" @update:open="requestWheelsetProductEditorOpenChange">
    <DialogContent
      size="full"
      data-wheelset-product-editor-dialog
      class="!flex h-[94dvh] max-h-[calc(100dvh-1rem)] !w-[95dvw] !max-w-[95dvw] flex-col gap-0 overflow-hidden p-0"
      style="overflow: hidden;"
      @open-auto-focus="handleWheelsetEditorOpenAutoFocus"
    >
      <form class="flex min-h-0 min-w-0 flex-1 flex-col" @submit.prevent="emit('submit')">
        <DialogHeader class="shrink-0 border-b px-5 py-4 pr-12">
          <DialogTitle>{{ mode === 'create' ? '新建 WHEELSET 商品' : '编辑 WHEELSET 商品' }}</DialogTitle>
          <DialogDescription>前轮与后轮分别录入产品级资料；没有可靠资料的可选字段留空。</DialogDescription>
        </DialogHeader>

        <div
          ref="scrollContainer"
          class="min-h-0 min-w-0 flex-1 space-y-4 overflow-x-hidden overflow-y-auto overscroll-contain px-5 pb-8 pt-4 [scrollbar-gutter:stable]"
          @wheel.stop
          @touchmove.stop
        >
          <ol class="grid gap-1.5 rounded-lg border border-dashed bg-muted/20 p-2 text-xs text-muted-foreground sm:grid-cols-2 xl:grid-cols-5">
            <li v-for="step in wheelsetEditorSteps" :key="step.no" class="flex min-w-0 items-center gap-2 rounded-md bg-background/70 px-2.5 py-1.5">
              <span class="font-mono text-[10px] font-black text-primary">{{ step.no }}</span>
              <strong class="min-w-0 truncate text-foreground">{{ step.label }}</strong>
            </li>
          </ol>

          <div v-if="Object.keys(errors).length" role="alert" class="rounded-lg border border-destructive/40 p-3 text-sm text-destructive">
            <p v-for="(message, key) in errors" :key="key">{{ message }}</p>
          </div>

          <AdminFormSection id="wheelset-basic-information" title="商品资料" description="填写商品名称、品牌和销售文案；分类固定为 WHEELSET。">
            <div class="grid gap-4 md:grid-cols-3">
              <AdminFormField label="商品名称" required :error="errors.name">
                <Input v-model="form.name" placeholder="请输入商品名称" @input="emit('clear-error', 'name')" />
              </AdminFormField>
              <AdminFormField label="商品品牌" description="选择商品的实际品牌。">
                <Select :model-value="brandSelectValue" @update:model-value="emit('product-brand-select', $event)">
                  <SelectTrigger class="w-full"><SelectValue placeholder="未设置品牌" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">未设置品牌</SelectItem>
                    <SelectItem v-for="brand in brands" :key="brand.id" :value="String(brand.id)" :disabled="brand.is_enabled === false && String(form.brand_id) !== String(brand.id)">
                      {{ brand.name }}{{ brand.is_enabled === false ? '（停用）' : '' }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <AdminFormField label="商品分类" required :error="errors.product_category_id" description="WHEELSET 商品固定归入此系统分类。">
                <Input :model-value="wheelsetSystemCategoryDisplayPath" disabled />
              </AdminFormField>
              <AdminFormField label="Slug" required :error="errors.slug">
                <Input v-model="form.slug" placeholder="例如 road-wheelset-700c" @input="emit('clear-error', 'slug')" />
              </AdminFormField>
              <AdminFormField label="语言" required :error="errors.locale" :description="mode === 'edit' ? '编辑商品时语言已锁定；如需其他语言，请从翻译管理创建。' : ''">
                <StorefrontLocaleSelect v-model="form.locale" :language-options="languageOptions" :disabled="mode === 'edit'" :locked="mode === 'edit'" locked-title="商品语言已锁定" />
              </AdminFormField>
              <AdminFormField label="主基准币种" required :error="errors.currency">
                <Input v-model="form.currency" class="font-mono uppercase" disabled />
              </AdminFormField>
              <AdminFormField label="简短描述" class="md:col-span-3">
                <Textarea v-model="form.short_description" class="min-h-20" placeholder="用于列表和摘要展示" />
              </AdminFormField>
              <AdminFormField label="详细描述" class="md:col-span-3" :error="errors.description">
                <ProductDescriptionEditor v-model="form.description" @update:model-value="emit('clear-error', 'description')" />
              </AdminFormField>
            </div>
          </AdminFormSection>

          <ProductCustomsSection
            :form="form"
            :errors="errors"
            :customs-classifications="customsClassifications"
            customs-description-placeholder="例如 Bicycle wheelset"
            manufacturer-identifier-label="制造商产品识别码（MPN）"
            manufacturer-identifier-description="填写制造商真实产品编号；该值会作为 WHEELSET Product Schema 的 mpn，最多 70 个字符"
            standard-identifier-label="EAN / GTIN（选填）"
            standard-identifier-description="填写校验位正确的 GTIN-8、GTIN-12、GTIN-13 或 GTIN-14；只允许数字，没有编码可以留空。"
            description="清关资料沿用商品公共字段；EAN / GTIN 使用现有标准商品编号字段。"
            @customs-classification-select="emit('customs-classification-select', $event)"
            @customs-classification-manual-edit="emit('customs-classification-manual-edit', $event)"
            @clear-error="emit('clear-error', $event)"
          />

          <WheelsetDedicatedProductSpecificationFields
            :specification="form.wheelset_specification"
            :locked="Boolean(form.parent_id)"
            :error="errors.wheelset_specification"
          />

          <AdminFormSection id="wheelset-sales-variants" title="销售选项" description="产品级技术规格保存在上方独立字段；本区维护实际销售编码、价格、履约重量、物流模板和库存。">
            <div class="min-w-0 rounded-md border">
              <ProductVariantEditor
                :variants="form.variants"
                :default-index="defaultVariantIndex"
                :shipping-templates="shippingTemplates"
                class="min-w-0 p-3"
                @add="emit('add-variant')"
                @remove="emit('remove-variant', $event)"
                @set-default="emit('set-default-variant', $event)"
                @set-active="(...args) => emit('set-variant-active', ...args)"
              />
            </div>
            <p v-if="errors.variants" class="mt-2 text-xs font-medium text-destructive">{{ errors.variants }}</p>
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
            @retry="emit('retry-supplier-cost')"
          />

          <ProductMediaSection
            id="wheelset-media-section"
            :media-items="form.media"
            :variants="form.variants"
            :uploading="uploadingMedia"
            :error="errors.media"
            @upload="(...args) => emit('upload-media', ...args)"
            @add-url="emit('add-media-url', $event)"
            @clear-error="emit('clear-error', 'media')"
            @set-primary="emit('set-primary-media', $event)"
            @move="(...args) => emit('move-media', ...args)"
            @remove="emit('remove-media', $event)"
          />

          <ProductPublishingSection
            id-prefix="wheelset"
            :form="form"
            :after-sales-templates="afterSalesTemplates"
            :packaging-templates="packagingTemplates"
            description="控制 WHEELSET 商品的公开状态、售后资料和前台精选展示。"
            notice="销售选项的履约重量和线路计费模板在上方维护。"
            @product-information-template-select="(...args) => emit('product-information-template-select', ...args)"
          />
        </div>

        <DialogFooter class="mx-0 mb-0 shrink-0 border-t px-5 py-4">
          <Button type="button" variant="outline" @click="requestWheelsetProductEditorOpenChange(false)">取消</Button>
          <Button type="submit" :disabled="submitting || supplierCostPending || uploadingMedia">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ supplierCostPending ? '请先处理成本资料' : submitting ? '保存中' : mode === 'create' ? '创建 WHEELSET 商品' : '保存 WHEELSET 商品' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import ProductCustomsSection from '@/components/admin/product/ProductCustomsSection.vue'
import ProductDescriptionEditor from '@/components/admin/product/ProductDescriptionEditor.vue'
import ProductMediaSection from '@/components/admin/product/ProductMediaSection.vue'
import ProductProfitabilitySection from '@/components/admin/product/ProductProfitabilitySection.vue'
import ProductPublishingSection from '@/components/admin/product/ProductPublishingSection.vue'
import ProductVariantEditor from '@/components/admin/product/ProductVariantEditor.vue'
import WheelsetDedicatedProductSpecificationFields from '@/components/admin/product/WheelsetDedicatedProductSpecificationFields.vue'
import type { ProductSupplierCostProfitDraft } from '@/composables/product/useProductSupplierCostProfitDraft'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type { ProductMediaType, ProductFormRecord, ShippingTemplateRecord } from '@/modules/product/productEditorTypes'

interface ProductBrandRecord {
  id: number | string
  name: string
  is_enabled?: boolean
}

interface ProductInformationTemplateRecord {
  id: number | string
  name: string
  locale?: string
  is_enabled?: boolean
}

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

interface LanguageOption {
  value: string
  label: string
}

const wheelsetEditorSteps = [
  { no: '01', label: '商品资料' },
  { no: '02', label: '清关与编号' },
  { no: '03', label: '前后轮规格' },
  { no: '04', label: '销售选项' },
  { no: '05', label: '媒体与发布' },
]

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: 'create' | 'edit'
  submitting?: boolean
  form: ProductFormRecord
  errors?: Record<string, string>
  wheelsetSystemCategoryDisplayPath: string
  brandSelectValue: string
  brands?: ProductBrandRecord[]
  shippingTemplates?: ShippingTemplateRecord[]
  afterSalesTemplates?: ProductInformationTemplateRecord[]
  packagingTemplates?: ProductInformationTemplateRecord[]
  customsClassifications?: CustomsClassificationRecord[]
  languageOptions?: LanguageOption[]
  defaultVariantIndex?: number
  uploadingMedia?: boolean
  supplierCostVisible?: boolean
  supplierCostCanEdit?: boolean
  supplierCostLoading?: boolean
  supplierCostSaving?: boolean
  supplierCostPending?: boolean
  supplierCostError?: string
  supplierCostLastSavedAt?: string
  supplierCostDrafts?: ProductSupplierCostProfitDraft[]
}>(), {
  open: false,
  mode: 'create',
  submitting: false,
  errors: () => ({}),
  brands: () => [],
  shippingTemplates: () => [],
  afterSalesTemplates: () => [],
  packagingTemplates: () => [],
  customsClassifications: () => [],
  languageOptions: () => [],
  defaultVariantIndex: 0,
  uploadingMedia: false,
  supplierCostVisible: false,
  supplierCostCanEdit: false,
  supplierCostLoading: false,
  supplierCostSaving: false,
  supplierCostPending: false,
  supplierCostError: '',
  supplierCostLastSavedAt: '',
  supplierCostDrafts: () => [],
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', field: string): void
  (event: 'product-brand-select', value: unknown): void
  (event: 'product-information-template-select', field: 'after_sales_template_id' | 'packaging_template_id', value: unknown): void
  (event: 'customs-classification-select', value: unknown): void
  (event: 'customs-classification-manual-edit', field: string): void
  (event: 'add-variant'): void
  (event: 'remove-variant', index: number): void
  (event: 'set-default-variant', index: number): void
  (event: 'set-variant-active', index: number, active: boolean): void
  (event: 'upload-media', value: Event, mediaType: ProductMediaType): void
  (event: 'add-media-url', mediaType: ProductMediaType): void
  (event: 'set-primary-media', index: number): void
  (event: 'move-media', index: number, direction: -1 | 1): void
  (event: 'remove-media', index: number): void
  (event: 'retry-supplier-cost'): void
}>()

const scrollContainer = ref<HTMLElement | null>(null)
const initialFormSnapshot = ref('')

watch(() => props.open, async (open) => {
  if (!open) return
  await nextTick()
  initialFormSnapshot.value = JSON.stringify(props.form)
})

const handleWheelsetEditorOpenAutoFocus = (event: Event): void => {
  event.preventDefault()
  void nextTick(() => requestAnimationFrame(() => scrollContainer.value?.scrollTo({ top: 0, behavior: 'auto' })))
}

const requestWheelsetProductEditorOpenChange = (open: boolean): void => {
  if (!open && (props.submitting || props.uploadingMedia)) return
  if (!open && initialFormSnapshot.value !== JSON.stringify(props.form) && !window.confirm('有未保存的修改，确定关闭 WHEELSET 编辑窗口？')) return
  emit('update:open', open)
}
</script>

<style scoped>
:global([data-wheelset-product-editor-dialog]) {
  display: flex !important;
  overflow: hidden !important;
}

:global([data-wheelset-product-editor-dialog] > form) {
  min-height: 0;
}
</style>

<template>
  <Dialog :open="open" @update:open="requestProductEditorOpenChange">
    <DialogContent
      size="full"
      data-product-editor-dialog
      class="!flex h-[94dvh] max-h-[calc(100dvh-1rem)] !w-[95dvw] !max-w-[95dvw] flex-col gap-0 overflow-hidden p-0"
      style="overflow: hidden;"
      @open-auto-focus="handleOpenAutoFocus"
    >
      <form class="flex min-h-0 min-w-0 flex-1 flex-col" @submit.prevent="emit('submit')">
        <DialogHeader class="shrink-0 border-b px-5 py-4 pr-12">
          <DialogTitle>{{ mode === 'create' ? (isDedicatedCarbonRim ? '添加碳纤维车圈' : '添加商品') : (isDedicatedCarbonRim ? '编辑碳纤维车圈' : '编辑商品') }}</DialogTitle>
          <DialogDescription>{{ isDedicatedCarbonRim ? '录入单只碟刹车圈的型号、尺寸和实际销售规格，再维护图片、物流及发布信息。' : '填写商品基础资料，并按实际销售方式维护 SKU、价格和库存。' }}</DialogDescription>
        </DialogHeader>

        <div
          ref="scrollContainer"
          class="product-editor-dialog__scroll min-h-0 min-w-0 flex-1 space-y-4 overflow-x-hidden overflow-y-auto overscroll-contain px-5 pb-8 pt-4 [scrollbar-gutter:stable]"
          @wheel.stop
          @touchmove.stop
        >
          <ol class="grid gap-1.5 rounded-lg border border-dashed bg-muted/20 p-2 text-xs text-muted-foreground sm:grid-cols-2 xl:grid-cols-4">
            <li
              v-for="step in (isDedicatedCarbonRim ? carbonRimEditorSteps : productEditorSteps)"
              :key="step.no"
              class="flex min-w-0 items-center gap-2 rounded-md bg-background/70 px-2.5 py-1.5"
            >
              <span class="font-mono text-[10px] font-black text-primary">{{ step.no }}</span>
              <button v-if="isDedicatedCarbonRim" type="button" class="min-w-0 truncate text-left font-semibold text-foreground" @click="navigateToCarbonRimEditorSection(step.no)">{{ step.label }}</button>
              <strong v-else class="min-w-0 truncate text-foreground">{{ step.label }}</strong>
            </li>
          </ol>

          <div v-if="Object.keys(errors).length" role="alert" class="rounded-lg border border-destructive/40 p-3 text-sm text-destructive">
            <p v-for="(message, key) in errors" :key="key">{{ message }}</p>
          </div>
          <AdminFormSection id="product-basic-information" title="基础信息" :description="isDedicatedCarbonRim ? '填写品牌、名称与销售文案；车圈工程参数在下一分区维护。' : '填写商品自身的品牌、分类、识别和销售文案。'">
            <div class="grid gap-4 md:grid-cols-3">
              <AdminFormField label="商品名称" required :error="errors.name">
                <Input v-model="form.name" placeholder="请输入商品名称" @input="emit('clear-error', 'name')" />
              </AdminFormField>
              <AdminFormField label="商品品牌" description="品牌是商品一级数据，前台详情、Google SEO 和 Merchant 可直接读取。">
                <Select :model-value="brandSelectValue" @update:model-value="emit('product-brand-select', $event)">
                  <SelectTrigger class="w-full"><SelectValue placeholder="未设置品牌" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">未设置品牌</SelectItem>
                    <SelectItem
                      v-for="brand in brands"
                      :key="brand.id"
                      :value="String(brand.id)"
                      :disabled="brand.is_enabled === false && String(form.brand_id) !== String(brand.id)"
                    >
                      {{ brand.name }}{{ brand.is_enabled === false ? '（停用）' : '' }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <AdminFormField label="商品分类" :required="isDedicatedCarbonRim" :error="errors.product_category_id" :description="isDedicatedCarbonRim ? '碳圈商品固定归入此分类。' : '可选；商品可以不设置分类。'">
                <Input v-if="isDedicatedCarbonRim" :model-value="carbonRimSystemCategoryDisplayPath" disabled />
                <Select v-else :model-value="productCategorySelectValue" @update:model-value="emit('product-category-select', $event)">
                  <SelectTrigger class="w-full"><SelectValue placeholder="未设置分类" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">未设置分类</SelectItem>
                    <SelectItem
                      v-for="category in productCategories.filter(item => item.slug !== carbonRimSystemCategorySlug)"
                      :key="category.id"
                      :value="String(category.id)"
                      :disabled="category.is_enabled === false && String(form.product_category_id) !== String(category.id)"
                      :style="{ paddingLeft: `${0.75 + Math.max(0, Number(category.depth || 1) - 1) * 1}rem` }"
                    >
                      {{ category.name }}{{ category.is_enabled === false ? '（停用）' : '' }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <AdminFormField label="Slug" required :error="errors.slug">
                <Input v-model="form.slug" placeholder="例如 crystal-bracelet" @input="emit('clear-error', 'slug')" />
              </AdminFormField>
              <AdminFormField
                label="语言"
                required
                :error="errors.locale"
                :description="mode === 'edit' ? '编辑商品时语言已锁定；如需其他语言，请新建对应语种商品。' : ''"
              >
                <StorefrontLocaleSelect
                  v-model="form.locale"
                  :language-options="languageOptions"
                  :disabled="mode === 'edit'"
                  :locked="mode === 'edit'"
                  locked-title="商品语言已锁定"
                />
              </AdminFormField>
              <AdminFormField label="主基准币种" required :error="errors.currency" description="商品和 SKU 的录入金额使用这个基准币种；展示价由独立读模型刷新任务生成。">
                <Input v-model="form.currency" class="font-mono uppercase" disabled />
              </AdminFormField>
              <AdminFormField label="简短描述" class="md:col-span-3">
                <Textarea v-model="form.short_description" class="min-h-20" placeholder="用于列表和摘要展示" />
              </AdminFormField>
              <AdminFormField label="详细描述" class="md:col-span-3" :error="errors.description">
                <ProductDescriptionEditor
                  v-model="form.description"
                  @update:model-value="emit('clear-error', 'description')"
                />
              </AdminFormField>
            </div>
          </AdminFormSection>

          <ProductCustomsSection
            :form="form"
            :errors="errors"
            :customs-classifications="customsClassifications"
            @customs-classification-select="emit('customs-classification-select', $event)"
            @customs-classification-manual-edit="emit('customs-classification-manual-edit')"
            @clear-error="emit('clear-error', $event)"
          />

          <AdminFormSection
            v-if="!isDedicatedCarbonRim && !form.product_type"
            id="product-schema-facts"
            title="产品结构化信息"
            description="普通商品可选填写；用于前台 Product Schema。未知字段可以留空，不影响商品创建。"
          >
            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <AdminFormField label="型号" description="对应 Schema.org model">
                <Input v-model="form.generic_product_schema.model" placeholder="例如 ABC-100" />
              </AdminFormField>
              <AdminFormField label="材料" description="对应 Schema.org material">
                <Input v-model="form.generic_product_schema.material" placeholder="例如 Aluminum" />
              </AdminFormField>
              <AdminFormField label="颜色" description="对应 Schema.org color">
                <Input v-model="form.generic_product_schema.color" placeholder="例如 Black" />
              </AdminFormField>
              <AdminFormField label="尺寸" description="对应 Schema.org size">
                <Input v-model="form.generic_product_schema.size" placeholder="例如 700C" />
              </AdminFormField>
              <AdminFormField label="产品重量（克）" description="产品级净重；SKU 重量仍用于履约和运费">
                <Input v-model.number="form.generic_product_schema.weight_g" type="number" min="0" step="0.01" placeholder="可留空" />
              </AdminFormField>
            </div>
          </AdminFormSection>

          <CarbonRimDedicatedProductSpecificationFields v-if="isDedicatedCarbonRim" :specification="form.carbon_rim_specification" :locked="Boolean(form.parent_id)" :error="errors.carbon_rim_specification" />
          <AdminFormSection id="product-sales-variants" :title="isDedicatedCarbonRim ? '销售规格与库存' : 'SKU 与库存'" :description="isDedicatedCarbonRim ? '每行代表一种实际销售的孔数、钻孔和涂装组合。已保存组合保持稳定；停卖请取消启用。重量按单只录入。' : '按实际销售方式添加 SKU；名称、编码、价格、重量、线路计费和库存由每个 SKU 独立维护。'">
            <div class="min-w-0 rounded-lg border">
              <ProductVariantEditor
                :dedicated-carbon-rim="isDedicatedCarbonRim"
                :engineering-facts-locked="Boolean(form.parent_id)"
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

          <ProductMediaSection id="product-media-section"
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

          <ProductPublishingSection id="product-publishing-section"
            :form="form"
            :after-sales-templates="afterSalesTemplates"
            :packaging-templates="packagingTemplates"
            description="控制商品的公开状态和前台可见性。"
            notice="重量和线路计费模板按 SKU 独立维护，请在销售规格中设置。"
            @product-information-template-select="(...args) => emit('product-information-template-select', ...args)"
          />

        </div>

        <div v-if="isDedicatedCarbonRim" id="carbon-rim-save-review" class="shrink-0 border-t px-5 py-2 text-xs text-muted-foreground">
          单只碳纤维碟刹车圈 · {{ form.carbon_rim_specification?.model || '型号待填写' }} · {{ form.variants.length }} 个销售 SKU · {{ form.media.length }} 项媒体
        </div>
        <DialogFooter class="mx-0 mb-0 shrink-0 rounded-b-[32px] border-t bg-background/95 px-5 py-4 backdrop-blur">
          <Button type="button" variant="outline" @click="requestProductEditorOpenChange(false)">取消</Button>
          <Button type="submit" :disabled="submitting || supplierCostPending || uploadingMedia">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ supplierCostPending ? '请先处理成本资料' : submitting ? '保存中' : '保存商品' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import CarbonRimDedicatedProductSpecificationFields from './CarbonRimDedicatedProductSpecificationFields.vue'
import type { PropType } from 'vue'
import { carbonRimSystemCategorySlug } from '@/modules/product/carbonRimRequiredSystemCategorySelection'
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import ProductDescriptionEditor from '@/components/admin/product/ProductDescriptionEditor.vue'
import ProductCustomsSection from '@/components/admin/product/ProductCustomsSection.vue'
import ProductMediaSection from '@/components/admin/product/ProductMediaSection.vue'
import ProductProfitabilitySection from '@/components/admin/product/ProductProfitabilitySection.vue'
import ProductPublishingSection from '@/components/admin/product/ProductPublishingSection.vue'
import ProductVariantEditor from '@/components/admin/product/ProductVariantEditor.vue'
import type { ProductSupplierCostProfitDraft } from '@/composables/product/useProductSupplierCostProfitDraft'
import type { ProductFormRecord } from '@/modules/product/productEditorTypes'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

interface ProductBrandRecord {
  id: number
  name: string
  slug: string
  is_enabled?: boolean
}

interface ProductCategoryRecord {
  id: number
  slug: string
  name: string
  depth: number
  is_enabled?: boolean
}

interface ShippingTemplateRecord {
  id: number
  name: string
  enabled?: boolean
}

interface InformationTemplateRecord {
  id: number
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

const scrollContainer = ref<HTMLElement | null>(null)

const handleOpenAutoFocus = (event: Event) => {
  event.preventDefault()
  void nextTick(() => {
    requestAnimationFrame(() => {
      scrollContainer.value?.scrollTo({ top: 0, behavior: 'auto' })
    })
  })
}

const props = defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  submitting: { type: Boolean, default: false },
  form: { type: Object as PropType<ProductFormRecord>, required: true },
  errors: { type: Object as PropType<Record<string, string>>, default: () => ({}) },
  brands: { type: Array as PropType<ProductBrandRecord[]>, default: () => [] },
  productCategories: { type: Array as PropType<ProductCategoryRecord[]>, default: () => [] },
  defaultVariantIndex: { type: Number, default: 0 },
  productCategorySelectValue: { type: String, default: '__none__' },
  carbonRimSystemCategoryDisplayPath: { type: String, required: true },
  brandSelectValue: { type: String, default: '__none__' },
  shippingTemplates: { type: Array as PropType<ShippingTemplateRecord[]>, default: () => [] },
  afterSalesTemplates: { type: Array as PropType<InformationTemplateRecord[]>, default: () => [] },
  packagingTemplates: { type: Array as PropType<InformationTemplateRecord[]>, default: () => [] },
  customsClassifications: { type: Array as PropType<CustomsClassificationRecord[]>, default: () => [] },
  uploadingMedia: { type: Boolean, default: false },
  supplierCostVisible: { type: Boolean, default: false },
  supplierCostCanEdit: { type: Boolean, default: false },
  supplierCostLoading: { type: Boolean, default: false },
  supplierCostSaving: { type: Boolean, default: false },
  supplierCostPending: { type: Boolean, default: false },
  supplierCostError: { type: String, default: '' },
  supplierCostLastSavedAt: { type: String, default: '' },
  supplierCostDrafts: { type: Array as PropType<ProductSupplierCostProfitDraft[]>, default: () => [] },
  languageOptions: { type: Array as PropType<LanguageOption[]>, default: () => [] },
})

const isDedicatedCarbonRim = computed(() => props.form.product_type === 'carbon_rim')
const productEditorSteps = [
  { no: '01', label: '商品资料' }, { no: '02', label: '销售 SKU' },
  { no: '03', label: '图片视频' }, { no: '04', label: '物流与售后' },
]
const carbonRimEditorSteps = [
  { no: '01', label: '商品资料' }, { no: '02', label: '车圈规格' },
  { no: '03', label: '销售规格' }, { no: '04', label: '图片视频' },
  { no: '05', label: '物流与售后' }, { no: '06', label: '核对保存' },
]
const navigateToCarbonRimEditorSection = (section: string): void => {
  const ids: Record<string, string> = { '01': 'product-basic-information', '02': 'carbon-rim-specifications', '03': 'product-sales-variants', '04': 'product-media-section', '05': 'product-publishing-section', '06': 'carbon-rim-save-review' }
  document.getElementById(ids[section] || '')?.scrollIntoView({ block: 'start', behavior: 'smooth' })
}
const initialProductEditorValues = ref('')
watch(() => props.open, async (open) => { if (open) { await nextTick(); initialProductEditorValues.value = JSON.stringify(props.form) } })
const requestProductEditorOpenChange = (open: boolean): void => {
  if (!open && (props.submitting || props.uploadingMedia)) return
  if (!open && isDedicatedCarbonRim.value && initialProductEditorValues.value !== JSON.stringify(props.form) && !window.confirm('有未保存的修改，确定关闭车圈编辑窗口？')) return
  emit('update:open', open)
}

const emit = defineEmits([
  'update:open',
  'submit',
  'clear-error',
  'product-category-select',
  'product-brand-select',
  'product-information-template-select',
  'customs-classification-select',
  'customs-classification-manual-edit',
  'add-variant',
  'remove-variant',
  'set-default-variant',
  'set-variant-active',
  'upload-media',
  'add-media-url',
  'set-primary-media',
  'move-media',
  'remove-media',
  'retry-supplier-cost',
])
</script>

<style scoped>
:global([data-product-editor-dialog]) {
  display: flex !important;
  overflow: hidden !important;
}

:global([data-product-editor-dialog] > form) {
  min-height: 0;
}

.product-editor-dialog__scroll {
  max-height: 100%;
}
</style>


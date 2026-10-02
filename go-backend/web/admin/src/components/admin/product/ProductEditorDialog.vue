<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      size="full"
      data-product-editor-dialog
      class="!flex h-[94dvh] max-h-[calc(100dvh-1rem)] !w-[95dvw] !max-w-[95dvw] flex-col gap-0 overflow-hidden p-0"
      style="overflow: hidden;"
      @open-auto-focus="handleOpenAutoFocus"
    >
      <form class="flex min-h-0 min-w-0 flex-1 flex-col" @submit.prevent="emit('submit')">
        <DialogHeader class="shrink-0 border-b px-5 py-4 pr-12">
          <DialogTitle>{{ mode === 'create' ? '添加商品' : '编辑商品' }}</DialogTitle>
          <DialogDescription>先录入商品基础识别信息，再绑定商品规格模板；模板决定下方商品字段和 SKU 选项列。</DialogDescription>
        </DialogHeader>

        <div
          ref="scrollContainer"
          class="product-editor-dialog__scroll min-h-0 min-w-0 flex-1 space-y-4 overflow-x-hidden overflow-y-auto overscroll-contain px-5 pb-8 pt-4 [scrollbar-gutter:stable]"
          @wheel.stop
          @touchmove.stop
        >
          <ol class="grid gap-1.5 rounded-lg border border-dashed bg-muted/20 p-2 text-xs text-muted-foreground sm:grid-cols-2 xl:grid-cols-4">
            <li
              v-for="step in editorSteps"
              :key="step.no"
              class="flex min-w-0 items-center gap-2 rounded-md bg-background/70 px-2.5 py-1.5"
            >
              <span class="font-mono text-[10px] font-black text-primary">{{ step.no }}</span>
              <strong class="min-w-0 truncate text-foreground">{{ step.label }}</strong>
            </li>
          </ol>

          <AdminFormSection title="基础信息" description="这里不放规格字段；规格字段必须先通过商品规格模板统一定义，再在下方录入具体值。">
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
              <AdminFormField label="商品分类" description="可选；商品可以不设置分类。">
                <Select :model-value="productCategorySelectValue" @update:model-value="emit('product-category-select', $event)">
                  <SelectTrigger class="w-full"><SelectValue placeholder="未设置分类" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">未设置分类</SelectItem>
                    <SelectItem
                      v-for="category in productCategories"
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

          <AdminFormSection title="绑定商品规格模板" description="选择模板后，下方才会出现对应的商品参数字段和 SKU 选项列。">
            <div class="grid gap-3 2xl:grid-cols-[minmax(20rem,0.68fr)_minmax(0,1.32fr)]">
              <div class="grid gap-3 rounded-xl border bg-muted/20 p-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end 2xl:block 2xl:space-y-3">
                <AdminFormField label="商品规格模板">
                  <Select :model-value="productSpecTemplateSelectValue" @update:model-value="emit('product-spec-template-select', $event)">
                    <SelectTrigger class="w-full"><SelectValue placeholder="请选择商品规格模板" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__none__">未选择模板</SelectItem>
                      <SelectItem v-for="type in productSpecTemplates" :key="type.id" :value="String(type.id)">
                        {{ type.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </AdminFormField>
                <Button type="button" variant="outline" size="sm" as-child>
                  <RouterLink to="/catalog/templates">
                    <Tags class="size-3.5" />
                    维护商品规格模板
                  </RouterLink>
                </Button>
                <p class="text-xs leading-5 text-muted-foreground lg:col-span-2 2xl:col-span-1">
                  模板只定义字段结构，当前商品的重量、价格、库存和尺寸值仍在这里维护。
                </p>
              </div>

              <div class="rounded-xl border border-dashed bg-muted/20 p-3">
                <div v-if="selectedProductSpecTemplate" class="grid gap-3 xl:grid-cols-[minmax(14rem,0.72fr)_minmax(0,1.28fr)]">
                  <div class="min-w-0 space-y-2">
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="min-w-0 truncate text-sm font-bold">{{ selectedProductSpecTemplate.name }}</span>
                      <span class="rounded-full bg-background px-2 py-0.5 font-mono text-[10px] text-muted-foreground">
                        {{ selectedProductSpecTemplate.slug }}
                      </span>
                    </div>
                    <p v-if="selectedProductSpecTemplate.description" class="line-clamp-2 text-xs leading-5 text-muted-foreground">
                      {{ selectedProductSpecTemplate.description }}
                    </p>
                   <div class="flex flex-wrap gap-2">
                      <span class="rounded-full bg-background px-2.5 py-1 text-[11px] font-black text-foreground">
                        商品字段 {{ selectedSpecDefinitions.length }}
                      </span>
                      <span class="rounded-full bg-background px-2.5 py-1 text-[11px] font-black text-foreground">
                        SKU 字段 {{ variantSpecDefinitions.length }}
                      </span>
                    </div>
                    <div v-if="templateScopedValuesTouched" class="rounded-xl border border-amber-500/25 bg-amber-500/10 px-3 py-2 text-xs leading-5 text-amber-800 dark:text-amber-200">
                      切换模板会清空旧模板下的字段值和 SKU 选项值；SKU 价格、重量、库存和商品媒体会保留。
                     </div>
                     <Button
                       v-if="mode === 'edit' && form.id"
                       type="button"
                       variant="outline"
                       size="sm"
                       :disabled="templateSyncLoading"
                       @click="emit('preview-template-sync')"
                     >
                       <LoaderCircle v-if="templateSyncLoading" class="size-3.5 animate-spin" />
                       <Tags v-else class="size-3.5" />
                       {{ templateSyncLoading ? '读取差异中' : '预览模板同步' }}
                     </Button>
                   </div>

                  <div class="grid min-w-0 gap-2 lg:grid-cols-2">
                    <div class="min-w-0 rounded-xl bg-background/70 p-2.5">
                      <div class="flex items-center justify-between gap-2">
                        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">商品参数字段</span>
                        <span class="font-mono text-xs font-black text-foreground">{{ selectedSpecDefinitions.length }}</span>
                      </div>
                      <div v-if="selectedSpecDefinitions.length" class="mt-2 flex max-h-20 flex-wrap gap-1.5 overflow-y-auto pr-1 [scrollbar-gutter:stable]">
                        <span
                          v-for="spec in selectedSpecDefinitions"
                          :key="`product-${spec.id || spec.slug}`"
                          class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground"
                        >
                          {{ getSpecLabel(spec) }}
                        </span>
                      </div>
                      <p v-else class="mt-2 text-xs text-muted-foreground">该模板没有商品级参数字段。</p>
                    </div>
                    <div class="min-w-0 rounded-xl bg-background/70 p-2.5">
                      <div class="flex items-center justify-between gap-2">
                        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SKU 选项字段</span>
                        <span class="font-mono text-xs font-black text-primary">{{ variantSpecDefinitions.length }}</span>
                      </div>
                      <div v-if="variantSpecDefinitions.length" class="mt-2 flex max-h-20 flex-wrap gap-1.5 overflow-y-auto pr-1 [scrollbar-gutter:stable]">
                        <span
                          v-for="spec in variantSpecDefinitions"
                          :key="`variant-${spec.id || spec.slug}`"
                          class="rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-medium text-primary"
                        >
                          {{ getSpecLabel(spec) }}
                        </span>
                      </div>
                      <p v-else class="mt-2 text-xs text-muted-foreground">没有 SKU 选项字段，仅维护默认 SKU。</p>
                    </div>
                  </div>
                </div>
                <p v-else class="text-xs leading-5 text-muted-foreground">
                  未选择模板时，下方不会出现商品参数字段，也不会给 SKU 生成选项列。建议先在“商品规格模板”页面创建车圈、车架等模板，再回到这里绑定。
                </p>
              </div>
            </div>
          </AdminFormSection>

          <AdminFormSection
            title="商品参数（来自模板）"
            :description="selectedProductSpecTemplate?.slug === 'schwalbe_tire' ? '每个 Article No. 单独建商品，并保留一个默认销售 SKU；不同 Article No. 或尺寸要分成不同商品，不能做成同一商品的 SKU 变体。官网未提供的字段留空。' : selectedSpecDefinitions.length ? '这里填写当前商品自己的参数值；字段来源于已绑定商品规格模板。' : '当前模板没有商品级参数字段；可以直接继续维护 SKU。'"
          >
            <div v-if="selectedProductSpecTemplate?.slug === 'schwalbe_tire'" class="mb-4 rounded-xl border border-primary/20 bg-primary/5 p-3">
              <AdminFormField label="Schwalbe 型号 / Article No.">
                <Select
                  :model-value="selectedSchwalbeTireArticleNo || '__none__'"
                  :disabled="schwalbeTireCatalogLoading || schwalbeTireCatalogItems.length === 0"
                  @update:model-value="emit('schwalbe-tire-model-select', $event)"
                >
                  <SelectTrigger class="w-full"><SelectValue :placeholder="schwalbeTireCatalogLoading ? '正在读取型号目录…' : '请选择官网型号'" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">请选择型号</SelectItem>
                    <SelectItem v-for="item in schwalbeTireCatalogItems" :key="item.article_no" :value="item.article_no">
                      {{ item.model_name }} · {{ item.etrto }} · {{ item.article_no }}{{ item.product_exists ? '（已有销售商品）' : '' }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <p v-if="schwalbeTireCatalogLoading" class="mt-2 text-xs text-muted-foreground">正在读取 Schwalbe 全谱系型号。</p>
              <p v-else-if="schwalbeTireCatalogError" class="mt-2 flex items-center gap-2 text-xs text-destructive">
                {{ schwalbeTireCatalogError }}
                <Button type="button" variant="link" size="sm" class="h-auto p-0" @click="emit('retry-schwalbe-tire-catalog')">重试</Button>
              </p>
              <p v-else-if="schwalbeTireCatalogItems.length === 0" class="mt-2 text-xs text-muted-foreground">目录中暂无型号数据。目录数据由官方来源导入，不在此处逐条录入。</p>
              <p v-else class="mt-2 text-xs text-muted-foreground">选择型号后会自动填充下方 19 项官网字段；是否存在销售商品仅作状态提示。</p>
            </div>
            <div v-if="selectedSpecDefinitions.length" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-6">
              <AdminFormField
                v-for="spec in selectedSpecDefinitions"
                :key="spec.id || spec.slug"
                class="min-w-0"
                :label="getSpecLabel(spec)"
                :required="spec.is_required"
                :error="errors[`spec:${spec.slug}`]"
              >
                <template v-if="spec.field_type === 'number'">
                  <Input
                    v-model.number="form.specs[spec.slug]"
                    type="number"
                    :min="selectedProductSpecTemplate?.slug === 'schwalbe_tire' ? (spec.slug === 'epi' ? '1' : spec.unit === 'bar' ? '0.1' : '0.01') : '0'"
                    :step="selectedProductSpecTemplate?.slug === 'schwalbe_tire' ? (spec.slug === 'epi' ? '1' : spec.unit === 'bar' ? '0.1' : '0.01') : (spec.unit === 'bar' ? '0.1' : '1')"
                    @input="emit('clear-error', `spec:${spec.slug}`)"
                  />
                  <p
                    v-if="selectedProductSpecTemplate?.slug === 'schwalbe_tire' && spec.slug === 'epi'"
                    class="mt-1 text-xs leading-4 text-muted-foreground"
                  >
                    官网写成 2x67 等双层胎体时，只填写单层 EPI：67。
                  </p>
                </template>
                <Select
                  v-else-if="spec.field_type === 'select' && parseSpecOptions(spec).length"
                  :model-value="specSelectValue(form.specs[spec.slug])"
                  @update:model-value="emit('set-spec-select-value', spec.slug, $event)"
                >
                  <SelectTrigger class="w-full"><SelectValue placeholder="请选择" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__empty__">未设置</SelectItem>
                    <SelectItem v-for="option in parseSpecOptions(spec)" :key="String(option)" :value="String(option)">
                      {{ formatSpecOption(option) }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <Input
                  v-else-if="spec.field_type === 'select'"
                  v-model="form.specs[spec.slug]"
                  :placeholder="`请输入${spec.name}（可动态录入）`"
                  @input="emit('clear-error', `spec:${spec.slug}`)"
                />
                <div v-else-if="spec.field_type === 'boolean'" class="flex h-9 items-center gap-2">
                  <Switch v-model="form.specs[spec.slug]" :aria-label="spec.name" />
 <span class="text-xs text-muted-foreground">{{ form.specs[spec.slug] ? '是': '否'}}</span>
                </div>
                <Input
                  v-else
                  v-model="form.specs[spec.slug]"
                  :placeholder="`请输入${spec.name}`"
                  @input="emit('clear-error', `spec:${spec.slug}`)"
                />
              </AdminFormField>
            </div>
            <div v-else class="rounded-xl border border-dashed bg-muted/20 px-4 py-5 text-center text-xs text-muted-foreground">
              {{ selectedProductSpecTemplate ? '这个模板没有商品级字段。' : '选择商品规格模板后，商品级字段会显示在这里。' }}
            </div>
          </AdminFormSection>

          <AdminFormSection title="SKU 变体矩阵" description="SKU 选项列来自商品规格模板；价格、重量和库存永远按每一行 SKU 独立维护，前台按用户选中的 SKU 显示对应重量。">
            <div class="min-w-0 rounded-lg border">
              <ProductVariantEditor
                :variants="form.variants"
                :spec-definitions="variantSpecDefinitions"
                :option-values="form.variant_option_values"
                :option-value-relations="form.option_value_relations"
                :product-id="form.id"
                :custom-option-definitions="customOptionDefinitions"
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
            <p v-if="errors.option_value_relations" class="mt-2 text-xs font-medium text-destructive">{{ errors.option_value_relations }}</p>
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
            :media-items="form.media"
            :variants="form.variants"
            :variant-option-values="form.variant_option_values"
            :spec-definitions="variantSpecDefinitions"
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
            :form="form"
            :shipping-templates="shippingTemplates"
            :after-sales-templates="afterSalesTemplates"
            :packaging-templates="packagingTemplates"
            description="控制商品的公开状态和前台可见性。"
            notice="重量现在只在 SKU 变体里维护，前台会按当前选中的 SKU 显示对应重量。"
            @product-shipping-template-select="emit('product-shipping-template-select', $event)"
            @product-information-template-select="(...args) => emit('product-information-template-select', ...args)"
          />

        </div>

        <DialogFooter class="mx-0 mb-0 shrink-0 rounded-b-[32px] border-t bg-background/95 px-5 py-4 backdrop-blur">
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting || supplierCostPending">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ supplierCostPending ? '请先处理成本资料' : submitting ? '保存中' : '保存商品' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
  <ProductTemplateSyncDialog
    :open="templateSyncDialogVisible"
    :diff="templateSyncDiff"
    :applying="templateSyncApplying"
    @update:open="emit('update-template-sync-open', $event)"
    @confirm="emit('confirm-template-sync')"
  />
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue'
import type { PropType } from 'vue'
import { RouterLink } from 'vue-router'
import { LoaderCircle, Tags } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import ProductDescriptionEditor from '@/components/admin/product/ProductDescriptionEditor.vue'
import ProductCustomsSection from '@/components/admin/product/ProductCustomsSection.vue'
import ProductMediaSection from '@/components/admin/product/ProductMediaSection.vue'
import ProductProfitabilitySection from '@/components/admin/product/ProductProfitabilitySection.vue'
import ProductPublishingSection from '@/components/admin/product/ProductPublishingSection.vue'
import ProductTemplateSyncDialog from '@/components/admin/product/ProductTemplateSyncDialog.vue'
import ProductVariantEditor from '@/components/admin/product/ProductVariantEditor.vue'
import type { SchwalbeTireCatalogItem } from '@/api/schwalbeTireCatalog'
import type { ProductSupplierCostProfitDraft } from '@/composables/product/useProductSupplierCostProfitDraft'
import type { ProductFormRecord, ProductTemplateSyncDiff } from '@/modules/product/productEditorTypes'
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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

const editorSteps = [
  { no: '01', label: '基础识别' },
  { no: '02', label: '绑定模板' },
  { no: '03', label: '填写参数' },
  { no: '04', label: '维护 SKU' },
  { no: '05', label: '商品成本' },
]

interface ProductSpecDefinition {
  id?: number
  slug: string
  name: string
  field_type: string
  presentation?: string
  is_required?: boolean
  unit?: string
  option_items?: Array<{ value_key?: string; default_label?: string }>
}

interface ProductSpecTemplateRecord {
  id: number
  name: string
  slug: string
  description?: string
  spec_definitions?: ProductSpecDefinition[]
}

interface ProductBrandRecord {
  id: number
  name: string
  slug: string
  is_enabled?: boolean
}

interface ProductCategoryRecord {
  id: number
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
  status?: string
}

interface LanguageOption {
  value: string
  label: string
}

type ProductFormValue = string | number | boolean | null | undefined

const scrollContainer = ref<HTMLElement | null>(null)

const handleOpenAutoFocus = (event: Event) => {
  event.preventDefault()
  void nextTick(() => {
    requestAnimationFrame(() => {
      scrollContainer.value?.scrollTo({ top: 0, behavior: 'auto' })
    })
  })
}

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  submitting: { type: Boolean, default: false },
  form: { type: Object as PropType<ProductFormRecord>, required: true },
  errors: { type: Object as PropType<Record<string, string>>, default: () => ({}) },
  productSpecTemplates: { type: Array as PropType<ProductSpecTemplateRecord[]>, default: () => [] },
  brands: { type: Array as PropType<ProductBrandRecord[]>, default: () => [] },
  productCategories: { type: Array as PropType<ProductCategoryRecord[]>, default: () => [] },
  selectedProductSpecTemplate: { type: Object as PropType<ProductSpecTemplateRecord | null>, default: null },
  schwalbeTireCatalogItems: { type: Array as PropType<SchwalbeTireCatalogItem[]>, default: () => [] },
  schwalbeTireCatalogLoading: { type: Boolean, default: false },
  schwalbeTireCatalogError: { type: String, default: '' },
  selectedSchwalbeTireArticleNo: { type: String, default: '' },
  selectedSpecDefinitions: { type: Array as PropType<ProductSpecDefinition[]>, default: () => [] },
  variantSpecDefinitions: { type: Array as PropType<ProductSpecDefinition[]>, default: () => [] },
  customOptionDefinitions: { type: Array as PropType<ProductSpecDefinition[]>, default: () => [] },
  defaultVariantIndex: { type: Number, default: 0 },
  productSpecTemplateSelectValue: { type: String, default: '__none__' },
  productCategorySelectValue: { type: String, default: '__none__' },
  brandSelectValue: { type: String, default: '__none__' },
  shippingTemplates: { type: Array as PropType<ShippingTemplateRecord[]>, default: () => [] },
  afterSalesTemplates: { type: Array as PropType<InformationTemplateRecord[]>, default: () => [] },
  packagingTemplates: { type: Array as PropType<InformationTemplateRecord[]>, default: () => [] },
  customsClassifications: { type: Array as PropType<CustomsClassificationRecord[]>, default: () => [] },
  templateScopedValuesTouched: { type: Boolean, default: false },
  templateSyncDialogVisible: { type: Boolean, default: false },
  templateSyncDiff: { type: Object as PropType<ProductTemplateSyncDiff | null>, default: null },
  templateSyncLoading: { type: Boolean, default: false },
  templateSyncApplying: { type: Boolean, default: false },
  uploadingMedia: { type: Boolean, default: false },
  supplierCostVisible: { type: Boolean, default: false },
  supplierCostCanEdit: { type: Boolean, default: false },
  supplierCostLoading: { type: Boolean, default: false },
  supplierCostSaving: { type: Boolean, default: false },
  supplierCostPending: { type: Boolean, default: false },
  supplierCostError: { type: String, default: '' },
  supplierCostLastSavedAt: { type: String, default: '' },
  supplierCostDrafts: { type: Array as PropType<ProductSupplierCostProfitDraft[]>, default: () => [] },
  parseSpecOptions: { type: Function as PropType<(spec: ProductSpecDefinition) => ProductFormValue[]>, required: true },
  formatSpecOption: { type: Function as PropType<(option: ProductFormValue) => string>, required: true },
  getSpecLabel: { type: Function as PropType<(spec: ProductSpecDefinition) => string>, required: true },
  specSelectValue: { type: Function as PropType<(value: ProductFormValue) => string>, required: true },
  languageOptions: { type: Array as PropType<LanguageOption[]>, default: () => [] },
})

const emit = defineEmits([
  'update:open',
  'submit',
  'clear-error',
  'product-spec-template-select',
  'schwalbe-tire-model-select',
  'retry-schwalbe-tire-catalog',
  'preview-template-sync',
  'update-template-sync-open',
  'confirm-template-sync',
  'product-category-select',
  'product-brand-select',
  'product-shipping-template-select',
  'product-information-template-select',
  'customs-classification-select',
  'customs-classification-manual-edit',
  'set-spec-select-value',
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


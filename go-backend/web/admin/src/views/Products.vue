<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品管理" description="管理商品资料、规格、SKU 变体和库存状态">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/product-types">
            <Tags class="size-4" />
            产品类型
          </RouterLink>
        </Button>
        <Button v-if="hasPermission('product:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加商品
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_repeat(3,minmax(140px,0.7fr))_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1.5">
          <span class="text-xs font-medium text-muted-foreground">搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="商品名称、SKU 或描述" />
          </div>
        </label>

        <FilterSelect v-model="filters.status" label="状态" :options="statusFilterOptions" />
        <FilterSelect v-model="filters.locale" label="语言" :options="localeFilterOptions" />
        <FilterSelect v-model="filters.featured" label="精选" :options="featuredFilterOptions" />

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9">
            <Search class="size-4" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9" @click="resetFilters">
            <RotateCcw class="size-4" />
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" :batch-visible="selectedProducts.length > 0">
      <template #batch>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-medium">已选择 {{ selectedProducts.length }} 个商品</span>
          <div class="flex flex-wrap gap-2">
            <Button v-if="hasPermission('product:edit')" size="sm" @click="requestBatchStatus('active')">
              <CircleCheck class="size-3.5" />
              批量上架
            </Button>
            <Button v-if="hasPermission('product:edit')" variant="outline" size="sm" @click="requestBatchStatus('inactive')">
              <CircleOff class="size-3.5" />
              批量下架
            </Button>
            <Button v-if="hasPermission('product:delete')" variant="destructive" size="sm" @click="requestBatchDelete">
              <Trash2 class="size-3.5" />
              批量删除
            </Button>
          </div>
        </div>
      </template>

      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-11">
              <Checkbox
                :model-value="selectionState"
                aria-label="选择当前页商品"
                @update:model-value="toggleAllProducts"
              />
            </TableHead>
            <TableHead class="w-16">ID</TableHead>
            <TableHead class="w-36">SKU</TableHead>
            <TableHead>商品名称</TableHead>
            <TableHead class="w-32">价格</TableHead>
            <TableHead class="w-24">库存</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-20 text-center">精选</TableHead>
            <TableHead class="w-20">语言</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="products.length === 0" :colspan="11">
            <div class="flex flex-col items-center text-muted-foreground">
              <PackageOpen class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无商品</span>
            </div>
          </TableEmpty>

          <TableRow v-for="product in products" :key="product.id">
            <TableCell>
              <Checkbox
                :model-value="isSelected(product.id)"
                :aria-label="`选择商品 ${product.name}`"
                @update:model-value="toggleProduct(product, $event)"
              />
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ product.id }}</TableCell>
            <TableCell class="font-mono text-xs">{{ product.sku || '-' }}</TableCell>
            <TableCell class="max-w-72 truncate font-medium">{{ product.name }}</TableCell>
            <TableCell>
              <div class="flex items-baseline gap-1.5 tabular-nums">
                <span v-if="product.sale_price" class="font-semibold text-destructive">¥{{ formatMoney(product.sale_price) }}</span>
                <span :class="product.sale_price ? 'text-xs text-muted-foreground line-through' : 'font-medium'">
                  ¥{{ formatMoney(product.price) }}
                </span>
              </div>
            </TableCell>
            <TableCell>
              <AdminStatusBadge v-if="Number(product.stock) === 0" tone="coral">缺货</AdminStatusBadge>
              <AdminStatusBadge v-else-if="Number(product.stock) < 10" tone="amber">{{ product.stock }}</AdminStatusBadge>
              <span v-else class="tabular-nums">{{ product.stock }}</span>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(product.status)">{{ getStatusName(product.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="text-center">
              <Star v-if="product.featured" class="mx-auto size-4 fill-amber-400 text-amber-500" aria-label="精选商品" />
              <span v-else class="text-muted-foreground">-</span>
            </TableCell>
            <TableCell>{{ localeName(product.locale) }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(product.created_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理商品 ${product.name}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem v-if="hasPermission('product:edit')" @select="showEditDialog(product)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('product:edit')" @select="requestToggleStatus(product)">
                    <CircleCheck v-if="product.status !== 'active'" class="size-4" />
                    <CircleOff v-else class="size-4" />
                    {{ product.status === 'active' ? '下架' : '上架' }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('product:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('product:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(product)"
                  >
                    <Trash2 class="size-4" />
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="dialogVisible">
      <DialogContent
        size="full"
        class="h-[94dvh] max-h-[calc(100dvh-1rem)] overflow-hidden p-0"
        @open-auto-focus.prevent
      >
        <form class="flex h-full min-h-0 min-w-0 flex-col" @submit.prevent="submitForm">
          <DialogHeader class="shrink-0 border-b px-5 py-4 pr-12">
            <DialogTitle>{{ dialogMode === 'create' ? '添加商品' : '编辑商品' }}</DialogTitle>
            <DialogDescription>维护商品基础资料、规格模板和 SKU 级价格库存。</DialogDescription>
          </DialogHeader>

          <div class="min-h-0 min-w-0 flex-1 space-y-7 overflow-x-hidden overflow-y-auto overscroll-contain px-5 py-5 [scrollbar-gutter:stable]">
            <FormSection title="基础信息" description="用于商品识别、展示和多语言归属。">
              <div class="grid gap-4 md:grid-cols-2">
                <FormField label="商品名称" required :error="formErrors.name">
                  <Input v-model="productForm.name" placeholder="请输入商品名称" @input="clearFieldError('name')" />
                </FormField>
                <FormField label="Slug" required :error="formErrors.slug">
                  <Input v-model="productForm.slug" placeholder="例如 crystal-bracelet" @input="clearFieldError('slug')" />
                </FormField>
                <FormField label="语言" required>
                  <Select v-model="productForm.locale">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="zh">中文</SelectItem>
                      <SelectItem value="en">English</SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField label="产品类型">
                  <Select :model-value="productTypeSelectValue" @update:model-value="handleProductTypeSelect">
                    <SelectTrigger class="w-full"><SelectValue placeholder="请选择产品类型" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__none__">未选择</SelectItem>
                      <SelectItem v-for="type in productTypes" :key="type.id" :value="String(type.id)">
                        {{ type.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField label="简短描述" class="md:col-span-2">
                  <Textarea v-model="productForm.short_description" class="min-h-20" placeholder="用于列表和摘要展示" />
                </FormField>
                <FormField label="详细描述" class="md:col-span-2">
                  <Textarea v-model="productForm.description" class="min-h-28" placeholder="请输入商品详细描述" />
                </FormField>
              </div>
            </FormSection>

            <FormSection
              v-if="selectedSpecDefinitions.length"
              title="规格模板"
              description="这些规格来自所选产品类型，保存为商品级属性。"
            >
              <div class="grid gap-4 md:grid-cols-2">
                <FormField
                  v-for="spec in selectedSpecDefinitions"
                  :key="spec.id"
                  :label="getSpecLabel(spec)"
                  :required="spec.is_required"
                  :error="formErrors[`spec:${spec.slug}`]"
                >
                  <Input
                    v-if="spec.field_type === 'number'"
                    v-model.number="productForm.specs[spec.slug]"
                    type="number"
                    min="0"
                    @input="clearFieldError(`spec:${spec.slug}`)"
                  />
                  <Select
                    v-else-if="spec.field_type === 'select'"
                    :model-value="specSelectValue(productForm.specs[spec.slug])"
                    @update:model-value="setSpecSelectValue(spec.slug, $event)"
                  >
                    <SelectTrigger class="w-full"><SelectValue placeholder="请选择" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__empty__">未设置</SelectItem>
                      <SelectItem v-for="option in parseSpecOptions(spec)" :key="String(option)" :value="String(option)">
                        {{ formatSpecOption(option) }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <div v-else-if="spec.field_type === 'boolean'" class="flex h-9 items-center gap-2">
                    <Switch v-model="productForm.specs[spec.slug]" :aria-label="spec.name" />
                    <span class="text-xs text-muted-foreground">{{ productForm.specs[spec.slug] ? '是' : '否' }}</span>
                  </div>
                  <Input
                    v-else
                    v-model="productForm.specs[spec.slug]"
                    :placeholder="`请输入${spec.name}`"
                    @input="clearFieldError(`spec:${spec.slug}`)"
                  />
                </FormField>
              </div>
            </FormSection>

            <FormSection title="SKU 变体矩阵" description="每个商品至少保留一个变体，价格和库存均以变体为准。">
              <div class="min-w-0 rounded-lg border">
                <ProductVariantEditor
                  :variants="productForm.variants"
                  :spec-definitions="variantSpecDefinitions"
                  :default-index="defaultVariantIndex"
                  class="min-w-0 p-3"
                  @add="addVariant"
                  @remove="removeVariant"
                  @set-default="setDefaultVariant"
                />
              </div>
              <p v-if="formErrors.variants" class="mt-2 text-xs font-medium text-destructive">{{ formErrors.variants }}</p>
            </FormSection>

            <FormSection title="发布设置" description="控制商品的公开状态和前台可见性。">
              <div class="grid gap-4 md:grid-cols-2">
                <div class="md:col-span-2 rounded-lg border bg-muted/30 px-3 py-2.5 text-xs text-muted-foreground">
                  重量现在只在 SKU 变体里维护，前台会按当前选中的 SKU 显示对应重量。
                </div>
                <FormField label="状态" required>
                  <Select v-model="productForm.status">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="active">在售</SelectItem>
                      <SelectItem value="inactive">下架</SelectItem>
                      <SelectItem value="out_of_stock">缺货</SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <div class="flex items-center justify-between gap-4 rounded-lg border px-3 py-2.5 md:col-span-2">
                  <div>
                    <Label for="product-featured">精选商品</Label>
                    <p class="mt-0.5 text-xs text-muted-foreground">在前台精选区域优先展示该商品。</p>
                  </div>
                  <Switch id="product-featured" v-model="productForm.featured" />
                </div>
              </div>
            </FormSection>

            <FormSection title="SEO" description="可选的搜索结果标题和描述。">
              <div class="grid gap-4">
                <FormField label="SEO 标题">
                  <Input v-model="productForm.meta_title" placeholder="请输入 SEO 标题" />
                </FormField>
                <FormField label="SEO 描述">
                  <Textarea v-model="productForm.meta_description" class="min-h-20" placeholder="请输入 SEO 描述" />
                </FormField>
              </div>
            </FormSection>
          </div>

          <DialogFooter class="mx-0 mb-0 shrink-0 rounded-b-lg border-t bg-background px-5 py-4">
            <Button type="button" variant="outline" @click="dialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '保存中' : '保存商品' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

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

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Boxes,
  CircleCheck,
  CircleOff,
  LoaderCircle,
  MoreHorizontal,
  PackageOpen,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Star,
  Tags,
  Trash2,
  TriangleAlert
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import ProductVariantEditor from '@/components/admin/product/ProductVariantEditor.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const FilterSelect = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    options: { type: Array, required: true }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'space-y-1.5' }, [
      h('span', { class: 'text-xs font-medium text-muted-foreground' }, props.label),
      h(Select, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value) => emit('update:modelValue', value)
      }, {
        default: () => [
          h(SelectTrigger, { class: 'h-9 w-full' }, { default: () => h(SelectValue) }),
          h(SelectContent, {}, {
            default: () => props.options.map((option) => h(SelectItem, { value: option.value }, { default: () => option.label }))
          })
        ]
      })
    ])
  }
})

const FormSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'space-y-4 border-t pt-6 first:border-t-0 first:pt-0' }, [
      h('div', { class: 'space-y-1' }, [
        h('h3', { class: 'text-sm font-semibold' }, props.title),
        props.description ? h('p', { class: 'max-w-2xl text-xs leading-5 text-muted-foreground' }, props.description) : null
      ]),
      h('div', { class: 'min-w-0' }, slots.default?.())
    ])
  }
})

const FormField = defineComponent({
  props: {
    label: { type: String, required: true },
    required: { type: Boolean, default: false },
    error: { type: String, default: '' }
  },
  setup(props, { slots, attrs }) {
    return () => h('label', { ...attrs, class: ['block space-y-1.5', attrs.class] }, [
      h('span', { class: 'text-xs font-medium' }, [
        props.label,
        props.required ? h('span', { class: 'ml-0.5 text-destructive', 'aria-hidden': 'true' }, '*') : null
      ]),
      slots.default?.(),
      props.error ? h('span', { class: 'block text-xs font-medium text-destructive' }, props.error) : null
    ])
  }
})

const authStore = useAuthStore()
const loading = ref(false)
const products = ref([])
const selectedProducts = ref([])
const productTypes = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const stats = ref({})
const formErrors = reactive({})

const filters = reactive({ search: '', status: 'all', locale: 'all', featured: 'all' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const productForm = reactive({
  id: null,
  product_type_id: null,
  name: '',
  slug: '',
  description: '',
  short_description: '',
  weight_grams: 0,
  status: 'active',
  locale: 'zh',
  featured: false,
  meta_title: '',
  meta_description: '',
  specs: {},
  variants: []
})
const confirmation = reactive({
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
const localeFilterOptions = [
  { label: '全部语言', value: 'all' },
  { label: '中文', value: 'zh' },
  { label: 'English', value: 'en' }
]
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
const selectedProductType = computed(() => productTypes.value.find((type) => type.id === productForm.product_type_id) || null)
const selectedSpecDefinitions = computed(() => (selectedProductType.value?.spec_definitions || []).filter((spec) => !spec.is_variant_option))
const variantSpecDefinitions = computed(() => (selectedProductType.value?.spec_definitions || []).filter((spec) => spec.is_variant_option))
const defaultVariantIndex = computed(() => {
  const index = productForm.variants.findIndex((variant) => variant.is_default)
  return index >= 0 ? index : 0
})
const productTypeSelectValue = computed(() => productForm.product_type_id == null ? '__none__' : String(productForm.product_type_id))
const selectionState = computed(() => {
  if (products.value.length === 0 || selectedProducts.value.length === 0) return false
  return selectedProducts.value.length === products.value.length ? true : 'indeterminate'
})

const hasPermission = (permission) => authStore.hasPermission(permission)
const getStatusName = (status) => ({ active: '在售', inactive: '下架', out_of_stock: '缺货' })[status] || status
const statusTone = (status) => ({ active: 'green', inactive: 'gray', out_of_stock: 'coral' })[status] || 'gray'
const localeName = (locale) => ({ zh: '中文', en: 'English' })[locale] || locale || '-'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatMoney = (amount) => Number(amount || 0).toFixed(2)

const parseSpecOptions = (spec) => {
  if (!spec?.options) return []
  try {
    const parsed = JSON.parse(spec.options)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}
const formatSpecOption = (option) => String(option).replace(/_/g, ' ')
const getSpecLabel = (spec) => spec.unit ? `${spec.name} (${spec.unit})` : spec.name
const specSelectValue = (value) => value === undefined || value === null || value === '' ? '__empty__' : String(value)
const setSpecSelectValue = (slug, value) => {
  productForm.specs[slug] = value === '__empty__' ? '' : value
  clearFieldError(`spec:${slug}`)
}

const coerceSpecValueForForm = (definition, value) => {
  if (!definition) return value
  if (definition.field_type === 'number') {
    const numberValue = Number(value)
    return Number.isFinite(numberValue) ? numberValue : undefined
  }
  if (definition.field_type === 'boolean') return value === true || value === 'true' || value === '1'
  return value
}
const buildSpecFormValues = (product) => {
  const values = {}
  ;(product.spec_values || []).forEach((item) => {
    if (item.definition?.slug) values[item.definition.slug] = coerceSpecValueForForm(item.definition, item.value)
  })
  return values
}
const parseVariantOptions = (variant) => {
  if (!variant?.option_values) return {}
  if (typeof variant.option_values === 'object') return { ...variant.option_values }
  try {
    const parsed = JSON.parse(variant.option_values)
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}
const createEmptyVariant = (overrides = {}) => ({
  id: null,
  sku: '',
  title: '',
  option_values: {},
  price: 0,
  sale_price: null,
  stock: 0,
  weight_grams: 0,
  is_default: false,
  is_active: true,
  sort_order: productForm.variants.length * 10,
  ...overrides
})
const buildVariantFormValues = (product) => {
  const variants = (product.variants || []).map((variant, index) => createEmptyVariant({
    id: variant.id || null,
    sku: variant.sku || '',
    title: variant.title || '',
    option_values: parseVariantOptions(variant),
    price: Number(variant.price || 0),
    sale_price: variant.sale_price ?? null,
    stock: Number(variant.stock || 0),
    weight_grams: variant.weight_grams ?? variant.weight ?? 0,
    is_default: Boolean(variant.is_default),
    is_active: variant.is_active !== false,
    sort_order: variant.sort_order ?? index * 10
  }))
  if (variants.length === 0) variants.push(createEmptyVariant({ weight_grams: product.weight_grams ?? product.weight ?? 0, is_default: true }))
  if (!variants.some((variant) => variant.is_default)) variants[0].is_default = true
  return variants
}

const getPrimaryVariantWeight = (variants) => {
  const primaryVariant = variants.find((variant) => variant.is_default) || variants[0]
  const primaryWeight = Number(primaryVariant?.weight_grams ?? 0)
  if (Number.isFinite(primaryWeight) && primaryWeight > 0) {
    return primaryWeight
  }

  const fallbackWeight = Number(productForm.weight_grams || 0)
  return Number.isFinite(fallbackWeight) && fallbackWeight > 0 ? fallbackWeight : 0
}

const addVariant = () => {
  productForm.variants.push(createEmptyVariant({ is_default: productForm.variants.length === 0 }))
  clearFieldError('variants')
}
const removeVariant = (index) => {
  if (productForm.variants.length <= 1) {
    toast.warning('至少保留一个变体')
    return
  }
  const wasDefault = productForm.variants[index]?.is_default
  productForm.variants.splice(index, 1)
  if (wasDefault) setDefaultVariant(0)
}
const setDefaultVariant = (index) => {
  productForm.variants.forEach((variant, currentIndex) => { variant.is_default = currentIndex === index })
}
const normalizeFormVariants = () => {
  if (!productForm.variants.length) return []
  if (!productForm.variants.some((variant) => variant.is_default)) productForm.variants[0].is_default = true
  return productForm.variants.map((variant, index) => {
    const optionValues = {}
    variantSpecDefinitions.value.forEach((spec) => {
      const value = variant.option_values?.[spec.slug]
      if (value !== undefined && value !== null && value !== '') optionValues[spec.slug] = value
    })
    return {
      id: variant.id || undefined,
      sku: String(variant.sku || '').trim(),
      title: String(variant.title || '').trim(),
      option_values: optionValues,
      price: Number(variant.price || 0),
      sale_price: variant.sale_price === '' || variant.sale_price == null ? null : Number(variant.sale_price),
      stock: Number(variant.stock || 0),
      weight_grams: Number(variant.weight_grams || 0),
      is_default: Boolean(variant.is_default),
      is_active: variant.is_active !== false,
      sort_order: Number(variant.sort_order ?? index * 10)
    }
  })
}
const buildProductPayload = () => {
  const variants = normalizeFormVariants()
  return {
    id: productForm.id,
    product_type_id: productForm.product_type_id,
    name: productForm.name.trim(),
    slug: productForm.slug.trim(),
    description: productForm.description,
    short_description: productForm.short_description,
    weight_grams: getPrimaryVariantWeight(variants),
    status: productForm.status,
    locale: productForm.locale,
    featured: productForm.featured,
    meta_title: productForm.meta_title,
    meta_description: productForm.meta_description,
    specs: { ...productForm.specs },
    variants,
  }
}

const clearFormErrors = () => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (field) => { delete formErrors[field] }
const validateForm = (payload) => {
  clearFormErrors()
  if (!payload.name) formErrors.name = '请输入商品名称'
  if (!payload.slug) formErrors.slug = '请输入 URL slug'
  selectedSpecDefinitions.value.forEach((spec) => {
    const value = payload.specs[spec.slug]
    if (spec.is_required && (value === undefined || value === null || value === '')) {
      formErrors[`spec:${spec.slug}`] = `请填写${spec.name}`
    }
  })
  if (!payload.variants.length) formErrors.variants = '请至少添加一个 SKU 变体'
  else if (payload.variants.some((variant) => !variant.sku)) formErrors.variants = '每个变体都必须填写 SKU'
  else if (new Set(payload.variants.map((variant) => variant.sku.toLowerCase())).size !== payload.variants.length) formErrors.variants = '变体 SKU 不能重复'
  else if (payload.variants.some((variant) => Number(variant.price) <= 0)) formErrors.variants = '每个变体价格必须大于 0'
  else if (payload.variants.some((variant) => Number(variant.stock) < 0)) formErrors.variants = '变体库存不能为负数'
  if (Object.keys(formErrors).length > 0) {
    toast.error('请检查商品表单中的必填项')
    return false
  }
  return true
}

const handleProductTypeSelect = (value) => {
  productForm.product_type_id = value === '__none__' ? null : Number(value)
  const nextSpecs = {}
  selectedSpecDefinitions.value.forEach((spec) => {
    if (spec.field_type === 'boolean') nextSpecs[spec.slug] = false
  })
  productForm.specs = nextSpecs
  productForm.variants.forEach((variant) => { variant.option_values = {} })
  clearFormErrors()
}
const resetForm = () => {
  Object.assign(productForm, {
    id: null,
    product_type_id: null,
    name: '',
    slug: '',
    description: '',
    short_description: '',
    weight_grams: 0,
    status: 'active',
    locale: 'zh',
    featured: false,
    meta_title: '',
    meta_description: '',
    specs: {},
    variants: []
  })
  productForm.variants = [createEmptyVariant({ is_default: true })]
  clearFormErrors()
}

const buildFilterParams = () => ({
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.status !== 'all' ? { status: filters.status } : {}),
  ...(filters.locale !== 'all' ? { locale: filters.locale } : {}),
  ...(filters.featured !== 'all' ? { featured: filters.featured } : {})
})
const fetchProductTypes = async () => {
  try {
    const response = await axios.get('/api/admin/product-types')
    productTypes.value = response.data?.data || []
  } catch (error) {
    console.error('Failed to fetch product types:', error)
  }
}
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/products/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch product stats:', error)
  }
}
const fetchProducts = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/products', {
      params: { page: pagination.page, page_size: pagination.pageSize, ...buildFilterParams() }
    })
    products.value = response.data.products || []
    pagination.total = response.data.pagination?.total || 0
    selectedProducts.value = []
  } catch (error) {
    console.error('Failed to fetch products:', error)
  } finally {
    loading.value = false
  }
}
const refreshProducts = () => Promise.all([fetchProducts(), fetchStats()])
const applyFilters = () => { pagination.page = 1; fetchProducts() }
const resetFilters = () => {
  Object.assign(filters, { search: '', status: 'all', locale: 'all', featured: 'all' })
  pagination.page = 1
  fetchProducts()
}
const updatePage = (page) => { pagination.page = page; fetchProducts() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchProducts() }

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}
const showEditDialog = async (product) => {
  dialogMode.value = 'edit'
  let detail = product
  try {
    if (productTypes.value.length === 0) await fetchProductTypes()
    const response = await axios.get(`/api/admin/products/${product.id}`)
    detail = response.data?.product || product
    if (detail.product_type && !productTypes.value.some((type) => type.id === detail.product_type.id)) {
      productTypes.value.push(detail.product_type)
    }
  } catch (error) {
    toast.warning('获取商品详情失败，已使用列表数据编辑')
  }
  Object.assign(productForm, {
    id: detail.id,
    product_type_id: detail.product_type_id || detail.product_type?.id || null,
    name: detail.name || '',
    slug: detail.slug || '',
    description: detail.description || '',
    short_description: detail.short_description || detail.short_desc || '',
    weight_grams: detail.weight_grams ?? detail.weight ?? 0,
    status: detail.status || 'active',
    locale: detail.locale || 'zh',
    featured: Boolean(detail.featured),
    meta_title: detail.meta_title || '',
    meta_description: detail.meta_description || detail.meta_desc || '',
    specs: buildSpecFormValues(detail),
    variants: buildVariantFormValues(detail)
  })
  clearFormErrors()
  dialogVisible.value = true
}
const submitForm = async () => {
  const payload = buildProductPayload()
  if (!validateForm(payload)) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/products', payload)
      toast.success('商品创建成功')
    } else {
      const { id, ...data } = payload
      await axios.put(`/api/admin/products/${id}`, data)
      toast.success('商品更新成功')
    }
    dialogVisible.value = false
    await refreshProducts()
  } catch (error) {
    console.error('Failed to save product:', error)
  } finally {
    submitting.value = false
  }
}

const isSelected = (productId) => selectedProducts.value.some((product) => product.id === productId)
const toggleAllProducts = (checked) => { selectedProducts.value = checked === true ? [...products.value] : [] }
const toggleProduct = (product, checked) => {
  if (checked === true && !isSelected(product.id)) selectedProducts.value = [...selectedProducts.value, product]
  else if (checked !== true) selectedProducts.value = selectedProducts.value.filter((selected) => selected.id !== product.id)
}
const setConfirmation = (values) => Object.assign(confirmation, {
  open: true,
  type: '',
  target: null,
  status: '',
  confirmLabel: '确定',
  destructive: false,
  ...values
})
const requestToggleStatus = (product) => {
  const status = product.status === 'active' ? 'inactive' : 'active'
  const action = status === 'active' ? '上架' : '下架'
  setConfirmation({
    type: 'status', target: product, status, title: `${action}商品？`,
    description: `商品“${product.name}”将被${action}。`, confirmLabel: action
  })
}
const requestDelete = (product) => setConfirmation({
  type: 'delete', target: product, title: '删除商品？',
  description: `商品“${product.name}”将被永久删除，此操作不可恢复。`, confirmLabel: '删除', destructive: true
})
const requestBatchStatus = (status) => {
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
      await axios.patch(`/api/admin/products/${target.id}/status`, { status })
      toast.success(status === 'active' ? '商品已上架' : '商品已下架')
    } else if (type === 'delete') {
      await axios.delete(`/api/admin/products/${target.id}`)
      toast.success('商品已删除')
    } else if (type === 'batch-status') {
      await axios.post('/api/admin/products/batch-status', { product_ids: target.map((product) => product.id), status })
      toast.success(status === 'active' ? '商品已批量上架' : '商品已批量下架')
    } else if (type === 'batch-delete') {
      await axios.post('/api/admin/products/batch-delete', { product_ids: target.map((product) => product.id) })
      toast.success('商品已批量删除')
    }
    await refreshProducts()
  } catch (error) {
    console.error('Failed to update products:', error)
  }
}

onMounted(() => Promise.all([fetchProductTypes(), fetchStats(), fetchProducts()]))
</script>

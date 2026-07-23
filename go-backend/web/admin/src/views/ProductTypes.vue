<template>
  <div class="space-y-4">
    <AdminPageHeader title="产品模板" description="按车圈、车架等模板维护字段结构，具体商品参数在商品/SKU 中填写">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/products">
            <Package class="size-4" />
            商品管理
          </RouterLink>
        </Button>
        <Button v-if="hasPermission('product:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加模板
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <div class="grid gap-3 md:grid-cols-[minmax(240px,1fr)_180px_auto]">
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="名称或标识" />
          </div>
        </label>
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 状态</span>
          <Select v-model="filters.status">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="enabled">已启用</SelectItem>
              <SelectItem value="disabled">已停用</SelectItem>
            </SelectContent>
          </Select>
        </label>
        <div class="flex items-end">
          <Button variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </div>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[920px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>标识</TableHead>
            <TableHead class="w-28 text-right">字段模板</TableHead>
            <TableHead class="w-28 text-right">变体字段</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-20 text-right">排序</TableHead>
            <TableHead class="w-44">更新时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="filteredTypes.length === 0" :colspan="9">
            <div class="flex flex-col items-center text-muted-foreground">
              <Tags class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无产品模板</span>
            </div>
          </TableEmpty>

          <TableRow v-for="type in filteredTypes" :key="type.id">
            <TableCell class="font-mono text-[10px] font-bold text-muted-foreground">{{ type.id }}</TableCell>
            <TableCell>
              <div class="max-w-80">
                <p class="truncate font-bold text-xs">{{ type.name }}</p>
                <p v-if="type.description" class="mt-0.5 truncate text-[10px] text-muted-foreground/70">{{ type.description }}</p>
              </div>
            </TableCell>
            <TableCell class="font-mono text-[11px] font-bold text-muted-foreground/80">{{ type.slug }}</TableCell>
            <TableCell class="text-right font-mono text-xs font-bold tabular-nums">{{ type.spec_definitions?.length || 0 }}</TableCell>
            <TableCell class="text-right font-mono text-xs font-bold tabular-nums">{{ variantSpecCount(type) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="type.is_enabled ? 'green' : 'gray'">
                {{ type.is_enabled ? '已启用' : '已停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right font-mono text-xs font-bold tabular-nums">{{ type.sort_order || 0 }}</TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground/80">{{ formatDate(type.updated_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理产品模板 ${type.name}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem v-if="hasPermission('product:edit')" @select="showEditDialog(type)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('product:edit')" @select="toggleType(type)">
                    <CircleCheck v-if="!type.is_enabled" class="size-4" />
                    <CircleOff v-else class="size-4" />
                    {{ type.is_enabled ? '停用' : '启用' }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('product:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('product:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(type)"
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
    </AdminTablePanel>

    <Dialog v-model:open="dialogVisible">
      <DialogContent
        size="full"
        class="flex h-[92dvh] max-h-[calc(100dvh-1rem)] gap-0 overflow-hidden p-0"
        @open-auto-focus.prevent
      >
        <form class="flex min-h-0 min-w-0 flex-1 flex-col" @submit.prevent="submitForm">
          <DialogHeader class="shrink-0 border-b px-5 py-3 pr-12">
            <DialogTitle>{{ dialogMode === 'create' ? '添加产品模板' : '编辑产品模板' }}</DialogTitle>
            <DialogDescription>
              产品模板只定义字段结构；具体重量、价格、库存和每个 SKU 的实际取值在商品编辑里维护。
            </DialogDescription>
          </DialogHeader>

          <div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
            <section class="rounded-2xl border border-dashed border-border/80 bg-card/70 p-4">
              <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 class="text-sm font-black tracking-tighter italic uppercase">基础信息</h3>
                  <p class="mt-1 text-xs text-muted-foreground">两行内完成模板名称、标识、排序和启用状态。</p>
                </div>
                <label class="inline-flex items-center gap-2 rounded-full border border-dashed px-3 py-1.5 text-xs font-medium">
                  <Switch v-model="typeForm.is_enabled" aria-label="启用产品模板" />
                  启用模板
                </label>
              </div>
              <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-[minmax(220px,1fr)_minmax(220px,1fr)_120px]">
                <FormField label="模板名称" required :error="formErrors.name">
                  <Input v-model="typeForm.name" @input="clearFieldError('name')" />
                </FormField>
                <FormField label="模板标识" required :error="formErrors.slug">
                  <Input v-model="typeForm.slug" class="font-mono" placeholder="例如：carbon_rim / carbon_frame" @input="clearFieldError('slug')" />
                </FormField>
                <FormField label="排序">
                  <Input v-model.number="typeForm.sort_order" type="number" min="0" step="1" />
                </FormField>
                <FormField label="说明" class="md:col-span-2 xl:col-span-3">
                  <Textarea v-model="typeForm.description" class="min-h-14 resize-y" placeholder="可选，给后台识别用，不会作为具体商品参数" />
                </FormField>
              </div>
            </section>

            <section class="rounded-2xl border border-dashed border-border/80 bg-card/70 p-4">
              <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
                <div class="space-y-1">
                  <h3 class="text-sm font-black tracking-tighter italic uppercase">字段模板</h3>
                  <p class="text-xs leading-5 text-muted-foreground">
                    这里只定义“该类商品编辑时出现哪些字段”。不要在这里填写某个具体产品的重量、尺寸、库存或价格。
                  </p>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <span class="rounded-full bg-muted px-2.5 py-1 text-xs text-muted-foreground">
                    {{ typeForm.spec_definitions.length }} 个字段
                  </span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    :aria-pressed="showSpecAdvanced"
                    @click="showSpecAdvanced = !showSpecAdvanced"
                  >
                    <SlidersHorizontal class="size-3.5" />
                    {{ showSpecAdvanced ? '隐藏属性' : '字段属性' }}
                  </Button>
                  <Button type="button" variant="outline" size="sm" @click="addSpecDefinition">
                    <Plus class="size-3.5" />
                    添加字段
                  </Button>
                </div>
              </div>

              <div class="grid min-w-0 gap-3 lg:grid-cols-2 xl:grid-cols-3">
                <div v-if="typeForm.spec_definitions.length === 0" class="rounded-xl border border-dashed py-8 text-center text-xs text-muted-foreground lg:col-span-2 xl:col-span-3">
                  暂无字段模板。添加后，这些字段会出现在商品编辑页。
                </div>

                <section
                  v-for="(spec, index) in typeForm.spec_definitions"
                  :key="spec.clientKey"
                  class="rounded-xl border bg-background/80 p-3"
                >
                  <div class="mb-2 flex items-center justify-between gap-2">
                    <strong class="text-xs font-black uppercase tracking-wider text-muted-foreground">字段 {{ index + 1 }}</strong>
                    <Button type="button" variant="ghost" size="icon" class="size-8" :aria-label="`删除字段 ${index + 1}`" @click="removeSpecDefinition(index)">
                      <Trash2 class="size-4 text-destructive" />
                    </Button>
                  </div>

                  <div
                    v-if="isProductSpecificSelect(spec)"
                    class="mb-2 rounded-lg border border-amber-500/25 bg-amber-500/10 px-3 py-2 text-xs leading-5 text-amber-800 dark:text-amber-200"
                  >
                    这个字段看起来像每个产品/SKU 自己决定的值。若不同产品的可选值不同，请把字段类型改成“文本/数字”，不要在产品模板里固定列出选项。
                  </div>

                  <div class="grid gap-2 sm:grid-cols-2">
                    <FormField label="字段名称" required :error="formErrors[`spec:${index}:name`]">
                      <Input v-model="spec.name" placeholder="字段显示名" @input="clearFieldError(`spec:${index}:name`)" />
                    </FormField>
                    <FormField label="字段标识" required :error="formErrors[`spec:${index}:slug`]">
                      <Input v-model="spec.slug" class="font-mono" placeholder="field_slug" @input="clearFieldError(`spec:${index}:slug`)" />
                    </FormField>
                    <FormField label="字段类型" required>
                      <Select v-model="spec.field_type">
                        <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="text">文本</SelectItem>
                          <SelectItem value="number">数字</SelectItem>
                          <SelectItem value="select">选项</SelectItem>
                          <SelectItem value="boolean">开关</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormField>
                    <FormField label="单位">
                      <Input v-model="spec.unit" placeholder="可选" />
                    </FormField>
                    <FormField label="排序">
                      <Input v-model.number="spec.sort_order" type="number" min="0" step="1" />
                    </FormField>
                    <FormField
                      v-if="spec.field_type === 'select'"
                      label="固定选项"
                      required
                      class="sm:col-span-2"
                      :error="formErrors[`spec:${index}:options`]"
                      description="只填写所有该类型产品共用的选项；每行一个。"
                    >
                      <Textarea
                        v-model="spec.optionsText"
                        class="min-h-12 font-mono text-xs"
                        placeholder="每行一个全类型共用选项，例如：Black&#10;White"
                        @input="clearFieldError(`spec:${index}:options`)"
                      />
                    </FormField>
                  </div>

                  <div v-if="showSpecAdvanced" class="mt-2 grid gap-2 border-t border-dashed pt-2 sm:grid-cols-2">
                    <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                      <span>必填</span>
                      <Switch v-model="spec.is_required" :aria-label="`${spec.name || '字段'}必填`" />
                    </label>
                    <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                      <span>可筛选</span>
                      <Switch v-model="spec.is_filterable" :aria-label="`${spec.name || '字段'}可筛选`" />
                    </label>
                    <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                      <span>前台可见</span>
                      <Switch v-model="spec.is_visible" :aria-label="`${spec.name || '字段'}前台可见`" />
                    </label>
                    <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                      <span>SKU 选项</span>
                      <Switch v-model="spec.is_variant_option" :aria-label="`${spec.name || '字段'}作为变体选项`" />
                    </label>
                  </div>
                </section>
              </div>
            </section>
          </div>

          <DialogFooter class="mx-0 mb-0 shrink-0 rounded-none border-t bg-background/95 px-5 py-3 backdrop-blur sm:flex-row sm:flex-nowrap sm:justify-end">
            <Button type="button" variant="outline" @click="dialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '保存中' : '保存模板' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      title="删除产品模板？"
      :description="`产品模板“${confirmation.target?.name || ''}”及其字段模板将被永久删除。`"
      confirm-label="删除"
      destructive
      @confirm="deleteType"
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
  ListChecks,
  LoaderCircle,
  MoreHorizontal,
  Package,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  SlidersHorizontal,
  Tags,
  Trash2
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const FormField = defineComponent({
  props: {
    label: { type: String, required: true },
    required: { type: Boolean, default: false },
    error: { type: String, default: '' },
    description: { type: String, default: '' }
  },
  setup(props, { slots, attrs }) {
    return () => h('label', { ...attrs, class: ['block space-y-1', attrs.class] }, [
      h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, [
        props.label,
        props.required ? h('span', { class: 'ml-0.5 text-destructive', 'aria-hidden': 'true' }, '*') : null
      ]),
      slots.default?.(),
      props.error
        ? h('span', { class: 'block text-xs font-medium text-destructive' }, props.error)
        : props.description
          ? h('span', { class: 'block text-xs leading-5 text-muted-foreground' }, props.description)
          : null
    ])
  }
})

const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogMode = ref('create')
const showSpecAdvanced = ref(false)
const productTypes = ref([])
const filters = reactive({ search: '', status: 'all' })
const formErrors = reactive({})
const confirmation = reactive({ open: false, target: null })
let nextSpecKey = 1

const typeForm = reactive({
  id: null,
  name: '',
  slug: '',
  description: '',
  sort_order: 0,
  is_enabled: true,
  spec_definitions: []
})

const filteredTypes = computed(() => {
  const keyword = filters.search.trim().toLowerCase()
  return productTypes.value.filter((type) => {
    if (filters.status === 'enabled' && !type.is_enabled) return false
    if (filters.status === 'disabled' && type.is_enabled) return false
    if (!keyword) return true
    return String(type.name || '').toLowerCase().includes(keyword) || String(type.slug || '').toLowerCase().includes(keyword)
  })
})

const statItems = computed(() => [
  { key: 'total', label: '模板总数', value: productTypes.value.length, icon: Tags, tone: 'gray' },
  { key: 'enabled', label: '已启用', value: productTypes.value.filter((type) => type.is_enabled).length, icon: CircleCheck, tone: 'green' },
  { key: 'specs', label: '字段模板', value: productTypes.value.reduce((total, type) => total + (type.spec_definitions?.length || 0), 0), icon: ListChecks, tone: 'blue' },
  { key: 'variants', label: '变体字段', value: productTypes.value.reduce((total, type) => total + variantSpecCount(type), 0), icon: Boxes, tone: 'amber' }
])

const hasPermission = (permission) => authStore.hasPermission(permission)
const formatDate = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'
const variantSpecCount = (type) => (type.spec_definitions || []).filter((spec) => spec.is_variant_option).length
const productSpecificSpecPattern = /(weight|重量|size|尺寸|diameter|直径|width|宽|height|高|depth|深|length|长|pack|包装|数量|count|qty)/i
const isProductSpecificSelect = (spec) => (
  spec.field_type === 'select' &&
  Boolean(String(spec.optionsText || '').trim()) &&
  productSpecificSpecPattern.test(`${spec.name || ''} ${spec.slug || ''}`)
)

const createEmptySpec = (overrides = {}) => ({
  id: 0,
  clientKey: nextSpecKey++,
  group: '',
  name: '',
  slug: '',
  field_type: 'text',
  unit: '',
  is_required: false,
  is_filterable: false,
  is_visible: true,
  is_variant_option: false,
  sort_order: 0,
  optionsText: '',
  validation: '',
  ...overrides
})

const optionsToText = (options) => {
  if (!options) return ''
  try {
    const parsed = JSON.parse(options)
    return Array.isArray(parsed) ? parsed.join('\n') : ''
  } catch {
    return ''
  }
}

const apiSpecToForm = (spec) => ({
  ...createEmptySpec(),
  ...spec,
  clientKey: nextSpecKey++,
  optionsText: optionsToText(spec.options)
})

const resetForm = () => {
  Object.assign(typeForm, {
    id: null,
    name: '',
    slug: '',
    description: '',
    sort_order: 0,
    is_enabled: true,
    spec_definitions: []
  })
  clearFormErrors()
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  showSpecAdvanced.value = false
  dialogVisible.value = true
}

const showEditDialog = (type) => {
  dialogMode.value = 'edit'
  showSpecAdvanced.value = false
  Object.assign(typeForm, {
    id: type.id,
    name: type.name || '',
    slug: type.slug || '',
    description: type.description || '',
    sort_order: Number(type.sort_order || 0),
    is_enabled: type.is_enabled !== false,
    spec_definitions: (type.spec_definitions || []).map(apiSpecToForm)
  })
  clearFormErrors()
  dialogVisible.value = true
}

const addSpecDefinition = () => {
  const spec = createEmptySpec()
  spec.sort_order = typeForm.spec_definitions.length * 10
  typeForm.spec_definitions.push(spec)
}

const removeSpecDefinition = (index) => {
  typeForm.spec_definitions.splice(index, 1)
  clearFormErrors()
}

const clearFormErrors = () => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (key) => { delete formErrors[key] }

const specOptions = (spec) => spec.optionsText
  .split(/\r?\n/)
  .map((value) => value.trim())
  .filter(Boolean)
  .filter((value, index, values) => values.indexOf(value) === index)

const validateForm = () => {
  clearFormErrors()
  const slugPattern = /^[a-z0-9]+(?:[_-][a-z0-9]+)*$/
  if (!typeForm.name.trim()) formErrors.name = '请输入模板名称'
  if (!slugPattern.test(typeForm.slug.trim())) formErrors.slug = '请输入有效的模板标识'

  const seenSlugs = new Set()
  typeForm.spec_definitions.forEach((spec, index) => {
    if (!spec.name.trim()) formErrors[`spec:${index}:name`] = '请输入字段名称'
    const slug = spec.slug.trim()
    if (!slugPattern.test(slug)) formErrors[`spec:${index}:slug`] = '请输入有效的字段标识'
    else if (seenSlugs.has(slug)) formErrors[`spec:${index}:slug`] = '字段标识不能重复'
    else seenSlugs.add(slug)
    if (spec.field_type === 'select' && specOptions(spec).length === 0) {
      formErrors[`spec:${index}:options`] = '请至少填写一个选项'
    }
  })

  if (Object.keys(formErrors).length > 0) {
    toast.error('请检查产品模板表单')
    return false
  }
  return true
}

const buildPayload = (source = typeForm, enabled = source.is_enabled) => ({
  name: String(source.name || '').trim(),
  slug: String(source.slug || '').trim().toLowerCase(),
  description: String(source.description || '').trim(),
  sort_order: Number(source.sort_order || 0),
  is_enabled: Boolean(enabled),
  spec_definitions: (source.spec_definitions || []).map((spec) => ({
    id: Number(spec.id || 0),
    group: String(spec.group || '').trim(),
    name: String(spec.name || '').trim(),
    slug: String(spec.slug || '').trim().toLowerCase(),
    field_type: spec.field_type || 'text',
    unit: String(spec.unit || '').trim(),
    is_required: Boolean(spec.is_required),
    is_filterable: Boolean(spec.is_filterable),
    is_visible: Boolean(spec.is_visible),
    is_variant_option: Boolean(spec.is_variant_option),
    sort_order: Number(spec.sort_order || 0),
    options: spec.field_type === 'select'
      ? JSON.stringify(spec.optionsText === undefined ? optionsToText(spec.options).split(/\r?\n/).filter(Boolean) : specOptions(spec))
      : '',
    validation: String(spec.validation || '')
  }))
})

const fetchProductTypes = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/product-types', { params: { include_disabled: true } })
    productTypes.value = response.data?.data || []
  } catch (error) {
    console.error('Failed to fetch product types:', error)
  } finally {
    loading.value = false
  }
}

const submitForm = async () => {
  if (!validateForm()) return
  submitting.value = true
  try {
    const payload = buildPayload()
    if (dialogMode.value === 'create') await axios.post('/api/admin/product-types', payload)
    else await axios.put(`/api/admin/product-types/${typeForm.id}`, payload)
    toast.success(dialogMode.value === 'create' ? '产品模板已创建' : '产品模板已更新')
    dialogVisible.value = false
    await fetchProductTypes()
  } catch (error) {
    console.error('Failed to save product type:', error)
  } finally {
    submitting.value = false
  }
}

const toggleType = async (type) => {
  try {
    await axios.put(`/api/admin/product-types/${type.id}`, buildPayload(type, !type.is_enabled))
    toast.success(type.is_enabled ? '产品模板已停用' : '产品模板已启用')
    await fetchProductTypes()
  } catch (error) {
    console.error('Failed to toggle product type:', error)
  }
}

const requestDelete = (type) => Object.assign(confirmation, { open: true, target: type })
const deleteType = async () => {
  const type = confirmation.target
  confirmation.open = false
  if (!type) return
  try {
    await axios.delete(`/api/admin/product-types/${type.id}`)
    toast.success('产品模板已删除')
    await fetchProductTypes()
  } catch (error) {
    console.error('Failed to delete product type:', error)
  } finally {
    confirmation.target = null
  }
}

const resetFilters = () => Object.assign(filters, { search: '', status: 'all' })

onMounted(fetchProductTypes)
</script>

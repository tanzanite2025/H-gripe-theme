<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品规格模板" description="按车圈、车架等模板维护字段结构，具体商品参数在商品/SKU 中填写">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">
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

    <ProductSpecificationTemplateFilterPanel
      :filters="filters"
      @reset="resetFilters"
    />

    <ProductSpecificationTemplateTablePanel
      :loading="loading"
      :templates="filteredTemplates"
      :can-edit="hasPermission('product:edit')"
      :can-delete="hasPermission('product:delete')"
      :variant-spec-count="variantSpecCount"
      :format-date="formatDate"
      @edit="showEditTemplateDialog"
      @toggle="toggleTemplate"
      @delete="requestDelete"
    />

    <ProductSpecificationTemplateEditorDialog
      v-model:open="dialogVisible"
      v-model:show-spec-advanced="showSpecAdvanced"
      :mode="dialogMode"
      :form="templateForm"
      :errors="formErrors"
      :submitting="submitting"
      :system-managed="templateForm.is_system_managed"
      :is-product-specific-select="isProductSpecificSelect"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @add-spec="addSpecDefinition"
      @remove-spec="removeSpecDefinition"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      title="删除商品规格模板？"
      :description="`商品规格模板“${confirmation.target?.name || ''}”及其字段模板将被永久删除。`"
      confirm-label="删除"
      destructive
      @confirm="deleteTemplate"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Boxes,
  CircleCheck,
  ListChecks,
  Package,
  Plus,
  Tags
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import ProductSpecificationTemplateEditorDialog from '@/components/admin/product/ProductSpecificationTemplateEditorDialog.vue'
import ProductSpecificationTemplateFilterPanel from '@/components/admin/product/ProductSpecificationTemplateFilterPanel.vue'
import ProductSpecificationTemplateTablePanel from '@/components/admin/product/ProductSpecificationTemplateTablePanel.vue'
import type {
  ProductSpecFieldType,
  ProductSpecPresentation,
  ProductSpecTemplateDialogMode,
  ProductSpecTemplateFilters,
  ProductSpecTemplateForm,
  ProductSpecTemplateFormErrors,
  ProductSpecTemplatePayload,
  ProductSpecTemplateRecord,
  ProductSpecTemplateSpecDefinition,
  ProductSpecTemplateSpecForm
} from '@/modules/product/productSpecificationTemplateTypes'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import productSpecTemplateApi from '@/api/productSpecificationTemplates'

const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogMode = ref<ProductSpecTemplateDialogMode>('create')
const showSpecAdvanced = ref(false)
const productSpecTemplates = ref<ProductSpecTemplateRecord[]>([])
const filters = reactive<ProductSpecTemplateFilters>({ search: '', status: 'all' })
const formErrors = reactive<ProductSpecTemplateFormErrors>({})
const confirmation = reactive<{ open: boolean; target: ProductSpecTemplateRecord | null }>({ open: false, target: null })
let nextSpecKey = 1

const templateForm = reactive<ProductSpecTemplateForm>({
  id: null,
  is_system_managed: false,
  name: '',
  slug: '',
  description: '',
  sort_order: 0,
  is_enabled: true,
  spec_definitions: []
})

const filteredTemplates = computed<ProductSpecTemplateRecord[]>(() => {
  const keyword = filters.search.trim().toLowerCase()
  return productSpecTemplates.value.filter((template) => {
    if (filters.status === 'enabled' && !template.is_enabled) return false
    if (filters.status === 'disabled' && template.is_enabled) return false
    if (!keyword) return true
    return String(template.name || '').toLowerCase().includes(keyword)
      || String(template.slug || '').toLowerCase().includes(keyword)
  })
})

const statItems = computed(() => [
  { key: 'total', label: '模板总数', value: productSpecTemplates.value.length, icon: Tags, tone: 'gray' },
  { key: 'enabled', label: '已启用', value: productSpecTemplates.value.filter((template) => template.is_enabled).length, icon: CircleCheck, tone: 'green' },
  { key: 'specs', label: '字段模板', value: productSpecTemplates.value.reduce((total, template) => total + (template.spec_definitions?.length || 0), 0), icon: ListChecks, tone: 'blue' },
  { key: 'variants', label: '变体字段', value: productSpecTemplates.value.reduce((total, template) => total + variantSpecCount(template), 0), icon: Boxes, tone: 'amber' }
])

const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
const formatDate = (value?: string | null): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
const variantSpecCount = (template: ProductSpecTemplateRecord): number => (template.spec_definitions || []).filter((spec) => spec.is_variant_option).length
const productSpecificSpecPattern = /(weight|重量|size|尺寸|diameter|直径|width|宽|height|高|depth|深|length|长|pack|包装|数量|count|qty)/i
const isProductSpecificSelect = (spec: ProductSpecTemplateSpecForm): boolean => (
  spec.field_type === 'select' &&
  Boolean(String(spec.optionsText || '').trim()) &&
  productSpecificSpecPattern.test(`${spec.name || ''} ${spec.slug || ''}`)
)
const fieldTypes: ProductSpecFieldType[] = ['text', 'number', 'select', 'boolean']
const normalizeFieldType = (fieldType?: string | null): ProductSpecFieldType => (
  fieldTypes.includes(fieldType as ProductSpecFieldType) ? fieldType as ProductSpecFieldType : 'text'
)
const presentations: ProductSpecPresentation[] = ['text', 'color', 'image']
const normalizePresentation = (presentation?: string | null): ProductSpecPresentation => (
  presentations.includes(presentation as ProductSpecPresentation) ? presentation as ProductSpecPresentation : 'text'
)

const createEmptySpec = (overrides: Partial<ProductSpecTemplateSpecForm> = {}): ProductSpecTemplateSpecForm => ({
  id: 0,
  clientKey: nextSpecKey++,
  group: '',
  name: '',
  slug: '',
  field_type: 'text',
  presentation: 'text',
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

const optionsToText = (options?: string | null): string => {
  if (!options) return ''
  try {
    const parsed = JSON.parse(options)
    return Array.isArray(parsed) ? parsed.join('\n') : ''
  } catch {
    return ''
  }
}

const apiSpecToForm = (spec: ProductSpecTemplateSpecDefinition): ProductSpecTemplateSpecForm => ({
  ...createEmptySpec(),
  ...spec,
  clientKey: nextSpecKey++,
  id: spec.id ?? 0,
  group: String(spec.group || ''),
  name: String(spec.name || ''),
  slug: String(spec.slug || ''),
  field_type: normalizeFieldType(spec.field_type),
  presentation: normalizePresentation(spec.presentation),
  unit: String(spec.unit || ''),
  sort_order: Number(spec.sort_order || 0),
  validation: String(spec.validation || ''),
  optionsText: optionsToText(spec.options)
})

const resetForm = (): void => {
  Object.assign(templateForm, {
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

const showCreateDialog = (): void => {
  dialogMode.value = 'create'
  resetForm()
  showSpecAdvanced.value = false
  dialogVisible.value = true
}

const showEditTemplateDialog = (template: ProductSpecTemplateRecord): void => {
  dialogMode.value = 'edit'
  showSpecAdvanced.value = false
  Object.assign(templateForm, {
    id: template.id,
    is_system_managed: Boolean(template.is_system_managed),
    name: template.name || '',
    slug: template.slug || '',
    description: template.description || '',
    sort_order: Number(template.sort_order || 0),
    is_enabled: template.is_enabled !== false,
    spec_definitions: (template.spec_definitions || []).map(apiSpecToForm)
  })
  clearFormErrors()
  dialogVisible.value = true
}

const addSpecDefinition = (): void => {
  const spec = createEmptySpec()
  spec.sort_order = templateForm.spec_definitions.length * 10
  templateForm.spec_definitions.push(spec)
}

const removeSpecDefinition = (index: number): void => {
  templateForm.spec_definitions.splice(index, 1)
  clearFormErrors()
}

const clearFormErrors = (): void => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (key: string): void => { delete formErrors[key] }

const specOptionsFromText = (text?: string | null): string[] => String(text || '')
  .split(/\r?\n/)
  .map((value) => value.trim())
  .filter(Boolean)
  .filter((value, index, values) => values.indexOf(value) === index)

const specOptions = (spec: ProductSpecTemplateSpecForm): string[] => specOptionsFromText(spec.optionsText)

const specPayloadOptions = (spec: ProductSpecTemplateSpecForm | ProductSpecTemplateSpecDefinition): string[] => {
  if ('optionsText' in spec) return specOptionsFromText(spec.optionsText)
  return specOptionsFromText(optionsToText(spec.options))
}

const validateForm = (): boolean => {
  clearFormErrors()
  const slugPattern = /^[a-z0-9]+(?:[_-][a-z0-9]+)*$/
  if (!templateForm.name.trim()) formErrors.name = '请输入模板名称'
  if (!slugPattern.test(templateForm.slug.trim())) formErrors.slug = '请输入有效的模板标识'

  const seenSlugs = new Set<string>()
  templateForm.spec_definitions.forEach((spec, index) => {
    if (!spec.name.trim()) formErrors[`spec:${index}:name`] = '请输入字段名称'
    const slug = spec.slug.trim()
    if (!slugPattern.test(slug)) formErrors[`spec:${index}:slug`] = '请输入有效的字段标识'
    else if (seenSlugs.has(slug)) formErrors[`spec:${index}:slug`] = '字段标识不能重复'
    else seenSlugs.add(slug)
  })

  if (Object.keys(formErrors).length > 0) {
      toast.error('请检查商品规格模板表单')
    return false
  }
  return true
}

const buildPayload = (
  source: ProductSpecTemplateForm | ProductSpecTemplateRecord = templateForm,
  enabled = source.is_enabled !== false
): ProductSpecTemplatePayload => ({
  name: String(source.name || '').trim(),
  slug: String(source.slug || '').trim().toLowerCase(),
  description: String(source.description || '').trim(),
  sort_order: Number(source.sort_order || 0),
  is_enabled: Boolean(enabled),
  spec_definitions: (source.spec_definitions || []).map((spec) => {
    const fieldType = normalizeFieldType(spec.field_type)
    return {
      id: Number(spec.id || 0),
      group: String(spec.group || '').trim(),
      name: String(spec.name || '').trim(),
      slug: String(spec.slug || '').trim().toLowerCase(),
      field_type: fieldType,
      presentation: fieldType === 'select' && spec.is_variant_option ? normalizePresentation(spec.presentation) : 'text',
      unit: String(spec.unit || '').trim(),
      is_required: Boolean(spec.is_required),
      is_filterable: Boolean(spec.is_filterable),
      is_visible: Boolean(spec.is_visible),
      is_variant_option: Boolean(spec.is_variant_option),
      sort_order: Number(spec.sort_order || 0),
      options: fieldType === 'select' ? JSON.stringify(specPayloadOptions(spec)) : '',
      validation: String(spec.validation || '')
    }
  })
})

const fetchProductSpecificationTemplates = async (): Promise<void> => {
  loading.value = true
  try {
    productSpecTemplates.value = await productSpecTemplateApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product specification templates:', error)
  } finally {
    loading.value = false
  }
}

const submitForm = async (): Promise<void> => {
  if (!validateForm()) return
  submitting.value = true
  const wasCreating = dialogMode.value === 'create'
  try {
    const payload = buildPayload()
    let savedTemplate: ProductSpecTemplateRecord | null = null
    if (dialogMode.value === 'create') {
      savedTemplate = await productSpecTemplateApi.create(payload)
    } else if (templateForm.id !== null) {
      savedTemplate = await productSpecTemplateApi.update(templateForm.id, payload)
    }

    if (!savedTemplate?.id) {
      throw new Error('Product specification template save returned no record')
    }

    templateForm.id = savedTemplate.id
    dialogMode.value = 'edit'

    toast.success(wasCreating ? '商品规格模板已创建' : '商品规格模板已更新')
    dialogVisible.value = false
    await fetchProductSpecificationTemplates()
  } catch (error) {
    console.error('Failed to save product specification template:', error)
    toast.error('商品规格模板保存失败，请检查名称、标识和字段设置')
  } finally {
    submitting.value = false
  }
}

const toggleTemplate = async (template: ProductSpecTemplateRecord): Promise<void> => {
  try {
    await productSpecTemplateApi.update(template.id, buildPayload(template, !template.is_enabled))
    toast.success(template.is_enabled ? '商品规格模板已停用' : '商品规格模板已启用')
    await fetchProductSpecificationTemplates()
  } catch (error) {
    console.error('Failed to toggle product specification template:', error)
  }
}

const requestDelete = (template: ProductSpecTemplateRecord): void => {
  Object.assign(confirmation, { open: true, target: template })
}
const deleteTemplate = async (): Promise<void> => {
  const template = confirmation.target
  confirmation.open = false
  if (!template) return
  try {
    await productSpecTemplateApi.deleteProductSpecTemplate(template.id)
    toast.success('商品规格模板已删除')
    await fetchProductSpecificationTemplates()
  } catch (error) {
    console.error('Failed to delete product specification template:', error)
  } finally {
    confirmation.target = null
  }
}

const resetFilters = (): void => { Object.assign(filters, { search: '', status: 'all' }) }

onMounted(() => {
  void fetchProductSpecificationTemplates()
})
</script>


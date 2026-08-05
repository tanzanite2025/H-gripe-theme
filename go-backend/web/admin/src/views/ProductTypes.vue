<template>
  <div class="space-y-4">
    <AdminPageHeader title="产品模板" description="按车圈、车架等模板维护字段结构，具体商品参数在商品/SKU 中填写">
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

    <ProductTypeFilterPanel
      :filters="filters"
      @reset="resetFilters"
    />

    <ProductTypeTablePanel
      :loading="loading"
      :types="filteredTypes"
      :can-edit="hasPermission('product:edit')"
      :can-delete="hasPermission('product:delete')"
      :variant-spec-count="variantSpecCount"
      :format-date="formatDate"
      @edit="showEditDialog"
      @toggle="toggleType"
      @delete="requestDelete"
    />

    <ProductTypeEditorDialog
      v-model:open="dialogVisible"
      v-model:show-spec-advanced="showSpecAdvanced"
      :mode="dialogMode"
      :form="typeForm"
      :errors="formErrors"
      :submitting="submitting"
      :is-product-specific-select="isProductSpecificSelect"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @add-spec="addSpecDefinition"
      @remove-spec="removeSpecDefinition"
    />

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
import ProductTypeEditorDialog from '@/components/admin/product/ProductTypeEditorDialog.vue'
import ProductTypeFilterPanel from '@/components/admin/product/ProductTypeFilterPanel.vue'
import ProductTypeTablePanel from '@/components/admin/product/ProductTypeTablePanel.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import productTypeApi from '@/api/productTypes'

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
    productTypes.value = await productTypeApi.list({ include_disabled: true })
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
    if (dialogMode.value === 'create') await productTypeApi.create(payload)
    else await productTypeApi.update(typeForm.id, payload)
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
    await productTypeApi.update(type.id, buildPayload(type, !type.is_enabled))
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
    await productTypeApi.deleteProductType(type.id)
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

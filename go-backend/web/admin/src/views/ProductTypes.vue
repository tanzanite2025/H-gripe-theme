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
      :language-options="languageOptions"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @add-spec="addSpecDefinition"
      @remove-spec="removeSpecDefinition"
      @image-selected="handleImageSelected"
      @image-cleared="handleImageCleared"
      @image-error="handleImageError"
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

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
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
import type {
  ProductSpecFieldType,
  ProductSpecPresentation,
  ProductTypeDialogMode,
  ProductTypeFilters,
  ProductTypeForm,
  ProductTypeFormErrors,
  ProductTypePayload,
  ProductTypeRecord,
  ProductTypeSpecDefinition,
  ProductTypeSpecForm,
  ProductTypeTranslation,
  ProductTypeTranslationForm
} from '@/components/admin/product/productTypeTypes'
import { Button } from '@/components/ui/button'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { normalizeLocaleCode } from '@/lib/languages'
import { useAuthStore } from '@/stores/auth'
import productTypeApi from '@/api/productTypes'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogMode = ref<ProductTypeDialogMode>('create')
const showSpecAdvanced = ref(false)
const productTypes = ref<ProductTypeRecord[]>([])
const filters = reactive<ProductTypeFilters>({ search: '', status: 'all' })
const formErrors = reactive<ProductTypeFormErrors>({})
const confirmation = reactive<{ open: boolean; target: ProductTypeRecord | null }>({ open: false, target: null })
let nextSpecKey = 1

const typeForm = reactive<ProductTypeForm>({
  id: null,
  name: '',
  slug: '',
  description: '',
  image_media_asset_id: null,
  image_url: '',
  pending_image_file: null,
  remove_image: false,
  sort_order: 0,
  is_enabled: true,
  translations: [],
  spec_definitions: []
})

const filteredTypes = computed<ProductTypeRecord[]>(() => {
  const keyword = filters.search.trim().toLowerCase()
  return productTypes.value.filter((type) => {
    if (filters.status === 'enabled' && !type.is_enabled) return false
    if (filters.status === 'disabled' && type.is_enabled) return false
    if (!keyword) return true
    return String(type.name || '').toLowerCase().includes(keyword)
      || String(type.slug || '').toLowerCase().includes(keyword)
      || (type.translations || []).some((translation) => String(translation.name || '').toLowerCase().includes(keyword))
  })
})

const statItems = computed(() => [
  { key: 'total', label: '模板总数', value: productTypes.value.length, icon: Tags, tone: 'gray' },
  { key: 'enabled', label: '已启用', value: productTypes.value.filter((type) => type.is_enabled).length, icon: CircleCheck, tone: 'green' },
  { key: 'specs', label: '字段模板', value: productTypes.value.reduce((total, type) => total + (type.spec_definitions?.length || 0), 0), icon: ListChecks, tone: 'blue' },
  { key: 'variants', label: '变体字段', value: productTypes.value.reduce((total, type) => total + variantSpecCount(type), 0), icon: Boxes, tone: 'amber' }
])

const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
const formatDate = (value?: string | null): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
const variantSpecCount = (type: ProductTypeRecord): number => (type.spec_definitions || []).filter((spec) => spec.is_variant_option).length
const productSpecificSpecPattern = /(weight|重量|size|尺寸|diameter|直径|width|宽|height|高|depth|深|length|长|pack|包装|数量|count|qty)/i
const isProductSpecificSelect = (spec: ProductTypeSpecForm): boolean => (
  spec.field_type === 'select' &&
  Boolean(String(spec.optionsText || '').trim()) &&
  productSpecificSpecPattern.test(`${spec.name || ''} ${spec.slug || ''}`)
)
const usesProductScopedOptions = (spec: ProductTypeSpecForm): boolean => (
  spec.field_type === 'select' &&
  spec.is_variant_option &&
  (spec.presentation === 'color' || spec.presentation === 'image')
)

const fieldTypes: ProductSpecFieldType[] = ['text', 'number', 'select', 'boolean']
const normalizeFieldType = (fieldType?: string | null): ProductSpecFieldType => (
  fieldTypes.includes(fieldType as ProductSpecFieldType) ? fieldType as ProductSpecFieldType : 'text'
)
const presentations: ProductSpecPresentation[] = ['text', 'color', 'image']
const normalizePresentation = (presentation?: string | null): ProductSpecPresentation => (
  presentations.includes(presentation as ProductSpecPresentation) ? presentation as ProductSpecPresentation : 'text'
)

const createEmptySpec = (overrides: Partial<ProductTypeSpecForm> = {}): ProductTypeSpecForm => ({
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

const apiSpecToForm = (spec: ProductTypeSpecDefinition): ProductTypeSpecForm => ({
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

const translationRowsFor = (source: ProductTypeTranslation[] = []): ProductTypeTranslationForm[] => {
  const existing = new Map<string, ProductTypeTranslation>()
  source.forEach((translation) => {
    const locale = normalizeLocaleCode(translation.locale)
    if (locale) existing.set(locale, translation)
  })

  const rows = languageOptions.value.map((option) => {
    const translation = existing.get(option.value)
    return {
      id: translation?.id ?? null,
      locale: option.value,
      name: String(translation?.name || ''),
      description: String(translation?.description || '')
    }
  })
  const displayedLocales = new Set(rows.map((translation) => translation.locale))

  for (const [locale, translation] of existing) {
    if (displayedLocales.has(locale)) continue
    rows.push({
      id: translation.id ?? null,
      locale,
      name: String(translation.name || ''),
      description: String(translation.description || '')
    })
  }

  return rows
}

watch(languageOptions, () => {
  if (typeForm.id === null && typeForm.translations.length === 0) {
    typeForm.translations = translationRowsFor()
  }
})

const resetForm = (): void => {
  Object.assign(typeForm, {
    id: null,
    name: '',
    slug: '',
    description: '',
    image_media_asset_id: null,
    image_url: '',
    pending_image_file: null,
    remove_image: false,
    sort_order: 0,
    is_enabled: true,
    translations: translationRowsFor(),
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

const showEditDialog = (type: ProductTypeRecord): void => {
  dialogMode.value = 'edit'
  showSpecAdvanced.value = false
  Object.assign(typeForm, {
    id: type.id,
    name: type.name || '',
    slug: type.slug || '',
    description: type.description || '',
    image_media_asset_id: type.image_media_asset_id ?? null,
    image_url: String(type.image_url || ''),
    pending_image_file: null,
    remove_image: false,
    sort_order: Number(type.sort_order || 0),
    is_enabled: type.is_enabled !== false,
    translations: translationRowsFor(type.translations || []),
    spec_definitions: (type.spec_definitions || []).map(apiSpecToForm)
  })
  clearFormErrors()
  dialogVisible.value = true
}

const addSpecDefinition = (): void => {
  const spec = createEmptySpec()
  spec.sort_order = typeForm.spec_definitions.length * 10
  typeForm.spec_definitions.push(spec)
}

const removeSpecDefinition = (index: number): void => {
  typeForm.spec_definitions.splice(index, 1)
  clearFormErrors()
}

const clearFormErrors = (): void => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (key: string): void => { delete formErrors[key] }

const handleImageSelected = (file: File): void => {
  typeForm.pending_image_file = file
  typeForm.remove_image = false
}

const handleImageCleared = (): void => {
  const hadExistingImage = Boolean(typeForm.image_url || typeForm.image_media_asset_id)
  typeForm.pending_image_file = null
  typeForm.image_url = ''
  typeForm.image_media_asset_id = null
  typeForm.remove_image = hadExistingImage
}

const handleImageError = (message: string): void => {
  toast.error(message)
}

const specOptionsFromText = (text?: string | null): string[] => String(text || '')
  .split(/\r?\n/)
  .map((value) => value.trim())
  .filter(Boolean)
  .filter((value, index, values) => values.indexOf(value) === index)

const specOptions = (spec: ProductTypeSpecForm): string[] => specOptionsFromText(spec.optionsText)

const specPayloadOptions = (spec: ProductTypeSpecForm | ProductTypeSpecDefinition): string[] => {
  if ('optionsText' in spec) return specOptionsFromText(spec.optionsText)
  return specOptionsFromText(optionsToText(spec.options))
}

const validateForm = (): boolean => {
  clearFormErrors()
  const slugPattern = /^[a-z0-9]+(?:[_-][a-z0-9]+)*$/
  if (!typeForm.name.trim()) formErrors.name = '请输入模板名称'
  if (!slugPattern.test(typeForm.slug.trim())) formErrors.slug = '请输入有效的模板标识'

  const seenSlugs = new Set<string>()
  typeForm.spec_definitions.forEach((spec, index) => {
    if (!spec.name.trim()) formErrors[`spec:${index}:name`] = '请输入字段名称'
    const slug = spec.slug.trim()
    if (!slugPattern.test(slug)) formErrors[`spec:${index}:slug`] = '请输入有效的字段标识'
    else if (seenSlugs.has(slug)) formErrors[`spec:${index}:slug`] = '字段标识不能重复'
    else seenSlugs.add(slug)
    if (spec.field_type === 'select' && specOptions(spec).length === 0 && !usesProductScopedOptions(spec)) {
      formErrors[`spec:${index}:options`] = '请至少填写一个选项'
    }
  })

  if (Object.keys(formErrors).length > 0) {
    toast.error('请检查产品模板表单')
    return false
  }
  return true
}

const buildPayload = (
  source: ProductTypeForm | ProductTypeRecord = typeForm,
  enabled = source.is_enabled !== false
): ProductTypePayload => ({
  name: String(source.name || '').trim(),
  slug: String(source.slug || '').trim().toLowerCase(),
  description: String(source.description || '').trim(),
  sort_order: Number(source.sort_order || 0),
  is_enabled: Boolean(enabled),
  translations: (source.translations || [])
    .map((translation) => ({
      id: Number(translation.id || 0),
      locale: String(translation.locale || '').trim(),
      name: String(translation.name || '').trim(),
      description: String(translation.description || '').trim()
    }))
    .filter((translation) => translation.locale && translation.name),
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

const fetchProductTypes = async (): Promise<void> => {
  loading.value = true
  try {
    productTypes.value = await productTypeApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product types:', error)
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
    let savedType: ProductTypeRecord | null = null
    if (dialogMode.value === 'create') {
      savedType = await productTypeApi.create(payload)
    } else if (typeForm.id !== null) {
      savedType = await productTypeApi.update(typeForm.id, payload)
    }

    if (!savedType?.id) {
      throw new Error('Product type save returned no record')
    }

    typeForm.id = savedType.id
    dialogMode.value = 'edit'

    const pendingImageFile = typeForm.pending_image_file
    const removingImage = typeForm.remove_image
    if (pendingImageFile || removingImage) {
      try {
        if (pendingImageFile) {
          const withImage = await productTypeApi.uploadImage(savedType.id, pendingImageFile)
          typeForm.image_media_asset_id = withImage.image_media_asset_id ?? null
          typeForm.image_url = String(withImage.image_url || '')
          typeForm.pending_image_file = null
          typeForm.remove_image = false
        } else {
          const withoutImage = await productTypeApi.deleteImage(savedType.id)
          typeForm.image_media_asset_id = withoutImage.image_media_asset_id ?? null
          typeForm.image_url = String(withoutImage.image_url || '')
          typeForm.remove_image = false
        }
      } catch (imageError) {
        console.error('Failed to sync product type image:', imageError)
        await fetchProductTypes()
        toast.warning(`产品模板已${wasCreating ? '创建' : '更新'}，但分类图片${pendingImageFile ? '上传' : '移除'}失败，请重试`)
        return
      }
    }

    toast.success(wasCreating ? '产品模板已创建' : '产品模板已更新')
    dialogVisible.value = false
    await fetchProductTypes()
  } catch (error) {
    console.error('Failed to save product type:', error)
    toast.error('产品模板保存失败，请检查名称、标识和字段设置')
  } finally {
    submitting.value = false
  }
}

const toggleType = async (type: ProductTypeRecord): Promise<void> => {
  try {
    await productTypeApi.update(type.id, buildPayload(type, !type.is_enabled))
    toast.success(type.is_enabled ? '产品模板已停用' : '产品模板已启用')
    await fetchProductTypes()
  } catch (error) {
    console.error('Failed to toggle product type:', error)
  }
}

const requestDelete = (type: ProductTypeRecord): void => {
  Object.assign(confirmation, { open: true, target: type })
}
const deleteType = async (): Promise<void> => {
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

const resetFilters = (): void => { Object.assign(filters, { search: '', status: 'all' }) }

onMounted(() => {
  void Promise.all([
    supportedLanguages.fetchLanguages(),
    fetchProductTypes()
  ])
})
</script>

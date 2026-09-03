<template>
  <div v-if="mode === 'manage'" class="space-y-4">
    <Card>
      <CardHeader class="border-b">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <CardTitle>{{ title }}</CardTitle>
            <CardDescription>{{ description }}</CardDescription>
          </div>
          <Button v-if="canCreate" size="sm" @click="openCreate">
            <Plus class="size-3.5" />
            添加品牌
          </Button>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <div v-if="loading" class="p-6 text-center text-sm text-muted-foreground">加载中...</div>
        <div v-else-if="brands.length" class="divide-y">
          <div v-for="brand in brands" :key="brand.id" class="flex items-start justify-between gap-4 p-4">
            <div class="flex min-w-0 items-start gap-3">
              <div class="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-md border bg-muted/20">
                <img v-if="brand.logo_url" :src="brand.logo_url" :alt="brand.name" class="size-full object-contain" loading="lazy">
                <Tag v-else class="size-4 text-muted-foreground" />
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="font-semibold">{{ brand.name }}</p>
                  <span
                    class="rounded-full px-2 py-0.5 text-[11px]"
                    :class="brand.is_enabled ? 'bg-emerald-500/10 text-emerald-700' : 'bg-muted text-muted-foreground'"
                  >
                    {{ brand.is_enabled ? '启用' : '停用' }}
                  </span>
                </div>
                <p class="mt-1 font-mono text-xs text-muted-foreground">{{ brand.slug }}</p>
                <p v-if="brand.website_url" class="mt-1 truncate text-xs text-muted-foreground">{{ brand.website_url }}</p>
                <p v-if="brand.description" class="mt-2 line-clamp-2 text-sm text-muted-foreground">{{ brand.description }}</p>
              </div>
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <Button v-if="canEdit" variant="ghost" size="icon" aria-label="编辑品牌" @click="openEdit(brand)">
                <Pencil class="size-4" />
              </Button>
              <Button v-if="canDelete" variant="ghost" size="icon" aria-label="删除品牌" @click="removeBrand(brand)">
                <Trash2 class="size-4 text-destructive" />
              </Button>
            </div>
          </div>
        </div>
        <p v-else class="p-6 text-center text-sm text-muted-foreground">暂无品牌，请先添加商品品牌。</p>
      </CardContent>
    </Card>
  </div>

  <Dialog v-model:open="dialogOpen">
    <DialogContent size="sm" class="max-h-[90dvh]">
      <template v-if="mode === 'picker' && dialogMode === 'select'">
        <DialogHeader>
          <DialogTitle>选择商品品牌</DialogTitle>
          <DialogDescription>品牌来自全局商品品牌主数据；选择后会在当前页面配置对应的轮圈型号。</DialogDescription>
        </DialogHeader>
        <div class="space-y-4 pt-1">
          <div v-if="loading" class="py-8 text-center text-sm text-muted-foreground">加载品牌中...</div>
          <template v-else>
            <AdminFormField v-if="availableBrands.length" label="商品品牌" required>
              <Select v-model="selectedBrandSlug">
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="选择品牌" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="brand in availableBrands"
                    :key="brand.id"
                    :value="brand.slug"
                  >
                    {{ brand.name }} · {{ brand.slug }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
            <p v-else class="rounded-lg border border-dashed p-4 text-center text-sm font-bold text-muted-foreground">
              暂无可添加的启用商品品牌
            </p>
          </template>
        </div>
        <DialogFooter>
          <Button v-if="canCreate" type="button" variant="outline" @click="openCreate">
            <Plus class="size-3.5" />
            新增商品品牌
          </Button>
          <Button type="button" :disabled="!selectedBrandSlug || !availableBrands.length" @click="confirmSelection">
            使用此品牌
          </Button>
        </DialogFooter>
      </template>

      <template v-else>
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑商品品牌' : '添加商品品牌' }}</DialogTitle>
          <DialogDescription>品牌名称会进入商品公开 JSON、结构化数据和 Google SEO 诊断。</DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="save">
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminFormField label="品牌名称" required>
              <Input v-model="form.name" placeholder="例如 DT Swiss" />
            </AdminFormField>
            <AdminFormField label="Slug" required description="用于品牌筛选和稳定的数据标识。">
              <Input v-model="form.slug" class="font-mono" placeholder="例如 dt-swiss" />
            </AdminFormField>
            <AdminFormField label="Logo URL">
              <Input v-model="form.logo_url" type="url" placeholder="https://..." />
            </AdminFormField>
            <AdminFormField label="官网 URL">
              <Input v-model="form.website_url" type="url" placeholder="https://..." />
            </AdminFormField>
            <AdminFormField label="排序">
              <Input v-model.number="form.sort_order" type="number" min="0" />
            </AdminFormField>
            <div class="flex items-center justify-between gap-4 rounded-lg border px-3 py-2.5 sm:mt-6">
              <span class="text-sm">启用品牌</span>
              <Switch v-model="form.is_enabled" />
            </div>
          </div>
          <AdminFormField label="品牌描述">
            <Textarea v-model="form.description" class="min-h-24" placeholder="可选，供后台和 SEO 数据使用。" />
          </AdminFormField>
          <DialogFooter>
            <Button type="button" variant="outline" @click="closeDialog">取消</Button>
            <Button type="submit" :disabled="saving">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              保存
            </Button>
          </DialogFooter>
        </form>
      </template>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Pencil, Plus, Tag, Trash2 } from '@lucide/vue'
import { productBrandApi, type ProductBrandRecord } from '@/api/products'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'

type ProductBrandManagerMode = 'manage' | 'picker'
type ProductBrandDialogMode = 'select' | 'create' | 'edit'

const props = withDefaults(defineProps<{
  mode?: ProductBrandManagerMode
  open?: boolean
  excludedSlugs?: string[]
  title?: string
  description?: string
}>(), {
  mode: 'manage',
  open: false,
  excludedSlugs: () => [],
  title: '品牌目录',
  description: '停用品牌不会影响已绑定商品；删除只允许用于尚未被商品引用的品牌。',
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'select', brand: ProductBrandRecord): void
  (event: 'brands-loaded', brands: ProductBrandRecord[]): void
}>()

const authStore = useAuthStore()
const brands = ref<ProductBrandRecord[]>([])
const loading = ref(false)
const saving = ref(false)
const internalDialogOpen = ref(false)
const dialogMode = ref<ProductBrandDialogMode>(props.mode === 'picker' ? 'select' : 'create')
const selectedBrandSlug = ref('')
const form = reactive<Partial<ProductBrandRecord>>({})

const canCreate = computed(() => authStore.hasPermission('product:create'))
const canEdit = computed(() => authStore.hasPermission('product:edit'))
const canDelete = computed(() => authStore.hasPermission('product:delete'))
const dialogOpen = computed({
  get: () => props.mode === 'picker' ? props.open : internalDialogOpen.value,
  set: (value: boolean) => {
    if (props.mode === 'picker') {
      emit('update:open', value)
    } else {
      internalDialogOpen.value = value
    }
    if (!value) dialogMode.value = props.mode === 'picker' ? 'select' : 'create'
  },
})

const normalizedSlug = (value: string) => value.trim().toLowerCase()

const availableBrands = computed(() => {
  const excluded = new Set(props.excludedSlugs.map(normalizedSlug))
  return brands.value.filter((brand) => (
    brand.is_enabled !== false &&
    !excluded.has(normalizedSlug(brand.slug))
  ))
})

const resetForm = () => Object.assign(form, {
  id: undefined,
  name: '',
  slug: '',
  description: '',
  logo_url: '',
  website_url: '',
  is_enabled: true,
  sort_order: 0,
})

const fetchBrands = async () => {
  loading.value = true
  try {
    brands.value = await productBrandApi.list({ include_disabled: true })
    emit('brands-loaded', brands.value)
    if (!availableBrands.value.some((brand) => normalizedSlug(brand.slug) === normalizedSlug(selectedBrandSlug.value))) {
      selectedBrandSlug.value = availableBrands.value[0]?.slug || ''
    }
  } catch (error) {
    console.error('Failed to fetch product brands:', error)
    toast.error('品牌加载失败')
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  if (!canCreate.value) return
  resetForm()
  dialogMode.value = 'create'
  dialogOpen.value = true
}

const openEdit = (brand: ProductBrandRecord) => {
  if (!canEdit.value) return
  Object.assign(form, brand)
  dialogMode.value = 'edit'
  dialogOpen.value = true
}

const closeDialog = () => {
  if (props.mode === 'picker' && dialogMode.value !== 'select') {
    dialogMode.value = 'select'
    return
  }
  dialogOpen.value = false
}

const save = async () => {
  if (!form.name?.trim() || !form.slug?.trim()) {
    toast.error('请填写品牌名称和 slug')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      slug: form.slug.trim(),
      description: form.description?.trim() || '',
      logo_url: form.logo_url?.trim() || '',
      website_url: form.website_url?.trim() || '',
      is_enabled: form.is_enabled !== false,
      sort_order: Number(form.sort_order || 0),
    }
    const savedBrand = form.id
      ? await productBrandApi.update(form.id, payload)
      : await productBrandApi.create(payload)
    await fetchBrands()
    const nextBrand = brands.value.find((brand) => brand.id === Number(savedBrand.id))
      || brands.value.find((brand) => normalizedSlug(brand.slug) === normalizedSlug(payload.slug))

    if (props.mode === 'picker' && dialogMode.value === 'create' && nextBrand) {
      emit('select', nextBrand)
      dialogOpen.value = false
    } else {
      dialogOpen.value = false
    }
    toast.success('品牌已保存')
  } catch (error) {
    console.error('Failed to save product brand:', error)
  } finally {
    saving.value = false
  }
}

const removeBrand = async (brand: ProductBrandRecord) => {
  if (!canDelete.value || !window.confirm(`确定删除品牌“${brand.name}”？已被商品引用的品牌不能删除。`)) return
  try {
    await productBrandApi.remove(brand.id)
    toast.success('品牌已删除')
    await fetchBrands()
  } catch (error) {
    console.error('Failed to delete product brand:', error)
  }
}

const confirmSelection = () => {
  const selected = availableBrands.value.find((brand) => normalizedSlug(brand.slug) === normalizedSlug(selectedBrandSlug.value))
  if (!selected) {
    toast.error('请选择有效的商品品牌')
    return
  }
  emit('select', selected)
  dialogOpen.value = false
}

watch(() => props.open, (open) => {
  if (open && props.mode === 'picker') {
    dialogMode.value = 'select'
    selectedBrandSlug.value = availableBrands.value[0]?.slug || ''
  }
})

watch(availableBrands, (nextBrands) => {
  if (!nextBrands.some((brand) => normalizedSlug(brand.slug) === normalizedSlug(selectedBrandSlug.value))) {
    selectedBrandSlug.value = nextBrands[0]?.slug || ''
  }
})

onMounted(() => {
  void fetchBrands()
})

defineExpose({
  refresh: fetchBrands,
  openCreate,
  openEdit,
  getBrands: () => brands.value,
})
</script>

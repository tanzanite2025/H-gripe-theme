<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品品牌" description="维护所有商品共用的品牌主数据，商品详情和 SEO 会直接读取这里的品牌。">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">商品管理</RouterLink>
        </Button>
        <Button v-if="hasPermission('product:create')" @click="openCreate">
          <Plus class="size-4" />
          添加品牌
        </Button>
      </template>
    </AdminPageHeader>

    <Card>
      <CardHeader class="border-b">
        <CardTitle>品牌目录</CardTitle>
        <CardDescription>停用品牌不会影响已绑定商品；删除只允许用于尚未被商品引用的品牌。</CardDescription>
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
              <Button v-if="hasPermission('product:edit')" variant="ghost" size="icon" aria-label="编辑品牌" @click="openEdit(brand)">
                <Pencil class="size-4" />
              </Button>
              <Button v-if="hasPermission('product:delete')" variant="ghost" size="icon" aria-label="删除品牌" @click="removeBrand(brand)">
                <Trash2 class="size-4 text-destructive" />
              </Button>
            </div>
          </div>
        </div>
        <p v-else class="p-6 text-center text-sm text-muted-foreground">暂无品牌，请先添加商品品牌。</p>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-2xl">
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
            <Button type="button" variant="outline" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import { LoaderCircle, Pencil, Plus, Tag, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { productBrandApi } from '@/api/products'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'

interface ProductBrandRecord {
  id: number
  name: string
  slug: string
  description: string
  logo_url: string
  website_url: string
  is_enabled: boolean
  sort_order: number
}

const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const brands = ref<ProductBrandRecord[]>([])
const form = reactive<Partial<ProductBrandRecord>>({
  id: undefined,
  name: '',
  slug: '',
  description: '',
  logo_url: '',
  website_url: '',
  is_enabled: true,
  sort_order: 0,
})

const hasPermission = (permission: string) => authStore.hasPermission(permission)

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

const openCreate = () => {
  resetForm()
  dialogOpen.value = true
}

const openEdit = (brand: ProductBrandRecord) => {
  Object.assign(form, brand)
  dialogOpen.value = true
}

const fetchBrands = async () => {
  loading.value = true
  try {
    brands.value = await productBrandApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product brands:', error)
    toast.error('品牌加载失败')
  } finally {
    loading.value = false
  }
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
    if (form.id) {
      await productBrandApi.update(form.id, payload)
    } else {
      await productBrandApi.create(payload)
    }
    dialogOpen.value = false
    toast.success('品牌已保存')
    await fetchBrands()
  } catch (error) {
    console.error('Failed to save product brand:', error)
  } finally {
    saving.value = false
  }
}

const removeBrand = async (brand: ProductBrandRecord) => {
  if (!window.confirm(`确定删除品牌“${brand.name}”？已被商品引用的品牌不能删除。`)) return
  try {
    await productBrandApi.remove(brand.id)
    toast.success('品牌已删除')
    await fetchBrands()
  } catch (error) {
    console.error('Failed to delete product brand:', error)
  }
}

onMounted(() => {
  void fetchBrands()
})
</script>

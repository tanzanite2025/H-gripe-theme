<template>
  <div class="space-y-4">
    <AdminPageHeader title="产品信息模板" description="维护商品详情页的 After-sales 与 Packaging 内容模板">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">商品管理</RouterLink>
        </Button>
        <Button v-if="hasPermission('product:create')" @click="openCreate">
          <Plus class="size-4" />
          添加模板
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid gap-4 xl:grid-cols-2">
      <Card v-for="group in groups" :key="group.kind" class="overflow-hidden">
        <CardHeader class="border-b">
          <CardTitle>{{ group.label }}</CardTitle>
          <CardDescription>{{ group.description }}</CardDescription>
        </CardHeader>
        <CardContent class="p-0">
          <div v-if="group.items.length" class="divide-y">
            <div v-for="item in group.items" :key="item.id" class="flex items-start justify-between gap-3 p-4">
              <div class="min-w-0">
                <p class="font-semibold">{{ item.name }}</p>
                <p class="mt-1 font-mono text-xs text-muted-foreground">{{ item.slug }} · {{ localeName(item.locale) }}</p>
                <p class="mt-2 line-clamp-2 text-sm text-muted-foreground" v-html="item.content || '暂无内容'" />
              </div>
              <div class="flex shrink-0 items-center gap-1">
                <Button v-if="hasPermission('product:edit')" variant="ghost" size="icon" aria-label="编辑模板" @click="openEdit(item)">
                  <Pencil class="size-4" />
                </Button>
                <Button v-if="hasPermission('product:delete')" variant="ghost" size="icon" aria-label="删除模板" @click="remove(item)">
                  <Trash2 class="size-4 text-destructive" />
                </Button>
              </div>
            </div>
          </div>
          <p v-else class="p-6 text-center text-sm text-muted-foreground">暂无模板</p>
        </CardContent>
      </Card>
    </div>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑产品信息模板' : '添加产品信息模板' }}</DialogTitle>
          <DialogDescription>内容会经过后台富文本净化后保存，并在商品详情页对应标签中展示。</DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="save">
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminFormField label="模板类型" required>
              <Select v-model="form.kind" :disabled="Boolean(form.id)">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="after_sales">After-sales</SelectItem>
                  <SelectItem value="packaging">Packaging</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
            <AdminFormField
              label="语言"
              required
              :description="form.id ? '编辑模板时语言已锁定；如需其他语言，请新建对应语种模板。' : '固定为前端支持的 20 个语种。'"
            >
              <StorefrontLocaleSelect
                v-model="form.locale"
                :language-options="languageOptions"
                :loading="languageLoading"
                :disabled="Boolean(form.id)"
                :locked="Boolean(form.id)"
                locked-title="模板语言已锁定"
              />
            </AdminFormField>
            <AdminFormField label="模板名称" required>
              <Input v-model="form.name" />
            </AdminFormField>
            <AdminFormField label="模板标识" required>
              <Input v-model="form.slug" class="font-mono" />
            </AdminFormField>
          </div>
          <AdminFormField label="HTML 内容" description="支持 p、h2-h4、列表、链接、图片和视频等安全标签。">
            <ProductDescriptionEditor v-model="form.content" />
          </AdminFormField>
          <div class="flex items-center justify-between gap-4 rounded-lg border px-3 py-2.5">
            <span class="text-sm">启用模板</span>
            <Switch v-model="form.is_enabled" />
          </div>
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
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import { LoaderCircle, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import ProductDescriptionEditor from '@/components/admin/product/ProductDescriptionEditor.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { productInformationTemplateApi } from '@/api/products'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { normalizeLocaleCode } from '@/lib/languages'
import { useAuthStore } from '@/stores/auth'

interface TemplateRecord {
  id: number
  kind: 'after_sales' | 'packaging'
  name: string
  slug: string
  content: string
  locale: string
  is_enabled: boolean
  sort_order: number
}

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const items = ref<TemplateRecord[]>([])
const form = reactive<Partial<TemplateRecord>>({
  id: undefined,
  kind: 'after_sales',
  name: '',
  slug: '',
  content: '',
  locale: 'en',
  is_enabled: true,
  sort_order: 0
})

const languageOptions = supportedLanguages.languageOptions
const languageLoading = supportedLanguages.loading
const localeName = supportedLanguages.localeName
const supportedLocaleCodes = computed(() => new Set(languageOptions.value.map((language) => language.value)))
const defaultTemplateLocale = computed(() => supportedLocaleCodes.value.has('en')
  ? 'en'
  : languageOptions.value[0]?.value || 'en')

const hasPermission = (permission: string) => authStore.hasPermission(permission)
const groups = computed(() => [
  {
    kind: 'after_sales',
    label: 'After-sales',
    description: '商品详情页的售后服务内容。',
    items: items.value.filter((item) => item.kind === 'after_sales')
  },
  {
    kind: 'packaging',
    label: 'Packaging',
    description: '商品详情页的包装说明内容。',
    items: items.value.filter((item) => item.kind === 'packaging')
  }
])

const resetForm = () => Object.assign(form, {
  id: undefined,
  kind: 'after_sales',
  name: '',
  slug: '',
  content: '',
  locale: defaultTemplateLocale.value,
  is_enabled: true,
  sort_order: 0
})

const normalizeTemplateLocale = (locale?: string | null) => {
  const normalized = normalizeLocaleCode(locale)
  return supportedLocaleCodes.value.has(normalized) ? normalized : ''
}

const openCreate = () => {
  resetForm()
  dialogOpen.value = true
}

const openEdit = (item: TemplateRecord) => {
  Object.assign(form, {
    ...item,
    locale: normalizeTemplateLocale(item.locale) || ''
  })
  dialogOpen.value = true
}

const fetchItems = async () => {
  loading.value = true
  try {
    items.value = await productInformationTemplateApi.list({ include_disabled: true })
  } catch (error) {
    console.error('Failed to fetch product information templates:', error)
    toast.error('模板加载失败')
  } finally {
    loading.value = false
  }
}

const save = async () => {
  if (!form.name?.trim() || !form.slug?.trim()) {
    toast.error('请填写模板名称和标识')
    return
  }
  const locale = normalizeTemplateLocale(form.locale)
  if (!locale) {
    toast.error('请选择前端支持的固定语种')
    return
  }
  saving.value = true
  try {
    const payload = {
      kind: form.kind,
      name: form.name,
      slug: form.slug,
      content: form.content || '',
      locale,
      is_enabled: form.is_enabled !== false,
      sort_order: Number(form.sort_order || 0)
    }
    if (form.id) {
      await productInformationTemplateApi.update(form.id, payload)
    } else {
      await productInformationTemplateApi.create(payload)
    }
    dialogOpen.value = false
    toast.success('模板已保存')
    await fetchItems()
  } catch (error) {
    console.error('Failed to save product information template:', error)
  } finally {
    saving.value = false
  }
}

const remove = async (item: TemplateRecord) => {
  if (!window.confirm(`确定删除模板“${item.name}”？已绑定商品会自动取消绑定。`)) return
  try {
    await productInformationTemplateApi.remove(item.id)
    toast.success('模板已删除')
    await fetchItems()
  } catch (error) {
    console.error('Failed to delete product information template:', error)
  }
}

onMounted(() => {
  void Promise.all([supportedLanguages.fetchLanguages(), fetchItems()])
})
</script>

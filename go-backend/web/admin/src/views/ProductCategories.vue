<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品分类" description="独立维护商品分类树，最多 5 级。">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">商品管理</RouterLink>
        </Button>
        <Button variant="outline" :disabled="loading || savingAll" @click="fetchCategories">
          <RefreshCcw class="size-4" />
          重新加载
        </Button>
        <Button v-if="hasPermission('product:create')" variant="outline" :disabled="savingAll" @click="addRootCategory">
          <Plus class="size-4" />
          添加顶级分类
        </Button>
        <Button v-if="hasPermission('product:edit') || hasPermission('product:create')" :disabled="!hasChanges || savingAll" @click="saveAll">
          <LoaderCircle v-if="savingAll" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          保存全部
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid gap-3 md:grid-cols-5">
      <Card v-for="item in statCards" :key="item.label">
        <CardHeader class="space-y-1 pb-3">
          <CardDescription>{{ item.label }}</CardDescription>
          <CardTitle class="text-2xl">{{ item.value }}</CardTitle>
        </CardHeader>
      </Card>
    </div>

    <Card>
      <CardHeader class="border-b">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <CardTitle class="flex items-center gap-2">
              <FolderTree class="size-5" />
              分类树
            </CardTitle>
            <CardDescription>直接在页面上编辑整棵树；新增的父级和子级可以一起保存。</CardDescription>
          </div>
          <Input v-model="search" class="max-w-sm" placeholder="搜索名称、slug 或描述" />
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <div v-if="loading" class="flex items-center justify-center gap-2 p-8 text-sm text-muted-foreground">
          <LoaderCircle class="size-4 animate-spin" />
          加载中...
        </div>
        <div v-else>
          <div class="product-category-tree-header hidden border-b bg-muted/30 px-4 py-2 text-xs font-medium text-muted-foreground xl:grid">
            <div>名称</div>
            <div>分类图片</div>
            <div>Slug</div>
            <div>父级</div>
            <div>描述</div>
            <div>翻译</div>
            <div>状态</div>
            <div class="text-right">操作</div>
          </div>

          <div v-if="visibleRows.length" class="divide-y">
            <ProductCategoryTreeRow
              v-for="row in visibleRows"
              :key="row.key"
              :row="row"
              :parent-options="parentOptions(row)"
              :max-depth="maxDepth"
              :saving="savingAll"
              :can-create="hasPermission('product:create')"
              :can-delete="hasPermission('product:delete')"
              :can-translate="hasPermission('product:edit')"
              @add-sibling="addSibling"
              @add-child="addChild"
              @change-parent="changeParent"
              @pick-image="openImagePicker"
              @clear-image="clearImage"
              @delete="requestDelete"
              @edit-translations="openTranslationDialog"
              @normalize-slug="normalizeSlug"
              @mark-dirty="markDirty"
            />
          </div>
          <div v-else class="p-8 text-center text-sm text-muted-foreground">
            {{ search.trim() ? '没有匹配的分类。' : '暂无分类，先添加一个顶级分类。' }}
          </div>
        </div>
      </CardContent>
    </Card>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      title="删除商品分类？"
      :description="`分类“${confirmation.target?.name || '未命名分类'}”会被删除；已保存分类必须先没有子分类。`"
      confirm-label="删除"
      destructive
      @confirm="deleteCategory"
    />

    <ProductCategoryTranslationDialog
      v-model:open="translationEditor.open"
      :category-name="translationEditor.category?.name"
      :translations="translationEditor.translations"
      :language-options="languageOptions"
      :loading="translationEditor.loading"
      :saving="translationEditor.saving"
      @save="saveTranslations"
    />

    <MediaAssetPickerDialog
      v-model:open="mediaPickerVisible"
      :selected-urls="selectedMediaUrls"
      @select="selectCategoryImage"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import { FolderTree, LoaderCircle, Plus, RefreshCcw, Save } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import MediaAssetPickerDialog from '@/components/admin/media/MediaAssetPickerDialog.vue'
import ProductCategoryTreeRow from '@/components/admin/product/ProductCategoryTreeRow.vue'
import ProductCategoryTranslationDialog from '@/components/admin/product/ProductCategoryTranslationDialog.vue'
import type {
  DraftCategoryRow,
  ProductCategoryTranslationPayload,
} from '@/modules/product/productCategoryTypes'
import { useProductCategoryTreeEditor } from '@/composables/product/useProductCategoryTreeEditor'
import productCategoryApi, { type ProductCategoryTranslationRecord } from '@/api/productCategories'
import type { MediaAsset } from '@/api/media'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const search = ref('')
const mediaPickerVisible = ref(false)
const mediaPickerTarget = ref<DraftCategoryRow | null>(null)
const confirmation = reactive<{ open: boolean; target: DraftCategoryRow | null }>({ open: false, target: null })
const translationEditor = reactive<{
  open: boolean
  loading: boolean
  saving: boolean
  category: DraftCategoryRow | null
  translations: ProductCategoryTranslationRecord[]
}>({
  open: false,
  loading: false,
  saving: false,
  category: null,
  translations: [],
})
const {
  loading,
  savingAll,
  rows,
  maxDepth,
  stats,
  hasChanges,
  addRootCategory,
  addSibling,
  addChild,
  changeParent,
  fetchCategories,
  markDirty,
  normalizeSlug,
  parentOptions,
  removeUnsavedRow,
  removeUnsavedSubtree,
  saveAll,
  subtreeKeys,
} = useProductCategoryTreeEditor()

const hasPermission = (permission: string) => authStore.hasPermission(permission)
const selectedMediaUrls = computed(() => (
  mediaPickerTarget.value?.image_url ? [mediaPickerTarget.value.image_url] : []
))
const statCards = computed(() => [
  { label: '分类总数', value: stats.value.total },
  { label: '已启用', value: stats.value.enabled },
  { label: '待保存', value: stats.value.changed },
  { label: '翻译待完善', value: stats.value.translationIncomplete },
  { label: '当前最深层级', value: `${stats.value.deepestLevel} / ${stats.value.maxDepth}` },
])
const visibleRows = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return rows.value
  return rows.value.filter((row) => (
    row.name.toLowerCase().includes(keyword)
    || row.slug.toLowerCase().includes(keyword)
    || row.description.toLowerCase().includes(keyword)
  ))
})

const openImagePicker = (row: DraftCategoryRow): void => {
  mediaPickerTarget.value = row
  mediaPickerVisible.value = true
}

const selectCategoryImage = (selection: { url: string; asset: MediaAsset }): void => {
  const row = mediaPickerTarget.value
  const assetID = Number(selection.asset?.id)
  if (!row || !Number.isFinite(assetID) || assetID <= 0 || !selection.url) return
  row.image_media_asset_id = assetID
  row.image_url = selection.url
  markDirty(row)
  mediaPickerVisible.value = false
}

const clearImage = (row: DraftCategoryRow): void => {
  row.image_media_asset_id = null
  row.image_url = ''
  markDirty(row)
}

const requestDelete = (row: DraftCategoryRow) => {
  const keys = subtreeKeys(row.key)
  if (keys.size > 1) {
    if (removeUnsavedSubtree(row)) return
    toast.error('已保存分类有下级时，请先删除或移动下级分类')
    return
  }
  if (removeUnsavedRow(row)) return
  Object.assign(confirmation, { open: true, target: row })
}

const deleteCategory = async () => {
  const row = confirmation.target
  confirmation.open = false
  if (!row?.id) return
  try {
    await productCategoryApi.remove(row.id)
    toast.success('商品分类已删除')
    await fetchCategories()
  } catch (error) {
    console.error('Failed to delete product category:', error)
    toast.error('商品分类删除失败，请先删除或移动它的子分类')
  } finally {
    confirmation.target = null
  }
}

const openTranslationDialog = async (row: DraftCategoryRow): Promise<void> => {
  if (!row.id) return
  Object.assign(translationEditor, {
    open: true,
    loading: true,
    saving: false,
    category: row,
    translations: [],
  })
  try {
    translationEditor.translations = await productCategoryApi.translations(row.id)
  } catch (error) {
    console.error('Failed to fetch product category translations:', error)
    toast.error('分类翻译加载失败')
    translationEditor.open = false
  } finally {
    translationEditor.loading = false
  }
}

const saveTranslations = async (translations: ProductCategoryTranslationPayload[]): Promise<void> => {
  if (!translationEditor.category?.id) return
  translationEditor.saving = true
  try {
    translationEditor.translations = await productCategoryApi.updateTranslations(
      translationEditor.category.id,
      translations,
    )
    translationEditor.open = false
    toast.success('分类翻译已保存')
  } catch (error) {
    console.error('Failed to save product category translations:', error)
    toast.error('分类翻译保存失败')
  } finally {
    translationEditor.saving = false
  }
}

onMounted(() => {
  void Promise.all([
    supportedLanguages.fetchLanguages(),
    fetchCategories(),
  ])
})
</script>

<style scoped>
.product-category-tree-header {
  width: 100%;
  min-width: 0;
  grid-template-columns: minmax(220px, 1.3fr) minmax(128px, 0.5fr) minmax(140px, 0.8fr) minmax(160px, 0.9fr) minmax(160px, 1fr) minmax(92px, 0.5fr) 56px minmax(140px, max-content);
  gap: 0.75rem;
}
</style>


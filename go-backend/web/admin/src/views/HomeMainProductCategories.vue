<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="首页主力产品"
      description="管理首页本店精选展示，支持按语言配置、排序、跳转和上下架"
    >
      <template #actions>
        <Button
          variant="outline"
          size="sm"
          :disabled="loading || saving"
          @click="loadItems"
        >
          <LoaderCircle v-if="loading" class="size-3.5 animate-spin" />
          <RefreshCw v-else class="size-3.5" />
          刷新
        </Button>
        <Button
          v-if="canEdit"
          size="sm"
          :disabled="saving || loading"
          @click="save"
        >
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ saving ? '保存中' : '保存主力产品' }}
        </Button>
      </template>
    </AdminPageHeader>

    <Card>
      <CardContent class="grid gap-4 md:grid-cols-[minmax(0,18rem)_minmax(0,1fr)] md:items-end">
        <AdminFormField
          label="内容语言"
          description="首页会按当前前台语种读取对应配置。"
        >
          <StorefrontLocaleSelect
            v-model="locale"
            :language-options="languageOptions"
            :disabled="loading || saving"
          />
        </AdminFormField>
        <div class="rounded-xl border border-dashed border-emerald-500/30 bg-emerald-500/5 px-3 py-2.5">
          <p class="text-xs font-black uppercase tracking-wider text-emerald-700 dark:text-emerald-300">
            前台发布逻辑
          </p>
          <p class="mt-1 text-[11px] leading-relaxed text-muted-foreground">
            未配置或全部关闭时，前台会保留模块框架并显示未配置提示；配置保存后只展示发布状态打开的卡片。
          </p>
        </div>
      </CardContent>
    </Card>

    <Card class="overflow-hidden">
      <CardHeader class="border-b">
        <CardTitle>6 个本店精选位</CardTitle>
        <CardDescription>
          图片沿用 visual-showcase 专用上传目录，必须使用 16:9 横图；链接可指向商品详情、全部商品页或外部地址。
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-2.5 p-2.5 sm:p-3">
        <div class="rounded-xl border border-dashed border-border bg-muted/30 px-2.5 py-1.5">
          <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/70">
            统一上传规范
          </p>
          <UploadSpecHint code="visual_showcase_home_categories" compact />
        </div>
        <div v-if="loading" class="flex min-h-64 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载首页主力产品配置
        </div>
        <div v-else class="space-y-2.5">
          <HomeMainProductCategoryEditorRow
            v-for="(item, index) in items"
            :key="item.client_id"
            :item="item"
            :index="index"
            :can-edit="canEdit"
            :uploading="uploadingIndex === index"
            @update:item="updateItem(index, $event)"
            @upload-image="uploadImage"
          />
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import HomeMainProductCategoryEditorRow from '@/components/admin/visual-showcase/HomeMainProductCategoryEditorRow.vue'
import {
  applyVisualShowcaseUploadToFormState,
  createVisualShowcaseHomeMainProductCategoryAdministrationItemFormState,
  visualShowcaseHomeMainProductCategoriesAdministrationRowsFromApiItems,
  visualShowcaseHomeMainProductCategoriesAdministrationSavePayloadFromFormRow,
  visualShowcaseHomeMainProductCategoriesAdministrationValidationMessage,
} from '@/modules/visual-showcase/visualShowcaseFormState'
import {
  HOME_MAIN_PRODUCT_CATEGORIES_REQUIRED_ITEM_COUNT,
  HOME_MAIN_PRODUCT_CATEGORIES_SHOWCASE_KEY,
} from '@/modules/visual-showcase/visualShowcaseTypes'
import type {
  VisualShowcaseAdministrationItemFormState,
  VisualShowcaseAdministrationUploadRequest,
} from '@/modules/visual-showcase/visualShowcaseTypes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/stores/auth'
import { buildLanguageOptions, STOREFRONT_SUPPORTED_LANGUAGES } from '@/lib/languages'
import visualShowcaseApi from '@/api/visualShowcase'

const authStore = useAuthStore()
const languageOptions = buildLanguageOptions(STOREFRONT_SUPPORTED_LANGUAGES)
const locale = ref('en')
const items = ref<VisualShowcaseAdministrationItemFormState[]>(
  Array.from(
    { length: HOME_MAIN_PRODUCT_CATEGORIES_REQUIRED_ITEM_COUNT },
    (_, index) => createVisualShowcaseHomeMainProductCategoryAdministrationItemFormState(index),
  ),
)
const loading = ref(false)
const saving = ref(false)
const uploadingIndex = ref<number | null>(null)

const canEdit = computed(() => authStore.hasPermission('content:edit'))

const updateItem = (index: number, item: VisualShowcaseAdministrationItemFormState): void => {
  items.value = items.value.map((current, itemIndex) => itemIndex === index ? item : current)
}

const loadItems = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await visualShowcaseApi.getItems(HOME_MAIN_PRODUCT_CATEGORIES_SHOWCASE_KEY, locale.value)
    items.value = visualShowcaseHomeMainProductCategoriesAdministrationRowsFromApiItems(response.items)
  } catch (error) {
    console.error('Failed to load home main product categories:', error)
    toast.error('首页主力产品配置加载失败')
    items.value = visualShowcaseHomeMainProductCategoriesAdministrationRowsFromApiItems([])
  } finally {
    loading.value = false
  }
}

const uploadImage = async ({ index, file }: VisualShowcaseAdministrationUploadRequest): Promise<void> => {
  if (!canEdit.value || uploadingIndex.value !== null) return

  uploadingIndex.value = index
  try {
    const upload = await visualShowcaseApi.uploadImage(HOME_MAIN_PRODUCT_CATEGORIES_SHOWCASE_KEY, locale.value, file)
    const current = items.value[index]
    if (current) updateItem(index, applyVisualShowcaseUploadToFormState(current, upload))
    toast.success(`第 ${index + 1} 个入口图片已上传，保存后生效`)
  } catch (error) {
    console.error('Failed to upload home main product category image:', error)
    toast.error('图片上传失败，请确认图片为 16:9 比例')
  } finally {
    uploadingIndex.value = null
  }
}

const save = async (): Promise<void> => {
  if (!canEdit.value) return

  const validationMessage = visualShowcaseHomeMainProductCategoriesAdministrationValidationMessage(items.value)
  if (validationMessage) {
    toast.error(validationMessage)
    return
  }

  saving.value = true
  try {
    const response = await visualShowcaseApi.replaceItems(
      HOME_MAIN_PRODUCT_CATEGORIES_SHOWCASE_KEY,
      locale.value,
      items.value.map((item, index) => visualShowcaseHomeMainProductCategoriesAdministrationSavePayloadFromFormRow(item, index)),
    )
    items.value = visualShowcaseHomeMainProductCategoriesAdministrationRowsFromApiItems(response.items)
    toast.success('首页主力产品已保存')
  } catch (error) {
    console.error('Failed to save home main product categories:', error)
    toast.error('首页主力产品保存失败')
  } finally {
    saving.value = false
  }
}

watch(locale, () => {
  if (!loading.value && !saving.value) void loadItems()
})

onMounted(() => {
  void loadItems()
})
</script>


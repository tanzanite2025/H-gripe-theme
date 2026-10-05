<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="首页视觉目录"
      description="管理首页首屏最多 9 张 1:1 展示图，不进入媒体库；可以只上传部分图片"
    >
      <template #actions>
        <Button
          variant="outline"
          size="sm"
          :disabled="loading || busy"
          @click="loadItems"
        >
          <LoaderCircle v-if="loading" class="size-3.5 animate-spin" />
          <RefreshCw v-else class="size-3.5" />
          刷新
        </Button>
        <Button
          v-if="canEdit"
          size="sm"
          :disabled="busy || loading"
          @click="save"
        >
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ saving ? '保存中' : '保存整组配置' }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid gap-4 md:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)] md:items-stretch">
      <AdminStorefrontLanguageDisplayCard
        :model-value="locale"
        :language-options="languageOptions"
        :disabled="loading || busy"
        :loading="loading || busy"
        aria-label="首页视觉目录内容语言"
        @update:model-value="locale = $event"
      />
      <Card>
        <CardContent class="flex h-full items-center">
          <div class="rounded-xl border border-dashed border-emerald-500/30 bg-emerald-500/5 px-3 py-2.5">
            <p class="text-xs font-black uppercase tracking-wider text-emerald-700 dark:text-emerald-300">
              专用对象生命周期
            </p>
            <p class="mt-1 text-[11px] leading-relaxed text-muted-foreground">
              上传文件只写入 visual-showcase 专用目录，不创建媒体库记录。保存新配置后，不再引用的旧文件会被删除。
            </p>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card class="overflow-hidden">
      <CardHeader class="border-b">
        <CardTitle>最多 9 张首页展示图</CardTitle>
        <CardDescription>
          图片必须为 600×600 px；可以只配置部分位置，未上传的位置不会在前台显示。左侧编号就是前台位置，不需要手动排序。每张卡片可以单独保存，也可以使用右上角按钮一次保存整组配置。白色文案条使用标题和备注，ALT 文本用于图片可访问性与搜索引擎上下文。
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-3 p-3 sm:p-4">
        <div class="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2">
          <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/70">
            统一上传规范
          </p>
          <UploadSpecHint code="visual_showcase_home_hero" compact />
        </div>
        <div v-if="loading" class="flex min-h-64 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载首页视觉目录
        </div>
        <div v-else class="space-y-3">
          <VisualShowcaseItemEditorRow
            v-for="(item, index) in items"
            :key="item.client_id"
            :item="item"
            :index="index"
            :can-edit="canEdit"
            :uploading="uploadingIndex === index"
            :saving="savingIndex === index"
            :busy="busy"
            @update:item="updateItem(index, $event)"
            @upload-image="uploadImage"
            @save-card-configuration="saveSingleVisualShowcaseCardConfiguration(index)"
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
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStorefrontLanguageDisplayCard from '@/components/admin/AdminStorefrontLanguageDisplayCard.vue'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import VisualShowcaseItemEditorRow from '@/components/admin/visual-showcase/VisualShowcaseItemEditorRow.vue'
import {
  createVisualShowcaseHomeHeroAdministrationItemFormState,
  applyVisualShowcaseUploadToFormState,
  visualShowcaseHomeHeroAdministrationRowsFromApiItems,
  visualShowcaseHomeHeroAdministrationSavePayloadFromFormRow,
  visualShowcaseHomeHeroAdministrationValidationMessage,
} from '@/modules/visual-showcase/visualShowcaseFormState'
import type {
  VisualShowcaseAdministrationItemFormState,
  VisualShowcaseAdministrationUploadRequest,
} from '@/modules/visual-showcase/visualShowcaseTypes'
import { HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT } from '@/modules/visual-showcase/visualShowcaseTypes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/stores/auth'
import { buildLanguageOptions, STOREFRONT_SUPPORTED_LANGUAGES } from '@/lib/languages'
import visualShowcaseApi from '@/api/visualShowcase'

const SHOWCASE_KEY = 'home-hero'

const authStore = useAuthStore()
const languageOptions = buildLanguageOptions(STOREFRONT_SUPPORTED_LANGUAGES)
const locale = ref('en')
const items = ref<VisualShowcaseAdministrationItemFormState[]>(
  Array.from(
    { length: HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT },
    (_, index) => createVisualShowcaseHomeHeroAdministrationItemFormState(index),
  ),
)
const loading = ref(false)
const saving = ref(false)
const savingIndex = ref<number | null>(null)
const uploadingIndex = ref<number | null>(null)

const canEdit = computed(() => authStore.hasPermission('content:edit'))
const busy = computed(() => saving.value || savingIndex.value !== null)

const updateItem = (index: number, item: VisualShowcaseAdministrationItemFormState): void => {
  items.value = items.value.map((current, itemIndex) => itemIndex === index ? item : current)
}

const loadItems = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await visualShowcaseApi.getItems(SHOWCASE_KEY, locale.value)
    const loadedItems = visualShowcaseHomeHeroAdministrationRowsFromApiItems(response.items)
    items.value = loadedItems
  } catch (error) {
    console.error('Failed to load visual showcase:', error)
    toast.error('首页视觉目录加载失败')
    const emptyItems = visualShowcaseHomeHeroAdministrationRowsFromApiItems([])
    items.value = emptyItems
  } finally {
    loading.value = false
  }
}

const uploadImage = async ({ index, file }: VisualShowcaseAdministrationUploadRequest): Promise<void> => {
  if (!canEdit.value || busy.value || uploadingIndex.value !== null) return

  uploadingIndex.value = index
  try {
    const upload = await visualShowcaseApi.uploadImage(SHOWCASE_KEY, locale.value, file)
    const current = items.value[index]
    if (current) updateItem(index, applyVisualShowcaseUploadToFormState(current, upload))
    toast.success(`第 ${index + 1} 张图片已上传，保存当前卡片后生效`)
  } catch (error) {
    console.error('Failed to upload visual showcase image:', error)
    toast.error('图片上传失败，请确认图片为 600×600 px')
  } finally {
    uploadingIndex.value = null
  }
}

const save = async (): Promise<void> => {
  if (!canEdit.value || busy.value) return

  const validationMessage = visualShowcaseHomeHeroAdministrationValidationMessage(items.value)
  if (validationMessage) {
    toast.error(validationMessage)
    return
  }

  saving.value = true
  try {
    const response = await visualShowcaseApi.replaceItems(
      SHOWCASE_KEY,
      locale.value,
      items.value
        .filter((item) => item.image_url.trim() || item.storage_key.trim())
        .map((item) => visualShowcaseHomeHeroAdministrationSavePayloadFromFormRow(item, item.desktop_order - 1)),
    )
    const savedItems = visualShowcaseHomeHeroAdministrationRowsFromApiItems(response.items)
    items.value = savedItems
    toast.success('首页视觉目录已保存，旧图片已按引用关系清理')
  } catch (error) {
    console.error('Failed to save visual showcase:', error)
    toast.error('首页视觉目录保存失败')
  } finally {
    saving.value = false
  }
}

const saveSingleVisualShowcaseCardConfiguration = async (index: number): Promise<void> => {
  if (!canEdit.value || saving.value || savingIndex.value !== null) return

  const item = items.value[index]
  if (!item) return

  if (!item.image_url.trim() || !item.storage_key.trim()) {
    toast.error(`第 ${index + 1} 张图片请先上传图片`)
    return
  }

  const validationMessage = visualShowcaseHomeHeroAdministrationValidationMessage([item])
  if (validationMessage) {
    toast.error(validationMessage)
    return
  }

  savingIndex.value = index
  try {
    const response = await visualShowcaseApi.saveSingleVisualShowcaseItem(
      SHOWCASE_KEY,
      locale.value,
      index + 1,
      visualShowcaseHomeHeroAdministrationSavePayloadFromFormRow(item, index),
    )
    const savedItems = visualShowcaseHomeHeroAdministrationRowsFromApiItems(response.items)
    items.value = items.value.map((currentItem, itemIndex) => (
      itemIndex === index ? savedItems[itemIndex] || currentItem : currentItem
    ))
    toast.success(`第 ${index + 1} 张配置已保存`)
  } catch (error) {
    console.error('Failed to save visual showcase item:', error)
    toast.error(`第 ${index + 1} 张图片保存失败`)
  } finally {
    savingIndex.value = null
  }
}

watch(locale, () => {
  if (!loading.value && !busy.value) void loadItems()
})

onMounted(() => {
  void loadItems()
})
</script>

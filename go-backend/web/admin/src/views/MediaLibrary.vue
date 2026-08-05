<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader class="shrink-0" title="媒体库" description="后台通用媒体资源仓库">
      <template #actions>
        <Button v-if="canCreate" size="sm" :disabled="uploading" @click="triggerUpload">
          <UploadCloud v-if="!uploading" class="size-3.5" />
          <LoaderCircle v-else class="size-3.5 animate-spin" />
          上传资源
        </Button>
        <input
          ref="fileInput"
          class="hidden"
          type="file"
          accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime,video/webm"
          @change="handleUpload"
        />
      </template>
    </AdminPageHeader>

    <MediaLibraryFilters
      :search="filters.search"
      :media-type="filters.mediaType"
      :status="filters.status"
      :total="pagination.total"
      :loading="loading"
      @update:search="filters.search = $event"
      @update:media-type="filters.mediaType = $event"
      @update:status="filters.status = $event"
      @apply="applyFilters"
      @reset="resetFilters"
      @refresh="fetchAssets"
    />

    <Card class="min-h-0 flex-1 overflow-hidden">
      <CardContent class="min-h-0 flex-1 overflow-y-auto p-4">
        <div v-if="loading" class="flex min-h-64 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载媒体资源
        </div>
        <div v-else-if="assets.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 text-muted-foreground">
          <ImageOff class="size-9 opacity-50" />
          <span class="text-xs font-bold">暂无媒体资源</span>
        </div>
        <div v-else class="grid auto-rows-fr grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4 2xl:grid-cols-5">
          <MediaAssetCard
            v-for="asset in assets"
            :key="asset.id"
            :asset="asset"
            :can-edit="canEdit"
            :can-delete="canDelete"
            @preview="previewAsset"
            @copy-url="copyAssetURL"
            @export-evidence="downloadCopyrightEvidence"
            @edit="editAsset"
            @delete="requestDelete"
          />
        </div>
      </CardContent>
      <CardContent class="border-t py-3">
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 40, 80, 100]"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </CardContent>
    </Card>
  </div>

  <MediaAssetEditorDialog
    v-model:open="editorOpen"
    v-model:alt="editor.alt"
    v-model:caption="editor.caption"
    v-model:status="editor.status"
    v-model:visibility="editor.visibility"
    :asset="editingAsset"
    :saving="saving"
    @save="saveEditor"
  />

  <MediaAssetPreviewDialog v-model:open="previewOpen" :asset="preview" />

  <MediaDeleteDialog
    :open="deleteDialog.open"
    v-model:confirmation="deleteDialog.confirmation"
    :asset="deleteDialog.asset"
    :references="deleteDialog.references"
    :total="deleteDialog.total"
    :loading="deleteDialog.loading"
    :deleting="deleteDialog.deleting"
    @update:open="setDeleteDialogOpen"
    @confirm="deleteAsset"
  />
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ImageOff, LoaderCircle, UploadCloud } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import MediaAssetCard from '@/components/admin/media/MediaAssetCard.vue'
import MediaAssetEditorDialog from '@/components/admin/media/MediaAssetEditorDialog.vue'
import MediaAssetPreviewDialog from '@/components/admin/media/MediaAssetPreviewDialog.vue'
import MediaDeleteDialog from '@/components/admin/media/MediaDeleteDialog.vue'
import MediaLibraryFilters from '@/components/admin/media/MediaLibraryFilters.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { useMediaDeletion } from '@/composables/media/useMediaDeletion'
import { useMediaLibraryAssets } from '@/composables/media/useMediaLibraryAssets'
import { assetAccessURL } from '@/lib/mediaPresentation'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const fileInput = ref(null)
const preview = ref(null)
const previewOpen = ref(false)
const editorOpen = ref(false)
const editingAsset = ref(null)
const editor = reactive({ alt: '', caption: '', status: 'active', visibility: 'public' })

const {
  assets,
  filters,
  pagination,
  loading,
  uploading,
  saving,
  fetchAssets,
  applyFilters,
  resetFilters,
  updatePage,
  updatePageSize,
  uploadFile,
  saveAsset,
  copyURL,
  downloadCopyrightEvidence,
} = useMediaLibraryAssets()
const { deleteDialog, requestDelete, deleteAsset, setDeleteDialogOpen } = useMediaDeletion(fetchAssets)

const canCreate = computed(() => authStore.hasPermission('media:create'))
const canEdit = computed(() => authStore.hasPermission('media:edit'))
const canDelete = computed(() => authStore.hasPermission('media:delete'))

const triggerUpload = () => fileInput.value?.click()

const handleUpload = async (event) => {
  const file = event.target.files?.[0]
  event.target.value = ''
  await uploadFile(file)
}

const previewAsset = (asset) => {
  preview.value = asset
  previewOpen.value = true
}

const editAsset = (asset) => {
  editingAsset.value = asset
  Object.assign(editor, {
    alt: asset.alt || '',
    caption: asset.caption || '',
    status: asset.status || 'active',
    visibility: asset.visibility || 'public',
  })
  editorOpen.value = true
}

const saveEditor = async () => {
  const saved = await saveAsset(editingAsset.value, editor)
  if (saved) {
    editorOpen.value = false
  }
}

const copyAssetURL = (asset) => copyURL(asset, assetAccessURL(asset))

onMounted(fetchAssets)
</script>

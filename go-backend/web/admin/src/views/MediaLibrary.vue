<template>
  <div class="flex min-h-full flex-col gap-4 lg:h-full lg:min-h-0 lg:overflow-hidden">
    <AdminPageHeader class="shrink-0" title="媒体库" description="按业务归属管理素材，使用位置单独查看">
      <template #actions>
        <Button v-if="canCreate" size="sm" :disabled="uploading" @click="triggerUpload">
          <UploadCloud v-if="!uploading" class="size-3.5" />
          <LoaderCircle v-else class="size-3.5 animate-spin" />
          {{ uploadButtonLabel }}
        </Button>
        <input
          ref="fileInput"
          class="hidden"
          type="file"
          :accept="`${uploadSpecAccept('media_library_image')},video/mp4,video/quicktime,video/webm`"
          @change="handleUpload"
        />
      </template>
    </AdminPageHeader>
    <UploadSpecHint v-if="canCreate" code="media_library_image" compact class="shrink-0" />

    <MediaLibraryBusinessPartitionNavigation :selected-partition="filters.libraryPartition" :counts="partitionCounts" @select="selectAndRememberMediaLibraryPartition" />
    <div v-if="filters.libraryPartition === 'unclassified' || (filters.libraryPartition === 'operational' && partitionCounts.unclassified)" class="flex shrink-0 flex-wrap items-center justify-between gap-2 rounded-lg border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
      <span>历史来源未确认的文件保留在待整理区，已有内容中的使用保持有效。</span>
      <Button v-if="filters.libraryPartition !== 'unclassified'" size="xs" variant="outline" @click="selectAndRememberMediaLibraryPartition('unclassified')">查看待整理文件</Button>
    </div>

    <MediaLibraryFilters
      :search="filters.search"
      :media-type="filters.mediaType"
      :status="filters.status"
      :total="pagination.total"
      :loading="loading"
      :library-partition="filters.libraryPartition"
      :upload-origin="filters.uploadOrigin"
      :customer-service-media-purpose="filters.customerServiceMediaPurpose"
      :can-view-conversation-attachments="canViewConversationAttachments"
      @update:search="filters.search = $event"
      @update:media-type="filters.mediaType = $event"
      @update:status="filters.status = $event"
      @update:upload-origin="filters.uploadOrigin = $event"
      @update:customer-service-media-purpose="setCustomerServiceMediaPurposeFilter"
      @apply="applyFilters"
      @reset="resetFilters"
      @refresh="fetchAssets"
    />

    <Card class="min-h-0 shrink-0 gap-0 overflow-hidden py-0 lg:flex-1 lg:shrink">
      <CardContent class="min-h-0 p-4 lg:flex-1 lg:overflow-y-auto">
        <div v-if="loading" class="flex min-h-64 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载媒体资源
        </div>
        <div v-else-if="loadError" role="alert" class="flex min-h-64 items-center justify-center text-xs text-destructive">{{ loadError }}</div>
        <div v-else-if="assets.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 text-muted-foreground">
          <ImageOff class="size-9 opacity-50" />
          <span class="text-xs font-bold">当前分区暂无符合筛选的资源</span>
        </div>
        <div v-else class="grid auto-rows-min grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4 2xl:grid-cols-5">
          <MediaAssetCard
            v-for="asset in assets"
            :key="asset.id"
            :asset="asset"
            :can-edit="canEdit"
            :can-delete="canDelete"
            :can-configure="canConfigure"
            @preview="previewAsset"
            @copy-url="copyAssetURL"
            @export-evidence="downloadCopyrightEvidence"
            @edit="editAsset"
            @delete="requestDelete"
            @references="showAssetBusinessReferences"
            @review-ownership="manualOwnershipReview.openMediaAssetManualOwnershipReview"
          />
        </div>
      </CardContent>
      <CardContent class="border-t py-3">
        <AdminPagination
          compact
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
  <MediaAssetBusinessReferenceDetailsDialog v-model:open="referencesOpen" :asset="referencedAsset" />
  <MediaAssetManualLibraryOwnershipReviewAndAuditedClassificationDialog :ownership-review="manualOwnershipReview" />

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

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ImageOff, LoaderCircle, UploadCloud } from '@lucide/vue'
import type { MediaAsset, OperationalMediaUploadContext } from '@/api/media'
import MediaLibraryBusinessPartitionNavigation from '@/components/admin/media/MediaLibraryBusinessPartitionNavigation.vue'
import MediaAssetBusinessReferenceDetailsDialog from '@/components/admin/media/MediaAssetBusinessReferenceDetailsDialog.vue'
import MediaAssetManualLibraryOwnershipReviewAndAuditedClassificationDialog from '@/components/admin/media/MediaAssetManualLibraryOwnershipReviewAndAuditedClassificationDialog.vue'
import { useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave } from '@/composables/media/useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave'
import { normalizeMediaLibraryPartitionView, type MediaLibraryPartitionView } from '@/modules/media/mediaLibraryBusinessPartitionUploadOriginAndCustomerServicePurposePresentation'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import MediaAssetCard from '@/components/admin/media/MediaAssetCard.vue'
import MediaAssetEditorDialog from '@/components/admin/media/MediaAssetEditorDialog.vue'
import MediaAssetPreviewDialog from '@/components/admin/media/MediaAssetPreviewDialog.vue'
import MediaDeleteDialog from '@/components/admin/media/MediaDeleteDialog.vue'
import MediaLibraryFilters from '@/components/admin/media/MediaLibraryFilters.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { useMediaDeletion } from '@/composables/media/useMediaDeletion'
import { useMediaLibraryAssets } from '@/composables/media/useMediaLibraryAssets'
import { assetAccessURL } from '@/lib/mediaPresentation'
import { uploadSpecAccept } from '@/lib/uploadSpecs'
import { useAuthStore } from '@/stores/auth'

interface MediaEditorState extends Record<string, unknown> {
  alt: string
  caption: string
  status: string
  visibility: string
}

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const referencedAsset = ref<MediaAsset | null>(null)
const referencesOpen = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const preview = ref<MediaAsset | null>(null)
const previewOpen = ref(false)
const editorOpen = ref(false)
const editingAsset = ref<MediaAsset | null>(null)
const editor = reactive<MediaEditorState>({ alt: '', caption: '', status: 'active', visibility: 'public' })

const {
  assets,
  loadError,
  partitionCounts,
  canViewConversationAttachments,
  filters,
  pagination,
  loading,
  uploading,
  saving,
  fetchAssets,
  applyFilters,
  selectLibraryPartition,
  resetFilters,
  updatePage,
  updatePageSize,
  uploadFile,
  saveAsset,
  copyURL,
  downloadCopyrightEvidence,
} = useMediaLibraryAssets(route.query.library_partition)
const { deleteDialog, requestDelete, deleteAsset, setDeleteDialogOpen } = useMediaDeletion(fetchAssets)
const manualOwnershipReview = useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave(fetchAssets)

const uploadContext = computed<OperationalMediaUploadContext>(() => {
  if (filters.libraryPartition === 'catalog') return 'catalog'
  if (filters.libraryPartition === 'content') return 'content'
  if (filters.libraryPartition === 'customer_service') return 'customer-service-reusable'
  return 'shared'
})
const canCreate = computed(() => {
  if (!authStore.hasPermission('media:create') || filters.libraryPartition === 'customer_submissions' || filters.libraryPartition === 'unclassified') return false
  if (uploadContext.value === 'catalog') return authStore.hasPermission('product:create') || authStore.hasPermission('product:edit')
  if (uploadContext.value === 'content') return authStore.hasPermission('content:create') || authStore.hasPermission('content:edit')
  if (uploadContext.value === 'customer-service-reusable') return authStore.hasPermission('ticket:edit')
  return true
})
const uploadButtonLabel = computed(() => ({ catalog: '上传商品素材', content: '上传页面素材', shared: '上传公共素材', 'customer-service-reusable': '上传常用素材', 'website-profile': '上传网站资料' })[uploadContext.value])
const canEdit = computed(() => authStore.hasPermission('media:edit'))
const canDelete = computed(() => authStore.hasPermission('media:delete'))
const canConfigure = computed(() => authStore.hasPermission('media:configure'))

const triggerUpload = (): void => fileInput.value?.click()

const handleUpload = async (event: Event): Promise<void> => {
  const input = event.target instanceof HTMLInputElement ? event.target : null
  const file = input?.files?.[0] || null
  if (input) input.value = ''
  const targetContext = uploadContext.value
  if (!canCreate.value) return
  await uploadFile(file, targetContext)
}

const previewAsset = (asset: MediaAsset): void => {
  if (asset.requires_conversation_access) return
  preview.value = asset
  previewOpen.value = true
}

const editAsset = (asset: MediaAsset): void => {
  editingAsset.value = asset
  Object.assign(editor, {
    alt: asset.alt || '',
    caption: asset.caption || '',
    status: asset.status || 'active',
    visibility: asset.visibility || 'public',
  })
  editorOpen.value = true
}

const saveEditor = async (): Promise<void> => {
  const saved = await saveAsset(editingAsset.value, editor)
  if (saved) {
    editorOpen.value = false
  }
}

const copyAssetURL = (asset: MediaAsset): Promise<void> => copyURL(asset, assetAccessURL(asset))

const selectAndRememberMediaLibraryPartition = (value: MediaLibraryPartitionView): void => {
  selectLibraryPartition(value)
  void router.replace({ query: { ...route.query, library_partition: value } })
}
watch(() => route.query.library_partition, (value) => {
  const partition = normalizeMediaLibraryPartitionView(value)
  if (partition !== filters.libraryPartition) selectLibraryPartition(partition)
})
const setCustomerServiceMediaPurposeFilter = (value: string): void => {
  if (value === 'all' || value === 'conversation_attachment' || value === 'reusable_reply') filters.customerServiceMediaPurpose = value
}
const showAssetBusinessReferences = (asset: MediaAsset): void => { referencedAsset.value = asset; referencesOpen.value = true }

onMounted(() => {
  void fetchAssets()
})
</script>

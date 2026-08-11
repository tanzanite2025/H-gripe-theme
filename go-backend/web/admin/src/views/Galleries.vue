<template>
  <div class="space-y-4">
    <AdminPageHeader title="品牌图库" description="管理前台展示的品牌图片、封面和图库素材">
      <template #actions>
        <Button v-if="hasPermission('gallery:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          创建图库
        </Button>
      </template>
    </AdminPageHeader>

    <GalleryTablePanel
      :loading="loading"
      :galleries="galleries"
      :pagination="pagination"
      :can-edit="hasPermission('gallery:edit')"
      :can-delete="hasPermission('gallery:delete')"
      :gallery-title="galleryTitle"
      :gallery-cover="galleryCover"
      :gallery-image-count="galleryImageCount"
      :format-date="formatDate"
      @preview="showImagePreview"
      @view-images="viewImages"
      @edit="showEditDialog"
      @delete="requestDeleteGallery"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <GalleryEditorDialog
      v-model:open="dialogVisible"
      :mode="dialogMode"
      :form="galleryForm"
      :errors="galleryErrors"
      :image-errors="galleryImageErrors"
      :submitting="submitting"
      @submit="submitGalleryForm"
      @clear-error="clearGalleryError"
      @open-product-picker="productPickerVisible = true"
      @remove-product="removeGalleryProduct"
      @add-image="addGalleryImage"
      @pick-image="openGalleryImagePicker"
      @clear-image-error="clearGalleryImageError"
      @remove-image="removeGalleryImage"
    />
    <GalleryProductPickerDialog
      v-model:open="productPickerVisible"
      :selected-product-ids="selectedGalleryProductIds"
      @select="addGalleryProduct"
      @remove="removeGalleryProduct"
    />

    <GalleryImagesDialog
      v-model:open="imagesDialogVisible"
      :current-gallery="currentGallery"
      :images="images"
      :selected-images="selectedImages"
      :image-selection-state="imageSelectionState"
      :loading="imagesLoading"
      :can-create="hasPermission('gallery:create')"
      :can-edit="hasPermission('gallery:edit')"
      :can-delete="hasPermission('gallery:delete')"
      :gallery-title="galleryTitle"
      @create="showAddImageDialog"
      @batch-delete="requestBatchDeleteImages"
      @toggle-all="toggleAllImages"
      @toggle-image="toggleImage"
      @preview="showImagePreview"
      @edit="showEditImageDialog"
      @delete="requestDeleteImage"
    />

    <GalleryImageEditorDialog
      v-model:open="imageDialogVisible"
      :mode="imageDialogMode"
      :form="imageForm"
      :errors="imageErrors"
      :submitting="imageSubmitting"
      @submit="submitImageForm"
      @clear-error="clearImageError"
      @pick-media="openStandaloneMediaPicker"
    />
    <MediaAssetPickerDialog
      v-model:open="mediaPickerVisible"
      :selected-urls="selectedMediaUrls"
      @select="selectGalleryImageAsset"
    />

    <GalleryPreviewDialog
      v-model:open="previewDialogVisible"
      :image="previewImage"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      destructive
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Plus } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import GalleryEditorDialog from '@/components/admin/gallery/GalleryEditorDialog.vue'
import GalleryImageEditorDialog from '@/components/admin/gallery/GalleryImageEditorDialog.vue'
import GalleryImagesDialog from '@/components/admin/gallery/GalleryImagesDialog.vue'
import GalleryPreviewDialog from '@/components/admin/gallery/GalleryPreviewDialog.vue'
import GalleryProductPickerDialog from '@/components/admin/gallery/GalleryProductPickerDialog.vue'
import GalleryTablePanel from '@/components/admin/gallery/GalleryTablePanel.vue'
import MediaAssetPickerDialog from '@/components/admin/media/MediaAssetPickerDialog.vue'
import type {
  GalleryConfirmation,
  GalleryDetailResponse,
  GalleryDialogMode,
  GalleryForm,
  GalleryFormErrors,
  GalleryId,
  GalleryImage,
  GalleryImageForm,
  GalleryImagePayload,
  GalleryImagesResponse,
  GalleryListResponse,
  GalleryPagination,
  GalleryPayload,
  GalleryPreviewImage,
  GalleryProductLink,
  GalleryRecord,
  GallerySelectionState
} from '@/components/admin/gallery/galleryTypes'
import type { MediaAsset } from '@/api/media'
import type { ProductRecord } from '@/components/admin/product/productEditorTypes'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const loading = ref(false)
const galleries = ref<GalleryRecord[]>([])
const dialogVisible = ref(false)
const dialogMode = ref<GalleryDialogMode>('create')
const submitting = ref(false)
const galleryErrors = reactive<GalleryFormErrors>({})
const galleryImageErrors = ref<GalleryFormErrors[]>([])
const productPickerVisible = ref(false)

const imagesDialogVisible = ref(false)
const imagesLoading = ref(false)
const images = ref<GalleryImage[]>([])
const currentGallery = ref<GalleryRecord | null>(null)
const selectedImages = ref<GalleryImage[]>([])

const imageDialogVisible = ref(false)
const imageDialogMode = ref<GalleryDialogMode>('create')
const imageSubmitting = ref(false)
const imageErrors = reactive<GalleryFormErrors>({})
const mediaPickerVisible = ref(false)
const mediaPickerTarget = ref<{ kind: 'gallery', index: number } | { kind: 'standalone' }>({ kind: 'standalone' })
const previewDialogVisible = ref(false)
const previewImage = reactive<GalleryPreviewImage>({ url: '', title: '' })
const removedGalleryImageIds = ref<number[]>([])

const pagination = reactive<GalleryPagination>({ page: 1, pageSize: 20, total: 0 })
const galleryForm = reactive<GalleryForm>({
  id: null,
  title: '',
  slug: '',
  description: '',
  product_links: [],
  images: []
})
const imageForm = reactive<GalleryImageForm>({ id: null, media_asset_id: null, url: '', thumbnail: '', title: '', description: '', tags: '', order: 0 })
const confirmation = reactive<GalleryConfirmation>({
  open: false,
  type: '',
  target: null,
  title: '',
  description: '',
  confirmLabel: '删除'
})

const imageSelectionState = computed<GallerySelectionState>(() => {
  if (images.value.length === 0 || selectedImages.value.length === 0) return false
  return selectedImages.value.length === images.value.length ? true : 'indeterminate'
})
const selectedGalleryProductIds = computed<GalleryId[]>(() => galleryForm.product_links.map((link) => link.product_id))
const selectedMediaUrls = computed<string[]>(() => {
  if (mediaPickerTarget.value.kind === 'gallery') {
    const image = galleryForm.images[mediaPickerTarget.value.index]
    return image?.url ? [image.url] : []
  }
  return imageForm.url ? [imageForm.url] : []
})

const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
const galleryTitle = (gallery?: GalleryRecord | null): string => gallery?.name || gallery?.title || '-'
const galleryCover = (gallery?: GalleryRecord | null): string => gallery?.cover_image || gallery?.images?.[0]?.thumbnail || gallery?.images?.[0]?.url || ''
const galleryImageCount = (gallery?: GalleryRecord | null): number | string => gallery?.image_count ?? gallery?.images_count ?? '-'
const formatDate = (dateString?: string | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const clearErrors = (errors: GalleryFormErrors): void => Object.keys(errors).forEach((key) => delete errors[key])
const clearGalleryError = (field: string): void => { delete galleryErrors[field] }
const clearImageError = (field: string): void => { delete imageErrors[field] }
const clearGalleryImageError = (index: number, field: string): void => {
  if (galleryImageErrors.value[index]) delete galleryImageErrors.value[index][field]
}
const toPositiveNumber = (value: GalleryId | null | undefined): number | null => {
  const id = Number(value)
  return Number.isFinite(id) && id > 0 ? id : null
}
function createEmptyGalleryImageForm(order = 0): GalleryImageForm {
  return {
    id: null,
    media_asset_id: null,
    url: '',
    thumbnail: '',
    title: '',
    description: '',
    tags: '',
    order
  }
}
const resetGalleryForm = (): void => {
  Object.assign(galleryForm, {
    id: null,
    title: '',
    slug: '',
    description: '',
    product_links: [],
    images: [createEmptyGalleryImageForm()]
  })
  galleryImageErrors.value = [{}]
  removedGalleryImageIds.value = []
  clearErrors(galleryErrors)
}
const resetImageForm = (): void => {
  Object.assign(imageForm, { id: null, media_asset_id: null, url: '', thumbnail: '', title: '', description: '', tags: '', order: 0 })
  clearErrors(imageErrors)
}
const buildGalleryImagePayload = (form: GalleryImageForm): GalleryImagePayload => ({
  media_asset_id: toPositiveNumber(form.media_asset_id),
  title: form.title.trim(),
  description: form.description.trim(),
  tags: form.tags.trim(),
  order: Math.max(0, Number(form.order || 0))
})
const buildGalleryPayload = (): GalleryPayload => ({
  title: galleryForm.title.trim(),
  slug: galleryForm.slug.trim(),
  description: galleryForm.description.trim(),
  product_ids: Array.from(
    new Set(
      galleryForm.product_links
        .map((link) => link.product_id)
        .map((id) => toPositiveNumber(id))
        .filter((id): id is number => id !== null)
    )
  ),
  ...(dialogMode.value === 'create'
    ? { images: galleryForm.images.map((image) => buildGalleryImagePayload(image)) }
    : {})
})
const buildImagePayload = (): GalleryImagePayload => ({
  ...buildGalleryImagePayload(imageForm)
})
const validateGallery = (payload: GalleryPayload): boolean => {
  clearErrors(galleryErrors)
  if (!payload.title) galleryErrors.title = '请输入图库标题'
  if (!payload.slug) galleryErrors.slug = '请输入图库 Slug'
  if (Object.keys(galleryErrors).length) {
    toast.error('请检查图库表单中的必填项')
    return false
  }
  return true
}
const validateImage = (payload: GalleryImagePayload): boolean => {
  clearErrors(imageErrors)
  if (!payload.media_asset_id) imageErrors.media_asset_id = '请从媒体仓库选择图片'
  if (!payload.title) imageErrors.title = '请输入图片标题'
  if (Object.keys(imageErrors).length) {
    toast.error('请检查图片表单中的必填项')
    return false
  }
  return true
}
const validateGalleryImages = (): boolean => {
  galleryImageErrors.value = galleryForm.images.map(() => ({}))
  if (dialogMode.value === 'create' && galleryForm.images.length === 0) {
    toast.error('请至少添加一张图库图片')
    return false
  }

  let valid = true
  galleryForm.images.forEach((image, index) => {
    const errors: GalleryFormErrors = {}
    const needsMediaAsset = dialogMode.value === 'create' || image.id === null
    if (needsMediaAsset && !toPositiveNumber(image.media_asset_id)) {
      errors.media_asset_id = '请从媒体仓库选择图片'
    }
    if (!image.title.trim()) {
      errors.title = '请输入图片标题'
    }
    if (Object.keys(errors).length) {
      galleryImageErrors.value[index] = errors
      valid = false
    }
  })

  if (!valid) toast.error('请检查图库图片配置')
  return valid
}

const normalizeProductLinks = (links?: GalleryProductLink[] | null): GalleryProductLink[] => {
  if (!Array.isArray(links)) return []
  return links
    .map((link) => ({
      product_id: link.product_id,
      name: link.name || '',
      slug: link.slug || '',
      locale: link.locale || ''
    }))
    .filter((link) => link.product_id !== null && link.product_id !== undefined)
}

const addGalleryProduct = (product: ProductRecord): void => {
  const productId = toPositiveNumber(product.id)
  if (!productId) return
  if (galleryForm.product_links.some((link) => String(link.product_id) === String(productId))) return
  galleryForm.product_links = [
    ...galleryForm.product_links,
    {
      product_id: productId,
      name: product.name || '',
      slug: product.slug || '',
      locale: product.locale || ''
    }
  ]
}

const removeGalleryProduct = (productId: GalleryId): void => {
  galleryForm.product_links = galleryForm.product_links.filter((link) => String(link.product_id) !== String(productId))
}

const addGalleryImage = (): void => {
  galleryForm.images = [...galleryForm.images, createEmptyGalleryImageForm(galleryForm.images.length)]
  galleryImageErrors.value = [...galleryImageErrors.value, {}]
}

const removeGalleryImage = (index: number): void => {
  const image = galleryForm.images[index]
  const imageId = toPositiveNumber(image?.id)
  if (dialogMode.value === 'edit' && imageId) {
    removedGalleryImageIds.value = [...removedGalleryImageIds.value, imageId]
  }
  galleryForm.images = galleryForm.images.filter((_, imageIndex) => imageIndex !== index)
  galleryImageErrors.value = galleryImageErrors.value.filter((_, imageIndex) => imageIndex !== index)
}

const openGalleryImagePicker = (index: number): void => {
  mediaPickerTarget.value = { kind: 'gallery', index }
  mediaPickerVisible.value = true
}

const openStandaloneMediaPicker = (): void => {
  mediaPickerTarget.value = { kind: 'standalone' }
  mediaPickerVisible.value = true
}

const selectGalleryImageAsset = (selection: { url: string, asset: MediaAsset }): void => {
  const asset = selection.asset
  const mediaAssetId = toPositiveNumber(asset.id)
  if (!mediaAssetId) return
  if (mediaPickerTarget.value.kind === 'gallery') {
    const target = galleryForm.images[mediaPickerTarget.value.index]
    if (!target) return
    Object.assign(target, {
      media_asset_id: mediaAssetId,
      url: selection.url,
      thumbnail: selection.url,
      title: target.title || asset.alt || asset.original_filename || asset.filename || ''
    })
    clearGalleryImageError(mediaPickerTarget.value.index, 'media_asset_id')
  } else {
    Object.assign(imageForm, {
      media_asset_id: mediaAssetId,
      url: selection.url,
      thumbnail: selection.url,
      title: imageForm.title || asset.alt || asset.original_filename || asset.filename || ''
    })
    delete imageErrors.media_asset_id
  }
  mediaPickerVisible.value = false
}

const isGalleryRecord = (target: GalleryConfirmation['target']): target is GalleryRecord => (
  Boolean(target) && !Array.isArray(target)
)
const isGalleryImage = (target: GalleryConfirmation['target']): target is GalleryImage => (
  Boolean(target) && !Array.isArray(target)
)

const fetchGalleries = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await axios.get<GalleryListResponse>('/api/admin/galleries', {
      params: { page: pagination.page, page_size: pagination.pageSize }
    })
    galleries.value = response.data.galleries || []
    pagination.total = response.data.pagination?.total ?? response.data.total ?? 0
  } catch (error) {
    console.error('Failed to fetch galleries:', error)
  } finally {
    loading.value = false
  }
}
const updatePage = (page: number): void => {
  pagination.page = page
  void fetchGalleries()
}
const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchGalleries()
}

const showCreateDialog = (): void => {
  dialogMode.value = 'create'
  resetGalleryForm()
  dialogVisible.value = true
}
const showEditDialog = async (gallery: GalleryRecord): Promise<void> => {
  dialogMode.value = 'edit'
  try {
    const response = await axios.get<GalleryDetailResponse>(`/api/admin/galleries/${gallery.id}`)
    const detail = response.data.gallery || gallery
    const detailImages = Array.isArray(detail.images) ? detail.images : []
    Object.assign(galleryForm, {
      id: detail.id,
      title: galleryTitle(detail),
      slug: detail.slug || '',
      description: detail.description || '',
      product_links: normalizeProductLinks(detail.product_links),
      images: detailImages.map((image, index) => ({
        id: toPositiveNumber(image.id),
        media_asset_id: toPositiveNumber(image.media_asset_id),
        url: image.url || '',
        thumbnail: image.thumbnail || image.url || '',
        title: image.title || '',
        description: image.description || '',
        tags: image.tags || '',
        order: Number(image.order ?? image.sort_order ?? index)
      }))
    })
    clearErrors(galleryErrors)
    galleryImageErrors.value = galleryForm.images.map(() => ({}))
    removedGalleryImageIds.value = []
    dialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch gallery detail:', error)
  }
}
const syncGalleryImages = async (galleryId: GalleryId): Promise<void> => {
  const requests = [
    ...galleryForm.images.map((image) => {
      const payload = buildGalleryImagePayload(image)
      const imageId = toPositiveNumber(image.id)
      return imageId
        ? axios.put(`/api/admin/galleries/${galleryId}/images/${imageId}`, payload)
        : axios.post(`/api/admin/galleries/${galleryId}/images`, payload)
    }),
    ...removedGalleryImageIds.value.map((imageId) =>
      axios.delete(`/api/admin/galleries/${galleryId}/images/${imageId}`)
    )
  ]

  await Promise.all(requests)
}
const submitGalleryForm = async (): Promise<void> => {
  const payload = buildGalleryPayload()
  if (!validateGallery(payload)) return
  if (!validateGalleryImages()) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/galleries', payload)
      toast.success('图库创建成功')
    } else if (galleryForm.id !== null) {
      await axios.put(`/api/admin/galleries/${galleryForm.id}`, payload)
      await syncGalleryImages(galleryForm.id)
      toast.success('图库更新成功')
    }
    dialogVisible.value = false
    await fetchGalleries()
  } catch (error) {
    console.error('Failed to save gallery:', error)
  } finally {
    submitting.value = false
  }
}

const viewImages = async (gallery: GalleryRecord): Promise<void> => {
  currentGallery.value = gallery
  images.value = []
  selectedImages.value = []
  imagesDialogVisible.value = true
  await fetchImages(gallery.id)
}
const fetchImages = async (galleryId: GalleryId): Promise<void> => {
  imagesLoading.value = true
  try {
    const response = await axios.get<GalleryImagesResponse>(`/api/admin/galleries/${galleryId}/images`)
    images.value = response.data.images || []
    selectedImages.value = []
  } catch (error) {
    console.error('Failed to fetch gallery images:', error)
  } finally {
    imagesLoading.value = false
  }
}
const showAddImageDialog = (): void => {
  imageDialogMode.value = 'create'
  resetImageForm()
  imageDialogVisible.value = true
}
const showEditImageDialog = (image: GalleryImage): void => {
  imageDialogMode.value = 'edit'
  Object.assign(imageForm, {
    id: image.id,
    media_asset_id: image.media_asset_id || null,
    url: image.url || '',
    thumbnail: image.thumbnail || '',
    title: image.title || '',
    description: image.description || '',
    tags: image.tags || '',
    order: image.order ?? image.sort_order ?? 0
  })
  clearErrors(imageErrors)
  imageDialogVisible.value = true
}
const submitImageForm = async (): Promise<void> => {
  if (!currentGallery.value) return

  const payload = buildImagePayload()
  if (!validateImage(payload)) return
  imageSubmitting.value = true
  try {
    if (imageDialogMode.value === 'create') {
      await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images`, payload)
      toast.success('图片添加成功')
    } else if (imageForm.id !== null) {
      await axios.put(`/api/admin/galleries/${currentGallery.value.id}/images/${imageForm.id}`, payload)
      toast.success('图片更新成功')
    }
    imageDialogVisible.value = false
    await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
  } catch (error) {
    console.error('Failed to save gallery image:', error)
  } finally {
    imageSubmitting.value = false
  }
}

const showImagePreview = (url?: string | null, title?: string | null): void => {
  if (!url) return
  Object.assign(previewImage, { url, title: title || '图片预览' })
  previewDialogVisible.value = true
}
const isImageSelected = (imageId: GalleryId): boolean => selectedImages.value.some((image) => image.id === imageId)
const toggleAllImages = (checked: GallerySelectionState): void => { selectedImages.value = checked === true ? [...images.value] : [] }
const toggleImage = (image: GalleryImage, checked: GallerySelectionState): void => {
  if (checked === true && !isImageSelected(image.id)) selectedImages.value = [...selectedImages.value, image]
  else if (checked !== true) selectedImages.value = selectedImages.value.filter((selected) => selected.id !== image.id)
}
const setConfirmation = (values: Partial<GalleryConfirmation>): void => {
  Object.assign(confirmation, {
    open: true,
    type: '',
    target: null,
    confirmLabel: '删除',
    ...values
  })
}
const requestDeleteGallery = (gallery: GalleryRecord): void => setConfirmation({
  type: 'delete-gallery', target: gallery, title: '删除图库？',
  description: `图库“${galleryTitle(gallery)}”及其中图片将被永久删除，此操作不可恢复。`
})
const requestDeleteImage = (image: GalleryImage): void => setConfirmation({
  type: 'delete-image', target: image, title: '删除图片？',
  description: `图片“${image.title || image.id}”将被永久删除，此操作不可恢复。`
})
const requestBatchDeleteImages = (): void => setConfirmation({
  type: 'batch-delete-images', target: [...selectedImages.value], title: '批量删除图片？',
  description: `${selectedImages.value.length} 张图片将被永久删除，此操作不可恢复。`, confirmLabel: '批量删除'
})
const executeConfirmedAction = async (): Promise<void> => {
  const { type, target } = confirmation
  confirmation.open = false
  try {
    if (type === 'delete-gallery' && isGalleryRecord(target)) {
      await axios.delete(`/api/admin/galleries/${target.id}`)
      toast.success('图库已删除')
      await fetchGalleries()
    } else if (type === 'delete-image' && isGalleryImage(target) && currentGallery.value) {
      await axios.delete(`/api/admin/galleries/${currentGallery.value.id}/images/${target.id}`)
      toast.success('图片已删除')
      await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
    } else if (type === 'batch-delete-images' && Array.isArray(target) && currentGallery.value) {
      await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images/batch-delete`, {
        image_ids: target.map((image) => image.id)
      })
      toast.success('图片已批量删除')
      await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
    }
  } catch (error) {
    console.error('Failed to delete gallery content:', error)
  }
}

onMounted(() => {
  void fetchGalleries()
})
</script>

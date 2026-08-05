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
      :submitting="submitting"
      @submit="submitGalleryForm"
      @clear-error="clearGalleryError"
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

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Plus } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import GalleryEditorDialog from '@/components/admin/gallery/GalleryEditorDialog.vue'
import GalleryImageEditorDialog from '@/components/admin/gallery/GalleryImageEditorDialog.vue'
import GalleryImagesDialog from '@/components/admin/gallery/GalleryImagesDialog.vue'
import GalleryPreviewDialog from '@/components/admin/gallery/GalleryPreviewDialog.vue'
import GalleryTablePanel from '@/components/admin/gallery/GalleryTablePanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const loading = ref(false)
const galleries = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const galleryErrors = reactive({})

const imagesDialogVisible = ref(false)
const imagesLoading = ref(false)
const images = ref([])
const currentGallery = ref(null)
const selectedImages = ref([])

const imageDialogVisible = ref(false)
const imageDialogMode = ref('create')
const imageSubmitting = ref(false)
const imageErrors = reactive({})
const previewDialogVisible = ref(false)
const previewImage = reactive({ url: '', title: '' })

const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const galleryForm = reactive({ id: null, title: '', slug: '', description: '' })
const imageForm = reactive({ id: null, url: '', thumbnail: '', title: '', description: '', tags: '', order: 0 })
const confirmation = reactive({
  open: false,
  type: '',
  target: null,
  title: '',
  description: '',
  confirmLabel: '删除'
})

const imageSelectionState = computed(() => {
  if (images.value.length === 0 || selectedImages.value.length === 0) return false
  return selectedImages.value.length === images.value.length ? true : 'indeterminate'
})

const hasPermission = (permission) => authStore.hasPermission(permission)
const galleryTitle = (gallery) => gallery?.name || gallery?.title || '-'
const galleryCover = (gallery) => gallery?.cover_image || gallery?.images?.[0]?.thumbnail || gallery?.images?.[0]?.url || ''
const galleryImageCount = (gallery) => gallery?.image_count ?? gallery?.images_count ?? '-'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const clearErrors = (errors) => Object.keys(errors).forEach((key) => delete errors[key])
const clearGalleryError = (field) => { delete galleryErrors[field] }
const clearImageError = (field) => { delete imageErrors[field] }
const resetGalleryForm = () => {
  Object.assign(galleryForm, { id: null, title: '', slug: '', description: '' })
  clearErrors(galleryErrors)
}
const resetImageForm = () => {
  Object.assign(imageForm, { id: null, url: '', thumbnail: '', title: '', description: '', tags: '', order: 0 })
  clearErrors(imageErrors)
}
const buildGalleryPayload = () => ({
  title: galleryForm.title.trim(),
  slug: galleryForm.slug.trim(),
  description: galleryForm.description.trim()
})
const buildImagePayload = () => ({
  url: imageForm.url.trim(),
  thumbnail: imageForm.thumbnail.trim(),
  title: imageForm.title.trim(),
  description: imageForm.description.trim(),
  tags: imageForm.tags.trim(),
  order: Math.max(0, Number(imageForm.order || 0))
})
const validateGallery = (payload) => {
  clearErrors(galleryErrors)
  if (!payload.title) galleryErrors.title = '请输入图库标题'
  if (!payload.slug) galleryErrors.slug = '请输入图库 Slug'
  if (Object.keys(galleryErrors).length) {
    toast.error('请检查图库表单中的必填项')
    return false
  }
  return true
}
const validateImage = (payload) => {
  clearErrors(imageErrors)
  if (!payload.url) imageErrors.url = '请输入图片 URL'
  if (!payload.title) imageErrors.title = '请输入图片标题'
  if (Object.keys(imageErrors).length) {
    toast.error('请检查图片表单中的必填项')
    return false
  }
  return true
}

const fetchGalleries = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/galleries', {
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
const updatePage = (page) => { pagination.page = page; fetchGalleries() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchGalleries() }

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetGalleryForm()
  dialogVisible.value = true
}
const showEditDialog = async (gallery) => {
  dialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/galleries/${gallery.id}`)
    const detail = response.data.gallery || gallery
    Object.assign(galleryForm, {
      id: detail.id,
      title: galleryTitle(detail),
      slug: detail.slug || '',
      description: detail.description || ''
    })
    clearErrors(galleryErrors)
    dialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch gallery detail:', error)
  }
}
const submitGalleryForm = async () => {
  const payload = buildGalleryPayload()
  if (!validateGallery(payload)) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/galleries', payload)
      toast.success('图库创建成功')
    } else {
      await axios.put(`/api/admin/galleries/${galleryForm.id}`, payload)
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

const viewImages = async (gallery) => {
  currentGallery.value = gallery
  images.value = []
  selectedImages.value = []
  imagesDialogVisible.value = true
  await fetchImages(gallery.id)
}
const fetchImages = async (galleryId) => {
  imagesLoading.value = true
  try {
    const response = await axios.get(`/api/admin/galleries/${galleryId}/images`)
    images.value = response.data.images || []
    selectedImages.value = []
  } catch (error) {
    console.error('Failed to fetch gallery images:', error)
  } finally {
    imagesLoading.value = false
  }
}
const showAddImageDialog = () => {
  imageDialogMode.value = 'create'
  resetImageForm()
  imageDialogVisible.value = true
}
const showEditImageDialog = (image) => {
  imageDialogMode.value = 'edit'
  Object.assign(imageForm, {
    id: image.id,
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
const submitImageForm = async () => {
  const payload = buildImagePayload()
  if (!validateImage(payload)) return
  imageSubmitting.value = true
  try {
    if (imageDialogMode.value === 'create') {
      await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images`, payload)
      toast.success('图片添加成功')
    } else {
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

const showImagePreview = (url, title) => {
  Object.assign(previewImage, { url, title: title || '图片预览' })
  previewDialogVisible.value = true
}
const isImageSelected = (imageId) => selectedImages.value.some((image) => image.id === imageId)
const toggleAllImages = (checked) => { selectedImages.value = checked === true ? [...images.value] : [] }
const toggleImage = (image, checked) => {
  if (checked === true && !isImageSelected(image.id)) selectedImages.value = [...selectedImages.value, image]
  else if (checked !== true) selectedImages.value = selectedImages.value.filter((selected) => selected.id !== image.id)
}
const setConfirmation = (values) => Object.assign(confirmation, {
  open: true,
  type: '',
  target: null,
  confirmLabel: '删除',
  ...values
})
const requestDeleteGallery = (gallery) => setConfirmation({
  type: 'delete-gallery', target: gallery, title: '删除图库？',
  description: `图库“${galleryTitle(gallery)}”及其中图片将被永久删除，此操作不可恢复。`
})
const requestDeleteImage = (image) => setConfirmation({
  type: 'delete-image', target: image, title: '删除图片？',
  description: `图片“${image.title || image.id}”将被永久删除，此操作不可恢复。`
})
const requestBatchDeleteImages = () => setConfirmation({
  type: 'batch-delete-images', target: [...selectedImages.value], title: '批量删除图片？',
  description: `${selectedImages.value.length} 张图片将被永久删除，此操作不可恢复。`, confirmLabel: '批量删除'
})
const executeConfirmedAction = async () => {
  const { type, target } = confirmation
  confirmation.open = false
  try {
    if (type === 'delete-gallery') {
      await axios.delete(`/api/admin/galleries/${target.id}`)
      toast.success('图库已删除')
      await fetchGalleries()
    } else if (type === 'delete-image') {
      await axios.delete(`/api/admin/galleries/${currentGallery.value.id}/images/${target.id}`)
      toast.success('图片已删除')
      await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
    } else if (type === 'batch-delete-images') {
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

onMounted(fetchGalleries)
</script>

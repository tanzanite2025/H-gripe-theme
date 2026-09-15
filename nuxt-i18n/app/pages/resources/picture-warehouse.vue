<template>
  <div class="picture-warehouse-page">
    <h1 class="sr-only">{{ t('resourcesPictureWarehouseShared.title') }}</h1>

    <PictureWarehouseRidersTab
      v-if="activeTab === 'riders'"
      :photos="userPhotos"
      :visible-photos="visibleUserPhotos"
      :loading="userLoading"
      :show-all="showAllUserPhotos"
      :has-more="hasMoreUserPhotos"
      @open="openLightbox('user', $event)"
      @toggle="toggleUserPhotos"
    />
    <PictureWarehouseBrandTab
      v-else
      :photos="brandPhotos"
      :visible-photos="visibleBrandPhotos"
      :loading="brandLoading"
      :show-all="showAllBrandPhotos"
      :has-more="hasMoreBrandPhotos"
      @open="openLightbox('brand', $event)"
      @toggle="toggleBrandPhotos"
    />

    <PictureWarehouseUploadPanel
      :upload-order-id="uploadOrderID"
      :upload-orders="uploadOrders"
      :upload-orders-loading="uploadOrdersLoading"
      :upload-orders-error="uploadOrdersError"
      :has-eligible-upload-order="hasEligibleUploadOrder"
      :selected-upload-order-eligible="selectedUploadOrderEligible"
      :upload-region="uploadRegion"
      :upload-location="uploadLocation"
      :upload-nickname="uploadNickname"
      :upload-notes="uploadNotes"
      :upload-files="uploadFiles"
      :upload-previews="uploadPreviews"
      :upload-accept="uploadSpecAccept('user_showcase_image')"
      :upload-hint="uploadSpecHint('user_showcase_image')"
      :uploading="uploading"
      :upload-error="uploadError"
      :upload-success="uploadSuccess"
      @update:upload-order-id="uploadOrderID = $event"
      @update:upload-region="uploadRegion = $event"
      @update:upload-location="uploadLocation = $event"
      @update:upload-nickname="uploadNickname = $event"
      @update:upload-notes="uploadNotes = $event"
      @file-change="onUploadFileChange"
      @remove-file="removeUploadFile"
      @submit="submitUpload"
    />

    <PictureWarehouseRiderLightbox
      :open="isLightboxOpen && activeKind === 'user'"
      :active-photo="activePhoto && activePhoto.kind === 'user' ? activePhoto : null"
      :current-image-url="currentImageUrl"
      :current-gallery-index="currentGalleryIndex"
      :comments="comments"
      :comments-loading="commentsLoading"
      :comments-error="commentsError"
      :comment-content="commentContent"
      :comment-location="commentLocation"
      :comment-submitting="commentSubmitting"
      :comment-success="commentSuccess"
      :comment-error="commentError"
      :share-copying="shareCopying"
      :share-message="shareMessage"
      :active-photo-product-links="activePhotoProductLinks"
      :product-link-path="productLinkPath"
      @close="closeLightbox"
      @previous="goPrev"
      @next="goNext"
      @select-image="currentGalleryIndex = $event"
      @copy-share-link="copyShareLink"
      @update:comment-content="commentContent = $event"
      @update:comment-location="commentLocation = $event"
      @submit-comment="submitComment"
    />

    <GlobalBrandGalleryLightbox
      :open="isLightboxOpen && activeKind === 'brand'"
      :gallery="activeBrandPhoto"
      :labels="brandLightboxLabels"
      @close="closeLightbox"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  definePageMeta,
  useHead,
  useI18n,
  useLocalePath,
  useRuntimeConfig,
} from '#imports'
import GlobalBrandGalleryLightbox from '~/components/global/gallery/GlobalBrandGalleryLightbox.vue'
import PictureWarehouseBrandTab from '~/components/picture-warehouse/PictureWarehouseBrandTab.vue'
import PictureWarehouseRiderLightbox from '~/components/picture-warehouse/PictureWarehouseRiderLightbox.vue'
import PictureWarehouseRidersTab from '~/components/picture-warehouse/PictureWarehouseRidersTab.vue'
import PictureWarehouseUploadPanel from '~/components/picture-warehouse/PictureWarehouseUploadPanel.vue'
import { useAuth } from '~/composables/useAuth'
import { useBrandGalleryPhotos } from '~/composables/useBrandGalleryPhotos'
import { usePageMessages } from '~/composables/usePageMessages'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import type { BrandGalleryPhoto, PictureWarehouseProductLink } from '~/types/brandGalleryPhotos'
import type { PhotoComment, PictureWarehousePhoto, RiderPhoto, UploadOrderOption, UploadPreview } from '~/types/pictureWarehouse'
import { pictureWarehouseTabs } from '~/utils/pageSubNavigation'
import { buildProductPath } from '~/utils/seo/urls'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'
import {
  uploadSpecAccept,
  uploadSpecHint,
  validateStorefrontUploadFiles,
} from '~/utils/uploadSpecs'
import {
  defaultGalleryLightboxLabels,
  type GalleryLightboxLabels,
} from '~/types/brandGalleryLightbox'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.pictureWarehouse',
  footerLabelFallback: 'Picture Warehouse',
})

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('resourcesPictureWarehouseShared')
const { activeTab } = usePageSubNavigationTab({
  tabs: pictureWarehouseTabs,
  basePath: '/resources/picture-warehouse',
  defaultValue: 'riders',
})

await loadPageMessages(locale.value)
watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('resourcesPictureWarehouseShared.metaTitle'),
}))

const auth = useAuth()
const mediaContext = createStorefrontMediaContext(useRuntimeConfig())
const {
  brandPhotos,
  brandLoading,
  fetchBrandPhotos,
  loadBrandGalleryDetails,
} = useBrandGalleryPhotos()

const userPhotos = ref<RiderPhoto[]>([])
const userLoading = ref(true)
const uploadRegion = ref('')
const uploadLocation = ref('')
const uploadNickname = ref('')
const uploadNotes = ref('')
const uploadOrders = ref<UploadOrderOption[]>([])
const uploadOrdersLoading = ref(false)
const uploadOrdersError = ref<string | null>(null)
const uploadOrderID = ref('')
const uploadFiles = ref<File[]>([])
const uploadPreviews = ref<UploadPreview[]>([])
const uploading = ref(false)
const uploadError = ref<string | null>(null)
const uploadSuccess = ref<string | null>(null)
const comments = ref<PhotoComment[]>([])
const commentsLoading = ref(false)
const commentsError = ref<string | null>(null)
const commentContent = ref('')
const commentLocation = ref('')
const commentSubmitting = ref(false)
const commentSuccess = ref<string | null>(null)
const commentError = ref<string | null>(null)
const shareCopying = ref(false)
const shareMessage = ref<string | null>(null)
const showAllUserPhotos = ref(false)
const showAllBrandPhotos = ref(false)
const activeKind = ref<'user' | 'brand' | null>(null)
const activeIndex = ref<number | null>(null)
const currentGalleryIndex = ref(0)

const selectedUploadOrderEligible = computed(() => uploadOrders.value.some(
  order => String(order.id) === uploadOrderID.value && order.eligible,
))
const hasEligibleUploadOrder = computed(() => uploadOrders.value.some(order => order.eligible))
const visibleUserPhotos = computed(() => showAllUserPhotos.value ? userPhotos.value : userPhotos.value.slice(0, 6))
const visibleBrandPhotos = computed(() => showAllBrandPhotos.value ? brandPhotos.value : brandPhotos.value.slice(0, 6))
const hasMoreUserPhotos = computed(() => userPhotos.value.length > 6)
const hasMoreBrandPhotos = computed(() => brandPhotos.value.length > 6)
const isLightboxOpen = computed(() => activeKind.value !== null && activeIndex.value !== null)
const activeList = computed<PictureWarehousePhoto[] | null>(() => {
  if (!activeKind.value) return null
  return activeKind.value === 'user' ? userPhotos.value : brandPhotos.value
})
const activePhoto = computed<PictureWarehousePhoto | null>(() => {
  if (!activeList.value || activeIndex.value === null) return null
  return activeList.value[activeIndex.value] || null
})
const activeBrandPhoto = computed<BrandGalleryPhoto | null>(() => (
  activeKind.value === 'brand' && activeIndex.value !== null
    ? brandPhotos.value[activeIndex.value] || null
    : null
))
const activePhotoProductLinks = computed<PictureWarehouseProductLink[]>(() => (
  (activePhoto.value?.productLinks || []).filter(product => Boolean(product.slug || product.name))
))
const currentImageUrl = computed(() => {
  const photo = activePhoto.value
  if (!photo?.galleryImages?.length) return ''
  return photo.galleryImages[currentGalleryIndex.value] || photo.galleryImages[0] || ''
})
const brandLightboxLabels = computed<GalleryLightboxLabels>(() => ({
  ...defaultGalleryLightboxLabels,
  close: t('resourcesPictureWarehouseShared.lightbox.close'),
  previousImage: t('resourcesPictureWarehouseShared.lightbox.previousPhoto'),
  nextImage: t('resourcesPictureWarehouseShared.lightbox.nextPhoto'),
  imageThumbnails: t('resourcesPictureWarehouseShared.lightbox.imageThumbnails'),
  loadingDetails: t('resourcesPictureWarehouseShared.lightbox.loadingDetails'),
  noImage: t('resourcesPictureWarehouseShared.lightbox.noImage'),
  relatedProducts: t('resourcesPictureWarehouseShared.lightbox.relatedProducts'),
  noRelatedProducts: t('resourcesPictureWarehouseShared.lightbox.noRelatedProducts'),
}))

const uploadFileKey = (file: File) => `${file.name}:${file.size}:${file.lastModified}`

const onUploadFileChange = async (event: Event) => {
  if (!selectedUploadOrderEligible.value) return
  const target = event.target as HTMLInputElement | null
  if (!target?.files?.length) return
  uploadError.value = null
  uploadSuccess.value = null
  const selectedFiles = Array.from(target.files)
  const validation = await validateStorefrontUploadFiles(
    [...uploadFiles.value, ...selectedFiles],
    'user_showcase_image',
  )
  if (!validation.ok) {
    uploadError.value = validation.error || t('resourcesPictureWarehouseShared.errors.invalidFiles')
    target.value = ''
    return
  }
  const existingKeys = new Set(uploadFiles.value.map(uploadFileKey))
  const newFiles = selectedFiles.filter(file => !existingKeys.has(uploadFileKey(file)))
  const availableSlots = Math.max(0, 10 - uploadFiles.value.length)
  if (newFiles.length > availableSlots) {
    uploadError.value = t('resourcesPictureWarehouseShared.errors.maximumFiles')
  }
  const filesToAdd = newFiles.slice(0, availableSlots)
  uploadFiles.value.push(...filesToAdd)
  uploadPreviews.value.push(...filesToAdd.map(file => ({
    key: uploadFileKey(file),
    file,
    url: URL.createObjectURL(file),
  })))
  target.value = ''
}

const removeUploadFile = (index: number) => {
  const preview = uploadPreviews.value[index]
  if (!preview) return
  URL.revokeObjectURL(preview.url)
  uploadPreviews.value.splice(index, 1)
  uploadFiles.value.splice(index, 1)
}

const clearUploadFiles = () => {
  uploadPreviews.value.forEach(preview => URL.revokeObjectURL(preview.url))
  uploadPreviews.value = []
  uploadFiles.value = []
}

onBeforeUnmount(clearUploadFiles)

const submitUpload = async () => {
  uploadError.value = null
  uploadSuccess.value = null
  if (!selectedUploadOrderEligible.value) {
    uploadError.value = t('resourcesPictureWarehouseShared.errors.selectOrder')
    return
  }
  if (!uploadRegion.value.trim()) {
    uploadError.value = t('resourcesPictureWarehouseShared.errors.regionRequired')
    return
  }
  if (!uploadFiles.value.length) {
    uploadError.value = t('resourcesPictureWarehouseShared.errors.choosePhoto')
    return
  }
  if (uploadFiles.value.length > 10) {
    uploadError.value = t('resourcesPictureWarehouseShared.errors.maximumFiles')
    return
  }
  const validation = await validateStorefrontUploadFiles(uploadFiles.value, 'user_showcase_image')
  if (!validation.ok) {
    uploadError.value = validation.error || t('resourcesPictureWarehouseShared.errors.invalidFiles')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    uploadFiles.value.forEach(file => formData.append('file[]', file))
    formData.append('order_id', uploadOrderID.value)
    formData.append('region', uploadRegion.value.trim())
    if (uploadLocation.value.trim()) formData.append('location', uploadLocation.value.trim())
    if (uploadNickname.value.trim()) formData.append('nickname', uploadNickname.value.trim())
    if (uploadNotes.value.trim()) formData.append('notes', uploadNotes.value.trim())

    try {
      await auth.request('/showcase/upload', {
        method: 'POST',
        headers: { accept: 'application/json' },
        body: formData,
      }, t('resourcesPictureWarehouseShared.errors.uploadFailed'))
    } catch (error: any) {
      const message = error?.message || t('resourcesPictureWarehouseShared.errors.uploadFailed')
      if (error?.code === 'showcase_upload_order_not_eligible' || message.includes('showcase_upload_order_not_eligible')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.orderNotEligible')
      } else if (error?.code === 'showcase_upload_order_required' || message.includes('showcase_upload_order_required')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.selectOrder')
      } else if (error?.status === 401 || message.includes('401') || message.toLowerCase().includes('login')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.loginRequired')
      } else if (message.includes('429')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.tooManyUploads')
      } else if (message.includes('invalid_type') || message.includes('Only WEBP')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.onlyWebp')
      } else if (message.includes('file_too_large') || message.includes('too large')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.fileTooLarge')
      } else if (message.includes('missing_region')) {
        uploadError.value = t('resourcesPictureWarehouseShared.errors.regionRequired')
      } else {
        uploadError.value = message
      }
      return
    }
    uploadSuccess.value = t('resourcesPictureWarehouseShared.upload.success')
    clearUploadFiles()
    uploadNotes.value = ''
  } catch {
    uploadError.value = t('resourcesPictureWarehouseShared.errors.uploadFailed')
  } finally {
    uploading.value = false
  }
}

const photoLinkPath = (product: PictureWarehouseProductLink) => {
  const slug = String(product.slug || '').trim()
  return slug ? localePath(buildProductPath(slug)) : localePath('/shop')
}
const productLinkPath = photoLinkPath

const toggleUserPhotos = () => {
  if (hasMoreUserPhotos.value) showAllUserPhotos.value = !showAllUserPhotos.value
}
const toggleBrandPhotos = () => {
  if (hasMoreBrandPhotos.value) showAllBrandPhotos.value = !showAllBrandPhotos.value
}

const mapPayloadToUserPhotos = (payload: any[]): RiderPhoto[] => payload
  .map((item: any): RiderPhoto | null => {
    const id = item?.id ?? item?.ID
    if (!id) return null
    const galleryImages = Array.isArray(item.gallery_images)
      ? item.gallery_images
        .map((image: unknown) => normalizeStorefrontMediaUrl(image, mediaContext))
        .filter((image: string): image is string => Boolean(image))
      : []
    return {
      id: String(id),
      kind: 'user',
      title: String(item.title ?? item.post_title ?? 'Rider photo'),
      region: String(item.region ?? 'Unknown'),
      nickname: typeof item.nickname === 'string' ? item.nickname : undefined,
      galleryImages,
    }
  })
  .filter((photo): photo is RiderPhoto => photo !== null)

const fetchUserPhotos = async () => {
  userLoading.value = true
  try {
    const payload = await auth.request<any[]>('/showcase/gallery?type=user&status=approved')
    userPhotos.value = Array.isArray(payload) ? mapPayloadToUserPhotos(payload) : []
  } catch {
    userPhotos.value = []
  } finally {
    userLoading.value = false
  }
}

const fetchUploadOrders = async () => {
  uploadOrdersLoading.value = true
  uploadOrdersError.value = null
  uploadOrders.value = []
  uploadOrderID.value = ''
  try {
    const payload = await auth.request<UploadOrderOption[]>(
      '/showcase/upload-orders',
      {},
      t('resourcesPictureWarehouseShared.errors.ordersLoadFailed'),
    )
    uploadOrders.value = Array.isArray(payload) ? payload : []
  } catch (error: any) {
    uploadOrdersError.value = error?.status === 401 || String(error?.message || '').toLowerCase().includes('login')
      ? t('resourcesPictureWarehouseShared.errors.loginRequired')
      : t('resourcesPictureWarehouseShared.errors.ordersLoadFailed')
  } finally {
    uploadOrdersLoading.value = false
  }
}

watch(() => auth.isAuthenticated.value, (authenticated, wasAuthenticated) => {
  if (authenticated && !wasAuthenticated) void fetchUploadOrders()
  if (!authenticated) {
    uploadOrders.value = []
    uploadOrderID.value = ''
    clearUploadFiles()
  }
})

onMounted(() => {
  void fetchUserPhotos()
  void fetchBrandPhotos()
  void fetchUploadOrders()
})

const loadActiveBrandGalleryDetails = () => {
  if (activeKind.value === 'brand' && activeIndex.value !== null) {
    void loadBrandGalleryDetails(activeIndex.value)
  }
}
const loadCommentsForActivePhoto = async () => {
  const current = activePhoto.value
  if (!current || current.kind !== 'user') {
    comments.value = []
    commentsError.value = null
    commentsLoading.value = false
    return
  }
  commentsLoading.value = true
  commentsError.value = null
  try {
    const payload = await auth.request<any[]>(`/showcase/comments?photo_id=${encodeURIComponent(current.id)}&per_page=20`)
    comments.value = Array.isArray(payload) ? payload.map((item: any): PhotoComment => {
      const rawDate = String(item.date_gmt ?? '')
      const date = rawDate ? new Date(`${rawDate}Z`) : null
      return {
        id: Number(item.id ?? 0),
        author: String(item.author ?? 'Anonymous'),
        content: String(item.content ?? ''),
        dateGmt: rawDate,
        dateGmtFormatted: date && !Number.isNaN(date.getTime())
          ? date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
          : rawDate,
        location: item.location ? String(item.location) : '',
      }
    }) : []
  } catch {
    commentsError.value = 'load_failed'
    comments.value = []
  } finally {
    commentsLoading.value = false
  }
}

const openLightbox = (kind: 'user' | 'brand', index: number) => {
  activeKind.value = kind
  activeIndex.value = index
  currentGalleryIndex.value = 0
  if (kind === 'user') void loadCommentsForActivePhoto()
  else loadActiveBrandGalleryDetails()
}
const closeLightbox = () => {
  activeKind.value = null
  activeIndex.value = null
  comments.value = []
  commentsError.value = null
  commentsLoading.value = false
  commentContent.value = ''
  commentLocation.value = ''
  commentSuccess.value = null
  commentError.value = null
}
const moveLightbox = (direction: 1 | -1) => {
  if (!activeList.value?.length || activeIndex.value === null) return
  activeIndex.value = (activeIndex.value + direction + activeList.value.length) % activeList.value.length
  currentGalleryIndex.value = 0
  if (activeKind.value === 'user') void loadCommentsForActivePhoto()
  else loadActiveBrandGalleryDetails()
}
const goNext = () => moveLightbox(1)
const goPrev = () => moveLightbox(-1)

const submitComment = async () => {
  commentError.value = null
  commentSuccess.value = null
  const current = activePhoto.value
  if (!current) {
    commentError.value = t('resourcesPictureWarehouseShared.lightbox.noPhotoSelected')
    return
  }
  if (current.kind !== 'user') {
    commentError.value = t('resourcesPictureWarehouseShared.lightbox.riderCommentsOnly')
    return
  }
  if (!commentContent.value.trim()) {
    commentError.value = t('resourcesPictureWarehouseShared.lightbox.commentRequired')
    return
  }
  commentSubmitting.value = true
  try {
    const body: Record<string, unknown> = {
      photo_id: Number(current.id),
      content: commentContent.value.trim(),
    }
    if (commentLocation.value.trim()) body.location = commentLocation.value.trim()
    try {
      await auth.request('/showcase/comments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', accept: 'application/json' },
        body: JSON.stringify(body),
      }, t('resourcesPictureWarehouseShared.lightbox.commentSubmitFailed'))
    } catch (error: any) {
      const message = error?.message || t('resourcesPictureWarehouseShared.lightbox.commentSubmitFailed')
      commentError.value = message.includes('401') || message.includes('403') || message.toLowerCase().includes('login')
        ? t('resourcesPictureWarehouseShared.lightbox.loginToComment')
        : message.includes('empty_comment') || message.includes('cannot be empty')
          ? t('resourcesPictureWarehouseShared.lightbox.commentRequired')
          : message
      return
    }
    commentSuccess.value = t('resourcesPictureWarehouseShared.lightbox.commentSubmitted')
    commentContent.value = ''
    commentLocation.value = ''
  } catch {
    commentError.value = t('resourcesPictureWarehouseShared.lightbox.commentSubmitFailed')
  } finally {
    commentSubmitting.value = false
  }
}

const copyShareLink = async () => {
  shareMessage.value = null
  const current = activePhoto.value
  if (!current) {
    shareMessage.value = t('resourcesPictureWarehouseShared.lightbox.noPhotoSelected')
    return
  }
  let shareUrl = ''
  try {
    const url = new URL(window.location.href)
    url.searchParams.set('photo', current.id)
    url.searchParams.set('kind', current.kind)
    shareUrl = url.toString()
  } catch {
    shareMessage.value = t('resourcesPictureWarehouseShared.lightbox.unableToBuildLink')
    return
  }
  shareCopying.value = true
  try {
    if (navigator.clipboard?.writeText) await navigator.clipboard.writeText(shareUrl)
    else {
      const textarea = document.createElement('textarea')
      textarea.value = shareUrl
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    shareMessage.value = t('resourcesPictureWarehouseShared.lightbox.linkCopied')
  } catch {
    shareMessage.value = t('resourcesPictureWarehouseShared.lightbox.copyFailed')
  } finally {
    shareCopying.value = false
  }
}
</script>

<style scoped>
.picture-warehouse-page {
  width: 100%;
}

.picture-lightbox-panel {
  max-height: min(90vh, var(--tz-mobile-safe-viewport-height, 90vh));
}

.picture-lightbox-image {
  max-height: min(65vh, calc(var(--tz-mobile-safe-viewport-height, 100vh) - 12rem));
}
</style>

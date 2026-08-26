import { ref } from 'vue'
import { toast } from 'vue-sonner'
import mediaApi from '@/api/media'
import { validateUploadFile } from '@/lib/uploadSpecs'
import {
  createProductMediaItem,
  getProductMediaRoleOptions,
  getProductMediaTypeLabel,
  normalizeProductMediaForPayload,
  normalizeProductMediaOrder
} from '@/lib/productMedia'

const imageVariantsFromAsset = (asset: any): Record<string, any> => {
  const derivatives = Array.isArray(asset?.derivatives) ? asset.derivatives : []
  const result: Record<string, any> = {}
  derivatives.forEach((derivative) => {
    const preset = String(derivative?.preset || '').trim()
    const url = String(derivative?.url || '').trim()
    if (!preset || !url) return
    result[preset] = {
      url,
      width: Number(derivative?.width || 0) || undefined,
      height: Number(derivative?.height || 0) || undefined,
      mime_type: String(derivative?.mime_type || '').trim() || undefined
    }
  })
  return result
}

const thumbnailFromImageVariants = (variants: Record<string, any>): string => (
  variants.thumbnail?.url || variants.card?.url || variants.large?.url || ''
)

export const useProductMediaManager = (productForm: any, options: Record<string, any> = {}) => {
  const uploadingMedia = ref(false)
  const clearFieldError = options.clearFieldError || (() => {})

  const normalizeMediaOrder = () => normalizeProductMediaOrder(productForm.media)
  const mediaTypeLabel = getProductMediaTypeLabel
  const mediaRoleOptions = getProductMediaRoleOptions

  const addMediaUrl = (type: string) => {
    const hasPrimaryImage = productForm.media.some((item) => item.media_type === 'image' && item.is_primary)
    productForm.media.push(createProductMediaItem({
      media_type: type,
      role: type === 'video' ? 'video' : hasPrimaryImage ? 'gallery' : 'primary',
      is_primary: type === 'image' && !hasPrimaryImage
    }, productForm.media.length * 10))
    normalizeMediaOrder()
    clearFieldError('media')
  }

  const appendUploadedMedia = (asset: any, type: string) => {
    const mediaType = asset?.media_type || type
    const imageVariants = mediaType === 'image' ? imageVariantsFromAsset(asset) : {}
    const hasPrimaryImage = productForm.media.some((item) => item.media_type === 'image' && item.is_primary)
    productForm.media.push(createProductMediaItem({
      media_asset_id: asset?.id || null,
      media_type: mediaType,
      role: mediaType === 'video' ? 'video' : hasPrimaryImage ? 'gallery' : 'primary',
      url: asset?.url || '',
      thumbnail_url: mediaType === 'image' ? thumbnailFromImageVariants(imageVariants) : '',
      image_variants: imageVariants,
      alt: asset?.alt || '',
      title: asset?.original_filename || asset?.filename || '',
      is_primary: mediaType === 'image' && !hasPrimaryImage
    }, productForm.media.length * 10))
    normalizeMediaOrder()
  }

  const handleMediaUpload = async (event: Event, type: string) => {
    const input = event.target as HTMLInputElement
    const files = Array.from(input.files || [])
    input.value = ''
    if (!files.length) return

    uploadingMedia.value = true
    try {
      for (const file of files) {
        if (type === 'image') {
          const validation = await validateUploadFile(file, 'product_image')
          if (!validation.ok) {
            toast.error(validation.error || '商品图片不符合上传规范')
            continue
          }
          if (validation.warning) toast.warning(validation.warning)
        }
        const formData = new FormData()
        formData.append('file', file)
        formData.append('media_type', type)
        if (type === 'image') formData.append('image_purpose', 'product_image')
        const asset = await mediaApi.uploadAsset(formData)
        appendUploadedMedia(asset, type)
      }
      clearFieldError('media')
      toast.success(`${files.length} 个商品媒体已上传`)
    } catch (error) {
      console.error('Failed to upload product media:', error)
      toast.error('商品媒体上传失败，请检查文件类型和大小')
    } finally {
      uploadingMedia.value = false
    }
  }

  const setPrimaryMedia = (index: number) => {
    productForm.media.forEach((item, currentIndex) => {
      if (item.media_type !== 'image') return
      item.is_primary = currentIndex === index
      item.role = currentIndex === index ? 'primary' : (item.role === 'primary' ? 'gallery' : item.role)
    })
  }

  const moveMedia = (index: number, direction: number) => {
    const nextIndex = index + direction
    if (nextIndex < 0 || nextIndex >= productForm.media.length) return
    const [item] = productForm.media.splice(index, 1)
    productForm.media.splice(nextIndex, 0, item)
    normalizeMediaOrder()
  }

  const removeMedia = (index: number) => {
    const [removed] = productForm.media.splice(index, 1)
    if (removed?.is_primary) {
      const nextImage = productForm.media.find((item) => item.media_type === 'image')
      if (nextImage) {
        nextImage.is_primary = true
        nextImage.role = 'primary'
      }
    }
    normalizeMediaOrder()
  }

  const normalizeFormMedia = () => normalizeProductMediaForPayload(productForm.media)

  return {
    uploadingMedia,
    mediaTypeLabel,
    mediaRoleOptions,
    addMediaUrl,
    handleMediaUpload,
    setPrimaryMedia,
    moveMedia,
    removeMedia,
    normalizeFormMedia
  }
}

export default useProductMediaManager

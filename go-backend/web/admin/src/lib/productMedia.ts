export const getPublicSiteOrigin = () => {
  const configured = String(import.meta.env.VITE_PUBLIC_SITE_URL || '').trim().replace(/\/+$/, '')
  if (configured) return configured

  if (typeof window !== 'undefined' && window.location?.hostname) {
    const { protocol, hostname, port } = window.location
    if (hostname.startsWith('admin.')) {
      return `${protocol}//${hostname.replace(/^admin\./, '')}${port ? `:${port}` : ''}`
    }
  }

  return 'https://tanzanite.site'
}

export const resolveProductMediaUrl = (url: unknown) => {
  const value = String(url || '').trim()
  if (!value) return ''
  if (/^(?:https?:)?\/\//i.test(value) || /^data:/i.test(value) || /^blob:/i.test(value)) return value

  const origin = getPublicSiteOrigin()
  const path = value.startsWith('/') ? value : `/${value}`
  return `${origin}${path}`
}

export const getProductThumbnail = (product: any) => {
  const mediaItems = Array.isArray(product?.media) ? product.media : []
  const visibleItems = mediaItems.filter((item) => item && item.is_visible !== false)
  const hasUrl = (item) => String(item?.url || '').trim().length > 0

  const primaryImage = visibleItems.find((item) => (
    item.media_type === 'image' && hasUrl(item) && (item.is_primary || item.role === 'primary')
  ))
  const fallbackImage = visibleItems.find((item) => item.media_type === 'image' && hasUrl(item))
  const image = primaryImage || fallbackImage

  if (image) {
    return {
      kind: 'image',
      src: resolveProductMediaUrl(image.url),
      alt: String(image.alt || image.title || product?.name || '商品图片').trim(),
      label: image.is_primary || image.role === 'primary' ? '主图' : '图片'
    }
  }

  const primaryVideo = visibleItems.find((item) => (
    item.media_type === 'video' && hasUrl(item) && (item.is_primary || item.role === 'video' || item.role === 'detail')
  ))
  const fallbackVideo = visibleItems.find((item) => item.media_type === 'video' && hasUrl(item))
  const video = primaryVideo || fallbackVideo

  if (video) {
    return {
      kind: 'video',
      src: resolveProductMediaUrl(video.poster_url || video.thumbnail_url || ''),
      alt: String(video.alt || video.title || product?.name || '商品视频').trim(),
      label: '视频'
    }
  }

  return {
    kind: 'empty',
    src: '',
    alt: String(product?.name || '商品').trim(),
    label: '无图'
  }
}

export const createProductMediaItem = (overrides: Record<string, any> = {}, sortOrder = 0) => ({
  id: null,
  local_key: `media-${Date.now()}-${Math.random().toString(16).slice(2)}`,
  variant_id: null,
  media_asset_id: null,
  media_type: 'image',
  role: 'gallery',
  url: '',
  thumbnail_url: '',
  poster_url: '',
  alt: '',
  title: '',
  locale: '',
  sort_order: sortOrder,
  is_primary: false,
  is_visible: true,
  ...overrides
})

export const buildProductMediaFormValues = (product: any) => (
  (product.media || []).map((item, index) => createProductMediaItem({
    id: item.id || null,
    variant_id: item.variant_id || null,
    media_asset_id: item.media_asset_id || null,
    media_type: item.media_type || 'image',
    role: item.role || (item.media_type === 'video' ? 'video' : 'gallery'),
    url: item.url || '',
    thumbnail_url: item.thumbnail_url || '',
    poster_url: item.poster_url || '',
    alt: item.alt || '',
    title: item.title || '',
    locale: item.locale || '',
    sort_order: item.sort_order ?? index * 10,
    is_primary: Boolean(item.is_primary),
    is_visible: item.is_visible !== false
  }, index * 10))
)

export const getProductMediaTypeLabel = (type: string) => ({ image: '图片', video: '视频' } as Record<string, string>)[type] || type

export const getProductMediaRoleOptions = (type: string) => type === 'video'
  ? [
      { label: '商品视频', value: 'video' },
      { label: '详情视频', value: 'detail' }
    ]
  : [
      { label: '主图', value: 'primary' },
      { label: '轮播图', value: 'gallery' },
      { label: '详情图', value: 'detail' }
    ]

export const normalizeProductMediaOrder = (items: Array<Record<string, any>>) => {
  items.forEach((item, index) => {
    item.sort_order = index * 10
  })
}

export const ensureSinglePrimaryProductImage = (items: Array<Record<string, any>>) => {
  let primaryIndex = items.findIndex((item) => (
    item.media_type === 'image' && (item.is_primary || item.role === 'primary')
  ))
  if (primaryIndex === -1) {
    primaryIndex = items.findIndex((item) => item.media_type === 'image' && String(item.url || '').trim())
  }
  items.forEach((item, index) => {
    if (item.media_type !== 'image') return
    const isPrimary = index === primaryIndex
    item.is_primary = isPrimary
    if (isPrimary) {
      item.role = 'primary'
    } else if (item.role === 'primary') {
      item.role = 'gallery'
    }
  })
}

export const normalizeProductMediaForPayload = (items: Array<Record<string, any>>) => {
  ensureSinglePrimaryProductImage(items)
  return items
    .filter((item) => String(item.url || '').trim())
    .map((item, index) => ({
      id: item.id || undefined,
      variant_id: item.variant_id || undefined,
      media_asset_id: item.media_asset_id || undefined,
      media_type: item.media_type || 'image',
      role: item.role || (item.media_type === 'video' ? 'video' : 'gallery'),
      url: String(item.url || '').trim(),
      thumbnail_url: String(item.thumbnail_url || '').trim(),
      poster_url: String(item.poster_url || '').trim(),
      alt: String(item.alt || '').trim(),
      title: String(item.title || '').trim(),
      locale: String(item.locale || '').trim(),
      sort_order: Number(item.sort_order ?? index * 10),
      is_primary: Boolean(item.is_primary),
      is_visible: item.is_visible !== false
    }))
}

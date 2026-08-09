export interface MediaAssetLike {
  id?: number | string | null
  alt?: string | null
  original_filename?: string | null
  filename?: string | null
  access_url?: string | null
  url?: string | null
}

export const assetTitle = (asset?: MediaAssetLike | null): string => (
  asset?.alt
  || asset?.original_filename
  || asset?.filename
  || `媒体 #${asset?.id || '-'}`
)

export const assetAccessURL = (asset?: MediaAssetLike | null): string => asset?.access_url || asset?.url || ''

export const statusLabel = (status?: string | null): string => status === 'archived' ? '归档' : '启用'

export const mediaTypeLabel = (type?: string | null): string => type === 'video' ? '视频' : '图片'

export const formatMediaDate = (date?: string | number | Date | null): string => date ? new Date(date).toLocaleDateString('zh-CN') : '-'

export const formatMediaSize = (bytes?: number | string | null): string => {
  const value = Number(bytes || 0)
  if (value >= 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`
  if (value >= 1024) return `${Math.round(value / 1024)} KB`
  return `${value} B`
}
